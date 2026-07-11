.PHONY: build build-server build-processor build-gencerts build-all \
	build-linux-amd64 build-windows-amd64 build-darwin-amd64 \
	certs run-server run-processor run-server-prod run-processor-prod \
	docker-up docker-down docker-ps docker-migrate observability-smoke \
	k8s-build-images helm-sync-configs \
	helm-template-dev helm-template-prod helm-install-dev helm-install-prod helm-upgrade-dev helm-upgrade-prod helm-uninstall \
	vault-install vault-uninstall vault-port-forward vault-bootstrap vault-seed \
	eso-install eso-uninstall prod-secrets-stack \
	test test-unit test-contract test-component test-integration test-e2e test-all \
	cover cover-unit cover-integration migrate generate-mocks \
	generate-easyjson generate-goverter lint

APP_DIR := app
BIN_DIR := $(APP_DIR)/bin
DIST_DIR := $(BIN_DIR)/dist
HELM_CHART := deploy/helm/gophprofile
HELM_CONFIGS := $(HELM_CHART)/configs
HELM_VALUES_DEV := $(HELM_CHART)/values-dev.yaml
HELM_VALUES_PROD := $(HELM_CHART)/values-prod.yaml
HELM_RELEASE ?= gophprofile
HELM_NAMESPACE ?= gophprofile
VAULT_NAMESPACE ?= vault
VAULT_RELEASE ?= vault
ESO_NAMESPACE ?= external-secrets
ESO_RELEASE ?= external-secrets
VAULT_VALUES := deploy/vault/values.yaml
ESO_VALUES := deploy/eso/values.yaml

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

k8s-build-images:
	docker build -f deploy/docker/Dockerfile.server -t gophprofile-server:latest .
	docker build -f deploy/docker/Dockerfile.processor -t gophprofile-processor:latest .
	docker build -f deploy/docker/Dockerfile.migrate -t gophprofile-migrate:latest .

# Sync deploy/rabbitmq and deploy/observability into the Helm chart (required by .Files.Get).
helm-sync-configs:
	rm -rf $(HELM_CONFIGS)/rabbitmq $(HELM_CONFIGS)/observability
	mkdir -p $(HELM_CONFIGS)/rabbitmq $(HELM_CONFIGS)/observability
	cp -r deploy/rabbitmq/. $(HELM_CONFIGS)/rabbitmq/
	cp -r deploy/observability/. $(HELM_CONFIGS)/observability/

helm-template-dev: helm-sync-configs
	helm template $(HELM_RELEASE) $(HELM_CHART) -f $(HELM_VALUES_DEV)

helm-template-prod: helm-sync-configs
	helm template $(HELM_RELEASE) $(HELM_CHART) -f $(HELM_VALUES_PROD)

helm-install-dev: k8s-build-images helm-sync-configs
	helm install $(HELM_RELEASE) $(HELM_CHART) \
		-n $(HELM_NAMESPACE) --create-namespace \
		-f $(HELM_VALUES_DEV)

helm-install-prod: helm-sync-configs
	helm install $(HELM_RELEASE) $(HELM_CHART) \
		-n $(HELM_NAMESPACE) --create-namespace \
		-f $(HELM_VALUES_PROD)

helm-upgrade-dev: helm-sync-configs
	helm upgrade $(HELM_RELEASE) $(HELM_CHART) \
		-n $(HELM_NAMESPACE) \
		-f $(HELM_VALUES_DEV)

helm-upgrade-prod: helm-sync-configs
	helm upgrade $(HELM_RELEASE) $(HELM_CHART) \
		-n $(HELM_NAMESPACE) \
		-f $(HELM_VALUES_PROD)

helm-uninstall:
	helm uninstall $(HELM_RELEASE) -n $(HELM_NAMESPACE)

# Vault + ESO secrets stack (prod).
vault-install:
	helm repo add hashicorp https://helm.releases.hashicorp.com 2>/dev/null || true
	helm upgrade --install $(VAULT_RELEASE) hashicorp/vault \
		-n $(VAULT_NAMESPACE) --create-namespace \
		-f $(VAULT_VALUES)

vault-uninstall:
	helm uninstall $(VAULT_RELEASE) -n $(VAULT_NAMESPACE)

vault-port-forward:
	kubectl port-forward -n $(VAULT_NAMESPACE) svc/$(VAULT_RELEASE) 8200:8200

eso-install:
	helm repo add external-secrets https://charts.external-secrets.io 2>/dev/null || true
	helm upgrade --install $(ESO_RELEASE) external-secrets/external-secrets \
		-n $(ESO_NAMESPACE) --create-namespace \
		-f $(ESO_VALUES)

eso-uninstall:
	helm uninstall $(ESO_RELEASE) -n $(ESO_NAMESPACE)

vault-bootstrap:
	bash deploy/vault/bootstrap/vault-bootstrap.sh

vault-seed:
	bash deploy/vault/bootstrap/vault-seed.sh

prod-secrets-stack: vault-install eso-install vault-bootstrap
	@echo "Next: copy deploy/vault/bootstrap/.env.vault.example to .env.vault, run 'make vault-port-forward' in another terminal, then 'make vault-seed'"

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
