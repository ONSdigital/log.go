package main

import (
	"context"
	"errors"
	"fmt"
	golog "log"
	"log/slog"
	"time"

	"github.com/ONSdigital/log.go/v3/log"
	pkgerrors "github.com/pkg/errors"
	"golang.org/x/xerrors"
)

var exampleStruct = struct {
	A string
	B int
}{"aaaa", 123}

func main() {
	log.Initialise("v3example")

	ctx := context.Background()

	slog.Info("slog with attrs", slog.String("a", "b"), slog.Int("int", 123))

	slog.Info("slog with group", slog.Group("stuff", slog.String("a", "b"), slog.Int("int", 123)))

	// Simple Info
	log.Info(ctx, "v3 simple info")

	// Complex Info
	log.Info(ctx, "v3 info with data", log.Data{"key": "value", "struct": exampleStruct})

	// Error
	err := errors.New("some error")
	log.Error(ctx, "v3 error", err)

	// Fatal
	log.Fatal(ctx, "v3 fatal", err)

	// Examples of different libraries providing error wrapping and stacktraces
	errorWrapping()

	// An example of go standard logging
	golog.Println("go standard logging")

	// An example of go standard structured logging
	slog.Info("standard slog with data",
		slog.String("string_attr", "a string"),
		slog.Group("a_group",
			slog.String("string_attr2", "a string"),
			slog.Int("int_attr", 123)))

	// Give the human formatter time to finish
	time.Sleep(time.Second)
}

func errorWrapping() {
	ctx := context.Background()
	// Standard Library Errors

	err := errors.New("basic error")
	log.Error(ctx, "basic error", err)

	werr := fmt.Errorf("basic wrapped error [%w]", err)
	log.Error(ctx, "basic wrapped error", werr)

	w2err := fmt.Errorf("double wrapped error [%w]", werr)
	log.Error(ctx, "double wrapped error", w2err)

	// github.com/pkg/errors examples

	perr := pkgerrors.New("pkg error")
	log.Error(ctx, "pkg error", perr)

	pwerr := pkgerrors.Wrap(perr, "pkg wrapped error")
	log.Error(ctx, "pkg wrapped error", pwerr)

	pw2err := pkgerrors.Wrap(pwerr, "double pkg wrapped error")
	log.Error(ctx, "double pkg wrapped error", pw2err)

	// golang.org/x/xerrors examples

	xerr := xerrors.New("x error")
	log.Error(ctx, "x error", xerr)

	xwerr := xerrors.Errorf("x wrapped %w", xerr)
	log.Error(ctx, "x wrapped error", xwerr)

	xw2err := xerrors.Errorf("x double wrapped %w", xwerr)
	log.Error(ctx, "x double wrapped error", xw2err)

	// Mixed errors

	pb := pkgerrors.Wrap(err, "pkg wrapped")
	log.Error(ctx, "pkg wrapped basic error", pb)

	xb := xerrors.Errorf("x wrapped %w", err)
	log.Error(ctx, "x wrapped basic error", xb)

	xpb := xerrors.Errorf("x wrapped %w", pb)
	log.Error(ctx, "x wrapped pkg wrapped basic error", xpb)

	pxb := pkgerrors.Wrap(xb, "pkg wrapped")
	log.Error(ctx, "pkg wrapped x wrapped basic error", pxb)
}
