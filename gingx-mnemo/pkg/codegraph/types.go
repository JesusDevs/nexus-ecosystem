package codegraph

// Node represents a code symbol.
type Node struct {
	ID            string `json:"id"`
	Kind          string `json:"kind"`
	Name          string `json:"name"`
	QualifiedName string `json:"qualified_name"`
	FilePath      string `json:"file_path"`
	Language      string `json:"language"`
	StartLine     int    `json:"start_line"`
	EndLine       int    `json:"end_line"`
	Signature     string `json:"signature,omitempty"`
	IsExported    bool   `json:"is_exported"`
	ParentID      string `json:"parent_id,omitempty"`
}

// Edge represents a relationship between two nodes.
type Edge struct {
	ID     int64  `json:"id"`
	Source string `json:"source"`
	Target string `json:"target"`
	Kind   string `json:"kind"`
	Line   int    `json:"line,omitempty"`
}

// FileInfo tracks an indexed source file.
type FileInfo struct {
	Path        string `json:"path"`
	ContentHash string `json:"content_hash"`
	Language    string `json:"language"`
	Size        int64  `json:"size"`
	IndexedAt   string `json:"indexed_at"`
	NodeCount   int    `json:"node_count"`
}

// SearchResult from FTS5.
type SearchResult struct {
	Node     Node   `json:"node"`
	Rank     float64 `json:"rank"`
	Snippet  string `json:"snippet,omitempty"`
}

// CallerResult: who calls a given symbol.
type CallerResult struct {
	Caller Node `json:"caller"`
	Edge   Edge `json:"edge"`
}

// CalleeResult: who a given symbol calls.
type CalleeResult struct {
	Callee Node `json:"callee"`
	Edge   Edge `json:"edge"`
}

// IndexResult from indexing a project.
type IndexResult struct {
	FilesIndexed int `json:"files_indexed"`
	NodesCreated int `json:"nodes_created"`
	EdgesCreated int `json:"edges_created"`
	FilesErrored int `json:"files_errored"`
}
