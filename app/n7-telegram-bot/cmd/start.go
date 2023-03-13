package cmd

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/eviltomorrow/project-n7/app/n7-telegram-bot/conf"
	"github.com/eviltomorrow/project-n7/app/n7-telegram-bot/handler/telegrambot"
	"github.com/eviltomorrow/project-n7/app/n7-telegram-bot/server"
	"github.com/eviltomorrow/project-n7/lib/cleanup"
	"github.com/eviltomorrow/project-n7/lib/fs"
	"github.com/eviltomorrow/project-n7/lib/grpc/middleware"
	"github.com/eviltomorrow/project-n7/lib/helper"
	"github.com/eviltomorrow/project-n7/lib/pid"
	"github.com/eviltomorrow/project-n7/lib/procutil"
	"github.com/eviltomorrow/project-n7/lib/zlog"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

var workflowsFunc = []func() error{
	setRuntime,
	loadConfig,
	printCfg,
	setGlobal,
	runBot,
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
	middleware.LogDir = filepath.Join(helper.Runtime.RootDir, "/log")

	server.ListenHost = cfg.Server.Host
	server.Port = cfg.Server.Port
	return nil
}

func setRuntime() error {
	for _, dir := range []string{
		filepath.Join(helper.Runtime.RootDir, "/log"),
		filepath.Join(helper.Runtime.RootDir, "/var/run"),
		filepath.Join(helper.Runtime.RootDir, "/db"),
	} {
		if err := fs.CreateDir(dir); err != nil {
			return fmt.Errorf("create dir failure, nest error: %v", err)
		}
	}
	return nil
}

func runBot() error {
	botC, err := conf.FindTelegramBot(filepath.Join(helper.Runtime.RootDir, cfg.BotFile))
	if err != nil {
		return err
	}
	var bot = &telegrambot.Bot{
		DomainName:  botC.DomainName,
		Pattern:     botC.Pattern,
		Port:        botC.Port,
		AccessToken: botC.AccessToken,
	}
	if err := bot.Run(); err != nil {
		return err
	}

	server.Bot = bot
	cleanup.RegisterCleanupFuncs(bot.Stop)
	return nil
}

func runServer() error {
	tb, err := conf.FindTelegramBot(filepath.Join(helper.Runtime.RootDir, cfg.BotFile))
	if err != nil {
		return err
	}

	if err := middleware.InitLogger(); err != nil {
		return err
	}

	var g = &server.GRPC{
		AppName: helper.App.Name,
		TB:      tb,
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
