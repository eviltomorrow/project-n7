package mathutil

import (
	"math"
	"math/rand"
	"time"
)

var (
	n10_1 = math.Pow10(2)
	n10_4 = math.Pow10(4)
)

func GenRandInt(min, max int) int {
	rand.Seed(time.Now().Unix())
	randNum := rand.Intn(max-min) + min
	return randNum
}

// Trunc2 trunc2
func Trunc2(val float64) float64 {
	return math.Trunc((val+0.5/n10_1)*n10_1) / n10_1
}

func Trunc4(val float64) float64 {
	return math.Trunc((val+0.5/n10_4)*n10_4) / n10_4
}

func TruncN(f float64, n int) float64 {
	pow10_n := math.Pow10(n)
	return math.Trunc((f+0.5/pow10_n)*pow10_n) / pow10_n
}

func Max[T int | uint | int64 | uint64 | float64](data []T) T {
	if len(data) == 0 {
		return 0
	}
	var max = data[0]
	for i := 1; i <= len(data)-1; i++ {
		if data[i] > max {
			max = data[i]
		}
	}
	return max
}

func Min[T int | uint | int64 | uint64 | float64](data []T) T {
	if len(data) == 0 {
		return 0
	}
	var min = data[0]
	for i := 1; i <= len(data)-1; i++ {
		if data[i] < min {
			min = data[i]
		}
	}
	return min
}

func Sum[T int | uint | int64 | uint64 | float64](data []T) T {
	var sum T
	for _, d := range data {
		sum += d
	}
	return sum
}
