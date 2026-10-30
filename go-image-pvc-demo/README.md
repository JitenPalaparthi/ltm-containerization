# Go image uploader with persistent Kubernetes storage (KIND)

A small Go HTTP application with a web form to upload JPG/PNG/GIF/WEBP images. Images are stored on a PersistentVolumeClaim (PVC) at `/data/images`. **Deleting the application Deployment does not delete the PVC or images.** No Go dependencies beyond the standard library.

## Requirements

- Docker running
- KIND and kubectl installed
- A KIND cluster named `kind` (create with `kind create cluster --name kind` if needed)

## 1. Build and load the image

From the project directory:

```bash
docker build -t go-image-pvc-demo:v1 .
kind load docker-image go-image-pvc-demo:v1 --name kind
```

For a different cluster name, replace `kind` after `--name`.

## 2. Create persistent storage (only once)

```bash
kubectl apply -f k8s/01-storage.yaml
kubectl get pv,pvc
```

Wait for `images-pvc` to be `Bound`. This demo uses a static hostPath PV on the KIND node, with a `Retain` reclaim policy. **Do not delete the PVC during the demonstration.**

## 3. Deploy the Go app

```bash
kubectl apply -f k8s/02-app.yaml
kubectl rollout status deployment/image-uploader
kubectl get pods,pv,pvc,svc
kubectl port-forward svc/image-uploader 8080:80
```

Open **http://localhost:8080** in your browser. Upload one or more pictures and confirm they appear on the page.

You can also upload with curl:

```bash
curl -F 'image=@/path/to/photo.jpg' http://localhost:8080/upload
```

## 4. Remove the application, NOT the persistent storage

Stop port-forward with Ctrl+C, then:

```bash
kubectl delete -f k8s/02-app.yaml
kubectl get pvc images-pvc
kubectl get pv images-pv
```

The Deployment, Pod and Service are gone; the PVC/PV remain. The pictures are still on the KIND node.

Inspect files directly inside the KIND node:

```bash
docker exec kind-control-plane ls -lah /var/local/kind-image-demo
```

## 5. Redeploy and prove images persist

```bash
kubectl apply -f k8s/02-app.yaml
kubectl rollout status deployment/image-uploader
kubectl port-forward svc/image-uploader 8080:80
```

Open **http://localhost:8080** again: previously uploaded images should still appear.

## 6. Troubleshooting

```bash
kubectl get pods -o wide
kubectl describe pod -l app=image-uploader
kubectl logs deployment/image-uploader
kubectl describe pvc images-pvc
kubectl get pv,pvc
```

If upload fails with `permission denied`, note that hostPath permissions on a KIND node can vary. For this local demonstration, run:

```bash
docker exec -u root kind-control-plane sh -c 'mkdir -p /var/local/kind-image-demo && chmod 0777 /var/local/kind-image-demo'
kubectl rollout restart deployment/image-uploader
```

`chmod 0777` is for a disposable local demo only; use an appropriate storage driver and restricted permissions in production.

## 7. Cleanup (DELETES DATA if you remove node directory)

```bash
kubectl delete -f k8s/02-app.yaml --ignore-not-found
kubectl delete -f k8s/01-storage.yaml --ignore-not-found
# Optional: manually delete stored images from the KIND node:
# docker exec -u root kind-control-plane rm -rf /var/local/kind-image-demo
```

### Persistence scope

- Deleting/recreating the **Deployment**: images survive.
- Deleting/recreating a **Pod**: images survive.
- Deleting the **PVC**: the PV uses `Retain`, but manual recovery/rebinding may be required; don't delete the PVC in this exercise.
- Deleting the **KIND cluster**: the KIND node container and its hostPath storage are normally destroyed. For data that survives KIND cluster deletion, use an `extraMounts` configuration to bind a Mac host directory into the KIND node, or an external storage provider.

### Design notes

- Single replica only: `ReadWriteOnce` and `Recreate` strategy keep this demo simple.
- `hostPath` is node-local, not production-grade shared storage.
- Files uploaded through this application are never deleted by the application itself. Deleting an image from inside the volume would remove it from persistent storage too.
- Maximum upload is approximately 10 MiB; image types are validated by content sniffing.
