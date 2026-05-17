# Changelog — Nexus-Mnemo

## v0.2.0 — Memoria Versionable (2026-05-13)

### Nuevas Features

- **`mem_save` MCP tool**: Los agentes ahora pueden guardar memorias autónomamente sin ejecutar comandos shell. Soporta `title`, `content`, `type`, `project`, `outcome`, `tags`, `media_type`, `version`.
- **Versionado de memorias**: Nuevo comando `mnemo release <project> <version>` que crea un snapshot de todas las memorias no versionadas. Ideal para capturar el conocimiento generado en cada ciclo de desarrollo.
- **Diff entre releases**: `mnemo diff <project> <v1> <v2>` muestra qué conocimiento se agregó, actualizó o removió entre versiones.
- **Listado de releases**: `mnemo releases <project>` lista todos los snapshots con metadata.
- **Export portable**: `mnemo pack export <project> [--version]` exporta memorias a JSON portable para compartir entre equipos.
- **Soporte multimodal**: Nuevo campo `media_type` (text, image, pdf, audio, video) y `media_file` (blob) en todas las memorias.
- **Nuevas MCP tools**: `mem_release`, `mem_diff`, `mem_list_releases` expuestas vía MCP para que cualquier agente las use.

### Mejoras

- El MCP server ahora se identifica como `nexus-mnemo v0.2.0`
- Los resultados de búsqueda incluyen versión y tipo de medio
- `mem_save` acepta `version` como parámetro opcional
- `mem_search_semantic` acepta `version` como filtro opcional
- La CLI usa el nombre `mnemo` en todos los mensajes

### Cambios Internos

- Nueva tabla `vec_releases` en SQLite para tracking de versiones
- Nuevo índice `idx_vec_memories_version`
- Nuevo struct `ReleaseSnapshot`, `VersionDiff`, `MemorySummary`, `DiffPair`, `MemoryPackEntry`
- Métodos `Release()`, `Diff()`, `ListReleases()`, `ExportPack()` en el Store
- Columna `version` agregada a `vec_memories`

---

## v0.1.0 — Initial Release (2026-05-02)

- Búsqueda semántica vía Ollama embeddings (bge-large-en-v1.5, 1024-dim)
- Cosine similarity en SQLite sin índices externos
- 4 tools MCP: `mem_search_semantic`, `mem_similar`, `mem_transfer`, `mem_detect_conflicts_semantic`
- CLI: search, save, similar, transfer, conflicts, stats, setup
- Zero APIs externas, todo local
- Standalone vector memory system with its own DB
