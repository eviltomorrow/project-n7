package cmd

import (
	"fmt"

	"github.com/eviltomorrow/project-n7/lib/buildinfo"
	"github.com/urfave/cli/v2"
)

var VersionCommand = &cli.Command{
	Name:  "version",
	Usage: "print version info",
	Action: func(ctx *cli.Context) error {
		fmt.Println(buildinfo.GetVersion())
		return nil
	},
}
