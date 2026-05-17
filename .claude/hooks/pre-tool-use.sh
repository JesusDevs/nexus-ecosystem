#!/usr/bin/env bash
# ═══════════════════════════════════════════════════════════════════════
# Nexus SDD Gate — PreToolUse Hook
# Blocks code writes without an approved spec.
# "Sin spec no hay código. Sin BDD no hay spec."
# ═══════════════════════════════════════════════════════════════════════
set -euo pipefail

INPUT=$(cat 2>/dev/null || echo "{}")
TOOL_NAME=$(echo "$INPUT" | python3 -c "import sys,json; print(json.load(sys.stdin).get('tool_name',''))" 2>/dev/null || echo "")

# ── Only enforce Write/Edit operations ──────────────────────────────
if [[ "$TOOL_NAME" != "Write" && "$TOOL_NAME" != "Edit" ]]; then
    exit 0
fi

# ── Extract file path ───────────────────────────────────────────────
FILE_PATH=$(echo "$INPUT" | python3 -c "
import sys, json
d = json.load(sys.stdin)
ti = d.get('tool_input', {})
# Handle both string and dict tool_input
if isinstance(ti, str):
    import re
    m = re.search(r'file_path[\"']?\s*[:=]\s*[\"']([^\"']+)', ti)
    print(m.group(1) if m else '')
else:
    print(ti.get('file_path', ''))
" 2>/dev/null || echo "")

# ── Allow-list: always permit these ─────────────────────────────────
# Read-only operations (handled above), docs, configs, tests
if [[ -z "$FILE_PATH" ]]; then
    exit 0
fi

# Allow .md, .txt files (documentation)
if [[ "$FILE_PATH" =~ \.(md|txt)$ ]]; then
    exit 0
fi

# Allow .yaml/.json config files (NOT in source directories)
if [[ "$FILE_PATH" =~ \.(yaml|yml|json|toml|cfg)$ ]]; then
    if [[ ! "$FILE_PATH" =~ /(src|lib|app|internal|pkg|cmd)/ ]]; then
        exit 0  # Config file outside source dirs — allow
    fi
fi

# Allow test files
if [[ "$FILE_PATH" =~ _test\.|test_|/tests/|/test/|\.test\.|_spec\.|spec_ ]]; then
    exit 0
fi

# Allow .nexus/ and .claude/ internal files
if [[ "$FILE_PATH" =~ ^\.(nexus|claude)/ ]]; then
    exit 0
fi

# ── Source code file detected ───────────────────────────────────────
# Check for active HDU tracking file
TRACKING_FILE=".nexus/current_task.yaml"

if [[ -f "$TRACKING_FILE" ]]; then
    HDU=$(grep -E '^hdu_id:' "$TRACKING_FILE" 2>/dev/null | awk '{print $2}' || echo "active")
    # Spec exists — allow
    exit 0
fi

# Check if there are ANY HDUs in openspec
HDU_COUNT=$(ls -d openspec/changes/HDU-* 2>/dev/null | wc -l | tr -d ' ')
if [[ "$HDU_COUNT" -gt 0 ]]; then
    # There are HDUs but none active — warn but allow
    echo '{"systemMessage": "SDD: Hay HDUs en openspec/changes/ pero ninguno activo. Considera activar uno con: echo hdu_id: HDU-XX > .nexus/current_task.yaml o ejecuta nexus-sdd orchestrate HDU-XX --phase apply"}'
    exit 0
fi

# ── BLOCK: No spec found ────────────────────────────────────────────
echo '{"systemMessage": "SDD GATE BLOCKED: No hay spec aprobado. El harness Nexus requiere spec antes de escribir codigo. Crea uno: nexus-sdd spec \"<titulo>\" o invoca /supervisor para iniciar el workflow SDD. Si es un fix urgente: echo fix: descripcion > .nexus/current_task.yaml"}'
exit 1
