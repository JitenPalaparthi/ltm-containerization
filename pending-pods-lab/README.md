# Pending Pods Lab: Scheduling, Selectors, Taints, and Capacity

A hands-on Kubernetes lab for diagnosing `Pending` Pods. It includes a small Go HTTP service, Dockerfile, and separate manifests for node-selector mismatch, unsatisfied resource requests, and taints/tolerations.

The oversized resource-request example deliberately stays Pending on ordinary clusters. It does not consume the requested CPU or memory because the scheduler will not place it. The taint lab changes one node label and adds a `NoSchedule` taint; do that only on a disposable lab node, never a shared or production node.

## Requirements

- Go 1.24+ (for local run/build)
- Docker
- `kubectl` pointed at a disposable Kubernetes cluster
- Optional: `kind` or `minikube`

Before applying manifests, check the active cluster:

```sh
kubectl config current-context
kubectl get nodes -o wide
```

## 1. Run locally

From this directory:

```sh
go run .
```

In another terminal:

```sh
curl http://localhost:8080/healthz
curl http://localhost:8080/
```

Stop the server with Ctrl+C.

## 2. Build and run with Docker

```sh
docker build -t pending-pods-lab:local .
docker run --rm -p 8080:8080 pending-pods-lab:local
```

Check `http://localhost:8080/healthz`, then stop with Ctrl+C.

## 3. Make the image available to Kubernetes

Build it first:

```sh
docker build -t pending-pods-lab:local .
```

For Docker Desktop Kubernetes, the local Docker image is generally visible to its cluster. For kind or minikube, load the image explicitly:

```sh
kind load docker-image pending-pods-lab:local
```

or:

```sh
minikube image load pending-pods-lab:local
```

For a remote cluster, push the image to a registry it can access, then change `image:` in the manifests to that registry image.

Create the lab namespace:

```sh
kubectl apply -f k8s/namespace.yaml
```

## 4. Start the healthy baseline app

```sh
kubectl apply -f k8s/app.yaml
kubectl -n pending-lab rollout status deployment/pending-lab-app
kubectl -n pending-lab get pods -o wide
```

Access it from your machine:

```sh
kubectl -n pending-lab port-forward service/pending-lab-app 8080:8080
```

In another terminal, run `curl http://localhost:8080/`. Stop port-forwarding with Ctrl+C.

## 5. Diagnose a node-selector mismatch

The selector asks for a label value that should not exist on any node:

```sh
kubectl apply -f k8s/01-node-selector-pending.yaml
kubectl -n pending-lab get pod selector-pending -w
```

In another terminal, inspect why it is pending:

```sh
kubectl -n pending-lab describe pod selector-pending
kubectl -n pending-lab get events --sort-by=.lastTimestamp
kubectl get nodes --show-labels
```

Look for `FailedScheduling` and the node affinity/selector mismatch in Events. Remove the demo Pod:

```sh
kubectl -n pending-lab delete pod selector-pending
```

## 6. Diagnose insufficient schedulable capacity

This Pod requests 100,000 CPUs and 100,000 GiB of memory, so it should remain Pending on ordinary clusters:

```sh
kubectl apply -f k8s/02-resource-pending.yaml
kubectl -n pending-lab get pod resource-pending -w
```

Inspect its scheduling event and node capacity:

```sh
kubectl -n pending-lab describe pod resource-pending
kubectl describe nodes
kubectl get nodes -o custom-columns=NAME:.metadata.name,CPU-CAPACITY:.status.capacity.cpu,CPU-ALLOCATABLE:.status.allocatable.cpu,MEMORY-CAPACITY:.status.capacity.memory,MEMORY-ALLOCATABLE:.status.allocatable.memory
```

`kubectl describe nodes` includes Allocated resources, which helps compare requests already placed on the node with the new Pod's requests. Scheduling uses requests and node allocatable capacity; current low CPU usage from `kubectl top` does not mean the node has that much unrequested capacity.

If Metrics Server is installed, compare live usage too:

```sh
kubectl top nodes
kubectl top pods -A
```

