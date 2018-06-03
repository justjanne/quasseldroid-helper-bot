#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

IMAGE=k8r.eu/justjanne/quasseldroid-helper-bot
TAGS=$(git describe --always --tags HEAD)

docker build -t ${IMAGE}:${TAGS} .
docker tag ${IMAGE}:${TAGS} ${IMAGE}:latest
echo Successfully tagged ${IMAGE}:latest
docker push ${IMAGE}:${TAGS}
docker push ${IMAGE}:latest