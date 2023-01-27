package calculate

func MA[T int64 | float64](data []T) T {
	if len(data) == 0 {
		return 0
	}

	var sum T
	for _, c := range data {
		sum += c
	}
	return sum / T(len(data))
}
