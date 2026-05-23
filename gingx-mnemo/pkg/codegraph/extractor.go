package codegraph

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// Extractor parses source code and emits Nodes and Edges.
// Uses regex-based extraction; swappable for tree-sitter later.
type Extractor struct{}

func NewExtractor() *Extractor {
	return &Extractor{}
}

// ExtractResult holds nodes and edges found in a single file.
type ExtractResult struct {
	Nodes []Node
	Edges []Edge
}

// ExtractFile parses a single source file and returns symbols.
func (e *Extractor) ExtractFile(path string, content []byte) (*ExtractResult, error) {
	lang := DetectLanguage(path)
	switch lang {
	case "go":
		return e.extractGo(path, string(content))
	case "python":
		return e.extractPython(path, string(content))
	case "typescript", "javascript":
		return e.extractTSJS(path, string(content), lang)
	default:
		return nil, fmt.Errorf("unsupported language: %s", lang)
	}
}

// ── Go Extractor ────────────────────────────────────────────────

var (
	goFuncRe     = regexp.MustCompile(`func\s+(?:\(([^)]*)\)\s+)?(\w+)\s*\(([^)]*)\)`)
	goTypeRe     = regexp.MustCompile(`type\s+(\w+)\s+(struct|interface)\s*\{`)
	goVarRe      = regexp.MustCompile(`var\s+(\w+)\s+`)
	goImportRe   = regexp.MustCompile(`import\s+(?:\w+\s+)?\"([^\"]+)\"`)
	goImportBlk  = regexp.MustCompile(`import\s*\(([^)]+)\)`)
	goPackageRe  = regexp.MustCompile(`package\s+(\w+)`)
)

func (e *Extractor) extractGo(path string, content string) (*ExtractResult, error) {
	result := &ExtractResult{}
	lines := strings.Split(content, "\n")
	fileName := filepath.Base(path)
	modName := strings.TrimSuffix(fileName, ".go")

	// Package declaration
	pkgMatch := goPackageRe.FindStringSubmatch(content)
	pkgName := modName
	if len(pkgMatch) >= 2 {
		pkgName = pkgMatch[1]
	}

	// Module node
	modID := fmt.Sprintf("%s:%s:module", path, pkgName)
	result.Nodes = append(result.Nodes, Node{
		ID: modID, Kind: KindModule, Name: pkgName,
		QualifiedName: pkgName, FilePath: path, Language: "go",
		StartLine: 1, EndLine: len(lines), IsExported: true,
	})

	// Functions and methods
	for _, m := range goFuncRe.FindAllStringSubmatch(content, -1) {
		for i := range m {
			m[i] = strings.TrimSpace(m[i])
		}
		receiver := m[1]
		name := m[2]
		params := m[3]

		if name == "" || isBuiltin(name) {
			continue
		}

		lineNum := findLine(lines, "func "+m[2])
		if lineNum == 0 {
			lineNum = findLine(lines, "func ("+receiver+") "+name)
		}

		kind := KindFunction
		qualified := pkgName + "." + name
		parentID := modID

		if receiver != "" {
			kind = KindMethod
			recvType := strings.Split(receiver, " ")[0]
			recvType = strings.TrimPrefix(recvType, "*")
			qualified = pkgName + "." + recvType + "." + name
		}

		id := fmt.Sprintf("%s:%s:%s", path, qualified, kind)
		isExported := len(name) > 0 && name[0] >= 'A' && name[0] <= 'Z'

		node := Node{
			ID: id, Kind: kind, Name: name,
			QualifiedName: qualified, FilePath: path, Language: "go",
			StartLine: lineNum, EndLine: lineNum,
			Signature: fmt.Sprintf("func %s(%s)", name, params),
			IsExported: isExported, ParentID: parentID,
		}
		result.Nodes = append(result.Nodes, node)
		result.Edges = append(result.Edges, Edge{
			Source: parentID, Target: id, Kind: EdgeContains,
		})
	}

	// Types (struct, interface)
	for _, m := range goTypeRe.FindAllStringSubmatch(content, -1) {
		name := m[1]
		typeKind := m[2]
		id := fmt.Sprintf("%s:%s.%s:%s", path, pkgName, name, typeKind)
		lineNum := findLine(lines, "type "+name)

		nodeKind := KindStruct
		if typeKind == "interface" {
			nodeKind = KindInterface
		}

		isExported := len(name) > 0 && name[0] >= 'A' && name[0] <= 'Z'
		node := Node{
			ID: id, Kind: nodeKind, Name: name,
			QualifiedName: pkgName + "." + name, FilePath: path, Language: "go",
			StartLine: lineNum, EndLine: lineNum,
			Signature: fmt.Sprintf("type %s %s", name, typeKind),
			IsExported: isExported, ParentID: modID,
		}
		result.Nodes = append(result.Nodes, node)
		result.Edges = append(result.Edges, Edge{
			Source: modID, Target: id, Kind: EdgeContains,
		})
	}

	// Imports
	e.extractGoImports(path, content, modID, result)

	return result, nil
}

