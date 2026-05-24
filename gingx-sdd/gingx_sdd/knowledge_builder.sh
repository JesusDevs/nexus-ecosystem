#!/usr/bin/env bash
# ═══════════════════════════════════════════════════════════════════════
# Gingx Knowledge Builder — explorer-agent toolkit
# Builds and updates .gingx/knowledge/ from codegraph + mnemo.
# Portable: all output in .gingx/knowledge/ travels with the repo.
# ═══════════════════════════════════════════════════════════════════════
set -euo pipefail

GINGX_DIR=".gingx"
KNOWLEDGE_DIR="$GINGX_DIR/knowledge"
PROJECT=$(basename "$(pwd)")

cmd="${1:-status}"
shift || true

# ── status ──────────────────────────────────────────────────────────

do_status() {
    echo "=== Knowledge Graph Status ==="
    echo ""

    if [[ -f "$KNOWLEDGE_DIR/domain-map.yaml" ]]; then
        local domains=$(python3 -c "
import yaml
with open('$KNOWLEDGE_DIR/domain-map.yaml') as f:
    data = yaml.safe_load(f) or {}
print(len(data.get('domains', [])))
" 2>/dev/null || echo "0")
        echo "Domains mapped: $domains"
    else
        echo "Domains mapped: 0 (no domain-map.yaml)"
    fi

    if [[ -f "$KNOWLEDGE_DIR/component-index.yaml" ]]; then
        local comps=$(python3 -c "
import yaml
with open('$KNOWLEDGE_DIR/component-index.yaml') as f:
    data = yaml.safe_load(f) or {}
print(len(data.get('components', [])))
" 2>/dev/null || echo "0")
        echo "Components indexed: $comps"
    else
        echo "Components indexed: 0 (no component-index.yaml)"
    fi

    if [[ -f "$KNOWLEDGE_DIR/decisions-log.yaml" ]]; then
        local decs=$(python3 -c "
import yaml
with open('$KNOWLEDGE_DIR/decisions-log.yaml') as f:
    data = yaml.safe_load(f) or {}
print(len(data.get('decisions', [])))
" 2>/dev/null || echo "0")
        echo "Decisions logged: $decs"
    else
        echo "Decisions logged: 0 (no decisions-log.yaml)"
    fi

    # Mnemo stats
    if command -v mnemo &>/dev/null; then
        echo ""
        echo "--- Mnemo ---"
        mnemo stats 2>/dev/null || echo "  (mnemo unavailable)"
    fi

    # Codegraph stats
    if command -v codegraph &>/dev/null && [[ -d ".codegraph" ]]; then
        echo ""
        echo "--- CodeGraph ---"
        codegraph status 2>/dev/null || echo "  (codegraph unavailable)"
    fi
}

# ── index ───────────────────────────────────────────────────────────

do_index() {
    echo "=== Building Knowledge Graph ==="

    mkdir -p "$KNOWLEDGE_DIR"

    # Index code if codegraph available
    if command -v codegraph &>/dev/null; then
        echo "[1/3] Indexing code symbols..."
        codegraph sync 2>/dev/null || true
        mnemo code index . 2>/dev/null || true
    fi

    # Generate domain map from directory structure
    echo "[2/3] Generating domain map..."
    python3 -c "
import yaml, os
from pathlib import Path

domains = []
top_dirs = [d for d in Path('.').iterdir() if d.is_dir() and not d.name.startswith('.')][:20]

for d in sorted(top_dirs):
    py_files = list(d.rglob('*.py'))
    go_files = list(d.rglob('*.go'))
    has_code = len(py_files) + len(go_files) > 0
    if has_code:
        domains.append({
            'name': d.name,
            'path': str(d) + '/',
            'description': f'{len(py_files)} .py files, {len(go_files)} .go files',
            'key_symbols': [],
            'dependencies': [],
            'patterns': [],
        })

output = {
    'domains': domains,
    'cross_domain_edges': [],
}
with open('$KNOWLEDGE_DIR/domain-map.yaml', 'w') as f:
    yaml.dump(output, f, default_flow_style=False, allow_unicode=True, sort_keys=False)
print(f'  -> {len(domains)} domains mapped')
"

    # Save to mnemo
    echo "[3/3] Saving to mnemo..."
    if command -v mnemo &>/dev/null; then
        mnemo save "Knowledge Graph: $PROJECT" \
            "Project knowledge graph indexed. $(cat $KNOWLEDGE_DIR/domain-map.yaml 2>/dev/null | head -20)" \
            --type progress --outcome in_progress --project "$PROJECT" \
            --tags knowledge-graph,explorer,index \
            2>/dev/null || true
    fi

    echo "Done. Use 'gingx-sdd knowledge status' to see stats."
}

# ── search ──────────────────────────────────────────────────────────

do_search() {
    local query="${1:-}"
    if [[ -z "$query" ]]; then
        echo "Usage: gingx-sdd knowledge search <query>"
        exit 1
    fi

    echo "=== Searching: $query ==="

    # 1. Mnemo semantic search
    if command -v mnemo &>/dev/null; then
        echo "--- Mnemo (semantic) ---"
        mnemo search "$query" --project "$PROJECT" --limit 5 2>/dev/null || echo "  (no results)"
    fi

    # 2. Codegraph symbol search
    if command -v codegraph &>/dev/null; then
        echo "--- CodeGraph (symbols) ---"
        codegraph search "$query" 2>/dev/null || echo "  (no results)"
        echo "--- CodeGraph (impact) ---"
        codegraph impact "$query" 2>/dev/null || echo "  (no results)"
    fi

    # 3. Local knowledge grep
    echo "--- Knowledge files ---"
    grep -rl "$query" "$KNOWLEDGE_DIR/" 2>/dev/null || echo "  (no matches in knowledge files)"
}

# ── explore ─────────────────────────────────────────────────────────

do_explore() {
    local target="${1:-}"
    if [[ -z "$target" ]]; then
        echo "Usage: gingx-sdd knowledge explore <symbol|path|domain>"
        exit 1
    fi

    echo "=== Exploring: $target ==="

    # What is it?
    echo "--- Symbol info ---"
    if command -v codegraph &>/dev/null; then
        codegraph search "$target" 2>/dev/null || true
        echo ""
        echo "--- Callers ---"
        codegraph callers "$target" 2>/dev/null || echo "  (none)"
        echo ""
        echo "--- Callees ---"
        codegraph callees "$target" 2>/dev/null || echo "  (none)"
        echo ""
        echo "--- Impact ---"
        codegraph impact "$target" 2>/dev/null || echo "  (none)"
    fi

    # Semantic neighbors
    echo ""
    echo "--- Semantic neighbors (mnemo) ---"
    if command -v mnemo &>/dev/null; then
        mnemo similar "$target" --project "$PROJECT" --limit 5 2>/dev/null || echo "  (none)"
    fi
}

# ── save-decision ───────────────────────────────────────────────────

do_save_decision() {
    local title="${1:-}"
    local rationale="${2:-}"
    local trade_off="${3:-}"

    if [[ -z "$title" ]]; then
        echo "Usage: gingx-sdd knowledge save-decision <title> <rationale> [trade-off]"
        exit 1
    fi

    mkdir -p "$KNOWLEDGE_DIR"

    python3 -c "
import yaml
from datetime import datetime
from pathlib import Path

log_file = Path('$KNOWLEDGE_DIR/decisions-log.yaml')

if log_file.exists():
    data = yaml.safe_load(log_file.read_text()) or {'decisions': []}
else:
    data = {'decisions': []}

dec_id = f'ADR-{len(data[\"decisions\"]) + 1:03d}'
data['decisions'].append({
    'id': dec_id,
    'title': '$title',
    'rationale': '$rationale',
    'trade_off': '$trade_off',
    'date': datetime.now().strftime('%Y-%m-%d %H:%M'),
    'discovered_by': 'explorer-agent',
})

log_file.write_text(yaml.dump(data, default_flow_style=False, allow_unicode=True, sort_keys=False))
print(f'Decision {dec_id} saved: $title')
"

    # Also save to mnemo
    if command -v mnemo &>/dev/null; then
        mnemo save "ADR: $title" \
            "Rationale: $rationale. Trade-off: ${trade_off:-N/A}" \
            --type decision --outcome success --project "$PROJECT" \
            --tags architecture,decision,adr \
            2>/dev/null || true
    fi
}

# ── load ────────────────────────────────────────────────────────────

do_load() {
    # Output knowledge context for session start (compact JSON for hooks)
    python3 -c "
import yaml, json
from pathlib import Path

kd = Path('$KNOWLEDGE_DIR')
output = {
    'domains': 0,
    'components': 0,
    'decisions': 0,
    'top_domains': [],
    'recent_decisions': [],
}

dm = kd / 'domain-map.yaml'
if dm.exists():
    data = yaml.safe_load(dm.read_text()) or {}
    domains = data.get('domains', [])
    output['domains'] = len(domains)
    output['top_domains'] = [d.get('name', '?') for d in domains[:6]]

ci = kd / 'component-index.yaml'
if ci.exists():
    data = yaml.safe_load(ci.read_text()) or {}
    comps = data.get('components', [])
    output['components'] = len(comps)

dl = kd / 'decisions-log.yaml'
if dl.exists():
    data = yaml.safe_load(dl.read_text()) or {}
    decs = data.get('decisions', [])
    output['decisions'] = len(decs)
    output['recent_decisions'] = [d.get('title', '?') for d in decs[-3:]]

print(json.dumps(output))
"
}

# ── dispatch ────────────────────────────────────────────────────────

case "$cmd" in
    status)       do_status ;;
    index)        do_index ;;
    search)       do_search "$@" ;;
    explore)      do_explore "$@" ;;
    save-decision) do_save_decision "$@" ;;
    load)         do_load ;;
    *)
        echo "Usage: gingx-sdd knowledge {status|index|search|explore|save-decision|load}"
        echo ""
        echo "  status          Show knowledge graph statistics"
        echo "  index           Rebuild knowledge graph from codebase"
        echo "  search <q>      Search across mnemo + codegraph + knowledge files"
        echo "  explore <sym>   Deep-dive on a symbol (callers, callees, impact, semantic neighbors)"
        echo "  save-decision   Log an architecture decision to decisions-log.yaml"
        echo "  load            Output compact JSON for session-start hook"
        exit 1
        ;;
esac
