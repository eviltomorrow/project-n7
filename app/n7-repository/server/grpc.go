package server

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/eviltomorrow/project-n7/app/n7-repository/handler/data"
	"github.com/eviltomorrow/project-n7/app/n7-repository/handler/db"
	"github.com/eviltomorrow/project-n7/lib/etcd"
	"github.com/eviltomorrow/project-n7/lib/grpc/client"
	"github.com/eviltomorrow/project-n7/lib/grpc/middleware"
	pb "github.com/eviltomorrow/project-n7/lib/grpc/pb/n7-repository"
	"github.com/eviltomorrow/project-n7/lib/model"
	"github.com/eviltomorrow/project-n7/lib/mysql"
	"github.com/eviltomorrow/project-n7/lib/netutil"
	"github.com/eviltomorrow/project-n7/lib/zlog"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
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

	pb.UnimplementedRepositoryServer
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
// GetStockFull(*emptypb.Empty, Repository_GetStockFullServer) error
// GetQuoteLatest(*QuoteRequest, Repository_GetQuoteLatestServer) error
func (g *GRPC) Sync(ctx context.Context, req *wrapperspb.StringValue) (*pb.SyncInfo, error) {
	if req == nil {
		return nil, fmt.Errorf("invalid request, date is nil")
	}
	d, err := time.Parse("2006-01-02", req.Value)
	if err != nil {
		return nil, err
	}

	stub, closeFunc, err := client.NewCollector()
	if err != nil {
		return nil, err
	}
	defer closeFunc()

	resp, err := stub.GetMetadata(context.Background(), &wrapperspb.StringValue{Value: req.Value})
	if err != nil {
		return nil, err
	}

	var (
		pipe   = make(chan *model.Metadata, 128)
		signal = make(chan struct{}, 1)
	)

	go func() {
		defer func() {
			close(pipe)
		}()
		for {
			md, err := resp.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				zlog.Error("GetMetadata recv failure", zap.Error(err))
				return
			}
			select {
			case <-signal:
				return
			default:
				pipe <- &model.Metadata{
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
				}
			}
		}
	}()
	affectedS, affectedD, affectedW, err := data.TransmissionMetadata(d, pipe)
	if err != nil {
		signal <- struct{}{}
		return nil, err
	}
	return &pb.SyncInfo{AffectedStock: affectedS, AffectedQuoteDay: affectedD, AffectedQuoteWeek: affectedW}, nil
}

func (g *GRPC) GetStockFull(_ *emptypb.Empty, resp pb.Repository_GetStockFullServer) error {
	var (
		offset, limit int64 = 0, 100
		timeout             = 10 * time.Second
	)

	for {
		stocks, err := db.StockWithSelectRange(mysql.DB, offset, limit, timeout)
		if err != nil {
			return err
		}

		for _, stock := range stocks {
			if err := resp.Send(&pb.Stock{Code: stock.Code, Name: stock.Name, Suspend: stock.Suspend}); err != nil {
				return err
			}
		}

		if int64(len(stocks)) < limit {
			break
		}
		offset += limit
	}
	return nil
}

func (g *GRPC) GetQuoteLatest(req *pb.QuoteRequest, resp pb.Repository_GetQuoteLatestServer) error {
	var (
		limit   int64 = req.Limit
		mode    string
		timeout = 10 * time.Second
	)
	if limit > 250 {
		return fmt.Errorf("limit should be less than 250")
	}

	switch req.Mode {
	case pb.QuoteRequest_Day:
		mode = db.Day
	case pb.QuoteRequest_Week:
		mode = db.Week
	default:
		mode = db.Day
	}

	quotes, err := db.QuoteWithSelectManyLatest(mysql.DB, mode, req.Code, req.Date, limit, timeout)
	if err != nil {
		return err
	}

	for _, quote := range quotes {
		if err := resp.Send(&pb.Quote{
			Code:            quote.Code,
			Open:            quote.Open,
			Close:           quote.Close,
			High:            quote.High,
			Low:             quote.Low,
			YesterdayClosed: quote.YesterdayClosed,
			Volume:          quote.Volume,
			Account:         quote.Account,
			Date:            quote.Date.Format("2006-01-02"),
			NumOfYear:       int32(quote.NumOfYear),
		}); err != nil {
			return err
		}
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
	pb.RegisterRepositoryServer(g.server, g)
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
