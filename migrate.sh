#!/bin/bash

# Прекращение выполнения при ошибках
set -e

# Применение миграций
echo "Applying database migrations..."
./goose -dir db/migrations postgres "postgres://validator:val1dat0r@db:5432/project-sem-1" up

# Запуск основного процесса
echo "Starting backend..."
exec "$@"