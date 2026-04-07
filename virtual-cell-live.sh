#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SAGE_WIKI_BIN="${SCRIPT_DIR}/sage-wiki"
PROJECT_DIR="${1:-/Users/richardwang/Documents/virtual-cell}"
PORT="${PORT:-3333}"

if [[ -f "${HOME}/.zshrc" ]]; then
  # Load API keys if the script is launched outside an interactive shell.
  # shellcheck disable=SC1090
  source "${HOME}/.zshrc" >/dev/null 2>&1 || true
fi

if [[ ! -x "${SAGE_WIKI_BIN}" ]]; then
  echo "Missing executable: ${SAGE_WIKI_BIN}" >&2
  echo "Build it first with: go build -tags webui -o sage-wiki ./cmd/sage-wiki" >&2
  exit 1
fi

if [[ ! -d "${PROJECT_DIR}" ]]; then
  echo "Project directory not found: ${PROJECT_DIR}" >&2
  exit 1
fi

WATCH_PID=""

cleanup() {
  if [[ -n "${WATCH_PID}" ]] && kill -0 "${WATCH_PID}" 2>/dev/null; then
    kill "${WATCH_PID}" 2>/dev/null || true
    wait "${WATCH_PID}" 2>/dev/null || true
  fi
}

trap cleanup EXIT INT TERM

echo "Project: ${PROJECT_DIR}"
echo "UI: http://127.0.0.1:${PORT}"
echo "Starting compile watch..."

"${SAGE_WIKI_BIN}" compile --watch --project "${PROJECT_DIR}" &
WATCH_PID=$!

echo "Starting web UI..."
"${SAGE_WIKI_BIN}" serve --ui --project "${PROJECT_DIR}" --port "${PORT}"
