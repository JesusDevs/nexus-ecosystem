package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/nexus-sdd/nexus-mnemo/vec"
)

// ── Protocolo MCP (JSON-RPC 2.0 sobre stdio) ──────────────────────

type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ToolDef struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	InputSchema JSONSchema `json:"inputSchema"`
}

type JSONSchema struct {
	Type       string              `json:"type"`
	Properties map[string]PropDef  `json:"properties"`
	Required   []string            `json:"required,omitempty"`
}

type PropDef struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

// ── Servidor ────────────────────────────────────────────────────────

type Server struct {
	store    *vec.Store
	embedder vec.Embedder
	reader   *bufio.Reader
	writer   io.Writer
}

func NewServer(store *vec.Store, embedder vec.Embedder) *Server {
	return &Server{
		store:    store,
		embedder: embedder,
		reader:   bufio.NewReader(os.Stdin),
		writer:   os.Stdout,
	}
}

func (s *Server) Run() error {
	for {
		line, err := s.reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var req JSONRPCRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			s.writeError(nil, -32700, "Parse error: "+err.Error())
			continue
		}

		s.handleRequest(&req)
	}
}

func (s *Server) handleRequest(req *JSONRPCRequest) {
	switch req.Method {
	case "initialize":
		s.handleInitialize(req)
	case "tools/list":
		s.handleToolsList(req)
	case "tools/call":
		s.handleToolsCall(req)
	default:
		s.writeError(req.ID, -32601, fmt.Sprintf("Method not found: %s", req.Method))
	}
}

func (s *Server) handleInitialize(req *JSONRPCRequest) {
	result := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"serverInfo": map[string]interface{}{
			"name":    "nexus-mnemo",
			"version": "0.2.0",
		},
		"capabilities": map[string]interface{}{
			"tools": map[string]bool{},
		},
	}
	s.writeResult(req.ID, result)
}

// ── tools/list ───────────────────────────────────────────────────

