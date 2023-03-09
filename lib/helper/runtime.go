package helper

import (
	"os"
	"runtime"
	"time"

	"github.com/eviltomorrow/project-n7/lib/timeutil"
)

var (
	now     = time.Now()
	Runtime = runtimeHelper{
		ARCH:       runtime.GOARCH,
		OS:         runtime.GOOS,
		Pid:        os.Getpid(),
		LaunchTime: now,
		RunningDuration: func() string {
			return timeutil.FormatDuration(time.Since(now))
		},
	}
)

type runtimeHelper struct {
	ExecuteDir      string        `json:"execute-dir"`
	RootDir         string        `json:"root-dir"`
	Pid             int           `json:"pid"`
	LaunchTime      time.Time     `json:"launch-time"`
	HostName        string        `json:"host-name"`
	OS              string        `json:"os"`
	ARCH            string        `json:"arch"`
	RunningDuration func() string `json:"-"`
	IP              string        `json:"ip"`
}
