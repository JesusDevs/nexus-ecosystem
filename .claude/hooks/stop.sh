#!/usr/bin/env bash
# ═══════════════════════════════════════════════════════════════════════
# Nexus Memory — Stop Hook
# Auto-guarda progreso en mnemo al terminar sesión.
# También actualiza el knowledge graph si graphify está instalado.
# ═══════════════════════════════════════════════════════════════════════
set -euo pipefail

PROJECT=$(basename "$(pwd)")
TRACKING_FILE=".nexus/current_task.yaml"

# ── Save progress to mnemo if there's an active task ───────────────
if [[ -f "$TRACKING_FILE" ]]; then
    HDU=$(grep -E '^hdu_id:' "$TRACKING_FILE" 2>/dev/null | awk '{print $2}' || echo "unknown")
    echo "{\"systemMessage\": \"Session ended. Save progress: mnemo save 'Session checkpoint: $HDU' 'Progress saved automatically by Nexus Stop hook. HDU: $HDU' --type progress --outcome in_progress --project $PROJECT\"}"
fi

# ── Update graphify graph if installed ──────────────────────────────
if command -v graphify &>/dev/null && [[ -d "graphify-out" ]]; then
    graphify update . --no-cluster 2>/dev/null || true
fi

exit 0
