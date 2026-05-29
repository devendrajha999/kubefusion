# Installation Guide

## Step 1: Prerequisites
- Kubernetes 1.28+
- Docker
- kubectl
- Helm 3

## Step 2: Build containers
- `docker build -f Dockerfile.backend -t kubefusion/backend:latest .`
- `docker build -f Dockerfile.frontend -t kubefusion/frontend:latest .`

## Step 3: Deploy PostgreSQL
- `kubectl apply -f deploy/k8s/base/postgres.yaml`

## Step 4: Deploy Redis
- `kubectl apply -f deploy/k8s/base/redis.yaml`

## Step 5: Deploy Backend
- `kubectl apply -f deploy/k8s/base/namespace.yaml`
- `kubectl apply -f deploy/k8s/base/config.yaml`
- `kubectl apply -f deploy/k8s/base/rbac.yaml`
- `kubectl apply -f deploy/k8s/base/backend.yaml`

## Step 6: Deploy Frontend
- `kubectl apply -f deploy/k8s/base/frontend.yaml`

## Step 7: Deploy Monitoring
- `kubectl apply -f deploy/k8s/monitoring/servicemonitor.yaml`
- Import `grafana/dashboards/cluster-overview.json`

## Step 8: Configure Ingress
- Update host in `deploy/k8s/base/frontend.yaml`

## Step 9: Configure Authentication
- Set JWT and external IdP configs in `kubefusion-secrets`

## Step 10: Register First Cluster
- Use `/api/v1/clusters` registration extension endpoint.

## Step 11: Create First GitOps Application
- `POST /api/v1/applications`

## Step 12: Verify Metrics Collection
- Check Prometheus targets and Grafana panel data

## Step 13: Verify Pod Shell Access
- Use backend exec capability extension endpoint

## Step 14: Verify Logs
- Verify logs stream in Logs view

## Step 15: Production Hardening
- Enforce TLS, external secrets, network policies, and RBAC least privilege.
