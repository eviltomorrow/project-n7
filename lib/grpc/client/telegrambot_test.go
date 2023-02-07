package client

import (
	"context"
	"testing"

	pb "github.com/eviltomorrow/project-n7/lib/grpc/pb/n7-telegram-bot"
)

func TestSendBot(t *testing.T) {
	TelegrambotTarget = "206.190.237.98:5274"
	stub, closeFunc, err := NewTelegrambot()
	if err != nil {
		t.Fatal(err)
	}
	defer closeFunc()

	resp, err := stub.Send(context.Background(), &pb.Chat{
		Username: "eviltomorrow",
		Text:     "Hello world",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("id: %s\r\n", resp.Value)
}
