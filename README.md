# GophProfile

Сервис управления аватарками на Go: загрузка, хранение, асинхронная обработка (миниатюры) и раздача через REST API.

## Архитектура

```mermaid
flowchart TB
  Client[Client]

  subgraph k8s [Kubernetes]
    Ingress[Ingress]

    subgraph app [Application]
      Server[server]
      Processor[processor]
    end

    subgraph data [Data]
      Postgres[(postgres)]
      MinIO[(minio)]
      RabbitMQ[rabbitmq]
    end

    Migrate[migrate]

    subgraph obs [Observability]
      OTEL[otel-collector]
      Jaeger[jaeger]
      Prometheus[prometheus]
      Loki[loki]
      Grafana[grafana]
      Alertmanager[alertmanager]
      AlertWebhook[alert-webhook]
      NodeExporter[node-exporter]
    end

    Vault[Vault]
    ESO[ESO]
  end

  Client --> Ingress
  Ingress --> Server
  Server --> Postgres
  Server --> MinIO
  Server --> RabbitMQ
  RabbitMQ --> Processor
  Processor --> Postgres
  Processor --> MinIO
  Server --> OTEL
  Processor --> OTEL
  OTEL --> Jaeger
  OTEL --> Loki
  OTEL --> Prometheus
  Grafana --> Prometheus
  Grafana --> Loki
  Grafana --> Jaeger
  Prometheus --> Alertmanager
  Loki --> Alertmanager
  Alertmanager --> AlertWebhook
  Vault --> ESO
  ESO -.-> Server
  ESO -.-> Processor
```

Два сервиса — **`cmd/server`** (HTTP API) и **`cmd/processor`** (worker обработки изображений). Написаны по **Clean Architecture**: `presentation` → `application` → `domain` → `adapters`.

Общая инфраструктура вынесена в **`app/internal/pkg`**, бизнес-модули не дублируют инфраструктурный код.

## Функциональность

**Server:** REST API для загрузки, получения, удаления и списка аватарок; health check; transactional outbox для публикации событий в RabbitMQ; фоновые задачи (GC просроченных upload, drain outbox); статический веб-UI на `/web`.

**Processor:** потребление событий `avatar.uploaded` / `avatar.deleted`; создание миниатюр **100×100** и **300×300**; retry через TTL-очередь; DLQ для poison messages.

**Сквозной сценарий:** клиент загружает JPEG/PNG/WebP (до 10 MB) → server сохраняет оригинал в MinIO и метаданные в Postgres → outbox публикует событие → processor ресайзит и сохраняет thumbnails → клиент запрашивает `?size=100x100` или `?size=300x300`.

---

## Что понадобится

| Компонент | Назначение |
|-----------|------------|
| **Go 1.26+** | локальная сборка server/processor/migrate |
| **Docker + Docker Compose** | инфраструктура; contract/component/integration/e2e-тесты |
| **golangci-lint** | статический анализ (`make lint`) |
| **make** | цели сборки, тестов, docker (опционально — команды можно запускать напрямую) |

Приоритет настроек: **флаги CLI → переменные окружения `GOPHPROFILE_*` → yaml → defaults**. Примеры env — [`.env.example`](.env.example) (Docker Compose) и [`app/.env.example`](app/.env.example) (локальный server/processor).

---

## Быстрый старт (Docker Compose)

Пошаговый сценарий: env → поднять весь стек → проверить health → загрузить аватарку → дождаться миниатюр → скачать изображение.

### Шаг 1. Клонирование и env

```bash
git clone <repo-url> gophprofile
cd gophprofile
cp .env.example .env
```

Файл `.env` не коммитится. Значения по умолчанию подходят для локальной разработки.

### Шаг 2. Запуск стека

```bash
make docker-up
# эквивалент: docker compose up -d --build
```

Compose поднимает сервисы в порядке зависимостей:

