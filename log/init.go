package log

import (
	"io"
	"log/slog"
	"os"

	"github.com/ONSdigital/log.go/v3/config"
	"github.com/ONSdigital/log.go/v3/log/pretty"
)

// Initialise is a helper function that creates a new [slog.Logger] with common options and sets it as the default logger
// for the app. Specifically it adds a namespace attribute to all logs, gets other config from environment variables.
// It can be configured with extra options passed in as arguments
func Initialise(ns string, opts ...config.Option) {
	// Add EnvVar processor and namespace to config options
	opts = append([]config.Option{config.FromEnv, config.Namespace(ns)}, opts...)

	handler := Handler(opts...)

	SetDefault(slog.New(handler))

	stdHandler := ModifyingHandler{handler}
	slog.SetDefault(slog.New(stdHandler))
}

// Handler is a function that returns a [slog.Logger] based on the supplied [config.Option] vararg.
func Handler(opts ...config.Option) slog.Handler {
	cfg := config.FromOptions(opts...)

	var out io.Writer = os.Stdout
	if cfg.Pretty {
		out = pretty.NewPrettyWriter(out)
	}

	var hdlr slog.Handler = slog.NewJSONHandler(out, &slog.HandlerOptions{Level: cfg.Level, ReplaceAttr: replaceAttr})

	if cfg.Namespace != "" {
		hdlr = hdlr.WithAttrs([]slog.Attr{slog.String("namespace", cfg.Namespace)})
	}

	return hdlr
}

// replaceAttr function to be used by a slog handler to rename attributes and expand errors
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
