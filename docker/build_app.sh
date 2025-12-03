#!/bin/sh

CMD_NAME=$1
IMAGE_NAME=$2

if [ "${CMD_NAME}" = "" ] || [ "${IMAGE_NAME}" = "" ]; then
    echo "./build <cmd name> <image name>"
    exit 1
fi

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
cd "$SCRIPT_DIR" || exit 1

cd ../app || exit 1

GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
    go build -o "$SCRIPT_DIR/bin/${CMD_NAME}" \
    -ldflags="-s -w" \
    -trimpath \
    -buildvcs=false \
    -mod=readonly \
    "./cmd/${CMD_NAME}/main.go"

cd "$SCRIPT_DIR" || exit 1

docker build -t "${IMAGE_NAME}:latest" --target "${CMD_NAME}" .