1. `postgres`, `minio`, `rabbitmq` — инфраструктура (healthcheck)
2. `jaeger`, `otel-collector`, `prometheus`, `alertmanager`, `alert-webhook`, `node-exporter`, `loki`, `grafana` — observability
3. `minio-init` — создание бакета `gophprofile`
4. `migrate` — накат SQL из `migrations/gophprofile/`
5. `server` — HTTP API на порту **8080**
6. `processor` — worker обработки; метрики через OTLP → collector → Prometheus `:8889`

Проверка:

```bash
make docker-ps
# или: docker compose ps
make observability-smoke   # E2E: targets, upload, metrics, logs, traces
```

Все сервисы `server` и `processor` должны быть в статусе `running`. Контейнер `migrate` — `exited (0)`.

### Шаг 3. Health check

```bash
curl -s http://127.0.0.1:8080/health
```

Ожидаемый ответ (все компоненты доступны):

```json
{
  "status": "ok",
  "database": "ok",
  "storage": "ok",
  "broker": "ok"
}
```

### Шаг 4. Загрузка аватарки

Подготовьте тестовое изображение (JPEG/PNG/WebP, ≤ 10 MB) или создайте минимальный PNG:

```bash
printf '\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x01\x00\x00\x00\x01\x08\x02\x00\x00\x00\x90wS\xde\x00\x00\x00\x0cIDATx\x9cc\xf8\x0f\x00\x00\x01\x01\x00\x05\x18\xd8N\x00\x00\x00\x00IEND\xaeB`\x82' > /tmp/test.png
```

Загрузка через API:

```bash
curl -s -X POST http://127.0.0.1:8080/api/v1/avatars \
  -H "X-User-ID: alice" \
  -F "file=@/tmp/test.png;type=image/png"
```

Пример ответа `201 Created`:

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "alice",
  "url": "/api/v1/avatars/550e8400-e29b-41d4-a716-446655440000",
  "status": "pending",
  "created_at": "2026-06-21T12:00:00Z"
}
```

Сохраните `id`:

```bash
AVATAR_ID="<id из ответа>"
```

> **Альтернатива:** откройте http://127.0.0.1:8080/web — форма загрузки и галерея без `curl`.

### Шаг 5. Ожидание обработки

Processor создаёт миниатюры асинхронно. Проверка метаданных:

```bash
curl -s "http://127.0.0.1:8080/api/v1/avatars/${AVATAR_ID}/metadata"
```

Когда обработка завершена, в массиве `thumbnails` появятся записи с `size` `100x100` и `300x300`. Повторяйте запрос или подождите ~5–10 секунд после upload.

### Шаг 6. Получение изображения

```bash
# оригинал
curl -s -o /tmp/avatar-original.png \
  "http://127.0.0.1:8080/api/v1/avatars/${AVATAR_ID}"

# миниатюра 100×100
curl -s -o /tmp/avatar-100.png \
  "http://127.0.0.1:8080/api/v1/avatars/${AVATAR_ID}?size=100x100"

# миниатюра 300×300
curl -s -o /tmp/avatar-300.png \
  "http://127.0.0.1:8080/api/v1/avatars/${AVATAR_ID}?size=300x300"
```

### Шаг 7. Удаление

```bash
curl -s -o /dev/null -w "%{http_code}\n" \
  -X DELETE "http://127.0.0.1:8080/api/v1/avatars/${AVATAR_ID}" \
  -H "X-User-ID: alice"
# ожидается: 204
```

### Остановка стека

```bash
make docker-down
# эквивалент: docker compose down
```

Данные Postgres, MinIO и RabbitMQ сохраняются в docker volumes. Полная очистка:

```bash
docker compose down -v
```

### Веб-интерфейсы инфраструктуры

| Сервис | URL | Логин / пароль (из `.env`) |
|--------|-----|----------------------------|
| RabbitMQ Management | http://127.0.0.1:15672 | `RABBITMQ_DEFAULT_USER` / `RABBITMQ_DEFAULT_PASS` |
| MinIO Console | http://127.0.0.1:9001 | `MINIO_ROOT_USER` / `MINIO_ROOT_PASSWORD` |

---

## Быстрый старт

Пошаговый сценарий: сборка → инфраструктура в Docker → миграции → env → server + processor на хосте → проверка API.

