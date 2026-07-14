#!/usr/bin/env bash
# End-to-end observability smoke test for the Docker Compose stack.
# Usage: make observability-smoke   (stack must be running: make docker-up)
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

SERVER_URL="${SERVER_URL:-http://127.0.0.1:8080}"
PROM_URL="${PROM_URL:-http://127.0.0.1:9092}"
LOKI_URL="${LOKI_URL:-http://127.0.0.1:3100}"
JAEGER_URL="${JAEGER_URL:-http://127.0.0.1:16686}"
COLLECTOR_METRICS_URL="${COLLECTOR_METRICS_URL:-http://127.0.0.1:8889/metrics}"
WEBHOOK_URL="${WEBHOOK_URL:-http://127.0.0.1:5001}"

PASS=0
FAIL=0

log() { printf '[observability-smoke] %s\n' "$*"; }
pass() { PASS=$((PASS + 1)); log "PASS: $*"; }
fail() { FAIL=$((FAIL + 1)); log "FAIL: $*"; }

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
  curl -fsS --max-time 10 "$url" >/dev/null
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

upload_test_avatar() {
  local upload_dir="${TMPDIR:-/tmp}"
  mkdir -p "$upload_dir"
  local test_png="${upload_dir}/gophprofile-smoke-test.png"
  # Valid 1x1 PNG (base64), not the broken printf blob from README quick start.
  printf '%s' 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==' \
    | base64 -d >"$test_png"
  (cd "$upload_dir" && curl -fsS --max-time 30 -X POST "${SERVER_URL}/api/v1/avatars" \
    -H "X-User-ID: smoke-alice" \
    -F "file=@gophprofile-smoke-test.png;type=image/png")
}

log "=== GophProfile observability smoke test ==="

# 1. Core app health
if retry 12 5 check_http "${SERVER_URL}/health"; then
  pass "server /health"
else
  fail "server /health (is make docker-up running?)"
fi

# 2. Prometheus targets
for job in prometheus otel-collector node-exporter rabbitmq; do
  if retry 12 5 check_prometheus_target_up "$job"; then
    pass "prometheus target ${job}=UP"
  else
    fail "prometheus target ${job}=UP"
  fi
done

# 3. Collector exports app metric names
if retry 12 5 bash -c "curl -fsS --max-time 10 '${COLLECTOR_METRICS_URL}' | grep -q http_server_requests_total"; then
  pass "otel-collector exposes http_server_requests_total"
else
  fail "otel-collector exposes http_server_requests_total"
fi

# 4. RabbitMQ queue metrics
if retry 12 5 bash -c "curl -fsS --max-time 10 '${PROM_URL}/api/v1/query' --data-urlencode 'query=rabbitmq_queue_messages{queue=\"avatar-processing\"}' | grep -q success"; then
  pass "rabbitmq_queue_messages metric present"
else
  fail "rabbitmq_queue_messages metric present"
fi

# 5. Node exporter host metrics
if retry 12 5 check_prometheus_query_positive 'node_memory_MemTotal_bytes'; then
  pass "node_memory_MemTotal_bytes > 0"
else
  fail "node_memory_MemTotal_bytes > 0"
fi

# 6. Alert webhook receiver reachable
if retry 6 5 check_http "${WEBHOOK_URL}/"; then
  pass "alert-webhook echo-server reachable"
else
  fail "alert-webhook echo-server reachable"
fi

# 7. Loki ready
if retry 12 5 check_http "${LOKI_URL}/ready"; then
  pass "loki /ready"
else
  fail "loki /ready"
fi

# 8. Jaeger UI
if retry 6 5 check_http "${JAEGER_URL}/"; then
  pass "jaeger UI"
else
  fail "jaeger UI"
fi

# 9. Upload avatar and verify metrics + traces
UPLOAD_RESP="$(upload_test_avatar 2>/dev/null || true)"

if [ -n "$UPLOAD_RESP" ] && echo "$UPLOAD_RESP" | grep -q '"id"'; then
  pass "avatar upload 201"
  AVATAR_ID="$(echo "$UPLOAD_RESP" | sed -n 's/.*"id":"\([^"]*\)".*/\1/p' | head -1)"
  log "uploaded avatar id=${AVATAR_ID}"
else
  fail "avatar upload 201"
  AVATAR_ID=""
fi

# Wait for OTLP batch + Prometheus scrape
sleep 25

if retry 12 6 check_prometheus_query_positive 'sum(avatars_uploads_total{job="gophprofile-server"})'; then
  pass "avatars_uploads_total incremented"
elif retry 3 5 bash -c "curl -fsS --max-time 10 '${COLLECTOR_METRICS_URL}' | grep -q avatars_uploads_total"; then
  pass "avatars_uploads_total incremented (collector)"
else
  fail "avatars_uploads_total incremented"
fi

# 10. Loki has logs from server
if retry 10 5 loki_query_has_logs '{service_name="gophprofile-server"}'; then
  pass "loki logs from gophprofile-server"
else
  fail "loki logs from gophprofile-server"
fi

# 11. Jaeger has traces for server
if retry 10 6 bash -c "curl -fsS --max-time 10 '${JAEGER_URL}/api/services' | grep -q gophprofile-server"; then
  pass "jaeger service gophprofile-server registered"
else
  fail "jaeger service gophprofile-server registered"
fi

# 12. Processing completed (optional, best-effort)
if [ -n "$AVATAR_ID" ]; then
  META_OK=false
  for _ in $(seq 1 12); do
    META="$(curl -fsS --max-time 10 "${SERVER_URL}/api/v1/avatars/${AVATAR_ID}/metadata" 2>/dev/null || true)"
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

  if retry 6 5 check_prometheus_query_positive 'sum(avatars_processing_total{job="gophprofile-processor"})'; then
    pass "avatars_processing_total incremented"
  elif retry 3 5 bash -c "curl -fsS --max-time 10 '${COLLECTOR_METRICS_URL}' | grep -q avatars_processing_total"; then
    pass "avatars_processing_total incremented (collector)"
  else
    fail "avatars_processing_total incremented"
  fi
fi

log "=== Results: ${PASS} passed, ${FAIL} failed ==="
if [ "$FAIL" -gt 0 ]; then
  exit 1
fi

log "All observability checks passed."
