#!/usr/bin/env bash
set -euo pipefail

SCRIPT_PATH=$(realpath "${0}")
SCRIPT_DIR=$(dirname "${SCRIPT_PATH}")
BUILD_DIR="${SCRIPT_DIR}/build"

function build() {
  local binary="${1}"

  pushd "cmd/${binary}" >>/dev/null

  echo "Building ${binary}"
  go build
  mv "${binary}" "${BUILD_DIR}"

  popd >>/dev/null
}

if [[ -d "${BUILD_DIR}" ]]; then
  rm -rf "${BUILD_DIR}"
fi

mkdir "${BUILD_DIR}"

build "eventserver"
build "recordkeeper"
