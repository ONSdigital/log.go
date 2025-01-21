#!/bin/bash -eux

cwd=$(pwd)

pushd $cwd/log.go
# Install golangci-lint
  go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.62.0
  make lint
popd
