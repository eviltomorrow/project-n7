package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/eviltomorrow/project-n7/app/n7-collector/handler/db"
	"github.com/eviltomorrow/project-n7/app/n7-collector/handler/sync"
	"github.com/eviltomorrow/project-n7/lib/etcd"
	"github.com/eviltomorrow/project-n7/lib/grpc/middleware"
	pb "github.com/eviltomorrow/project-n7/lib/grpc/pb/n7-collector"
	"github.com/eviltomorrow/project-n7/lib/mongodb"
	"github.com/eviltomorrow/project-n7/lib/netutil"
	clientv3 "go.etcd.io/etcd/client/v3"
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
	Client  *clientv3.Client

	ctx        context.Context
	cancel     func()
	revokeFunc func() error
	server     *grpc.Server

	pb.UnimplementedCollectorServer
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

// Sync(context.Context, *wrapperspb.StringValue) (*emptypb.Empty, error)
// GetMetadata(*wrapperspb.StringValue, Collector_GetMetadataServer) error
func (g *GRPC) Sync(ctx context.Context, req *wrapperspb.StringValue) (*pb.SyncInfo, error) {
	if req == nil {
		return nil, fmt.Errorf("invalid request, source is nil")
	}
	if req.Value != "sina" && req.Value != "net126" {
		return nil, fmt.Errorf("invalid request, source is %s", req.Value)
	}
	total, ignore, err := sync.DataQuick(req.Value)
	if err != nil {
		return nil, err
	}
	return &pb.SyncInfo{Total: total, Ignore: ignore}, nil
}

func (g *GRPC) GetMetadata(req *wrapperspb.StringValue, resp pb.Collector_GetMetadataServer) error {
	if req == nil {
		return fmt.Errorf("invalid request, date is nil")
	}

	d, err := time.Parse("2006-01-02", req.Value)
	if err != nil {
		return err
	}

	var (
		offset, limit int64 = 0, 100
		lastID        string
		timeout       = 20 * time.Second
	)
	for {
		metadata, err := db.SelectMetadataRange(mongodb.DB, offset, limit, d.Format("2006-01-02"), lastID, timeout)
		if err != nil {
			return err
		}
		if len(metadata) == 0 {
			break
		}
		for _, md := range metadata {
			if err := resp.Send(&pb.Metadata{
				Source:          md.Source,
				Code:            md.Code,
				Name:            md.Name,
				Open:            md.Open,
				YesterdayClosed: md.YesterdayClosed,
				Latest:          md.Latest,
				High:            md.High,
				Low:             md.Low,
				Volume:          md.Volume,
				Account:         md.Account,
				Date:            md.Date,
				Time:            md.Time,
				Suspend:         md.Suspend,
			}); err != nil {
				return err
			}
		}
		if len(metadata) < int(limit) {
			break
		}
		offset += limit
		// if len(metadata) >= 2 {
		// 	lastID = metadata[len(metadata)-2].ObjectID
		// }
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
	g.revokeFunc, err = etcd.RegisterService(g.ctx, g.AppName, AccessHost, Port, 10, g.Client)
	if err != nil {
		return err
	}

	reflection.Register(g.server)
	pb.RegisterCollectorServer(g.server, g)
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
