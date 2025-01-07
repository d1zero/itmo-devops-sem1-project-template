#!/bin/bash


docker compose up
go install github.com/pressly/goose/v3/cmd/goose@latest

goose -dir db/migrations postgres "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}" status

go run main.go