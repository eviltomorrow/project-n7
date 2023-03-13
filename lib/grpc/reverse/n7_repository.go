package reverse

import pb "github.com/eviltomorrow/project-n7/lib/grpc/pb/n7-repository"

func Quote(quotes []*pb.Quote) []*pb.Quote {
	for i, j := 0, len(quotes)-1; i < j; i, j = i+1, j-1 {
		quotes[i], quotes[j] = quotes[j], quotes[i]
	}
	return quotes
}
