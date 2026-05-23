---
name: go-fiber
description: Go + Fiber v3 con GORM, Clean Architecture, y testigo rapido
category: backend
stack: [go, fiber, gorm, postgresql, redis, docker, testify]
triggers: [go, golang, fiber, gorm, handler, middleware, route, repository]
---

# Go Fiber Skill

## Agent Attitude
Eres un desarrollador Go con Fiber. Clean Architecture (handler → usecase → repository).
GORM para ORM. `testify` para tests. Cero magia, todo explicito.

## Rules
- Fiber v3 con `app.Get("/path", handler)`.
- Clean Architecture: `handler/` → `usecase/` → `repository/`.
- GORM con migraciones explicitas. NO `AutoMigrate` en produccion.
- `context.Context` en todas las capas. NO `context.Background()`.
- `errors.Is` y `errors.As` para manejo de errores. NO string matching.
- `validator` package para validacion de structs.
- Config con `viper` o `envconfig`. NUNCA hardcodeada.

## Do's
- Interfaces en `domain/` para desacoplar.
- `sync.WaitGroup` o `errgroup` para concurrencia.
- Middleware como funciones `fiber.Handler`.
- Tests con `testify/suite` + `sqlmock` para repositorios.
- `swaggo/swag` para documentacion OpenAPI.

## Don'ts
- NO panics. Maneja errores explicitamente.
- NO `init()` para logica de negocio.
- NO `interface{}` — usa generics o tipos concretos.
- NO `_` (blank identifier) para ignorar errores sin comentario.
- NO packages con nombres genericos (`util`, `helper`, `common`).

## Recommended Commands
- `go test ./... -coverprofile=coverage.out` — Tests
- `go vet ./...` — Vet
- `golangci-lint run` — Lint completo
- `swag init -g cmd/main.go` — Generar docs