func (e *Extractor) extractGoImports(path, content, modID string, result *ExtractResult) {
	// Single imports
	for _, m := range goImportRe.FindAllStringSubmatch(content, -1) {
		pkgPath := m[1]
		parts := strings.Split(pkgPath, "/")
		name := parts[len(parts)-1]
		id := fmt.Sprintf("%s:import:%s", path, name)
		result.Nodes = append(result.Nodes, Node{
			ID: id, Kind: KindImport, Name: name,
			QualifiedName: pkgPath, FilePath: path, Language: "go",
		})
		result.Edges = append(result.Edges, Edge{
			Source: modID, Target: id, Kind: EdgeImports,
		})
	}
}

// ── Python Extractor ────────────────────────────────────────────

var (
	pyFuncRe   = regexp.MustCompile(`def\s+(\w+)\s*\(([^)]*)\)`)
	pyClassRe  = regexp.MustCompile(`class\s+(\w+)\s*(?:\(([^)]*)\))?\s*:`)
	pyImportRe = regexp.MustCompile(`(?:from\s+(\S+)\s+)?import\s+(.+)`)
	pyAssignRe = regexp.MustCompile(`^(\w+)\s*[:=]\s*`)
)

func (e *Extractor) extractPython(path string, content string) (*ExtractResult, error) {
	result := &ExtractResult{}
	lines := strings.Split(content, "\n")
	fileName := filepath.Base(path)
	modName := strings.TrimSuffix(fileName, ".py")

	// Module node
	modID := fmt.Sprintf("%s:%s:module", path, modName)
	result.Nodes = append(result.Nodes, Node{
		ID: modID, Kind: KindModule, Name: modName,
		QualifiedName: modName, FilePath: path, Language: "python",
		StartLine: 1, EndLine: len(lines),
	})

	// Functions
	for _, m := range pyFuncRe.FindAllStringSubmatch(content, -1) {
		name := m[1]
		params := m[2]
		if name == "" || strings.HasPrefix(name, "_") && name != "__init__" {
			continue
		}

		lineNum := findLine(lines, "def "+name)
		kind := KindFunction
		qualified := modName + "." + name

		id := fmt.Sprintf("%s:%s:%s", path, qualified, kind)
		node := Node{
			ID: id, Kind: kind, Name: name,
			QualifiedName: qualified, FilePath: path, Language: "python",
			StartLine: lineNum, EndLine: lineNum,
			Signature: fmt.Sprintf("def %s(%s)", name, params),
			ParentID: modID,
		}
		result.Nodes = append(result.Nodes, node)
		result.Edges = append(result.Edges, Edge{
			Source: modID, Target: id, Kind: EdgeContains,
		})
	}

	// Classes
	currentClassID := ""
	for _, m := range pyClassRe.FindAllStringSubmatch(content, -1) {
		name := m[1]
		if name == "" {
			continue
		}

		lineNum := findLine(lines, "class "+name)
		qualified := modName + "." + name
		id := fmt.Sprintf("%s:%s:%s", path, qualified, KindClass)

		node := Node{
			ID: id, Kind: KindClass, Name: name,
			QualifiedName: qualified, FilePath: path, Language: "python",
			StartLine: lineNum, EndLine: lineNum,
			ParentID: modID,
		}
		result.Nodes = append(result.Nodes, node)
		result.Edges = append(result.Edges, Edge{
			Source: modID, Target: id, Kind: EdgeContains,
		})

		// Track for method assignment
		if m[2] != "" {
			for _, base := range strings.Split(m[2], ",") {
				base = strings.TrimSpace(base)
				baseID := fmt.Sprintf("%s:%s.%s:%s", path, modName, base, KindClass)
				result.Edges = append(result.Edges, Edge{
					Source: id, Target: baseID, Kind: EdgeExtends,
				})
			}
		}

		currentClassID = id
		_ = currentClassID // methods can be detected if we track class scope
	}

	// Imports
	for _, m := range pyImportRe.FindAllStringSubmatch(content, -1) {
		fromMod := m[1]
		importWhat := m[2]
		if importWhat == "" {
			continue
		}
		for _, imp := range strings.Split(importWhat, ",") {
			imp = strings.TrimSpace(imp)
			if imp == "" {
				continue
			}
			id := fmt.Sprintf("%s:import:%s", path, imp)
			result.Nodes = append(result.Nodes, Node{
				ID: id, Kind: KindImport, Name: imp,
				QualifiedName: fromMod + "." + imp, FilePath: path, Language: "python",
			})
			result.Edges = append(result.Edges, Edge{
				Source: modID, Target: id, Kind: EdgeImports,
			})
		}
	}

	return result, nil
}

