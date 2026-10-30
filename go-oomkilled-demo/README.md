# Go OOMKilled Demo (KIND + Docker)

A self-contained Go application that **retains 150 MiB of touched heap memory**. The first Kubernetes manifest limits its container to **80 MiB**, so Linux cgroup memory enforcement should kill the process (`OOMKilled`, typically exit code `137`). The second manifest provides a **256 MiB** limit and a **192 MiB** request so the same bounded workload can run.

## Prerequisites
- Docker, KIND, kubectl
- A running KIND cluster (default cluster name: `kind`)

## Run
```bash
chmod +x build-and-load.sh
./build-and-load.sh kind
kubectl apply -f k8s/01-oom.yaml
kubectl get pods -l app=go-oom-demo -w
```

In a separate terminal:
```bash
kubectl get pods -l app=go-oom-demo
kubectl describe pod -l app=go-oom-demo
kubectl logs deployment/go-oom-demo --previous
kubectl get pods -l app=go-oom-demo -o jsonpath='{range .items[*]}{.metadata.name}{" reason="}{.status.containerStatuses[0].lastState.terminated.reason}{" exit="}{.status.containerStatuses[0].lastState.terminated.exitCode}{" restarts="}{.status.containerStatuses[0].restartCount}{"\n"}{end}'
```

Expect `OOMKilled`, `137`, and repeated restarts. `CrashLoopBackOff` may follow. Timing varies. If the process is currently running, the last termination state can still show `OOMKilled`.

## Correct the resource configuration
```bash
kubectl apply -f k8s/02-fixed.yaml
kubectl rollout status deployment/go-oom-demo
kubectl get pods -l app=go-oom-demo
kubectl logs deployment/go-oom-demo
```

Expected log: `Target reached; keeping memory resident...` and no new OOMKilled events for the new Pod. **The corrected deployment replaces the previous Pod**, so inspect the new Pod's restart count.

## Cleanup
```bash
kubectl delete deployment go-oom-demo
```

## Run in Docker alone (optional)
```bash
docker run --rm --memory=80m --memory-swap=80m -e GOGC=off go-oom-demo:v1
# Docker may show exit 137 if the kernel kills the process.
docker run --rm --memory=256m --memory-swap=256m -e GOGC=off go-oom-demo:v1
```

## Concepts
- `requests.memory`: scheduler resource requirement; not a cap.
- `limits.memory`: cgroup-enforced memory ceiling.
- Exit code 137 means SIGKILL, **not necessarily OOM**; confirm the termination reason.
- This demo is **bounded** memory consumption, not a leak. Raising the limit is suitable only when the workload legitimately needs that memory.
- This is a deliberate resource-exhaustion test. Run only in an isolated training cluster.
