#! /usr/bin/env bash

for i in ${@}; do
  case ${i} in
    cli)
    echo go build ./examples/cli
    go build ./examples/cli
    ;;
  esac
done
