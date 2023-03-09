package cmd

import (
	"fmt"
	"path/filepath"
	"syscall"

	"github.com/eviltomorrow/project-n7/lib/helper"
	"github.com/eviltomorrow/project-n7/lib/procutil"
	"github.com/urfave/cli/v2"
)

var StopCommand = &cli.Command{
	Name:  "stop",
	Usage: "stop  app",
	Action: func(ctx *cli.Context) error {
		var pidFile = filepath.Join(helper.Runtime.RootDir, fmt.Sprintf("/var/run/%s.pid", helper.App.Name))
		process, err := procutil.FindWithPidFile(pidFile)
		if err != nil {
			return err
		} else {
			if err := process.Signal(syscall.SIGQUIT); err != nil {
				return nil
			}
		}
		return nil
	},
}
