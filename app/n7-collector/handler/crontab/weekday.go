package crontab

import (
	"context"
	"fmt"
	"time"

	"github.com/eviltomorrow/project-n7/app/n7-collector/handler/sync"
	"google.golang.org/protobuf/types/known/wrapperspb"

	grpcclient "github.com/eviltomorrow/project-n7/lib/grpc/client"
	emailpb "github.com/eviltomorrow/project-n7/lib/grpc/pb/n7-email"
	repositorypb "github.com/eviltomorrow/project-n7/lib/grpc/pb/n7-repository"
	"github.com/eviltomorrow/project-n7/lib/zlog"
	"go.uber.org/zap"
)

var (
	Source        = "sina"
	lastDay int64 = -1
)

func EveryWeekDay() error {
	zlog.Info("Sync data slow begin")
	var (
		total, ignore int64
		e             error
		begin         = time.Now()
	)
	defer func() {
		if e != nil {
			client, closeFunc, err := grpcclient.NewEmail()
			if err != nil {
				zlog.Error("Create email client failure", zap.Error(err))
				return
			}
			defer closeFunc()

			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()

			if _, err := client.Send(ctx, &emailpb.Mail{
				To: []*emailpb.Contact{
					{Name: "Shepard", Address: "eviltomorrow@163.com"},
				},
				Subject: fmt.Sprintf("同步数据失败-[%s]", time.Now().Format("2006-01-02")),
				Body:    fmt.Sprintf("错误描述, nest error: %v", e),
			}); err != nil {
				zlog.Error("Send email failure, notify [sync data slow]", zap.Error(err))
				return
			}
		}
	}()

	total, ignore, e = sync.DataSlow(Source)
	if e != nil {
		return e
	}

	var (
		client    repositorypb.RepositoryClient
		closeFunc func() error
	)
	client, closeFunc, e = grpcclient.NewRepository()
	if e != nil {
		return e
	}
	defer closeFunc()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if _, e = client.Sync(ctx, &wrapperspb.StringValue{Value: begin.Format("2006-01-02")}); e != nil {
		return e
	}
	if lastDay != -1 {
		if lastDay > total && (lastDay-total) > int64(float64(lastDay)*0.1) {
			e = fmt.Errorf("sync data slow possible missing data, nest last: %v, nest count: %v", lastDay, total)
		}
	}
	lastDay = total
	zlog.Info("Sync data slow complete", zap.Int64("total", total), zap.Int64("ignore", ignore), zap.Duration("cost", time.Since(begin)))
	return nil
}
