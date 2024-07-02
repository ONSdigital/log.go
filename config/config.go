package config

import (
	"errors"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Namespace string
	Pretty    bool
	Level     slog.Level
}
type Option func(cfg *Config)

// FromEnv is a config option that processes environment variables and populates the config with the contents of those vars
var FromEnv Option = func(cfg *Config) {
	humanLog, _ := strconv.ParseBool(os.Getenv("HUMAN_LOG"))
	cfg.Pretty = humanLog

	if s := os.Getenv("LOG_LEVEL"); s != "" {
		if l, err := levelFromString(s); err == nil {
			cfg.Level = l
		}
	}
}

// Pretty is an option that makes the output human-readable
var Pretty Option = func(cfg *Config) {
	cfg.Pretty = true
}

func levelFromString(s string) (slog.Level, error) {
	switch strings.ToUpper(s) {
	case "DEBUG":
		return slog.LevelDebug, nil
	case "INFO":
		return slog.LevelInfo, nil
	case "WARN":
		return slog.LevelWarn, nil
	case "ERROR":
		return slog.LevelError, nil
	case "FATAL":
		return slog.Level(12), nil
	}

	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return slog.Level(i), nil
	}

	return 0, errors.New("log level string unrecognised:" + s)
}

// Namespace returns a config option that populates a config with the namespace of the logger
func Namespace(ns string) Option {
	return func(cfg *Config) {
		cfg.Namespace = ns
	}
}

// Level returns a config option that populates a config with the log level of the logger
func Level(l int) Option {
	return func(cfg *Config) {
		cfg.Level = slog.Level(l)
	}
}

// FromOptions returns a populated config derived from the provided Options
func FromOptions(opts ...Option) *Config {
	cfg := &Config{}

	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}
