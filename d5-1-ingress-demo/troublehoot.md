# Troubleshooting kind Ingress and Services

Check the request path in order: **Pod -> Service -> Ingress -> host port**. This isolates where routing breaks.

## 1. Confirm the cluster and provider

```bash
kubectl config current-context
kind get clusters
kubectl get nodes
kubectl get ingress path-routing -o wide
```

- Make sure kubectl is using the intended kind context, such as `kind-ingress-demo`.
- The Ingress `ADDRESS` should appear after the provider processes it.
- kind's native Ingress support requires cloud-provider-kind v0.9.0 or later. Keep the provider process running in its terminal.
- On macOS, run the provider with `sudo`; on Windows, use an Administrator PowerShell window.

Official guide: [kind Ingress](https://kind.sigs.k8s.io/docs/user/ingress/).

## 2. Check the Pods

```bash
kubectl get pods -o wide
kubectl describe pod -l app=foo
kubectl describe pod -l app=bar
kubectl logs deployment/foo-app
kubectl logs deployment/bar-app
```

Expected: each Deployment has a Pod in `Running` and `Ready` state. If a Pod is not ready, inspect its Events and logs before troubleshooting its Service or Ingress.

## 3. Check the Services

```bash
kubectl get service foo-service -o yaml
kubectl get service bar-service -o yaml
kubectl get endpointslice -l kubernetes.io/service-name=foo-service
kubectl get endpointslice -l kubernetes.io/service-name=bar-service
```

Check each Service:

- `spec.selector` matches its Pod labels (`app: foo` or `app: bar`).
- `port` is the port clients and Ingress use (`8080` here).
- `targetPort` is the port the container listens on (`8080` here).
- The EndpointSlice lists one or more ready Pod addresses. An empty EndpointSlice usually means the selector matches no ready Pods.

Test a Service without Ingress. Run this in one terminal:

```bash
kubectl port-forward service/foo-service 8081:8080
```

Then in another terminal:

```bash
curl http://localhost:8081
```

If this fails, fix the Pod or Service first. Kubernetes' Service guide recommends checking selectors, ports, EndpointSlices, and whether the Pods respond directly: [Debug Services](https://kubernetes.io/docs/tasks/debug/debug-application/debug-service/).

## 4. Check the Ingress rules

```bash
kubectl get ingress path-routing -o yaml
kubectl describe ingress path-routing
kubectl get events --sort-by=.metadata.creationTimestamp
```

Verify that:

- Host is `demo.local`.
- `/foo` points to `foo-service`, and `/bar` points to `bar-service`.
- Both backend port numbers are the Service port `8080`, not a different container port.
- The Ingress has an address and its rules match the hostname and path sent by curl.
- The example leaves `ingressClassName` unset, matching the kind native Ingress example. If you add a class, make sure the installed controller handles that class.

## 5. Check the macOS or Windows host port

On macOS and Windows, kind nodes are inside Docker Desktop's Linux VM. cloud-provider-kind maps the Ingress port to a host port. Find the mapping:

macOS Terminal:

```bash
docker ps --filter name=kindccm --format 'table {{.Names}}\t{{.Ports}}'
```

Windows PowerShell:

```powershell
docker ps --filter "name=kindccm" --format "table {{.Names}}`t{{.Ports}}"
```

If Docker shows `0.0.0.0:42381->80/tcp`, the host port is `42381`. Use your actual port and include the expected Host header:

macOS:

```bash
PORT=42381
curl -v -H 'Host: demo.local' "http://localhost:$PORT/foo"
curl -v -H 'Host: demo.local' "http://localhost:$PORT/bar"
```

Windows PowerShell:

```powershell
$port = 42381
curl.exe -v -H "Host: demo.local" "http://localhost:$port/foo"
curl.exe -v -H "Host: demo.local" "http://localhost:$port/bar"
```

Replace `42381` with your Docker port mapping. The host header matters because the Ingress rule is for `demo.local`; requesting `localhost` without that header may not match the rule.

## Common symptoms

| Symptom | Check |
|---|---|
| Ingress `ADDRESS` is empty | Provider process is running; kind context is correct; cloud-provider-kind is v0.9.0 or later. |
| HTTP 404 | Host header and path match the Ingress rule (`demo.local`, `/foo` or `/bar`). |
| HTTP 502 or 503 | Service has ready EndpointSlice addresses; Service selector matches Pod labels; `port` and `targetPort` are correct. |
| Connection refused or timeout | Provider is still running; use the host port shown by `docker ps`; Pods and Services are healthy. |
| EndpointSlice is empty | Compare Service selector to Pod labels and check Pod readiness. |
| `kubectl port-forward` works but Ingress does not | Focus on Ingress host/path/backend settings, provider status, and host port mapping. |

For kind's macOS/Windows port mapping, the provider documentation notes that the Envoy container uses an ephemeral host port. Use the mapped port from Docker rather than assuming port 80 on localhost: [cloud-provider-kind](https://github.com/kubernetes-sigs/cloud-provider-kind).
