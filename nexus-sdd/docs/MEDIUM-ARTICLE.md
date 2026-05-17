# La Fábrica de Software: Cómo Dejé de Corregir Código de IA y Empecé a Dirigirla

> *El "vibe coding" está generando deuda técnica más rápido de lo que los equipos pueden pagarla. Esta es la solución que construimos.*

---

## El momento "esto no puede seguir así"

Eran las 11pm. Llevaba 3 horas debuggeando un bug que Claude Code había introducido en el módulo de pagos. El agente, en su entusiasmo, había implementado Stripe Checkout sin idempotency keys. Dos webhooks simultáneos = doble cargo a clientes.

Lo peor: **este mismo bug ya lo habíamos resuelto hace 6 meses en otro proyecto.**

El desarrollador que lo arregló ya no estaba en la empresa. El conocimiento se fue con él. Y yo estaba ahí, a las 11pm, pagando el precio de no tener memoria institucional.

Esa noche decidí construir Nexus-SDD.

## El problema no es la IA. Es la falta de control.

Los agentes de IA son brillantes escribiendo código. Pero:

- **No planifican.** Escriben código sin spec. Sin plan. Sin diseño.
- **No recuerdan.** Cada sesión empieza desde cero. El conocimiento se evapora.
- **No se coordinan.** Un agente usa MongoDB, otro Postgres. Nadie detecta la contradicción.
- **No validan seguridad.** Hardcodean API keys, tokens, connection strings.

No es que la IA sea mala. Es que **le dimos un Ferrari sin volante.**

## Lo que construimos

Nexus-SDD es un framework open-source que transforma la IA de "chatbot que escribe código" a "fábrica de software industrial".

```bash
curl -fsSL https://raw.githubusercontent.com/JesusDevs/nexus-sdd/main/install.sh | bash
nexus-sdd init --suite fullstack
nexus-sdd spec "Login con OAuth2"
nexus-sdd status
```

**Un comando. Cero fricción. Todo listo.**

## El corazón: Spec-Driven Development real

Nexus-SDD obliga a la IA a seguir fases estrictas:

```
SPEC  → ¿Qué vamos a construir? (OpenSpec + escenarios BDD)
PLAN  → ¿Cómo lo vamos a construir? (con memoria de proyectos pasados)
CODE  → Implementación (un archivo → su test → siguiente)
TEST  → BDD + Unit + Integration (Ralph Loop: falla → corrige → pasa)
SEC   → Escaneo de secrets antes del commit (18 patrones)
```

**La IA no puede escribir una línea de código hasta que el spec esté aprobado.** Esto elimina el 80% del "vibe coding".

## El cerebro: memoria que sobrevive a sesiones

