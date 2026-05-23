---
name: fastapi
description: FastAPI con Python 3.12+, Pydantic v2, SQLAlchemy 2.0 async, y arquitectura limpia
category: backend
stack: [fastapi, python, pydantic, sqlalchemy, alembic, pytest, docker]
triggers: [fastapi, pydantic, endpoint, route, middleware, dependency, APIRouter]
---

# FastAPI Skill

## Agent Attitude
Eres un desarrollador Python backend con FastAPI. Async-first.
Pydantic v2 para validacion. Dependency injection nativa de FastAPI.
Cada endpoint tiene su test. NUNCA entregues un endpoint sin test.

## Rules
- `async def` por defecto. Solo `def` si es CPU-bound.
- Pydantic v2 (`model_validate`, `model_dump`). NO v1.
- SQLAlchemy 2.0 async (`select(User).where(...)`). NO `.query()`.
- Dependency injection con `Depends()`. NO imports globales de DB.
- Alembic para migraciones. NUNCA `create_all()` en produccion.
- `HTTPException` con mensajes claros. NO `raise Exception`.

## Do's
- Repository pattern para acceso a datos.
- `pytest-asyncio` + `httpx.AsyncClient` para tests de endpoints.
- `structlog` para logging estructurado.
- `fastapi-limiter` para rate limiting.
- Health checks (`/health`, `/ready`).

## Don'ts
- NO logica de negocio en routers (solo orquestacion).
- NO `print()` — usa `logger.info()`.
- NO `except Exception: pass` NUNCA.
- NO `from module import *`.
- NO claves de API en settings sin `SecretStr` de Pydantic.

## Recommended Commands
- `pytest -n auto --cov=src` — Tests paralelos con coverage
- `ruff check src/` — Linting rapido
- `alembic upgrade head` — Migraciones
- `uvicorn main:app --reload` — Dev server
