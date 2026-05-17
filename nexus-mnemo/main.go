package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/nexus-sdd/nexus-mnemo/mcp"
	"github.com/nexus-sdd/nexus-mnemo/vec"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]

	switch cmd {
	case "mcp":
		runMCP()
	case "search":
		runSearch()
	case "save":
		runSave()
	case "similar":
		runSimilar()
	case "transfer":
		runTransfer()
	case "release":
		runRelease()
	case "diff":
		runDiff()
	case "releases":
		runReleases()
	case "pack":
		runPack()
	case "conflicts":
		runConflicts()
	case "stats":
		runStats()
	case "setup":
		runSetup()
	case "config":
		if len(os.Args) >= 4 && os.Args[2] == "set" {
			runConfigSet()
		} else {
			runConfig()
		}
	case "version", "--version", "-v":
		fmt.Println("nexus-mnemo v0.2.0")
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Print(`nexus-mnemo — Versionable Vector Memory for AI Agents

Usage:
  mnemo mcp                   Start MCP server (stdio)
  mnemo search <query>        Semantic search
  mnemo save <title>          Save memory with embedding
  mnemo similar <id>          Find similar memories
  mnemo transfer <query>      Transfer knowledge between projects
  mnemo release <proj> <ver>  Create release snapshot
  mnemo diff <proj> <v1> <v2> Compare memories between versions
  mnemo releases <proj>       List releases for a project
  mnemo pack export <proj>    Export memories to portable JSON
  mnemo conflicts             Detect semantic conflicts
  mnemo stats                 Show store statistics
  mnemo setup                 Install and configure dependencies
  mnemo config                Show current configuration
  mnemo config set <k> <v>    Set a configuration value
  mnemo version               Show version

Configuration is stored in ~/.mnemo/mnemo.db (vec_config table).
Environment variables override DB config:
  MNEMO_DIR           Database directory (default: ~/.mnemo)
  OLLAMA_HOST         Override ollama.host
  OLLAMA_EMBED_MODEL  Override embed.model
  EMBEDDER_MOCK       Use mock embedder for testing (true/false)
  MNEMO_PROJECT       Default project (default: current directory name)
`)
}

func getDefaultProject() string {
	if p := os.Getenv("MNEMO_PROJECT"); p != "" {
		return p
	}
	dir, err := os.Getwd()
	if err != nil {
		return "default"
	}
	parts := strings.Split(dir, "/")
	return parts[len(parts)-1]
}

func openStore() (*vec.Store, vec.Embedder) {
	dbDir := os.Getenv("MNEMO_DIR")
	store, err := vec.NewStore(dbDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening Mnemo DB: %v\n", err)
		os.Exit(1)
	}

	embedder, err := vec.NewEmbedderFromStore(store)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error starting embedder: %v\n", err)
		os.Exit(1)
	}

	return store, embedder
}

func runMCP() {
	store, embedder := openStore()
	defer store.Close()

	server := mcp.NewServer(store, embedder)
	fmt.Fprintf(os.Stderr, "nexus-mnemo MCP server started (stdio)\n")
	if err := server.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "MCP server error: %v\n", err)
		os.Exit(1)
	}
}

func runSearch() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Usage: mnemo search <query> [--project <name>] [--limit <n>] [--min-sim <0.0-1.0>]")
		os.Exit(1)
	}

	query := os.Args[2]
	project := flagArg("--project", "")
	limit := flagArgInt("--limit", 5)
	minSim := flagArgFloat("--min-sim", 0.7)

	store, embedder := openStore()
	defer store.Close()

	embedding, err := embedder.Embed(query)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating embedding: %v\n", err)
		os.Exit(1)
	}

	results, err := store.SearchSemantic(embedding, project, limit, minSim)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Search error: %v\n", err)
		os.Exit(1)
	}

	if len(results) == 0 {
		fmt.Printf("No semantically similar memories found for: \"%s\"\n", query)
		return
	}

	fmt.Printf("═══ %d memories found for: \"%s\" ═══\n\n", len(results), query)
	for i, r := range results {
		fmt.Printf("%d. %s (%.0f%% match)\n", i+1, r.Title, r.Similarity*100)
		fmt.Printf("   Project: %s | Type: %s | Outcome: %s", r.Project, r.Type, r.Outcome)
		if r.Version != "" {
			fmt.Printf(" | v%s", r.Version)
		}
		fmt.Println()
		fmt.Printf("   %s\n", r.Content)
		fmt.Printf("   ID: %s\n\n", r.ID)
	}
}

