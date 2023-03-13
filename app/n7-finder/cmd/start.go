package cmd

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/eviltomorrow/project-n7/app/n7-finder/conf"
	"github.com/eviltomorrow/project-n7/app/n7-finder/server"
	"github.com/eviltomorrow/project-n7/lib/cleanup"
	"github.com/eviltomorrow/project-n7/lib/etcd"
	"github.com/eviltomorrow/project-n7/lib/fs"
	"github.com/eviltomorrow/project-n7/lib/grpc/lb"
	"github.com/eviltomorrow/project-n7/lib/grpc/middleware"
	"github.com/eviltomorrow/project-n7/lib/helper"
	"github.com/eviltomorrow/project-n7/lib/pid"
	"github.com/eviltomorrow/project-n7/lib/procutil"
	"github.com/eviltomorrow/project-n7/lib/zlog"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"google.golang.org/grpc/resolver"
)

var workflowsFunc = []func() error{
	setRuntime,
	loadConfig,
	printCfg,
	setGlobal,
	runServer,
	buildPidFile,
	rewritePaniclog,
}

var StartCommand = &cli.Command{
	Name:  "start",
	Usage: "start app in background",
	Flags: []cli.Flag{
		&cli.BoolFlag{Name: "daemon", Value: false, Usage: "run app in background", Aliases: []string{"d"}},
	},
	Action: func(ctx *cli.Context) error {
		var begin = time.Now()
		if isDaemon := ctx.Bool("daemon"); isDaemon {
			if err := procutil.RunInBackground(helper.App.Name, []string{"start"}, nil, nil); err != nil {
				log.Fatalf("[F] Run app in background failure, nest error: %v", err)
			}
			return nil
		}

		defer func() {
			for _, err := range cleanup.RunCleanupFuncs() {
				zlog.Error("run clear funcs failure", zap.Error(err))
			}
			zlog.Info("Stop app complete", zap.String("app-name", helper.App.Name), zap.String("running-duration", helper.Runtime.RunningDuration()))
		}()

		for _, f := range workflowsFunc {
			if err := f(); err != nil {
				return err
			}
		}
		zlog.Info("Start app success", zap.String("app-name", helper.App.Name), zap.Duration("cost", time.Since(begin)))

		procutil.WaitForSigterm()
		return nil
	},
}

var cfg = conf.DefaultGlobal

func loadConfig() error {
	if err := cfg.ParseFile(filepath.Join(helper.Runtime.RootDir, "/etc/global.conf")); err != nil {
		return err
	}

	closeFuncs, err := conf.SetupLogger(cfg.Log)
	if err != nil {
		return err
	}
	cleanup.RegisterCleanupFuncs(closeFuncs...)

	return nil
}

func setGlobal() error {
	etcd.Endpoints = cfg.Etcd.Endpoints
	middleware.LogDir = filepath.Join(helper.Runtime.RootDir, "/log")

	server.ListenHost = cfg.Server.Host
	server.Port = cfg.Server.Port
	return nil
}

func setRuntime() error {
	for _, dir := range []string{
		filepath.Join(helper.Runtime.RootDir, "/log"),
		filepath.Join(helper.Runtime.RootDir, "/var/run"),
	} {
		if err := fs.CreateDir(dir); err != nil {
			return fmt.Errorf("create dir failure, nest error: %v", err)
		}
	}
	return nil
}

func runServer() error {
	client, err := etcd.NewClient()
	if err != nil {
		return err
	}
	cleanup.RegisterCleanupFuncs(client.Close)

	if err := middleware.InitLogger(); err != nil {
		return err
	}
	resolver.Register(lb.NewBuilder(client))

	var g = &server.GRPC{
		AppName: helper.App.Name,
		Client:  client,
	}
	if err := g.Startup(); err != nil {
		return err
	}
	cleanup.RegisterCleanupFuncs(g.Shutdown)

	zlog.Info("Startup GRPC Server complete", zap.String("addrs", fmt.Sprintf("%s:%d", server.ListenHost, server.Port)))
	return nil
}

func rewritePaniclog() error {
	fs.StderrFilePath = filepath.Join(helper.Runtime.RootDir, "/log/panic.log")
	if err := fs.RewriteStderrFile(); err != nil {
		zlog.Error("RewriteStderrFile failure", zap.Error(err))
	}
	return nil
}

func buildPidFile() error {
	closeFunc, err := pid.CreatePidFile(filepath.Join(helper.Runtime.RootDir, fmt.Sprintf("/var/run/%s.pid", helper.App.Name)))
	if err != nil {
		return err
	}
	cleanup.RegisterCleanupFuncs(closeFunc)
	return nil
}

func printCfg() error {
	zlog.Info("Load config success", zap.String("config", cfg.String()))
	return nil
}
