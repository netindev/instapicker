#!/bin/bash
set -e

cd "$(dirname "$0")"

if [ ! -f .env ]; then
  echo "error: .env file not found. copy .env.example and fill in your credentials."
  exit 1
fi

echo "building instapicker..."
go build -o bin/instapicker ./cmd/instapicker

echo "running instapicker..."
./bin/instapicker
