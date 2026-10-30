# Go Kubernetes Probe Troubleshooting Lab

Four independent demos using **one simple Go HTTP server** and four Kubernetes Deployments.

## Requirements
- Docker, kubectl, a running Kubernetes cluster (KIND, Minikube or Docker Desktop)
- Docker Hub account if pushing publicly

## Build and push

```bash
docker build -t jpalaparthi/go-probe-lab:v1 .
docker login
docker push jpalaparthi/go-probe-lab:v1
```

Replace `jpalaparthi` in the image tag and all YAML files if you use another Docker Hub account. To use KIND without pushing, run `kind load docker-image jpalaparthi/go-probe-lab:v1 --name kind` (change cluster name if needed). For Minikube, use `minikube image load jpalaparthi/go-probe-lab:v1`.

## Deploy each example

```bash
kubectl config current-context
kubectl apply -f k8s/00-namespace.yaml
kubectl apply -f k8s/01-startup-probe-failure.yaml
kubectl apply -f k8s/02-readiness-failure.yaml
kubectl apply -f k8s/03-liveness-failure.yaml
kubectl apply -f k8s/04-app-crash.yaml
kubectl get pods -n probe-lab -w
```

To run just one example, apply namespace and the matching numbered YAML file.

## What to observe

| Example | Failure | Result | Root cause |
|---|---|---|---|
| 01 startup-failure | `/startup` returns 503 for 45s; probe allows ~15s | Repeated kubelet restarts; eventually CrashLoopBackOff | Startup timeout too short |
| 02 readiness-failure | `/ready` returns 503 | Running, 0/1 Ready, **no restart** | Readiness failure alone does NOT cause CrashLoopBackOff |
| 03 liveness-failure | `/live` returns 503 after 20s | Kubelet restarts; eventually CrashLoopBackOff | Liveness probe failure |
| 04 app-crash | Go process calls `os.Exit(1)` after 15s | Restart and eventually CrashLoopBackOff | Application failure |

## Diagnose (repeat with each deployment name)

```bash
kubectl get pods -n probe-lab -o wide
kubectl describe deployment startup-failure -n probe-lab
kubectl describe pods -n probe-lab -l app=startup-failure
kubectl logs -n probe-lab deployment/startup-failure --timestamps
kubectl logs -n probe-lab deployment/startup-failure --previous --timestamps
kubectl get events -n probe-lab --sort-by=.lastTimestamp
```

For app crash:
```bash
kubectl logs -n probe-lab deployment/app-crash --previous
kubectl describe pods -n probe-lab -l app=app-crash
```

For readiness failure:
```bash
kubectl get endpointslices -n probe-lab -l kubernetes.io/service-name=readiness-failure
kubectl describe pods -n probe-lab -l app=readiness-failure
```

## Exit codes
- **1**: application terminated with an error (demo 04).
- **137**: SIGKILL; often OOMKilled, but inspect `Last State`, `Reason` and events before concluding.
- **143**: SIGTERM; may occur when kubelet terminates container after failed startup/liveness probe, or during normal termination.
- **0**: process exited successfully; a Deployment may still restart it because restartPolicy is Always.

`CrashLoopBackOff` is a **waiting status**, not an exit code. For probe-triggered restarts, the application may have no stack trace; inspect **Events** and **Last State**. `kubectl logs --previous` is available only when a previous terminated container instance exists and its logs remain accessible.

## Fixes

**Startup**: In `01-startup-probe-failure.yaml`, change `failureThreshold: 3` to `failureThreshold: 12` (about 60 seconds at 5-second intervals), then `kubectl apply -f ...`.

**Readiness**: In `02-readiness-failure.yaml`, change `READY_AFTER_SECONDS` from `999999` to `2`, then reapply. Readiness recovers without restarting the container.

**Liveness**: In `03-liveness-failure.yaml`, set `FAIL_AFTER_SECONDS` to `0` and reapply.

**App crash**: In `04-app-crash.yaml`, set `CRASH_AFTER_SECONDS` to `0` and reapply.

## Cleanup
```bash
kubectl delete namespace probe-lab
```

## HTTP endpoints
- `/` returns uptime
- `/startup` returns 503 until STARTUP_SECONDS
- `/ready` returns 503 until READY_AFTER_SECONDS (or after FAIL_AFTER_SECONDS)
- `/live` returns 503 after FAIL_AFTER_SECONDS when nonzero

All delays are configurable via environment variables. This lab deliberately uses HTTP 503 responses to teach probe behavior.
