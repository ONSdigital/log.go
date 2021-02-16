#!/bin/bash -eux

cwd=$(pwd)

pushd $cwd/log.go
  make build
popd