# Design: GraphRAG Hybrid Layer for nexus-mnemo

## Architecture Overview

```
nexus-mnemo/
├── vec/store.go          # Vector DB (existing, unchanged)
├── vec/embed.go          # Embeddings (existing, unchanged)
├── mcp/server.go         # MCP server (existing, extended with 6 graph tools)
├── main.go               # CLI (existing, extended with 4 graph commands)
│
├── graph/                # NEW: GraphRAG Hybrid Layer
│   ├── store.go          # GraphStore: SQLite adjacency list, CRUD for nodes/edges
│   ├── traverse.go       # BFS/DFS traversal engine with depth/type filtering
│   ├── hybrid.go         # Hybrid query: vector search entry → graph traversal
│   ├── path.go           # Shortest path between nodes (BFS, unweighted)
│   └── types.go          # Node, Edge, TraversalResult, HybridSearchResult types
```

## 1. Data Model

### Rationale: SQLite Adjacency List over Neo4j/Property Graph

| Factor | SQLite Adjacency | Neo4j / dedicated graph DB |
|--------|-----------------|---------------------------|
| Deployment | Zero additional services | Requires Neo4j process, Java runtime |
| Consistency with architecture | Same SQLite DB as vec_memories | Separate storage, separate backups |
| Query power | BFS/DFS via Go code, not Cypher | Full Cypher, but we don't need it |
| Multi-hop performance | Acceptable: ~100K edges in memory per traversal | Optimized for millions of edges |
| Operational complexity | None (existing DB handle) | New process, monitoring, auth |

Decision: SQLite adjacency list. The graph layer is for knowledge graphs (hundreds to low thousands of nodes per project), not social graphs (millions). BFS in application code over an adjacency list loaded from SQLite is fast enough and avoids introducing a new infrastructure dependency. This is the same reasoning that led to SQLite for vectors -- local-first, single-binary, no services.

### Table Schema

```sql
CREATE TABLE IF NOT EXISTS graph_nodes (
    id TEXT PRIMARY KEY,                          -- e.g. "node-nexus-ecosystem-paymentservice"
    project TEXT NOT NULL DEFAULT '',              -- project namespace
    memory_id TEXT,                               -- FK to vec_memories.id (nullable: standalone nodes)
    label TEXT NOT NULL,                          -- human-readable name, e.g. "PaymentService"
    node_type TEXT NOT NULL DEFAULT 'entity',     -- entity|component|field|function|decision|milestone
    properties TEXT NOT NULL DEFAULT '{}',        -- JSON: arbitrary key-value metadata
    embedding BLOB,                               -- optional: node-level embedding for semantic entry
    embedding_model TEXT NOT NULL DEFAULT '',
    embedding_dim INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (memory_id) REFERENCES vec_memories(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_graph_nodes_project ON graph_nodes(project);
CREATE INDEX IF NOT EXISTS idx_graph_nodes_memory_id ON graph_nodes(memory_id);
CREATE INDEX IF NOT EXISTS idx_graph_nodes_type ON graph_nodes(node_type);

CREATE TABLE IF NOT EXISTS graph_edges (
    id TEXT PRIMARY KEY,                          -- e.g. "edge-nexus-ecosystem-paymentservice-taxrate"
    project TEXT NOT NULL DEFAULT '',
    source_node_id TEXT NOT NULL,                 -- FK to graph_nodes.id
    target_node_id TEXT NOT NULL,                 -- FK to graph_nodes.id
    relationship_type TEXT NOT NULL,              -- DEPENDS_ON|CALLS|READS_FROM|WRITES_TO|etc.
    properties TEXT NOT NULL DEFAULT '{}',        -- JSON: edge metadata (e.g. {"frequency": "daily"})
    weight REAL NOT NULL DEFAULT 1.0,             -- edge weight for traversal priority (higher = stronger)
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (source_node_id) REFERENCES graph_nodes(id) ON DELETE CASCADE,
    FOREIGN KEY (target_node_id) REFERENCES graph_nodes(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_graph_edges_source ON graph_edges(source_node_id);
CREATE INDEX IF NOT EXISTS idx_graph_edges_target ON graph_edges(target_node_id);
CREATE INDEX IF NOT EXISTS idx_graph_edges_project ON graph_edges(project);
CREATE INDEX IF NOT EXISTS idx_graph_edges_type ON graph_edges(relationship_type);
```

