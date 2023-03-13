package fs

import (
	"os"
	"runtime"
	"syscall"
)

var (
	stdErrFileHandler *os.File

	StderrFilePath = "/log/panic.log"
)

func RewriteStderrFile() error {
	file, err := os.OpenFile(StderrFilePath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	stdErrFileHandler = file

	if err = syscall.Dup2(int(file.Fd()), int(os.Stderr.Fd())); err != nil {
		return err
	}
	runtime.SetFinalizer(stdErrFileHandler, func(fd *os.File) {
		fd.Close()
	})

	return nil
}