### Шаг 1. Сборка

Из корня репозитория:

```bash
make build
# бинарники: app/bin/server, app/bin/processor
# version/date вшиваются через -ldflags (см. Makefile)
```

Без `make`:

```bash
go build -tags=go_json -o app/bin/server ./app/cmd/server
go build -tags=go_json -o app/bin/processor ./app/cmd/processor
```

Проверка:

```bash
./app/bin/server --version
./app/bin/processor --version
```

### Шаг 2. Инфраструктура (Postgres, MinIO, RabbitMQ)

```bash
cp .env.example .env
docker compose up -d postgres minio minio-init rabbitmq
```

Дождаться готовности:

```bash
docker compose ps
# postgres, rabbitmq — healthy
# minio-init — exited (0)
```

> RabbitMQ поднимается с топологией и пользователем из [`deploy/rabbitmq/definitions.json`](deploy/rabbitmq/definitions.json) (`management.load_definitions` в [`deploy/rabbitmq/rabbitmq.conf`](deploy/rabbitmq/rabbitmq.conf)). При импорте definitions env `RABBITMQ_DEFAULT_USER/PASS` **не создаёт** пользователя в брокере — они только подставляются в `GOPHPROFILE_BROKER_URL` для server/processor. Для Docker-сети не используйте `guest` (удалённые подключения запрещены); в `.env.example` — `gophprofile` / `gophprofile`.

### Шаг 3. Миграции PostgreSQL

```bash
export DATABASE_DSN='postgres://gophprofile:gophprofile@localhost:5432/gophprofile?sslmode=disable'
make migrate
```

или:

```bash
go run ./app/cmd/migrate -d "$DATABASE_DSN"
```

Миграции лежат в [`migrations/gophprofile/`](migrations/gophprofile/).

### Шаг 4. Переменные окружения

Создайте `app/.env` (можно скопировать из [`app/.env.example`](app/.env.example) и раскомментировать строки):

```bash
export GOPHPROFILE_DATABASE_URI='postgres://gophprofile:gophprofile@localhost:5432/gophprofile?sslmode=disable'
export GOPHPROFILE_SERVER_ADDRESS=127.0.0.1:8080
export GOPHPROFILE_MINIO_ENDPOINT=127.0.0.1:9000
export GOPHPROFILE_MINIO_ACCESS_KEY=minioadmin
export GOPHPROFILE_MINIO_SECRET_KEY=minioadmin
export GOPHPROFILE_MINIO_USE_SSL=false
export GOPHPROFILE_BROKER_URL=amqp://gophprofile:gophprofile@localhost:5672/
```

Server и processor читают `app/.env` или `.env` в корне при старте. Processor использует те же `GOPHPROFILE_DATABASE_URI`, `GOPHPROFILE_MINIO_*`, `GOPHPROFILE_BROKER_URL`.

### Шаг 5. Запуск server

Терминал 1:

```bash
make run-server
# или: ./app/bin/server -c app/configs/server.yaml
```

Сервер слушает `http://127.0.0.1:8080`.

### Шаг 6. Запуск processor

Терминал 2:

```bash
make run-processor
# или: ./app/bin/processor -c app/configs/processor.yaml
```

### Шаг 7. Проверка

Те же команды из раздела «Быстрый старт (Docker Compose)», шаги 3–7, или веб-UI: http://127.0.0.1:8080/web .

---

## Веб-интерфейс

Статический SPA на **`GET /web`** ([`web/static/index.html`](web/static/index.html)):

| Экран | URL | Описание |
|-------|-----|----------|
| Загрузка | http://127.0.0.1:8080/web | форма: User ID + файл → `POST /api/v1/avatars` |
| Галерея | http://127.0.0.1:8080/web#gallery/{user_id} | список аватарок пользователя, просмотр и удаление |

Заголовок **`X-User-ID`** выставляется из поля User ID. После успешной загрузки можно перейти в галерею по ссылке на странице.

---

## REST API

OpenAPI 3.0: [`api/openapi.yaml`](api/openapi.yaml).

