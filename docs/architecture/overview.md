# Architecture

KubeFusion has these runtime components:
- `kubefusion-server`: REST API, authn/authz, GitOps orchestration.
- `kubefusion-controller`: Kubernetes watcher/sync engine (module scaffolded).
- `postgres`: source of truth for applications, projects, audit, history.
- `redis`: caching, session coordination, pub/sub notifications.
- `frontend`: UI and WebSocket clients.
