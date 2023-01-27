//go:build darwin
// +build darwin

package procutil

import (
	"fmt"
	"io"
)

var (
	HomeDir = "."
)

func RunInBackground(name string, args []string, reader io.Reader, writer io.WriteCloser) error {
	return fmt.Errorf("not support")
}
