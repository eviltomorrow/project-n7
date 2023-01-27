package client

import (
	"context"
	"testing"

	"github.com/eviltomorrow/project-n7/lib/etcd"
	"github.com/eviltomorrow/project-n7/lib/grpc/lb"
	pb "github.com/eviltomorrow/project-n7/lib/grpc/pb/n7-email"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/resolver"
)

func TestSendEmail(t *testing.T) {
	_assert := assert.New(t)

	etcd.Endpoints = []string{"127.0.0.1:2379"}
	client, err := etcd.NewClient()
	_assert.Nil(err)
	defer client.Close()

	resolver.Register(lb.NewBuilder(client))

	stub, closeFunc, err := NewEmail()
	_assert.Nil(err)
	defer closeFunc()

	resp, err := stub.Send(context.Background(), &pb.Mail{
		To: []*pb.Contact{
			{Name: "shepard", Address: "eviltomorrow@163.com"},
		},
		Subject:     "Test",
		Body:        "This is one test",
		ContentType: pb.Mail_TEXT_HTML,
	})
	_assert.Nil(err)
	t.Logf("id: %s\r\n", resp.Value)
}
