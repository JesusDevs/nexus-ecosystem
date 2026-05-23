---
name: docker-kubernetes
description: Container patterns — Docker multi-stage builds, Compose, K8s manifests, security
category: infra
stack: [docker, kubernetes, containers]
triggers: [docker, kubernetes, container, deploy, infra]
---

# Docker & Kubernetes Patterns

## Rules
- Multi-stage builds: build in one stage, run in another (minimal image)
- `.dockerignore` before `COPY .` — never copy `.git`, `node_modules`, `.venv`
- Run as non-root user in container
- Health checks for every service
- Secrets via env vars (dev) or K8s secrets (prod) — never baked into image

## Do's
- Pin base image versions: `FROM golang:1.22-alpine`, not `FROM golang:latest`
- Layer caching: copy dependency files first, install deps, then copy source
- Use `docker compose` for local dev, K8s for production
- Tag images with git SHA: `image:app:${GIT_SHA}`
- Signal handling: use `tini` or exec form `CMD ["app"]`

## Don'ts
- No `latest` tag in production
- No `COPY . .` before `go mod download` / `npm install`
- No hardcoded ports — use env vars
- No secrets in Dockerfile or docker-compose.yml
- No `docker compose up` for production — use K8s or managed service

## Recommended Commands
```bash
docker build -t app:$(git rev-parse --short HEAD) .
docker compose up -d
docker scan app:latest
kubectl apply -f k8s/
kubectl rollout status deployment/app
```
