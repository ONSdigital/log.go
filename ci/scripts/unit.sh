#!/bin/bash -eux

pushd log.go
  make test
popd
