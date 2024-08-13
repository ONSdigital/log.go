package config_test

import (
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/ONSdigital/log.go/v3/config"
	"github.com/ONSdigital/log.go/v3/log"
	. "github.com/smartystreets/goconvey/convey"
)

func TestFromEnv(t *testing.T) {
	Convey("Starting with a clear environment", t, func() {
		os.Clearenv()

		Convey("No vars leaves config unchanged", func() {
			wanted := config.Config{Namespace: "some namespace"}
			cfg := wanted
			config.FromEnv(&cfg)
			So(cfg, ShouldEqual, wanted)
		})

		Convey("Setting HUMAN_LOG updates pretty config", func() {
			cfg := config.Config{}

			tests := []struct {
				value  string
				wanted bool
			}{
				{"1", true},
				{"0", false},
				{"t", true},
				{"f", false},
				{"true", true},
				{"false", false},
			}

			for _, test := range tests {
				Convey("When HUMAN_LOG="+test.value, func() {
					err := os.Setenv("HUMAN_LOG", test.value)
					So(err, ShouldBeNil)
					config.FromEnv(&cfg)
					So(cfg.Pretty, ShouldEqual, test.wanted)
				})
			}
		})

		Convey("Setting LOG_LEVEL updates level in config", func() {
			initialLevel := slog.Level(327)
			cfg := config.Config{Level: initialLevel}

			tests := []struct {
				value  string
				wanted slog.Level
			}{
				{"DEBUG", log.LevelDebug},
				{"debug", log.LevelDebug},
				{"Debug", log.LevelDebug},
				{"-4", log.LevelDebug},

				{"INFO", log.LevelInfo},
				{"info", log.LevelInfo},
				{"Info", log.LevelInfo},
				{"0", log.LevelInfo},

				{"WARN", log.LevelWarn},
				{"warn", log.LevelWarn},
				{"Warn", log.LevelWarn},
				{"4", log.LevelWarn},

				{"ERROR", log.LevelError},
				{"error", log.LevelError},
				{"Error", log.LevelError},
				{"8", log.LevelError},

				{"FATAL", log.LevelFatal},
				{"fatal", log.LevelFatal},
				{"Fatal", log.LevelFatal},
				{"12", log.LevelFatal},

				{"99", slog.Level(99)},

				// Invalid config leaves level in config unchanged
				{"nonsense", initialLevel},
			}

			for _, test := range tests {
				Convey("When HUMAN_LOG="+test.value, func() {
					cfg.Level = initialLevel
					err := os.Setenv("LOG_LEVEL", test.value)
					So(err, ShouldBeNil)
					config.FromEnv(&cfg)
					So(cfg.Level, ShouldEqual, test.wanted)
				})
			}
		})
	})
}

func TestFromOptions(t *testing.T) {
	const (
		ns = "some_namespace"
	)

	Convey("With some predefined config options", t, func() {
		prettyTrueOpt := func(cfg *config.Config) { cfg.Pretty = true }
		prettyFalseOpt := func(cfg *config.Config) { cfg.Pretty = false }
		nsOpt := func(cfg *config.Config) { cfg.Namespace = ns }

		Convey("No opts leaves config unchanged", func() {
			cfg := config.FromOptions()
			So(cfg, ShouldNotBeNil)
			So(*cfg, ShouldEqual, config.Config{})
		})

		Convey("Single opt should set appropriate value", func() {
			cfg := config.FromOptions(prettyTrueOpt)
			So(cfg, ShouldNotBeNil)
			So(cfg.Pretty, ShouldBeTrue)
		})

		Convey("Duplicate opts should use last value", func() {
			cfg := config.FromOptions(prettyTrueOpt, prettyFalseOpt)
			So(cfg, ShouldNotBeNil)
			So(cfg.Pretty, ShouldBeFalse)
		})

		Convey("Multiple opts should not conflict", func() {
			cfg := config.FromOptions(prettyTrueOpt, nsOpt)
			So(cfg, ShouldNotBeNil)
			So(cfg.Pretty, ShouldBeTrue)
			So(cfg.Namespace, ShouldEqual, ns)
		})
	})
}

func TestLevel(t *testing.T) {
	initialLevel := slog.Level(327)

	Convey("Starting with a blank config", t, func() {
		cfg := config.Config{Level: initialLevel}

		tests := []slog.Level{
			log.LevelDebug,
			log.LevelWarn,
			log.LevelInfo,
			log.LevelError,
			log.LevelFatal,
			slog.Level(123),
		}
		for _, test := range tests {
			Convey(fmt.Sprintf("With an option for level=%d", test), func() {
				cfg.Level = initialLevel

				// create an opt for the level and apply it to the config struct
				opt := config.Level(int(test))
				opt(&cfg)
				So(cfg.Level, ShouldEqual, test)
			})
		}
	})
}

func TestNamespace(t *testing.T) {
	initialNS := "initial"

	Convey("Starting with a blank config", t, func() {
		cfg := config.Config{Namespace: initialNS}

		tests := []string{
			"",
			"something",
			"something_else",
			"rAnDoM CAPs!!!",
		}
		for _, test := range tests {
			Convey(fmt.Sprintf("With an option for namespace=%s", test), func() {
				cfg.Namespace = initialNS

				// create an opt for the level and apply it to the config struct
				opt := config.Namespace(test)
				opt(&cfg)
				So(cfg.Namespace, ShouldEqual, test)
			})
		}
	})
}

func TestPretty(t *testing.T) {
	Convey("Starting with a blank config", t, func() {
		cfg := config.Config{}

		Convey("Config Pretty defaults to false", func() {
			So(cfg.Pretty, ShouldBeFalse)
		})

		Convey("Pretty true when option applied", func() {
			opt := config.Pretty
			opt(&cfg)
			So(cfg.Pretty, ShouldBeTrue)
		})

		Convey("Pretty true when multiple options applied", func() {
			opt := config.Pretty
			opt(&cfg)
			opt(&cfg)
			So(cfg.Pretty, ShouldBeTrue)
		})
	})
}
