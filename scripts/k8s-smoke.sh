#!/usr/bin/env bash
# End-to-end smoke test for GophProfile on Kubernetes (Helm dev or prod).
# Usage:
#   make k8s-smoke-dev    # after make helm-install-dev
#   make k8s-smoke-prod   # after make helm-install-prod-local
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

# Resolve profile and ingress host.
PROFILE="${PROFILE:-dev}"
NAMESPACE="${HELM_NAMESPACE:-gophprofile}"
RELEASE="${HELM_RELEASE:-gophprofile}"
FULLNAME="${RELEASE}"

if [ "$PROFILE" = "dev" ]; then
  INGRESS_HOST="${INGRESS_HOST:-gophprofile.local}"
elif [ "$PROFILE" = "prod" ]; then
  INGRESS_HOST="${INGRESS_HOST:-gophprofile-prod.local}"
else
  echo "PROFILE must be dev or prod, got: ${PROFILE}" >&2
  exit 1
fi

PASS=0
FAIL=0
PF_PIDS=()

log() { printf '[k8s-smoke:%s] %s\n' "$PROFILE" "$*"; }
pass() { PASS=$((PASS + 1)); log "PASS: $*"; }
fail() { FAIL=$((FAIL + 1)); log "FAIL: $*"; }

# Stop port-forwards on exit.
cleanup() {
  for pid in "${PF_PIDS[@]:-}"; do
    kill "$pid" 2>/dev/null || true
  done
}
trap cleanup EXIT

# Retry helper for eventually-ready checks.
retry() {
  local attempts="$1"
  local delay="$2"
  shift 2
  local i=1
  while [ "$i" -le "$attempts" ]; do
    if "$@"; then
      return 0
    fi
    sleep "$delay"
    i=$((i + 1))
  done
  return 1
}

check_http() {
  local url="$1"
  curl -fsS --max-time 15 "$url" >/dev/null
}

# Resolve ingress base URL (Rancher Desktop uses localhost, not EXTERNAL-IP).
ingress_base_url() {
  if [ -n "${SERVER_BASE_URL:-}" ]; then
    echo "${SERVER_BASE_URL}"
    return
  fi
  local ip
  ip="$(kubectl get svc -n kube-system traefik -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || true)"
  if [ -z "$ip" ]; then
    ip="$(kubectl get svc -n kube-system traefik -o jsonpath='{.status.loadBalancer.ingress[0].hostname}' 2>/dev/null || true)"
  fi
  if [ "$ip" = "192.168.127.2" ] || curl -fsS --connect-timeout 2 -o /dev/null -H "Host: ${INGRESS_HOST}" "http://127.0.0.1/health" 2>/dev/null; then
    echo "http://127.0.0.1"
    return
  fi
  if [ -n "$ip" ]; then
    echo "http://${ip}"
    return
  fi
  echo "http://127.0.0.1"
}

server_url() {
  if [ -n "${SERVER_URL:-}" ]; then
    echo "$SERVER_URL"
    return
  fi
  local base
  base="$(ingress_base_url)"
  echo "${base}"
}

# Call server API through ingress with Host header.
server_curl() {
  local path="$1"
  shift
  local base url host
  base="$(server_url)"
  url="${base}${path}"
  host="${INGRESS_HOST}"
  curl -fsS --max-time 30 -H "Host: ${host}" "$@" "$url"
}

# Expose ClusterIP observability services on localhost.
port_forward() {
  local svc="$1"
  local local_port="$2"
  local remote_port="$3"
  kubectl port-forward -n "$NAMESPACE" "svc/${svc}" "${local_port}:${remote_port}" >/dev/null 2>&1 &
  PF_PIDS+=("$!")
  sleep 2
}

check_prometheus_target_up() {
  local job="$1"
  curl -fsS --max-time 10 "${PROM_URL}/api/v1/targets" \
    | grep -q "\"job\":\"${job}\".*\"health\":\"up\""
}

