package conf

import (
	"encoding/json"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/eviltomorrow/project-n7/lib/runtimeutil"
	"github.com/eviltomorrow/project-n7/lib/zlog"
)

type Config struct {
	ServiceName string  `json:"service-name" toml:"service-name"`
	Server      Server  `json:"server" toml:"server"`
	Etcd        Etcd    `json:"etcd" toml:"etcd"`
	MongoDB     MongoDB `json:"mongodb" toml:"mongodb"`
	Watcher     Watcher `json:"watcher" toml:"watcher"`

	Log Log `json:"log" toml:"log"`
}

func (cg *Config) String() string {
	buf, _ := json.Marshal(cg)
	return string(buf)
}

func (c *Config) ParseFile(path string) error {
	_, err := toml.DecodeFile(path, c)
	return err
}

type Server struct {
	Host string `json:"host" toml:"host"`
	Port int    `json:"port" toml:"port"`
}

type Etcd struct {
	Endpoints []string `json:"endpoints" toml:"endpoints"`
}

type MongoDB struct {
	DSN string `json:"dsn" toml:"dsn"`
}

type Watcher struct {
	Source     string   `json:"source" toml:"source"`
	CodeList   []string `json:"code-list" toml:"code-list"`
	Crontab    string   `json:"crontab" toml:"crontab"`
	RandomWait string   `json:"random-wait" toml:"random-wait"`
}

type Log struct {
	DisableTimestamp bool   `json:"disable-timestamp" toml:"disable-timestamp"`
	Level            string `json:"level" toml:"level"`
	Format           string `json:"format" toml:"format"`
	MaxSize          int    `json:"maxsize" toml:"maxsize"`
	MaxDays          int    `toml:"max-days" json:"max-days"`
	Dir              string `toml:"dir" json:"dir"`
	Compress         bool   `toml:"compress" json:"compress"`
}

var DefaultGlobal = &Config{
	ServiceName: "n7-collector",
	Server: Server{
		Host: "0.0.0.0",
		Port: 5270,
	},
	MongoDB: MongoDB{
		DSN: "mongodb://127.0.0.1:27017",
	},
	Etcd: Etcd{
		Endpoints: []string{
			"127.0.0.1:2379",
		},
	},
	Watcher: Watcher{
		Source: "sina",
		CodeList: []string{
			"sh688***",
			"sh605***",
			"sh603***",
			"sh601***",
			"sh600***",
			"sz300***",
			"sz0030**",
			"sz002***",
			"sz001**",
			"sz000***",
		},
		Crontab:    "05 19 * * MON,TUE,WED,THU,FRI",
		RandomWait: "20s,60s",
	},
	Log: Log{
		DisableTimestamp: false,
		Level:            "info",
		Format:           "text",
		MaxSize:          30,
		MaxDays:          180,
		Dir:              "../log",
		Compress:         true,
	},
}

func SetupLogger(l Log) ([]func() error, error) {
	global, prop, err := zlog.InitLogger(&zlog.Config{
		Level:            l.Level,
		Format:           l.Format,
		DisableTimestamp: l.DisableTimestamp,
		File: zlog.FileLogConfig{
			Filename:   filepath.Join(runtimeutil.ExecutableDir, filepath.Join(l.Dir, "data.log")),
			MaxSize:    l.MaxSize,
			MaxDays:    30,
			MaxBackups: 30,
			Compress:   true,
		},
		DisableStacktrace:   true,
		DisableErrorVerbose: true,
	})
	if err != nil {
		return nil, err
	}
	zlog.ReplaceGlobals(global, prop)

	return []func() error{
		func() error { return global.Sync() },
	}, nil
}
