package stl

import (
	"math"

	"github.com/chewxy/stl/loess"
	"github.com/pkg/errors"
	"gorgonia.org/tensor"
)

type subcycleState struct {
	// (2, P, L) tensor - first 2 are rawdata and weight
	// P: period length/periodicity
	// L cycle length = periods+1
	data     *tensor.Dense
	smoothed *tensor.Dense // (P, L+Fwd+Bwd+1)

	periodicity, periods, rem, fwd, bwd int

	Config // for any loess smoothing
}

// fwd = fowward extrapolation period
// bwd = backfill period
func newSubcycleState(conf Config, size, periodicity, fwd, bwd int) *subcycleState {
	periods := size / periodicity
	rem := size % periodicity
	cycleLength := periods + 1
	if rem == 0 {
		cycleLength = periods
	}
	smoothedLength := cycleLength + fwd + bwd
	retVal := &subcycleState{
		data:     tensor.New(tensor.WithShape(2, periodicity, cycleLength), tensor.Of(tensor.Float64), tensor.WithEngine(tensor.Float64Engine{})),
		smoothed: tensor.New(tensor.WithShape(periodicity, smoothedLength), tensor.Of(tensor.Float64), tensor.WithEngine(tensor.Float64Engine{})),

		periodicity: periodicity,
		periods:     periods,
		rem:         rem,
		fwd:         fwd,
		bwd:         bwd,

		Config: conf,
	}

	return retVal
}

func (s *subcycleState) smoothSeasonal(X []float64, weights []float64, retVal []float64) error {
	s.setupWorkspace(X, weights)
	if err := s.computeSmoothedSubSeries(); err != nil {
		return err
	}
	copy(retVal, s.smoothed.Data().([]float64))
	return nil
}

// setupWorkspace sets up the workspace by copying the data to the tensor.
// Due to the periodicity in question, a direct copy wouldn't really work
func (s *subcycleState) setupWorkspace(X, weights []float64) {
	// populate data
	// strides := s.data.Strides()
	// raw := s.data.Data().([]float64)
	// copy(raw[0:strides[0]], X)
	// copy(raw[strides[0]:], weights)

	var xxx [][][]float64
	var err error
	if xxx, err = tensor.Native3TensorF64(s.data); err != nil {
		panic(errors.Wrap(err, "setup workspace"))
	}

	for p := 0; p < s.periodicity; p++ {
		l := s.periods
		if p < s.rem {
			l = s.periods + 1
		}
		for i := 0; i < l; i++ {
			xxx[0][p][i] = X[i*s.periodicity+p]
			if len(weights) > 0 {
				xxx[1][p][i] = weights[i*s.periodicity+p]
			}
		}
	}
}

func (s *subcycleState) computeSmoothedSubSeries() (err error) {
	var xxx [][][]float64
	var xx [][]float64
	if xxx, err = tensor.Native3TensorF64(s.data); err != nil {
		return errors.Wrap(err, "Compute Smoothed Series")
	}
	if xx, err = tensor.NativeMatrixF64(s.smoothed); err != nil {
		return errors.Wrap(err, "Compute Smoothed Series")
	}

	periods := len(xxx[0])
	for p := 0; p < periods; p++ {
		data := xxx[0][p]
		weights := xxx[1][p]
		smoothed := xx[p]

		s.do(weights, data, smoothed)
	}
	return nil
}

func (s *subcycleState) do(weights, data, smoothed []float64) {
	cycleLength := float64(len(data))
	l := loess.NewWithExternal(s.Width, data, weights)

	if _, err := loess.UnsafeSmooth(l, s.Config.Width, s.Config.Jump, loess.Linear, smoothed[1:]); err != nil {
		panic(err)
	}

	var left, right float64
	width := float64(s.Width)
	right = left + width - 1
	right = math.Min(cycleLength-1, right)
	leftVal := s.bwd

	for i := 1; i <= s.bwd; i++ {
		j := -float64(i)
		point, err := loess.Regress(l, loess.Linear, j, left, right)
		if err != nil {
			smoothed[leftVal-i] = smoothed[leftVal]
		} else {
			smoothed[leftVal-i] = point
		}
	}

	right = cycleLength - 1
	left = right - width + 1
	left = math.Max(0, left)
	rightVal := s.bwd + int(right)

	for i := 1; i <= s.fwd; i++ {
		j := float64(i)
		point, err := loess.Regress(l, loess.Linear, right+j, left, right)
		if err != nil {
			smoothed[rightVal+i] = smoothed[rightVal]
		} else {
			smoothed[rightVal+i] = point
		}
	}
}
