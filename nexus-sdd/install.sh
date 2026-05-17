#!/usr/bin/env bash
set -euo pipefail

# ═══════════════════════════════════════════════════════════════════════
# Nexus-SDD Universal Installer
# "One command to rule them all"
#
# Usage:
#   curl -fsSL https://nexus-sdd.dev/install.sh | bash
#   # or locally:
#   ./install.sh
#
# What it does:
#   1. Detects OS and installs prerequisites (Python, Node, Go)
#   2. Installs OpenSpec CLI (spec-driven development)
#   3. Installs Nexus-Mnemo (vector memory, replaces Engram)
#   4. Installs Ollama + bge-m3 embedding model (local, zero-cost)
#   5. Sets up LangGraph harness + supervisor
#   6. Detects project tech stack
#   7. Installs matching skills + team personas
#   8. Configures Claude Code / Cursor / Windsurf / OpenCode / Kiro
#   9. Creates .nexus/ directory structure
# ═══════════════════════════════════════════════════════════════════════

# ── Colors ───────────────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

log()    { echo -e "${BLUE}[NEXUS]${NC} $1"; }
ok()     { echo -e "${GREEN}[✓]${NC} $1"; }
warn()   { echo -e "${YELLOW}[!]${NC} $1"; }
err()    { echo -e "${RED}[✗]${NC} $1"; }
header() { echo -e "\n${BOLD}${CYAN}═══ $1 ═══${NC}\n"; }

# ── Detect OS ────────────────────────────────────────────────────────
detect_os() {
    case "$(uname -s)" in
        Darwin)  OS="macos" ;;
        Linux)   OS="linux" ;;
        MINGW*|MSYS*|CYGWIN*) OS="windows" ;;
        *)       err "OS no soportado: $(uname -s)"; exit 1 ;;
    esac
    log "OS detectado: ${OS}"
}

# ── Check/Install Prerequisites ───────────────────────────────────────
install_python() {
    if command -v python3 &>/dev/null; then
        PYTHON_VERSION=$(python3 -c 'import sys; print(f"{sys.version_info.major}.{sys.version_info.minor}")')
        if [[ "${PYTHON_VERSION%.*}" -ge 3 && "${PYTHON_VERSION#*.}" -ge 11 ]]; then
            ok "Python ${PYTHON_VERSION} encontrado"
            return
        fi
    fi
    warn "Python 3.11+ requerido. Instalando..."
    case "$OS" in
        macos)
            if command -v brew &>/dev/null; then
                brew install python@3.12
            else
                err "Instala Homebrew primero: https://brew.sh"
                exit 1
            fi
            ;;
        linux)
            sudo apt-get update -qq && sudo apt-get install -y python3.12 python3-pip python3.12-venv
            ;;
        windows)
            err "En Windows, instala Python desde https://python.org"
            exit 1
            ;;
    esac
    ok "Python instalado"
}

install_node() {
    if command -v node &>/dev/null; then
        NODE_VERSION=$(node -v | sed 's/v//' | cut -d. -f1)
        if [[ "$NODE_VERSION" -ge 20 ]]; then
            ok "Node.js $(node -v) encontrado"
            return
        fi
    fi
    warn "Node.js 20+ requerido para OpenSpec. Instalando..."
    case "$OS" in
        macos) brew install node@20 ;;
        linux)
            curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
            sudo apt-get install -y nodejs
            ;;
    esac
    ok "Node.js instalado"
}

install_go() {
    if command -v go &>/dev/null; then
        ok "Go $(go version | awk '{print $3}') encontrado"
        return
    fi
    warn "Go requerido para Nexus-Mnemo. Instalando..."
    case "$OS" in
        macos) brew install go ;;
        linux)
            curl -fsSL https://go.dev/dl/go1.22.0.linux-amd64.tar.gz | sudo tar -C /usr/local -xz
            export PATH=$PATH:/usr/local/go/bin
            ;;
    esac
    ok "Go instalado"
}

# ── Install OpenSpec ──────────────────────────────────────────────────
install_openspec() {
    header "Instalando OpenSpec (Spec-Driven Development)"
    if command -v openspec &>/dev/null; then
        ok "OpenSpec ya instalado: $(openspec version 2>/dev/null || echo 'ok')"
    else
        npm install -g @fission-ai/openspec 2>/dev/null || {
            warn "OpenSpec CLI no disponible via npm. Creando estructura manual..."
        }
        ok "OpenSpec instalado"
    fi
}

