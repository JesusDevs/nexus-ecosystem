# SPEC GATE — SDD Enforcement

"No code without an approved spec."
Este documento es el checkpoint obligatorio antes de escribir una sola linea de codigo.

## Gate 1: Problem Definition
Responder SI o NO. Si alguna es NO, no se avanza.

- [ ] Hay un problema de usuario claramente definido? (no "quiero usar X tecnologia")
- [ ] Se identifico QUIEN tiene el problema? (rol, persona, stakeholder)
- [ ] Se definio como se MIDE el exito? (metricas, comportamiento esperado)
- [ ] Se verifico que no existe ya una solucion en el codebase? (check mnemo + git log)
- [ ] Se busco en mnemo decisiones previas relacionadas? (`mnemo search`)

## Gate 2: Context Gathering
La IA debe hacer estas preguntas antes de proponer solucion:

```
1. "Que clases/componentes existen ya que se relacionan con esto?"
   → Buscar en mnemo + explorar el codebase

2. "Que patrones de diseño se usaron en features similares?"
   → mnemo search "pattern <context>"

3. "Que trade-offs aplican para este caso especifico?"
   → Cargar memoria de decisiones previas similares

4. "Que dependencias (internas/externas) se ven afectadas?"
   → Revisar DAG de componentes (HDU-06)

5. "Cual es el alcance minimo viable (MVP) vs el alcance completo?"
   → Separar MUST de SHOULD/COULD
```

## Gate 3: Spec Artifacts
Antes de implementar, deben existir estos archivos:

```
openspec/changes/<HDU-ID>/
├── proposal.md     # Que, por que, para quien, metricas de exito
├── specs/          # Especificaciones detalladas
│   └── <feature>.md   # Con escenarios BDD (Gherkin)
├── design.md       # Enfoque tecnico, trade-offs, alternativas consideradas
└── tasks.md        # Tareas discretas, cada una con criterio de aceptacion
```

## Gate 4: AI Smart Questions
Para features, la IA DEBE preguntar con contexto antes de implementar:

| Contexto | Pregunta |
|----------|----------|
| Nueva feature | "Que patrones similares existen en el codebase? Que aprendimos de HDU anteriores?" |
| Cambio de UI | "Que componentes existentes se reusan? Que patron de diseño visual seguimos?" |
| Cambio de API/DB | "Que otros servicios dependen de esto? Cual es el plan de migracion?" |
| Bug fix | "Cual fue la decision original? Por que se hizo asi? Hay tests que cubran esto?" |
| Refactor | "Que HDU introdujo este codigo? Que trade-offs se documentaron?" |

## Gate 5: Delivery Mode
Determinar estrategia de entrega:

- [ ] **Single PR**: scope chico, ≤400 lineas, ≤3 areas
- [ ] **Stacked PRs**: feature grande, dividir en PRs encadenados
- [ ] **Feature track**: branch vacia como target, PRs incrementales
- [ ] **Ask on risk**: avanzar, frenar si se detecta riesgo

## Enforcement
Este gate se aplica via supervisor en la fase PROPOSE. Si los gates 1-3 no pasan:
- No se crean tasks
- No se escribe codigo
- Se documenta en mnemo el bloqueo y el motivo

```bash
mnemo save "SPEC GATE: <HDU-ID> blocked" \
  "Gate check failed. Missing: <list>. Action: <what's needed to unblock>." \
  --type decision --outcome blocked --project $(basename $(pwd)) --tags spec-gate,<hdu-id>
```