Integramos [Engram](https://github.com/Gentleman-Programming/engram) (3.4k estrellas, MIT) y construimos [Engram-Vec](https://github.com/JesusDevs/engram-vec), una extensión vectorial que entiende **significado, no palabras clave**.

```
FTS5 (búsqueda léxica):
  "login con Google" → encuentra documentos con "login" o "Google"
  ❌ NO encuentra "autenticación OAuth 2.0 con PKCE"

Vectorial (búsqueda semántica):
  "login con Google" → embedding de 1024 dimensiones
  ✅ ENCUENTRA "autenticación OAuth", "SSO provider", "social login"
```

Y la killer feature: **transferencia de conocimiento entre proyectos.**

```bash
engram-vec transfer "implementar pagos con Stripe" ecommerce-app
# → Encuentra en fintech-app (proyecto de hace 8 meses):
#   "Race condition en webhooks de Stripe. Fix: idempotency keys."
# → El desarrollador nuevo evita el bug que ya resolvimos.
```

Esto es memoria institucional real. El conocimiento ya no se va cuando el dev se va.

## El detector: 14 skills instaladas automáticamente

Nexus-SDD escanea tu proyecto y detecta el stack:

```bash
nexus-sdd init
# Detecta: Python + FastAPI + PostgreSQL + React + Next.js + Vitest
# Instala: fastapi + react + nextjs + bdd-behave + playwright
# Configura: perfiles de equipo + security middleware + Engram
```

O elegís una suite predefinida:

```bash
nexus-sdd init --suite mobile          # Android + Kotlin KMP + Flutter + SwiftUI
nexus-sdd init --suite "backend,ai"    # FastAPI + Django + LangGraph + AWS
nexus-sdd init --suite mobile --only kotlin  # Solo Android/Kotlin
```

14 skills. 12 suites. Desde React hasta LangGraph en AWS con Bedrock.

## El escudo: seguridad que bloquea antes del commit

El security middleware escanea cada archivo que la IA genera:

```
❌ sk-ant-api03-...hardcoded     → BLOQUEADO
❌ pk_live_...                    → BLOQUEADO
❌ mongodb://admin:pass@...       → BLOQUEADO
❌ -----BEGIN PRIVATE KEY-----    → BLOQUEADO
```

18 patrones de detección. Si encuentra algo, bloquea el commit, revierte el archivo, y genera una alerta en `.nexus/alerts/security_leak_blocked.md`.

## Los números: esto es lo que ahorra

| Escenario | Sin Nexus-SDD | Con Nexus-SDD | Ahorro |
|---|---|---|---|
| Feature simple | 3,000 tokens | 2,500 + 500 skills | Empate |
| Feature media (5 archivos) | 12,000 tokens | 5,000 + 800 skills | **-52%** |
| Feature compleja (múltiples HDUs) | 45,000 tokens | 18,000 + 2,000 skills | **-55%** |
| Bug recurrente (ya resuelto antes) | 8h debugging | 0h (lo encuentra Engram-Vec) | **-100%** |
| Onboarding dev nuevo | 2-3 semanas | 2-3 días | **-80%** |

**Cada token gastado en skills ahorra 5-20 tokens en correcciones.**

## ¿Por qué Open Source?

Porque el estándar debe ser de la comunidad. Las empresas Fortune 500 ya están armando sus "frameworks internos de IA". Pero lo hacen a puertas cerradas, en silos, sin compartir.

Nosotros creemos que:

- **Las skills deben ser abiertas.** Una skill de "scoring crediticio" escrita por un banco en Chile puede servirle a una fintech en México.
- **El protocolo debe ser abierto.** `.nexus/` como formato universal que cualquier IDE pueda leer.
- **La memoria debe ser transportable.** Si cambiás de Claude a Codex, tus skills y tu memoria te siguen.

## El roadmap

| Fase | Qué | Cuándo |
|---|---|---|
| **Ahora** | Nexus-SDD CLI + 14 skills + 12 suites + Security + Engram-Vec | ✅ Listo |
| **Q3 2026** | Dashboard web (token usage, HDU progress, alertas) | En desarrollo |
| **Q4 2026** | Catálogo comunitario de skills (PRs abiertos) | Planificado |
| **Q1 2027** | Nexus Cloud (SaaS para equipos) | Beta cerrada |

## El equipo detrás de esto

Somos desarrolladores que vivimos el problema en carne propia. No somos una startup respaldada por VC. Somos gente que programa con IA todos los días y dijo "esto no puede seguir así".

## Cómo empezar

```bash
# Un comando. Todo listo.
curl -fsSL https://raw.githubusercontent.com/JesusDevs/nexus-sdd/main/install.sh | bash

# O cloná y ejecutá
git clone https://github.com/JesusDevs/nexus-sdd.git
cd nexus-sdd && ./install.sh

# Creá tu primer feature con SDD real
nexus-sdd spec "Mi primer feature con IA disciplinada"
nexus-sdd status
```

## Lo que viene

Esto es el inicio. La visión es un ecosistema donde:

- Las skills son el "lenguaje común" entre humanos e IA
- La memoria institucional sobrevive a las personas
- La seguridad es automática, no opcional
- Cualquier agente (Claude, Codex, Cursor, OpenCode) entiende el estándar

**La fábrica de software ya no es un sueño. Es open source. Es hoy.**

---

*¿Querés contribuir? [github.com/JesusDevs/nexus-sdd](https://github.com/JesusDevs/nexus-sdd)*
*¿Querés la extensión vectorial? [github.com/JesusDevs/engram-vec](https://github.com/JesusDevs/engram-vec)*
