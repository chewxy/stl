package stl

import (
	"testing"
	"testing/quick"
)

func TestDefaults(t *testing.T) {
	seasonal := func(width int) (ok bool) {
		defer func() {
			if r := recover(); r != nil {
				if width <= 0 {
					ok = true
					return
				}
				ok = false
			}
		}()
		conf := DefaultSeasonal(width)
		if conf.Width <= 0 {
			return false
		}
		if conf.Fn == nil {
			return false
		}
		if conf.Jump <= 0 {
			return false
		}
		return true
	}
	if err := quick.Check(seasonal, nil); err != nil {
		t.Fatal(err)
	}

	trend := func(periodicity, width int) (ok bool) {
		defer func() {
			if r := recover(); r != nil {
				if width <= 0 || periodicity <= 0 || (r == "Default trend width calculation overflowed" && periodicity > 0 && width > 0) {
					ok = true
					return
				}
				ok = false
			}
		}()
		conf := DefaultTrend(periodicity, width)
		if conf.Width <= 0 {
			return false
		}
		if conf.Fn == nil {
			return false
		}
		if conf.Jump <= 0 {
			return false
		}
		return true
	}
	if err := quick.Check(trend, nil); err != nil {
		t.Fatal(err)
	}

	lowpass := func(periodicity int) (ok bool) {
		defer func() {
			if r := recover(); r != nil {
				if periodicity <= 0 {
					ok = true
					return
				}
				ok = false
			}
		}()
		conf := DefaultLowPass(periodicity)
		if conf.Width <= 0 {
			return false
		}
		if conf.Fn == nil {
			return false
		}
		if conf.Jump <= 0 {
			return false
		}
		return true
	}

	if err := quick.Check(lowpass, nil); err != nil {
		t.Fatal(err)
	}
}