### Relationship Types (Standardized Vocabulary)

| Type | Semantics | Example |
|------|-----------|---------|
| `DEPENDS_ON` | A requires B to function | "PaymentService DEPENDS_ON AuthService" |
| `CALLS` | A invokes B | "Checkout CALLS PaymentService.processPayment" |
| `READS_FROM` | A reads data from B | "ReportGenerator READS_FROM TransactionsDB" |
| `WRITES_TO` | A writes data to B | "PaymentService WRITES_TO LedgerDB" |
| `DEFINES_DATA` | A defines the schema/structure of B | "TaxRateField DEFINES_DATA tax_percentage(DB column)" |
| `IMPLEMENTS` | A implements the specification in B | "GoHandler IMPLEMENTS REST-spec-payment" |
| `RELATES_TO` | General semantic relationship | "AuthService RELATES_TO SecurityPolicy" |
| `CONTRADICTS` | A contradicts B | "Decision-JWT-15min CONTRADICTS Decision-JWT-24h" |
| `SUPERSEDES` | A replaces B | "v2.0-auth SUPERSEDES v1.0-auth" |
| `DERIVES_FROM` | A was derived from B (knowledge transfer) | "nexus-sdd-auth DERIVES_FROM banking-app-auth" |
| `CONTAINS` | A is a parent/container of B | "PaymentModule CONTAINS PaymentService" |
| `PRECEDES` | A must happen before B (temporal/ordering) | "Phase-Spec PRECEDES Phase-Design" |

This vocabulary is intentionally constrained -- free-form strings lead to inconsistent querying. Edge types map directly to the problem domain identified in the GraphRAG research (CALLS, READS_FROM, DEFINES_DATA for impact analysis) and in HDU-06 (DEPENDS_ON for task DAG).

## 2. Go API Design

### `graph/types.go`

```go
package graph

type Node struct {
    ID             string            `json:"id"`
    Project        string            `json:"project"`
    MemoryID       string            `json:"memory_id,omitempty"`
    Label          string            `json:"label"`
    NodeType       string            `json:"node_type"`
    Properties     map[string]any    `json:"properties"`
    Embedding      []float32         `json:"embedding,omitempty"`
    EmbeddingModel string            `json:"embedding_model,omitempty"`
    EmbeddingDim   int               `json:"embedding_dim,omitempty"`
    CreatedAt      string            `json:"created_at"`
}

type Edge struct {
    ID               string         `json:"id"`
    Project          string         `json:"project"`
    SourceNodeID     string         `json:"source_node_id"`
    TargetNodeID     string         `json:"target_node_id"`
    RelationshipType string         `json:"relationship_type"`
    Properties       map[string]any `json:"properties"`
    Weight           float64        `json:"weight"`
    CreatedAt        string         `json:"created_at"`
}

type TraversalResult struct {
    StartNode Node             `json:"start_node"`
    Paths     []TraversalPath  `json:"paths"`
    Nodes     map[string]Node  `json:"nodes"`   // deduplicated set of reached nodes
    Edges     []Edge           `json:"edges"`   // edges traversed
}

type TraversalPath struct {
    Nodes []string `json:"nodes"` // ordered list of node IDs
    Edges []string `json:"edges"` // ordered list of edge IDs
    Depth int      `json:"depth"`
}

type HybridSearchResult struct {
    Query           string          `json:"query"`
    EntryNodes      []EntryNode     `json:"entry_nodes"`      // vector search results → graph entry points
    Traversal       TraversalResult `json:"traversal"`         // graph traversal from entry points
    Explanation     string          `json:"explanation"`       // human-readable impact chain
}

type EntryNode struct {
    Node       Node    `json:"node"`
    Similarity float32 `json:"similarity"` // cosine similarity from vector search
}
```

