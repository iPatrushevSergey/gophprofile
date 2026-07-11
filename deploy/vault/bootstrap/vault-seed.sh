#!/usr/bin/env bash
# Writes application secrets to Vault KV. Values from .env.vault.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
ENV_FILE="${ROOT_DIR}/.env.vault"

if [[ ! -f "${ENV_FILE}" ]]; then
  echo "Missing ${ENV_FILE}. Copy deploy/vault/bootstrap/.env.vault.example and fill values." >&2
  exit 1
fi

# shellcheck disable=SC1090
source "${ENV_FILE}"

: "${VAULT_ADDR:?VAULT_ADDR required in .env.vault}"
: "${VAULT_TOKEN:?VAULT_TOKEN required in .env.vault}"
: "${POSTGRES_PASSWORD:?POSTGRES_PASSWORD required}"
: "${MINIO_ROOT_USER:?MINIO_ROOT_USER required}"
: "${MINIO_ROOT_PASSWORD:?MINIO_ROOT_PASSWORD required}"
: "${RABBITMQ_PASSWORD:?RABBITMQ_PASSWORD required}"
: "${GRAFANA_ADMIN_PASSWORD:?GRAFANA_ADMIN_PASSWORD required}"

echo "==> Port-forward Vault if VAULT_ADDR is localhost (make vault-port-forward in another terminal)"

vault kv put secret/gophprofile/postgres \
  password="${POSTGRES_PASSWORD}"

vault kv put secret/gophprofile/minio \
  rootUser="${MINIO_ROOT_USER}" \
  rootPassword="${MINIO_ROOT_PASSWORD}"

vault kv put secret/gophprofile/rabbitmq \
  password="${RABBITMQ_PASSWORD}"

vault kv put secret/gophprofile/grafana \
  adminPassword="${GRAFANA_ADMIN_PASSWORD}"

echo "==> Secrets written to Vault KV (mount secret/, paths gophprofile/*)"
