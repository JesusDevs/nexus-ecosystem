package codegraph

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"strings"
)

// Store manages code intelligence data in SQLite.
// Lives alongside vec_memories in the same mnemo.db.
type Store struct {
	db *sql.DB
}

// NewStore opens a codegraph store using an existing *sql.DB connection.
func NewStore(db *sql.DB) (*Store, error) {
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("codegraph migrate: %w", err)
	}
	return s, nil
}

func (s *Store) migrate() error {
	stmts := []string{
		// Tables
		`CREATE TABLE IF NOT EXISTS cg_nodes (
			id TEXT PRIMARY KEY, kind TEXT NOT NULL, name TEXT NOT NULL,
			qualified_name TEXT NOT NULL, file_path TEXT NOT NULL,
			language TEXT NOT NULL, start_line INTEGER NOT NULL,
			end_line INTEGER NOT NULL, signature TEXT,
			is_exported INTEGER DEFAULT 0, parent_id TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS cg_edges (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			source TEXT NOT NULL, target TEXT NOT NULL, kind TEXT NOT NULL,
			line INTEGER,
			FOREIGN KEY (source) REFERENCES cg_nodes(id) ON DELETE CASCADE,
			FOREIGN KEY (target) REFERENCES cg_nodes(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS cg_files (
			path TEXT PRIMARY KEY, content_hash TEXT NOT NULL,
			language TEXT NOT NULL, size INTEGER NOT NULL,
			indexed_at TEXT NOT NULL DEFAULT (datetime('now')),
			node_count INTEGER DEFAULT 0
		)`,
		// Indexes
		`CREATE INDEX IF NOT EXISTS idx_cg_nodes_kind ON cg_nodes(kind)`,
		`CREATE INDEX IF NOT EXISTS idx_cg_nodes_name ON cg_nodes(name)`,
		`CREATE INDEX IF NOT EXISTS idx_cg_nodes_file ON cg_nodes(file_path)`,
		`CREATE INDEX IF NOT EXISTS idx_cg_nodes_lang ON cg_nodes(language)`,
		`CREATE INDEX IF NOT EXISTS idx_cg_nodes_file_line ON cg_nodes(file_path, start_line)`,
		`CREATE INDEX IF NOT EXISTS idx_cg_edges_source ON cg_edges(source, kind)`,
		`CREATE INDEX IF NOT EXISTS idx_cg_edges_target ON cg_edges(target, kind)`,
		`CREATE INDEX IF NOT EXISTS idx_cg_files_lang ON cg_files(language)`,
		// FTS5 — optional, skipped if not compiled in
		`CREATE VIRTUAL TABLE IF NOT EXISTS cg_nodes_fts USING fts5(
			id, name, qualified_name, signature,
			content='cg_nodes', content_rowid='rowid'
		)`,
		// Triggers (single statements — no internal splitting)
		`CREATE TRIGGER IF NOT EXISTS cg_nodes_ai AFTER INSERT ON cg_nodes BEGIN
			INSERT INTO cg_nodes_fts(rowid, id, name, qualified_name, signature)
			VALUES (NEW.rowid, NEW.id, NEW.name, NEW.qualified_name, NEW.signature);
		END`,
		`CREATE TRIGGER IF NOT EXISTS cg_nodes_ad AFTER DELETE ON cg_nodes BEGIN
			INSERT INTO cg_nodes_fts(cg_nodes_fts, rowid, id, name, qualified_name, signature)
			VALUES ('delete', OLD.rowid, OLD.id, OLD.name, OLD.qualified_name, OLD.signature);
		END`,
		`CREATE TRIGGER IF NOT EXISTS cg_nodes_au AFTER UPDATE ON cg_nodes BEGIN
			INSERT INTO cg_nodes_fts(cg_nodes_fts, rowid, id, name, qualified_name, signature)
			VALUES ('delete', OLD.rowid, OLD.id, OLD.name, OLD.qualified_name, OLD.signature);
			INSERT INTO cg_nodes_fts(rowid, id, name, qualified_name, signature)
			VALUES (NEW.rowid, NEW.id, NEW.name, NEW.qualified_name, NEW.signature);
		END`,
	}

	hasFTS5 := true
	for _, q := range stmts {
		q = strings.TrimSpace(q)
		// Skip FTS5-dependent objects when FTS5 module is missing
		if !hasFTS5 && strings.Contains(q, "cg_nodes_fts") {
			continue
		}
		if _, err := s.db.Exec(q); err != nil {
			errMsg := err.Error()
			if strings.Contains(errMsg, "fts5") || strings.Contains(errMsg, "no such module") {
				hasFTS5 = false
				continue
			}
			return fmt.Errorf("schema exec [%.60s]: %w", q, err)
		}
	}
	return nil
}

// ── Nodes ──────────────────────────────────────────────────────

