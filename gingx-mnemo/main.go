package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/gingx-sdd/gingx-mnemo/mcp"
	"github.com/gingx-sdd/gingx-mnemo/pkg/codegraph"
	"github.com/gingx-sdd/gingx-mnemo/vec"
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
	case "sync":
		runSync()
	case "hdu":
		runHDU()
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
	case "code":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: mnemo code <index|search|callers|callees|impact|status>")
			os.Exit(1)
		}
		runCodeCommand()
	case "version", "--version", "-v":
		fmt.Println("gingx-mnemo v0.2.0")
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Print(`gingx-mnemo — Versionable Vector Memory for AI Agents

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
  mnemo pack import <file>    Import memories from a pack JSON
  mnemo sync push|pull|status Sync mnemo DB via git remote
  mnemo hdu save|search|list  Manage HDUs in vector memory
  mnemo conflicts             Detect semantic conflicts
  mnemo stats                 Show store statistics
  mnemo setup                 Install and configure dependencies
  mnemo config                Show current configuration
  mnemo config set <k> <v>    Set a configuration value
  mnemo version               Show version

Code Intelligence (native, no npm):
  mnemo code index [path]     Index source code in current directory
  mnemo code search <name>    Search symbols by name
  mnemo code callers <name>   Find callers of a symbol
  mnemo code callees <name>   Find callees of a symbol
  mnemo code impact <name>    Calculate impact radius
  mnemo code status           Show code index statistics
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
	fmt.Fprintf(os.Stderr, "gingx-mnemo MCP server started (stdio)\n")
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
	if len(os.Args) < 4 {
		fmt.Fprintln(os.Stderr, "Usage: mnemo pack export <project> [--version <ver>] [--output <file>]")
		fmt.Fprintln(os.Stderr, "       mnemo pack import <file>")
		os.Exit(1)
	}

	subCmd := os.Args[2]

	switch subCmd {
	case "export":
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

	case "import":
		filePath := os.Args[3]

		store, _ := openStore()
		defer store.Close()

		data, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
			os.Exit(1)
		}

		var pack struct {
			PackVersion string                `json:"pack_version"`
			Project     string                `json:"project"`
			Entries     []vec.MemoryPackEntry `json:"entries"`
		}
		if err := json.Unmarshal(data, &pack); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing pack: %v\n", err)
			os.Exit(1)
		}

		count, err := store.ImportPack(pack.Entries)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error importing: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Pack imported: %d/%d memories\n", count, len(pack.Entries))

	default:
		fmt.Fprintf(os.Stderr, "Unknown pack command: %s\n", subCmd)
		os.Exit(1)
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
	fmt.Println("gingx-mnemo setup")
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

// ── Sync Commands ─────────────────────────────────────────────────────

func runSync() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Usage: mnemo sync <push|pull|status>")
		os.Exit(1)
	}

	store, _ := openStore()
	defer store.Close()

	switch os.Args[2] {
	case "push":
		msg, err := store.SyncPush()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Sync push error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(msg)

	case "pull":
		msg, err := store.SyncPull()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Sync pull error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(msg)

	case "status":
		msg, err := store.SyncStatus()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Sync status error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(msg)

	default:
		fmt.Fprintf(os.Stderr, "Unknown sync command: %s\n", os.Args[2])
		fmt.Fprintln(os.Stderr, "Usage: mnemo sync <push|pull|status>")
		os.Exit(1)
	}
}

// ── HDU Commands ──────────────────────────────────────────────────────

func runHDU() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Usage: mnemo hdu <save|search|list|transfer|get> [...]")
		os.Exit(1)
	}

	switch os.Args[2] {
	case "save":
		if len(os.Args) < 5 {
			fmt.Fprintln(os.Stderr, "Usage: mnemo hdu save <id> --title <title> [--phase <phase>] [--status <status>] [--project <name>] [--content <content>]")
			os.Exit(1)
		}
		hduID := os.Args[3]
		title := flagArg("--title", hduID)
		project := flagArg("--project", getDefaultProject())
		phase := flagArg("--phase", "init")
		status := flagArg("--status", "active")
		content := flagArg("--content", title)

		store, embedder := openStore()
		defer store.Close()

		textToEmbed := fmt.Sprintf("HDU: %s\nTitle: %s\nPhase: %s\nStatus: %s\n%s", hduID, title, phase, status, content)
		embedding, err := embedder.Embed(textToEmbed)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating embedding: %v\n", err)
			os.Exit(1)
		}

		memID := fmt.Sprintf("hdu-%s-%s", project, hduID)
		mem := &vec.VectorMemory{
			ID:             memID,
			Project:        project,
			Title:          fmt.Sprintf("HDU: %s", title),
			Content:        textToEmbed,
			Type:           "hdu",
			Embedding:      embedding,
			EmbeddingModel: embedder.ModelName(),
			EmbeddingDim:   embedder.Dimension(),
			Tags:           []string{hduID, phase, status},
			Outcome:        status,
		}

		if err := store.Save(mem); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving HDU: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("HDU saved: %s\n", memID)
		fmt.Printf("   Title: %s\n", title)
		fmt.Printf("   Phase: %s | Status: %s\n", phase, status)

	case "search":
		if len(os.Args) < 4 {
			fmt.Fprintln(os.Stderr, "Usage: mnemo hdu search <query> [--project <name>] [--limit <n>]")
			os.Exit(1)
		}
		query := os.Args[3]
		project := flagArg("--project", "")
		limit := flagArgInt("--limit", 10)

		store, embedder := openStore()
		defer store.Close()

		embedding, err := embedder.Embed(query)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating embedding: %v\n", err)
			os.Exit(1)
		}

		// Search semantic + filter by type=hdu via SearchByType
		results, err := store.SearchSemantic(embedding, project, limit*2, 0.3)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("═══ HDUs matching: \"%s\" ═══\n\n", query)
		count := 0
		for _, r := range results {
			if r.Type == "hdu" {
				fmt.Printf("  %s (%.0f%% match) [%s]\n", r.Title, r.Similarity*100, r.Project)
				fmt.Printf("  Status: %s | %s\n", r.Outcome, r.CreatedAt)
				fmt.Println()
				count++
				if count >= limit {
					break
				}
			}
		}
		if count == 0 {
			fmt.Println("  No HDUs found.")
		}

	case "list":
		project := flagArg("--project", getDefaultProject())
		limit := flagArgInt("--limit", 20)

		store, _ := openStore()
		defer store.Close()

		results, err := store.SearchByType("hdu", project, limit)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if len(results) == 0 {
			fmt.Printf("No HDUs in project '%s'\n", project)
			return
		}

		fmt.Printf("═══ %d HDUs in '%s' ═══\n\n", len(results), project)
		for _, r := range results {
			fmt.Printf("  %s (%.50s...)\n", r.Title, r.Content)
			fmt.Printf("  Status: %s | %s\n\n", r.Outcome, r.CreatedAt)
		}

	case "transfer":
		if len(os.Args) < 5 {
			fmt.Fprintln(os.Stderr, "Usage: mnemo hdu transfer <hdu_id> <to_project>")
			os.Exit(1)
		}
		hduID := os.Args[3]
		toProject := os.Args[4]

		store, _ := openStore()
		defer store.Close()

		memID := fmt.Sprintf("hdu-%s-%s", flagArg("--from-project", getDefaultProject()), hduID)
		mem, err := store.GetByID(memID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "HDU not found: %s\n", memID)
			os.Exit(1)
		}

		mem.ID = fmt.Sprintf("hdu-%s-%s", toProject, hduID)
		mem.Project = toProject
		if err := store.Save(mem); err != nil {
			fmt.Fprintf(os.Stderr, "Error transferring HDU: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("HDU transferred: %s → %s\n", hduID, toProject)

	case "get":
		if len(os.Args) < 4 {
			fmt.Fprintln(os.Stderr, "Usage: mnemo hdu get <id> [--project <name>]")
			os.Exit(1)
		}
		hduID := os.Args[3]
		project := flagArg("--project", getDefaultProject())

		store, _ := openStore()
		defer store.Close()

		memID := fmt.Sprintf("hdu-%s-%s", project, hduID)
		mem, err := store.GetByID(memID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "HDU not found: %s\n", memID)
			os.Exit(1)
		}

		fmt.Printf("═══ HDU: %s ═══\n", hduID)
		fmt.Printf("Title: %s\n", mem.Title)
		fmt.Printf("Project: %s\n", mem.Project)
		fmt.Printf("Type: %s\n", mem.Type)
		fmt.Printf("Outcome: %s\n", mem.Outcome)
		fmt.Printf("Tags: %v\n", mem.Tags)
		fmt.Printf("Version: %s\n", mem.Version)
		fmt.Printf("Content:\n%s\n", mem.Content)

	default:
		fmt.Fprintf(os.Stderr, "Unknown hdu command: %s\n", os.Args[2])
		fmt.Fprintln(os.Stderr, "Usage: mnemo hdu <save|search|list|transfer|get> [...]")
		os.Exit(1)
	}
}

// ── Code Intelligence Commands ───────────────────────────────────────

func runCodeCommand() {
	subCmd := os.Args[2]

	store, _ := openStore()
	defer store.Close()

	codeStore, err := codegraph.NewStore(store.DB())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing codegraph: %v\n", err)
		os.Exit(1)
	}

	indexer := codegraph.NewIndexer(codeStore)

	switch subCmd {
	case "index":
		rootPath := flagArg("--path", ".")
		if len(os.Args) > 3 && !strings.HasPrefix(os.Args[3], "--") {
			rootPath = os.Args[3]
		}
		fmt.Fprintf(os.Stderr, "Indexing %s ...\n", rootPath)
		result, err := indexer.IndexProject(rootPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Indexed %d files, %d symbols, %d edges\n",
			result.FilesIndexed, result.NodesCreated, result.EdgesCreated)
		if result.FilesErrored > 0 {
			fmt.Printf("%d files failed\n", result.FilesErrored)
		}

	case "search":
		if len(os.Args) < 4 {
			fmt.Fprintln(os.Stderr, "Usage: mnemo code search <name> [--language <lang>] [--limit <n>]")
			os.Exit(1)
		}
		query := os.Args[3]
		lang := flagArg("--language", "")
		limit := flagArgInt("--limit", 10)

		results, err := indexer.FindSymbolByName(query, lang, limit)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Print(codegraph.FormatSearchResults(results, query))

	case "callers":
		if len(os.Args) < 4 {
			fmt.Fprintln(os.Stderr, "Usage: mnemo code callers <symbol>")
			os.Exit(1)
		}
		callers, err := indexer.GetCallersByName(os.Args[3])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Print(codegraph.FormatCallers(callers))

	case "callees":
		if len(os.Args) < 4 {
			fmt.Fprintln(os.Stderr, "Usage: mnemo code callees <symbol>")
			os.Exit(1)
		}
		callees, err := indexer.GetCalleesByName(os.Args[3])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Print(codegraph.FormatCallees(callees))

	case "impact":
		if len(os.Args) < 4 {
			fmt.Fprintln(os.Stderr, "Usage: mnemo code impact <symbol>")
			os.Exit(1)
		}
		impacted, err := indexer.GetImpactByName(os.Args[3])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Print(codegraph.FormatImpact(impacted, os.Args[3]))

	case "status":
		stats, err := indexer.Stats()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Code Intelligence Index\n")
		fmt.Printf("  Nodes: %d\n", stats["nodes"])
		fmt.Printf("  Edges: %d\n", stats["edges"])
		fmt.Printf("  Files: %d\n\n", stats["files"])
		for k, v := range stats {
			if strings.HasPrefix(k, "nodes_") {
				fmt.Printf("  %s: %d symbols\n", strings.TrimPrefix(k, "nodes_"), v)
			}
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown code command: %s\n", subCmd)
		fmt.Fprintln(os.Stderr, "Usage: mnemo code <index|search|callers|callees|impact|status>")
		os.Exit(1)
	}
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
