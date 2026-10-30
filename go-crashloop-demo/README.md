# Go CrashLoopBackOff Kubernetes Lab

This application intentionally exits with status 1 after **45 seconds** (configurable). Kubernetes restarts the container because Deployments use `restartPolicy: Always`; after repeated failures, restart backoff increases and the pod can show `CrashLoopBackOff`. The precise time before the status appears varies by cluster.

## Prerequisites
- Docker Desktop running
- kubectl and an accessible Kubernetes cluster (KIND, Minikube, or Docker Desktop)
- Docker Hub account if pushing to Docker Hub

## 1. Build on Mac (works on Apple Silicon too)

```bash
docker build --platform linux/amd64 -t jpalaparthi/go-crash-demo:v1 .
```

**Architecture:** The command above produces amd64 images. For ARM64 Kubernetes nodes use `--platform linux/arm64` instead. For a multi-platform Docker Hub image, use `docker buildx build --platform linux/amd64,linux/arm64 -t jpalaparthi/go-crash-demo:v1 --push .` after logging in.

## 2. Test locally

```bash
docker run --rm -p 8080:8080 -e CRASH_AFTER_SECONDS=15 jpalaparthi/go-crash-demo:v1
# In another terminal before 15 seconds:
curl http://localhost:8080/
```

## 3. Push to Docker Hub

The sample assumes the Docker Hub username `jpalaparthi`. Replace it in the build/tag/push commands **and in k8s/deployment.yaml** if your username differs.

```bash
docker login
docker push jpalaparthi/go-crash-demo:v1
```

The Docker Hub repository must exist or be creatable under your account, and the cluster must have pull access (public repository or image pull credentials).

## 4. Select your cluster

```bash
kubectl config get-contexts
kubectl config use-context minikube
# or: kubectl config use-context docker-desktop
# or: kubectl config use-context kind-kind
```

To use a separate file instead:

```bash
kubectl --kubeconfig="$HOME/kind-config.yaml" get nodes
# Add --kubeconfig="$HOME/kind-config.yaml" to the following kubectl commands.
```

## 5. Deploy

```bash
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
kubectl -n crash-demo get pods -w
```

Initially the pod should run; after repeated intentional crashes the STATUS may become `CrashLoopBackOff`. Press Ctrl+C to exit watch.

## 6. Investigate the crash

```bash
kubectl -n crash-demo get pods -o wide
kubectl -n crash-demo get pods
kubectl -n crash-demo describe pod -l app=go-crash-demo
kubectl -n crash-demo logs deployment/go-crash-demo --previous
kubectl -n crash-demo get events --sort-by=.lastTimestamp
```

`--previous` shows the previous terminated container instance. If no prior instance exists yet, wait for a restart.

## 7. Reach the service while it is running

```bash
kubectl -n crash-demo port-forward svc/go-crash-demo 8080:80
# In another terminal:
curl http://localhost:8080/
```

During restarts the service may temporarily have no ready endpoints.

## 8. Fix the crash and observe recovery

Edit `k8s/deployment.yaml` and change `CRASH_AFTER_SECONDS` from `"45"` to `"86400"`, then:

```bash
kubectl apply -f k8s/deployment.yaml
kubectl -n crash-demo rollout status deployment/go-crash-demo
kubectl -n crash-demo get pods
```

The new pod runs for 24 hours before the intentional exit. For a **permanent fix**, remove the intentional `os.Exit(1)` logic from `main.go`, rebuild with a **new image tag**, push, and update the Deployment image.

## 9. Cleanup

```bash
kubectl delete namespace crash-demo
```

## Note on KIND and local images

If you want to skip Docker Hub and use a local KIND image, build for the KIND node architecture and run:

```bash
kind load docker-image jpalaparthi/go-crash-demo:v1 --name kind
```

Then apply the manifests. `IfNotPresent` permits using the locally loaded image. If your KIND cluster has a different name, replace `kind`.
