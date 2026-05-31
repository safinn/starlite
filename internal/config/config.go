package config

import (
	"github.com/peterbourgon/ff/v4"
)

type RunMode string

const (
	RunModeServer  RunMode = "server"
	RunModeAuto    RunMode = "auto"
	RunModeMigrate RunMode = "migrate"
	RunModeDown    RunMode = "down"
)

func (c Config) IsDev() bool {
	return c.Env == "dev"
}

func (c Config) IsProd() bool {
	return c.Env == "prod"
}

type Config struct {
	Env         string
	Port        int
	Host        string
	LogLevel    string
	LogFormat   string
	DatabaseURL string
	RunMode     RunMode
	configFile  string
}

func Load(args []string) (Config, error) {
	var cfg Config

	fs := ff.NewFlagSet("app")
	fs.StringEnumVar(&cfg.Env, 'e', "env",
		"environment: dev, stage/staging, prod",
		"dev", "stage", "staging", "prod",
	)
	fs.IntVar(&cfg.Port, 'p', "port", 8080, "port")
	fs.StringVar(&cfg.Host, 'h', "host", "", "host")
	fs.StringEnumVar(&cfg.LogLevel, 'l', "log",
		"log level: debug, info, warn/warning, error",
		"info", "debug", "warn", "warning", "error",
	)
	fs.StringEnumVar(&cfg.LogFormat, 'f', "format",
		"log format: text, json",
		"text", "json",
	)
	fs.StringVar(&cfg.DatabaseURL, 'd', "db-url", "", "database connection URL")
	fs.StringEnumVar((*string)(&cfg.RunMode), 'r', "run-mode",
		"run mode: server, auto (migrate then serve), migrate (migrate only), down (rollback 1)",
		"server", "auto", "migrate", "down",
	)
	fs.StringVar(&cfg.configFile, 'c', "config", "config", "config file name")

	err := ff.Parse(fs, args,
		ff.WithEnvVars(),
		ff.WithConfigFileParser(ff.PlainParser),
		ff.WithConfigFileFlag("config"),
		ff.WithConfigAllowMissingFile(),
	)

	return cfg, err
}