### `graph/store.go` -- GraphStore

```go
package graph

import "database/sql"

type GraphStore struct {
    db      *sql.DB         // shared handle from vec.Store
    project string
}

func NewGraphStore(db *sql.DB, project string) *GraphStore {
    g := &GraphStore{db: db, project: project}
    g.migrate()
    return g
}

func (g *GraphStore) migrate() error {
    // CREATE TABLE IF NOT EXISTS graph_nodes ...
    // CREATE TABLE IF NOT EXISTS graph_edges ...
    // Called from vec.Store.migrate() so tables exist alongside vec_memories
}

// AddNode creates a node. If memoryID is set, links to an existing memory.
// If properties contain text to embed, generates embedding via the shared embedder.
func (g *GraphStore) AddNode(node *Node) error { ... }

// AddEdge creates a relationship between two existing nodes.
// Rejects if source or target don't exist.
func (g *GraphStore) AddEdge(edge *Edge) error { ... }

// GetNode retrieves a single node by ID.
func (g *GraphStore) GetNode(id string) (*Node, error) { ... }

// GetNeighbors returns all edges and neighbor nodes for a given node.
func (g *GraphStore) GetNeighbors(nodeID string, direction string, edgeTypes []string) ([]Edge, []Node, error) { ... }

// DeleteNode removes a node and all its edges (CASCADE).
func (g *GraphStore) DeleteNode(id string) error { ... }

// DeleteEdge removes an edge.
func (g *GraphStore) DeleteEdge(id string) error { ... }

// ListNodes returns nodes filtered by project, type, label pattern.
func (g *GraphStore) ListNodes(nodeType string, limit int) ([]Node, error) { ... }

// ExportGraph returns all nodes and edges for a project (used by pack export).
func (g *GraphStore) ExportGraph(project string) ([]Node, []Edge, error) { ... }
```

### `graph/traverse.go` -- BFS/DFS Traversal

```go
// Traverse performs multi-hop BFS from a start node.
// Parameters:
//   - maxDepth: maximum hops (0 = unlimited, be careful)
//   - direction: "out" (follow outgoing), "in" (follow incoming), "both"
//   - edgeTypes: filter by relationship types, empty = all types
//
// Implementation: BFS with a visited set and a queue. At each level:
//   1. Dequeue node
//   2. Query edges matching direction + type filters
//   3. For each neighbor not in visited: enqueue, record path
//   4. Continue until queue empty or maxDepth reached
func (g *GraphStore) Traverse(startNodeID string, maxDepth int, direction string, edgeTypes []string) (*TraversalResult, error) { ... }
```

BFS is chosen over DFS because:
- Impact analysis is inherently breadth-first (what does X directly affect?)
- BFS guarantees shortest paths in unweighted graphs
- DFS can produce misleadingly deep chains

### `graph/hybrid.go` -- Hybrid Query (the killer feature)

```go
// HybridSearch implements the core GraphRAG pattern:
//   1. Vector search finds entry nodes matching the query semantically
//   2. Graph traversal follows relationships from entry nodes
//   3. Results merged with explainability path
//
// Algorithm:
//   Step 1: Embed query via embedder.Embed(query)
//   Step 2: For each node with an embedding in graph_nodes, compute cosine similarity
//           (or search node labels via vec_memories if node has memory_id)
//   Step 3: Select top-k entry nodes above minSimilarity
//   Step 4: For each entry node, run Traverse(entry, maxDepth, "both", edgeTypes)
//   Step 5: Merge subgraphs: union of all reached nodes and edges, deduplicated
//   Step 6: Rank by combined score = alpha * vector_similarity + (1-alpha) * (1 / graph_distance)
//   Step 7: Generate human-readable explanation chain
func (g *GraphStore) HybridSearch(
    queryEmbedding []float32,
    project string,
    topK int,
    minSimilarity float32,
    maxDepth int,
    edgeTypes []string,
) (*HybridSearchResult, error) { ... }
```

