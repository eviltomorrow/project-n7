package handler

import pb "github.com/eviltomorrow/project-n7/lib/grpc/pb/n7-repository"

type Direction int

const (
	UP Direction = iota
	DOWN
	HORIZONTAL
)

func Findtrend(data []*pb.Quote) Direction {
	return HORIZONTAL
}