package conf

import (
	"encoding/json"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/eviltomorrow/project-n7/lib/helper"
	"github.com/eviltomorrow/project-n7/lib/zlog"
)

type Config struct {
	ServiceName string `json:"service-name" toml:"service-name"`
	Server      Server `json:"server" toml:"server"`
	Etcd        Etcd   `json:"etcd" toml:"etcd"`
	MySQL       MySQL  `json:"mysql" toml:"mysql"`

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

type MySQL struct {
	DSN     string `json:"dsn" toml:"dsn"`
	MinOpen int    `json:"min-open" toml:"min-open"`
	MaxOpen int    `json:"max-open" toml:"max-open"`
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
	ServiceName: "n7-repository",
	Server: Server{
		Host: "0.0.0.0",
		Port: 5272,
	},
	Etcd: Etcd{
		Endpoints: []string{
			"127.0.0.1:2379",
		},
	},
	MySQL: MySQL{
		DSN:     "root:root@tcp(127.0.0.1:3306)/rogue_repo?charset=utf8mb4&parseTime=true&loc=Local",
		MinOpen: 3,
		MaxOpen: 10,
	},
	Log: Log{
		DisableTimestamp: false,
		Level:            "info",
		Format:           "text",
		MaxSize:          30,
		MaxDays:          180,
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
