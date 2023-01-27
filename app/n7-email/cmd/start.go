package cmd

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/eviltomorrow/project-n7/app/n7-email/conf"
	"github.com/eviltomorrow/project-n7/app/n7-email/server"
	"github.com/eviltomorrow/project-n7/lib/etcd"
	"github.com/eviltomorrow/project-n7/lib/fs"
	"github.com/eviltomorrow/project-n7/lib/grpc/lb"
	"github.com/eviltomorrow/project-n7/lib/grpc/middleware"
	"github.com/eviltomorrow/project-n7/lib/pid"
	"github.com/eviltomorrow/project-n7/lib/procutil"
	"github.com/eviltomorrow/project-n7/lib/runtimeutil"
	"github.com/eviltomorrow/project-n7/lib/self"
	"github.com/eviltomorrow/project-n7/lib/zlog"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"google.golang.org/grpc/resolver"
)

var workflowsFunc = []func() error{
	setRuntime,
	loadConfig,
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
		if isDaemon := ctx.Bool("daemon"); isDaemon {
			if err := procutil.RunInBackground(runtimeutil.AppName, []string{"start"}, nil, nil); err != nil {
				log.Fatalf("[F] Run app in background failure, nest error: %v", err)
			}
			return nil
		}

		defer func() {
			for _, err := range self.RunClearFuncs() {
				zlog.Error("run clear funcs failure", zap.Error(err))
			}
		}()

		for _, f := range workflowsFunc {
			if err := f(); err != nil {
				return err
			}
		}
		procutil.WaitForSigterm()
		return nil
	},
}

var cfg = conf.DefaultGlobal

func loadConfig() error {
	if err := cfg.ParseFile(filepath.Join(runtimeutil.ExecutableDir, "../etc/global.conf")); err != nil {
		return err
	}

	closeFuncs, err := conf.SetupLogger(cfg.Log)
	if err != nil {
		return err
	}
	self.RegisterClearFuncs(closeFuncs...)

	zlog.Info("Load config file complete", zap.String("conf", cfg.String()))
	return nil
}

func setGlobal() error {
	etcd.Endpoints = cfg.Etcd.Endpoints
	middleware.LogDir = filepath.Join(runtimeutil.ExecutableDir, "../log")

	server.ListenHost = cfg.Server.Host
	server.Port = cfg.Server.Port
	return nil
}

func setRuntime() error {
	for _, dir := range []string{
		filepath.Join(runtimeutil.ExecutableDir, "../log"),
		filepath.Join(runtimeutil.ExecutableDir, "../var/run"),
	} {
		if err := fs.CreateDir(dir); err != nil {
			return fmt.Errorf("create dir failure, nest error: %v", err)
		}
	}
	return nil
}

func runServer() error {
	smtp, err := conf.FindSMTP(filepath.Join(runtimeutil.ExecutableDir, cfg.SmtpFile))
	if err != nil {
		return err
	}
	client, err := etcd.NewClient()
	if err != nil {
		return err
	}
	self.RegisterClearFuncs(client.Close)

	if err := middleware.InitLogger(); err != nil {
		return err
	}
	resolver.Register(lb.NewBuilder(client))

	var g = &server.GRPC{
		AppName: runtimeutil.AppName,
		Client:  client,
		SMTP:    smtp,
	}
	if err := g.Startup(); err != nil {
		return err
	}
	self.RegisterClearFuncs(g.Shutdown)

	zlog.Info("Startup GRPC Server complete", zap.String("addrs", fmt.Sprintf("%s:%d", server.ListenHost, server.Port)))
	return nil
}

func rewritePaniclog() error {
	fs.StderrFilePath = filepath.Join(runtimeutil.ExecutableDir, "../log/panic.log")
	if err := fs.RewriteStderrFile(); err != nil {
		zlog.Error("RewriteStderrFile failure", zap.Error(err))
	}
	return nil
}

func buildPidFile() error {
	closeFunc, err := pid.CreatePidFile(filepath.Join(runtimeutil.ExecutableDir, fmt.Sprintf("../var/run/%s.pid", runtimeutil.AppName)))
	if err != nil {
		return err
	}
	self.RegisterClearFuncs(closeFunc)
	return nil
}
