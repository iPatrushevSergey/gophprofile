#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
MOCKGEN="go run go.uber.org/mock/mockgen@v0.6.0"

gen() {
  local src="$1"
  local dest="$2"
  mkdir -p "$(dirname "$ROOT/$dest")"
  (cd "$ROOT" && $MOCKGEN -source="$src" -destination="$dest" -package=mocks)
}

# shared
gen app/internal/pkg/port/logger.go app/internal/pkg/port/mocks/mock_logger.go
gen app/internal/pkg/port/tracing.go app/internal/pkg/port/mocks/mock_tracer.go

# server avatar
gen app/internal/server/avatar/application/port/avatar_repository.go app/internal/server/avatar/application/port/mocks/mock_avatar_repository.go
gen app/internal/server/avatar/application/port/avatar_storage.go app/internal/server/avatar/application/port/mocks/mock_avatar_storage.go
gen app/internal/server/avatar/application/port/outbox.go app/internal/server/avatar/application/port/mocks/mock_outbox.go
gen app/internal/server/avatar/application/port/event_publisher.go app/internal/server/avatar/application/port/mocks/mock_event_publisher.go
gen app/internal/server/avatar/application/port/transactor.go app/internal/server/avatar/application/port/mocks/mock_transactor.go
gen app/internal/server/avatar/application/port/clock.go app/internal/server/avatar/application/port/mocks/mock_clock.go
gen app/internal/server/avatar/application/port/id_generator.go app/internal/server/avatar/application/port/mocks/mock_id_generator.go

# processor processing
gen app/internal/processor/processing/application/port/avatar_repository.go app/internal/processor/processing/application/port/mocks/mock_avatar_repository.go
gen app/internal/processor/processing/application/port/avatar_storage.go app/internal/processor/processing/application/port/mocks/mock_avatar_storage.go
gen app/internal/processor/processing/application/port/event_consumer.go app/internal/processor/processing/application/port/mocks/mock_event_consumer.go
gen app/internal/processor/processing/application/port/image_processor.go app/internal/processor/processing/application/port/mocks/mock_image_processor.go
gen app/internal/processor/processing/application/port/transactor.go app/internal/processor/processing/application/port/mocks/mock_transactor.go
gen app/internal/processor/processing/application/port/clock.go app/internal/processor/processing/application/port/mocks/mock_clock.go

echo "mocks generated"
