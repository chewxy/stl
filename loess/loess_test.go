package loess

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"

	"gorgonia.org/dawson"
)

type gen1 struct {
	a           []float64
	width, jump int
}

var float64sType = reflect.TypeOf([]float64{})

// Generate implements quick.Generator
func (g gen1) Generate(r *rand.Rand, size int) reflect.Value {
	elements := r.Intn(size)
	if elements < 5 {
		elements = 6
	}

	a := reflect.MakeSlice(float64sType, elements, elements).Interface().([]float64)
	width := r.Intn(len(a) / 2)

	var jump int
	for jump = r.Intn(5); jump == 0; jump = r.Intn(5) {
	}

	data := gen1{
		a:     a,
		width: width,
		jump:  jump,
	}
	return reflect.ValueOf(data)
}

func TestSmooth(t *testing.T) {

	// assert := assert.New(t)

	f := func(data gen1) bool {
		a := data.a
		width := data.width
		jump := data.jump

		a1 := make([]float64, len(a))
		a2 := make([]float64, len(a))

		copy(a1, a)
		copy(a2, a)

		b1, err := Smooth(a1, width, jump, Linear)
		if err != nil {
			return false
		}
		s := New(width, a2)

		b2 := make([]float64, len(a))
		if _, err := UnsafeSmooth(s, width, jump, Linear, b2); err != nil {
			return false
		}
		return dawson.AllClose(b1, b2)
	}

	if err := quick.Check(f, nil); err != nil {
		t.Fatalf("Error %v", err)
	}
}

func Example_smooth() {
	a := []float64{
		5, 6.0, 2.0, 4.5, 5,
		5, 6.5, 3.5, 4.0, 5,
		5, 5.5, 3.5, 5.0, 5,
		5, 6.5, 2.5, 4.5, 5,
	}

	// Smooth on a seasonality width of 1, and period of 5 (i.e. look ahead 5 and regress on it)
	if smoothed, err := Smooth(a, 1, 5, Linear); err == nil {
		fmt.Printf("Smoothed %1.2f\n", smoothed)
	}

	// Smoothed on a periodic window of of 5
	if smoothed, err := Smooth(a, 5, 1, Linear); err == nil {
		fmt.Printf("Smoothed %1.2f\n", smoothed)
	}

	// if smoothed, err := Smooth(a, 5, 1, Quadratic); err == nil {
	// 	fmt.Printf("Smoothed %1.1v\n", smoothed)
	// }

	// Smoothed [1e+01 1e+01 1e+01 9 7 5 5 5 5 5 5 3 1 -0.7 -3 -5 0.1 5 9 1e+01]

	// Output:
	// Smoothed [5.00 5.00 5.00 5.00 5.00 5.00 5.00 5.00 5.00 5.00 5.00 5.00 5.00 5.00 5.00 5.00 5.00 5.00 5.00 5.00]
	// Smoothed [5.50 4.60 3.86 3.93 4.86 5.43 5.21 4.50 4.14 4.71 5.14 4.78 4.50 4.57 5.00 5.43 4.93 4.22 4.35 4.78]

}

func Example_externalWeights() {
	a := []float64{
		5, 6.0, 2.0, 4.5, 5,
		5, 6.5, 3.5, 4.0, 5,
		5, 5.5, 3.5, 5.0, 5,
		5, 6.5, 2.5, 4.5, 5,
	}
	w := []float64{
		1, 0, 1, 0, 1,
		1, 0, 1, 0, 1,
		1, 0, 1, 0, 1,
		1, 0, 1, 0, 1,
	}

	width := 5
	state := NewWithExternal(width, a, w)
	retVal := make([]float64, len(a))

	if smoothed, err := UnsafeSmooth(state, width, 1, Linear, retVal); err == nil {
		fmt.Printf("Smoothed %1.2f\n", smoothed)
	} else {
		fmt.Printf("ERR %v", err)
	}

	// Output:
	// Smoothed [5.00 3.50 2.00 3.50 5.00 5.00 4.25 3.50 4.25 5.00 5.00 4.25 3.50 4.25 5.00 5.00 3.75 2.50 3.75 5.00]
}