The entry point embedding strategy has two paths:
1. **Node with memory_id**: delegate to the existing vector search on vec_memories, then map memory → node
2. **Node with its own embedding**: compute cosine similarity directly against the node's embedding BLOB

Option 2 is preferred for graph-native nodes (components, fields) because it avoids an unnecessary join. But option 1 is supported for nodes that directly mirror memories.

### `graph/path.go` -- Shortest Path

```go
// FindPath finds the shortest path between two nodes using BFS (unweighted edges).
// If weighted=true, uses Dijkstra with edge weights.
// Returns the ordered list of node IDs and edge IDs forming the path.
// Returns nil if no path exists (disconnected components).
func (g *GraphStore) FindPath(fromNodeID, toNodeID string, maxDepth int, weighted bool) (*TraversalPath, error) { ... }
```

BFS is the default because most edges have weight=1.0. Dijkstra is available for weighted traversal when edge weights encode connection strength.

## 3. MCP Tool Definitions

The six new tools extend the existing 14-tool MCP server (8 current + 6 from HDU-06):

```go
// mcp/server.go — add to tools slice

{
    Name:        "graph_add_node",
    Description: "Create a node in the knowledge graph. Nodes represent entities, components, fields, functions, or decisions. Optionally link to an existing vector memory.",
    InputSchema: JSONSchema{
        Type: "object",
        Properties: map[string]PropDef{
            "label":      {Type: "string", Description: "Human-readable name (e.g. 'PaymentService')"},
            "node_type":  {Type: "string", Description: "entity|component|field|function|decision|milestone"},
            "memory_id":  {Type: "string", Description: "Optional: link to a vec_memories entry"},
            "project":    {Type: "string", Description: "Project namespace (default: current)"},
            "properties": {Type: "string", Description: "JSON string of key-value metadata"},
        },
        Required: []string{"label", "node_type"},
    },
},
{
    Name:        "graph_add_edge",
    Description: "Create a typed relationship between two nodes. Types: DEPENDS_ON, CALLS, READS_FROM, WRITES_TO, DEFINES_DATA, IMPLEMENTS, RELATES_TO, CONTRADICTS, SUPERSEDES, DERIVES_FROM, CONTAINS, PRECEDES.",
    InputSchema: JSONSchema{
        Type: "object",
        Properties: map[string]PropDef{
            "source_node_id": {Type: "string", Description: "Source node ID"},
            "target_node_id": {Type: "string", Description: "Target node ID"},
            "type":           {Type: "string", Description: "Relationship type from the standard vocabulary"},
            "properties":     {Type: "string", Description: "JSON string of edge metadata"},
            "weight":         {Type: "number", Description: "Edge weight (default 1.0, higher = stronger connection)"},
        },
        Required: []string{"source_node_id", "target_node_id", "type"},
    },
},
{
    Name:        "graph_traverse",
    Description: "Multi-hop graph traversal from a starting node. Follows relationships up to a maximum depth. Use for impact analysis: 'show me everything that depends on this component.'",
    InputSchema: JSONSchema{
        Type: "object",
        Properties: map[string]PropDef{
            "node_id":    {Type: "string", Description: "Starting node ID"},
            "max_depth":  {Type: "integer", Description: "Maximum traversal depth (default: 3)"},
            "direction":  {Type: "string", Description: "out|in|both (default: both)"},
            "edge_types": {Type: "string", Description: "Comma-separated relationship types to follow (default: all)"},
        },
        Required: []string{"node_id"},
    },
},
{
    Name:        "graph_hybrid_search",
    Description: "The core GraphRAG operation: (1) vector search to find semantically relevant entry nodes, (2) graph traversal from those nodes to find related entities. Returns the combined subgraph with an explanation chain. Use this for queries like 'which components break if I change X?'",
    InputSchema: JSONSchema{
        Type: "object",
        Properties: map[string]PropDef{
            "query":          {Type: "string", Description: "Natural language query (e.g. 'impact of changing Tax-Rate field')"},
            "project":        {Type: "string", Description: "Project to search within"},
            "top_k":          {Type: "integer", Description: "Number of vector-matched entry nodes (default: 3)"},
            "min_similarity": {Type: "number", Description: "Minimum cosine similarity for entry (default: 0.5)"},
            "max_depth":      {Type: "integer", Description: "Graph traversal depth from entry nodes (default: 3)"},
            "edge_types":     {Type: "string", Description: "Comma-separated types to traverse (default: all)"},
        },
        Required: []string{"query"},
    },
},
{
    Name:        "graph_list_neighbors",
    Description: "List immediate neighbors (1-hop) of a graph node. Shows all incoming and outgoing relationships.",
    InputSchema: JSONSchema{
        Type: "object",
        Properties: map[string]PropDef{
            "node_id":    {Type: "string", Description: "Node ID"},
            "direction":  {Type: "string", Description: "out|in|both (default: both)"},
            "edge_types": {Type: "string", Description: "Filter by relationship types (comma-separated)"},
        },
        Required: []string{"node_id"},
    },
},
{
    Name:        "graph_find_path",
    Description: "Find the shortest path between two nodes in the knowledge graph. Returns the chain of nodes and relationships connecting them.",
    InputSchema: JSONSchema{
        Type: "object",
        Properties: map[string]PropDef{
            "from_node_id": {Type: "string", Description: "Source node ID"},
            "to_node_id":   {Type: "string", Description: "Target node ID"},
            "max_depth":    {Type: "integer", Description: "Maximum path length to search (default: 6)"},
        },
        Required: []string{"from_node_id", "to_node_id"},
    },
},
```

