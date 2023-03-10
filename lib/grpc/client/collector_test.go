package client

import (
	"context"
	"io"
	"testing"

	"github.com/eviltomorrow/project-n7/lib/etcd"
	"github.com/eviltomorrow/project-n7/lib/grpc/lb"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/resolver"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestGetMetadata(t *testing.T) {
	_assert := assert.New(t)

	etcd.Endpoints = []string{"127.0.0.1:2379"}
	client, err := etcd.NewClient()
	_assert.Nil(err)
	defer client.Close()

	resolver.Register(lb.NewBuilder(client))

	stub, closeFunc, err := NewCollector()
	_assert.Nil(err)
	defer closeFunc()

	resp, err := stub.GetMetadata(context.Background(), &wrapperspb.StringValue{Value: "2023-01-20"})
	_assert.Nil(err)

	var i int
	for {
		md, err := resp.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			_assert.Nil(err)
			break
		}
		i++
		t.Logf("[%4d]%s\r\n", i, md.String())
	}
}
