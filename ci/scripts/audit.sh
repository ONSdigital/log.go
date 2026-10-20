#!/bin/bash -eux

pushd log.go
  make audit
popd
