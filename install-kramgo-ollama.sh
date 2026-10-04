#!/usr/bin/env bash
set -euo pipefail

INSTALL_DIR="${INSTALL_DIR:-/home/dtk1376}"
SOURCE_DIR="${1:-$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)}"
SERVICE_NAME="kramgo-ollama"
WATCHDOG_SERVICE_NAME="kramgo-watchdog"

if [[ $EUID -ne 0 ]]; then
  echo "This script must be run as root." >&2
  exit 1
fi

for required in \
  "${SOURCE_DIR}/kramgo" \
  "${SOURCE_DIR}/kramgo-watchdog" \
  "${SOURCE_DIR}/kramgo-ollama.service" \
  "${SOURCE_DIR}/kramgo-watchdog.service"; do
  if [[ ! -f "${required}" ]]; then
    echo "Missing required file: ${required}" >&2
    exit 1
  fi
done

install -d "${INSTALL_DIR}"

install_if_needed() {
  local src="$1"
  local dst="$2"
  local mode="$3"

  if [[ "$(realpath "${src}")" == "$(realpath "${dst}")" ]]; then
    echo "Skipping copy: ${src} already matches ${dst}"
    return
  fi

  install -m "${mode}" "${src}" "${dst}"
}

install_if_needed "${SOURCE_DIR}/kramgo" "${INSTALL_DIR}/kramgo" "0755"
install_if_needed "${SOURCE_DIR}/kramgo-watchdog" "${INSTALL_DIR}/kramgo-watchdog" "0755"
install_if_needed "${SOURCE_DIR}/kramgo-ollama.service" "/etc/systemd/system/${SERVICE_NAME}.service" "0644"
install_if_needed "${SOURCE_DIR}/kramgo-watchdog.service" "/etc/systemd/system/${WATCHDOG_SERVICE_NAME}.service" "0644"

systemctl daemon-reload
systemctl enable --now "${SERVICE_NAME}.service" "${WATCHDOG_SERVICE_NAME}.service"

systemctl status "${SERVICE_NAME}.service" --no-pager
systemctl status "${WATCHDOG_SERVICE_NAME}.service" --no-pager
