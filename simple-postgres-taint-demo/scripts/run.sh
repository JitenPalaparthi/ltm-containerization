#!/usr/bin/env sh
set -eu

if [ "$#" -ne 1 ]; then
  echo "Usage: $0 <minikube-or-kind-node-name>" >&2
  echo "Find node names with: kubectl get nodes" >&2
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
kubectl get node "$node" >/dev/null
printf 'Using local context %s and node %s\n' "$context" "$node"

kubectl create namespace taint-demo --dry-run=client -o yaml | kubectl apply -f -
: "${POSTGRES_PASSWORD:=lab-only-change-me}"
kubectl create secret generic postgres-password \
  --namespace taint-demo \
  --from-literal="password=$POSTGRES_PASSWORD" \
  --dry-run=client -o yaml | kubectl apply -f -

kubectl label node "$node" workload=database --overwrite
kubectl taint node "$node" workload=database:NoSchedule --overwrite
kubectl apply -f "$(dirname "$0")/../k8s/postgres.yaml"
kubectl -n taint-demo wait --for=condition=Ready pod/postgres --timeout=180s
printf '\nPostgreSQL is ready. Connect with:\n'
printf '  kubectl exec -it -n taint-demo postgres -- psql -U labuser -d labdb\n'
printf '\nTo see the rejection example:\n'
printf '  kubectl apply -f %s\n' "$(dirname "$0")/../k8s/no-toleration.yaml"
