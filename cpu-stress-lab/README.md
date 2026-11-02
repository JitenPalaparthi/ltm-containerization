# CPU Stress Lab (Go)

A small, bounded CPU load generator. It starts one Go worker per configured worker, reports progress once per second, and exits after a chosen duration. It has no third-party Go dependencies.

This is a CPU load demonstration, not a performance benchmark. It intentionally keeps CPU busy while it runs. Start with a short duration and monitor temperature and system responsiveness.

## Requirements

- Go 1.24+ for local runs
- Docker for container builds
- A Kubernetes cluster and `kubectl` for the Kubernetes Job

## Run locally

From this directory:

```sh
go run .
```

Default run: 30 seconds, using Go's default parallelism. To request all 8 logical processors and run for one minute:

```sh
go run . -workers=8 -duration=60s
```

Stop early with Ctrl+C. You can also use environment variables:

```sh
WORKERS=8 DURATION=60s go run .
```

`WORKERS` sets the number of CPU-bound goroutines. `-workers` overrides it. `DURATION=0` runs until stopped by Ctrl+C or a termination signal.

## Build and run with Docker

Build the image:

```sh
docker build -t cpu-stress-lab:local .
```

Run for 60 seconds with 8 workers:

```sh
docker run --rm --cpus=8 -e WORKERS=8 -e DURATION=60s cpu-stress-lab:local
```

If your machine or Docker Desktop VM has fewer than 8 CPUs available, lower both `--cpus` and `WORKERS` to the available amount. Docker's `--cpus` sets a CPU quota; it does not create extra processors.

## Run as a Kubernetes Job

The manifest requests and limits 8 CPUs and runs for 60 seconds. A cluster needs enough available CPU for the Pod to schedule. Adjust `WORKERS`, `resources.requests.cpu`, and `resources.limits.cpu` together if your cluster has less capacity.

Build the image, then make it available to your cluster.

### Docker Desktop Kubernetes

```sh
docker build -t cpu-stress-lab:local .
kubectl apply -f k8s/job.yaml
kubectl logs -f job/cpu-stress-lab
```

### Minikube

```sh
docker build -t cpu-stress-lab:local .
minikube image load cpu-stress-lab:local
kubectl apply -f k8s/job.yaml
kubectl logs -f job/cpu-stress-lab
```

### kind

```sh
docker build -t cpu-stress-lab:local .
kind load docker-image cpu-stress-lab:local
kubectl apply -f k8s/job.yaml
kubectl logs -f job/cpu-stress-lab
```

Check status and output:

```sh
kubectl get job,pods -l app=cpu-stress-lab
kubectl logs job/cpu-stress-lab
```

The manifest removes the completed Job about five minutes after it finishes. To clean it up immediately:

```sh
kubectl delete -f k8s/job.yaml
```

If the Pod stays `Pending`, inspect its events with `kubectl describe pod -l app=cpu-stress-lab`; the cluster may not have 8 CPUs available. Lower the CPU values in `k8s/job.yaml` and `WORKERS` to fit.

## Inspect the compiled machine instructions

The program cannot print every hardware instruction the processor supports. To view the architecture-specific assembly generated for this program on your build machine, run:

```sh
go build -gcflags='-S' . 2> assembly.txt
grep -A40 'main.work' assembly.txt
```

This prints the compiler's assembly for the `work` function. If your Go installation includes the `objdump` tool, you can also inspect the final binary:

```sh
go build -o cpu-stress-lab .
go tool objdump -s 'main.work' ./cpu-stress-lab
```

These show instructions compiled for this program; they are not a complete list of every instruction the CPU supports. A container image uses Linux machine code; on macOS, build locally to inspect instructions for macOS/your local CPU architecture.
