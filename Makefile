SHELL=bash

test:
	go test -v -count=1 -race -cover ./...

.PHONY: test

audit:
	dis-vulncheck
.PHONY: audit

build:
	go build ./...
.PHONY: build

.PHONY: lint
lint:
	golangci-lint run ./...
