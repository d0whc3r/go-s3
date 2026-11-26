#!/bin/sh
# Copyright © 2020-2026 d0whc3r
# Licensed under the Apache License, Version 2.0.
#
# Install gos3 from GitHub Releases:
#   curl -fsSL https://raw.githubusercontent.com/d0whc3r/go-s3/master/install.sh | sh
#
# Optional environment variables:
#   VERSION      Specific version to install (e.g. v2.0.0 or 2.0.0). Defaults to latest.
#   INSTALL_DIR  Destination directory (default: /usr/local/bin).

set -eu

REPO="d0whc3r/go-s3"
BINARY="gos3"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

log() {
  printf '[gos3-install] %s\n' "$*"
}

fail() {
  printf '[gos3-install] error: %s\n' "$*" >&2
  exit 1
}

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || fail "required command not found: $1"
}

detect_os() {
  os=$(uname -s | tr '[:upper:]' '[:lower:]')
  case "$os" in
    linux)   printf 'linux' ;;
    darwin)  printf 'darwin' ;;
    freebsd) printf 'freebsd' ;;
    *)       fail "unsupported OS: $os (download manually from https://github.com/${REPO}/releases)" ;;
  esac
}

detect_arch() {
  arch=$(uname -m)
  case "$arch" in
    x86_64|amd64)     printf 'amd64' ;;
    aarch64|arm64)    printf 'arm64' ;;
    armv7l|armv6l|arm) printf 'arm' ;;
    ppc64le)          printf 'ppc64le' ;;
    *)                fail "unsupported architecture: $arch" ;;
  esac
}

resolve_version() {
  if [ -n "${VERSION:-}" ]; then
    printf '%s' "${VERSION#v}"
    return
  fi

  # Resolve latest tag via HTTP redirect (avoids GitHub API rate limits).
  latest_url=$(curl -fsSL -o /dev/null -w '%{url_effective}' "https://github.com/${REPO}/releases/latest") || \
    fail "could not determine latest release version"
  tag=$(printf '%s' "$latest_url" | sed 's|.*/tag/||')
  if [ -z "$tag" ] || [ "$tag" = "$latest_url" ]; then
    fail "unexpected latest release URL: $latest_url"
  fi
  printf '%s' "${tag#v}"
}

verify_checksum() {
  file="$1"
  sums_file="$2"
  asset_name="$3"

  expected=$(awk -v name="$asset_name" '$2 == name || $2 == ("*" name) { print $1; exit }' "$sums_file")
  if [ -z "$expected" ]; then
    log "checksum entry not found for ${asset_name}; skipping verification"
    return 0
  fi

  if command -v sha256sum >/dev/null 2>&1; then
    actual=$(sha256sum "$file" | awk '{ print $1 }')
  elif command -v shasum >/dev/null 2>&1; then
    actual=$(shasum -a 256 "$file" | awk '{ print $1 }')
  else
    log "neither sha256sum nor shasum found; skipping checksum verification"
    return 0
  fi

  if [ "$actual" != "$expected" ]; then
    fail "checksum mismatch for ${asset_name} (expected ${expected}, got ${actual})"
  fi
  log "SHA-256 checksum verified"
}

main() {
  need_cmd curl
  need_cmd uname

  os=$(detect_os)
  arch=$(detect_arch)
  ver=$(resolve_version)

  asset="${BINARY}_v${ver}_${os}-${arch}"
  base_url="https://github.com/${REPO}/releases/download/v${ver}"

  tmp_dir=$(mktemp -d 2>/dev/null || mktemp -d -t gos3-install)
  trap 'rm -rf "$tmp_dir"' EXIT

  log "downloading ${BINARY} v${ver} (${os}/${arch})..."
  curl -fsSL "${base_url}/${asset}" -o "${tmp_dir}/${BINARY}" || \
    fail "failed to download ${base_url}/${asset}"

  if curl -fsSL "${base_url}/SHA256SUMS.txt" -o "${tmp_dir}/SHA256SUMS.txt" 2>/dev/null; then
    verify_checksum "${tmp_dir}/${BINARY}" "${tmp_dir}/SHA256SUMS.txt" "${asset}"
  fi

  chmod +x "${tmp_dir}/${BINARY}"

  if [ ! -d "$INSTALL_DIR" ]; then
    mkdir -p "$INSTALL_DIR" 2>/dev/null || sudo mkdir -p "$INSTALL_DIR"
  fi

  if [ -w "$INSTALL_DIR" ]; then
    mv "${tmp_dir}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
  elif command -v sudo >/dev/null 2>&1; then
    log "installing to ${INSTALL_DIR} (using sudo)..."
    sudo mv "${tmp_dir}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
  else
    fail "cannot write to ${INSTALL_DIR}; re-run with INSTALL_DIR=\"\$HOME/.local/bin\""
  fi

  log "installed ${BINARY} to ${INSTALL_DIR}/${BINARY}"
  "${INSTALL_DIR}/${BINARY}" --version

  case ":${PATH}:" in
    *":${INSTALL_DIR}:"*) ;;
    *) log "add ${INSTALL_DIR} to your PATH to run '${BINARY}' directly" ;;
  esac
}

main "$@"