Базовый префикс: `/api/v1`. Аутентификация для защищённых маршрутов — заголовок **`X-User-ID`**.

| Метод | Путь | Auth | Описание |
|-------|------|------|----------|
| `GET` | `/health` | нет | health check (database, storage, broker) |
| `POST` | `/api/v1/avatars` | `X-User-ID` | загрузка (`multipart/form-data`, поле **`file`**, max 10 MB) |
| `GET` | `/api/v1/avatars/{avatar_id}` | нет | получение изображения |
| `GET` | `/api/v1/avatars/{avatar_id}/metadata` | нет | метаданные и список thumbnails |
| `GET` | `/api/v1/users/{user_id}/avatars` | нет | список аватарок пользователя |
| `DELETE` | `/api/v1/avatars/{avatar_id}` | `X-User-ID` | удаление (только свой avatar) |

**Query-параметры `GET /api/v1/avatars/{avatar_id}`:**

| Параметр | Значения | По умолчанию |
|----------|----------|--------------|
| `size` | `original`, `100x100`, `300x300` | `original` |
| `format` | `jpeg`, `png`, `webp` | формат оригинала / thumbnail |

Поддерживаемые форматы upload: **JPEG, PNG, WebP**.

Примеры:

```bash
# upload
curl -X POST http://127.0.0.1:8080/api/v1/avatars \
  -H "X-User-ID: bob" \
  -F "file=@avatar.png;type=image/png"

# metadata
curl http://127.0.0.1:8080/api/v1/avatars/{avatar_id}/metadata

# список аватарок пользователя
curl http://127.0.0.1:8080/api/v1/users/bob/avatars

# thumbnail
curl -o thumb.png "http://127.0.0.1:8080/api/v1/avatars/{avatar_id}?size=100x100"

# delete
curl -X DELETE http://127.0.0.1:8080/api/v1/avatars/{avatar_id} -H "X-User-ID: bob"
```

---

## Асинхронная обработка

```
Client → POST /avatars
           ↓
        Server: MinIO (original) + Postgres (metadata, status=pending)
           ↓
        Outbox → RabbitMQ exchange "avatars"
           ↓
        Queue "avatar-processing"
           ↓
        Processor: resize → MinIO (100x100, 300x300) → Postgres
```

При ошибках processor использует retry-очередь с TTL и dead-letter queue. Параметры — [`app/configs/processor.yaml`](app/configs/processor.yaml), топология и пользователь брокера — [`deploy/rabbitmq/definitions.json`](deploy/rabbitmq/definitions.json) + [`deploy/rabbitmq/rabbitmq.conf`](deploy/rabbitmq/rabbitmq.conf).

---

## Production (HTTPS + TLS)

### 1. Сертификаты (dev CA)

```bash
make certs
# certs/ca.pem, certs/server.crt, certs/server.key
```

### 2. Env

```bash
export GOPHPROFILE_DATABASE_URI='postgres://gophprofile:gophprofile@localhost:5432/gophprofile?sslmode=disable'
export GOPHPROFILE_MINIO_ENDPOINT=localhost:9000
export GOPHPROFILE_MINIO_ACCESS_KEY=minioadmin
export GOPHPROFILE_MINIO_SECRET_KEY=minioadmin
export GOPHPROFILE_MINIO_USE_SSL=false
export GOPHPROFILE_BROKER_URL=amqp://gophprofile:gophprofile@localhost:5672/
export GOPHPROFILE_SERVER_CERT_FILE=certs/server.crt
export GOPHPROFILE_SERVER_KEY_FILE=certs/server.key
```

Если MinIO в Docker на HTTP, а в `server.prod.yaml` стоит `minio.use_ssl: true`, переопределите `GOPHPROFILE_MINIO_USE_SSL=false`.

### 3. Запуск

```bash
make run-server-prod   # https://0.0.0.0:8443
make run-processor-prod
```

