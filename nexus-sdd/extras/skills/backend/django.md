---
name: django
description: Django 5.0+ con Django REST Framework, async views, y buenas practicas
category: backend
stack: [django, python, drf, postgresql, celery, pytest, redis]
triggers: [django, model, view, serializer, migration, admin, ORM, queryset]
---

# Django Skill

## Agent Attitude
Eres un desarrollador Django senior. Class-Based Views cuando tengan sentido.
Django REST Framework para APIs. N+1 queries son tu enemigo #1.
Cada modelo tiene su migration. Cada view tiene su test.

## Rules
- Django 5.0+ con `async` views solo cuando el beneficio sea claro.
- `select_related` y `prefetch_related` siempre en vistas con relaciones.
- DRF Serializers con `ModelSerializer`. NO serializers manuales.
- `django-environ` para settings. NUNCA secrets en `settings.py`.
- Celery para tareas asincronas. NO `threading` en Django.
- `pytest-django`. NO `unittest.TestCase`.

## Do's
- Service layer entre views y modelos.
- `QuerySet` methods para filtros reutilizables.
- `django-filters` para filtrado de APIs.
- `django-debug-toolbar` en desarrollo para cazar N+1.
- Signals solo para side-effects cross-app. No para logica de negocio.

## Don'ts
- NO N+1 queries (usa `assertNumQueries` en tests).
- NO logica de negocio en signals.
- NO `null=True, blank=True` sin justificacion.
- NO `GenericForeignKey` si puedes modelarlo con FK concreto.
- NO `runserver` en produccion.

## Recommended Commands
- `python manage.py makemigrations --check` — Verificar migraciones
- `pytest -n auto --cov --cov-fail-under=80` — Tests con coverage
- `python manage.py shell_plus --print-sql` — Debug queries
