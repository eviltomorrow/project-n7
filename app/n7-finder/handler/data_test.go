package handler

import (
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/eviltomorrow/project-n7/lib/grpc/client"
	pb "github.com/eviltomorrow/project-n7/lib/grpc/pb/n7-repository"
	"github.com/eviltomorrow/project-n7/lib/grpc/reverse"
)

func TestNewData(t *testing.T) {
	client, closeFunc, err := client.NewRepositoryWithTarget("101.42.255.204:5272")
	if err != nil {
		t.Fatal(err)
	}
	defer closeFunc()

	resp, err := client.GetQuoteLatest(context.Background(), &pb.QuoteRequest{
		Code:  "sh601066",
		Date:  "2023-03-13",
		Limit: 50,
	})
	if err != nil {
		t.Fatal(err)
	}

	var datas = make([]*pb.Quote, 0, 50)
	for {
		data, err := resp.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		datas = append(datas, data)
		fmt.Println(data)
	}

	data, err := NewData(reverse.Quote(datas))
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("data: %v\r\n", data)
}
