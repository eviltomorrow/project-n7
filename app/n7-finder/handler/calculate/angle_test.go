package calculate

import "testing"

func TestIncludedangle(t *testing.T) {
	var (
		y0, y1 float64 = 0, 1
	)
	var angle = IncludedAngle(y0, y1)
	t.Logf("angle: %v", angle)
}
