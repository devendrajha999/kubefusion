#!/usr/bin/env bash
set -euo pipefail
kubectl apply -f deploy/k8s/base/rbac.yaml
kubectl apply -f deploy/k8s/base/backend.yaml
kubectl apply -f deploy/k8s/base/frontend.yaml
