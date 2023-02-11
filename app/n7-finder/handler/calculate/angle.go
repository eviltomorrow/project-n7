package calculate

import "math"

func IncludedAngle(y0, y1 float64) float64 {
	return math.Atan2((y1-y0), (1-0)) * 180 / math.Pi
}

func IncludedAngleN(y0, y1 float64, n float64) float64 {
	return math.Atan2((y1-y0), n) * 180 / math.Pi
}
