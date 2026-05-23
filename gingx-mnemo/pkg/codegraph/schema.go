package codegraph

// SQLite schema for code intelligence — nodes, edges, files, FTS5.
// Lives in the same mnemo.db alongside vec_memories.
// Inspired by gingx-codegraph but native Go, no npm.

const SchemaSQL = `
-- CodeGraph: code intelligence tables
-- Nodes: code symbols (functions, classes, variables, etc.)
CREATE TABLE IF NOT EXISTS cg_nodes (
    id TEXT PRIMARY KEY,
    kind TEXT NOT NULL,
    name TEXT NOT NULL,
    qualified_name TEXT NOT NULL,
    file_path TEXT NOT NULL,
    language TEXT NOT NULL,
    start_line INTEGER NOT NULL,
    end_line INTEGER NOT NULL,
    signature TEXT,
    is_exported INTEGER DEFAULT 0,
    parent_id TEXT
);

-- Edges: relationships between nodes
CREATE TABLE IF NOT EXISTS cg_edges (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    source TEXT NOT NULL,
    target TEXT NOT NULL,
    kind TEXT NOT NULL,
    line INTEGER,
    FOREIGN KEY (source) REFERENCES cg_nodes(id) ON DELETE CASCADE,
    FOREIGN KEY (target) REFERENCES cg_nodes(id) ON DELETE CASCADE
);

-- Files: tracked source files
CREATE TABLE IF NOT EXISTS cg_files (
    path TEXT PRIMARY KEY,
    content_hash TEXT NOT NULL,
    language TEXT NOT NULL,
    size INTEGER NOT NULL,
    indexed_at TEXT NOT NULL DEFAULT (datetime('now')),
    node_count INTEGER DEFAULT 0
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_cg_nodes_kind ON cg_nodes(kind);
CREATE INDEX IF NOT EXISTS idx_cg_nodes_name ON cg_nodes(name);
CREATE INDEX IF NOT EXISTS idx_cg_nodes_file ON cg_nodes(file_path);
CREATE INDEX IF NOT EXISTS idx_cg_nodes_lang ON cg_nodes(language);
CREATE INDEX IF NOT EXISTS idx_cg_nodes_file_line ON cg_nodes(file_path, start_line);
CREATE INDEX IF NOT EXISTS idx_cg_edges_source ON cg_edges(source, kind);
CREATE INDEX IF NOT EXISTS idx_cg_edges_target ON cg_edges(target, kind);
CREATE INDEX IF NOT EXISTS idx_cg_files_lang ON cg_files(language);

-- Full-text search on node names and signatures
CREATE VIRTUAL TABLE IF NOT EXISTS cg_nodes_fts USING fts5(
    id,
    name,
    qualified_name,
    signature,
    content='cg_nodes',
    content_rowid='rowid'
);

-- FTS sync triggers
CREATE TRIGGER IF NOT EXISTS cg_nodes_ai AFTER INSERT ON cg_nodes BEGIN
    INSERT INTO cg_nodes_fts(rowid, id, name, qualified_name, signature)
    VALUES (NEW.rowid, NEW.id, NEW.name, NEW.qualified_name, NEW.signature);
END;

CREATE TRIGGER IF NOT EXISTS cg_nodes_ad AFTER DELETE ON cg_nodes BEGIN
    INSERT INTO cg_nodes_fts(cg_nodes_fts, rowid, id, name, qualified_name, signature)
    VALUES ('delete', OLD.rowid, OLD.id, OLD.name, OLD.qualified_name, OLD.signature);
END;

CREATE TRIGGER IF NOT EXISTS cg_nodes_au AFTER UPDATE ON cg_nodes BEGIN
    INSERT INTO cg_nodes_fts(cg_nodes_fts, rowid, id, name, qualified_name, signature)
    VALUES ('delete', OLD.rowid, OLD.id, OLD.name, OLD.qualified_name, OLD.signature);
    INSERT INTO cg_nodes_fts(rowid, id, name, qualified_name, signature)
    VALUES (NEW.rowid, NEW.id, NEW.name, NEW.qualified_name, NEW.signature);
END;
`

// NodeKind constants — subset of the full gingx-codegraph taxonomy.
const (
	KindFunction  = "function"
	KindMethod    = "method"
	KindClass     = "class"
	KindInterface = "interface"
	KindVariable  = "variable"
	KindImport    = "import"
	KindModule    = "module"
	KindStruct    = "struct"
)

// EdgeKind constants.
const (
	EdgeContains    = "contains"
	EdgeCalls       = "calls"
	EdgeImports     = "imports"
	EdgeReferences  = "references"
	EdgeExtends     = "extends"
	EdgeImplements  = "implements"
)
