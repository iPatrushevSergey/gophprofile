.PHONY: build build-server build-processor build-gencerts build-all \
	build-linux-amd64 build-windows-amd64 build-darwin-amd64 \
	certs run-server run-processor run-server-prod run-processor-prod \
	docker-up docker-down docker-ps docker-migrate observability-smoke \
	test test-unit test-contract test-component test-integration test-e2e test-all \
	cover cover-unit cover-integration migrate generate-mocks \
	generate-easyjson generate-goverter lint

APP_DIR := app
BIN_DIR := $(APP_DIR)/bin
DIST_DIR := $(BIN_DIR)/dist

BUILD_FLAGS := -tags=go_json

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -X 'github.com/iPatrushevSergey/gophprofile/app/internal/pkg/apputil.Version=$(VERSION)' \
	-X 'github.com/iPatrushevSergey/gophprofile/app/internal/pkg/apputil.Date=$(BUILD_DATE)'

BUILD_LDFLAGS := -ldflags "$(LDFLAGS)"

build: build-server build-processor

build-server:
	go build $(BUILD_FLAGS) $(BUILD_LDFLAGS) -o $(BIN_DIR)/server ./app/cmd/server

build-processor:
	go build $(BUILD_FLAGS) $(BUILD_LDFLAGS) -o $(BIN_DIR)/processor ./app/cmd/processor

build-gencerts:
	go build -o $(BIN_DIR)/gencerts ./app/cmd/gencerts

build-all: build-linux-amd64 build-windows-amd64 build-darwin-amd64

build-linux-amd64:
	GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) $(BUILD_LDFLAGS) -o $(DIST_DIR)/linux-amd64/server ./app/cmd/server
	GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) $(BUILD_LDFLAGS) -o $(DIST_DIR)/linux-amd64/processor ./app/cmd/processor

build-windows-amd64:
	GOOS=windows GOARCH=amd64 go build $(BUILD_FLAGS) $(BUILD_LDFLAGS) -o $(DIST_DIR)/windows-amd64/server.exe ./app/cmd/server
	GOOS=windows GOARCH=amd64 go build $(BUILD_FLAGS) $(BUILD_LDFLAGS) -o $(DIST_DIR)/windows-amd64/processor.exe ./app/cmd/processor

build-darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build $(BUILD_FLAGS) $(BUILD_LDFLAGS) -o $(DIST_DIR)/darwin-amd64/server ./app/cmd/server
	GOOS=darwin GOARCH=amd64 go build $(BUILD_FLAGS) $(BUILD_LDFLAGS) -o $(DIST_DIR)/darwin-amd64/processor ./app/cmd/processor

certs:
	go run ./app/cmd/gencerts -out certs

run-server:
	go run $(BUILD_FLAGS) ./app/cmd/server -c app/configs/server.yaml

run-processor:
	go run $(BUILD_FLAGS) ./app/cmd/processor -c app/configs/processor.yaml

run-server-prod:
	go run $(BUILD_FLAGS) ./app/cmd/server -c app/configs/server.prod.yaml

run-processor-prod:
	go run $(BUILD_FLAGS) ./app/cmd/processor -c app/configs/processor.prod.yaml

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down

docker-ps:
	docker compose ps

docker-migrate:
	docker compose run --rm --build migrate

observability-smoke:
	bash scripts/observability-smoke.sh

# Tests pyramid: unit → contract → integration → component → e2e.
test:
	$(MAKE) test-unit

test-unit:
	go test $(BUILD_FLAGS) ./app/internal/... ./app/cmd/...

test-contract:
	go test $(BUILD_FLAGS) -tags=contract -p 1 -timeout 10m ./app/tests/contract/...

test-component:
	go test $(BUILD_FLAGS) -tags=component -p 1 -timeout 10m ./app/tests/component/...

test-integration:
	go test $(BUILD_FLAGS) -tags=integration -p 1 \
		./app/internal/server/avatar/adapters/repository/postgres/... \
		./app/internal/server/avatar/adapters/repository/minio/... \
		./app/internal/server/avatar/adapters/repository/rabbitmq/... \
		./app/internal/processor/processing/adapters/repository/postgres/... \
		./app/internal/processor/processing/adapters/repository/minio/...

test-e2e:
	go test $(BUILD_FLAGS) -tags=e2e -p 1 -timeout 15m ./app/tests/e2e/...

test-all: test-unit test-contract test-integration test-component test-e2e

cover:
	go test $(BUILD_FLAGS) -tags=integration -p 1 -coverpkg=./app/... ./app/... -coverprofile=coverage.out && go tool cover -func=coverage.out

cover-unit:
	go test $(BUILD_FLAGS) -coverpkg=./app/... ./app/... -coverprofile=coverage-unit.out && go tool cover -func=coverage-unit.out

cover-integration:
	go test $(BUILD_FLAGS) -tags=integration -p 1 -coverpkg=./app/... \
		./app/internal/server/avatar/adapters/repository/postgres/... \
		./app/internal/server/avatar/adapters/repository/minio/... \
		./app/internal/server/avatar/adapters/repository/rabbitmq/... \
		./app/internal/processor/processing/adapters/repository/postgres/... \
		./app/internal/processor/processing/adapters/repository/minio/... \
		-coverprofile=coverage-integration.out && go tool cover -func=coverage-integration.out

# Local DB: make migrate (override via DATABASE_DSN=...)
DATABASE_DSN ?= postgres://gophprofile:gophprofile@localhost:5432/gophprofile?sslmode=disable

migrate:
	go run ./app/cmd/migrate -d "$(DATABASE_DSN)"

# Regenerate gomock stubs.
generate-mocks:
	bash scripts/generate-mocks.sh

# Regenerate easyjson marshalers.
generate-easyjson:
	bash scripts/generate-easyjson.sh

# Regenerate goverter converters.
generate-goverter:
	bash scripts/generate-goverter.sh

# Static analysis (golangci-lint).
lint:
	golangci-lint run ./app/...
