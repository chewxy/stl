// Package stl implements functions and data structures necessary to perform a Seasonal-Trend decomposition by LOESS, as described by Cleveland et al. (1990).
//
// This library's code was ported over from the original  Netlib library which was written in Fortran '77.
package stl

import (
	"math"

	"github.com/pkg/errors"
)

// ModelType is the type of STL model we would like to perform. A STL model type is usually additive, or multiplicative,
// however, it can be somewhere in between. This is done by means of a Box-Cox transform (and the reverse when we're done applying STL)
type ModelType struct {
	Fwd BoxCox
	Bwd IsoBoxCox
}

// BoxCox is any function that performs a box cox transform
type BoxCox func(a []float64) []float64

// IsoBoxCox is the invese of the BoxCox
type IsoBoxCox func(a []float64) []float64

// Additive returns a BoxCox transform that does nothing.
func Additive() ModelType {
	return ModelType{func(a []float64) []float64 { return a }, func(a []float64) []float64 { return a }}
}

// Multiplicative returns a BoxCox transform that performs and unsafe transform of the input slice
func Multiplicative() ModelType {
	return ModelType{
		Fwd: func(a []float64) []float64 {
			for i := range a {
				a[i] = math.Log(a[i])
			}
			return a
		},
		Bwd: func(a []float64) []float64 {
			for i := range a {
				a[i] = math.Exp(a[i])
			}
			return a
		},
	}

}

// UnsafeTransform creates a transformation function that is somewhere between an additive and multiplicative model
func UnsafeTransform(lambda float64) ModelType {
	switch lambda {
	case 1.0:
		return Multiplicative()
	case 0.0:
		return Additive()
	}
	if lambda < 0 {
		panic("Lambda cannot be less than 0")
	}

	return ModelType{
		Fwd: func(a []float64) []float64 {
			for i := range a {
				a[i] = (math.Pow(a[i], lambda) - 1) / lambda
			}
			return a
		},
		Bwd: func(a []float64) []float64 {
			for i := range a {
				a[i] = math.Pow((a[i]*lambda + 1), (1 / lambda))
			}
			return a
		},
	}
}

// Decompose performs a STL decomposition.
//
// Following the R library conventions, this function uses a default of 2 iterations and 0 robust iterations.
func Decompose(X []float64, periodicity, width int, m ModelType, opts ...Opt) Result {
	if periodicity < 2 {
		return Result{Err: errors.Errorf("Periodicity must be greater than 2")}
	}
	if width < 1 {
		return Result{Err: errors.Errorf("Width must be greater than 1")}
	}

	bc := m.Fwd
	ibc := m.Bwd

	// transform the data
	X = bc(X)
	s := newState(X, periodicity, width, opts...)
	var useResidualWeights bool
	for o := 0; o <= s.robustIter; o++ {
		useResidualWeights = o > 0
		for i := 0; i < s.innerIter; i++ {
			s.doDetrend()

			if err := s.smoothSubcycles(useResidualWeights); err != nil {
				s.Err = errors.Wrapf(err, "Outer Iteration: %d, Inner Iteration %d - Failed to smooth subcycles", o, i)
				return s.Result
			}

			if err := s.removeSeasonality(); err != nil {
				s.Err = errors.Wrapf(err, "Outer Iteration: %d, Inner Iteration %d - Failed to remove seasonality", o, i)
				return s.Result
			}

			if err := s.updateSeasonalAndTrend(useResidualWeights); err != nil {
				s.Err = errors.Wrapf(err, "Outer Iteration: %d, Inner Iteration %d - Failed to update seasonal and trend", o, i)
				return s.Result
			}
		}
		s.updateWeights()
	}
	updateResiduals(&s.Result)

	// untransform the data
	s.Result.Data = ibc(s.Result.Data)
	s.Result.Seasonal = ibc(s.Result.Seasonal)
	s.Result.Trend = ibc(s.Result.Trend)
	s.Result.Resid = ibc(s.Result.Resid)
	return s.Result
}
