package stl_test

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/chewxy/stl"
	"github.com/chewxy/tightywhities"
)

const (
	w, h = 80, 10 // width and height of the chart producd
)

func readCheck(a io.Reader, err error) io.Reader {
	if err != nil {
		panic(err)
	}
	return a
}

func check(r stl.Result) stl.Result {
	if r.Err != nil {
		panic(r.Err)
	}
	return r
}

func printResult(res stl.Result) {
	fmt.Println("Data:")
	line := tightywhities.NewLine(nil, res.Data[:w])
	line.Plot(w, h, os.Stdout)
	fmt.Println("\nTrend:")
	line = tightywhities.NewLine(nil, res.Trend[:w])
	line.Plot(w, h, os.Stdout)
	fmt.Println("\nSeasonal:")
	line = tightywhities.NewLine(nil, res.Seasonal[:w])
	line.Plot(w, h, os.Stdout)
	fmt.Println("\nResiduals:")
	line = tightywhities.NewLine(nil, res.Resid[:w])
	line.Plot(w, h, os.Stdout)

}

func Example() {
	f := readCheck(os.Open("testdata/co2.csv"))
	r := csv.NewReader(f)
	var data []float64
	r.Read() // read header
	for rec, err := r.Read(); err == nil; rec, err = r.Read() {
		// here we're ignoring errors because we know the file to be correct
		if co2, err := strconv.ParseFloat(rec[0], 64); err == nil {
			data = append(data, co2)
		}
	}
	res := check(stl.Decompose(data, 12, 35, stl.Additive(), stl.WithRobustIter(2), stl.WithIter(2)))
	fmt.Println("ADDITIVE MODEL\n=====================")
	printResult(res)

	// Multiplicative model
	res = check(stl.Decompose(data, 12, 35, stl.Multiplicative(), stl.WithRobustIter(2), stl.WithIter(2)))
	fmt.Println("\nMULTIPLICATIVE MODEL\n=====================")
	printResult(res)

	// Output:
	// ADDITIVE MODEL
	// =====================
	// Data:
	// │
	// │                                                               ╭╮          ╭╮
	// │                                                              ╭╯╰╮        ╭╯╰╮
	// │+                                      ╭╮         ╭──╮        │  │       ╭╯  ╰╮
	// │                           ╭─╮        ╭╯╰╮       ╭╯  ╰╮      ╭╯  ╰╮    ╭─╯    │
	// │                          ╭╯ │       ╭╯  ╰╮     ╭╯    │    ╭─╯    │    │      ╰╮
	// │  ╭─╮         ╭──╮       ╭╯  ╰╮     ╭╯    │    ╭╯     │   ╭╯      ╰╮  ╭╯       │
	// │  │ ╰╮       ╭╯  ╰╮     ╭╯    │    ╭╯     ╰╮  ╭╯      ╰╮ ╭╯        │ ╭╯        ╰─
	// │  ╯  ╰╮    ╭─╯    │   ╭─╯     ╰╮  ╭╯       │ ╭╯        ╰╮│         ╰─╯
	// │      ╰╮  ╭╯      ╰╮ ╭╯        │ ╭╯        ╰─╯          ╰╯
	// │       │  │        ╰╮│         ╰─╯
	// │       ╰──╯         ╰╯
	//
	// Trend:
	// │
	// │+                                   ╭───╮       ╭──╮
	// │                         ╭───╮    ╭─╯   ╰─╮  ╭──╯  ╰─╮
	// │                       ╭─╯   ╰─╮╭─╯       ╰──╯       ╰╮
	// │                      ╭╯       ╰╯                     ╰╮ ╭─────╮
	// │                     ╭╯                                ╰─╯     ╰╮
	// │                ╭╮  ╭╯                                          ╰─╮
	// │             ╭──╯╰──╯                                             ╰╮   ╭───╮
	// │           ╭─╯                                                     ╰───╯   ╰─╮
	// │         ╭─╯                                                                 ╰╮
	// │  ───────╯                                                                    ╰╮
	// │                                                                               ╰─
	//
	// Seasonal:
	// │
	// │                                                                              ╭──
	// │+                                                                        ╭────╯
	// │                                                                     ╭───╯
	// │                                                                 ╭───╯
	// │                                                             ╭───╯
	// │                                                          ╭──╯
	// │                                                      ╭───╯
	// │                                                ╭─────╯
	// │  ─────╮                                   ╭────╯
	// │       ╰──────╮                    ╭───────╯
	// │              ╰────────────────────╯
	//
	// Residuals:
	// │
	// │                           ╭╮                                  ╭╮
	// │                           │╰╮         ╭╮          ╭╮          ││          ╭╮
	// │               ╭─╮        ╭╯ │        ╭╯╰╮        ╭╯╰╮        ╭╯╰╮        ╭╯╰╮
	// │+ ╭──╮        ╭╯ │        │  ╰╮       │  │        │  ╰╮       │  │        │  │
	// │  │  │       ╭╯  ╰╮       │   │      ╭╯  ╰╮      ╭╯   │      ╭╯  ╰╮      ╭╯  ╰╮
	// │  │  ╰╮     ╭╯    │     ╭─╯   │     ╭╯    │     ╭╯    │     ╭╯    │    ╭─╯    │
	// │  ╯   │    ╭╯     │    ╭╯     │    ╭╯     │    ╭╯     │    ╭╯     │    │      │
	// │      ╰╮  ╭╯      ╰╮  ╭╯      ╰╮  ╭╯      ╰╮  ╭╯      ╰╮  ╭╯      ╰╮  ╭╯      ╰╮
	// │       │  │        │ ╭╯        │  │        │ ╭╯        │ ╭╯        │  │        │
	// │       ╰╮╭╯        ╰╮│         │ ╭╯        │╭╯         ╰╮│         │ ╭╯        │
	// │        ╰╯          ╰╯         ╰─╯         ╰╯           ╰╯         ╰─╯         ╰─
	//
	// MULTIPLICATIVE MODEL
	// =====================
	// Data:
	// │
	// │                                                               ╭╮          ╭╮
	// │                                                              ╭╯╰╮        ╭╯╰╮
	// │+                                      ╭╮         ╭──╮        │  │       ╭╯  ╰╮
	// │                           ╭─╮        ╭╯╰╮       ╭╯  ╰╮      ╭╯  ╰╮    ╭─╯    │
	// │                          ╭╯ │       ╭╯  ╰╮     ╭╯    │    ╭─╯    │    │      ╰╮
	// │  ╭─╮         ╭──╮       ╭╯  ╰╮     ╭╯    │    ╭╯     │   ╭╯      ╰╮  ╭╯       │
	// │  │ ╰╮       ╭╯  ╰╮     ╭╯    │    ╭╯     ╰╮  ╭╯      ╰╮ ╭╯        │ ╭╯        ╰─
	// │  ╯  ╰╮    ╭─╯    │   ╭─╯     ╰╮  ╭╯       │ ╭╯        ╰╮│         ╰─╯
	// │      ╰╮  ╭╯      ╰╮ ╭╯        │ ╭╯        ╰─╯          ╰╯
	// │       │  │        ╰╮│         ╰─╯
	// │       ╰──╯         ╰╯
	//
	// Trend:
	// │+
	// │                                                 ╭╮
	// │                         ╭─╮        ╭────╮     ╭─╯╰─╮
	// │                        ╭╯ ╰──╮   ╭─╯    ╰─╮ ╭─╯    ╰╮
	// │                       ╭╯     ╰╮╭─╯        ╰─╯       ╰╮      ╭╮
	// │                      ╭╯       ╰╯                     ╰──────╯╰╮
	// │                     ╭╯                                        ╰─╮
	// │             ╭─────╮╭╯                                           ╰╮      ╭╮
	// │            ╭╯     ╰╯                                             ╰╮ ╭───╯╰─╮
	// │           ╭╯                                                      ╰─╯      ╰╮
	// │        ╭──╯                                                                 ╰╮
	// │  ──────╯                                                                     ╰──
	//
	// Seasonal:
	// │
	// │                                                                           ╭─────
	// │+                                                                      ╭───╯
	// │                                                                   ╭───╯
	// │                                                               ╭───╯
	// │                                                           ╭───╯
	// │                                                        ╭──╯
	// │                                                   ╭────╯
	// │                                             ╭─────╯
	// │  ────────╮                             ╭────╯
	// │          ╰──────────╮        ╭─────────╯
	// │                     ╰────────╯
	//
	// Residuals:
	// │
	// │                           ╭╮                                  ╭╮
	// │                           │╰╮         ╭╮          ╭╮          ││          ╭╮
	// │               ╭─╮        ╭╯ │         │╰╮        ╭╯╰╮        ╭╯╰╮        ╭╯╰╮
	// │+ ╭──╮        ╭╯ │        │  ╰╮       ╭╯ │        │  │        │  │        │  │
	// │  │  │       ╭╯  │        │   │      ╭╯  ╰╮      ╭╯  ╰╮      ╭╯  ╰╮      ╭╯  ╰╮
	// │  │  ╰╮     ╭╯   ╰╮     ╭─╯   │     ╭╯    │     ╭╯    │     ╭╯    │    ╭─╯    │
	// │  ╯   │    ╭╯     │    ╭╯     │    ╭╯     │    ╭╯     │    ╭╯     │    │      │
	// │      ╰╮  ╭╯      ╰╮  ╭╯      ╰╮  ╭╯      ╰╮  ╭╯      ╰╮  ╭╯      ╰╮  ╭╯      ╰╮
	// │       │  │        │ ╭╯        │  │        │ ╭╯        │  │        │  │        │
	// │       │ ╭╯        ╰╮│         │ ╭╯        │╭╯         ╰╮╭╯        │ ╭╯        │
	// │       ╰─╯          ╰╯         ╰─╯         ╰╯           ╰╯         ╰─╯         ╰─

}
