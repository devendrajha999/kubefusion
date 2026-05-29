# KubeFusion

Primary install document: `docs/INSTALL_K8S.md`

Quick start:
1. Build and push backend/frontend images.
2. Apply namespace, postgres, redis, config, RBAC, backend, frontend manifests.
3. Configure ingress host + DNS.
4. Login (`/api/v1/auth/login`) and use token for API/UI.

Phase status:
- GitOps baseline (create/sync/drift/rollback/history)
- Cluster operations (nodes/pods/logs/exec)
- Live logs streaming (SSE)
- JWT auth + RBAC
- PostgreSQL write path + read path fallback