## 4. Hybrid Query Algorithm in Detail

This is the intellectual core of the HDU. The algorithm answers: "Given a natural language query about impact/dependencies, return the relevant subgraph."

```
Input: query="Which services break if I change the TaxRate field?",
       project="nexus-ecosystem",
       topK=3, minSimilarity=0.5, maxDepth=3,
       edgeTypes=["DEPENDS_ON", "CALLS", "READS_FROM", "WRITES_TO", "DEFINES_DATA"]

Phase 1: Vector Entry Point Discovery
  1. embedQuery = ollama.Embed(query)  → []float32{1024 dims}
  2. candidateNodes = graphStore.ListNodes(project)
  3. For each node with embedding:
       sim = cosineSimilarity(embedQuery, node.Embedding)
       if sim >= minSimilarity: candidates.append(node, sim)
  4. Sort candidates by similarity descending, take topK
  5. entryNodes = [
       {node: TaxRateField, similarity: 0.89},
       {node: PaymentCalculator, similarity: 0.72},
       {node: TaxConfigModule, similarity: 0.65}
     ]

Phase 2: Multi-Source Graph Traversal
  6. For each entryNode in entryNodes:
       subgraph = Traverse(entryNode.Node.ID, maxDepth=3, direction="both", edgeTypes)
       // BFS from entry node, following only the specified edge types
       allNodes.union(subgraph.Nodes)
       allEdges.union(subgraph.Edges)

Phase 3: Scoring and Ranking
  7. For each node in allNodes:
       // Combined score: α * vector_sim(entry) + (1-α) * (1 / min_distance_to_any_entry)
       // α = 0.6 (biased toward semantic relevance)
       score = 0.6 * entrySimilarity + 0.4 * (1.0 / float64(graphDistance))
  8. Sort nodes by score descending
  9. Sort edges by minimum distance from entry nodes

Phase 4: Explanation Generation
  10. explanation = buildExplanationChain(entryNodes, allNodes, allEdges)
      // "TaxRateField (semantic match 89%) is defined by PaymentCalculator.
      //  PaymentCalculator DEPENDS_ON TaxConfigModule.
      //  TaxConfigModule CALLS 12 downstream services:
      //  - InvoiceService (READS_FROM tax_rate column)
      //  - ReportGenerator (READS_FROM tax_category)
      //  - ..."

Output: HybridSearchResult{
    EntryNodes: [TaxRateField, PaymentCalculator, TaxConfigModule],
    Traversal: {nodes: 15, edges: 18, paths: [3 paths of depth 1-3]},
    Explanation: "Changing TaxRateField impacts 15 components across 3 dependency chains..."
}
```

