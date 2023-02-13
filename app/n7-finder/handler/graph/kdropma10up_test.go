package graph

import (
	"context"
	"io"
	"testing"

	"github.com/eviltomorrow/project-n7/app/n7-finder/handler"
	"github.com/eviltomorrow/project-n7/lib/grpc/client"
	pb "github.com/eviltomorrow/project-n7/lib/grpc/pb/n7-repository"
)

func TestKDropMA10UPMatch(t *testing.T) {
	client, closed, err := client.NewRepositoryWithTarget("101.42.255.204:5272")
	if err != nil {
		t.Fatal(err)
	}
	defer closed()

	resp, err := client.GetQuoteLatest(context.Background(), &pb.QuoteRequest{Code: "sh601066", Date: "2023-22-10", Limit: 20})
	if err != nil {
		t.Fatal(err)
	}
	var data = make([]*pb.Quote, 0, 200)
	for {
		quote, err := resp.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		data = append(data, quote)
	}
	data = handler.ReverseQuote(data)

	var k = &KDropMA10UP{}
	var flag = k.Match(data)
	t.Logf("Flag: %v", flag)
}
