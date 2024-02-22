#!/bin/bash

set -eu

SCRIPT_NAME="$(basename "$0")"
SCRIPT_DIR="$(dirname "$0")"
REPO_ROOT="$(cd "${SCRIPT_DIR}" && git rev-parse --show-toplevel)"
SCRIPTS_DIR="${REPO_ROOT}/scripts"
source "${SCRIPTS_DIR}/helpers-source.sh"

echo "${SCRIPT_NAME} is running... "

export DOCKERFILE_PATH="${REPO_ROOT}/build/docker/cthulhu-mythos-tools/Dockerfile"
export IMAGE_NAME="${CTHULHU_MYTHOS_TOOLS_IMAGE:-${DOCKER_REPO}cthulhu-mythos-tools:${VERSION}}"

echo "Building ${IMAGE_NAME} of ${APP_NAME} ..."

docker buildx bake -f "${REPO_ROOT}/build/docker/bake.hcl" "${APP_NAME}"