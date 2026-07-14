#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
EASYJSON="go run github.com/mailru/easyjson/easyjson@v0.9.0"

DTO_DIRS=(
	app/internal/server/avatar/presentation/http/dto
	app/internal/server/avatar/adapters/repository/rabbitmq/model
	app/internal/processor/processing/adapters/broker/rabbitmq/model
)

for dir in "${DTO_DIRS[@]}"; do
	for src in "$ROOT/$dir"/*.go; do
		[[ "$(basename "$src")" == *_easyjson.go ]] && continue
		(cd "$ROOT" && $EASYJSON -all "$dir/$(basename "$src")")
	done
done

echo "easyjson generated"
