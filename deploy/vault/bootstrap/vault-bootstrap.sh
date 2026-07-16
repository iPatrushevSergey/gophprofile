#!/usr/bin/env bash
# Configures Vault KV, policy, Kubernetes auth, and ClusterSecretStore for ESO.
# Requires: kubectl, Vault pod running in namespace vault.
set -euo pipefail

# Bootstrap defaults.
VAULT_NAMESPACE="${VAULT_NAMESPACE:-vault}"
VAULT_POD="${VAULT_POD:-vault-0}"
ESO_NAMESPACE="${ESO_NAMESPACE:-external-secrets}"
ESO_SA="${ESO_SA:-external-secrets}"
VAULT_ROLE="${VAULT_ROLE:-eso-gophprofile}"
POLICY_NAME="${POLICY_NAME:-gophprofile-read}"
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"

# Run vault CLI inside the Vault pod.
vault_exec() {
  kubectl exec -n "${VAULT_NAMESPACE}" "${VAULT_POD}" -- vault "$@"
}

# Enable KV v2 secret mount.
echo "==> Enable KV v2 at secret/"
vault_exec secrets enable -path=secret kv-v2 2>/dev/null || true

# Apply read-only policy for ESO.
echo "==> Apply policy ${POLICY_NAME}"
kubectl exec -i -n "${VAULT_NAMESPACE}" "${VAULT_POD}" -- \
  vault policy write "${POLICY_NAME}" - \
  < "${ROOT_DIR}/deploy/vault/bootstrap/policies.hcl"

# Enable Kubernetes auth method.
echo "==> Enable Kubernetes auth"
vault_exec auth enable kubernetes 2>/dev/null || true

# Configure Kubernetes auth (host, CA, reviewer JWT).
echo "==> Configure Kubernetes auth"
KUBE_HOST="$(kubectl config view --minify -o jsonpath='{.clusters[0].cluster.server}')"
if [[ "${KUBE_HOST}" =~ ^https?://(127\.0\.0\.1|localhost): ]]; then
  KUBE_HOST="https://kubernetes.default.svc.cluster.local:443"
fi
KUBE_CA="$(kubectl config view --raw --minify -o jsonpath='{.clusters[0].cluster.certificate-authority-data}' | base64 -d)"
REVIEWER_JWT="$(kubectl create token vault -n "${VAULT_NAMESPACE}" --duration=8760h)"

vault_exec write auth/kubernetes/config \
  kubernetes_host="${KUBE_HOST}" \
  kubernetes_ca_cert="${KUBE_CA}" \
  token_reviewer_jwt="${REVIEWER_JWT}"

# Create Vault role bound to ESO service account.
echo "==> Create role ${VAULT_ROLE} for ESO service account"
vault_exec write auth/kubernetes/role/"${VAULT_ROLE}" \
  bound_service_account_names="${ESO_SA}" \
  bound_service_account_namespaces="${ESO_NAMESPACE}" \
  policies="${POLICY_NAME}" \
  ttl=24h

# Apply ClusterSecretStore for ESO.
echo "==> Apply ClusterSecretStore"
kubectl apply -f "${ROOT_DIR}/deploy/eso/cluster-secret-store.yaml"

# Restart ESO to pick up Vault auth config changes.
echo "==> Restart ESO (pick up Vault auth config changes)"
kubectl rollout restart deployment/external-secrets -n "${ESO_NAMESPACE}"
kubectl rollout status deployment/external-secrets -n "${ESO_NAMESPACE}" --timeout=120s

echo "==> Done. Run: make vault-seed && make helm-install-prod"