var tools = []ToolDef{
	{
		Name:        "mem_search_semantic",
		Description: "Búsqueda semántica de memorias por significado (no por keywords). Usa embeddings para encontrar memorias conceptualmente similares aunque usen palabras distintas.",
		InputSchema: JSONSchema{
			Type: "object",
			Properties: map[string]PropDef{
				"query":          {Type: "string", Description: "Texto de búsqueda (en lenguaje natural, no keywords)"},
				"project":        {Type: "string", Description: "Proyecto (opcional). Si se omite, busca en todos los proyectos."},
				"limit":          {Type: "integer", Description: "Máximo de resultados (default: 5)"},
				"min_similarity": {Type: "number", Description: "Similitud mínima 0.0-1.0 (default: 0.7)"},
				"version":        {Type: "string", Description: "Filtrar por versión/release (opcional)"},
			},
			Required: []string{"query"},
		},
	},
	{
		Name:        "mem_similar",
		Description: "Encuentra memorias similares a una existente. Útil para detectar patrones repetidos o decisiones contradictorias.",
		InputSchema: JSONSchema{
			Type: "object",
			Properties: map[string]PropDef{
				"memory_id":      {Type: "string", Description: "ID de la memoria de referencia"},
				"cross_project":  {Type: "boolean", Description: "Si es true, busca en TODOS los proyectos (default: false)"},
				"limit":          {Type: "integer", Description: "Máximo de resultados (default: 5)"},
				"min_similarity": {Type: "number", Description: "Similitud mínima (default: 0.7)"},
			},
			Required: []string{"memory_id"},
		},
	},
	{
		Name:        "mem_transfer",
		Description: "Transfiere conocimiento de otros proyectos al actual. Busca memorias relevantes en proyectos anteriores para no repetir errores.",
		InputSchema: JSONSchema{
			Type: "object",
			Properties: map[string]PropDef{
				"query":          {Type: "string", Description: "Qué estás por hacer (ej: 'implementar pagos con Stripe')"},
				"to_project":     {Type: "string", Description: "Proyecto actual (recibe el conocimiento)"},
				"limit":          {Type: "integer", Description: "Máximo de resultados (default: 5)"},
				"min_similarity": {Type: "number", Description: "Similitud mínima (default: 0.7)"},
			},
			Required: []string{"query", "to_project"},
		},
	},
	{
		Name:        "mem_save",
		Description: "Guarda una nueva memoria vectorial con embedding semántico. Usa esto para registrar bugs, decisiones, feedback, errores resueltos, o lecciones aprendidas. Soporta tipos de medio: text, image, pdf, audio, video. Opcionalmente asocia la memoria a una versión/release.",
		InputSchema: JSONSchema{
			Type: "object",
			Properties: map[string]PropDef{
				"title":      {Type: "string", Description: "Título descriptivo de la memoria"},
				"content":    {Type: "string", Description: "Contenido detallado: qué pasó, cómo se resolvió, qué se aprendió"},
				"type":       {Type: "string", Description: "Tipo: bug, decision, architecture, note, feedback, lesson"},
				"project":    {Type: "string", Description: "Proyecto (default: directorio actual)"},
				"outcome":    {Type: "string", Description: "Resultado: solucionado, falló, parcial, pendiente"},
				"tags":       {Type: "string", Description: "Tags separados por coma (ej: 'auth,stripe,webhook')"},
				"media_type": {Type: "string", Description: "Tipo de medio: text, image, pdf, audio, video (default: text)"},
				"version":    {Type: "string", Description: "Versión/release (opcional). Si no se especifica, queda sin versionar."},
			},
			Required: []string{"title", "content"},
		},
	},
	{
		Name:        "mem_release",
		Description: "Crea un snapshot de release: taggea todas las memorias no versionadas de un proyecto con un número de versión. Ideal para capturar el conocimiento generado en un ciclo de desarrollo.",
		InputSchema: JSONSchema{
			Type: "object",
			Properties: map[string]PropDef{
				"project":     {Type: "string", Description: "Proyecto a versionar"},
				"version":     {Type: "string", Description: "Número de versión (ej: 'v1.0.0', 'v2.1.0')"},
				"description": {Type: "string", Description: "Descripción del release (ej: 'MVP con auth OAuth2')"},
			},
			Required: []string{"project", "version"},
		},
	},
	{
		Name:        "mem_diff",
		Description: "Compara las memorias entre dos versiones de un proyecto. Muestra qué conocimiento se agregó, se actualizó o se removió.",
		InputSchema: JSONSchema{
			Type: "object",
			Properties: map[string]PropDef{
				"project":      {Type: "string", Description: "Proyecto"},
				"from_version": {Type: "string", Description: "Versión base (ej: 'v1.0.0')"},
				"to_version":   {Type: "string", Description: "Versión a comparar (ej: 'v1.1.0')"},
			},
			Required: []string{"project", "from_version", "to_version"},
		},
	},
	{
		Name:        "mem_list_releases",
		Description: "Lista todos los releases/snapshots de un proyecto con su metadata.",
		InputSchema: JSONSchema{
			Type: "object",
			Properties: map[string]PropDef{
				"project": {Type: "string", Description: "Proyecto"},
			},
			Required: []string{"project"},
		},
	},
	{
		Name:        "mem_detect_conflicts_semantic",
		Description: "Detecta conflictos semánticos en las memorias de un proyecto. Encuentra decisiones contradictorias o patrones incompatibles.",
		InputSchema: JSONSchema{
			Type: "object",
			Properties: map[string]PropDef{
				"project":        {Type: "string", Description: "Proyecto a analizar"},
				"min_similarity": {Type: "number", Description: "Umbral de similitud para considerar conflicto (default: 0.75)"},
			},
			Required: []string{"project"},
		},
	},
	{
		Name:        "agent_dispatch",
		Description: "Dispatches a task to a specialized Nexus agent persona (supervisor, po-agent, ux-agent, architect-agent, dev-agent, qa-agent, devops-agent). Loads the agent's persona definition, relevant tech stack skills from the profile, prior context from mnemo semantic search, and proactive interrogation prompts if the profile enables 'Tony Stark mode'. Returns the assembled agent prompt ready for execution.",
		InputSchema: JSONSchema{
			Type: "object",
			Properties: map[string]PropDef{
				"agent":      {Type: "string", Description: "Agent to dispatch: supervisor, po-agent, ux-agent, architect-agent, dev-agent, qa-agent, devops-agent"},
				"task":       {Type: "string", Description: "Task description in natural language"},
				"profile":    {Type: "string", Description: "Profile name (default: developer). Available: fullstack-python-langgraph, fullstack-go, react-nextjs, fullstack, minimal"},
				"hdu_id":     {Type: "string", Description: "OpenSpec HDU context (optional)"},
				"tech_stack": {Type: "string", Description: "Comma-separated tech stack override (optional)"},
				"mode":       {Type: "string", Description: "'prompt' returns assembled prompt (default: prompt)"},
			},
			Required: []string{"agent", "task"},
		},
	},
}

