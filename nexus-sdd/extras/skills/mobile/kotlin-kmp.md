---
name: kotlin-kmp
description: Kotlin Multiplatform (KMP) con Compose Multiplatform para iOS y Android
category: mobile
stack: [kotlin, kmp, compose-multiplatform, android, ios, gradle]
triggers: [kotlin, kmp, compose, multiplatform, android, ios, gradle, kts]
---

# Kotlin Multiplatform Skill

## Agent Attitude
Eres un desarrollador KMP que comparte logica entre plataformas.
`commonMain` es tu hogar. Solo bajas a `androidMain`/`iosMain` cuando es inevitable.
Compose Multiplatform para UI compartida.

## Rules
- Logica de negocio en `commonMain`. Siempre.
- `expect`/`actual` para platform-specific. Documenta por que.
- Ktor para networking. OkHttp en Android, Darwin en iOS.
- SQLDelight para base de datos local.
- Koin para dependency injection (KMP-native, sin reflection).
- Coroutines + Flow para async. NUNCA callbacks.

## Do's
- `sealed class` para estados de UI (Loading, Success, Error).
- ViewModel compartido en `commonMain` con `kotlinx-coroutines`.
- Compose Multiplatform para UI (Material 3).
- Tests en `commonTest` con `kotlin.test`.

## Don'ts
- NO logica de negocio en `androidMain` o `iosMain`.
- NO reflection (no funciona en KMP).
- NO Java-specific APIs en `commonMain`.
- NO `synchronized` (usa `Mutex` de coroutines).
- NO `System.currentTimeMillis()` (usa `kotlinx-datetime`).

## Recommended Commands
- `./gradlew :shared:allTests` — Tests multiplataforma
- `./gradlew :shared:detekt` — Static analysis
- `./gradlew :composeApp:assembleDebug` — Android debug
