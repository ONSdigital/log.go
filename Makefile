SHELL=bash

test:
	go test -v -count=1 -race -cover ./...

.PHONY: test

audit:
	go list -json -m all | nancy sleuth --exclude-vulnerability-file ./.nancy-ignore
.PHONY: audit

build:
	go build ./...
.PHONY: build

.PHONY: lint
lint:
	golangci-lint --timeout=10m --fast --enable=gosec --enable=gocritic --enable=gofmt --enable=gocyclo --enable=bodyclose --enable=gocognit run