func (s *Server) handleToolsList(req *JSONRPCRequest) {
	s.writeResult(req.ID, map[string]interface{}{
		"tools": tools,
	})
}

// ── tools/call ───────────────────────────────────────────────────

type ToolsCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

func (s *Server) handleToolsCall(req *JSONRPCRequest) {
	var params ToolsCallParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		s.writeError(req.ID, -32602, "Invalid params: "+err.Error())
		return
	}

	var result interface{}
	var err error

	switch params.Name {
	case "mem_search_semantic":
		result, err = s.handleSearchSemantic(params.Arguments)
	case "mem_similar":
		result, err = s.handleSimilar(params.Arguments)
	case "mem_transfer":
		result, err = s.handleTransfer(params.Arguments)
	case "mem_save":
		result, err = s.handleSave(params.Arguments)
	case "mem_release":
		result, err = s.handleRelease(params.Arguments)
	case "mem_diff":
		result, err = s.handleDiff(params.Arguments)
	case "mem_list_releases":
		result, err = s.handleListReleases(params.Arguments)
	case "mem_detect_conflicts_semantic":
		result, err = s.handleDetectConflicts(params.Arguments)
		case "agent_dispatch":
			result, err = s.handleAgentDispatch(params.Arguments)
	default:
		s.writeError(req.ID, -32602, "Unknown tool: "+params.Name)
		return
	}

	if err != nil {
		s.writeError(req.ID, -32000, err.Error())
		return
	}

	s.writeResult(req.ID, map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": result,
			},
		},
	})
}

// ── Handlers ──────────────────────────────────────────────────────────

func (s *Server) handleSearchSemantic(args map[string]interface{}) (string, error) {
	query := getString(args, "query")
	project := getString(args, "project")
	limit := getInt(args, "limit", 5)
	minSim := getFloat(args, "min_similarity", 0.7)

	embedding, err := s.embedder.Embed(query)
	if err != nil {
		return "", fmt.Errorf("error generando embedding: %w", err)
	}

	results, err := s.store.SearchSemantic(embedding, project, limit, minSim)
	if err != nil {
		return "", err
	}

	return formatSearchResults(results, "mem_search_semantic", query), nil
}

func (s *Server) handleSimilar(args map[string]interface{}) (string, error) {
	memoryID := getString(args, "memory_id")
	crossProject := getBool(args, "cross_project")
	limit := getInt(args, "limit", 5)
	minSim := getFloat(args, "min_similarity", 0.7)

	results, err := s.store.Similar(memoryID, crossProject, limit, minSim)
	if err != nil {
		return "", err
	}

	return formatSearchResults(results, "mem_similar", memoryID), nil
}

func (s *Server) handleTransfer(args map[string]interface{}) (string, error) {
	query := getString(args, "query")
	toProject := getString(args, "to_project")
	limit := getInt(args, "limit", 5)
	minSim := getFloat(args, "min_similarity", 0.7)

	embedding, err := s.embedder.Embed(query)
	if err != nil {
		return "", err
	}

	results, err := s.store.Transfer(toProject, embedding, limit, minSim)
	if err != nil {
		return "", err
	}

	return formatTransferResults(results, query, toProject), nil
}

func (s *Server) handleSave(args map[string]interface{}) (string, error) {
	title := getString(args, "title")
	content := getString(args, "content")
	memType := getString(args, "type")
	if memType == "" {
		memType = "note"
	}
	project := getString(args, "project")
	if project == "" {
		if wd, err := os.Getwd(); err == nil {
			parts := strings.Split(wd, "/")
			project = parts[len(parts)-1]
		}
	}
	outcome := getString(args, "outcome")
	tagsStr := getString(args, "tags")
	mediaType := getString(args, "media_type")
	if mediaType == "" {
		mediaType = "text"
	}
	version := getString(args, "version")

	var tags []string
	if tagsStr != "" {
		for _, t := range strings.Split(tagsStr, ",") {
			t = strings.TrimSpace(t)
			if t != "" {
				tags = append(tags, t)
			}
		}
	}

	textToEmbed := title + "\n" + content
	embedding, err := s.embedder.Embed(textToEmbed)
	if err != nil {
		return "", fmt.Errorf("error generando embedding: %w", err)
	}

	id := fmt.Sprintf("vec-%s-%s", project, sanitizeID(title))

	mem := &vec.VectorMemory{
		ID:             id,
		Project:        project,
		Title:          title,
		Content:        content,
		Type:           memType,
		Embedding:      embedding,
		EmbeddingModel: s.embedder.ModelName(),
		EmbeddingDim:   s.embedder.Dimension(),
		Tags:           tags,
		Outcome:        outcome,
		MediaType:      mediaType,
		Version:        version,
	}

	if err := s.store.Save(mem); err != nil {
		return "", fmt.Errorf("error guardando memoria: %w", err)
	}

	return fmt.Sprintf("✅ Memoria guardada: %s\n- Proyecto: %s\n- Tipo: %s\n- Medio: %s\n- Versión: %s\n- Outcome: %s\n- Tags: %s\n- Dims: %d",
		id, project, memType, mediaType, version, outcome, strings.Join(tags, ", "), s.embedder.Dimension()), nil
}

