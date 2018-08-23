package numberutil

import "testing"

func TestRangeInt(t *testing.T) {
	t.Log(RangeInt(30, 10, 4))
	t.Log(RangeInt(10, 30, 4))
}

func TestTimes(t *testing.T) {
	Times(10, func(i uint) {
		t.Log(i)
	})
}

func TestMaxInt64(t *testing.T) {
	t.Log(MaxInt64())
	t.Log(MaxInt64(1))
	t.Log(MaxInt64(1, 2, 3, 4, 5))
}

func TestMinInt64(t *testing.T) {
	t.Log(MinInt64())
	t.Log(MinInt64(1))
	t.Log(MinInt64(1, 2, 3, 4, 5))
}