func (s *Store) InsertNode(n *Node) error {
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO cg_nodes
			(id, kind, name, qualified_name, file_path, language,
			 start_line, end_line, signature, is_exported, parent_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, n.ID, n.Kind, n.Name, n.QualifiedName, n.FilePath, n.Language,
		n.StartLine, n.EndLine, n.Signature, boolToInt(n.IsExported), nullStr(n.ParentID))
	return err
}

func (s *Store) GetNode(id string) (*Node, error) {
	row := s.db.QueryRow(`
		SELECT id, kind, name, qualified_name, file_path, language,
		       start_line, end_line, signature, is_exported, parent_id
		FROM cg_nodes WHERE id = ?
	`, id)
	n := &Node{}
	var isExp int
	var sig, parent sql.NullString
	err := row.Scan(&n.ID, &n.Kind, &n.Name, &n.QualifiedName, &n.FilePath,
		&n.Language, &n.StartLine, &n.EndLine, &sig, &isExp, &parent)
	if err != nil {
		return nil, err
	}
	n.Signature = sig.String
	n.IsExported = isExp == 1
	n.ParentID = parent.String
	return n, nil
}

// ── Edges ──────────────────────────────────────────────────────

func (s *Store) InsertEdge(e *Edge) error {
	_, err := s.db.Exec(`
		INSERT INTO cg_edges (source, target, kind, line)
		VALUES (?, ?, ?, ?)
	`, e.Source, e.Target, e.Kind, e.Line)
	return err
}

// ── Files ──────────────────────────────────────────────────────