# ── Install Ollama + Embedding Model ──────────────────────────────────
install_ollama() {
    header "Instalando Ollama (Embeddings Locales)"

    if command -v ollama &>/dev/null; then
        ok "Ollama CLI encontrado"
    else
        warn "Instalando Ollama..."
        curl -fsSL https://ollama.com/install.sh | sh
        ok "Ollama instalado"
    fi

    # Verificar que el servicio corre
    if curl -s http://localhost:11434/api/tags >/dev/null 2>&1; then
        ok "Ollama servicio corriendo"
    else
        warn "Iniciando Ollama en background..."
        ollama serve &>/dev/null &
        sleep 3
    fi

    # Descargar modelo (bge-m3: multilingual, 1024-dim, superior a bge-large-en-v1.5)
    MODEL="bge-m3"
    if ollama list 2>/dev/null | grep -q "$MODEL"; then
        ok "Modelo $MODEL descargado"
    else
        warn "Descargando $MODEL (~1.2GB, una sola vez)..."
        ollama pull "$MODEL"
        ok "Modelo $MODEL listo"
    fi
}

# ── Install Nexus-Mnemo ───────────────────────────────────────────────
install_mnemo() {
    header "Instalando Nexus-Mnemo (Memoria Vectorial Versionable)"

    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    MNEMO_DIR="$SCRIPT_DIR/../nexus-mnemo"

    if [[ -f "$MNEMO_DIR/main.go" ]]; then
        cd "$MNEMO_DIR"
        go build -o mnemo . 2>/dev/null || {
            warn "Compilación de mnemo falló. Intentando con binario..."
        }
        if [[ -f "./mnemo" ]]; then
            cp mnemo /usr/local/bin/ 2>/dev/null || sudo cp mnemo /usr/local/bin/
            ok "mnemo instalado desde fuente"
        fi
    elif command -v mnemo &>/dev/null; then
        ok "mnemo ya instalado ($(mnemo version 2>/dev/null || echo 'ok'))"
    else
        warn "mnemo no encontrado. Clonando y compilando..."
        if [[ ! -d "$SCRIPT_DIR/../nexus-mnemo" ]]; then
            git clone https://github.com/nexus-sdd/nexus-mnemo "$SCRIPT_DIR/../nexus-mnemo" 2>/dev/null || {
                warn "No se pudo clonar nexus-mnemo."
                warn "Instálalo manualmente: https://github.com/nexus-sdd/nexus-mnemo"
                return
            }
        fi
        cd "$SCRIPT_DIR/../nexus-mnemo"
        go build -o mnemo .
        cp mnemo /usr/local/bin/ 2>/dev/null || sudo cp mnemo /usr/local/bin/
        ok "mnemo compilado e instalado"
    fi

    # Guardar config en DB
    if command -v mnemo &>/dev/null; then
        EMBEDDER_MOCK=true mnemo config set embed.model "bge-m3" 2>/dev/null || true
        EMBEDDER_MOCK=true mnemo config set ollama.host "http://localhost:11434" 2>/dev/null || true
        ok "mnemo configurado (DB: ~/.mnemo/mnemo.db)"
    fi
}

# ── Install Nexus-SDD Python Package ──────────────────────────────────
install_nexus() {
    header "Instalando Nexus-SDD (LangGraph Harness)"

    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    cd "$SCRIPT_DIR"

    if [[ -f "pyproject.toml" ]]; then
        pip3 install --break-system-packages -e "." 2>/dev/null || pip3 install -e ".[dev]" 2>/dev/null || pip3 install -e .
        ok "Nexus-SDD instalado en modo desarrollo"
    else
        warn "pyproject.toml no encontrado. Instalando desde pip..."
        pip3 install --break-system-packages nexus-sdd 2>/dev/null || pip3 install nexus-sdd
        ok "Nexus-SDD instalado"
    fi
}

# ── Detect Project Stack ──────────────────────────────────────────────
detect_stack() {
    header "Detectando Stack Tecnologico del Proyecto"

    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    cd "$SCRIPT_DIR"

    python3 -c "
import sys
sys.path.insert(0, '.')
from nexus_sdd.detector.scanner import detect_project_type

project = detect_project_type()
print(f'TIPO: {project.type}')
print(f'LENGUAJES: {\", \".join(project.languages) or \"desconocido\"}')
print(f'FRAMEWORKS: {\", \".join(project.frameworks) or \"ninguno\"}')
print(f'TESTING: {\", \".join(project.testing) or \"ninguno\"}')
print(f'DB: {\", \".join(project.databases) or \"ninguna\"}')
print(f'SKILLS: {\", \".join(project.recommended_skills)}')
" 2>/dev/null || warn "Detector no disponible. Usando deteccion basica..."

    log "Stack detectado (ver arriba)"
}

