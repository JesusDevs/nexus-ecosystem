#!/usr/bin/env bash
# ═══════════════════════════════════════════════════════════════════════
# Nexus Context — SessionStart Hook
# Carga contexto al iniciar sesión: mnemo memories + HDUs activos.
# ═══════════════════════════════════════════════════════════════════════
set -euo pipefail

PROJECT=$(basename "$(pwd)")

# ── Check for active HDUs ──────────────────────────────────────────
HDU_LIST=$(ls -d openspec/changes/HDU-* 2>/dev/null | wc -l | tr -d ' ')
ACTIVE_HDU="none"
if [[ -f ".nexus/current_task.yaml" ]]; then
    ACTIVE_HDU=$(grep -E '^hdu_id:' ".nexus/current_task.yaml" 2>/dev/null | awk '{print $2}' || echo "unknown")
fi

echo "{\"systemMessage\": \"Nexus Ecosystem ready. HDUs: $HDU_LIST, Active: $ACTIVE_HDU. Search context: mnemo search '<query>' --project $PROJECT --limit 5. If graphify graph exists, query it: graphify query '<question>'\"}"

exit 0
