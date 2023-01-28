package buildinfo

import (
	"runtime"

	"github.com/eviltomorrow/project-n7/lib/runtimeutil"
	"github.com/fatih/color"
)

var (
	MainVersion string
	GoVersion   = runtime.Version()
	GoOSArch    = runtime.GOOS + "/" + runtime.GOARCH
	GitSha      string
	BuildTime   string
)

var (
	bold     = color.New(color.Bold)
	bluebold = color.New(color.FgBlue, color.Bold)
)

/*
%s version RELEASE.%s (%s)
Runtime: %s %s
*/

func GetVersion() string {
	var line1 = bold.Sprintf("%s version[%s] (commit-id=%s)", runtimeutil.AppName, MainVersion, GitSha)
	var s1 = bluebold.Sprintf("Runtime: ")
	var s2 = bold.Sprintf("%s %s RELEASE.%s", GoVersion, GoOSArch, BuildTime)
	return line1 + "\r\n" + s1 + s2
}
