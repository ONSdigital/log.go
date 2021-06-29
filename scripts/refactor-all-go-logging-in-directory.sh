#!/bin/bash

# Runs sed script refactor-go-logging.sh against all .go files found in TARGET_DIR given as argument $1
# NB: will edit all files in place!

THIS_DIR=$(realpath $(dirname $0))
TARGET_DIR=$1
find ${TARGET_DIR} -type f -name "*.go" -print0 | while read -d $'\0' file
do
  ${THIS_DIR}/refactor-go-logging.sh < ${file} > ${file}.new && mv ${file}.new ${file}
done

