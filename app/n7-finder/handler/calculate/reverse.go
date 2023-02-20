package calculate

import pb "github.com/eviltomorrow/project-n7/lib/grpc/pb/n7-repository"

func ReverseQuote(data []*pb.Quote) []*pb.Quote {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
	return data
}
