package loess

import (
	"math"

	"github.com/pkg/errors"
)

var errTotal = errors.New("Total <= 0")

// WeightUpdate is a function that modifies the State.
type WeightUpdate func(s *State, x, left, right float64) error

// State represents a state for regression. Every field will be mutated by the Regress and Smooth functions
type State struct {
	width int
	w     []float64 // weights
	x     []float64 // data
	e     []float64 // external weights

}

// New creates a new LOESS state
func New(width int, x []float64) *State {
	return &State{
		width: width,
		x:     x,
		w:     make([]float64, len(x)),
	}
}

// NewWithExternal creates a new LOESS state with externally provided weights
func NewWithExternal(width int, x, e []float64) *State {
	return &State{
		width: width,
		w:     make([]float64, len(x)),
		x:     x,
		e:     e,
	}
}

// W returns the weights in the state
func (s State) W() []float64 { return s.w }

// X returns the Xs in the state.
func (s State) X() []float64 { return s.x }

// W returns the externally defined weights in the state.
func (s State) E() []float64 { return s.e }

// Regress performs local regression of x between left and right
func Regress(s *State, fn WeightUpdate, x, left, right float64) (retVal float64, err error) {
	if err = s.localWeights(x, left, right); err != nil {
		return
	}

	if err = fn(s, x, left, right); err != nil {
		return
	}

	l, r := int(left), int(right)
	for i := l; i <= r; i++ {
		retVal += s.w[i] * s.x[i]
	}
	return
}

// Smooth smooths a slice of float64, with the provided width, jumps and update functions
func Smooth(x []float64, width, jump int, fn WeightUpdate) ([]float64, error) {
	s := New(width, x)
	retVal := make([]float64, len(x))
	return smooth(s, width, jump, fn, retVal)
}

// UnsafeSmooth is like Smooth, except the state is passed in, as well as a optional return value to be mutated.
func UnsafeSmooth(regression *State, width, jump int, fn WeightUpdate, retVal []float64) ([]float64, error) {
	if regression.width != width {
		return nil, errors.Errorf("Regression width: %d. Smoothing width %d", regression.width, width)
	}
	if retVal == nil {
		retVal = make([]float64, len(regression.x))
	}

	if len(retVal) < len(regression.x) {
		return nil, errors.Errorf("Expected the preallocated value to have at least %d elements. Got %d elements.", len(regression.x), len(retVal))
	}

	return smooth(regression, width, jump, fn, retVal)
}

func smooth(s *State, width, jump int, fn WeightUpdate, retVal []float64) ([]float64, error) {
	x := s.x
	size := len(x)
	if size == 1 {
		retVal[0] = x[0]
		return retVal, nil
	}

	left, right := -1.0, -1.0
	half := (width + 1) / 2
	switch {
	case width >= size:
		left = 0
		right = float64(size - 1)

		for i := 0; i < size; i += jump {
			j := float64(i)
			if point, err := Regress(s, fn, j, left, right); err == nil {
				retVal[i] = point
			} else {
				retVal[i] = x[i]
			}
			if i+jump > size {
				break
			}
		}
	case jump == 1:
		left = 0
		right = float64(width - 1)
		for i := range x {
			j := float64(i)
			if i >= half && int(right) != size-1 {
				left++
				right++
			}
			if point, err := Regress(s, fn, j, left, right); err == nil {
				retVal[i] = point
			} else {
				retVal[i] = x[i]
			}
		}
	default:
		for i := 0; i < size; i += jump {
			j := float64(i)
			switch {
			case i < half-1:
				left = 0
			case i >= size-half:
				left = float64(size - width)
			default:
				left = float64(i - half + 1)
			}

			right = left + float64(width-1)
			if point, err := Regress(s, fn, j, left, right); err == nil {
				retVal[i] = point
			} else {
				retVal[i] = x[i]
			}
		}

	}

	if jump != 1 {
		jmp := float64(jump)
		for i := 0; i < size-jump; i += jump {
			slope := (retVal[i+jump] - retVal[i]) / jmp
			for j := i + 1; j < i+jump; j++ {
				retVal[j] = retVal[i] + slope*float64(j-i)
			}
		}
		last := size - 1
		lastSmoothedPos := (last / jump) * jump
		if lastSmoothedPos != last {
			if point, err := Regress(s, fn, float64(last), left, right); err == nil {
				retVal[last] = point
			} else {
				retVal[last] = x[last]
			}

			if lastSmoothedPos != last-1 {
				slope := (retVal[last] - retVal[lastSmoothedPos]) / float64(last-lastSmoothedPos)
				for j := lastSmoothedPos + 1; j < last; j++ {
					retVal[j] = retVal[lastSmoothedPos] + slope*float64(j-lastSmoothedPos)
				}

			}
		}
	}
	return retVal, nil
}

