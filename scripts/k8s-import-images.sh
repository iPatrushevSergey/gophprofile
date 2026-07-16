#!/usr/bin/env bash
# Import locally built app images into Rancher Desktop k3s when Docker engines differ.
# If the cluster shares the host docker engine, only verify images exist.
set -euo pipefail

# Bootstrap defaults.
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

TAR="${TMPDIR:-/tmp}/gophprofile-k8s.tar"
if [[ -n "${TMP:-}" ]]; then
  TAR="${TMP}/gophprofile-k8s.tar"
fi

log() { printf '[k8s-import-images] %s\n' "$*"; }

# Application images to import.
IMAGES=(gophprofile-server:latest gophprofile-processor:latest gophprofile-migrate:latest)

# Verify local images exist.
for img in "${IMAGES[@]}"; do
  if ! docker image inspect "$img" >/dev/null 2>&1; then
    log "missing image ${img}; run: make k8s-build-images"
    exit 1
  fi
done

# Load images into Rancher Desktop WSL distro.
load_via_wsl() {
  local distro="$1"
  local win_path="$TAR"
  if command -v cygpath >/dev/null 2>&1; then
    win_path="$(cygpath -w "$TAR")"
  fi
  local wsl_path
  wsl_path="$(MSYS_NO_PATHCONV=1 wsl -d "$distro" wslpath -u "$win_path")"
  log "loading into ${distro} docker via ${wsl_path}"
  MSYS_NO_PATHCONV=1 wsl -d "$distro" docker load -i "$wsl_path"
}

# Detect whether import is needed (separate Rancher Desktop engine).
import_needed=false
if command -v rdctl >/dev/null 2>&1; then
  import_needed=true
elif wsl -l -v 2>/dev/null | grep -qE 'rancher-desktop(-data)?'; then
  import_needed=true
fi

# Skip import when cluster shares host docker engine.
if ! $import_needed; then
  log "cluster shares host docker engine; images OK"
  exit 0
fi

# Save images to tar archive.
log "saving images to ${TAR}"
docker save "${IMAGES[@]}" -o "$TAR"

# Load images into cluster (rdctl or WSL).
if command -v rdctl >/dev/null 2>&1; then
  log "loading via rdctl"
  rdctl shell -- docker load -i "$TAR"
elif wsl -l -v 2>/dev/null | grep -qE 'rancher-desktop(-data)?'; then
  for distro in rancher-desktop rancher-desktop-data; do
    if wsl -l -v 2>/dev/null | grep -qw "$distro"; then
      load_via_wsl "$distro"
      log "done"
      exit 0
    fi
  done
fi

log "WARN: import path not found; images may be missing in cluster"
exit 1
