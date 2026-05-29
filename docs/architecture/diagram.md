# Architecture Diagram

```text
[User Browser]
   |
[Frontend React]
   |
[Backend API (REST/gRPC)] -- [Redis]
   |            |
   |            +-- [PostgreSQL]
   |
[Kubernetes client-go]
   |
[Managed Clusters]
```
