package log

import (
	"github.com/ONSdigital/log.go/v3/config"
	"github.com/ONSdigital/log.go/v3/log/pretty"
	"io"
	"log/slog"
	"os"
)

// Initialise is a helper function that creates a new [slog.Logger] with common options and sets it as the default logger
// for the app. Specifically it adds a namespace attribute to all logs, gets other config from environment variables.
// It can be configured with extra options passed in as arguments
func Initialise(ns string, opts ...config.Option) {
	// Add EnvVar processor and namespace to config options
	opts = append([]config.Option{config.FromEnv, config.Namespace(ns)}, opts...)

	logger := Logger(opts...)

	SetDefault(logger)

	stdLogger := logger.With(slog.Int("severity", int(INFO))).WithGroup("data")
	slog.SetDefault(stdLogger)
}

// Logger is a function that returns a [slog.Logger] based on the supplied [config.Option] vararg.
func Logger(opts ...config.Option) *slog.Logger {
	cfg := config.FromOptions(opts...)

	var out io.Writer = os.Stdout
	if cfg.Pretty {
		out = pretty.NewPrettyWriter(out)
	}

	logger := slog.New(slog.NewJSONHandler(out, &slog.HandlerOptions{Level: cfg.Level, ReplaceAttr: replaceAttr}))

	if cfg.Namespace != "" {
		logger = logger.With(slog.String("namespace", cfg.Namespace))
	}

	return logger
}

func replaceAttr(groups []string, a slog.Attr) slog.Attr {
	if a.Key == "msg" {
		a.Key = "event"
	}

	switch a.Value.Kind() {
	case slog.KindTime:
		if groups == nil && a.Key == "time" {
			a.Key = "created_at"
			t := a.Value.Time()
			a.Value = slog.TimeValue(t.UTC())
		}

	case slog.KindAny:
		switch v := a.Value.Any().(type) {
		case error:
			a = slog.Any("errors", FormatAsErrors(v))
		case slog.Level:
			if a.Value.Any().(slog.Level) == LevelFatal {
				a.Value = slog.StringValue("FATAL")
			}
		}
	}

	return a
}
