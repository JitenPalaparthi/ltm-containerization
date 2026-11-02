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

printf 'Kubernetes context: %s\nTarget node: %s\n' "$context" "$node"
kubectl get node "$node" >/dev/null

# Label directs database Pods to this node; the taint repels Pods without a match.
kubectl label node "$node" "$key=$value" --overwrite
kubectl taint node "$node" "$key=$value:NoSchedule" --overwrite

printf '\nNode prepared. Confirm with:\n'
printf '  kubectl describe node %s\n' "$node"