If `kubectl top` reports that Metrics API is unavailable, continue with node descriptions, Allocated resources, and Events. Delete the deliberately oversized Pod when finished:

```sh
kubectl -n pending-lab delete pod resource-pending
```

## 7. Taints and tolerations on one lab node

Choose a disposable worker node. Do not use a shared or production node: adding a taint can stop unrelated Pods from being scheduled there.

Set its name in your shell and add a lab-only label and taint:

```sh
export NODE=<disposable-worker-node-name>
kubectl label node "$NODE" lab.example.com/dedicated=pending-lab --overwrite
kubectl taint node "$NODE" lab.example.com/dedicated=pending-lab:NoSchedule
```

Check that both are present:

```sh
kubectl get node "$NODE" --show-labels
kubectl describe node "$NODE" | grep -A5 Taints
```

Now create a Pod that selects that node but has no toleration. It should stay Pending because the node repels it:

```sh
kubectl apply -f k8s/03-taint-no-toleration.yaml
kubectl -n pending-lab describe pod taint-pending
```

Delete it, then create the matching toleration Pod:

```sh
kubectl -n pending-lab delete pod taint-pending
kubectl apply -f k8s/04-taint-tolerated.yaml
kubectl -n pending-lab get pod taint-tolerated -o wide -w
```

A matching toleration permits the Pod onto a tainted node; it does not attract it. This manifest also uses a node selector so the Pod is directed to the labeled lab node. Once it is Running, verify its node and app:

```sh
kubectl -n pending-lab get pod taint-tolerated -o wide
kubectl -n pending-lab port-forward pod/taint-tolerated 8080:8080
```

Remove the test Pod, then restore the node:

```sh
kubectl -n pending-lab delete pod taint-tolerated
kubectl taint node "$NODE" lab.example.com/dedicated=pending-lab:NoSchedule-
kubectl label node "$NODE" lab.example.com/dedicated-
unset NODE
```

## 8. Review node pressure and capacity

Use these commands when a real Pod is Pending or a node is under pressure:

```sh
kubectl get nodes
kubectl describe nodes
kubectl get events -A --sort-by=.lastTimestamp
kubectl top nodes
kubectl top pods -A --sort-by=cpu
kubectl top pods -A --sort-by=memory
```

In `kubectl describe node`, inspect `Conditions`, `Taints`, `Allocatable`, and `Allocated resources`. A `MemoryPressure` or `DiskPressure` condition is a node health signal; Kubernetes may add a corresponding taint. Do not manually add Kubernetes' built-in pressure taints to simulate pressure. Use the deliberately oversized Pod to practice the scheduler-capacity path instead.

Quick diagnosis:

- `FailedScheduling` + `Insufficient cpu` or `Insufficient memory`: compare requests against allocatable resources and already allocated requests; reduce requests, free capacity, or add capacity.
- Selector/affinity mismatch: check node labels and correct the Pod selector.
- Untolerated taint: identify the taint and decide whether the workload should tolerate it; do not add broad tolerations just to force placement.
- `Pending` with a node assigned: scheduling succeeded; inspect container events, image pulls, volumes, and probes instead.
- `OOMKilled` after start: inspect container memory use and its limit; that differs from a scheduler `FailedScheduling` event.

## Cleanup

Remove the lab app and namespace when finished:

```sh
kubectl delete -f k8s/app.yaml --ignore-not-found
kubectl delete namespace pending-lab --ignore-not-found
```

Confirm the node taint and label were removed before ending the lab:

```sh
kubectl get nodes --show-labels
kubectl describe nodes | grep -A5 Taints
```

## Official Kubernetes references

- [Debug Pods](https://kubernetes.io/docs/tasks/debug/debug-application/debug-pods/)
- [Resource Management for Pods and Containers](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/)
- [Assigning Pods to Nodes](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/)
- [Taints and Tolerations](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/)
- [Resource Metrics Pipeline](https://kubernetes.io/docs/tasks/debug/debug-cluster/resource-metrics-pipeline/)