check_prometheus_query_positive() {
  local query="$1"
  local result
  result="$(curl -fsS --max-time 10 -G "${PROM_URL}/api/v1/query" \
    --data-urlencode "query=${query}")"
  echo "$result" | grep -q '"status":"success"'
  echo "$result" | grep -Eq '"value":\[[^]]+,"[0-9]+(\.[0-9]+)?"\]' \
    && ! echo "$result" | grep -qE '"value":\[[^]]+,"0"\]'
}

loki_query_has_logs() {
  local query="$1"
  local end start resp
  end=$(($(date +%s) * 1000000000))
  start=$((end - 1800 * 1000000000))
  resp="$(curl -fsS --max-time 15 -G "${LOKI_URL}/loki/api/v1/query_range" \
    --data-urlencode "query=${query}" \
    --data-urlencode "start=${start}" \
    --data-urlencode "end=${end}" \
    --data-urlencode "limit=5")"
  echo "$resp" | grep -q '"status":"success"' && echo "$resp" | grep -q '"values"'
}

# Upload minimal PNG and exercise upload → outbox → processor path.
upload_test_avatar() {
  local upload_dir="${TMPDIR:-/tmp}"
  mkdir -p "$upload_dir"
  local test_png="${upload_dir}/gophprofile-k8s-smoke.png"
  printf '%s' 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==' \
    | base64 -d >"$test_png"
  (cd "$upload_dir" && server_curl "/api/v1/avatars" -X POST \
    -H "X-User-ID: smoke-alice" \
    -F "file=@gophprofile-k8s-smoke.png;type=image/png")
}

log "=== GophProfile Kubernetes smoke (${PROFILE}) ==="

# Wait for namespace.
if kubectl get namespace "$NAMESPACE" >/dev/null 2>&1; then
  pass "namespace ${NAMESPACE} exists"
else
  fail "namespace ${NAMESPACE} exists"
fi

# Wait for application deployments.
for dep in server processor; do
  if retry 30 10 kubectl wait -n "$NAMESPACE" --for=condition=available \
    "deployment/${FULLNAME}-${dep}" --timeout=600s 2>/dev/null; then
    pass "deployment ${dep} available"
  else
    fail "deployment ${dep} available"
  fi
done

# Wait for data layer StatefulSets.
for sts in postgres minio rabbitmq; do
  if retry 30 10 kubectl wait -n "$NAMESPACE" --for=jsonpath='{.status.readyReplicas}'=1 \
    "statefulset/${FULLNAME}-${sts}" --timeout=600s 2>/dev/null; then
    pass "statefulset ${sts} ready"
  else
    fail "statefulset ${sts} ready"
  fi
done

# Check ingress and server health.
if retry 12 5 server_curl "/health"; then
  pass "ingress ${INGRESS_HOST} /health"
else
  fail "ingress ${INGRESS_HOST} /health (traefik + Host header?)"
fi

# Initialize observability URLs and port-forwards.
PROM_URL="${PROM_URL:-http://127.0.0.1:19090}"
LOKI_URL="${LOKI_URL:-http://127.0.0.1:13100}"
JAEGER_URL="${JAEGER_URL:-http://127.0.0.1:16686}"
COLLECTOR_METRICS_URL="${COLLECTOR_METRICS_URL:-http://127.0.0.1:18889/metrics}"
WEBHOOK_URL="${WEBHOOK_URL:-http://127.0.0.1:15001}"

port_forward "${FULLNAME}-prometheus" 19090 9090
port_forward "${FULLNAME}-loki" 13100 3100
port_forward "${FULLNAME}-jaeger" 16686 16686
port_forward "${FULLNAME}-otel-collector" 18889 8889
port_forward "${FULLNAME}-alert-webhook" 15001 80

# Check Prometheus scrape targets.
for job in prometheus otel-collector node-exporter rabbitmq; do
  if retry 12 5 check_prometheus_target_up "$job"; then
    pass "prometheus target ${job}=UP"
  else
    fail "prometheus target ${job}=UP"
  fi
done

# Check otel-collector metrics export.
if retry 12 5 bash -c "curl -fsS --max-time 10 '${COLLECTOR_METRICS_URL}' | grep -q http_server_requests"; then
  pass "otel-collector exposes http_server_requests"