func runSave() {
	if len(os.Args) < 4 {
		fmt.Fprintln(os.Stderr, "Usage: mnemo save <title> <content> [--type <type>] [--project <name>] [--outcome <outcome>] [--tags tag1,tag2] [--media-type <text|image|pdf|audio|video>] [--version <ver>]")
		os.Exit(1)
	}

	title := os.Args[2]
	content := os.Args[3]
	memType := flagArg("--type", "note")
	project := flagArg("--project", getDefaultProject())
	outcome := flagArg("--outcome", "")
	tagsStr := flagArg("--tags", "")
	mediaType := flagArg("--media-type", "text")
	version := flagArg("--version", "")

	var tags []string
	if tagsStr != "" {
		tags = strings.Split(tagsStr, ",")
	}

	store, embedder := openStore()
	defer store.Close()

	textToEmbed := title + "\n" + content
	embedding, err := embedder.Embed(textToEmbed)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating embedding: %v\n", err)
		os.Exit(1)
	}

	id := fmt.Sprintf("vec-%s-%s", project, sanitizeID(title))

	mem := &vec.VectorMemory{
		ID:             id,
		Project:        project,
		Title:          title,
		Content:        content,
		Type:           memType,
		Embedding:      embedding,
		EmbeddingModel: embedder.ModelName(),
		EmbeddingDim:   embedder.Dimension(),
		Tags:           tags,
		Outcome:        outcome,
		MediaType:      mediaType,
		Version:        version,
	}

	if err := store.Save(mem); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Memory saved: %s\n", id)
	fmt.Printf("   Project: %s | Type: %s | Media: %s", project, memType, mediaType)
	if version != "" {
		fmt.Printf(" | v%s", version)
	}
	fmt.Printf(" | Dims: %d\n", embedder.Dimension())
}

func runSimilar() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Usage: mnemo similar <memory_id> [--cross-project]")
		os.Exit(1)
	}

	memoryID := os.Args[2]
	crossProject := hasFlag("--cross-project")
	limit := flagArgInt("--limit", 5)
	minSim := flagArgFloat("--min-sim", 0.7)

	store, _ := openStore()
	defer store.Close()

	results, err := store.Similar(memoryID, crossProject, limit, minSim)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(results) == 0 {
		fmt.Printf("No similar memories found for %s\n", memoryID)
		return
	}

	fmt.Printf("═══ %d memories similar to %s ═══\n\n", len(results), memoryID)
	for i, r := range results {
		fmt.Printf("%d. %s (%.0f%% match) [%s]", i+1, r.Title, r.Similarity*100, r.Project)
		if r.Version != "" {
			fmt.Printf(" v%s", r.Version)
		}
		fmt.Printf("\n   %s\n\n", r.Content)
	}
}

func runTransfer() {
	if len(os.Args) < 4 {
		fmt.Fprintln(os.Stderr, "Usage: mnemo transfer <query> <to_project> [--version <ver>]")
		os.Exit(1)
	}

	query := os.Args[2]
	toProject := os.Args[3]
	limit := flagArgInt("--limit", 5)
	minSim := flagArgFloat("--min-sim", 0.7)

	store, embedder := openStore()
	defer store.Close()

	embedding, err := embedder.Embed(query)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating embedding: %v\n", err)
		os.Exit(1)
	}

	results, err := store.Transfer(toProject, embedding, limit, minSim)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(results) == 0 {
		fmt.Printf("No transferable knowledge found for \"%s\"\n", query)
		return
	}

	fmt.Printf("Transferred Knowledge → %s\n", toProject)
	fmt.Printf("   %d lessons from other projects:\n\n", len(results))
	for i, r := range results {
		fmt.Printf("%d. [%s] %s (%.0f%% relevance)", i+1, r.Project, r.Title, r.Similarity*100)
		if r.Version != "" {
			fmt.Printf(" v%s", r.Version)
		}
		fmt.Printf("\n   Outcome: %s\n", r.Outcome)
		fmt.Printf("   Lesson: %s\n\n", r.Content)
	}
}

