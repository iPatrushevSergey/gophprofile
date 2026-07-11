#!/usr/bin/env bash
# Configures Vault KV, policy, Kubernetes auth, and ClusterSecretStore for ESO.
# Requires: kubectl, vault CLI (for seed only), Vault pod running in namespace vault.
set -euo pipefail

VAULT_NAMESPACE="${VAULT_NAMESPACE:-vault}"
VAULT_POD="${VAULT_POD:-vault-0}"
ESO_NAMESPACE="${ESO_NAMESPACE:-external-secrets}"
ESO_SA="${ESO_SA:-external-secrets}"
VAULT_ROLE="${VAULT_ROLE:-eso-gophprofile}"
POLICY_NAME="${POLICY_NAME:-gophprofile-read}"
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"

vault_exec() {
  kubectl exec -n "${VAULT_NAMESPACE}" "${VAULT_POD}" -- vault "$@"
}

echo "==> Enable KV v2 at secret/"
vault_exec secrets enable -path=secret kv-v2 2>/dev/null || true

echo "==> Apply policy ${POLICY_NAME}"
kubectl cp "${ROOT_DIR}/deploy/vault/bootstrap/policies.hcl" \
  "${VAULT_NAMESPACE}/${VAULT_POD}:/tmp/policies.hcl"
vault_exec policy write "${POLICY_NAME}" /tmp/policies.hcl

echo "==> Enable Kubernetes auth"
vault_exec auth enable kubernetes 2>/dev/null || true

echo "==> Configure Kubernetes auth"
KUBE_HOST="$(kubectl config view --minify -o jsonpath='{.clusters[0].cluster.server}')"
KUBE_CA="$(kubectl config view --raw --minify -o jsonpath='{.clusters[0].cluster.certificate-authority-data}' | base64 -d)"
REVIEWER_JWT="$(kubectl create token "${ESO_SA}" -n "${ESO_NAMESPACE}" --duration=24h)"

vault_exec write auth/kubernetes/config \
  kubernetes_host="${KUBE_HOST}" \
  kubernetes_ca_cert="${KUBE_CA}" \
  token_reviewer_jwt="${REVIEWER_JWT}"

echo "==> Create role ${VAULT_ROLE} for ESO service account"
vault_exec write auth/kubernetes/role/"${VAULT_ROLE}" \
  bound_service_account_names="${ESO_SA}" \
  bound_service_account_namespaces="${ESO_NAMESPACE}" \
  policies="${POLICY_NAME}" \
  ttl=24h

echo "==> Apply ClusterSecretStore"
kubectl apply -f "${ROOT_DIR}/deploy/eso/cluster-secret-store.yaml"

echo "==> Done. Run: make vault-seed && make helm-install-prod"
