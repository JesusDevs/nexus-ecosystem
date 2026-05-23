package mcp

import (
	"fmt"
	"strings"

	"github.com/gingx-sdd/gingx-mnemo/pkg/codegraph"
)

// CodeTools returns the MCP tool definitions for code intelligence.
func CodeTools() []ToolDef {
	return []ToolDef{
		{
			Name:        "code_search",
			Description: "Search code symbols by name or keyword. Finds functions, classes, methods, interfaces, structs, variables, and imports across the indexed codebase. Use this to locate where a symbol is defined and understand code structure.",
			InputSchema: JSONSchema{
				Type: "object",
				Properties: map[string]PropDef{
					"query":    {Type: "string", Description: "Symbol name or keyword to search (e.g., 'UserService', 'authenticate', 'handleRequest')"},
					"language": {Type: "string", Description: "Filter by language: go, python, typescript, javascript (optional)"},
					"limit":    {Type: "integer", Description: "Maximum results (default: 10)"},
				},
				Required: []string{"query"},
			},
		},
		{
			Name:        "code_callers",
			Description: "Find all functions/methods that CALL a given symbol. Answers 'who depends on this code?' — essential for understanding impact before making changes.",
			InputSchema: JSONSchema{
				Type: "object",
				Properties: map[string]PropDef{
					"symbol": {Type: "string", Description: "Function or method name to find callers for (e.g., 'UserService.authenticate')"},
				},
				Required: []string{"symbol"},
			},
		},
		{
			Name:        "code_callees",
			Description: "Find all functions/methods CALLED BY a given symbol. Answers 'what does this code depend on?' — trace the dependency chain of a function.",
			InputSchema: JSONSchema{
				Type: "object",
				Properties: map[string]PropDef{
					"symbol": {Type: "string", Description: "Function or method name to find callees for (e.g., 'handleLogin')"},
				},
				Required: []string{"symbol"},
			},
		},
		{
			Name:        "code_impact",
			Description: "Calculate the impact radius of a symbol — shows all code that transitively depends on it (up to 2 levels deep). Use before renaming, refactoring, or changing a function's signature.",
			InputSchema: JSONSchema{
				Type: "object",
				Properties: map[string]PropDef{
					"symbol": {Type: "string", Description: "Function or method name to calculate impact for"},
				},
				Required: []string{"symbol"},
			},
		},
		{
			Name:        "code_index_status",
			Description: "Show code intelligence index status — how many files, symbols, and edges are indexed, with breakdown by language.",
			InputSchema: JSONSchema{
				Type:       "object",
				Properties: map[string]PropDef{},
				Required:   []string{},
			},
		},
	}
}

// CodeHandler handles code intelligence tool calls.
type CodeHandler struct {
	indexer *codegraph.Indexer
}

func NewCodeHandler(indexer *codegraph.Indexer) *CodeHandler {
	return &CodeHandler{indexer: indexer}
}

// Execute runs a code intelligence tool and returns the result string.
func (h *CodeHandler) Execute(toolName string, args map[string]interface{}) (string, error) {
	switch toolName {
	case "code_search":
		return h.handleSearch(args)
	case "code_callers":
		return h.handleCallers(args)
	case "code_callees":
		return h.handleCallees(args)
	case "code_impact":
		return h.handleImpact(args)
	case "code_index_status":
		return h.handleStatus()
	default:
		return "", fmt.Errorf("unknown code tool: %s", toolName)
	}
}

func (h *CodeHandler) handleSearch(args map[string]interface{}) (string, error) {
	query := getString(args, "query")
	language := getString(args, "language")
	limit := getInt(args, "limit", 10)

	results, err := h.indexer.FindSymbolByName(query, language, limit)
	if err != nil {
		return "", fmt.Errorf("search error: %w", err)
	}

	return codegraph.FormatSearchResults(results, query), nil
}

func (h *CodeHandler) handleCallers(args map[string]interface{}) (string, error) {
	symbol := getString(args, "symbol")

	callers, err := h.indexer.GetCallersByName(symbol)
	if err != nil {
		return "", err
	}

	return codegraph.FormatCallers(callers), nil
}

func (h *CodeHandler) handleCallees(args map[string]interface{}) (string, error) {
	symbol := getString(args, "symbol")

	callees, err := h.indexer.GetCalleesByName(symbol)
	if err != nil {
		return "", err
	}

	return codegraph.FormatCallees(callees), nil
}

func (h *CodeHandler) handleImpact(args map[string]interface{}) (string, error) {
	symbol := getString(args, "symbol")

	impacted, err := h.indexer.GetImpactByName(symbol)
	if err != nil {
		return "", err
	}

	return codegraph.FormatImpact(impacted, symbol), nil
}

func (h *CodeHandler) handleStatus() (string, error) {
	stats, err := h.indexer.Stats()
	if err != nil {
		return "", fmt.Errorf("stats error: %w", err)
	}

	var sb strings.Builder
	sb.WriteString("## Code Intelligence Index Status\n\n")
	sb.WriteString(fmt.Sprintf("- **Nodes**: %d\n", stats["nodes"]))
	sb.WriteString(fmt.Sprintf("- **Edges**: %d\n", stats["edges"]))
	sb.WriteString(fmt.Sprintf("- **Files**: %d\n\n", stats["files"]))

	sb.WriteString("### By Language\n\n")
	for k, v := range stats {
		if strings.HasPrefix(k, "nodes_") {
			lang := strings.TrimPrefix(k, "nodes_")
			sb.WriteString(fmt.Sprintf("- **%s**: %d symbols\n", lang, v))
		}
	}

	return sb.String(), nil
}