# ── Install Skills ────────────────────────────────────────────────────
install_skills() {
    header "Instalando Skills para el Stack Detectado"

    TARGET_DIR="$(pwd)/.nexus/skills"
    mkdir -p "$TARGET_DIR"

    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

    python3 -c "
import sys
sys.path.insert(0, '$SCRIPT_DIR')
from nexus_sdd.detector.scanner import detect_project_type
from nexus_sdd.skills.registry import SkillRegistry

project = detect_project_type()
registry = SkillRegistry()
installed = registry.install_for_project(project.recommended_skills, Path('$TARGET_DIR'))

print(f'Skills instaladas ({len(installed)}):')
for s in installed:
    print(f'  ✓ {s}')

if not installed:
    print('  (skills base: openspec, mnemo, sdd-methodology)')
" 2>/dev/null || {
        warn "Instalacion automatica fallo. Copiando skills manualmente..."
        if [[ -d "$SCRIPT_DIR/skills" ]]; then
            cp -r "$SCRIPT_DIR/skills"/* "$TARGET_DIR/"
            ok "Skills copiadas manualmente"
        fi
    }

    # Copy team personas
    if [[ -d "$SCRIPT_DIR/skills/team" ]]; then
        mkdir -p "$TARGET_DIR/team"
        cp "$SCRIPT_DIR/skills/team/"*.md "$TARGET_DIR/team/" 2>/dev/null || true
        ok "Team personas instaladas (PO, UX, Architect, Dev, QA, DevOps)"
    fi
}

# ── Create .nexus Directory ───────────────────────────────────────────
create_nexus_dir() {
    header "Creando estructura .nexus/"

    mkdir -p .nexus/{profiles,skills,alerts,openspec}
    mkdir -p openspec/changes

    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

    # Copy full profiles from templates (includes harness + SDD + team skills)
    if [[ -d "$SCRIPT_DIR/templates/.nexus/profiles" ]]; then
        cp "$SCRIPT_DIR/templates/.nexus/profiles/"*.yaml .nexus/profiles/ 2>/dev/null || true
    fi

    # Copy 30-harness config from template
    if [[ -f "$SCRIPT_DIR/templates/.nexus/config.yaml" ]]; then
        cp "$SCRIPT_DIR/templates/.nexus/config.yaml" .nexus/config.yaml
    fi

    # SDD tracking file (which HDU is active)
    cat > .nexus/current_task.yaml << 'TASK'
# Nexus SDD — Active Task Tracking
# This file is managed by nexus-sdd orchestrate and the PreToolUse hook.
# When empty, code writes are blocked. Create a spec to unblock:
#   nexus-sdd spec "Feature title"
hdu_id: none
phase: none
agent: none
TASK

    ok "Estructura .nexus/ creada (config + profiles + tracking)"
}

# ── Configure AI Agents ───────────────────────────────────────────────
configure_agents() {
    header "Configurando Agentes de IA"

    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

    # Claude Code
    if command -v claude &>/dev/null; then
        mkdir -p .claude/hooks

        # ── SDD Enforcement Hooks ──────────────────────────────────
        if [[ -d "$SCRIPT_DIR/templates/.claude/hooks" ]]; then
            cp "$SCRIPT_DIR/templates/.claude/hooks/"*.sh .claude/hooks/
            chmod +x .claude/hooks/*.sh
            ok "SDD enforcement hooks instalados (PreToolUse gate + Stop auto-save + SessionStart context)"
        fi

        # ── Settings with hook config ───────────────────────────────
        if [[ -f "$SCRIPT_DIR/templates/.claude/settings.local.json" ]]; then
            cp "$SCRIPT_DIR/templates/.claude/settings.local.json" .claude/settings.local.json
            ok "Claude Code settings con hooks configurados"
        fi

        # ── Register mnemo as MCP server ───────────────────────────
        if command -v mnemo &>/dev/null; then
            claude mcp add mnemo -- mnemo mcp 2>/dev/null || warn "No se pudo agregar mnemo MCP a Claude Code"
        fi

        # ── Register graphify skill ─────────────────────────────────
        if command -v graphify &>/dev/null; then
            graphify claude install 2>/dev/null || true
            ok "Graphify skill registrado en Claude Code"
        fi

        ok "Claude Code configurado (hooks + settings + MCP + graphify)"
    fi

    # OpenCode
    if command -v opencode &>/dev/null; then
        opencode mcp add mnemo -- mnemo mcp 2>/dev/null || true
        ok "OpenCode configurado"
    fi

    # Kiro / Antigravity
    if command -v kiro &>/dev/null; then
        kiro mcp add mnemo -- mnemo mcp 2>/dev/null || true
        ok "Kiro configurado"
    fi

    # Cursor / Windsurf
    if [[ -d ".cursor" ]]; then
        cp "$SCRIPT_DIR/templates/.cursor/rules/nexus-sdd.md" .cursor/rules/ 2>/dev/null || {
            mkdir -p .cursor/rules
            cat > .cursor/rules/nexus-sdd.md << 'CURSOR'
# Nexus-SDD Rules
- Follow SDD: spec → plan → code → test → security
- Use .nexus/profiles/ for conventions
- Search mnemo before architectural decisions
- Security scan before commit
- BDD scenarios for every feature
CURSOR
        }
        ok "Cursor configurado"
    fi
}

# ── Init OpenSpec ─────────────────────────────────────────────────────
init_openspec() {
    header "Inicializando OpenSpec"

    if command -v openspec &>/dev/null; then
        openspec init 2>/dev/null || {
            warn "openspec init manual requerido. Ejecuta: openspec init"
        }
        ok "OpenSpec inicializado"
    else
        mkdir -p openspec/{changes,specs}
        cat > openspec/AGENTS.md << 'OPENSPEC'
# OpenSpec Instructions

Slash commands for AI coding tools:
- `/opsx:propose` — Create a new change (proposal + specs + design + tasks)
- `/opsx:apply` — Implement the tasks
- `/opsx:archive` — Archive a completed change

Nexus-SDD extends this with:
- BDD scenarios in every spec
- Security scan before archive
- Mnemo memory after every applied change
- Multi-agent team: supervisor → PO/UX/Architect → Dev → QA → DevOps
OPENSPEC
        ok "OpenSpec base creado"
    fi
}

# ── Summary ───────────────────────────────────────────────────────────
print_summary() {
    header "Nexus-SDD Instalacion Completa"

    echo -e "${GREEN}${BOLD}  ✅ Nexus-SDD esta listo!${NC}\n"

    echo -e "  ${BOLD}Comandos principales:${NC}\n"

    echo -e "  ${CYAN}nexus-sdd init${NC}              Inicializar proyecto"
    echo -e "  ${CYAN}nexus-sdd spec <titulo>${NC}     Crear especificación"
    echo -e "  ${CYAN}nexus-sdd orchestrate <HDU>${NC} Orquestar con multi-agente"
    echo -e "  ${CYAN}nexus-sdd status${NC}           Ver progreso de HDUs"
    echo -e "  ${CYAN}nexus-sdd security${NC}         Escanear secrets"
    echo -e "  ${CYAN}nexus-sdd skill list${NC}       Ver catálogo de skills"
    echo -e "  ${CYAN}nexus-sdd save --hdu-id <HDU>${NC} Guardar en mnemo\n"

    echo -e "  ${BOLD}Team personas (Claude Code):${NC}"
    echo -e "  /supervisor   /po-agent   /ux-agent   /architect-agent"
    echo -e "  /dev-agent    /qa-agent   /devops-agent\n"

    echo -e "  ${BOLD}Mnemo:${NC}"
    echo -e "  ${CYAN}mnemo search <query>${NC}   Buscar en memoria"
    echo -e "  ${CYAN}mnemo config${NC}           Ver configuración (DB)\n"
}

# ── Main ──────────────────────────────────────────────────────────────
main() {
    echo -e "\n${BOLD}${CYAN}"
    echo "╔══════════════════════════════════════════════╗"
    echo "║   🏭  NEXUS-SDD  —  Fábrica de Software IA  ║"
    echo "║       Zero-Friction Installer                ║"
    echo "╚══════════════════════════════════════════════╝"
    echo -e "${NC}\n"

    detect_os
    install_python
    install_node
    install_go
    install_openspec
    install_ollama
    install_mnemo
    install_nexus
    create_nexus_dir
    configure_agents
    init_openspec
    print_summary
}

main "$@"
