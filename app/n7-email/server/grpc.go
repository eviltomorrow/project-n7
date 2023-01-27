package server

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/eviltomorrow/project-n7/app/n7-email/conf"
	"github.com/eviltomorrow/project-n7/lib/etcd"
	"github.com/eviltomorrow/project-n7/lib/grpc/middleware"
	pb "github.com/eviltomorrow/project-n7/lib/grpc/pb/n7-email"
	"github.com/eviltomorrow/project-n7/lib/netutil"
	"github.com/eviltomorrow/project-n7/lib/smtp"
	"github.com/eviltomorrow/project-n7/lib/zlog"
	"github.com/google/uuid"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var (
	ListenHost, AccessHost string
	Port                   int
)

type GRPC struct {
	AppName string
	SMTP    *conf.SMTP
	Client  *clientv3.Client

	ctx        context.Context
	cancel     func()
	revokeFunc func() error
	server     *grpc.Server

	pb.UnimplementedEmailServer
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

func (g *GRPC) Send(ctx context.Context, mail *pb.Mail) (*wrapperspb.StringValue, error) {
	if mail == nil {
		return nil, fmt.Errorf("illegal parameter, nest error: mail is nil")
	}
	if len(mail.To) == 0 {
		return nil, fmt.Errorf("illegal parameter, nest error: to is nil")
	}

	var contentType = smtp.TextHTML
	switch mail.ContentType {
	case pb.Mail_TEXT_PLAIN:
		contentType = smtp.TextPlain
	default:
	}
	var message = &smtp.Message{
		From: smtp.Contact{
			Name:    g.SMTP.Alias,
			Address: g.SMTP.Username,
		},
		Subject:     mail.Subject,
		Body:        mail.Body,
		ContentType: contentType,
	}

	var to = make([]smtp.Contact, 0, len(mail.To))
	for _, c := range mail.To {
		if c != nil {
			to = append(to, smtp.Contact{Name: c.Name, Address: c.Address})
		}
	}
	message.To = to

	var cc = make([]smtp.Contact, 0, len(mail.Cc))
	for _, c := range mail.Cc {
		if c != nil {
			cc = append(cc, smtp.Contact{Name: c.Name, Address: c.Address})
		}
	}
	message.Cc = cc

	var bcc = make([]smtp.Contact, 0, len(mail.Bcc))
	for _, c := range mail.Bcc {
		if c != nil {
			bcc = append(bcc, smtp.Contact{Name: c.Name, Address: c.Address})
		}
	}
	message.Bcc = bcc

	if err := smtp.SendWithSSL(g.SMTP.Server, g.SMTP.Username, g.SMTP.Password, message); err != nil {
		return nil, err
	}

	var uid = uuid.New()
	zlog.Info("Send email success", zap.String("id", uid.String()), zap.String("msg", message.String()))
	return &wrapperspb.StringValue{Value: uid.String()}, nil
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
	g.revokeFunc, err = etcd.RegisterService(g.ctx, g.AppName, AccessHost, Port, 10, g.Client)
	if err != nil {
		return err
	}

	reflection.Register(g.server)
	pb.RegisterEmailServer(g.server, g)
	go func() {
		if err := g.server.Serve(listen); err != nil {
			log.Fatalf("Startup grpc server failure, nest error: %v", err)
		}
	}()
	return nil
}

func (g *GRPC) Shutdown() error {
	if g.revokeFunc != nil {
		g.revokeFunc()
	}

	if g.server != nil {
		g.server.Stop()
	}
	g.cancel()
	return nil
}
