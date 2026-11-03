# Kubernetes Ingress Routing on kind (macOS and Windows)

This walkthrough creates a kind cluster and routes two HTTP paths through one Kubernetes Ingress:

- `http://demo.local/foo` -> `foo-service` -> `foo-app`
- `http://demo.local/bar` -> `bar-service` -> `bar-app`

## How Ingress works on kind now

Kubernetes Ingress needs an implementation to apply its routing rules. kind's current guide uses **cloud-provider-kind v0.9.0 or newer**, which provides native Ingress support. This walkthrough uses that instead of ingress-nginx, which was retired in March 2026. Keep the cloud-provider-kind process running while you test. On macOS, run it with `sudo`; on Windows, run it from an Administrator PowerShell window. On both platforms, `--enable-lb-port-mapping` exposes the route through a host port.

## Requirements

- Docker Desktop running with Linux containers enabled
- Go installed and available as `go`
- `kind` and `kubectl` available in your terminal
- Internet access for container images and Go modules

Check the tools before continuing:

```text
docker version
go version
kind version
kubectl version --client
```

Official install pages: [Go](https://go.dev/doc/install), [kind](https://kind.sigs.k8s.io/docs/user/quick-start/), [kubectl for macOS](https://kubernetes.io/docs/tasks/tools/install-kubectl-macos/), [kubectl for Windows](https://kubernetes.io/docs/tasks/tools/install-kubectl-windows/).

## macOS setup

Install Docker Desktop, Go, kind, and kubectl. If you use Homebrew for the command line tools:

```bash
brew install kind kubectl
```

Install Go from [go.dev/dl](https://go.dev/dl/) if `go version` does not work. Start Docker Desktop and wait until its engine is running.

Create the cluster:

```bash
kind create cluster --name ingress-demo
kubectl config use-context kind-ingress-demo
kubectl get nodes
```

Install cloud-provider-kind:

```bash
go install sigs.k8s.io/cloud-provider-kind@latest
```

Open a **second Terminal window** and start it. Keep this command running while you test Ingress:

```bash
DOCKER_HOST="$(docker context inspect "$(docker context show)" --format '{{.Endpoints.docker.Host}}')"
sudo env KUBECONFIG="$HOME/.kube/config" DOCKER_HOST="$DOCKER_HOST" \
  "$(go env GOPATH)/bin/cloud-provider-kind" --enable-lb-port-mapping
```

If the provider says it cannot connect to Docker, check that Docker Desktop is running and that `docker version` shows both Client and Server sections.

## Windows setup (PowerShell)

Install Docker Desktop with Linux containers enabled. Install Go from [go.dev/dl](https://go.dev/dl/). In an Administrator PowerShell window, install kind and kubectl with winget:

```powershell
winget install Kubernetes.kind
winget install -e --id Kubernetes.kubectl
```

Close and reopen PowerShell after installation. Confirm `docker version` shows a Server section, then confirm these commands work:

```powershell
go version
kind version
kubectl version --client
```

Create the cluster:

```powershell
kind create cluster --name ingress-demo
kubectl config use-context kind-ingress-demo
kubectl get nodes
```

Install cloud-provider-kind:

```powershell
go install sigs.k8s.io/cloud-provider-kind@latest
```

Open another **Administrator PowerShell window** using the same Windows account. Set kubeconfig and start the provider; keep it running while you test:

```powershell
$env:KUBECONFIG = Join-Path $HOME ".kube\config"
$cloudProvider = Join-Path (go env GOPATH) "bin\cloud-provider-kind.exe"
& $cloudProvider --enable-lb-port-mapping
```

If `cloud-provider-kind.exe` is not found, inspect the Go binary folder with `go env GOPATH` and check that its `bin` directory is on your PATH, or run it using the full path shown by the command above.

## Create the routing example

Save the following as `ingress-routing.yaml` in your working directory. The same manifest is included as a separate file alongside this README.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: foo
  template:
    metadata:
      labels:
        app: foo
    spec:
      containers:
        - name: app
          image: registry.k8s.io/e2e-test-images/agnhost:2.39
          command: ["/agnhost", "serve-hostname", "--http=true", "--port=8080"]
          ports:
            - containerPort: 8080
---
apiVersion: v1
kind: Service
metadata:
  name: foo-service
spec:
  selector:
    app: foo
  ports:
    - port: 8080
      targetPort: 8080
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: bar-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: bar
  template:
    metadata:
      labels:
        app: bar
    spec:
      containers:
        - name: app
          image: registry.k8s.io/e2e-test-images/agnhost:2.39
          command: ["/agnhost", "serve-hostname", "--http=true", "--port=8080"]
          ports:
            - containerPort: 8080
---
apiVersion: v1
kind: Service
metadata:
  name: bar-service
spec:
  selector:
    app: bar
  ports:
    - port: 8080
      targetPort: 8080
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: path-routing
spec:
  rules:
    - host: demo.local
      http:
        paths:
          - path: /foo
            pathType: Prefix
            backend:
              service:
                name: foo-service
                port:
                  number: 8080
          - path: /bar
            pathType: Prefix
            backend:
              service:
                name: bar-service
                port:
                  number: 8080
```

## Apply and wait

In a terminal with the `kind-ingress-demo` context:

```bash
kubectl apply -f ingress-routing.yaml
kubectl rollout status deployment/foo-app
kubectl rollout status deployment/bar-app
kubectl get ingress path-routing -o wide
```

Wait for an address to appear in the Ingress `ADDRESS` column:

```bash
kubectl get ingress path-routing --watch
```

Press Ctrl+C after the address appears.

## Test from macOS or Windows

On macOS and Windows, kind nodes run inside Docker Desktop's Linux VM. The provider maps the Ingress port to an **ephemeral host port**. List the mapping:

```bash
docker ps --filter name=kindccm --format 'table {{.Names}}\t{{.Ports}}'
```

PowerShell equivalent:

```powershell
docker ps --filter "name=kindccm" --format "table {{.Names}}`t{{.Ports}}"
```

Find the entry that looks like `0.0.0.0:42381->80/tcp`. Here, `42381` is the host port; yours will likely differ. Set `PORT` to the number before `->80/tcp`, then send requests with the required hostname:

macOS Terminal:

```bash
PORT=42381
curl -H 'Host: demo.local' "http://localhost:$PORT/foo"
curl -H 'Host: demo.local' "http://localhost:$PORT/bar"
```

Windows PowerShell:

```powershell
$port = 42381
curl.exe -H "Host: demo.local" "http://localhost:$port/foo"
curl.exe -H "Host: demo.local" "http://localhost:$port/bar"
```

Replace `42381` with the actual host port shown by `docker ps`. The `/foo` response should contain a `foo-app` Pod name; `/bar` should contain a `bar-app` Pod name.

If no `kindccm` container or port mapping appears, confirm the provider terminal is still running, then inspect:

```bash
kubectl get ingress path-routing -o yaml
kubectl get pods,services
```

On Windows, run those `kubectl` commands in PowerShell. The provider may need a few moments after the Ingress is created to publish its address and create the port mapping.

## Cleanup

Remove the example resources:

```bash
kubectl delete -f ingress-routing.yaml
```

When finished with the whole cluster, stop the provider with Ctrl+C in its terminal and delete kind:

```bash
kind delete cluster --name ingress-demo
```

## Notes

- No `IngressClass` is set because the kind native example uses cloud-provider-kind's default Ingress support.
- The `Host: demo.local` header is sent by curl, so you do not need to edit the system hosts file.
- Ingress is stable but frozen; Kubernetes recommends Gateway API for new development. This is an Ingress learning example. See [Kubernetes Ingress documentation](https://kubernetes.io/docs/concepts/services-networking/ingress/).
- kind's current native Ingress setup is documented at [kind Ingress](https://kind.sigs.k8s.io/docs/user/ingress/) and [cloud-provider-kind](https://github.com/kubernetes-sigs/cloud-provider-kind).
