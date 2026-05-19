# Proposal: GraphRAG Hybrid Layer for nexus-mnemo

## Why

Pure vector search has a fundamental blind spot: it retrieves **documents**, not **relationships**. When an architect asks "If I change the Tax-Rate field, which 50 programs break?", cosine similarity cannot answer this -- it finds semantically similar memories about "tax rate" but has no concept of dependency chains, impact propagation, or causal lineage. This is the **multi-hop reasoning gap**.

The research saved to mnemo (Decision: GraphRAG hybrid approach, Vineet Chachondia, Dec 2025) identifies three superpowers that knowledge graphs unlock:

1. **Multi-hop reasoning** -- traverse chains of relationships instead of chaining independent queries
2. **Built-in explainability** -- every answer comes with an auditable path of nodes and edges showing exactly why it was reached
3. **Contextual density** -- filter at the relationship level, not just content similarity, producing higher-signal results

Meanwhile, HDU-06 defines `swarm/dag.go` for Component Dependency DAG resolution. That DAG is currently an in-memory transient structure. By persisting it as a graph layer inside nexus-mnemo, we get three wins at once: the DAG becomes durable and queryable, cross-session impact analysis becomes possible, and the graph layer serves both the swarm orchestrator and end-user queries.

The hybrid pattern is: **vectors for fuzzy entry points** (natural language to entities), **graph traversal for relationship exploration** (entity to dependencies). This is not a replacement for vector search -- it is a complementary layer that answers the questions vector search cannot.

## What Changes

### 1. Graph layer in nexus-mnemo (`graph/` package)
New Go package adding SQLite-backed adjacency list on top of the existing vec/store.go. Two new tables (`graph_nodes`, `graph_edges`) coexist with `vec_memories` in the same SQLite database. No external graph database -- consistent with the project's zero-dependency, local-first architecture.

### 2. Node and edge model
- **Nodes** are entities linked to vec_memories. A memory can be a node (e.g., "Tax-Rate field definition") or nodes can be standalone (e.g., "PaymentService component").
- **Edges** carry typed relationships: `DEPENDS_ON`, `CALLS`, `READS_FROM`, `WRITES_TO`, `DEFINES_DATA`, `RELATES_TO`, `CONTRADICTS`, `SUPERSEDES`, `DERIVES_FROM`.
- Both nodes and edges support JSON properties for flexible metadata.

### 3. Hybrid query engine
The core innovation: `graph_hybrid_search(query, traversal_depth, edge_types)`:
1. Embed the query via Ollama (bge-m3, 1024-dim)
2. Vector search on node labels/properties to find entry nodes
3. From entry nodes, BFS traversal following specified edge types up to `traversal_depth`
4. Return the subgraph: entry nodes + reached nodes + traversed edges, ranked by combined vector similarity and graph distance

### 4. Six new MCP tools
| Tool | Purpose |
|------|---------|
| `graph_add_node` | Create a node linked to a memory or standalone |
| `graph_add_edge` | Create a typed relationship between two nodes |
| `graph_traverse` | Multi-hop BFS/DFS traversal from a node |
| `graph_hybrid_search` | Vector search entry point + graph traversal in one call |
| `graph_list_neighbors` | Immediate neighbors of a node (1-hop) |
| `graph_find_path` | Shortest path between two nodes (Dijkstra/BFS) |

Total MCP tools: 14 (HDU-06) to 20.

### 5. Automatic DAG persistence
`swarm/dag.go` (from HDU-06) will be instrumented to auto-persist its component dependency graph as graph nodes and edges. When the DAG resolver builds a dependency tree, `DEPENDS_ON` edges are written to the graph layer. This means:
- The swarm's dependency knowledge survives crashes and restarts
- Impact analysis queries can be answered instantly without rebuilding the DAG
- Cross-HDU dependency patterns become queryable

### 6. New CLI commands
- `mnemo graph add-node <label> [--memory-id <id>] [--project <name>] [--properties <json>]`
- `mnemo graph add-edge <source-id> <target-id> <type> [--properties <json>]`
- `mnemo graph traverse <node-id> [--depth <n>] [--edge-types <types>] [--direction out|in|both]`
- `mnemo graph hybrid-search <query> [--depth <n>] [--project <name>]`

### 7. Integration with existing features
- **Release system (HDU-07)**: Graph snapshots included in release exports
- **Conflict detection**: Auto-create `CONTRADICTS` edges when `mem_detect_conflicts_semantic` finds contradictions
- **Knowledge transfer**: `DERIVES_FROM` edges when knowledge transfers across projects
- **Swarm orchestration**: `DEPENDS_ON` edges auto-populated from task dependencies

## What Does NOT Change
- Mnemo DB schema -- new tables are additive, existing `vec_memories` unchanged
- Embedding pipeline -- still Ollama + bge-m3 + cosine similarity
- MCP protocol -- JSON-RPC 2.0, same server, same wire format
- Zero external APIs -- graph is pure SQLite adjacency, no Neo4j, no cloud graphs
- OpenSpec structure -- standard HDU format with proposal/design/specs/tasks
- Existing 14 MCP tools (8 current + 6 from HDU-06) -- graph tools are additive

## Impact
- HDU: HDU-graphrag-hybrid
- Complexity: medium (one new Go package, ~6 files, no external deps)
- New Go package: `graph/` within nexus-mnemo
- New database tables: `graph_nodes`, `graph_edges`
- New MCP tools: 6 (bringing total to 14 + 6 = 20)
- New CLI commands: 4 (`mnemo graph add-node`, `add-edge`, `traverse`, `hybrid-search`)
- Dependencies: HDU-06 (swarm/dag.go for auto-persistence), HDU-07 (release snapshots)
- Prerequisite: None -- can be built independently, integration points with HDU-06 are additive hooks

## Acceptance Criteria
1. `mnemo graph add-node "TaxRateField"` creates a node persisted in SQLite
2. `mnemo graph add-edge node-A node-B DEPENDS_ON` creates a relationship
3. `mnemo graph traverse node-A --depth 2` returns the 2-hop subgraph
4. `mnemo graph hybrid-search "tax rate impact" --depth 3` returns: vector-matched entry nodes + multi-hop dependency subgraph
5. Graph operations exposed as MCP tools and callable from Claude Code
6. DAG persistence hook in swarm/dag.go writes DEPENDS_ON edges automatically
7. Release export includes graph state alongside mempack
8. All operations are local-only (SQLite, no network calls for graph operations)
