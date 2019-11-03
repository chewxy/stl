# stl [![GoDoc](https://godoc.org/github.com/chewxy/stl?status.svg)](https://godoc.org/github.com/chewxy/stl) [![Build Status](https://travis-ci.org/chewxy/stl.svg?branch=master)](https://travis-ci.org/chewxy/stl) [![Coverage Status](https://coveralls.io/repos/github/chewxy/stl/badge.png)](https://coveralls.io/github/chewxy/stl) [![Go Report Card](https://goreportcard.com/badge/github.com/chewxy/stl)](https://goreportcard.com/report/github.com/chewxy/stl) #

`package stl` implements Seasonal-Trend Decompositions by LOESS of time series. It is a direct implementation of [Cleveland et al (1990)](https://search.proquest.com/openview/cc5001e8a0978a6c029ae9a41af00f21/1?pq-origsite=gscholar&cbl=105444), which was written in Fortran '77. 

This package does not implement the KD-Tree indexing for nearest neighbour search for the LOESS component as implemented in the [original Netlib library](http://www.netlib.org/a/stl). Instead, a straightforwards algorithm is implemented, focusing on clarity and understandability.

There were some parts that were "inlined" and unrolled manually for performance purposes - the tricube function and local neighbour functions for example. The original more mathematical implemetnations have been left in there for future reference. 

Additionally, because multiplicative models are common in time series decompositions, [Box-Cox transform](https://en.wikipedia.org/wiki/Power_transform#Box%E2%80%93Cox_transformation) functions (and generator) have also been provided.

# Installation #

This package has very minimal dependencies. These are the listing of the dependencies:

* [`gorgonia.org/tensor`](https://github.com/gorgonia/tensor) - used for storing the matrix data in the subcycle smoothing bits of the code
* [`github.com/pkg/error`](https://github.com/pkg/error) - general errors management package
* [`github.com/chewxy/tightywhities`](https://github.com/chewxy/tightywhities) - used in plotting ASCII charts in the examples/test.
* [`gorgonia.org/dawson`](https://github.com/gorgonia/dawson) - used in tests to compare floating point numbers.

This package supports modules

Automated testing only tests up from Go 1.8 onwards. While I'm fairly confident that this package will work on those lower versions as well

# Usage #

```golang
import (
	"encoding/csv"
	"os"
	"fmt"

	"github.com/chewxy/stl"
)

func main() {
	var data []float64
	f, _ := os.Open("testdata/co2.csv")
	r := csv.NewReader(f)

	r.Read() // read header
	for rec, err := r.Read(); err == nil; rec, err = r.Read() {
		// here we're ignoring errors because we know the file to be correct
		if co2, err := strconv.ParseFloat(rec[0], 64); err == nil {
			data = append(data, co2)
		}
	}

	// The main function: 
	res := stl.Decompose(data, 12, 35, stl.Additive(), stl.WithRobustIter(2), stl.WithIter(2))
	fmt.Printf("%v", res.Seasonal)
	fmt.Printf("%v", res.Trens)
	fmt.Printf("%v", res.Resid)

```

# Licence #
 
This package is licenced with a MIT licence. I thank Rob Hyndman for writing a very excellent guide to STL, both in the R standard lib and in principle.
