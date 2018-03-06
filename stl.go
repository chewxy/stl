package stl

import (
	"math"
	"sort"

	"github.com/chewxy/stl/loess"
)

type state struct {
	Result
	weights     []float64
	periodicity int                // >2
	width       int                // seasonal width for loess smoothing
	smoothFn    loess.WeightUpdate // smoothing function

	sConf Config // seasonal config
	tConf Config // trend config
	lConf Config // lowpass filter config

	innerIter  int
	robustIter int

	detrend          []float64
	extendedSeasonal []float64
	deseasonalized   []float64

	sstate  *loess.State
	scstate *subcycleState
}

// Result is the result of a decompositon
type Result struct {
	Data     []float64
	Trend    []float64
	Seasonal []float64
	Resid    []float64
	Err      error
}

func newState(data []float64, periodicity, width int, opts ...Opt) *state {
	s := &state{
		periodicity: periodicity,
		width:       width,
		smoothFn:    loess.Linear,

		sConf: DefaultSeasonal(width),
		tConf: DefaultTrend(periodicity, width),
		lConf: DefaultLowPass(periodicity),

		// R interface sets these to be the default
		innerIter:  2,
		robustIter: 0,
	}
	s.Data = data
	s.Trend = make([]float64, len(data))
	s.Seasonal = make([]float64, len(data))
	s.Resid = make([]float64, len(data))
	s.weights = make([]float64, len(data))
	s.extendedSeasonal = make([]float64, len(data)+2*periodicity)
	s.detrend = make([]float64, len(data))

	for i := range s.weights {
		s.weights[i] = 1
	}
	s.sstate = loess.NewWithExternal(s.tConf.Width, s.Trend, s.weights)
	s.scstate = newSubcycleState(s.sConf, len(s.Data), s.periodicity, 1, 1)

	for _, o := range opts {
		o(s)
	}
	return s
}

func updateResiduals(r *Result) {
	for i := range r.Data {
		r.Resid[i] = r.Data[i] - r.Seasonal[i] - r.Trend[i]
	}
}

// updateWeights is used to calculate the weights based on the residuals. This is used during a robust calculation to remove outliers
//
// It is entirely reliant on finding the median absolute deviation
func (s *state) updateWeights() {
	for i := range s.Data {
		s.Resid[i] = s.Data[i] - s.Seasonal[i] - s.Trend[i]
		s.weights[i] = math.Abs(s.Resid[i])
	}
	sort.Float64s(s.weights)

	// find the upper and lower bounds of the median deviation
	med0 := (len(s.Data)+1)/2 - 1
	med1 := (len(s.Data) - med0 - 1)
	mad6 := 6 * (s.weights[med0] + s.weights[med1]) // 6 times the MAD

	// numerical stability
	ceil := 0.999 * mad6
	flor := 0.001 * mad6

	for i := range s.Data {
		a := math.Abs(s.Resid[i])
		if a <= flor {
			s.weights[i] = 1
		} else if a <= ceil {
			h := a / mad6
			w := 1 - h*h
			s.weights[i] = w * w
		} else {
			s.weights[i] = 0
		}
	}
}

func (s *state) doDetrend() {
	for i := range s.Data {
		s.detrend[i] = s.Data[i] - s.Trend[i]
	}
}

func (s *state) smoothSubcycles(useWeights bool) error {
	var weights []float64
	if useWeights {
		weights = s.weights
	}
	return s.scstate.smoothSeasonal(s.detrend, weights, s.extendedSeasonal)
}

func (s *state) removeSeasonality() (err error) {
	passes := ma(ma(ma(s.extendedSeasonal, s.periodicity), s.periodicity), 3)
	s.deseasonalized, err = loess.Smooth(passes, s.lConf.Width, s.lConf.Jump, s.lConf.Fn)
	return err
}

func (s *state) updateSeasonalAndTrend(useWeights bool) (err error) {
	for i := range s.Seasonal {
		s.Seasonal[i] = s.extendedSeasonal[s.periodicity+1] - s.deseasonalized[i]
		s.Trend[i] = s.Data[i] - s.Seasonal[i]
	}

	s.Trend, err = loess.UnsafeSmooth(s.sstate, s.tConf.Width, s.tConf.Jump, s.tConf.Fn, s.Trend)
	return
}

// ma is a moving average
func ma(data []float64, window int) (retVal []float64) {
	retSize := len(data) - window + 1
	retVal = make([]float64, retSize)

	var sum float64
	w := float64(window)
	firstRange := data[0:window]
	for i := range firstRange {
		sum += firstRange[i]
	}
	retVal[0] = sum / w

	var start int
	end := window
	for i := 1; i < retSize; i++ {
		sum += data[end] - data[start]
		start++
		end++
		retVal[i] = sum / w
	}
	return retVal
}