else
  fail "otel-collector exposes http_server_requests"
fi

# Check alert webhook receiver.
if retry 6 5 check_http "${WEBHOOK_URL}/"; then
  pass "alert-webhook reachable"
else
  fail "alert-webhook reachable"
fi

# Check Loki readiness.
if retry 12 5 check_http "${LOKI_URL}/ready"; then
  pass "loki /ready"
else
  fail "loki /ready"
fi

# Check Jaeger UI.
if retry 6 5 check_http "${JAEGER_URL}/"; then
  pass "jaeger UI"
else
  fail "jaeger UI"
fi

# Prod-only: verify Vault → ESO → Secret chain.
if [ "$PROFILE" = "prod" ]; then
  if kubectl get externalsecret -n "$NAMESPACE" "${FULLNAME}-app" >/dev/null 2>&1; then
    pass "ExternalSecret app exists"
    if retry 12 10 kubectl get secret -n "$NAMESPACE" "${FULLNAME}-app" >/dev/null 2>&1; then
      pass "Secret app synced from Vault"
    else
      fail "Secret app synced from Vault"
    fi
  else
    fail "ExternalSecret app exists"
  fi
fi

# Prod-only: verify HPA objects.
if [ "$PROFILE" = "prod" ]; then
  if kubectl get hpa -n "$NAMESPACE" "${FULLNAME}-server" >/dev/null 2>&1; then
    pass "HPA server exists"
  else
    fail "HPA server exists"
  fi
fi

# Upload avatar and verify API response.
UPLOAD_RESP="$(upload_test_avatar 2>/dev/null || true)"
if [ -n "$UPLOAD_RESP" ] && echo "$UPLOAD_RESP" | grep -q '"id"'; then
  pass "avatar upload 201"
  AVATAR_ID="$(echo "$UPLOAD_RESP" | sed -n 's/.*"id":"\([^"]*\)".*/\1/p' | head -1)"
  log "uploaded avatar id=${AVATAR_ID}"
else
  fail "avatar upload 201"
  AVATAR_ID=""
fi

# Wait for OTLP batch and Prometheus scrape interval.
sleep 25

# Check upload metric in Prometheus or collector.
if retry 12 6 check_prometheus_query_positive 'sum(avatars_uploads_total{job="gophprofile-server"})'; then
  pass "avatars_uploads_total incremented"
elif retry 3 5 bash -c "curl -fsS --max-time 10 '${COLLECTOR_METRICS_URL}' | grep -q avatars_uploads_total"; then
  pass "avatars_uploads_total incremented (collector)"
else
  fail "avatars_uploads_total incremented"
fi

# Check server logs in Loki.
if retry 10 5 loki_query_has_logs '{service_name="gophprofile-server"}'; then
  pass "loki logs from gophprofile-server"
else
  fail "loki logs from gophprofile-server"
fi

# Check server traces in Jaeger.
if retry 10 6 bash -c "curl -fsS --max-time 10 '${JAEGER_URL}/api/services' | grep -q gophprofile-server"; then
  pass "jaeger service gophprofile-server registered"
else
  fail "jaeger service gophprofile-server registered"
fi

# Wait for processor thumbnails and verify metadata.
if [ -n "$AVATAR_ID" ]; then
  META_OK=false
  for _ in $(seq 1 18); do
    META="$(server_curl "/api/v1/avatars/${AVATAR_ID}/metadata" 2>/dev/null || true)"
    if echo "$META" | grep -q '100x100'; then
      META_OK=true
      break
    fi
    sleep 5
  done
  if $META_OK; then
    pass "avatar processing completed (thumbnails present)"
  else
    fail "avatar processing completed (thumbnails present)"
  fi
fi

# Report results.
log "=== Results: ${PASS} passed, ${FAIL} failed ==="
if [ "$FAIL" -gt 0 ]; then
  kubectl get pods -n "$NAMESPACE" 2>/dev/null || true
  exit 1
fi

log "All Kubernetes smoke checks passed."
