#!/bin/bash -eux

cwd=$(pwd)

# Install golangci-lint
  lint_ver=1.46.2
  curl --location --no-progress-meter https://github.com/golangci/golangci-lint/releases/download/v$lint_ver/golangci-lint-$lint_ver-linux-amd64.tar.gz | tar zxvf -
  PATH=$PATH:$cwd/golangci-lint-$lint_ver-linux-amd64

pushd $cwd/log.go
  make lint
popd
