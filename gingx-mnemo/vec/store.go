package vec

import (
	"bytes"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Store manages the vector tables inside the Mnemo DB.
type Store struct {
	db  *sql.DB
	dir string
}

// NewStore opens (or creates) the Mnemo database and adds the vec_memories table.
func NewStore(dbDir string) (*Store, error) {
	if dbDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("cannot get home dir: %w", err)
		}
		dbDir = filepath.Join(home, ".mnemo")
	}

	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("cannot create %s: %w", dbDir, err)
	}

	dbPath := filepath.Join(dbDir, "mnemo.db")
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("cannot open DB: %w", err)
	}

	db.Exec("PRAGMA journal_mode=WAL")
	db.Exec("PRAGMA synchronous=NORMAL")
	db.Exec("PRAGMA busy_timeout=5000")

	s := &Store{db: db, dir: dbDir}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	return s, nil
}

func (s *Store) migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS vec_memories (
			id TEXT PRIMARY KEY,
			project TEXT NOT NULL DEFAULT '',
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			type TEXT NOT NULL DEFAULT 'note',
			embedding BLOB,
			embedding_model TEXT NOT NULL DEFAULT '',
			embedding_dim INTEGER NOT NULL DEFAULT 0,
			tags TEXT NOT NULL DEFAULT '[]',
			outcome TEXT NOT NULL DEFAULT '',
			media_type TEXT NOT NULL DEFAULT 'text',
			media_file BLOB,
			version TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL DEFAULT (datetime('now')),
			updated_at TEXT NOT NULL DEFAULT (datetime('now'))
		)`,

		`CREATE INDEX IF NOT EXISTS idx_vec_memories_project ON vec_memories(project)`,
		`CREATE INDEX IF NOT EXISTS idx_vec_memories_type ON vec_memories(type)`,
		`CREATE INDEX IF NOT EXISTS idx_vec_memories_version ON vec_memories(version)`,

		`CREATE TABLE IF NOT EXISTS vec_projects (
			name TEXT PRIMARY KEY,
			description TEXT NOT NULL DEFAULT '',
			embedding_model TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL DEFAULT (datetime('now'))
		)`,

		`CREATE TABLE IF NOT EXISTS vec_releases (
			id TEXT PRIMARY KEY,
			project TEXT NOT NULL,
			version TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			memory_count INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL DEFAULT (datetime('now'))
		)`,
		`CREATE INDEX IF NOT EXISTS idx_vec_releases_project ON vec_releases(project)`,
		`CREATE INDEX IF NOT EXISTS idx_vec_releases_version ON vec_releases(project, version)`,

		`CREATE TABLE IF NOT EXISTS vec_config (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL DEFAULT ''
		)`,
	}

	for _, q := range queries {
		if _, err := s.db.Exec(q); err != nil {
			return fmt.Errorf("query failed [%s]: %w", q[:60], err)
		}
	}

	// Ensure defaults exist
	s.initDefaults()

	return nil
}

func (s *Store) initDefaults() {
	defaults := map[string]string{
		"embed.model":    "bge-m3",
		"embed.dims":     "1024",
		"ollama.host":    "http://localhost:11434",
		"mnemo.version":  "0.2.0",
		"swarm.mode":     "hybrid",
	}
	for k, v := range defaults {
		var exists string
		s.db.QueryRow("SELECT value FROM vec_config WHERE key = ?", k).Scan(&exists)
		if exists == "" {
			s.db.Exec("INSERT OR IGNORE INTO vec_config (key, value) VALUES (?, ?)", k, v)
		}
	}
}

// GetConfig retrieves a configuration value.
func (s *Store) GetConfig(key string) string {
	var value string
	s.db.QueryRow("SELECT value FROM vec_config WHERE key = ?", key).Scan(&value)
	return value
}

// SetConfig sets a configuration value.
func (s *Store) SetConfig(key, value string) error {
	_, err := s.db.Exec("INSERT OR REPLACE INTO vec_config (key, value) VALUES (?, ?)", key, value)
	return err
}

// AllConfig returns all configuration as a map.
func (s *Store) AllConfig() map[string]string {
	rows, err := s.db.Query("SELECT key, value FROM vec_config ORDER BY key")
	if err != nil {
		return nil
	}
	defer rows.Close()
	cfg := make(map[string]string)
	for rows.Next() {
		var k, v string
		rows.Scan(&k, &v)
		cfg[k] = v
	}
	return cfg
}

// Save stores a vector memory.
func (s *Store) Save(mem *VectorMemory) error {
	tagsJSON := "["
	for i, t := range mem.Tags {
		if i > 0 {
			tagsJSON += ","
		}
		tagsJSON += `"` + t + `"`
	}
	tagsJSON += "]"

	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO vec_memories
			(id, project, title, content, type,
			 embedding, embedding_model, embedding_dim,
			 tags, outcome, media_type, media_file, version, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'))
	`,
		mem.ID,
		mem.Project,
		mem.Title,
		mem.Content,
		mem.Type,
		floatsToBytes(mem.Embedding),
		mem.EmbeddingModel,
		mem.EmbeddingDim,
		tagsJSON,
		mem.Outcome,
		mem.MediaType,
		mem.MediaFile,
		mem.Version,
	)
	return err
}

