package stl

import "github.com/chewxy/stl/loess"

// Config is a configuration structure
type Config struct {
	// Width represents the width of the LOESS smoother in data points
	Width int

	// Jump is the number of points to skip between smoothing,
	Jump int

	// Which weight updating function should be used?
	Fn loess.WeightUpdate
}

// Opt is a function that helps build the conf
type Opt func(*state)

// WithSeasonalConfig configures the seasonal component of the decomposition process.
func WithSeasonalConfig(conf Config) Opt {
	return func(s *state) {
		s.sConf = conf
	}
}

// WithTrendConfig configures the trend component of the decomposition process.
func WithTrendConfig(conf Config) Opt {
	return func(s *state) {
		s.tConf = conf
	}
}

// WithLowpassConfig configures the operation that performs the lowpass filter in the decomposition process
func WithLowpassConfig(conf Config) Opt {
	return func(s *state) {
		s.lConf = conf
	}
}

// WithRobustIter indicates how many iterations of "robust" (i.e. outlier removal) to do.
// The default is 0.
func WithRobustIter(n int) Opt {
	return func(s *state) {
		s.robustIter = n
	}
}

// WithIter indicates how many iterations to run.
// The default is 2.
func WithIter(n int) Opt {
	return func(s *state) {
		s.innerIter = n
	}
}

// DefaultSeasonal returns the default configuration for the operation that works on the seasonal component.
func DefaultSeasonal(width int) Config {
	return Config{
		Width: width,
		Jump:  int(0.1 * float64(width)),
		Fn:    loess.Linear,
	}
}

// DefaulTrend returns the default configuration for the operation that works on the trend component.
func DefaultTrend(periodicity, width int) Config {
	return Config{
		Width: trendWidth(periodicity, width),
		Jump:  int(0.1 * float64(width)),
		Fn:    loess.Linear,
	}
}

// DefaultLowpass returns the default configuration for the operation that works on the lowpass component.
func DefaultLowPass(periodicity int) Config {
	return Config{
		Width: periodicity,
		Jump:  int(0.1 * float64(periodicity)),
		Fn:    loess.Linear,
	}
}

// from the original paper's numerical analysis section
func trendWidth(periodicity int, seasonalWidth int) int {
	p := float64(periodicity)
	w := float64(seasonalWidth)

	return int(1.5*p/(1-1.5/w) + 0.5)
}
