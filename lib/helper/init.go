package helper

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/eviltomorrow/project-n7/lib/netutil"
)

func init() {
	path, err := os.Executable()
	if err != nil {
		panic(fmt.Errorf("panic: Executable path failure, nest error: %v", err))
	}
	path, err = filepath.Abs(path)
	if err != nil {
		panic(fmt.Errorf("panic: Abs path failure, nest error: %v", err))
	}

	executeDir, executeFile := filepath.Dir(path), filepath.Base(path)
	Runtime.ExecuteDir = executeDir
	if strings.HasSuffix(executeDir, "/bin") {
		Runtime.RootDir = filepath.Dir(executeDir)
	} else {
		Runtime.RootDir = executeDir
	}
	App.Name = executeFile

	Runtime.HostName, _ = os.Hostname()
	Runtime.IP, _ = netutil.GetLocalIP2()
}