// Local Weights perform local weight smoothing via the tricube function
// left >= 1
func (s *State) localWeights(x, left, right float64) error {
	lambda := math.Max(x-left, right-x)

	// lambda is adjusted to conform to roughly width /2
	//
	// However the netlib fortran code seems to indicate that it's lambda = (width-n) / 2
	// After checking around, it seems that most implementations use something like the below instead
	X := s.X()
	if s.width > len(X) {
		lambda += float64(s.width-len(X)) / 2.0
	}

	if lambda <= 0 {
		return errors.Errorf("Lambda %v", lambda)
	}

	// Numerical stabilization
	ceil := 0.99999 * lambda
	flor := 0.00001 * lambda

	// compute neighbourhood weights, update with external weights if supplied
	var sum float64
	W := s.w
	E := s.e
	for i := left; i <= right; i++ {
		j := int(i)
		delta := math.Abs(x - i)

		// compute tricube neightborhood weight
		// This loop was optimized to "inline" the tricube function to reduce call overheads
		var w float64
		if delta <= ceil {
			// don't combine this and the previous if statements without checking!
			// at this point, this two if-statements is clearer than just having it in one.
			if delta <= flor {
				w = 1
			} else {
				frac := delta / lambda
				trix := 1.0 - frac*frac*frac
				w = trix * trix * trix
			}

			if len(E) > j {
				w *= E[j]
			}
			sum += w
		}
		W[j] = w
	}

	if sum <= 0 {
		return errTotal
	}

	// normalize weights
	for i := left; i <= right; i++ {
		j := int(i)
		W[j] /= sum
	}

	return nil
}

// tricbue is the tricube function.
func tricube(u float64) float64 {
	if u >= 1 {
		return 0
	}
	tmp := (1 - u*u*u)
	return tmp * tmp * tmp
}

// neightborhoodWeight calculates the neightborhood weight
//
// this represents the mathematical version of the code. It's optimized form is used in state.localWeights.
// additionally localweights also performs various numerical stabilization functions so things don't get out of hand
func neighborhoodWeight(lambda, xi, x float64) float64 {
	return tricube(math.Abs(xi-x) / lambda)
}

// Linear performs linear regression, but constrained to left and right values. The implementation is slightly different
// from the typical Y = (X'X)^-1(X'Y)
func Linear(s *State, x, left, right float64) error {
	// Notes for implementors in other variations of regression -
	// you can access the Weights and Xs of the state by calling the function.
	//
	// Additionally, this code is quite unrolled and quite difficult to understand
	// This was done for performance purposes.
	// I will have tried to explain the equivalence in the comments
	W := s.w // s.W()
	X := s.x // s.X()

	var mean, variance float64
	for i := left; i <= right; i++ {
		j := int(i)
		mean += i * W[j]
	}

	for i := left; i <= right; i++ {
		j := int(i)
		delta := i - mean
		variance += W[j] * delta * delta
	}

	// weight updates are done only if there is enough variance (which we defined to be (0.01 * r)).
	// if the variance is >= threshold^2, then we calculate slow, and then update the weights. Otherwise,
	// we'll leave it alone.
	r := float64(len(X) - 1)
	thresh := (0.01 * r)
	if variance >= thresh*thresh {
		beta := (x - mean) / variance
		for i := left; i <= right; i++ {
			j := int(i)
			W[j] *= 1.0 + beta*(i-mean)
		}
	}
	return nil
}

// Quadratic performs quadratic regression, constrained to left. and right.
func Quadratic(s *State, x, left, right float64) error {
	// Notes for implementors of other variations of the regression -
	// you can access weights and Xs of the state by calling the function.
	//
	// Additionally, this code is quite unrolled and quite difficult to understand
	// This was done for performance purposes.
	// I will have tried to explain the equivalence in the comments
	W := s.w                   // s.W()
	X := s.x                   // s.X()
	var x1, x2, x3, x4 float64 // unrolled calculation of means m2, m3, m4

	for i := left; i <= right; i++ {
		j := int(i)
		w := W[j]
		x1w := i * w
		x2w := i * x1w
		x3w := i * x2w
		x4w := i * x3w

		x1 += x1w
		x2 += x2w
		x3 += x3w
		x4 += x4w
	}

	m2 := x2 - x1*x1
	m3 := x3 - x2*x1
	m4 := x4 - x2*x2

	variance := m2*m4 - m3*m3

	// like in the linear function, we restrict gradient updates
	// to points that are spread out enough.

	r := float64(len(X) - 1)
	thresh := (0.01 * r)
	for variance > thresh*thresh {
		beta4 := m2 / variance
		beta3 := m3 / variance
		beta2 := m4 / variance

		x1 := x - x1
		x2 := x*x - x2

		a1 := beta2*x1 - beta3*x2
		a2 := beta4*x2 - beta3*x1

		for i := left; i <= right; i++ {
			j := int(i)
			W[j] *= (1 + a1*(i-x1) + a2*(i*i-x2))
		}
	}
	return nil
}
