package conf

import (
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/eviltomorrow/project-n7/lib/helper"
	"github.com/eviltomorrow/project-n7/lib/zlog"
	jsoniter "github.com/json-iterator/go"
)

type Config struct {
	ServiceName string `json:"service-name" toml:"service-name"`
	Server      Server `json:"server" toml:"server"`
	Etcd        Etcd   `json:"etcd" toml:"etcd"`
	SmtpFile    string `json:"smtp-file" toml:"smtp-file"`

	Log Log `json:"log" toml:"log"`
}

func (cg *Config) String() string {
	buf, _ := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(cg)
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

type Log struct {
	DisableTimestamp bool   `json:"disable-timestamp" toml:"disable-timestamp"`
	Level            string `json:"level" toml:"level"`
	Format           string `json:"format" toml:"format"`
	MaxSize          int    `json:"max-size" toml:"max-size"`
	MaxDays          int    `toml:"max-days" json:"max-days"`
	MaxBackups       int    `toml:"max-backups" json:"max-backups"`
	Dir              string `toml:"dir" json:"dir"`
	Compress         bool   `toml:"compress" json:"compress"`
}

var DefaultGlobal = &Config{
	ServiceName: "n7-email",
	SmtpFile:    "/etc/smtp.json",
	Server: Server{
		Host: "0.0.0.0",
		Port: 5271,
	},
	Etcd: Etcd{
		Endpoints: []string{
			"127.0.0.1:2379",
		},
	},
	Log: Log{
		DisableTimestamp: false,
		Level:            "info",
		Format:           "text",
		MaxSize:          30,
		MaxDays:          30,
		MaxBackups:       30,
		Dir:              "/log",
		Compress:         true,
	},
}

func SetupLogger(l Log) ([]func() error, error) {
	global, prop, err := zlog.InitLogger(&zlog.Config{
		Level:            l.Level,
		Format:           l.Format,
		DisableTimestamp: l.DisableTimestamp,
		File: zlog.FileLogConfig{
			Filename:   filepath.Join(helper.Runtime.RootDir, l.Dir, "data.log"),
			MaxSize:    l.MaxSize,
			MaxDays:    l.MaxDays,
			MaxBackups: l.MaxBackups,
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