func runRelease() {
	if len(os.Args) < 4 {
		fmt.Fprintln(os.Stderr, "Usage: mnemo release <project> <version> [--description <desc>]")
		os.Exit(1)
	}

	project := os.Args[2]
	version := os.Args[3]
	description := flagArg("--description", "")

	store, _ := openStore()
	defer store.Close()

	snapshot, err := store.Release(project, version, description)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating release: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Release created: %s\n", snapshot.ID)
	fmt.Printf("   Project: %s\n", snapshot.Project)
	fmt.Printf("   Version: %s\n", snapshot.Version)
	fmt.Printf("   Description: %s\n", snapshot.Description)
	fmt.Printf("   Memories versioned: %d\n", snapshot.MemoryCount)
	fmt.Printf("   Date: %s\n", snapshot.CreatedAt)
}

func runDiff() {
	if len(os.Args) < 5 {
		fmt.Fprintln(os.Stderr, "Usage: mnemo diff <project> <v1> <v2>")
		os.Exit(1)
	}

	project := os.Args[2]
	v1 := os.Args[3]
	v2 := os.Args[4]

	store, _ := openStore()
	defer store.Close()

	diff, err := store.Diff(project, v1, v2)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error comparing versions: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("═══ Diff: %s → %s (%s) ═══\n\n", v1, v2, project)

	if len(diff.Added) > 0 {
		fmt.Printf("Added (%d):\n", len(diff.Added))
		for _, m := range diff.Added {
			fmt.Printf("   + [%s] %s → %s\n", m.Type, m.Title, m.Outcome)
		}
		fmt.Println()
	}

	if len(diff.Updated) > 0 {
		fmt.Printf("Updated (%d):\n", len(diff.Updated))
		for _, p := range diff.Updated {
			fmt.Printf("   ~ %s: %s → %s\n", p.Old.Title, p.Old.Outcome, p.New.Outcome)
		}
		fmt.Println()
	}

	if len(diff.Removed) > 0 {
		fmt.Printf("Removed (%d):\n", len(diff.Removed))
		for _, m := range diff.Removed {
			fmt.Printf("   - [%s] %s → %s\n", m.Type, m.Title, m.Outcome)
		}
		fmt.Println()
	}

	if len(diff.Added) == 0 && len(diff.Updated) == 0 && len(diff.Removed) == 0 {
		fmt.Println("No changes between these versions.")
	}
}

func runReleases() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Usage: mnemo releases <project>")
		os.Exit(1)
	}

	project := os.Args[2]

	store, _ := openStore()
	defer store.Close()

	releases, err := store.ListReleases(project)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(releases) == 0 {
		fmt.Printf("No releases in project '%s'.\n", project)
		return
	}

	fmt.Printf("═══ Releases of '%s' ═══\n\n", project)
	for _, r := range releases {
		fmt.Printf("%s — %d memories — %s\n", r.Version, r.MemoryCount, r.CreatedAt)
		if r.Description != "" {
			fmt.Printf("   %s\n", r.Description)
		}
		fmt.Println()
	}
}

func runPack() {
	if len(os.Args) < 4 || os.Args[2] != "export" {
		fmt.Fprintln(os.Stderr, "Usage: mnemo pack export <project> [--version <ver>] [--output <file>]")
		os.Exit(1)
	}

	project := os.Args[3]
	version := flagArg("--version", "")
	output := flagArg("--output", "")

	store, _ := openStore()
	defer store.Close()

	entries, err := store.ExportPack(project, version)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error exporting: %v\n", err)
		os.Exit(1)
	}

	pack := map[string]interface{}{
		"pack_version": "1.0",
		"project":      project,
		"version":      version,
		"exported_at":  fmt.Sprintf("%d", os.Getpid()),
		"entries":      entries,
	}

	data, _ := json.MarshalIndent(pack, "", "  ")

	if output != "" {
		if err := os.WriteFile(output, data, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Pack exported: %s (%d memories)\n", output, len(entries))
	} else {
		fmt.Println(string(data))
	}
}

func runConflicts() {
	project := flagArg("--project", getDefaultProject())
	minSim := flagArgFloat("--min-sim", 0.75)

	store, _ := openStore()
	defer store.Close()

	conflicts, err := store.DetectConflicts(project, minSim)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(conflicts) == 0 {
		fmt.Printf("No semantic conflicts detected in '%s'\n", project)
		return
	}

	fmt.Printf("%d potential semantic conflicts in '%s'\n\n", len(conflicts), project)
	for i, c := range conflicts {
		fmt.Printf("%d. Conflict (%.0f%% similarity)\n", i+1, c.Similarity*100)
		fmt.Printf("   A: [%s] %s → %s\n", c.TypeA, c.TitleA, c.OutcomeA)
		fmt.Printf("   B: [%s] %s → %s\n", c.TypeB, c.TitleB, c.OutcomeB)
		fmt.Println()
	}
}