// Release creates a snapshot of all current memories for a project.
func (s *Store) Release(project, version, description string) (*ReleaseSnapshot, error) {
	var count int
	s.db.QueryRow("SELECT COUNT(*) FROM vec_memories WHERE project = ?", project).Scan(&count)

	_, err := s.db.Exec(`
		UPDATE vec_memories SET version = ?, updated_at = datetime('now')
		WHERE project = ? AND version = ''
	`, version, project)
	if err != nil {
		return nil, fmt.Errorf("error tagging memories: %w", err)
	}

	id := fmt.Sprintf("rel-%s-%s-%d", project, version, time.Now().Unix())
	_, err = s.db.Exec(`
		INSERT INTO vec_releases (id, project, version, description, memory_count)
		VALUES (?, ?, ?, ?, ?)
	`, id, project, version, description, count)
	if err != nil {
		return nil, fmt.Errorf("error saving release: %w", err)
	}

	return &ReleaseSnapshot{
		ID:          id,
		Project:     project,
		Version:     version,
		Description: description,
		MemoryCount: count,
		CreatedAt:   time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

// Diff compares memories between two versions of a project.
func (s *Store) Diff(project, v1, v2 string) (*VersionDiff, error) {
	rows1, err := s.db.Query("SELECT id, title, type, outcome FROM vec_memories WHERE project = ? AND version = ?", project, v1)
	if err != nil {
		return nil, err
	}
	defer rows1.Close()

	v1Memories := make(map[string]MemorySummary)
	for rows1.Next() {
		var m MemorySummary
		rows1.Scan(&m.ID, &m.Title, &m.Type, &m.Outcome)
		v1Memories[m.ID] = m
	}

	rows2, err := s.db.Query("SELECT id, title, type, outcome FROM vec_memories WHERE project = ? AND version = ?", project, v2)
	if err != nil {
		return nil, err
	}
	defer rows2.Close()

	v2Memories := make(map[string]MemorySummary)
	for rows2.Next() {
		var m MemorySummary
		rows2.Scan(&m.ID, &m.Title, &m.Type, &m.Outcome)
		v2Memories[m.ID] = m
	}

	diff := &VersionDiff{FromVersion: v1, ToVersion: v2}

	for id, m := range v2Memories {
		if _, exists := v1Memories[id]; !exists {
			diff.Added = append(diff.Added, m)
		}
	}
	for id, m := range v1Memories {
		if m2, exists := v2Memories[id]; exists {
			if m.Title != m2.Title || m.Outcome != m2.Outcome {
				diff.Updated = append(diff.Updated, DiffPair{Old: m, New: m2})
			}
		} else {
			diff.Removed = append(diff.Removed, m)
		}
	}

	return diff, nil
}

// ListReleases lists all releases for a project.
func (s *Store) ListReleases(project string) ([]ReleaseSnapshot, error) {
	rows, err := s.db.Query(`
		SELECT id, project, version, description, memory_count, created_at
		FROM vec_releases WHERE project = ? ORDER BY created_at DESC
	`, project)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var releases []ReleaseSnapshot
	for rows.Next() {
		var r ReleaseSnapshot
		rows.Scan(&r.ID, &r.Project, &r.Version, &r.Description, &r.MemoryCount, &r.CreatedAt)
		releases = append(releases, r)
	}
	return releases, nil
}

// SearchSemantic searches memories by cosine similarity.
func (s *Store) SearchSemantic(queryEmbedding []float32, project string, limit int, minSimilarity float32) ([]*SearchResult, error) {
	var rows *sql.Rows
	var err error

	if project != "" {
		rows, err = s.db.Query(`
			SELECT id, project, title, content, type, embedding, outcome, created_at, version
			FROM vec_memories
			WHERE project = ? AND embedding IS NOT NULL
			ORDER BY created_at DESC
			LIMIT 1000
		`, project)
	} else {
		rows, err = s.db.Query(`
			SELECT id, project, title, content, type, embedding, outcome, created_at, version
			FROM vec_memories
			WHERE embedding IS NOT NULL
			ORDER BY created_at DESC
			LIMIT 5000
		`)
	}
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var results []*SearchResult
	for rows.Next() {
		var (
			id, proj, title, content, memType, outcome, createdAt string
			embeddingBlob                                           []byte
			version                                                 string
		)
		if err := rows.Scan(&id, &proj, &title, &content, &memType, &embeddingBlob, &outcome, &createdAt, &version); err != nil {
			continue
		}

		emb := bytesToFloats(embeddingBlob)
		if len(emb) == 0 || len(emb) != len(queryEmbedding) {
			continue
		}

		similarity := cosineSimilarity(queryEmbedding, emb)
		if similarity >= minSimilarity {
			results = append(results, &SearchResult{
				ID:         id,
				Project:    proj,
				Title:      title,
				Content:    content,
				Type:       memType,
				Outcome:    outcome,
				Similarity: similarity,
				CreatedAt:  createdAt,
				Version:    version,
			})
		}
	}

	sortResults(results)
	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// Similar finds memories similar to an existing one.
func (s *Store) Similar(memoryID string, crossProject bool, limit int, minSimilarity float32) ([]*SearchResult, error) {
	var embeddingBlob []byte
	err := s.db.QueryRow(`
		SELECT embedding FROM vec_memories WHERE id = ? AND embedding IS NOT NULL
	`, memoryID).Scan(&embeddingBlob)
	if err != nil {
		return nil, fmt.Errorf("memory not found or no embedding: %w", err)
	}

	queryEmbedding := bytesToFloats(embeddingBlob)
	if len(queryEmbedding) == 0 {
		return nil, fmt.Errorf("empty embedding for memory %s", memoryID)
	}

	var rows *sql.Rows
	if crossProject {
		rows, err = s.db.Query(`
			SELECT id, project, title, content, type, embedding, outcome, created_at, version
			FROM vec_memories
			WHERE id != ? AND embedding IS NOT NULL
			LIMIT 5000
		`, memoryID)
	} else {
		rows, err = s.db.Query(`
			SELECT id, project, title, content, type, embedding, outcome, created_at, version
			FROM vec_memories
			WHERE id != ? AND embedding IS NOT NULL
			  AND project = (SELECT project FROM vec_memories WHERE id = ?)
			LIMIT 1000
		`, memoryID, memoryID)
	}
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var results []*SearchResult
	for rows.Next() {
		var (
			id, proj, title, content, memType, outcome, createdAt string
			embBlob                                                 []byte
			version                                                 string
		)
		if err := rows.Scan(&id, &proj, &title, &content, &memType, &embBlob, &outcome, &createdAt, &version); err != nil {
			continue
		}

		emb := bytesToFloats(embBlob)
		if len(emb) != len(queryEmbedding) {
			continue
		}

		similarity := cosineSimilarity(queryEmbedding, emb)
		if similarity >= minSimilarity {
			results = append(results, &SearchResult{
				ID:         id,
				Project:    proj,
				Title:      title,
				Content:    content,
				Type:       memType,
				Outcome:    outcome,
				Similarity: similarity,
				CreatedAt:  createdAt,
				Version:    version,
			})
		}
	}

	sortResults(results)
	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// DetectConflicts finds semantic conflicts within a project.
func (s *Store) DetectConflicts(project string, minSimilarity float32) ([]*ConflictPair, error) {
	rows, err := s.db.Query(`
		SELECT id, title, content, type, embedding, outcome
		FROM vec_memories
		WHERE project = ? AND embedding IS NOT NULL
		ORDER BY created_at DESC
		LIMIT 500
	`, project)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	type mem struct {
		id, title, content, memType, outcome string
		embedding                            []float32
	}
	var memories []mem

	for rows.Next() {
		var m mem
		var embBlob []byte
		if err := rows.Scan(&m.id, &m.title, &m.content, &m.memType, &embBlob, &m.outcome); err != nil {
			continue
		}
		m.embedding = bytesToFloats(embBlob)
		if len(m.embedding) > 0 {
			memories = append(memories, m)
		}
	}

	var conflicts []*ConflictPair
	for i := 0; i < len(memories); i++ {
		for j := i + 1; j < len(memories); j++ {
			sim := cosineSimilarity(memories[i].embedding, memories[j].embedding)
			if sim >= minSimilarity {
				if memories[i].memType != memories[j].memType ||
					memories[i].outcome != memories[j].outcome {
					conflicts = append(conflicts, &ConflictPair{
						MemoryA:    memories[i].id,
						TitleA:     memories[i].title,
						ContentA:   memories[i].content,
						TypeA:      memories[i].memType,
						OutcomeA:   memories[i].outcome,
						MemoryB:    memories[j].id,
						TitleB:     memories[j].title,
						ContentB:   memories[j].content,
						TypeB:      memories[j].memType,
						OutcomeB:   memories[j].outcome,
						Similarity: sim,
					})
				}
			}
		}
	}

	return conflicts, nil
}

// Transfer finds knowledge from other projects relevant to the target project.
func (s *Store) Transfer(toProject string, queryEmbedding []float32, limit int, minSimilarity float32) ([]*SearchResult, error) {
	rows, err := s.db.Query(`
		SELECT id, project, title, content, type, embedding, outcome, created_at, version
		FROM vec_memories
		WHERE project != ? AND embedding IS NOT NULL
		ORDER BY created_at DESC
		LIMIT 5000
	`, toProject)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*SearchResult
	for rows.Next() {
		var (
			id, proj, title, content, memType, outcome, createdAt string
			embBlob                                                 []byte
			version                                                 string
		)
		if err := rows.Scan(&id, &proj, &title, &content, &memType, &embBlob, &outcome, &createdAt, &version); err != nil {
			continue
		}

		emb := bytesToFloats(embBlob)
		if len(emb) != len(queryEmbedding) {
			continue
		}

		similarity := cosineSimilarity(queryEmbedding, emb)
		if similarity >= minSimilarity {
			results = append(results, &SearchResult{
				ID:         id,
				Project:    proj,
				Title:      title,
				Content:    content,
				Type:       memType,
				Outcome:    outcome,
				Similarity: similarity,
				CreatedAt:  createdAt,
				Version:    version,
			})
		}
	}

	sortResults(results)
	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// ExportPack exports all memories for a project (and optional release) to portable JSON.
func (s *Store) ExportPack(project, version string) ([]MemoryPackEntry, error) {
	var rows *sql.Rows
	var err error

	if version != "" {
		rows, err = s.db.Query(`
			SELECT id, project, title, content, type, embedding, embedding_model,
			       embedding_dim, tags, outcome, media_type, version
			FROM vec_memories
			WHERE project = ? AND version = ?
			ORDER BY created_at
		`, project, version)
	} else {
		rows, err = s.db.Query(`
			SELECT id, project, title, content, type, embedding, embedding_model,
			       embedding_dim, tags, outcome, media_type, version
			FROM vec_memories
			WHERE project = ?
			ORDER BY created_at
		`, project)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []MemoryPackEntry
	for rows.Next() {
		var e MemoryPackEntry
		var embBlob []byte
		var tagsStr string
		if err := rows.Scan(&e.ID, &e.Project, &e.Title, &e.Content, &e.Type,
			&embBlob, &e.EmbeddingModel, &e.EmbeddingDim,
			&tagsStr, &e.Outcome, &e.MediaType, &e.Version); err != nil {
			continue
		}
		json.Unmarshal([]byte(tagsStr), &e.Tags)
		e.Embedding = bytesToFloats(embBlob)
		entries = append(entries, e)
	}

	return entries, nil
}

// ImportPack imports MemoryPackEntry items into the store.
// Returns the number of memories imported.
func (s *Store) ImportPack(entries []MemoryPackEntry) (int, error) {
	count := 0
	for _, e := range entries {
		tagsJSON := "["
		for i, t := range e.Tags {
			if i > 0 {
				tagsJSON += ","
			}
			tagsJSON += `"` + t + `"`
		}
		tagsJSON += "]"

		_, err := s.db.Exec(`
			INSERT OR REPLACE INTO vec_memories
				(id, project, title, content, type,
				 embedding, embedding_model, embedding_dim,
				 tags, outcome, media_type, version, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'))
		`, e.ID, e.Project, e.Title, e.Content, e.Type,
			floatsToBytes(e.Embedding), e.EmbeddingModel, e.EmbeddingDim,
			tagsJSON, e.Outcome, e.MediaType, e.Version)
		if err != nil {
			continue
		}
		count++
	}
	return count, nil
}

// SearchByType returns memories of a given type without requiring an embedding query.
func (s *Store) SearchByType(memType, project string, limit int) ([]*SearchResult, error) {
	var rows *sql.Rows
	var err error

	if project != "" {
		rows, err = s.db.Query(`
			SELECT id, project, title, content, type, outcome, media_type, version, created_at
			FROM vec_memories
			WHERE type = ? AND project = ?
			ORDER BY updated_at DESC
			LIMIT ?
		`, memType, project, limit)
	} else {
		rows, err = s.db.Query(`
			SELECT id, project, title, content, type, outcome, media_type, version, created_at
			FROM vec_memories
			WHERE type = ?
			ORDER BY updated_at DESC
			LIMIT ?
		`, memType, limit)
	}
	if err != nil {
		return nil, fmt.Errorf("search by type: %w", err)
	}
	defer rows.Close()

	var results []*SearchResult
	for rows.Next() {
		var r SearchResult
		var version string
		if err := rows.Scan(&r.ID, &r.Project, &r.Title, &r.Content, &r.Type,
			&r.Outcome, &r.MediaType, &version, &r.CreatedAt); err != nil {
			continue
		}
		r.Version = version
		results = append(results, &r)
	}
	return results, nil
}

// GetByID retrieves a single memory by ID.
func (s *Store) GetByID(id string) (*VectorMemory, error) {
	row := s.db.QueryRow(`
		SELECT id, project, title, content, type, embedding, embedding_model,
		       embedding_dim, tags, outcome, media_type, version
		FROM vec_memories WHERE id = ?
	`, id)

	var m VectorMemory
	var embBlob []byte
	var tagsStr string
	err := row.Scan(&m.ID, &m.Project, &m.Title, &m.Content, &m.Type,
		&embBlob, &m.EmbeddingModel, &m.EmbeddingDim,
		&tagsStr, &m.Outcome, &m.MediaType, &m.Version)
	if err != nil {
		return nil, err
	}
	m.Embedding = bytesToFloats(embBlob)
	json.Unmarshal([]byte(tagsStr), &m.Tags)
	return &m, nil
}

// ── Sync (git-based portability) ────────────────────────────────

// SyncPush commits and pushes the mnemo database to the configured git remote.
func (s *Store) SyncPush() (string, error) {
	remote := s.GetConfig("sync.remote")
	if remote == "" {
		return "", fmt.Errorf("no sync.remote configured. Run: mnemo config set sync.remote <git-url>")
	}

	gitDir := filepath.Join(s.dir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		// Init git repo
		cmd := exec.Command("git", "-C", s.dir, "init")
		if out, err := cmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("git init: %s: %w", string(out), err)
		}
		// Set remote
		cmd = exec.Command("git", "-C", s.dir, "remote", "add", "origin", remote)
		if out, err := cmd.CombinedOutput(); err != nil {
			// Remote might already exist
			_ = out
		}
	}

	// Stage mnemo.db
	cmd := exec.Command("git", "-C", s.dir, "add", "mnemo.db")
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("git add: %s: %w", string(out), err)
	}

	// Check if there are changes to commit
	statusCmd := exec.Command("git", "-C", s.dir, "diff", "--cached", "--quiet", "mnemo.db")
	if err := statusCmd.Run(); err == nil {
		return "no changes to push", nil
	}

	// Commit
	msg := fmt.Sprintf("mnemo sync %s", time.Now().Format("2006-01-02 15:04"))
	cmd = exec.Command("git", "-C", s.dir, "commit", "-m", msg)
	out, err := cmd.CombinedOutput()
	if err != nil && !strings.Contains(string(out), "nothing to commit") {
		return "", fmt.Errorf("git commit: %s: %w", string(out), err)
	}

	// Push
	cmd = exec.Command("git", "-C", s.dir, "push", "-u", "origin", "main")
	out, err = cmd.CombinedOutput()
	if err != nil {
		// Try master if main fails
		cmd = exec.Command("git", "-C", s.dir, "push", "-u", "origin", "master")
		out, err = cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("git push: %s: %w", string(out), err)
		}
	}

	return fmt.Sprintf("pushed to %s", remote), nil
}

// SyncPull pulls the mnemo database from the configured git remote.
func (s *Store) SyncPull() (string, error) {
	remote := s.GetConfig("sync.remote")
	if remote == "" {
		return "", fmt.Errorf("no sync.remote configured. Run: mnemo config set sync.remote <git-url>")
	}

	gitDir := filepath.Join(s.dir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		// Clone instead of init
		parent := filepath.Dir(s.dir)
		base := filepath.Base(s.dir)
		cmd := exec.Command("git", "clone", remote, filepath.Join(parent, base+"_tmp"))
		out, err := cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("git clone: %s: %w", string(out), err)
		}
		// Move cloned DB into place
		cloneDB := filepath.Join(parent, base+"_tmp", "mnemo.db")
		if _, err := os.Stat(cloneDB); err == nil {
			// Close current DB, replace, reopen
			s.db.Close()
			if err := os.Rename(cloneDB, filepath.Join(s.dir, "mnemo.db")); err != nil {
				return "", fmt.Errorf("replace db: %w", err)
			}
			os.RemoveAll(filepath.Join(parent, base+"_tmp"))
			// Reopen
			db, err := sql.Open("sqlite3", filepath.Join(s.dir, "mnemo.db")+"?_journal_mode=WAL")
			if err != nil {
				return "", fmt.Errorf("reopen db: %w", err)
			}
			s.db = db
			// Init git with remote for future pushes
			exec.Command("git", "-C", s.dir, "init").Run()
			exec.Command("git", "-C", s.dir, "remote", "add", "origin", remote).Run()
		}
		return fmt.Sprintf("cloned from %s", remote), nil
	}

	// Pull
	cmd := exec.Command("git", "-C", s.dir, "pull", "origin", "main")
	out, err := cmd.CombinedOutput()
	if err != nil {
		cmd = exec.Command("git", "-C", s.dir, "pull", "origin", "master")
		out, err = cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("git pull: %s: %w", string(out), err)
		}
	}

	// Reopen DB to pick up changes
	s.db.Close()
	db, err := sql.Open("sqlite3", filepath.Join(s.dir, "mnemo.db")+"?_journal_mode=WAL")
	if err != nil {
		return "", fmt.Errorf("reopen db after pull: %w", err)
	}
	s.db = db

	return fmt.Sprintf("pulled from %s: %s", remote, strings.TrimSpace(string(out))), nil
}

// SyncStatus returns the git sync status.
func (s *Store) SyncStatus() (string, error) {
	remote := s.GetConfig("sync.remote")
	if remote == "" {
		return "no sync.remote configured", nil
	}

	gitDir := filepath.Join(s.dir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return "not initialized (run mnemo sync push first)", nil
	}

	var buf bytes.Buffer
	cmd := exec.Command("git", "-C", s.dir, "status", "--short")
	cmd.Stdout = &buf
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git status: %w", err)
	}
	out := strings.TrimSpace(buf.String())
	if out == "" {
		return "in sync with " + remote, nil
	}
	return out, nil
}

