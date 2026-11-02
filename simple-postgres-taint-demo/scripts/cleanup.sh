#!/usr/bin/env sh
set -eu

if [ "$#" -ne 1 ]; then
  echo "Usage: $0 <minikube-or-kind-node-name>" >&2
  exit 2
fi

context="$(kubectl config current-context)"
case "$context" in
  minikube|kind|kind-*) ;;
  *)
    echo "Refusing context '$context'. Switch kubectl to Minikube or kind first." >&2
    exit 2
    ;;
esac

node="$1"
kubectl delete namespace taint-demo --ignore-not-found
kubectl taint node "$node" workload=database:NoSchedule- || true
kubectl label node "$node" workload- || true
printf 'Removed demo namespace, taint, and label from %s.\n' "$node"