// ── TypeScript / JavaScript Extractor ──────────────────────────

var (
	tsFuncRe     = regexp.MustCompile(`(?:export\s+)?(?:async\s+)?function\s+(\w+)\s*\(([^)]*)\)`)
	tsArrowRe    = regexp.MustCompile(`(?:export\s+)?(?:const|let|var)\s+(\w+)\s*[:=]\s*(?:async\s+)?\(([^)]*)\)\s*=>`)
	tsClassRe    = regexp.MustCompile(`(?:export\s+)?class\s+(\w+)\s*(?:extends\s+(\w+))?`)
	tsInterfaceRe = regexp.MustCompile(`(?:export\s+)?interface\s+(\w+)`)
	tsImportRe   = regexp.MustCompile(`import\s+(?:\{[^}]+\}|\w+|\*\s+as\s+\w+)\s+from\s+['\"]([^'\"]+)['\"]`)
	tsImportAllRe = regexp.MustCompile(`import\s+\*\s+as\s+(\w+)\s+from\s+['\"]([^'\"]+)['\"]`)
	tsImportNamed = regexp.MustCompile(`import\s+\{([^}]+)\}\s+from\s+['\"]([^'\"]+)['\"]`)
	tsImportDef   = regexp.MustCompile(`import\s+(\w+)\s+from\s+['\"]([^'\"]+)['\"]`)
)

func (e *Extractor) extractTSJS(path string, content string, lang string) (*ExtractResult, error) {
	result := &ExtractResult{}
	lines := strings.Split(content, "\n")
	fileName := filepath.Base(path)
	modName := strings.TrimSuffix(fileName, filepath.Ext(fileName))

	modID := fmt.Sprintf("%s:%s:module", path, modName)
	result.Nodes = append(result.Nodes, Node{
		ID: modID, Kind: KindModule, Name: modName,
		QualifiedName: modName, FilePath: path, Language: lang,
		StartLine: 1, EndLine: len(lines),
	})

	// Functions (regular and arrow)
	e.extractTSFunctions(path, content, lines, modID, modName, result)
	// Classes
	e.extractTSClasses(path, content, lines, modID, modName, result)
	// Interfaces
	e.extractTSInterfaces(path, content, lines, modID, modName, result)
	// Imports
	e.extractTSImports(path, content, modID, result)

	return result, nil
}

func (e *Extractor) extractTSFunctions(path, content string, lines []string, modID, modName string, result *ExtractResult) {
	for _, m := range tsFuncRe.FindAllStringSubmatch(content, -1) {
		name := m[1]
		params := m[2]
		lineNum := findLine(lines, "function "+name)
		qualified := modName + "." + name
		id := fmt.Sprintf("%s:%s:function", path, qualified)

		result.Nodes = append(result.Nodes, Node{
			ID: id, Kind: KindFunction, Name: name,
			QualifiedName: qualified, FilePath: path, Language: "typescript",
			StartLine: lineNum, EndLine: lineNum,
			Signature: fmt.Sprintf("function %s(%s)", name, params),
			ParentID: modID,
		})
		result.Edges = append(result.Edges, Edge{
			Source: modID, Target: id, Kind: EdgeContains,
		})
	}

	for _, m := range tsArrowRe.FindAllStringSubmatch(content, -1) {
		name := m[1]
		params := m[2]
		lineNum := findLine(lines, name)
		qualified := modName + "." + name
		id := fmt.Sprintf("%s:%s:function", path, qualified)

		result.Nodes = append(result.Nodes, Node{
			ID: id, Kind: KindFunction, Name: name,
			QualifiedName: qualified, FilePath: path, Language: "typescript",
			StartLine: lineNum, EndLine: lineNum,
			Signature: fmt.Sprintf("const %s = (%s) =>", name, params),
			ParentID: modID,
		})
		result.Edges = append(result.Edges, Edge{
			Source: modID, Target: id, Kind: EdgeContains,
		})
	}
}

