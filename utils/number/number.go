package numberutil

import "strconv"

func RangeInt(from int, to int, step uint) []int {
	if step == 0 {
		step = 1
	}
	var numbers []int
	var i int
	var intStep = int(step)
	if from < to {
		for i = from; i <= to; i += intStep {
			numbers = append(numbers, i)
		}
	} else {
		for i = from; i >= to; i -= intStep {
			numbers = append(numbers, i)
		}
	}
	return numbers
}

func ParseInt64(s string, def int64) int64 {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return def
	}
	return i
}

func Times(n uint, iterator func(i uint)) {
	var i uint
	for i = 0; i < n; i ++ {
		iterator(i)
	}
}

func MaxInt64(n ... int64) int64 {
	count := len(n)
	if count == 0 {
		return 0
	}

	if count == 1 {
		return n[0]
	}

	first := n[0]
	for _, m := range n[1:] {
		if m > first {
			first = m
		}
	}
	return first
}

func MinInt64(n ... int64) int64 {
	count := len(n)
	if count == 0 {
		return 0
	}

	if count == 1 {
		return n[0]
	}

	first := n[0]
	for _, m := range n[1:] {
		if m < first {
			first = m
		}
	}
	return first
}

func MaxInt32(n ... int32) int32 {
	count := len(n)
	if count == 0 {
		return 0
	}

	if count == 1 {
		return n[0]
	}

	first := n[0]
	for _, m := range n[1:] {
		if m > first {
			first = m
		}
	}
	return first
}

func MinInt32(n ... int32) int32 {
	count := len(n)
	if count == 0 {
		return 0
	}

	if count == 1 {
		return n[0]
	}

	first := n[0]
	for _, m := range n[1:] {
		if m < first {
			first = m
		}
	}
	return first
}

func MaxInt(n ... int) int {
	count := len(n)
	if count == 0 {
		return 0
	}

	if count == 1 {
		return n[0]
	}

	first := n[0]
	for _, m := range n[1:] {
		if m > first {
			first = m
		}
	}
	return first
}

func MinInt(n ... int) int {
	count := len(n)
	if count == 0 {
		return 0
	}

	if count == 1 {
		return n[0]
	}

	first := n[0]
	for _, m := range n[1:] {
		if m < first {
			first = m
		}
	}
	return first
}