| | Dev | Prod |
|--|-----|------|
| Конфиги | `server.yaml` / `processor.yaml` | `server.prod.yaml` / `processor.prod.yaml` |
| URL server | `http://127.0.0.1:8080` | `https://127.0.0.1:8443` |
| TLS | нет | `server.crt` + `server.key` |
| MinIO `use_ssl` | `false` | `true` в prod yaml (переопределите env, если MinIO на HTTP) |

---

## Глобальные флаги server

| Флаг | Короткий | Env | Описание |
|------|----------|-----|----------|
| `--config` | `-c` | `CONFIG` | путь к yaml |
| `--address` | `-a` | `GOPHPROFILE_SERVER_ADDRESS` | listen address |
| `--database-uri` | `-d` | `GOPHPROFILE_DATABASE_URI` | Postgres DSN (**обязателен**) |
| `--log-level` | `-l` | `GOPHPROFILE_LOGGER_LEVEL` | уровень логов |
| `--shutdown-timeout` | | `GOPHPROFILE_SERVER_SHUTDOWN_TIMEOUT` | graceful shutdown |
| `--tls-cert` | | `GOPHPROFILE_SERVER_CERT_FILE` | TLS cert (prod) |
| `--tls-key` | | `GOPHPROFILE_SERVER_KEY_FILE` | TLS key (prod) |

MinIO и broker — через env / yaml, см. [`app/configs/server.yaml`](app/configs/server.yaml):

| Поле yaml | Env | Описание |
|-----------|-----|----------|
| `minio.endpoint` | `GOPHPROFILE_MINIO_ENDPOINT` | адрес MinIO |
| `minio.access_key` | `GOPHPROFILE_MINIO_ACCESS_KEY` | ключ |
| `minio.secret_key` | `GOPHPROFILE_MINIO_SECRET_KEY` | секрет |
| `minio.bucket` | `GOPHPROFILE_MINIO_BUCKET` | бакет (default `gophprofile`) |
| `minio.use_ssl` | `GOPHPROFILE_MINIO_USE_SSL` | HTTPS к MinIO |
| `broker.url` | `GOPHPROFILE_BROKER_URL` | AMQP URL |

Postgres, MinIO и RabbitMQ **обязательны** для старта server.

---

## Глобальные флаги processor

| Флаг | Короткий | Env | Описание |
|------|----------|-----|----------|
| `--config` | `-c` | `CONFIG` | путь к yaml |
| `--database-uri` | `-d` | `GOPHPROFILE_DATABASE_URI` | Postgres DSN (**обязателен**) |
| `--log-level` | `-l` | `GOPHPROFILE_LOGGER_LEVEL` | уровень логов |
| `--shutdown-timeout` | | `GOPHPROFILE_WORKER_SHUTDOWN_TIMEOUT` | graceful shutdown worker |

Очереди, retry, DLQ — в [`app/configs/processor.yaml`](app/configs/processor.yaml) (`GOPHPROFILE_BROKER_*` через env).

Postgres, MinIO и RabbitMQ **обязательны** для старта processor.

---

## Миграции

| Способ | Команда | Когда |
|--------|---------|-------|
| Локально | `make migrate` | dev на хосте, нужен `DATABASE_DSN` |
| Docker one-shot | `make docker-migrate` | ручной перезапуск после новых SQL |
| Compose (авто) | `make docker-up` | первый подъём стека (`migrate` → `server`/`processor`) |

SQL-файлы: [`migrations/gophprofile/`](migrations/gophprofile/).

> **После новых SQL-файлов** контейнер `migrate` при `docker compose up` **не перезапускается** автоматически, если уже завершился успешно. Выполните `make docker-migrate` или `make migrate` вручную.

---

## Трассировка

Для ручной телеметрии используем стабильный instrumentation key:

```text
<module>.<layer>.<component>.<operation>
```

Ключ пишется в атрибут `gophprofile.telemetry.key` и нужен как единая точка адресации для span-ов и будущего selective disable.

### Правила

