# Kubernetes Database Node Taint Lab

This lab dedicates one worker node to database workloads:

- `scripts/taint-db-node.sh` adds a database label and `NoSchedule` taint.
- PostgreSQL uses a matching node selector and toleration, plus a persistent volume claim.
- A test Pod with the same node selector but no toleration stays Pending.
- `scripts/untaint-db-node.sh` removes the lab label and taint.

Use a disposable worker node in a lab cluster. Do not taint a shared or production node as this can prevent unrelated Pods from being scheduled there. The PostgreSQL password below is for lab use; use a proper secrets workflow for real environments.

## Requirements

- `kubectl` configured for a test cluster
- A worker node you can dedicate to the lab
- A default StorageClass or another provisioner that can create a 1 GiB `ReadWriteOnce` PVC

Check your context and nodes first:

```sh
kubectl config current-context
kubectl get nodes -o wide
kubectl get storageclass
```

## 1. Create the namespace and database credential

From this directory:

```sh
kubectl apply -f k8s/namespace.yaml
kubectl create secret generic postgres-auth \
  --namespace database-lab \
  --from-literal=POSTGRES_PASSWORD='change-this-lab-password'
```

## 2. Label and taint one worker node

Pass the exact worker node name from `kubectl get nodes`:

```sh
./scripts/taint-db-node.sh worker-node-name
```

The script prints the current Kubernetes context and changes only the named node. It applies:

- Label `lab.example.com/dedicated=database`, used by database Pods' `nodeSelector`.
- Taint `lab.example.com/dedicated=database:NoSchedule`, which repels Pods without a matching toleration.

## 3. Run PostgreSQL on the dedicated node

```sh
kubectl apply -f k8s/postgres.yaml
kubectl -n database-lab get pods,pvc -o wide
kubectl -n database-lab rollout status statefulset/postgres
```

Confirm the PostgreSQL Pod is on the labeled node:

```sh
kubectl -n database-lab get pod postgres-0 -o wide
kubectl -n database-lab describe pod postgres-0
```

If the Pod is Pending, inspect both the Pod and its claim:

```sh
kubectl -n database-lab describe pod postgres-0
kubectl -n database-lab describe pvc data-postgres-0
kubectl get events -n database-lab --sort-by=.lastTimestamp
```

A Pending PVC usually means the cluster has no default StorageClass or no provisioner that can satisfy the claim. Configure a suitable class or add `storageClassName` to the claim template before proceeding.

## 4. See a Pod rejected by the taint

This test Pod selects the dedicated database node but has no toleration:

```sh
kubectl apply -f k8s/pod-no-toleration.yaml
kubectl -n database-lab get pod no-db-toleration -w
```

In another terminal, inspect the scheduling event:

```sh
kubectl -n database-lab describe pod no-db-toleration
```

Expect `FailedScheduling` with a message about an untolerated taint. Remove the test Pod:

```sh
kubectl -n database-lab delete pod no-db-toleration
```

## 5. Connect with a Pod that tolerates the database taint

```sh
kubectl apply -f k8s/db-client-tolerated.yaml
kubectl -n database-lab logs -f db-client
```

The client waits for PostgreSQL, runs a small `SELECT`, and then stays alive. Its matching toleration lets it run on the tainted node; the node selector directs it there. A toleration permits placement but does not select or attract a Pod to a node by itself.

## 6. Restore the node and clean up

Delete the lab workloads and namespace, then remove the taint and label from the selected node:

```sh
kubectl delete namespace database-lab
./scripts/untaint-db-node.sh worker-node-name
```

Verify that the lab taint and label are gone:

```sh
kubectl describe node worker-node-name
kubectl get node worker-node-name --show-labels
```

The namespace deletion also deletes the PVC. Whether the backing volume is retained or deleted depends on its StorageClass reclaim policy. This lab is intended for test data.

## Kubernetes reference

Kubernetes documents taints as a way for nodes to repel Pods that lack a matching toleration. Tolerations allow scheduling onto a tainted node but do not guarantee or attract placement. See [Taints and Tolerations](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/).
