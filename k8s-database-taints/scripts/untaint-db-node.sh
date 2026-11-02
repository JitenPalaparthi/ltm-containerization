#!/usr/bin/env sh
set -eu

if [ "$#" -ne 1 ]; then
  echo "Usage: $0 <worker-node-name>" >&2
  exit 2
fi

node="$1"
key='lab.example.com/dedicated'
value='database'
context="$(kubectl config current-context)"
printf 'Kubernetes context: %s\nRestoring node: %s\n' "$context" "$node"
kubectl get node "$node" >/dev/null

kubectl taint node "$node" "$key=$value:NoSchedule-"
kubectl label node "$node" "$key-"
printf 'Removed the lab database taint and label from %s.\n' "$node"