- `module` — бизнес-модуль: `avatar`, `processing`
- `layer` — слой CA: `presentation`, `application`, `domain`, `adapter`
- `component` — имя Go-файла без `.go` (`outbox_publisher.go` → `outbox_publisher`)
- `operation` — имя **Go-метода**, в котором создаётся span, в `snake_case` (`ReceiveMessages` → `receive_messages`, `PublishAvatarUploaded` → `publish_avatar_uploaded`);
  - исключение: если span в общем entrypoint (`Execute`, `Run`, `Handle`) — имя use case или логического действия (`publish_pending_outbox_events`, `process_uploaded_avatar`).

Где смотреть при добавлении span: `module` и `layer` — из пакета/пути; `component` — имя файла без `.go`; `operation` — метод, в теле которого вызывается `tracer.Start`. Существующие ключи — `grep` по `Key:` в репозитории или примеры ниже.

## Метрики

Метрики экспортируются через OTLP в **otel-collector**, который отдаёт их Prometheus на `:8889`. Имена series совпадают с прежними (`avatars_uploads_total`, `http_server_requests_total`, …); label `job` берётся из `service.name` (`gophprofile-server`, `gophprofile-processor`).

| Сервис | Что пишется |
|--------|-------------|
| **server** | HTTP RED, upload/delete, `avatars_storage_bytes`, `outbox_pending_messages`, pool (`db_pool_connections_*`) |
| **processor** | processing (`avatars_processing_*`), pool (`db_pool_connections_*`) |

В PromQL указывайте `job="gophprofile-server"` или `job="gophprofile-processor"`, чтобы брать series нужного процесса.

## Docker Compose

`make docker-up` поднимает полный стек наблюдаемости вместе с приложением. Конфиги — [`deploy/observability/`](deploy/observability/):

| UI | URL | Назначение |
|----|-----|------------|
| **Grafana** | http://127.0.0.1:3000 | дашборды (метрики + логи + трейсы) |
| **Jaeger** | http://127.0.0.1:16686 | трассировка |
| **Prometheus** | http://127.0.0.1:9092 | сбор метрик, алертинг |
| **Alertmanager** | http://127.0.0.1:9093 | алертинг |
| **Alert webhook** | http://127.0.0.1:5001 | приём webhook от Alertmanager (логи в `docker compose logs alert-webhook`) |
| **RabbitMQ metrics** | http://127.0.0.1:15692/metrics | метрики RabbitMQ |
| **Node exporter** | http://127.0.0.1:9100/metrics | метрики хоста Docker |

**Проверка стека:**

```bash
make docker-up
make observability-smoke
```

---

## Kubernetes (Rancher Desktop)

Helm chart: [`deploy/helm/gophprofile/`](deploy/helm/gophprofile/).

### Что понадобится

| Компонент | Назначение |
|-----------|------------|
| **Rancher Desktop** | Kubernetes + Traefik ingress |
| **kubectl, helm, make, docker** | в PATH |

---

### Dev — пошагово

Секреты заданы в `values-dev.yaml`, Vault не нужен.

```bash
# 1. Установка
make helm-install-dev

# 2. Дождаться готовности pod'ов
kubectl get pods -n gophprofile -w

# 3. Smoke test
make k8s-smoke-dev
```

Ручная проверка:

```bash
curl -H "Host: gophprofile.local" http://127.0.0.1/health
```

Обновление / удаление:

```bash
make helm-upgrade-dev      # после изменений в коде/chart
make helm-uninstall        # удалить release (PVC остаются)
```

---

### Prod-local — пошагово

Vault + External Secrets Operator + HPA + Traefik (без TLS).

**Стек секретов** (один раз, или после `make vault-uninstall`):

```bash
# 1. Vault + ESO + bootstrap (KV, policy, K8s auth, ClusterSecretStore)
make prod-secrets-stack

# 2. Дождаться готовности
kubectl wait -n vault --for=condition=ready pod/vault-0 --timeout=120s
kubectl wait -n external-secrets --for=condition=ready pod \
  -l app.kubernetes.io/name=external-secrets --timeout=120s

# 3. Пароли для Vault KV
cp deploy/vault/bootstrap/.env.vault.example .env.vault
# значения по умолчанию совпадают с dev; для prod-local менять не обязательно

# 4. Записать секреты в Vault
make vault-seed

# 5. Проверить ESO
kubectl get clustersecretstore vault-backend
```

