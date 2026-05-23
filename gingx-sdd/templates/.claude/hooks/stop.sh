#!/usr/bin/env bash
# ═══════════════════════════════════════════════════════════════════════
# Gingx Memory — Stop Hook
# Auto-guarda progreso en mnemo + actualiza HDU tracking.
# También codegraph sync y knowledge graph export.
# ═══════════════════════════════════════════════════════════════════════
set -euo pipefail

PROJECT=$(basename "$(pwd)")
TRACKING_FILE=".gingx/current_task.yaml"
HDU_DIR=".gingx/hdus"

# ── Check mode ──────────────────────────────────────────────────────
MODE_FILE=".gingx/mode.yaml"
if [[ -f "$MODE_FILE" ]]; then
    MODE=$(grep -E '^mode:' "$MODE_FILE" 2>/dev/null | awk '{print $2}' || echo "interactive")
    if [[ "$MODE" == "off" ]]; then
        echo '{"systemMessage": "Gingx harness off. Session ended."}'
        exit 0
    fi
fi

# ── Update HDU progress if there's an active task ───────────────────
if [[ -f "$TRACKING_FILE" ]]; then
    HDU=$(grep -E '^hdu_id:' "$TRACKING_FILE" 2>/dev/null | awk '{print $2}' || echo "")
    if [[ -n "$HDU" ]]; then
        # Touch HDU entry with updated timestamp
        if [[ -f "$HDU_DIR/$HDU.yaml" ]]; then
            python3 -c "
import yaml
from datetime import datetime
try:
    with open('$HDU_DIR/$HDU.yaml') as f:
        data = yaml.safe_load(f) or {}
    data['updated_at'] = datetime.now().strftime('%Y-%m-%d %H:%M')
    with open('$HDU_DIR/$HDU.yaml', 'w') as f:
        yaml.dump(data, f, default_flow_style=False, allow_unicode=True, sort_keys=False)
except Exception:
    pass
" 2>/dev/null || true
        fi

        # Save progress to mnemo
        if command -v mnemo &>/dev/null; then
            mnemo save "Session checkpoint: $HDU" \
                "Progress saved by Nexus Stop hook. HDU: $HDU" \
                --type progress --outcome in_progress --project "$PROJECT" \
                2>/dev/null || true
        fi
    fi
fi

# ── CodeGraph sync ──────────────────────────────────────────────────
if command -v codegraph &>/dev/null && [[ -d ".codegraph" ]]; then
    codegraph sync 2>/dev/null || true
fi

# ── Update graphify graph if installed ──────────────────────────────
if command -v graphify &>/dev/null && [[ -d "graphify-out" ]]; then
    graphify update . --no-cluster 2>/dev/null || true
fi

echo "{\"systemMessage\": \"Session ended. HDU progress saved. Run gingx-sdd hdu summary to see all.\"}"
exit 0
