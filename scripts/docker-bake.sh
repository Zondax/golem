#!/usr/bin/env bash
set -euo pipefail

detect_git_metadata() {
  if [ -z "${GIT_SHORT_HASH:-}" ]; then
    export GIT_SHORT_HASH=$(git rev-parse --short HEAD 2>/dev/null || echo "")
  fi

  if [ -z "${GIT_BRANCH:-}" ]; then
    export GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null | tr '/' '-' || echo "")
  fi

  if [ -z "${GIT_TAG:-}" ]; then
    export GIT_TAG=$(git describe --tags --exact-match HEAD 2>/dev/null || echo "")
  fi

  if [ -z "${GIT_COMMIT_TIMESTAMP:-}" ]; then
    local unix_ts
    unix_ts=$(git log -1 --format=%ct 2>/dev/null || echo "")
    if [ -n "$unix_ts" ]; then
      if date --version >/dev/null 2>&1; then
        export GIT_COMMIT_TIMESTAMP=$(date -d "@${unix_ts}" +"%Y%m%d%H%M%S")
      else
        export GIT_COMMIT_TIMESTAMP=$(date -r "${unix_ts}" +"%Y%m%d%H%M%S")
      fi
    fi
  fi
}

print_config() {
  echo "Git metadata detected:"
  echo "  GIT_SHORT_HASH:       ${GIT_SHORT_HASH:-<not set>}"
  echo "  GIT_BRANCH:           ${GIT_BRANCH:-<not set>}"
  echo "  GIT_TAG:              ${GIT_TAG:-<not set>}"
  echo "  GIT_COMMIT_TIMESTAMP: ${GIT_COMMIT_TIMESTAMP:-<not set>}"
}

main() {
  local push_flag=""
  local extra_args=()

  for arg in "$@"; do
    case "$arg" in
      --push)  push_flag="--push" ;;
      --print) detect_git_metadata; print_config; exit 0 ;;
      *)       extra_args+=("$arg") ;;
    esac
  done

  detect_git_metadata
  print_config

  export DOCKER_BUILDKIT=1
  export BUILDX_NO_DEFAULT_ATTESTATIONS=1

  docker buildx bake $push_flag "${extra_args[@]+"${extra_args[@]}"}"
}

main "$@"