**Приложение:**

```bash
# 6. Если dev release уже стоит — сначала удалить
make helm-uninstall

# 7. Установка prod-local
make helm-install-prod-local

# 8. Дождаться pod'ов и ExternalSecret
kubectl get pods,externalsecret -n gophprofile -w

# 9. Smoke test
make k8s-smoke-prod
```

---

### Переключение dev ↔ prod-local

```bash
make helm-uninstall
# при смене паролей RabbitMQ/Postgres — удалить PVC:
kubectl delete pvc -n gophprofile --all
make helm-install-dev          # или make helm-install-prod-local
```

---

## Тесты

Пирамида: unit → contract → integration → component → e2e. Для contract, component, integration и e2e нужен **Docker** (testcontainers: Postgres + MinIO + RabbitMQ).

```bash
make test              # unit
make test-contract     # wire-контракт server - processor (JSON событий)
make test-integration  # репозитории Postgres / MinIO / RabbitMQ
make test-component    # HTTP-сценарии server (upload + list)
make test-e2e          # полные сценарии API (upload, metadata, auth)
make test-all          # все цели выше
```

Покрытие:

```bash
make cover-unit         # unit
make cover-integration  # integration
make cover              # unit + integration
```

Статический анализ ([golangci-lint](https://golangci-lint.run/)):

```bash
# установка
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

make lint
```

Конфигурация — [`.golangci.yml`](.golangci.yml) (build tag `go_json`, исключены generated/mocks).

---

## Генерация кода

После изменения HTTP dto (struct, json-теги):

```bash
make generate-easyjson
```

После изменения converter-интерфейсов в `adapters/.../converter/`:

```bash
make generate-goverter
```

После изменения port-интерфейсов:

```bash
make generate-mocks
```

---

## Make-цели

```bash
make build              # app/bin/server + processor
make build-all          # кросс-сборка linux/windows/darwin → app/bin/dist/
make certs              # dev TLS (certs/)
make run-server         # HTTP dev server
make run-processor      # dev worker
make run-server-prod    # HTTPS prod server
make run-processor-prod # prod worker
make migrate            # Postgres миграции (DATABASE_DSN)
make docker-up          # полный стек в Docker
make docker-down        # остановить compose
make docker-ps          # статус контейнеров
make docker-migrate     # migrate в Docker (ручной запуск)
make observability-smoke # E2E smoke Docker observability stack
make k8s-build-images   # собрать server/processor/migrate образы
make k8s-import-images  # build + import образов в Rancher Desktop k3s
make helm-install-dev   # dev install (build + helm)
make helm-upgrade-dev   # dev upgrade
make helm-install-prod-local  # prod-local install на RD (Vault + ESO)
make helm-upgrade-prod-local  # prod-local upgrade
make helm-install-prod  # prod install (CI, образы из registry)
make helm-upgrade-prod  # prod upgrade
make helm-uninstall     # удалить Helm release
make k8s-smoke-dev      # E2E smoke dev
make k8s-smoke-prod     # E2E smoke prod-local
make prod-secrets-stack # Vault + ESO + bootstrap
make vault-install      # установить Vault
make vault-uninstall    # удалить Vault
make vault-bootstrap    # KV, policy, K8s auth, ClusterSecretStore
make vault-seed         # записать секреты в Vault из .env.vault
make eso-install        # установить External Secrets Operator
make eso-uninstall      # удалить ESO
make test               # unit-тесты (см. раздел «Тесты»)
make test-contract      # контракт server ↔ processor
make test-integration   # postgres/minio/rabbitmq integration
make test-component     # component-тесты server
make test-e2e           # e2e API (testcontainers)
make test-all           # всё вместе
make lint               # golangci-lint (см. раздел «Тесты»)
make cover-unit         # покрытие unit
make cover              # покрытие unit + integration
make generate-easyjson  # easyjson для HTTP dto
make generate-goverter  # goverter для adapters/.../converter
make generate-mocks     # gomock для application/port
```
