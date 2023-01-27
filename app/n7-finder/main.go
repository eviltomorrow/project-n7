package main

import (
	"log"
	"os"
	"time"

	"github.com/eviltomorrow/project-n7/app/n7-finder/cmd"
	"github.com/eviltomorrow/project-n7/lib/buildinfo"
	"github.com/eviltomorrow/project-n7/lib/runtimeutil"
	"github.com/urfave/cli/v2"
)

var (
	MainVersion = "unknown"
	GitSha      = "unknown"
	GitTag      = "unknown"
	GitBranch   = "unknown"
	BuildTime   = "unknown"
)

func init() {
	buildinfo.MainVersion = MainVersion
	buildinfo.GitSha = GitSha
	buildinfo.GitTag = GitTag
	buildinfo.GitBranch = GitBranch
	buildinfo.BuildTime = BuildTime
}

func main() {
	if err := runApp(); err != nil {
		log.Printf("[F] Run app failure, nest error: %v", err)
		os.Exit(1)
	}
}

func runApp() error {
	var commands = make([]*cli.Command, 0, 8)
	registerCommands := func(c *cli.Command) {
		commands = append(commands, c)
	}

	registerCommands(cmd.StartCommand)
	registerCommands(cmd.StopCommand)
	registerCommands(cmd.VersionCommand)

	var app = &cli.App{
		Name:     runtimeutil.AppName,
		Usage:    "service for finder",
		Version:  buildinfo.MainVersion,
		Compiled: time.Now(),
		Authors: []*cli.Author{
			{
				Name:  "shepard",
				Email: "eviltomorrow@163.com",
			},
		},
		Description:          "nothing",
		HideHelpCommand:      true,
		HideVersion:          true,
		EnableBashCompletion: true,
		Commands:             commands,
	}

	return app.Run(os.Args)
}
