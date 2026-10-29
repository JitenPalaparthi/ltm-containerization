kubectl version

kubectl api-resources

kubectl cluster-info 

kubectl get nodes

kubectl get all -A

kubectl get pods -A

kubectl apply -f node-demo-pod.yaml

kubectl get all -A

kubectl apply -f dev-ns.yaml

kubectl apply -f node-demo-deployment.yaml

kubectl apply -f node-demo-service.yaml

kubectl get all -n dev

kubectl scale deployment node-demo --replicas=20 -n dev

