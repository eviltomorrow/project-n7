package graph

import (
	"fmt"

	"github.com/eviltomorrow/project-n7/app/n7-finder/handler/calculate"
	pb "github.com/eviltomorrow/project-n7/lib/grpc/pb/n7-repository"
	"github.com/eviltomorrow/project-n7/lib/mathutil"
)

type KDropMA10UP struct {
}

func (k *KDropMA10UP) Match(data []*pb.Quote) bool {
	if len(data) <= 14 {
		return false
	}
	var (
		closed = make([]float64, 0, 10)
		ma10   = make([]float64, 0, len(data)-10+1)
	)
	for _, d := range data {
		closed = append(closed, d.Close)
		if len(closed) >= 10 {
			ma10 = append(ma10, mathutil.Trunc2(calculate.MA(closed)))
			closed = closed[1:]
		}
	}
	fmt.Println(ma10)

	return true
}
