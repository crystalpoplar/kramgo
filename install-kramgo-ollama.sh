#!/usr/bin/env bash
set -euo pipefail

SERVICE_NAME="kramgo-ollama"
SERVICE_FILE="${SERVICE_NAME}.service"
WATCHDOG_SERVICE_NAME="kramgo-watchdog"
WATCHDOG_SERVICE_FILE="${WATCHDOG_SERVICE_NAME}.service"
BINARY_PATH="/home/dtk1376/kramgo"
WATCHDOG_BINARY_PATH="/home/dtk1376/kramgo-watchdog"
WORK_DIR="/home/dtk1376"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [[ $EUID -ne 0 ]]; then
  echo "This script must be run as root." >&2
  exit 1
fi

if [[ ! -f "${REPO_ROOT}/${SERVICE_FILE}" ]]; then
  echo "Service file not found: ${REPO_ROOT}/${SERVICE_FILE}" >&2
  exit 1
fi

if [[ ! -f "${REPO_ROOT}/${WATCHDOG_SERVICE_FILE}" ]]; then
  echo "Watchdog service file not found: ${REPO_ROOT}/${WATCHDOG_SERVICE_FILE}" >&2
  exit 1
fi

if ! command -v go >/dev/null 2>&1; then
  echo "Go is not installed. Install Go first." >&2
  exit 1
fi

mkdir -p "${WORK_DIR}"

cd "${REPO_ROOT}"
CGO_ENABLED=0 go build -o "${BINARY_PATH}" ./cmd/kramgo
CGO_ENABLED=0 go build -o "${WATCHDOG_BINARY_PATH}" ./cmd/kramgo-watchdog

cp "${REPO_ROOT}/${SERVICE_FILE}" "/etc/systemd/system/${SERVICE_FILE}"
cp "${REPO_ROOT}/${WATCHDOG_SERVICE_FILE}" "/etc/systemd/system/${WATCHDOG_SERVICE_FILE}"

systemctl daemon-reload
systemctl enable --now "${SERVICE_NAME}.service" "${WATCHDOG_SERVICE_NAME}.service"

systemctl status "${SERVICE_NAME}.service" --no-pager
systemctl status "${WATCHDOG_SERVICE_NAME}.service" --no-pager
