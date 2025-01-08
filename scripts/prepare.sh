#!/bin/bash

go install github.com/pressly/goose/v3/cmd/goose@latest
go build -ldflags="-s -w" -o main cmd/api/main.go