func (s *Server) handleRelease(args map[string]interface{}) (string, error) {
	project := getString(args, "project")
	version := getString(args, "version")
	description := getString(args, "description")

	snapshot, err := s.store.Release(project, version, description)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("🏷️ Release creado: %s\n- Proyecto: %s\n- Versión: %s\n- Descripción: %s\n- Memorias versionadas: %d\n- Fecha: %s",
		snapshot.ID, snapshot.Project, snapshot.Version, snapshot.Description, snapshot.MemoryCount, snapshot.CreatedAt), nil
}

func (s *Server) handleDiff(args map[string]interface{}) (string, error) {
	project := getString(args, "project")
	fromVersion := getString(args, "from_version")
	toVersion := getString(args, "to_version")

	diff, err := s.store.Diff(project, fromVersion, toVersion)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## 📊 Diff: %s → %s (%s)\n\n", fromVersion, toVersion, project))

	if len(diff.Added) > 0 {
		sb.WriteString(fmt.Sprintf("### ✅ Agregadas (%d)\n", len(diff.Added)))
		for _, m := range diff.Added {
			sb.WriteString(fmt.Sprintf("- [%s] %s → *%s*\n", m.Type, m.Title, m.Outcome))
		}
		sb.WriteString("\n")
	}

	if len(diff.Updated) > 0 {
		sb.WriteString(fmt.Sprintf("### 🔄 Actualizadas (%d)\n", len(diff.Updated)))
		for _, p := range diff.Updated {
			sb.WriteString(fmt.Sprintf("- **%s**: *%s* → *%s*\n", p.Old.Title, p.Old.Outcome, p.New.Outcome))
		}
		sb.WriteString("\n")
	}

	if len(diff.Removed) > 0 {
		sb.WriteString(fmt.Sprintf("### ❌ Removidas (%d)\n", len(diff.Removed)))
		for _, m := range diff.Removed {
			sb.WriteString(fmt.Sprintf("- [%s] %s → *%s*\n", m.Type, m.Title, m.Outcome))
		}
		sb.WriteString("\n")
	}

	if len(diff.Added) == 0 && len(diff.Updated) == 0 && len(diff.Removed) == 0 {
		sb.WriteString("Sin cambios entre estas versiones.\n")
	}

	return sb.String(), nil
}

