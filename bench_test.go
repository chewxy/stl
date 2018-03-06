package stl

import (
	"encoding/csv"
	"os"
	"strconv"
	"testing"
)

func BenchmarkAPI(b *testing.B) {
	var data []float64
	f, _ := (os.Open("testdata/co2.csv"))
	r := csv.NewReader(f)
	r.Read() // read header
	for rec, err := r.Read(); err == nil; rec, err = r.Read() {
		// here we're ignoring errors because we know the file to be correct
		if co2, err := strconv.ParseFloat(rec[0], 64); err == nil {
			data = append(data, co2)
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res := Decompose(data, 12, 35, Additive(), WithRobustIter(2), WithIter(2))
		if res.Err != nil {
			b.Fatal(res.Err)
		}
	}
}
