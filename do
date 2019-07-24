#! /usr/bin/env bash

for i in ${@}; do
  case ${i} in
    cli)
    echo go clean
    go clean
    echo go build ./examples/cli
    go build ./examples/cli
    ;;
    gin|srv)
    echo go clean
    go clean
    echo go build ./examples/srv
    go build ./examples/srv
    ;;
  esac
done
