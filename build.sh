#!/usr/bin/env bash

mkdir -p temp

GOOS=linux GOARCH=386 go build -o temp/gdns cli/gdns.go

