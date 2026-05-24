---
name: goal
description: Autonomous goal-driven agent loops with LangGraph — GoalGraph, plan-act-observe-reflect cycle, checkpointed execution
category: ai
stack: [langgraph, python, goal]
triggers: [goal, objetivo, autonomo, loop, overnight, autonomous, self-paced, GoalGraph]
---

# Goal Skill — Autonomous Goal Execution

## Agent Attitude
Eres un agente autónomo que persigue objetivos (goals) sin supervisión humana. Trabajas en ciclos plan-act-observe-reflect.
Cada iteración avanza el goal, persiste el progreso, y decide si continuar o terminar. El checkpointer es tu memoria entre sesiones.
No preguntas — actúas, observas, y ajustas. Solo te detienes cuando el goal está completo o encuentras un bloqueo insalvable.

## Reglas

### GoalGraph Pattern
```python
from typing import TypedDict, Annotated
from langgraph.graph import StateGraph, END
from langgraph.checkpoint.memory import MemorySaver

class GoalState(TypedDict):
    objective: str
    key_results: list[str]
    progress: dict[str, float]        # KR -> 0.0..1.0
    current_step: str
    history: Annotated[list, operator.add]
    status: str                       # active | blocked | completed
    iteration: int

# Nodos del GoalGraph
def planner(state: GoalState) -> dict:
    """Evalúa el progreso y decide el siguiente paso concreto."""
    ...

def executor(state: GoalState) -> dict:
    """Ejecuta el paso planeado (escribir código, tests, docs)."""
    ...

def observer(state: GoalState) -> dict:
    """Observa los resultados: qué cambió, qué se completó."""
    ...

def reflector(state: GoalState) -> dict:
    """Reflexiona: ¿estamos más cerca del objetivo? ¿Qué ajustar?"""
    ...

# Armar el grafo
workflow = StateGraph(GoalState)
workflow.add_node("planner", planner)
workflow.add_node("executor", executor)
workflow.add_node("observer", observer)
workflow.add_node("reflector", reflector)

workflow.set_entry_point("planner")
workflow.add_edge("planner", "executor")
workflow.add_edge("executor", "observer")
workflow.add_edge("observer", "reflector")
workflow.add_conditional_edges("reflector", should_continue, {
    "continue": "planner",
    "completed": END,
    "blocked": END,
})

checkpointer = MemorySaver()  # o SqliteSaver para prod
graph = workflow.compile(checkpointer=checkpointer)
```

### Plan-Act-Observe-Reflect Cycle
- **Plan**: Revisar key results, progreso actual, historial. Elegir UN solo paso concreto.
- **Act**: Ejecutar el paso. Escribir código, tests, docs, o hacer research. Sin preguntar.
- **Observe**: Registrar qué se hizo, qué archivos cambiaron, qué tests pasan/fallan.
- **Reflect**: Comparar progreso contra key results. ¿Completado? → END. ¿Bloqueado? → flag + END. ¿Más trabajo? → continue.

### Checkpointing
- `MemorySaver` para desarrollo/testing (en memoria).
- `SqliteSaver` para persistencia local (`.gingx/goals/checkpoints.db`).
- `PostgresSaver` o `DynamoDBSaver` para producción.
- Cada iteración guarda checkpoint automáticamente via `graph.invoke(state, config)`.
- El config usa `thread_id` = goal_id para recuperar estado entre sesiones.

### Completion Criteria
- Todos los key results con progress ≥ 1.0 → goal completado.
- O el reflector detecta que no hay más pasos accionables → bloqueado.
- NUNCA loop infinito. Si iteration > max_iterations sin progreso → bloqueado.

### Autonomous Loop (ScheduleWakeup)
```python
# El agente se auto-programa
if should_continue(state) == "continue":
    # Programar siguiente wake-up en 60-300s (cache-friendly)
    schedule_wakeup(delay_seconds=180, reason="Goal iteration", prompt=goal_prompt)
else:
    # Goal completado o bloqueado — notificar y persistir
    save_goal_result(state)
```

## Do's

### Estructura de Goal
```yaml
# .gingx/goals/<goal-id>.yaml
goal_id: "autonomous-code-review"
objective: "Revisar todo el código del proyecto en busca de bugs y mejoras"
key_results:
  - "Todos los archivos .py revisados"           # KR1
  - "Bugs encontrados documentados en mnemo"      # KR2
  - "Mejoras implementadas con tests"             # KR3
status: active
max_iterations: 50
iteration: 12
progress:
  kr1: 0.65
  kr2: 0.40
  kr3: 0.15
history:
  - "Iter 1: Revisado src/models.py, encontré 2 bugs"
  - "Iter 2: Fix bug #1 en models.py, tests pasan"
```

### OKR Format
- **Objective**: Cualitativo, aspiracional. ¿Qué queremos lograr?
- **Key Results**: Cuantitativos, medibles. 2-5 KRs por goal.
- Cada KR tiene un `progress` de 0.0 a 1.0.
- El goal está completo cuando TODOS los KRs tienen progress ≥ 1.0.

### Progress Tracking
- Cada iteración actualiza al menos un KR.
- El history acumula entradas inmutables (Annotated[list, operator.add]).
- Mnemo guarda snapshot del goal state al final de cada iteración.

### Integration with Mnemo
```bash
mnemo save "Goal: <objective>" \
  "Iteration <N>: <summary>. KR progress: <details>. Next: <step>." \
  --type progress --outcome in_progress --project $(basename $(pwd)) --tags goal,autonomous,<goal-id>
```

## Don'ts
- NO preguntar al usuario durante la ejecución autónoma.
- NO loops infinitos sin checkpoint. Cada iteración persiste.
- NO modificar el objective durante la ejecución (los KRs son fijos).
- NO ejecutar más de una acción por iteración (un paso = un cambio).
- NO ignorar errores. Si un paso falla 3 veces, marcar como bloqueado.
- NO trabajar sin un goal_id configurado en `.gingx/goals/`.
- NO exceder `max_iterations` sin marcar el goal como blocked.

## Goal Loop Lifecycle
```
[CREATED] → planner → executor → observer → reflector
                ↑                              ↓
                └──── continue ────────┘       ↓
                                          [COMPLETED]
                                          [BLOCKED]
```

## Recommended Commands
- `gingx-sdd goal create "Title" --objective "Obj" --key-results "KR1,KR2"` — Crear goal
- `gingx-sdd goal list` — Listar goals
- `gingx-sdd goal status <id>` — Ver progreso
- `gingx-sdd goal loop <id>` — Iniciar loop autónomo
- `gingx-sdd goal complete <id>` — Marcar completado
- `mnemo search "Goal: <id>" --project <name>` — Buscar snapshots en mnemo
