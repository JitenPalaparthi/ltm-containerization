#!/usr/bin/env sh
set -eu
CLUSTER="${1:-kind}"
docker build -t go-oom-demo:v1 .
kind load docker-image go-oom-demo:v1 --name "$CLUSTER"
echo 'Image loaded. Run: kubectl apply -f k8s/01-oom.yaml'
