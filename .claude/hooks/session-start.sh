#!/usr/bin/env bash
# ═══════════════════════════════════════════════════════════════════════
# Gingx Context — SessionStart Hook
# Carga contexto: knowledge graph, HDUs activos, blockers, codegraph, mnemo.
# AUTO-GENERA el knowledge graph si no existe o codebase cambio.
# ═══════════════════════════════════════════════════════════════════════
set -euo pipefail

PROJECT=$(basename "$(pwd)")
TRACKING_FILE=".gingx/current_task.yaml"
HDU_DIR=".gingx/hdus"
KNOWLEDGE_DIR=".gingx/knowledge"

# ── Check mode ──────────────────────────────────────────────────────
MODE_FILE=".gingx/mode.yaml"
MODE="interactive"
if [[ -f "$MODE_FILE" ]]; then
    MODE=$(grep -E '^mode:' "$MODE_FILE" 2>/dev/null | awk '{print $2}' || echo "interactive")
    if [[ "$MODE" == "off" ]]; then
        echo "{\"systemMessage\": \"Gingx harness OFF. Free mode — no spec required. Use 'gingx-sdd mode set interactive' to re-enable.\"}"
        exit 0
    fi
fi

# ── Knowledge Graph — auto-build if missing ─────────────────────────
# Generates domain-map.yaml from directory structure + codegraph symbols.
# Only runs if the file is missing. Incremental updates happen in stop hook.
KNOWLEDGE_MSG=""
mkdir -p "$KNOWLEDGE_DIR"

if [[ ! -f "$KNOWLEDGE_DIR/domain-map.yaml" ]]; then
    # First run: auto-discover domains from directory structure
    python3 -c "
import yaml, os
from pathlib import Path

domains = []
root = Path('.')
for d in sorted(root.iterdir()):
    if not d.is_dir() or d.name.startswith('.'):
        continue
    py_files = list(d.rglob('*.py'))
    go_files = list(d.rglob('*.go'))
    md_files = list(d.rglob('*.md'))
    total = len(py_files) + len(go_files) + len(md_files)
    if total > 0:
        domains.append({
            'name': d.name,
            'path': str(d) + '/',
            'description': f'{len(py_files)} .py, {len(go_files)} .go, {len(md_files)} .md files',
            'key_symbols': [],
            'dependencies': [],
            'patterns': [],
            'last_indexed': '$(date +%Y-%m-%d\ %H:%M)',
        })

data = {'domains': domains, 'cross_domain_edges': []}
with open('$KNOWLEDGE_DIR/domain-map.yaml', 'w') as f:
    yaml.dump(data, f, default_flow_style=False, allow_unicode=True, sort_keys=False)
" 2>/dev/null || true
fi

# Load domain names for session context
if [[ -f "$KNOWLEDGE_DIR/domain-map.yaml" ]]; then
    KNOWLEDGE_CTX=$(python3 -c "
import yaml
with open('$KNOWLEDGE_DIR/domain-map.yaml') as f:
    data = yaml.safe_load(f) or {}
domains = [d['name'] for d in data.get('domains', [])[:5]]
if domains:
    print(', '.join(domains))
" 2>/dev/null || echo "")
    if [[ -n "$KNOWLEDGE_CTX" ]]; then
        KNOWLEDGE_MSG=" Knowledge: $KNOWLEDGE_CTX."
    fi
fi

# ── Auto-delegation status ─────────────────────────────────────────
AUTO_DELEGATION_MSG=""
if [[ "$MODE" == "automatic" ]]; then
    AUTO_DELEGATION_MSG=" Auto-delegation ACTIVE — tasks auto-classified and dispatched. Use: gingx-sdd auto '<task>'"
fi

# ── Gather context ──────────────────────────────────────────────────
ACTIVE_HDU="none"
BLOCKER_MSG=""
HDU_LIST_COUNT=$(ls -d openspec/changes/HDU-* 2>/dev/null | wc -l | tr -d ' ')

if [[ -f "$TRACKING_FILE" ]]; then
    ACTIVE_HDU=$(grep -E '^hdu_id:' "$TRACKING_FILE" 2>/dev/null | awk '{print $2}' || echo "unknown")

    if [[ -n "$ACTIVE_HDU" && "$ACTIVE_HDU" != "unknown" && -f "$HDU_DIR/$ACTIVE_HDU.yaml" ]]; then
        OPEN=$(python3 -c "
import yaml
try:
    with open('$HDU_DIR/$ACTIVE_HDU.yaml') as f:
        data = yaml.safe_load(f) or {}
    blockers = [b['question'] for b in data.get('blockers', []) if not b.get('answer')]
    if blockers:
        print(blockers[0])
except:
    pass
" 2>/dev/null || echo "")
        if [[ -n "$OPEN" ]]; then
            BLOCKER_MSG=" BLOCKED: $OPEN Responde: gingx-sdd hdu unblock $ACTIVE_HDU --answer \\\"...\\\""
        fi
    fi
fi

# ── CodeGraph status ────────────────────────────────────────────────
CG_MSG=""
if command -v codegraph &>/dev/null && [[ -d ".codegraph" ]]; then
    CG_STATS=$(codegraph status 2>/dev/null | head -3 | tr '\n' ' ' || echo "")
    if [[ -n "$CG_STATS" ]]; then
        CG_MSG=" CodeGraph: $CG_STATS"
    fi
fi

# ── Mnemo sync pull ──────────────────────────────────────────────────
MNEMO_SYNC_MSG=""
if command -v mnemo &>/dev/null; then
    SYNC_OUTPUT=$(mnemo sync pull 2>/dev/null || echo "")
    if [[ -n "$SYNC_OUTPUT" && "$SYNC_OUTPUT" != *"no sync.remote"* ]]; then
        MNEMO_SYNC_MSG=" Mnemo synced from remote."
    fi
    if [[ -f ".gingx/memory/entries.jsonl" ]]; then
        IMPORT_RESULT=$(mnemo import 2>/dev/null || echo "")
        if [[ -n "$IMPORT_RESULT" && "$IMPORT_RESULT" != *"nothing to import"* ]]; then
            MNEMO_IMPORT_MSG=" Mnemo: $IMPORT_RESULT from .gingx/memory/"
        fi
    fi
    MNEMO_HDU_COUNT=$(mnemo hdu list --project "$PROJECT" --limit 100 2>/dev/null | grep -c "^  HDU:" || echo "0")
fi

echo "{\"systemMessage\": \"Gingx Ecosystem ready. HDUs: $HDU_LIST_COUNT (mnemo: $MNEMO_HDU_COUNT), Active: $ACTIVE_HDU.$BLOCKER_MSG$KNOWLEDGE_MSG$MNEMO_SYNC_MSG$MNEMO_IMPORT_MSG$CG_MSG$AUTO_DELEGATION_MSG\n  Explorer: gingx-sdd knowledge {status|search|explore|index}\n  Mnemo: mnemo search '<query>' --project $PROJECT --limit 5\n  Create HDU: gingx-sdd hdu create '<title>' --question '<blocking question>'\n  Auto: gingx-sdd auto '<task>'\n  Mode: gingx-sdd mode status\"}"

exit 0
