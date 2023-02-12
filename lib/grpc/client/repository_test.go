package client

import (
	"context"
	"io"
	"testing"

	"github.com/eviltomorrow/project-n7/lib/etcd"
	"github.com/eviltomorrow/project-n7/lib/grpc/lb"
	pb "github.com/eviltomorrow/project-n7/lib/grpc/pb/n7-repository"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/resolver"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestGetStock(t *testing.T) {
	_assert := assert.New(t)

	etcd.Endpoints = []string{"127.0.0.1:2379"}
	client, err := etcd.NewClient()
	_assert.Nil(err)
	defer client.Close()

	resolver.Register(lb.NewBuilder(client))

	stub, closeFunc, err := NewRepository()
	_assert.Nil(err)
	defer closeFunc()

	resp, err := stub.GetStockFull(context.Background(), &emptypb.Empty{})
	_assert.Nil(err)

	var i int
	for {
		stock, err := resp.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			_assert.Nil(err)
			break
		}
		i++
		t.Logf("[%4d]%s\r\n", i, stock.String())
	}
}

func TestGetQuote(t *testing.T) {
	client, closed, err := NewRepositoryWithTarget("101.42.255.204:5272")
	if err != nil {
		t.Fatal(err)
	}
	defer closed()

	resp, err := client.GetQuoteLatest(context.Background(), &pb.QuoteRequest{Code: "sh601066", Date: "2023-22-10", Limit: 200})
	if err != nil {
		t.Fatal(err)
	}
	for {
		quote, err := resp.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("Quote: %v", quote)
	}
}
