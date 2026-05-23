package codegraph

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Indexer scans a project directory and indexes all supported source files.
type Indexer struct {
	store     *Store
	extractor *Extractor
	traverser *Traverser
}

func NewIndexer(store *Store) *Indexer {
	return &Indexer{
		store:     store,
		extractor: NewExtractor(),
		traverser: NewTraverser(store),
	}
}

// IndexProject scans all supported files under rootPath and indexes them.
func (idx *Indexer) IndexProject(rootPath string) (*IndexResult, error) {
	result := &IndexResult{}

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip inaccessible files
		}
		if info.IsDir() {
			base := filepath.Base(path)
			if strings.HasPrefix(base, ".") || base == "node_modules" || base == "vendor" ||
				base == "__pycache__" || base == "dist" || base == "build" {
				return filepath.SkipDir
			}
			return nil
		}

		lang := DetectLanguage(path)
		if lang == "unknown" {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil || len(content) > 1048576 { // skip files > 1MB
			result.FilesErrored++
			return nil
		}

		extractResult, err := idx.extractor.ExtractFile(path, content)
		if err != nil {
			result.FilesErrored++
			return nil
		}

		// Resolve calls within file
		callEdges := idx.traverser.ResolveCalls(path, string(content), extractResult.Nodes)
		extractResult.Edges = append(extractResult.Edges, callEdges...)

		// Write to DB
		for _, node := range extractResult.Nodes {
			if err := idx.store.InsertNode(&node); err != nil {
				continue
			}
			result.NodesCreated++
		}
		for _, edge := range extractResult.Edges {
			if err := idx.store.InsertEdge(&edge); err != nil {
				continue
			}
			result.EdgesCreated++
		}

		// Track file
		idx.store.UpsertFile(&FileInfo{
			Path:        path,
			ContentHash: HashFile(content),
			Language:    lang,
			Size:        info.Size(),
			NodeCount:   len(extractResult.Nodes),
		})

		result.FilesIndexed++
		return nil
	})

	if err != nil {
		return result, fmt.Errorf("walk error: %w", err)
	}

	return result, nil
}

// IndexFile indexes a single file and returns the extracted nodes.
func (idx *Indexer) IndexFile(path string) (*IndexResult, error) {
	result := &IndexResult{}

	content, err := os.ReadFile(path)
	if err != nil {
		return result, fmt.Errorf("read file: %w", err)
	}

	extractResult, err := idx.extractor.ExtractFile(path, content)
	if err != nil {
		return result, fmt.Errorf("extract: %w", err)
	}

	callEdges := idx.traverser.ResolveCalls(path, string(content), extractResult.Nodes)
	extractResult.Edges = append(extractResult.Edges, callEdges...)

	for _, node := range extractResult.Nodes {
		if err := idx.store.InsertNode(&node); err != nil {
			continue
		}
		result.NodesCreated++
	}
	for _, edge := range extractResult.Edges {
		if err := idx.store.InsertEdge(&edge); err != nil {
			continue
		}
		result.EdgesCreated++
	}

	info, _ := os.Stat(path)
	size := int64(0)
	if info != nil {
		size = info.Size()
	}
	idx.store.UpsertFile(&FileInfo{
		Path:        path,
		ContentHash: HashFile(content),
		Language:    DetectLanguage(path),
		Size:        size,
		NodeCount:   len(extractResult.Nodes),
		IndexedAt:   time.Now().Format(time.RFC3339),
	})

	result.FilesIndexed = 1
	return result, nil
}

// FindSymbolByName searches for a node by name (exact or prefix match).
func (idx *Indexer) FindSymbolByName(name string, language string, limit int) ([]SearchResult, error) {
	return idx.store.SearchSymbols(name, language, limit)
}

// GetCallers returns all callers of a symbol found by name.
func (idx *Indexer) GetCallersByName(name string) ([]CallerResult, error) {
	results, err := idx.store.SearchSymbols(name, "", 5)
	if err != nil || len(results) == 0 {
		return nil, fmt.Errorf("symbol not found: %s", name)
	}
	return idx.store.GetCallers(results[0].Node.ID)
}

// GetCalleesByName returns all callees of a symbol found by name.
func (idx *Indexer) GetCalleesByName(name string) ([]CalleeResult, error) {
	results, err := idx.store.SearchSymbols(name, "", 5)
	if err != nil || len(results) == 0 {
		return nil, fmt.Errorf("symbol not found: %s", name)
	}
	return idx.store.GetCallees(results[0].Node.ID)
}

// Stats returns code intelligence statistics.
func (idx *Indexer) Stats() (map[string]int, error) {
	return idx.store.Stats()
}

// GetImpactByName finds impact radius of a symbol by name.
func (idx *Indexer) GetImpactByName(name string) ([]Node, error) {
	results, err := idx.store.SearchSymbols(name, "", 5)
	if err != nil || len(results) == 0 {
		return nil, fmt.Errorf("symbol not found: %s", name)
	}
	return idx.traverser.GetImpactRadius(results[0].Node.ID, 2)
}
