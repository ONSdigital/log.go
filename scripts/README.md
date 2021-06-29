This directory contains scripts for refactoring logging statements in .go files.

# Historical logging refactoring: go-ns -> log.go

The file `edit-logs.sh` is from an earlier logging exercise. It refactors from the older `go-ns` logging library to the newer `log.go`, and updates logging statement syntax to match.

## Usage

Execute as `<path to>edit-logs.sh <target directory>`

`edit-logs.sh` will then scan `<target directory>` for .go files, and in-place modify each one found to:
- update logging imports from `go-ns` to `log.go`
- update logging statements to newer syntax (run `edit-logs.sh -h` or read the file itself for more details)

The script can be tested by running against this directory. You should see the file `test-update-logs.go` altered so that the logging statements match those stated in the comments immediately above each one.

# Logging refactoring June/July 2021: log.go v1 -> log.go v2

The file `refactor-go-logging.sh` is from a refactoring exercise run in June/July 2021. It refactors from the older `log.go v1` to the newer `log.go v2`, and updates logging statement syntax to match.

## Usage (single file)

Execute as `<path to>refactor-go-logging.sh < <target file> > <output file>` (note use of single `<` and `>` denote streaming from / to files here, not user-defined input)

`refactor-go-logging.sh` will create `<output file>`, which contains the content of `<target file>` modified as so:
- update logging imports from `log.go` to `log.go/v2`
- update logging statements to newer syntax, e.g.:
```go
log.Event(ctx, string, log.FATAL, log.Error(something1), something2)
	-> log.Fatal(ctx, string, something1, something2)

log.Event(ctx, string, log.ERROR log.Error(something1), something2)
	-> log.Error(ctx, string, something1, something2)

log.Event(ctx, string, log.INFO, something)
	-> log.Info(ctx, string, something)

log.Event(ctx, string, log.WARN, log.Error(something), something2)
	-> log.Warn(ctx, string, log.FormatErrors([]error{something}), something2)
```

## Usage (directory)

The script `refactor-all-go-logging-in-directory.sh` is a wrapper for `refactor-go-logging.sh` which runs it to in-place modify all .go files in `<target directory>`, similar to the original `edit-logs.sh`.

Execute as `<path to>refactor-all-go-logging-in-directory.sh <target directory>`

`refactor-all-go-logging-in-directory.sh` will then scan a target directory for .go files, and in-place modify each one found using the `refactor-go-logging.sh` script.
