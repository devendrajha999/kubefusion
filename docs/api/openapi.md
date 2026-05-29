# API Spec

Base path: `/api/v1`

## Login
`POST /auth/login`

Request:
```json
{"username":"admin","password":"admin"}
```

Response:
```json
{"token":"<jwt>","role":"admin"}
```