### Performance Characteristics
- Phase 1 (vector search on nodes): O(N * D) where N = nodes with embeddings, D = 1024. For 1000 nodes: ~1M float ops, < 5ms.
- Phase 2 (BFS traversal): O(V + E) per entry node. For 1000 nodes and 5000 edges: < 10ms.
- Phase 3 (scoring): O(V log V) for sorting. Negligible for < 10K nodes.
- Total: < 50ms for typical knowledge graph (1000 nodes, 5000 edges).

## 5. Integration with HDU-06 Swarm DAG

### Auto-Persistence Hook

In `swarm/dag.go`, after building the dependency DAG in memory:

```go
// swarm/dag.go — added to DAG resolution function
func (dag *DAG) Resolve() ([]string, error) {
    // ... existing resolution logic ...
    ordered := dag.topologicalSort()

    // [NEW] Persist DAG to graph layer
    if dag.graphStore != nil {
        for _, task := range dag.Tasks {
            // Ensure node exists for this task
            nodeID := "node-" + dag.Project + "-" + task.ID
            dag.graphStore.AddNode(&graph.Node{
                ID:       nodeID,
                Project:  dag.Project,
                Label:    task.Phase + ":" + task.HDUID,
                NodeType: "milestone",
                Properties: map[string]any{
                    "hdu_id": task.HDUID,
                    "phase":  task.Phase,
                    "status": task.Status,
                },
            })

            // Create edges for dependencies
            for _, depID := range task.Dependencies {
                depNodeID := "node-" + dag.Project + "-" + depID
                dag.graphStore.AddEdge(&graph.Edge{
                    SourceNodeID:     nodeID,
                    TargetNodeID:     depNodeID,
                    RelationshipType: "DEPENDS_ON",
                    Properties: map[string]any{
                        "source": "swarm-dag-auto",
                    },
                })
            }
        }
    }

    return ordered, nil
}
```

This hook is **optional** -- if the graph store is nil, DAG resolution works identically to the HDU-06 spec. The dependency is one-directional: graph layer does not depend on swarm; swarm optionally writes to graph.

### Impact Analysis Query

With the DAG persisted, the swarm orchestrator can answer questions like:
```
> mnemo graph hybrid-search "what depends on task-5" --project nexus-ecosystem
→ Returns all tasks that have task-5 in their dependency chain
→ Shows impact depth (1-hop = direct dependency, 2-hop = transitive)
→ Explains: "Delaying task-5 blocks 3 direct dependents and 7 transitive dependents"
```

## 6. Integration with Vector Search

The graph layer does not replace vector search -- it complements it. The existing 8 search/transfer/similar tools remain unchanged. The graph tools add a new dimension of query capability:

| Query type | Tool | Pattern |
|-----------|------|---------|
| "What have we learned about auth?" | `mem_search_semantic` | Vector only |
| "Is this similar to past bugs?" | `mem_similar` | Vector only |
| "What lessons from other projects?" | `mem_transfer` | Vector only |
| "What breaks if I change X?" | `graph_hybrid_search` | Vector → Graph |
| "Show me the dependency chain" | `graph_traverse` | Graph only |
| "How are A and B connected?" | `graph_find_path` | Graph only |
| "What relates to this decision?" | `graph_list_neighbors` | Graph only |

## 7. Release and Export Integration (HDU-07)

When `mnemo pack export` is called, graph state is included:

