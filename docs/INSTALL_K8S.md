# KubeFusion Kubernetes Installation Guide

This guide deploys KubeFusion into namespace `gitops-platform`.

## Step 1: Prerequisites
- Kubernetes cluster (v1.28+)
- `kubectl` configured to target cluster
- Docker (or build system) for images
- Ingress controller (NGINX recommended)
- Optional: cert-manager for TLS

Validate:
```bash
kubectl version --short
kubectl get nodes
```

## Step 2: Build Containers
From project root:
```bash
docker build -f Dockerfile.backend -t kubefusion/backend:latest .
docker build -f Dockerfile.frontend -t kubefusion/frontend:latest .
```

Push to your registry:
```bash
docker tag kubefusion/backend:latest <REGISTRY>/kubefusion/backend:latest
docker tag kubefusion/frontend:latest <REGISTRY>/kubefusion/frontend:latest
docker push <REGISTRY>/kubefusion/backend:latest
docker push <REGISTRY>/kubefusion/frontend:latest
```

Update image references in:
- `deploy/k8s/base/backend.yaml`
- `deploy/k8s/base/frontend.yaml`

## Step 3: Deploy PostgreSQL
```bash
kubectl apply -f deploy/k8s/base/namespace.yaml
kubectl apply -f deploy/k8s/base/postgres.yaml
kubectl -n gitops-platform rollout status deploy/postgres
```

## Step 4: Deploy Redis
```bash
kubectl apply -f deploy/k8s/base/redis.yaml
kubectl -n gitops-platform rollout status deploy/redis
```

## Step 5: Deploy Backend
1) Create/update secrets:
```bash
kubectl apply -f deploy/k8s/base/config.yaml
kubectl -n gitops-platform create secret generic kubefusion-secrets \
  --from-literal=postgres_dsn='postgres://kubefusion:kubefusion@postgres:5432/kubefusion?sslmode=disable' \
  --from-literal=jwt_secret='CHANGE-THIS-TO-A-STRONG-SECRET' \
  --dry-run=client -o yaml | kubectl apply -f -
```

2) Apply RBAC + backend:
```bash
kubectl apply -f deploy/k8s/base/rbac.yaml
kubectl apply -f deploy/k8s/base/backend.yaml
kubectl -n gitops-platform rollout status deploy/kubefusion-backend
```

## Step 6: Deploy Frontend
```bash
kubectl apply -f deploy/k8s/base/frontend.yaml
kubectl -n gitops-platform rollout status deploy/kubefusion-frontend
```

## Step 7: Deploy Monitoring
If you already have Prometheus Operator:
```bash
kubectl apply -f deploy/k8s/monitoring/servicemonitor.yaml
```

If not, install kube-prometheus-stack:
```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
helm upgrade --install monitoring prometheus-community/kube-prometheus-stack -n monitoring --create-namespace
```

Import dashboard file:
- `grafana/dashboards/cluster-overview.json`

## Step 8: Configure Ingress
Update host in `deploy/k8s/base/frontend.yaml` (`kubefusion.local` by default), then apply:
```bash
kubectl apply -f deploy/k8s/base/frontend.yaml
kubectl -n gitops-platform get ingress kubefusion
```

Add DNS record for your host to ingress controller address.

## Step 9: Configure Authentication
Current implementation supports local JWT login.

Login API:
```bash
curl -X POST http://<KUBEFUSION_HOST>/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"admin"}'
```

Save `token` from response and use:
```bash
-H "Authorization: Bearer <TOKEN>"
```

## Step 10: Register First Cluster
Current backend assumes in-cluster access using service account. Verify cluster APIs:
```bash
curl -H "Authorization: Bearer <TOKEN>" http://<KUBEFUSION_HOST>/api/v1/clusters
curl -H "Authorization: Bearer <TOKEN>" http://<KUBEFUSION_HOST>/api/v1/clusters/in-cluster/nodes
```

## Step 11: Create First GitOps Application
```bash
curl -X POST http://<KUBEFUSION_HOST>/api/v1/applications \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer <TOKEN>" \
  -d '{
    "name":"demo-app",
    "project":"default",
    "repoUrl":"https://github.com/example/repo.git",
    "path":"manifests",
    "targetRevision":"main",
    "destination":"in-cluster",
    "namespace":"default",
    "syncPolicy":"manual"
  }'
```

Manual sync:
```bash
curl -X POST http://<KUBEFUSION_HOST>/api/v1/applications/<APP_ID>/sync -H "Authorization: Bearer <TOKEN>"
```

## Step 12: Verify Metrics Collection
```bash
kubectl -n gitops-platform get svc kubefusion-backend
kubectl -n monitoring get servicemonitors | grep kubefusion
```

In Grafana/Prometheus verify target scrape and dashboard data.

## Step 13: Verify Pod Shell Access (Exec)
```bash
curl -X POST http://<KUBEFUSION_HOST>/api/v1/clusters/in-cluster/pods/exec \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer <TOKEN>" \
  -d '{
    "namespace":"default",
    "pod":"<POD_NAME>",
    "container":"<CONTAINER_NAME>",
    "command":["/bin/sh","-c","ls -la"]
  }'
```

## Step 14: Verify Logs
One-shot logs:
```bash
curl -X POST http://<KUBEFUSION_HOST>/api/v1/clusters/in-cluster/pods/logs \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer <TOKEN>" \
  -d '{"namespace":"default","pod":"<POD_NAME>","tailLines":100}'
```

Streaming logs (SSE):
```bash
curl -N "http://<KUBEFUSION_HOST>/api/v1/clusters/in-cluster/pods/logs/stream?namespace=default&pod=<POD_NAME>" \
  -H "Authorization: Bearer <TOKEN>"
```

## Step 15: Production Hardening
- Replace default secrets and rotate regularly.
- Enable TLS for ingress and internal service-to-service traffic.
- Use managed PostgreSQL + Redis with backups and HA.
- Apply strict network policies.
- Restrict RBAC scope to least privilege.
- Add PodDisruptionBudgets, resource limits, and HPA.
- Enable external secret manager integration.
- Configure persistent audit retention and SIEM forwarding.

## Optional: Helm Install
```bash
helm upgrade --install kubefusion deploy/helm/kubefusion -n gitops-platform --create-namespace \
  --set image.backend=<REGISTRY>/kubefusion/backend:latest \
  --set image.frontend=<REGISTRY>/kubefusion/frontend:latest
```

## Post-Install Checks
```bash
kubectl -n gitops-platform get all
kubectl -n gitops-platform get ingress
kubectl -n gitops-platform logs deploy/kubefusion-backend --tail=200
kubectl -n gitops-platform logs deploy/kubefusion-frontend --tail=200
```
