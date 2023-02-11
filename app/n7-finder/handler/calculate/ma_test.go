package calculate

import "testing"

func TestMA(t *testing.T) {
	var (
		data = []float64{1, 2, 3, 4, 5, 6, 7, 8}
	)
	ma := MA(data)
	t.Logf("ma: %v\r\n", ma)
}
