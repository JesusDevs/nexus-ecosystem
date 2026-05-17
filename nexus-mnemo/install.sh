#!/usr/bin/env bash
set -euo pipefail

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

log()  { echo -e "${CYAN}[mnemo]${NC} $1"; }
ok()   { echo -e "${GREEN}[✓]${NC} $1"; }
warn() { echo -e "${YELLOW}[!]${NC} $1"; }

echo -e "\n${BOLD}${CYAN}╔══════════════════════════════════════════╗"
echo "║  🧠  NEXUS-MNEMO — Memoria Versionable  ║"
echo "║       Zero-Friction Installer            ║"
echo -e "╚══════════════════════════════════════════╝${NC}\n"

# ── Verificar Go ─────────────────────────────────────────────────────
log "Verificando Go..."
if command -v go &>/dev/null; then
    ok "Go $(go version | awk '{print $3}')"
else
    warn "Go no encontrado. Instalando..."
    case "$(uname -s)" in
        Darwin) brew install go ;;
        Linux)
            curl -fsSL https://go.dev/dl/go1.22.0.linux-amd64.tar.gz | sudo tar -C /usr/local -xz
            export PATH=$PATH:/usr/local/go/bin
            ;;
    esac
    ok "Go instalado"
fi

# ── Compilar ──────────────────────────────────────────────────────────
log "Compilando nexus-mnemo..."
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

go build -o mnemo .
ok "Binario compilado: $(pwd)/mnemo"

# ── Instalar en PATH ──────────────────────────────────────────────────
log "Instalando en /usr/local/bin..."
if [[ -w /usr/local/bin ]]; then
    cp mnemo /usr/local/bin/
else
    sudo cp mnemo /usr/local/bin/
fi
ok "mnemo instalado en /usr/local/bin/"

# ── Verificar Ollama ─────────────────────────────────────────────────
log "Verificando Ollama..."
if curl -s http://localhost:11434/api/tags >/dev/null 2>&1; then
    ok "Ollama corriendo en localhost:11434"
else
    warn "Ollama no detectado. Instalando..."
    if command -v ollama &>/dev/null; then
        warn "Ollama CLI existe pero el servicio no está corriendo."
        warn "Ejecuta: ollama serve"
    else
        curl -fsSL https://ollama.com/install.sh | sh
        ok "Ollama instalado"
    fi
fi

# ── Configurar vía DB (no env vars) ──────────────────────────────────
log "Configurando mnemo (modelo, host)..."
MODEL="${OLLAMA_EMBED_MODEL:-bge-m3}"
HOST="${OLLAMA_HOST:-http://localhost:11434}"

# Run setup which saves config to ~/.mnemo/mnemo.db
EMBEDDER_MOCK=true ./mnemo config set embed.model "$MODEL" 2>/dev/null || true
EMBEDDER_MOCK=true ./mnemo config set ollama.host "$HOST" 2>/dev/null || true

ok "Config guardada en ~/.mnemo/mnemo.db"
ok "Modelo: $MODEL | Host: $HOST"

# ── Descargar modelo de embeddings ───────────────────────────────────
log "Verificando modelo de embeddings ($MODEL)..."
if ollama list 2>/dev/null | grep -q "$MODEL"; then
    ok "Modelo $MODEL ya descargado"
else
    warn "Descargando modelo $MODEL (~1.2GB, una sola vez)..."
    ollama pull "$MODEL"
    ok "Modelo $MODEL descargado"
fi

# ── Verificar que funcione ───────────────────────────────────────────
log "Probando mnemo..."
if ./mnemo version &>/dev/null; then
    ok "mnemo $(./mnemo version) funciona correctamente"
else
    warn "No se pudo ejecutar mnemo. Verifica la instalación."
fi

# ── Mostrar config actual ────────────────────────────────────────────
echo ""
log "Configuración actual:"
EMBEDDER_MOCK=true ./mnemo config 2>/dev/null || true

# ── Configurar MCP ───────────────────────────────────────────────────
echo ""
echo -e "${BOLD}${GREEN}✅ nexus-mnemo instalado!${NC}\n"
echo -e "  ${BOLD}Comandos:${NC}"
echo -e "  ${CYAN}mnemo mcp${NC}             — Iniciar servidor MCP (stdio)"
echo -e "  ${CYAN}mnemo search${NC}          — Búsqueda semántica"
echo -e "  ${CYAN}mnemo save${NC}            — Guardar memoria con embedding"
echo -e "  ${CYAN}mnemo similar${NC}         — Memorias similares"
echo -e "  ${CYAN}mnemo transfer${NC}        — Transferir conocimiento"
echo -e "  ${CYAN}mnemo release${NC}         — Crear snapshot de versión"
echo -e "  ${CYAN}mnemo diff${NC}            — Comparar releases"
echo -e "  ${CYAN}mnemo pack export${NC}     — Exportar pack portable"
echo -e "  ${CYAN}mnemo conflicts${NC}       — Detectar conflictos"
echo -e "  ${CYAN}mnemo stats${NC}           — Estadísticas"
echo -e "  ${CYAN}mnemo config${NC}          — Ver configuración (DB)"
echo -e "  ${CYAN}mnemo config set k v${NC}  — Cambiar configuración"
echo -e "  ${CYAN}mnemo setup${NC}           — Verificar + configurar\n"

echo -e "  ${BOLD}Configuración:${NC}"
echo -e "  La config se guarda en ~/.mnemo/mnemo.db (tabla vec_config)."
echo -e "  Variables de entorno (OLLAMA_HOST, OLLAMA_EMBED_MODEL, EMBEDDER_MOCK)"
echo -e "  solo actúan como overrides. La fuente de verdad es la DB.\n"

echo -e "  ${BOLD}Agregar a Claude Code:${NC}"
echo -e "  ${GREEN}claude mcp add mnemo -- mnemo mcp${NC}\n"

echo -e "  ${BOLD}Agregar a OpenCode:${NC}"
echo -e "  ${GREEN}opencode mcp add mnemo -- mnemo mcp${NC}\n"

echo -e "  ${BOLD}Agregar a Kiro / Antigravity / Codex:${NC}"
echo -e "  ${GREEN}<tool> mcp add mnemo -- mnemo mcp${NC}\n"
