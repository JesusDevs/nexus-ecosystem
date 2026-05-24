#!/usr/bin/env bash
# ═══════════════════════════════════════════════════════════════════════
# Gingx SDD Gate — PreToolUse Hook
# Blocks code writes without an approved spec.
# V2: HDU blocker support + mode awareness.
# "Sin spec no hay código. Sin respuesta no avanza."
# ═══════════════════════════════════════════════════════════════════════
set -euo pipefail

INPUT=$(cat 2>/dev/null || echo "{}")
TOOL_NAME=$(echo "$INPUT" | python3 -c "import sys,json; print(json.load(sys.stdin).get('tool_name',''))" 2>/dev/null || echo "")

# ── Only enforce Write/Edit operations ──────────────────────────────
if [[ "$TOOL_NAME" != "Write" && "$TOOL_NAME" != "Edit" ]]; then
    exit 0
fi

# ── Check mode — if off, pass everything ────────────────────────────
MODE_FILE=".gingx/mode.yaml"
if [[ -f "$MODE_FILE" ]]; then
    MODE=$(grep -E '^mode:' "$MODE_FILE" 2>/dev/null | awk '{print $2}' || echo "interactive")
    if [[ "$MODE" == "off" || "$MODE" == "dry_run" ]]; then
        exit 0
    fi
fi

# ── Extract file path ───────────────────────────────────────────────
FILE_PATH=$(echo "$INPUT" | python3 -c "
import sys, json
d = json.load(sys.stdin)
ti = d.get('tool_input', {})
if isinstance(ti, str):
    import re
    m = re.search(r'file_path[\"']?\s*[:=]\s*[\"']([^\"']+)', ti)
    print(m.group(1) if m else '')
else:
    print(ti.get('file_path', ''))
" 2>/dev/null || echo "")

# ── Allow-list: always permit these ─────────────────────────────────
if [[ -z "$FILE_PATH" ]]; then
    exit 0
fi

if [[ "$FILE_PATH" =~ \.(md|txt)$ ]]; then
    exit 0
fi

if [[ "$FILE_PATH" =~ \.(yaml|yml|json|toml|cfg)$ ]]; then
    if [[ ! "$FILE_PATH" =~ /(src|lib|app|internal|pkg|cmd)/ ]]; then
        exit 0
    fi
fi

if [[ "$FILE_PATH" =~ _test\.|test_|/tests/|/test/|\.test\.|_spec\.|spec_ ]]; then
    exit 0
fi

if [[ "$FILE_PATH" =~ ^\.(gingx|claude)/ ]]; then
    exit 0
fi

# ── Source code file detected ───────────────────────────────────────
TRACKING_FILE=".gingx/current_task.yaml"
HDU_DIR=".gingx/hdus"

# Check for active HDU
if [[ -f "$TRACKING_FILE" ]]; then
    HDU=$(grep -E '^hdu_id:' "$TRACKING_FILE" 2>/dev/null | awk '{print $2}' || echo "")

    if [[ -n "$HDU" && -f "$HDU_DIR/$HDU.yaml" ]]; then
        # Check for open blockers
        OPEN_BLOCKER=$(python3 -c "
import yaml, sys
try:
    with open('$HDU_DIR/$HDU.yaml') as f:
        data = yaml.safe_load(f) or {}
    for b in data.get('blockers', []):
        if not b.get('answer'):
            print(b.get('question', 'Unknown question'))
            sys.exit(0)
    sys.exit(1)
except:
    sys.exit(1)
" 2>/dev/null || echo "")

        if [[ -n "$OPEN_BLOCKER" ]]; then
            echo "{\"systemMessage\": \"SDD GATE BLOCKED: El HDU $HDU tiene una pregunta bloqueante sin responder: '$OPEN_BLOCKER'. Respóndela con: gingx-sdd hdu unblock $HDU --answer \\\"<tu respuesta>\\\"\"}"
            exit 1
        fi

        # Has active HDU with no open blockers — allow
        exit 0
    fi
fi

# Check if there are ANY HDUs in openspec
HDU_COUNT=$(ls -d openspec/changes/HDU-* 2>/dev/null | wc -l | tr -d ' ')
if [[ "$HDU_COUNT" -gt 0 ]]; then
    if [[ "$MODE" == "automatic" ]]; then
        echo '{"systemMessage": "SDD: Hay HDUs pero ninguno activo. Auto-delegacion sugiere: gingx-sdd auto \"<tu tarea>\" para clasificar y dispatchear automaticamente.\"}'
    else
        echo '{"systemMessage": "SDD: Hay HDUs en openspec/changes/ pero ninguno activo. Activa uno: gingx-sdd hdu status <HDU-ID> o crea uno nuevo: gingx-sdd hdu create \"titulo\"\""}'
    fi
    exit 0
fi

# ── BLOCK: No spec found ────────────────────────────────────────────
if [[ "$MODE" == "automatic" ]]; then
    echo '{"systemMessage": "SDD: Sin spec previo. Auto-delegacion activa — ejecuta: gingx-sdd auto \"<tu tarea>\" para clasificar y disparar el agente correcto.\"}'
else
    echo '{"systemMessage": "SDD GATE BLOCKED: No hay spec aprobado. Crea uno: gingx-sdd hdu create \"<titulo>\" --question \"<pregunta bloqueante>\" o usa gingx-sdd mode set off para desactivar el harness.\"}'
fi
exit 1
