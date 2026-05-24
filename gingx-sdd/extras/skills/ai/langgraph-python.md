---
name: langgraph-python
description: LangGraph patterns for building stateful AI agents — Supervisor, tool nodes, checkpointer, streaming, interrupts
category: ai
stack: [langgraph, python, langchain, fastapi, langfuse]
triggers: [langgraph, state graph, supervisor, agent, tool node, checkpointer, subgraph, interrupt, streaming, langchain, langfuse]
---

# LangGraph Python Skill

## Agent Attitude
Eres un desarrollador de agentes IA con LangGraph. Piensas en grafos de estado, no en cadenas lineales.
El supervisor orquesta sub-agentes especializados. Cada nodo es una función pura.
Siempre hay un checkpointer para resiliencia. Siempre hay streaming para UX.

## Reglas

### Grafo
- `StateGraph(State)` con tipado fuerte. NUNCA `dict` genérico.
- `add_node("name", function)` — una responsabilidad por nodo.
- `add_conditional_edges("node", router, mapping)` — ruteo explícito.
- `add_edge("a", "b")` para flujo fijo.
- `set_entry_point("start")` y `set_finish_point("end")`.

### Estado
- `TypedDict` o `Pydantic BaseModel` para el estado. Campos con tipo.
- Estado inmutable. Cada nodo retorna un dict parcial con `{key: new_value}`.
- `operator.add` para listas acumulativas (mensajes, eventos).

### Supervisor Pattern
```python
# Router que decide qué agente ejecutar
def supervisor_router(state: AgentState) -> str:
    phase = state.get("phase", "spec")
    if phase == "spec": return "spec_agent"
    if phase == "code": return "code_agent"
    if phase == "done": return END
    return END

# Cada agente retorna al supervisor
workflow.add_conditional_edges("supervisor", supervisor_router, {...})
for agent in ["spec_agent", "code_agent"]:
    workflow.add_edge(agent, "supervisor")
```

### Checkpointer
- `MemorySaver` para desarrollo/testing.
- `SqliteSaver` para producción local.
- `PostgresSaver` o `DynamoDBSaver` para producción cloud.
- `graph.compile(checkpointer=checkpointer)` siempre.

### Streaming
```python
# Streaming por token
for event in graph.stream(initial_state, config={"callbacks": [handler]}):
    print(event)  # Cada nodo emite sus cambios

# Modos de streaming
graph.stream(state, stream_mode="values")   # Estado completo
graph.stream(state, stream_mode="updates")  # Solo cambios
graph.stream(state, stream_mode="debug")    # Debug detallado
```

### Interrupts (Human-in-the-Loop)
```python
# Pausar antes de acción crítica
workflow.add_node("approval", approval_node, interrupt_before=["approval"])

# Retomar con input humano
graph.invoke(Command(resume={"approved": True}), config)
```

### Subgraphs
- Un subgraph es un `StateGraph` compilado que se usa como nodo.
- Útil para agentes jerárquicos: supervisor → team_agent (subgraph con 3 agentes internos).

### Tools
```python
from langgraph.prebuilt import ToolNode

tools = [search_engram, write_file, run_tests]
tool_node = ToolNode(tools)
workflow.add_node("tools", tool_node)
workflow.add_conditional_edges("agent", should_use_tools, {"tools": "tools", END: END})
workflow.add_edge("tools", "agent")  # Loop: agente → tools → agente
```

## Do's

### Estructura de proyecto
```
agent/
├── graph.py           # StateGraph + compile
├── state.py           # TypedDict del estado
├── nodes/
│   ├── supervisor.py  # Nodo de ruteo
│   ├── spec.py        # Agente especializado
│   └── code.py        # Agente especializado
├── tools/
│   └── engram.py      # Herramientas MCP
└── tests/
    └── test_graph.py  # Test del grafo completo
```

### Testing
- `pytest` + `langgraph.testing` para simular invocaciones.
- Testear cada nodo en aislamiento.
- Testear el grafo completo con `graph.invoke(state)`.
- Mockear tools externas (Engram, APIs).

### LangFuse
```python
from langfuse.langchain import CallbackHandler

handler = CallbackHandler(
    session_id=hdu_id,
    user_id=developer_id,
    tags=["gingx-sdd", project_name]
)
graph.invoke(state, config={"callbacks": [handler]})
```

### FastAPI + LangGraph
```python
from fastapi import FastAPI
from fastapi.responses import StreamingResponse

app = FastAPI()

@app.post("/agent/stream")
async def stream_agent(request: AgentRequest):
    async def event_stream():
        for event in graph.stream(request.state):
            yield f"data: {json.dumps(event)}\n\n"
    
    return StreamingResponse(event_stream(), media_type="text/event-stream")
```

## Don'ts
- NO `StateGraph` sin tipo. Siempre `StateGraph(YourState)`.
- NO nodos de +100 líneas. Dividir en sub-nodos.
- NO `while True` ni loops infinitos. El grafo debe tener finitud garantizada.
- NO reglas de negocio en el router. El router solo decide destino.
- NO `print()` para debug. Usar `stream_mode="debug"` o LangFuse.
- NO `graph.invoke()` sin `config` si usás checkpointer o callbacks.
- NO herramientas que muten estado global. Cada tool recibe estado, retorna resultado.

## LangGraph Versions
- **LangGraph ≥ 0.2.0**: StateGraph API estable, Command, subgraphs.
- **LangGraph Cloud**: Deploy serverless con streaming + checkpointer gestionado.
- **LangGraph Studio**: IDE visual para debuggear grafos.

### GoalGraph (Autonomous Loop)
```python
# GoalGraph: plan-act-observe-reflect para trabajo autónomo
class GoalState(TypedDict):
    objective: str
    key_results: list[str]
    progress: dict[str, float]
    history: Annotated[list, operator.add]
    iteration: int
    status: str  # active | blocked | completed

workflow = StateGraph(GoalState)
workflow.add_node("planner", planner)
workflow.add_node("executor", executor)
workflow.add_node("observer", observer)
workflow.add_node("reflector", reflector)
workflow.add_conditional_edges("reflector", should_continue, {
    "continue": "planner", "completed": END, "blocked": END,
})
```
Ver `goal` skill para el patrón completo (plan-act-observe-reflect, ScheduleWakeup, checkpointing).

## Recommended Commands
- `pip install langgraph langchain langchain-openai langfuse` — Core
- `pip install langgraph-cli` — Deploy a LangGraph Cloud
- `pytest -n auto --cov=agent` — Tests
- `langgraph dev` — Modo desarrollo con Studio
- `langgraph up` — Deploy local con Docker
