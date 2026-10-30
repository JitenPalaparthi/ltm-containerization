# Python Picture Upload — Docker + Kind + Persistent Volume

This complete Flask application uploads and displays JPG, PNG, and WebP photos. Photos are stored in `/app/uploads` inside the container, which is mounted from a Kubernetes PersistentVolumeClaim (PVC). The PV points to `/data/pictures` on the Kind node, which is mapped to a folder on your laptop.

## Prerequisites

Docker, `kind`, and `kubectl` installed and running. Use a new cluster name to avoid changing an existing cluster.

## 1. Create Kind cluster with laptop-backed storage

Run these commands from the **project root**:

```bash
mkdir -p "$HOME/kind-pictures"
sed "s|__HOST_PICTURES_DIR__|$HOME/kind-pictures|g" kind-config.yaml > kind-config.local.yaml
kind create cluster --name pictures-cluster --config kind-config.local.yaml
kubectl config use-context kind-pictures-cluster
kubectl get nodes
```

The generated `kind-config.local.yaml` contains your actual absolute path. On macOS with Docker Desktop, make sure the directory is accessible to Docker Desktop.

## 2. Build and load Docker image

```bash
docker build -t picture-upload:v1 .
docker images picture-upload
kind load docker-image picture-upload:v1 --name pictures-cluster
```

Optional standalone Docker test (not Kubernetes):

```bash
docker run --rm -p 8081:8080 -v "$HOME/kind-pictures:/app/uploads" --name picture-test picture-upload:v1
# Open http://localhost:8081; Ctrl+C to stop
```

## 3. Deploy Kubernetes storage, application, and Service

```bash
kubectl apply -f k8s/pv.yaml
kubectl apply -f k8s/pvc.yaml
kubectl get pv,pvc
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
kubectl rollout status deployment/picture-upload
kubectl get pods,deploy,svc,pv,pvc
```

Open **http://localhost:8080** and upload some pictures. The Kind port mapping sends laptop port 8080 to the node's NodePort 30080, then to the Service and Pod.

Alternative access method: `kubectl port-forward svc/picture-service 8081:8080` and open http://localhost:8081.

## 4. Verify the pictures really live in the volume

```bash
ls -lah "$HOME/kind-pictures"
kubectl exec deployment/picture-upload -- ls -lah /app/uploads
kubectl logs deployment/picture-upload
```

## 5. Delete Deployment — pictures survive

```bash
kubectl delete deployment picture-upload
kubectl get pvc,pv
ls -lah "$HOME/kind-pictures"
```

The Service and PVC are separate objects; they remain. The pictures remain on your laptop.

## 6. Recreate Deployment — pictures return

```bash
kubectl apply -f k8s/deployment.yaml
kubectl rollout status deployment/picture-upload
```

Refresh http://localhost:8080. Previously uploaded pictures will reappear.

## 7. Optional: Delete the Kind cluster — laptop pictures still survive

```bash
kind delete cluster --name pictures-cluster
ls -lah "$HOME/kind-pictures"
```

To recreate the same demo:

```bash
kind create cluster --name pictures-cluster --config kind-config.local.yaml
kubectl config use-context kind-pictures-cluster
kind load docker-image picture-upload:v1 --name pictures-cluster
kubectl apply -f k8s/pv.yaml -f k8s/pvc.yaml -f k8s/deployment.yaml -f k8s/service.yaml
kubectl rollout status deployment/picture-upload
```

## Architecture

```text
Browser → localhost:8080 → Kind NodePort 30080 → Kubernetes Service
                                                ↓
                                           Python Pod
                                                ↓
                                         /app/uploads
                                                ↓
                                        pictures-pvc
                                                ↓
                                         pictures-pv
                                                ↓
                                Kind node /data/pictures
                                                ↓
                                 Laptop ~/kind-pictures
```

## Troubleshooting

- `ImagePullBackOff`: run `kind load docker-image picture-upload:v1 --name pictures-cluster`.
- PVC `Pending`: verify `kubectl get pv,pvc` and ensure both use storage class `manual` and `pictures-pv`.
- `CrashLoopBackOff`: run `kubectl logs deployment/picture-upload` and `kubectl describe pod -l app=picture-upload`.
- Permission errors: ensure the laptop directory is writable and mounted into the Kind node. Check `docker exec pictures-cluster-control-plane ls -ld /data/pictures`.
- Port 8080 already used: change `hostPort` in the Kind config **before** creating the cluster, or use `kubectl port-forward` instead.
- This is a **single-node development demo**. `hostPath` is not recommended for production; `ReadWriteOnce` is not a cross-node shared filesystem. For production use CSI-backed storage or object storage, add authentication and CSRF protection, configure a real secret key, and apply stricter upload and resource limits.
- PV `Retain` keeps underlying data after PVC deletion, but a released PV generally requires administrator intervention before binding a new claim. This demo deliberately retains its PVC when deleting the Deployment.
