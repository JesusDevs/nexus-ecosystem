package codegraph

import (
	"fmt"
	"strings"
)

// Traverser resolves references and traverses the code graph.
type Traverser struct {
	store *Store
}

func NewTraverser(store *Store) *Traverser {
	return &Traverser{store: store}
}

// ResolveCalls scans all nodes and creates `calls` edges where one function
// calls another by name. This is a best-effort regex-based approach;
// tree-sitter would give more precision.
func (t *Traverser) ResolveCalls(filePath string, content string, nodes []Node) []Edge {
	var edges []Edge
	lines := strings.Split(content, "\n")

	for _, node := range nodes {
		if node.Kind != KindFunction && node.Kind != KindMethod {
			continue
		}
		// Find calls to other known functions within the same file
		for _, other := range nodes {
			if other.ID == node.ID {
				continue
			}
			if other.Kind != KindFunction && other.Kind != KindMethod {
				continue
			}
			// Check if node's body references other.Name
			start := node.StartLine
			end := node.EndLine
			if end <= start {
				end = start + 20 // approximate scope
			}
			for i := start; i < end && i <= len(lines); i++ {
				line := lines[i-1]
				if strings.Contains(line, other.Name+"(") {
					edges = append(edges, Edge{
						Source: node.ID,
						Target: other.ID,
						Kind:   EdgeCalls,
						Line:   i,
					})
					break
				}
			}
		}
	}
	return edges
}

// GetImpactRadius finds nodes that depend on the given node (callers → their callers).
// Limited to depth 2 for performance.
func (t *Traverser) GetImpactRadius(nodeID string, maxDepth int) ([]Node, error) {
	visited := make(map[string]bool)
	var impacted []Node

	var dfs func(id string, depth int)
	dfs = func(id string, depth int) {
		if depth > maxDepth || visited[id] {
			return
		}
		visited[id] = true

		callers, err := t.store.GetCallers(id)
		if err != nil {
			return
		}
		for _, cr := range callers {
			if !visited[cr.Caller.ID] {
				impacted = append(impacted, cr.Caller)
				dfs(cr.Caller.ID, depth+1)
			}
		}
	}

	dfs(nodeID, 0)
	return impacted, nil
}

// FormatCallers returns a formatted string for MCP output.
func FormatCallers(callers []CallerResult) string {
	if len(callers) == 0 {
		return "No callers found."
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## %d caller(s)\n\n", len(callers)))
	for i, cr := range callers {
		sb.WriteString(fmt.Sprintf("### %d. %s (%s)\n", i+1, cr.Caller.Name, cr.Caller.Kind))
		sb.WriteString(fmt.Sprintf("- **File**: %s:%d\n", cr.Caller.FilePath, cr.Caller.StartLine))
		sb.WriteString(fmt.Sprintf("- **Qualified**: %s\n", cr.Caller.QualifiedName))
		if cr.Caller.Signature != "" {
			sb.WriteString(fmt.Sprintf("- **Signature**: %s\n", cr.Caller.Signature))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

// FormatCallees returns a formatted string for MCP output.
func FormatCallees(callees []CalleeResult) string {
	if len(callees) == 0 {
		return "No callees found."
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## %d callee(s)\n\n", len(callees)))
	for i, cr := range callees {
		sb.WriteString(fmt.Sprintf("### %d. %s (%s)\n", i+1, cr.Callee.Name, cr.Callee.Kind))
		sb.WriteString(fmt.Sprintf("- **File**: %s:%d\n", cr.Callee.FilePath, cr.Callee.StartLine))
		sb.WriteString(fmt.Sprintf("- **Qualified**: %s\n", cr.Callee.QualifiedName))
		if cr.Callee.Signature != "" {
			sb.WriteString(fmt.Sprintf("- **Signature**: %s\n", cr.Callee.Signature))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

// FormatSearchResults returns a formatted string for MCP output.
func FormatSearchResults(results []SearchResult, query string) string {
	if len(results) == 0 {
		return fmt.Sprintf("No symbols found matching \"%s\"", query)
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## %d symbol(s) matching \"%s\"\n\n", len(results), query))
	for i, r := range results {
		sb.WriteString(fmt.Sprintf("### %d. %s (%s)\n", i+1, r.Node.Name, r.Node.Kind))
		sb.WriteString(fmt.Sprintf("- **File**: %s:%d\n", r.Node.FilePath, r.Node.StartLine))
		sb.WriteString(fmt.Sprintf("- **Qualified**: %s\n", r.Node.QualifiedName))
		if r.Node.Signature != "" {
			sb.WriteString(fmt.Sprintf("- **Signature**: %s\n", r.Node.Signature))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

// FormatImpact returns a formatted string for impact radius.
func FormatImpact(impacted []Node, centerName string) string {
	if len(impacted) == 0 {
		return fmt.Sprintf("No impact radius found for %s (nothing depends on it).", centerName)
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Impact radius for %s — %d dependents\n\n", centerName, len(impacted)))
	for i, n := range impacted {
		sb.WriteString(fmt.Sprintf("%d. **%s** (%s) — %s:%d\n", i+1, n.Name, n.Kind, n.FilePath, n.StartLine))
	}
	return sb.String()
}
