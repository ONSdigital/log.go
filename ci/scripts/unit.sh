#!/bin/bash -eux

cwd=$(pwd)

pushd $cwd/log.go
  make test
popd