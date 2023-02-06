package client

import (
	"context"
	"testing"

	pb "github.com/eviltomorrow/project-n7/lib/grpc/pb/n7-telegram-bot"
	"github.com/stretchr/testify/assert"
)

func TestSendBot(t *testing.T) {
	_assert := assert.New(t)

	stub, closeFunc, err := NewTelegrambot()
	_assert.Nil(err)
	defer closeFunc()

	resp, err := stub.Send(context.Background(), &pb.Chat{
		Username: "eviltomorrow",
		Text:     "Hello world",
	})
	_assert.Nil(err)
	t.Logf("id: %s\r\n", resp.Value)
}