func (e *Extractor) extractTSClasses(path, content string, lines []string, modID, modName string, result *ExtractResult) {
	for _, m := range tsClassRe.FindAllStringSubmatch(content, -1) {
		name := m[1]
		extends := m[2]
		lineNum := findLine(lines, "class "+name)
		qualified := modName + "." + name
		id := fmt.Sprintf("%s:%s:class", path, qualified)

		result.Nodes = append(result.Nodes, Node{
			ID: id, Kind: KindClass, Name: name,
			QualifiedName: qualified, FilePath: path, Language: "typescript",
			StartLine: lineNum, EndLine: lineNum,
			ParentID: modID,
		})
		result.Edges = append(result.Edges, Edge{
			Source: modID, Target: id, Kind: EdgeContains,
		})

		if extends != "" {
			extID := fmt.Sprintf("%s:%s.%s:class", path, modName, extends)
			result.Edges = append(result.Edges, Edge{
				Source: id, Target: extID, Kind: EdgeExtends,
			})
		}
	}
}

func (e *Extractor) extractTSInterfaces(path, content string, lines []string, modID, modName string, result *ExtractResult) {
	for _, m := range tsInterfaceRe.FindAllStringSubmatch(content, -1) {
		name := m[1]
		lineNum := findLine(lines, "interface "+name)
		qualified := modName + "." + name
		id := fmt.Sprintf("%s:%s:interface", path, qualified)

		result.Nodes = append(result.Nodes, Node{
			ID: id, Kind: KindInterface, Name: name,
			QualifiedName: qualified, FilePath: path, Language: "typescript",
			StartLine: lineNum, EndLine: lineNum,
			ParentID: modID,
		})
		result.Edges = append(result.Edges, Edge{
			Source: modID, Target: id, Kind: EdgeContains,
		})
	}
}

func (e *Extractor) extractTSImports(path, content, modID string, result *ExtractResult) {
	// import X from 'y'
	for _, m := range tsImportDef.FindAllStringSubmatch(content, -1) {
		name := m[1]
		from := m[2]
		id := fmt.Sprintf("%s:import:%s", path, name)
		result.Nodes = append(result.Nodes, Node{
			ID: id, Kind: KindImport, Name: name,
			QualifiedName: from, FilePath: path, Language: "typescript",
		})
		result.Edges = append(result.Edges, Edge{
			Source: modID, Target: id, Kind: EdgeImports,
		})
	}

	// import { a, b } from 'y'
	for _, m := range tsImportNamed.FindAllStringSubmatch(content, -1) {
		names := m[1]
		from := m[2]
		for _, name := range strings.Split(names, ",") {
			name = strings.TrimSpace(name)
			if name == "" {
				continue
			}
			id := fmt.Sprintf("%s:import:%s", path, name)
			result.Nodes = append(result.Nodes, Node{
				ID: id, Kind: KindImport, Name: name,
				QualifiedName: from, FilePath: path, Language: "typescript",
			})
			result.Edges = append(result.Edges, Edge{
				Source: modID, Target: id, Kind: EdgeImports,
			})
		}
	}
}

// ── Helpers ────────────────────────────────────────────────────

func findLine(lines []string, substr string) int {
	for i, line := range lines {
		if strings.Contains(line, substr) {
			return i + 1 // 1-based line numbers
		}
	}
	return 0
}

func isBuiltin(name string) bool {
	switch name {
	case "init", "main", "len", "cap", "make", "append", "copy", "close", "delete", "panic", "recover":
		return true
	}
	return false
}
