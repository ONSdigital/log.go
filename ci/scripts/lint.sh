#!/bin/bash -eux

cwd=$(pwd)

pushd $cwd/log.go
# Install golangci-lint
  go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.59.1
  make lint
popd
