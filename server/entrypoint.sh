#!/bin/sh
# Wait for database to be ready
until nc -z db 5433; do
  echo "Waiting for database..."
  sleep 1
done
# Run migrations
echo "Running database migrations..."
goose -dir /app/sql/schema postgres "postgres://denethor:denethor@db:5433/denethor?sslmode=disable" up

# Start the server
exec "$@"
