# Simple PostgreSQL Taint Demo (Minikube or kind)

This demo taints one local Kubernetes node so ordinary Pods are repelled, then runs PostgreSQL on that node with a matching toleration and node selector. It also includes a small Pod without a toleration so you can see it remain Pending.

PostgreSQL data uses `emptyDir` to keep the example short. Data is lost when the Pod or namespace is deleted. Do not use this storage setup for a real database.

## 1. Start a local cluster

Start Minikube:

```sh
minikube start
kubectl config use-context minikube
```

Or create a single-node kind cluster:

```sh
kind create cluster --name taint-demo
kubectl config use-context kind-taint-demo
```

Check the context and get the node name:

```sh
kubectl config current-context
kubectl get nodes
```

Typical node names are `minikube` for Minikube and `taint-demo-control-plane` for a one-node kind cluster. Use the exact name shown by `kubectl get nodes`. The manifest includes the control-plane toleration needed for this local, single-node kind example; do not copy that extra toleration into production workloads.

## 2. Taint the node and start PostgreSQL

From this directory, replace the sample node name with the one you saw above:

```sh
./scripts/run.sh minikube
```

For kind, it might be:

```sh
./scripts/run.sh taint-demo-control-plane
```

The script refuses contexts other than Minikube or kind. It labels the selected node `workload=database`, adds the taint `workload=database:NoSchedule`, creates a lab namespace and password Secret, then waits for PostgreSQL to become Ready. The pod uses `postgres:17-alpine`, a writable ephemeral volume, and a 512 MiB memory limit.

Set a different lab password before running if you want:

```sh
POSTGRES_PASSWORD='choose-a-lab-password' ./scripts/run.sh minikube
```

The default password is only for local learning. Do not use it in production. If the cluster cannot pull the image, preload it first:

```sh
# Minikube
docker pull postgres:17-alpine
minikube image load postgres:17-alpine

# kind
docker pull postgres:17-alpine
kind load docker-image postgres:17-alpine --name taint-demo
```

## 3. Connect to PostgreSQL

```sh
kubectl exec -it -n taint-demo postgres -- psql -U labuser -d labdb
```

At the `psql` prompt, try:

```sql
CREATE TABLE notes (id serial PRIMARY KEY, message text);
INSERT INTO notes (message) VALUES ('scheduled onto the database node');
SELECT * FROM notes;
\q
```

Confirm the Pod's node:

```sh
kubectl get pod -n taint-demo postgres -o wide
kubectl describe node minikube
```

## 4. See a Pod without a toleration stay Pending

This test Pod selects the same node but has no matching toleration:

```sh
kubectl apply -f k8s/no-toleration.yaml
kubectl get pod -n taint-demo no-toleration -w
```

In another terminal, inspect the scheduling event:

```sh
kubectl describe pod -n taint-demo no-toleration
```

Expect `FailedScheduling` because the Pod does not tolerate `workload=database:NoSchedule`. Remove the test Pod:

```sh
kubectl delete pod -n taint-demo no-toleration
```

## 5. Clean up

Pass the same node name used by `run.sh`:

```sh
./scripts/cleanup.sh minikube
```

Or for kind:

```sh
./scripts/cleanup.sh taint-demo-control-plane
```

Cleanup deletes the demo namespace, then removes the taint and label from the node. Verify with `kubectl describe node <node-name>`.

Kubernetes taints repel Pods that lack a matching toleration. A toleration permits scheduling but does not choose a node, which is why the PostgreSQL Pod also has a node selector. See [Taints and Tolerations](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/).