// Stats returns statistics about the vector store.
func (s *Store) Stats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	var total, withEmbedding int
	s.db.QueryRow("SELECT COUNT(*), COUNT(embedding) FROM vec_memories").Scan(&total, &withEmbedding)
	stats["total_memories"] = total
	stats["memories_with_embeddings"] = withEmbedding

	var projects int
	s.db.QueryRow("SELECT COUNT(DISTINCT project) FROM vec_memories").Scan(&projects)
	stats["projects"] = projects

	var releases int
	s.db.QueryRow("SELECT COUNT(*) FROM vec_releases").Scan(&releases)
	stats["releases"] = releases

	return stats, nil
}

// DB returns the underlying *sql.DB connection for use by other packages.
func (s *Store) DB() *sql.DB {
	return s.db
}

// Close closes the database.
func (s *Store) Close() error {
	return s.db.Close()
}

// ── Types ──────────────────────────────────────────────────────────

type VectorMemory struct {
	ID             string
	Project        string
	Title          string
	Content        string
	Type           string
	Embedding      []float32
	EmbeddingModel string
	EmbeddingDim   int
	Tags           []string
	Outcome        string
	MediaType      string
	MediaFile      []byte
	Version        string
}

type SearchResult struct {
	ID         string
	Project    string
	Title      string
	Content    string
	Type       string
	Outcome    string
	MediaType  string
	Version    string
	Similarity float32
	CreatedAt  string
}