func (s *Store) UpsertFile(f *FileInfo) error {
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO cg_files (path, content_hash, language, size, indexed_at, node_count)
		VALUES (?, ?, ?, ?, datetime('now'), ?)
	`, f.Path, f.ContentHash, f.Language, f.Size, f.NodeCount)
	return err
}

// ── Search (FTS5) ─────────────────────────────────────────────

func (s *Store) SearchSymbols(query string, language string, limit int) ([]SearchResult, error) {
	ftsQuery := strings.Join(strings.Fields(query), " AND ")
	rows, err := s.db.Query(`
		SELECT cg_nodes.id, cg_nodes.kind, cg_nodes.name, cg_nodes.qualified_name,
		       cg_nodes.file_path, cg_nodes.language, cg_nodes.start_line, cg_nodes.end_line,
		       cg_nodes.signature, cg_nodes.is_exported, cg_nodes.parent_id,
		       rank
		FROM cg_nodes_fts
		JOIN cg_nodes ON cg_nodes.rowid = cg_nodes_fts.rowid
		WHERE cg_nodes_fts MATCH ?
		ORDER BY rank
		LIMIT ?
	`, ftsQuery, limit)
	if err != nil {
		// FTS5 may error on malformed queries — fall back to LIKE
		return s.searchLike(query, language, limit)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var r SearchResult
		var sig, parent sql.NullString
		var isExp int
		if err := rows.Scan(&r.Node.ID, &r.Node.Kind, &r.Node.Name, &r.Node.QualifiedName,
			&r.Node.FilePath, &r.Node.Language, &r.Node.StartLine, &r.Node.EndLine,
			&sig, &isExp, &parent, &r.Rank); err != nil {
			continue
		}
		r.Node.Signature = sig.String
		r.Node.IsExported = isExp == 1
		r.Node.ParentID = parent.String
		results = append(results, r)
	}
	return results, nil
}

func (s *Store) searchLike(query string, language string, limit int) ([]SearchResult, error) {
	likeQ := "%" + query + "%"
	rows, err := s.db.Query(`
		SELECT id, kind, name, qualified_name, file_path, language,
		       start_line, end_line, signature, is_exported, parent_id
		FROM cg_nodes
		WHERE name LIKE ? OR qualified_name LIKE ?
		ORDER BY name
		LIMIT ?
	`, likeQ, likeQ, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var r SearchResult
		var sig, parent sql.NullString
		var isExp int
		if err := rows.Scan(&r.Node.ID, &r.Node.Kind, &r.Node.Name, &r.Node.QualifiedName,
			&r.Node.FilePath, &r.Node.Language, &r.Node.StartLine, &r.Node.EndLine,
			&sig, &isExp, &parent); err != nil {
			continue
		}
		r.Node.Signature = sig.String
		r.Node.IsExported = isExp == 1
		r.Node.ParentID = parent.String
		r.Rank = 0
		results = append(results, r)
	}
	return results, nil
}

// ── Callers (who calls me?) ────────────────────────────────────

func (s *Store) GetCallers(nodeID string) ([]CallerResult, error) {
	rows, err := s.db.Query(`
		SELECT cg_nodes.id, cg_nodes.kind, cg_nodes.name, cg_nodes.qualified_name,
		       cg_nodes.file_path, cg_nodes.language, cg_nodes.start_line, cg_nodes.end_line,
		       cg_nodes.signature, cg_nodes.is_exported, cg_nodes.parent_id,
		       cg_edges.id, cg_edges.kind, cg_edges.line
		FROM cg_edges
		JOIN cg_nodes ON cg_nodes.id = cg_edges.source
		WHERE cg_edges.target = ? AND cg_edges.kind = 'calls'
		ORDER BY cg_nodes.name
	`, nodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []CallerResult
	for rows.Next() {
		var cr CallerResult
		var sig, parent sql.NullString
		var isExp int
		if err := rows.Scan(&cr.Caller.ID, &cr.Caller.Kind, &cr.Caller.Name, &cr.Caller.QualifiedName,
			&cr.Caller.FilePath, &cr.Caller.Language, &cr.Caller.StartLine, &cr.Caller.EndLine,
			&sig, &isExp, &parent, &cr.Edge.ID, &cr.Edge.Kind, &cr.Edge.Line); err != nil {
			continue
		}
		cr.Caller.Signature = sig.String
		cr.Caller.IsExported = isExp == 1
		cr.Caller.ParentID = parent.String
		cr.Edge.Source = cr.Caller.ID
		cr.Edge.Target = nodeID
		results = append(results, cr)
	}
	return results, nil
}

// ── Callees (who do I call?) ───────────────────────────────────

func (s *Store) GetCallees(nodeID string) ([]CalleeResult, error) {
	rows, err := s.db.Query(`
		SELECT cg_nodes.id, cg_nodes.kind, cg_nodes.name, cg_nodes.qualified_name,
		       cg_nodes.file_path, cg_nodes.language, cg_nodes.start_line, cg_nodes.end_line,
		       cg_nodes.signature, cg_nodes.is_exported, cg_nodes.parent_id,
		       cg_edges.id, cg_edges.kind, cg_edges.line
		FROM cg_edges
		JOIN cg_nodes ON cg_nodes.id = cg_edges.target
		WHERE cg_edges.source = ? AND cg_edges.kind = 'calls'
		ORDER BY cg_nodes.name
	`, nodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []CalleeResult
	for rows.Next() {
		var cr CalleeResult
		var sig, parent sql.NullString
		var isExp int
		if err := rows.Scan(&cr.Callee.ID, &cr.Callee.Kind, &cr.Callee.Name, &cr.Callee.QualifiedName,
			&cr.Callee.FilePath, &cr.Callee.Language, &cr.Callee.StartLine, &cr.Callee.EndLine,
			&sig, &isExp, &parent, &cr.Edge.ID, &cr.Edge.Kind, &cr.Edge.Line); err != nil {
			continue
		}
		cr.Callee.Signature = sig.String
		cr.Callee.IsExported = isExp == 1
		cr.Callee.ParentID = parent.String
		cr.Edge.Source = nodeID
		cr.Edge.Target = cr.Callee.ID
		results = append(results, cr)
	}
	return results, nil
}

// ── Stats ──────────────────────────────────────────────────────

func (s *Store) Stats() (map[string]int, error) {
	stats := make(map[string]int)
	var nodes, edges, files int
	s.db.QueryRow("SELECT COUNT(*) FROM cg_nodes").Scan(&nodes)
	s.db.QueryRow("SELECT COUNT(*) FROM cg_edges").Scan(&edges)
	s.db.QueryRow("SELECT COUNT(*) FROM cg_files").Scan(&files)
	stats["nodes"] = nodes
	stats["edges"] = edges
	stats["files"] = files

	rows, err := s.db.Query("SELECT language, COUNT(*) FROM cg_nodes GROUP BY language")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var lang string
			var count int
			rows.Scan(&lang, &count)
			stats["nodes_"+lang] = count
		}
	}
	return stats, nil
}

// ── Clear ──────────────────────────────────────────────────────

func (s *Store) ClearAll() error {
	for _, table := range []string{"cg_edges", "cg_nodes", "cg_files"} {
		if _, err := s.db.Exec("DELETE FROM " + table); err != nil {
			return err
		}
	}
	// FTS5 content rebuild
	s.db.Exec("INSERT INTO cg_nodes_fts(cg_nodes_fts) VALUES ('rebuild')")
	return nil
}

// ── Helpers ────────────────────────────────────────────────────

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func nullStr(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// HashFile returns a sha256 hash of the content.
func HashFile(content []byte) string {
	h := sha256.Sum256(content)
	return fmt.Sprintf("%x", h)
}

// DetectLanguage returns the language for a file extension.
func DetectLanguage(path string) string {
	switch {
	case strings.HasSuffix(path, ".go"):
		return "go"
	case strings.HasSuffix(path, ".py"):
		return "python"
	case strings.HasSuffix(path, ".ts"):
		return "typescript"
	case strings.HasSuffix(path, ".tsx"):
		return "typescript"
	case strings.HasSuffix(path, ".js"):
		return "javascript"
	case strings.HasSuffix(path, ".jsx"):
		return "javascript"
	default:
		return "unknown"
	}
}
