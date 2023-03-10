package cmd

import (
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/eviltomorrow/project-n7/app/n7-collector/conf"
	"github.com/eviltomorrow/project-n7/app/n7-collector/handler/crontab"
	"github.com/eviltomorrow/project-n7/app/n7-collector/handler/sync"
	"github.com/eviltomorrow/project-n7/app/n7-collector/server"
	"github.com/eviltomorrow/project-n7/lib/cleanup"
	"github.com/eviltomorrow/project-n7/lib/etcd"
	"github.com/eviltomorrow/project-n7/lib/fs"
	"github.com/eviltomorrow/project-n7/lib/grpc/lb"
	"github.com/eviltomorrow/project-n7/lib/grpc/middleware"
	"github.com/eviltomorrow/project-n7/lib/helper"
	"github.com/eviltomorrow/project-n7/lib/mongodb"
	"github.com/eviltomorrow/project-n7/lib/pid"
	"github.com/eviltomorrow/project-n7/lib/procutil"

	"github.com/eviltomorrow/project-n7/lib/zlog"
	"github.com/robfig/cron/v3"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"google.golang.org/grpc/resolver"
)

var workflowsFunc = []func() error{
	setRuntime,
	loadConfig,
	printCfg,
	setGlobal,
	runDB,
	runCron,
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
				zlog.Error("run cleanup funcs failure", zap.Error(err))
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
	middleware.LogDir = filepath.Join(helper.Runtime.RootDir, "/log")
	server.ListenHost = cfg.Server.Host
	server.Port = cfg.Server.Port
	etcd.Endpoints = cfg.Etcd.Endpoints
	mongodb.DSN = cfg.MongoDB.DSN
	crontab.Source = cfg.Watcher.Source
	sync.CodeList = cfg.Watcher.CodeList
	if strings.Count(cfg.Watcher.RandomWait, ",") == 1 {
		var attrs = strings.Split(cfg.Watcher.RandomWait, ",")
		v1, _ := strconv.Atoi(attrs[0])
		v2, _ := strconv.Atoi(attrs[1])
		if v1 < v2 && v1 > 0 && v2 < 100 {
			sync.RandomWait = [2]int{v1, v2}
		}
	}
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

func runDB() error {
	if err := mongodb.Build(); err != nil {
		return err
	}
	cleanup.RegisterCleanupFuncs(mongodb.Close)

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
	zlog.Info("Startup GRPC server success", zap.String("host", server.ListenHost), zap.Int("port", server.Port))
	return nil
}

func runCron() error {
	var c = cron.New()
	_, err := c.AddFunc(cfg.Watcher.Crontab, func() {
		if err := crontab.EveryWeekDay(); err != nil {
			zlog.Error("Crontab run EveryWeekDay failure", zap.Error(err))
		}
	})
	if err != nil {
		return err
	}
	c.Start()

	cleanup.RegisterCleanupFuncs(func() error {
		c.Stop()
		return nil
	})
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