type ConflictPair struct {
	MemoryA    string
	TitleA     string
	ContentA   string
	TypeA      string
	OutcomeA   string
	MemoryB    string
	TitleB     string
	ContentB   string
	TypeB      string
	OutcomeB   string
	Similarity float32
}

type ReleaseSnapshot struct {
	ID          string `json:"id"`
	Project     string `json:"project"`
	Version     string `json:"version"`
	Description string `json:"description"`
	MemoryCount int    `json:"memory_count"`
	CreatedAt   string `json:"created_at"`
}

type VersionDiff struct {
	FromVersion string          `json:"from_version"`
	ToVersion   string          `json:"to_version"`
	Added       []MemorySummary `json:"added"`
	Removed     []MemorySummary `json:"removed"`
	Updated     []DiffPair      `json:"updated"`
}

type MemorySummary struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Type    string `json:"type"`
	Outcome string `json:"outcome"`
}

type DiffPair struct {
	Old MemorySummary `json:"old"`
	New MemorySummary `json:"new"`
}

type MemoryPackEntry struct {
	ID             string    `json:"id"`
	Project        string    `json:"project"`
	Title          string    `json:"title"`
	Content        string    `json:"content"`
	Type           string    `json:"type"`
	Embedding      []float32 `json:"embedding"`
	EmbeddingModel string    `json:"embedding_model"`
	EmbeddingDim   int       `json:"embedding_dim"`
	Tags           []string  `json:"tags"`
	Outcome        string    `json:"outcome"`
	MediaType      string    `json:"media_type"`
	Version        string    `json:"version"`
}

// ── Utilities ─────────────────────────────────────────────────────

func floatsToBytes(f []float32) []byte {
	buf := make([]byte, len(f)*4)
	for i, v := range f {
		binary.LittleEndian.PutUint32(buf[i*4:], math.Float32bits(v))
	}
	return buf
}

func bytesToFloats(b []byte) []float32 {
	if len(b)%4 != 0 {
		return nil
	}
	f := make([]float32, len(b)/4)
	for i := range f {
		f[i] = math.Float32frombits(binary.LittleEndian.Uint32(b[i*4:]))
	}
	return f
}

func cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return float32(dot / (math.Sqrt(normA) * math.Sqrt(normB)))
}

func sortResults(results []*SearchResult) {
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Similarity > results[i].Similarity {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
}
