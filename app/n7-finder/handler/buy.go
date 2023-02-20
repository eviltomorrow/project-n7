package handler

import pb "github.com/eviltomorrow/project-n7/lib/grpc/pb/n7-repository"

type Buy interface {
	Locate(data []*pb.Quote) (string, int, bool)
}
