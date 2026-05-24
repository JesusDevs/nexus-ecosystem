#!/usr/bin/env bash
# ═══════════════════════════════════════════════════════════════════════
# Gingx Context — SessionStart Hook
# Carga contexto: HDUs activos, blockers, codegraph stats, mnemo.
# ═══════════════════════════════════════════════════════════════════════
set -euo pipefail

PROJECT=$(basename "$(pwd)")
TRACKING_FILE=".gingx/current_task.yaml"
HDU_DIR=".gingx/hdus"

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

    # Check for open blockers
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
    # Auto-import portable memory from .gingx/memory/
    if [[ -f ".gingx/memory/entries.jsonl" ]]; then
        IMPORT_RESULT=$(mnemo import 2>/dev/null || echo "")
        if [[ -n "$IMPORT_RESULT" && "$IMPORT_RESULT" != *"nothing to import"* ]]; then
            MNEMO_IMPORT_MSG=" Mnemo: $IMPORT_RESULT from .gingx/memory/"
        fi
    fi
    # Count HDUs in mnemo
    MNEMO_HDU_COUNT=$(mnemo hdu list --project "$PROJECT" --limit 100 2>/dev/null | grep -c "^  HDU:" || echo "0")
fi

echo "{\"systemMessage\": \"Gingx Ecosystem ready. HDUs: $HDU_LIST_COUNT (mnemo: $MNEMO_HDU_COUNT), Active: $ACTIVE_HDU.$BLOCKER_MSG$MNEMO_SYNC_MSG$MNEMO_IMPORT_MSG$CG_MSG$AUTO_DELEGATION_MSG\n  Mnemo: mnemo search '<query>' --project $PROJECT --limit 5\n  Create HDU: gingx-sdd hdu create '<title>' --question '<blocking question>'\n  Auto: gingx-sdd auto '<task>'\n  Mode: gingx-sdd mode status\"}"

exit 0