func runStats() {
	store, _ := openStore()
	defer store.Close()

	stats, err := store.Stats()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	pretty, _ := json.MarshalIndent(stats, "", "  ")
	fmt.Println(string(pretty))
}

func runConfig() {
	store, _ := openStore()
	defer store.Close()

	cfg := store.AllConfig()
	if len(cfg) == 0 {
		fmt.Println("No configuration found.")
		return
	}

	fmt.Println("Current configuration:")
	for _, k := range []string{"embed.model", "embed.dims", "ollama.host", "mnemo.version"} {
		if v, ok := cfg[k]; ok {
			override := ""
			switch k {
			case "embed.model":
				if env := os.Getenv("OLLAMA_EMBED_MODEL"); env != "" && env != v {
					override = fmt.Sprintf(" (env override: %s)", env)
				}
			case "ollama.host":
				if env := os.Getenv("OLLAMA_HOST"); env != "" && env != v {
					override = fmt.Sprintf(" (env override: %s)", env)
				}
			}
			fmt.Printf("  %-20s = %s%s\n", k, v, override)
		}
	}
}

func runConfigSet() {
	if len(os.Args) < 4 {
		fmt.Fprintln(os.Stderr, "Usage: mnemo config set <key> <value>")
		os.Exit(1)
	}
	key := os.Args[3]
	value := os.Args[4]

	store, _ := openStore()
	defer store.Close()

	if err := store.SetConfig(key, value); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Config set: %s = %s\n", key, value)
}

func runSetup() {
	fmt.Println("nexus-mnemo setup")
	fmt.Println()

	// Open DB first
	dbDir := os.Getenv("MNEMO_DIR")
	store, err := vec.NewStore(dbDir)
	if err != nil {
		fmt.Printf("Error opening DB: %v\n", err)
		os.Exit(1)
	}
	defer store.Close()

	// Show current config
	fmt.Println("Current configuration:")
	cfg := store.AllConfig()
	for _, k := range []string{"embed.model", "embed.dims", "ollama.host", "mnemo.version"} {
		if v, ok := cfg[k]; ok {
			fmt.Printf("  %-20s = %s\n", k, v)
		}
	}
	fmt.Println()

	// Check Ollama connectivity
	fmt.Print("Ollama... ")
	currentModel := store.GetConfig("embed.model")
	embedder, err := vec.NewEmbedderFromStore(store)
	if err != nil {
		fmt.Printf("not available: %v\n", err)
		fmt.Println()
		fmt.Println("To install Ollama:")
		fmt.Println("  curl -fsSL https://ollama.com/install.sh | sh")
		fmt.Printf("  ollama pull %s\n", currentModel)
	} else {
		fmt.Printf("%s (%d dims)\n", embedder.ModelName(), embedder.Dimension())
	}

	// DB stats
	fmt.Print("Mnemo DB... ")
	stats, _ := store.Stats()
	fmt.Printf("%v memories, %v with embeddings, %v releases\n",
		stats["total_memories"], stats["memories_with_embeddings"], stats["releases"])

	fmt.Println()
	fmt.Println("To change the embedding model:")
	fmt.Printf("  mnemo config set embed.model <model-name>\n")
	fmt.Println()
	fmt.Println("To add to Claude Code:")
	fmt.Println("  claude mcp add mnemo -- mnemo mcp")
	fmt.Println()
	fmt.Println("To add to OpenCode:")
	fmt.Println("  opencode mcp add mnemo -- mnemo mcp")
}

// ── CLI Helpers ───────────────────────────────────────────────────────

func flagArg(name, defaultVal string) string {
	for i, arg := range os.Args {
		if arg == name && i+1 < len(os.Args) {
			return os.Args[i+1]
		}
	}
	return defaultVal
}

func flagArgInt(name string, defaultVal int) int {
	val := flagArg(name, "")
	if val == "" {
		return defaultVal
	}
	var n int
	fmt.Sscanf(val, "%d", &n)
	return n
}

func flagArgFloat(name string, defaultVal float32) float32 {
	val := flagArg(name, "")
	if val == "" {
		return defaultVal
	}
	var f float32
	fmt.Sscanf(val, "%f", &f)
	return f
}

func hasFlag(name string) bool {
	for _, arg := range os.Args {
		if arg == name {
			return true
		}
	}
	return false
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
