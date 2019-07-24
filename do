#! /usr/bin/env bash

for i in ${@}; do
  case ${i} in
    cli)
    echo go clean
    go clean
    echo go build ./examples/cli
    go build ./examples/cli
    ;;
    gin)
    echo go clean
    go clean
    echo go build ./examples/srv
    go build ./examples/srv
    ;;
    srv2)
    echo go clean
    go clean
    echo go build -tags sessembed ./examples/srv2
    go build -tags sessembed ./examples/srv2
    ;;
  esac
done