func (s *Server) handleListReleases(args map[string]interface{}) (string, error) {
	project := getString(args, "project")

	releases, err := s.store.ListReleases(project)
	if err != nil {
		return "", err
	}

	if len(releases) == 0 {
		return fmt.Sprintf("No hay releases en el proyecto '%s'.", project), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## 🏷️ Releases de '%s'\n\n", project))
	for _, r := range releases {
		sb.WriteString(fmt.Sprintf("- **%s** — %d memorias — %s\n  %s\n", r.Version, r.MemoryCount, r.CreatedAt, r.Description))
	}

	return sb.String(), nil
}

func (s *Server) handleDetectConflicts(args map[string]interface{}) (string, error) {
	project := getString(args, "project")
	minSim := getFloat(args, "min_similarity", 0.75)

	conflicts, err := s.store.DetectConflicts(project, minSim)
	if err != nil {
		return "", err
	}

	return formatConflicts(conflicts, project), nil
}

// ── Formateo ─────────────────────────────────────────────────────────

func formatSearchResults(results []*vec.SearchResult, toolName, query string) string {
	if len(results) == 0 {
		return fmt.Sprintf("No se encontraron memorias semánticamente similares a: \"%s\"", query)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## %d memorias encontradas para: \"%s\"\n\n", len(results), query))

	for i, r := range results {
		sb.WriteString(fmt.Sprintf("### %d. %s (%.0f%% match)\n", i+1, r.Title, r.Similarity*100))
		sb.WriteString(fmt.Sprintf("- **Proyecto**: %s\n", r.Project))
		sb.WriteString(fmt.Sprintf("- **Tipo**: %s\n", r.Type))
		sb.WriteString(fmt.Sprintf("- **Medio**: %s\n", r.MediaType))
		sb.WriteString(fmt.Sprintf("- **Resultado**: %s\n", r.Outcome))
		if r.Version != "" {
			sb.WriteString(fmt.Sprintf("- **Versión**: %s\n", r.Version))
		}
		sb.WriteString(fmt.Sprintf("- **Contenido**: %s\n", r.Content))
		sb.WriteString(fmt.Sprintf("- **ID**: %s\n\n", r.ID))
	}

	return sb.String()
}

func formatTransferResults(results []*vec.SearchResult, query, toProject string) string {
	if len(results) == 0 {
		return fmt.Sprintf("No se encontró conocimiento transferible para \"%s\" desde otros proyectos.", query)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## 🧠 Conocimiento Transferido → %s\n", toProject))
	sb.WriteString(fmt.Sprintf("Se encontraron %d memorias relevantes de otros proyectos:\n\n", len(results)))

	for i, r := range results {
		sb.WriteString(fmt.Sprintf("### %d. %s (%.0f%% relevancia, proyecto: %s)\n", i+1, r.Title, r.Similarity*100, r.Project))
		sb.WriteString(fmt.Sprintf("- **Tipo**: %s\n", r.Type))
		sb.WriteString(fmt.Sprintf("- **Medio**: %s\n", r.MediaType))
		sb.WriteString(fmt.Sprintf("- **Resultado original**: %s\n", r.Outcome))
		if r.Version != "" {
			sb.WriteString(fmt.Sprintf("- **Versión**: %s\n", r.Version))
		}
		sb.WriteString(fmt.Sprintf("- **Lección**: %s\n", r.Content))
		sb.WriteString(fmt.Sprintf("- **ID**: %s\n\n", r.ID))
	}

	return sb.String()
}

func formatConflicts(conflicts []*vec.ConflictPair, project string) string {
	if len(conflicts) == 0 {
		return fmt.Sprintf("✅ No se detectaron conflictos semánticos en el proyecto '%s'.", project)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## ⚠️ %d posibles conflictos semánticos en '%s'\n\n", len(conflicts), project))

	for i, c := range conflicts {
		sb.WriteString(fmt.Sprintf("### Conflicto %d (%.0f%% similitud semántica)\n", i+1, c.Similarity*100))
		sb.WriteString(fmt.Sprintf("- **A**: [%s] %s → *%s*\n", c.TypeA, c.TitleA, c.OutcomeA))
		sb.WriteString(fmt.Sprintf("- **B**: [%s] %s → *%s*\n", c.TypeB, c.TitleB, c.OutcomeB))
		sb.WriteString(fmt.Sprintf("- **Contenido A**: %s\n", truncate(c.ContentA, 100)))
		sb.WriteString(fmt.Sprintf("- **Contenido B**: %s\n\n", truncate(c.ContentB, 100)))
	}

	return sb.String()
}

// ── Helpers ──────────────────────────────────────────────────────────

func (s *Server) writeResult(id interface{}, result interface{}) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	data, _ := json.Marshal(resp)
	fmt.Fprintf(s.writer, "%s\n", data)
}

func (s *Server) writeError(id interface{}, code int, message string) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &RPCError{Code: code, Message: message},
	}
	data, _ := json.Marshal(resp)
	fmt.Fprintf(s.writer, "%s\n", data)
}

func getString(args map[string]interface{}, key string) string {
	if v, ok := args[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getInt(args map[string]interface{}, key string, defaultVal int) int {
	if v, ok := args[key]; ok {
		switch n := v.(type) {
		case float64:
			return int(n)
		case int:
			return n
		}
	}
	return defaultVal
}

func getFloat(args map[string]interface{}, key string, defaultVal float32) float32 {
	if v, ok := args[key]; ok {
		switch n := v.(type) {
		case float64:
			return float32(n)
		case float32:
			return n
		}
	}
	return defaultVal
}

func getBool(args map[string]interface{}, key string) bool {
	if v, ok := args[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func sanitizeID(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, ":", "")
	if len(s) > 40 {
		s = s[:40]
	}
	return s
}