```go
// main.go — runPack() extended
func runPack() {
    // ... existing export logic ...
    entries, _ := store.ExportPack(project, version)

    // [NEW] Include graph state
    graphStore := graph.NewGraphStore(store.DB(), project)
    nodes, edges, _ := graphStore.ExportGraph(project)

    pack := map[string]interface{}{
        "pack_version": "1.1",         // bumped from 1.0 for graph support
        "project":      project,
        "version":      version,
        "entries":      entries,
        "graph": map[string]interface{}{  // NEW
            "nodes": nodes,
            "edges": edges,
        },
    }
    // ...
}
```

## 8. Auto-Edge Creation from Conflict Detection

When `mem_detect_conflicts_semantic` finds a contradiction, the graph layer auto-creates a `CONTRADICTS` edge:

```go
// In the conflict detection handler:
if conflictFound {
    graphStore.AddEdge(&graph.Edge{
        SourceNodeID:     ensureNodeForMemory(conflict.MemoryA),
        TargetNodeID:     ensureNodeForMemory(conflict.MemoryB),
        RelationshipType: "CONTRADICTS",
        Properties: map[string]any{
            "similarity":  conflict.Similarity,
            "source":      "auto-conflict-detection",
        },
    })
}
```

## 9. Migration and Backward Compatibility

The graph tables are additive. Existing mnemo.db files work without modification. The migration runs alongside the existing vec_memories migration:

```go
// vec/store.go — migrate() extended
func (s *Store) migrate() error {
    // ... existing vec_memories, vec_projects, etc. tables ...
    
    // NEW: GraphRAG tables
    graphQueries := []string{
        `CREATE TABLE IF NOT EXISTS graph_nodes (...)`,
        `CREATE TABLE IF NOT EXISTS graph_edges (...)`,
        // ... indexes ...
    }
    for _, q := range graphQueries {
        s.db.Exec(q)
    }

    return nil
}
```

The `vec.Store` struct gains an optional `*graph.GraphStore` field. If nil (no graph operations requested), all graph code paths are no-ops. This keeps the vector-only path fast and the binary size impact minimal.

## 10. Alternatives Considered

### A. Separate graph database (Neo4j, Dgraph, ArangoDB)
**Rejected.** Introduces infrastructure complexity (separate process, Java runtime, authentication, backups), violates the local-first single-binary design principle, and adds network latency. A knowledge graph of 1000-10000 nodes does not need a dedicated graph DB.

### B. In-memory graph only (no persistence)
**Rejected.** The primary value of persisting the DAG from HDU-06 is durability across crashes and sessions. In-memory loses this. SQLite gives us persistence for free.

### C. Graph represented as vec_memories with special 'edge' type
**Rejected.** Pollutes the vector memory space with structural data. Edges don't need embeddings. Keeping them separate from semantic memories makes queries cleaner and avoids false-positive vector matches on structural data.

### D. Property graph in SQLite (nodes store edges inline)
**Rejected.** Adjacency list (separate edges table) is simpler, scales better for multi-hop traversal (indexed JOINs), and supports edge metadata naturally. Inline edge storage makes traversal queries complex.

## 11. Decision Log

| Decision | Rationale | Date |
|----------|-----------|------|
| SQLite adjacency list over Neo4j | Consistency with existing architecture, zero operational overhead | 2026-05-18 |
| BFS over DFS for traversal | Impact analysis is breadth-first; BFS gives shortest paths | 2026-05-18 |
| Hybrid scoring: α=0.6 vector, 0.4 graph distance | Semantic relevance is slightly more important than graph proximity | 2026-05-18 |
| Standardized edge types over free-form strings | Consistent querying, prevents taxonomy drift | 2026-05-18 |
| GraphStore as optional field in vec.Store | Keeps vector-only path fast, enables incremental adoption | 2026-05-18 |
| Graph data included in mempack 1.1 | Enables portable knowledge graphs, cross-tool sharing | 2026-05-18 |
