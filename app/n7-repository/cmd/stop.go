package cmd

import (
	"fmt"
	"path/filepath"
	"syscall"

	"github.com/eviltomorrow/project-n7/lib/procutil"
	"github.com/eviltomorrow/project-n7/lib/runtimeutil"
	"github.com/urfave/cli/v2"
)

var StopCommand = &cli.Command{
	Name:  "stop",
	Usage: "stop  app",
	Action: func(ctx *cli.Context) error {
		var pidFile = filepath.Join(runtimeutil.ExecutableDir, fmt.Sprintf("../var/run/%s.pid", runtimeutil.AppName))
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
