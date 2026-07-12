#!/usr/bin/env bash
# Writes application secrets to Vault KV via kubectl exec into the Vault pod.
set -euo pipefail

# Bootstrap defaults.
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
ENV_FILE="${ROOT_DIR}/.env.vault"

VAULT_NAMESPACE="${VAULT_NAMESPACE:-vault}"
VAULT_POD="${VAULT_POD:-vault-0}"

# Load secrets from .env.vault.
if [[ ! -f "${ENV_FILE}" ]]; then
  echo "Missing ${ENV_FILE}. Copy deploy/vault/bootstrap/.env.vault.example and fill values." >&2
  exit 1
fi

# shellcheck disable=SC1090
source "${ENV_FILE}"

# Validate required variables.
: "${VAULT_TOKEN:?VAULT_TOKEN required in .env.vault}"
: "${POSTGRES_PASSWORD:?POSTGRES_PASSWORD required}"
: "${MINIO_ROOT_USER:?MINIO_ROOT_USER required}"
: "${MINIO_ROOT_PASSWORD:?MINIO_ROOT_PASSWORD required}"
: "${RABBITMQ_PASSWORD:?RABBITMQ_PASSWORD required}"
: "${GRAFANA_ADMIN_PASSWORD:?GRAFANA_ADMIN_PASSWORD required}"

# Run vault CLI inside the Vault pod with root token.
vault_exec() {
  kubectl exec -n "${VAULT_NAMESPACE}" "${VAULT_POD}" -- \
    env VAULT_TOKEN="${VAULT_TOKEN}" vault "$@"
}

# Write application secrets to Vault KV.
echo "==> Write secrets to Vault KV via kubectl exec (${VAULT_NAMESPACE}/${VAULT_POD})"

vault_exec kv put secret/gophprofile/postgres \
  password="${POSTGRES_PASSWORD}"

vault_exec kv put secret/gophprofile/minio \
  rootUser="${MINIO_ROOT_USER}" \
  rootPassword="${MINIO_ROOT_PASSWORD}"

vault_exec kv put secret/gophprofile/rabbitmq \
  password="${RABBITMQ_PASSWORD}"

vault_exec kv put secret/gophprofile/grafana \
  adminPassword="${GRAFANA_ADMIN_PASSWORD}"

echo "==> Secrets written to Vault KV (mount secret/, paths gophprofile/*)"
