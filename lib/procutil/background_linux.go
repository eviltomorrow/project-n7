//go:build linux
// +build linux

package procutil

import (
	"io"
	"os"
	"os/exec"
)

var (
	HomeDir = "."
)

func RunInBackground(name string, args []string, reader io.Reader, writer io.WriteCloser) error {
	var data = make([]string, 0, len(args)+1)
	data = append(data, name)
	data = append(data, args...)
	var cmd = &exec.Cmd{
		Path:  "/proc/self/exe",
		Args:  data,
		Stdin: reader,
	}
	cmd.Dir = HomeDir
	cmd.Env = os.Environ()

	if writer == nil {
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
	}

	return cmd.Start()
}
