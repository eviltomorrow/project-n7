package server

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/eviltomorrow/project-n7/app/n7-telegram-bot/conf"
	"github.com/eviltomorrow/project-n7/lib/grpc/middleware"
	pb "github.com/eviltomorrow/project-n7/lib/grpc/pb/n7-telegram-bot"
	"github.com/eviltomorrow/project-n7/lib/netutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	ListenHost, AccessHost string
	Port                   int
)

type GRPC struct {
	AppName string
	TB      *conf.TelegramBot

	ctx    context.Context
	cancel func()
	server *grpc.Server

	pb.UnimplementedTelegramBotServer
}

func setDefault() error {
	h, err := netutil.GetLocalIP2()
	if err != nil {
		return err
	}
	AccessHost = h
	if ListenHost == "" {
		ListenHost = h
	}

	if Port == 0 {
		p, err := netutil.GetAvailablePort()
		if err != nil {
			return err
		}
		Port = p
	}

	if ListenHost == "" || AccessHost == "" || Port == 0 {
		return fmt.Errorf("panic: invalid ListenHost/AccessHost or Port")
	}
	return nil
}

func (g *GRPC) Startup() error {
	if err := setDefault(); err != nil {
		return err
	}

	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", ListenHost, Port))
	if err != nil {
		return err
	}

	g.server = grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			middleware.UnaryServerRecoveryInterceptor,
			middleware.UnaryServerLogInterceptor,
		),
		grpc.ChainStreamInterceptor(
			middleware.StreamServerRecoveryInterceptor,
			middleware.StreamServerLogInterceptor,
		),
	)
	g.ctx, g.cancel = context.WithCancel(context.Background())

	reflection.Register(g.server)
	pb.RegisterTelegramBotServer(g.server, g)
	go func() {
		if err := g.server.Serve(listen); err != nil {
			log.Fatalf("Startup grpc server failure, nest error: %v", err)
		}
	}()
	return nil
}

func (g *GRPC) Shutdown() error {
	if g.server != nil {
		g.server.Stop()
	}
	g.cancel()
	return nil
}
