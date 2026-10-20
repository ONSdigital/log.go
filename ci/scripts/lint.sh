#!/bin/bash -eux

pushd log.go
  make lint
popd
