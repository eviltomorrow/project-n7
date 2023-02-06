package client

import (
	"context"
	"fmt"
	"time"

	pb "github.com/eviltomorrow/project-n7/lib/grpc/pb/n7-telegram-bot"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	TelegrambotTarget = "127.0.0.1:5274"
)

func NewTelegrambot() (pb.TelegramBotClient, func() error, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		TelegrambotTarget,
		grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"LoadBalancingPolicy": "%s"}`, roundrobin.Name)),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, nil, err
	}
	return pb.NewTelegramBotClient(conn), func() error { return conn.Close() }, nil
}
