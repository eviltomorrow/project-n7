package handler

import (
	"errors"
	"fmt"

	"github.com/eviltomorrow/project-n7/app/n7-finder/handler/calculate"
	pb "github.com/eviltomorrow/project-n7/lib/grpc/pb/n7-repository"
	"github.com/eviltomorrow/project-n7/lib/mathutil"
	jsoniter "github.com/json-iterator/go"
)

var (
	ErrNoData = errors.New("no data")
)

type Data struct {
	Quote []*pb.Quote `json:"-"`

	Ma10, Ma50, Ma150, Ma200 []float64
}

func (d *Data) String() string {
	buf, err := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(d)
	if err != nil {
		return fmt.Sprintf("marshal failure, nest error: %v", err)
	}
	return string(buf)
}

func NewData(quotes []*pb.Quote) (*Data, error) {
	if len(quotes) == 0 {
		return nil, ErrNoData
	}

	var (
		closed                   = make([]float64, 0, len(quotes))
		ma10, ma50, ma150, ma200 []float64
	)

	for _, quote := range quotes {
		closed = append(closed, quote.Close)

		ma10 = maN(ma10, len(quotes)-10+1, 10, closed)
		ma50 = maN(ma50, len(quotes)-50+1, 50, closed)
		ma150 = maN(ma150, len(quotes)-150+1, 150, closed)
		ma200 = maN(ma200, len(quotes)-200+1, 200, closed)
	}

	return &Data{Quote: quotes, Ma10: ma10, Ma50: ma50, Ma150: ma150, Ma200: ma200}, nil
}

func maN(data []float64, size, n int, closed []float64) []float64 {
	if len(closed) < n {
		return data
	}
	if data == nil {
		data = make([]float64, 0, size)
	}
	data = append(data, mathutil.Trunc2(calculate.MA(closed[len(closed)-n:])))
	return data
}
