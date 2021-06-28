#!/bin/bash -eux

cwd=$(pwd)

pushd $cwd/log.go
  make lint
popd
