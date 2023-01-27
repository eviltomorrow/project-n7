package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/eviltomorrow/project-n7/lib/etcd"
	grpcclient "github.com/eviltomorrow/project-n7/lib/grpc/client"
	"github.com/eviltomorrow/project-n7/lib/grpc/lb"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc/resolver"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var SyncCommand = &cli.Command{
	Name:  "sync",
	Usage: "sync data quick",
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "source", Value: "sina", Usage: "sync data from sina", Aliases: []string{"s"}},
	},
	Action: func(ctx *cli.Context) error {
		var begin = time.Now()

		if err := loadConfig(); err != nil {
			return err
		}
		etcd.Endpoints = cfg.Etcd.Endpoints

		client, err := etcd.NewClient()
		if err != nil {
			return err
		}
		defer client.Close()

		resolver.Register(lb.NewBuilder(client))

		stub, closeFunc, err := grpcclient.NewCollector()
		if err != nil {
			return err
		}
		defer closeFunc()

		resp, err := stub.Sync(context.Background(), &wrapperspb.StringValue{Value: ctx.String("source")})
		if err != nil {
			return err
		}
		fmt.Printf("[Sync Info] Total: %d, Ignore: %d, Cost: %v\r\n", resp.Total, resp.Ignore, time.Since(begin))

		return nil
	},
}
