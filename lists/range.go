package lists

func Range(from int, to int, step uint) []int {
	result := []int{from}
	up := from < to
	for {
		if up {
			from = from + int(step)
			if from <= to {
				result = append(result, from)
			} else {
				break
			}
		} else {
			from = from - int(step)
			if from >= to {
				result = append(result, from)
			} else {
				break
			}
		}
	}
	return result
}
