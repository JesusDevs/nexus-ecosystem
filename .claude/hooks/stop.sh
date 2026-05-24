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
                "Progress saved by Gingx Stop hook. HDU: $HDU" \
                --type progress --outcome in_progress --project "$PROJECT" \
                2>/dev/null || true

            # Sync HDU to mnemo (dual-write)
            if [[ -f "$HDU_DIR/$HDU.yaml" ]]; then
                python3 -c "
import yaml, subprocess, sys
try:
    with open('$HDU_DIR/$HDU.yaml') as f:
        data = yaml.safe_load(f) or {}
    hdu_id = data.get('id', '$HDU')
    title = data.get('title', hdu_id)
    phase = data.get('phase', 'init')
    status = data.get('status', 'active')
    summary = data.get('executive_summary', '')
    subprocess.run([
        'mnemo', 'hdu', 'save', hdu_id,
        '--title', title,
        '--phase', phase,
        '--status', status,
        '--project', '$PROJECT',
        '--content', summary or title,
    ], capture_output=True, timeout=5)
except Exception:
    pass
" 2>/dev/null || true
            fi
        fi
    fi
fi

# ── CodeGraph sync ──────────────────────────────────────────────────
if command -v codegraph &>/dev/null && [[ -d ".codegraph" ]]; then
    codegraph sync 2>/dev/null || true
fi

# ── Knowledge Graph incremental update ──────────────────────────────
# Update domain last_indexed timestamps and export top symbols to mnemo
KNOWLEDGE_DIR=".gingx/knowledge"
if [[ -f "$KNOWLEDGE_DIR/domain-map.yaml" ]]; then
    python3 -c "
import yaml
from datetime import datetime
from pathlib import Path

# Update timestamp
kf = Path('$KNOWLEDGE_DIR/domain-map.yaml')
data = yaml.safe_load(kf.read_text()) or {}
for d in data.get('domains', []):
    d['last_indexed'] = datetime.now().strftime('%Y-%m-%d %H:%M')
data['_updated_at'] = datetime.now().strftime('%Y-%m-%d %H:%M')
kf.write_text(yaml.dump(data, default_flow_style=False, allow_unicode=True, sort_keys=False))
" 2>/dev/null || true
fi

# ── Regenerate Obsidian vault ──────────────────────────────────────
# Auto-refresh the vault so graph view stays current
if [[ -f "$KNOWLEDGE_DIR/domain-map.yaml" ]]; then
    python3 -c "
import yaml, json
from pathlib import Path
from datetime import datetime

kdir = Path('$KNOWLEDGE_DIR')
vdir = kdir / 'vault'
vdir.mkdir(parents=True, exist_ok=True)

domain_data = yaml.safe_load((kdir / 'domain-map.yaml').read_text()) if (kdir / 'domain-map.yaml').exists() else {}
comp_data = yaml.safe_load((kdir / 'component-index.yaml').read_text()) if (kdir / 'component-index.yaml').exists() else {}
dec_data = yaml.safe_load((kdir / 'decisions-log.yaml').read_text()) if (kdir / 'decisions-log.yaml').exists() else {}
domains = domain_data.get('domains', [])
edges = domain_data.get('cross_domain_edges', [])
components = comp_data.get('components', {})
decisions = dec_data.get('decisions', [])
proj = Path.cwd().name

domain_lookup = {d['name']: d for d in domains}
domain_deps = {}
for edge in edges:
    src, tgt = edge.get('from', ''), edge.get('to', '')
    if src and tgt:
        domain_deps.setdefault(src, []).append((tgt, edge.get('via', [])))

comp_domain = {}
for name, comp in components.items():
    cp = comp.get('path', '')
    for d in domains:
        dp = d.get('path', '')
        if dp and cp.startswith(dp):
            comp_domain[name] = d['name']
            break

today = datetime.now().strftime('%Y-%m-%d')
count = 0

# Home.md
lines = ['---', f'project: {proj}', f'date: {today}', 'type: moc', 'tags: [knowledge-graph, moc]', '---', '',
         f'# {proj} — Knowledge Graph', '', '## Domains', '']
for d in domains:
    dname = d['name']
    desc = d.get('description', '')[:120]
    pats = ', '.join(d.get('patterns', []))
    lines.append(f'- **[[domains/{dname}|{dname}]]** — {desc}')
    if pats:
        lines.append(f'  - _patterns_: {pats}')
lines.append('')
lines.append('## Cross-Domain Edges')
lines.append('')
for edge in edges:
    lines.append(f'- [[domains/{edge.get(\"from\",\"\")}|{edge.get(\"from\",\"\")}]] → [[domains/{edge.get(\"to\",\"\")}|{edge.get(\"to\",\"\")}]]')
lines.append('')
if components:
    lines.append('## Key Components'); lines.append('')
    for name in sorted(components.keys()):
        lines.append(f'- [[components/{name}]]')
    lines.append('')
lines.append('## Graph View')
lines.append('')
lines.append('Open in [Obsidian](https://obsidian.md) → Graph View (Ctrl+G).')
lines.append('')
lines.append('> Auto-refreshed by Gingx stop hook.')
(vdir / 'Home.md').write_text('\n'.join(lines) + '\n')
count += 1

# Domain notes
dom_dir = vdir / 'domains'; dom_dir.mkdir(exist_ok=True)
for d in domains:
    dname, dpath = d['name'], d.get('path', '')
    desc = d.get('description', '')
    symbols = d.get('key_symbols', [])
    patterns = d.get('patterns', [])
    deps = d.get('dependencies', [])
    dedges = domain_deps.get(dname, [])
    dl = ['---', f'domain: {dname}', f'path: {dpath}', f'tags: [domain, {dname}]']
    if patterns: dl.append(f'patterns: [{chr(44).join(patterns)}]')
    if deps: dl.append(f'dependencies: [{chr(44).join(deps)}]')
    dl.extend(['---', '', f'# {dname}', '', desc, ''])
    if symbols:
        dl.append('## Key Symbols'); dl.append('')
        for s in symbols: dl.append(f'- `{s}`')
        dl.append('')
    if deps:
        dl.append('## Dependencies'); dl.append('')
        for dep in deps: dl.append(f'- [[{dep}]]')
        dl.append('')
    if dedges:
        dl.append('## Cross-Domain Edges'); dl.append('')
        for tgt, via in dedges:
            dl.append(f'- → [[{tgt}]]')
            for v in via: dl.append(f'  - `{v}`')
        dl.append('')
    if patterns:
        dl.append('## Patterns'); dl.append('')
        for p in patterns: dl.append(f'- `{p}`')
        dl.append('')
    backlinks = [(e.get('from',''), e.get('via',[])) for e in edges if e.get('to') == dname]
    if backlinks:
        dl.append('## Depended On By'); dl.append('')
        for src, via in backlinks:
            dl.append(f'- [[{src}]]')
            for v in via: dl.append(f'  - `{v}`')
        dl.append('')
    dc = [c for c, cd in comp_domain.items() if cd == dname]
    if dc:
        dl.append('## Components'); dl.append('')
        for c in sorted(dc): dl.append(f'- [[../components/{c}|{c}]]')
        dl.append('')
    (dom_dir / f'{dname}.md').write_text('\n'.join(dl) + '\n')
    count += 1

# Component notes
if components:
    comp_dir = vdir / 'components'; comp_dir.mkdir(exist_ok=True)
    for name, comp in components.items():
        cp = comp.get('path', ''); ct = comp.get('type', '')
        cr = comp.get('role', ''); callers = comp.get('callers', [])
        callees = comp.get('callees', []); ctests = comp.get('tests', [])
        cd = comp_domain.get(name, '')
        cl = ['---', f'component: {name}', f'type: {ct}']
        if cd: cl.append(f'domain: {cd}')
        cl.extend(['---', '', f'# {name}', '', f'**Type:** `{ct}`', f'**Path:** `{cp}`', '', cr, ''])
        if cd: cl.append(f'**Domain:** [[../domains/{cd}|{cd}]]'); cl.append('')
        if callers: cl.append('## Callers'); cl.append('')
        for c in callers: cl.append(f'- `{c}`')
        if callers: cl.append('')
        if callees: cl.append('## Callees'); cl.append('')
        for c in callees: cl.append(f'- `{c}`')
        if callees: cl.append('')
        if ctests: cl.append('## Tests'); cl.append('')
        for t in ctests: cl.append(f'- `{t}`')
        if ctests: cl.append('')
        (comp_dir / f'{name}.md').write_text('\n'.join(cl) + '\n')
        count += 1

# Decision notes
if decisions:
    dec_dir = vdir / 'decisions'; dec_dir.mkdir(exist_ok=True)
    for dec in decisions:
        dt = dec.get('title', 'untitled'); dd = dec.get('date', '')
        dr = dec.get('rationale', ''); dto = dec.get('trade_off', '')
        df = dt.lower().replace(' ', '-').replace('/', '-')
        dl2 = ['---', f'title: {dt}', f'date: {dd}', 'tags: [decision, adr]', '---', '', f'# {dt}', '']
        if dr: dl2.append(f'## Rationale\n\n{dr}\n')
        if dto: dl2.append(f'## Trade-off\n\n{dto}\n')
        (dec_dir / f'{df}.md').write_text('\n'.join(dl2) + '\n')
        count += 1

print(f'[knowledge] vault refreshed: {count} files', file=__import__('sys').stderr)
" 2>/dev/null || true
fi

# Export knowledge to mnemo for cross-session memory
if command -v mnemo &>/dev/null && [[ -f "$KNOWLEDGE_DIR/domain-map.yaml" ]]; then
    SUMMARY=$(python3 -c "
import yaml
with open('$KNOWLEDGE_DIR/domain-map.yaml') as f:
    data = yaml.safe_load(f) or {}
domains = [d['name'] for d in data.get('domains', [])]
print(f'{len(domains)} domains: ' + ', '.join(domains[:8]))
" 2>/dev/null || echo "knowledge graph")
    mnemo save "Knowledge Graph: $PROJECT — session end" \
        "$SUMMARY" \
        --type progress --outcome in_progress --project "$PROJECT" \
        --tags knowledge-graph,explorer,session-end \
        2>/dev/null || true
fi

# ── Update graphify graph (incremental, code-only via AST) ─────────
# Fast path: AST-only for code changes. Docs changes need manual /graphify .
if command -v graphify &>/dev/null && [[ -d "graphify-out" ]] && [[ -f "graphify-out/.graphify_python" ]]; then
    GRAPHIFY_PY=$(cat graphify-out/.graphify_python 2>/dev/null)
    if [[ -n "$GRAPHIFY_PY" ]]; then
        "$GRAPHIFY_PY" -c "
import json, sys
from pathlib import Path
from graphify.detect import detect_incremental, save_manifest

try:
    result = detect_incremental(Path('.'))
    new_total = result.get('new_total', 0)
    if new_total == 0:
        print('', file=sys.stderr)  # no changes
        sys.exit(0)
    code_exts = {'.py','.ts','.js','.go','.rs','.java','.cpp','.c','.rb','.swift','.kt','.cs','.scala','.php','.sh'}
    changed = [f for files in result.get('new_files', {}).values() for f in files]
    code_only = all(Path(f).suffix.lower() in code_exts for f in changed)
    if code_only and changed:
        # AST-only update — fast, no LLM needed
        from graphify.extract import collect_files, extract
        code_files = []
        for f in changed:
            if Path(f).suffix.lower() in code_exts:
                code_files.extend(collect_files(Path(f)) if Path(f).is_dir() else [Path(f)])
        if code_files:
            ast_result = extract(code_files, cache_root=Path('.'))
            from networkx.readwrite import json_graph
            import networkx as nx
            old_data = json.loads(Path('graphify-out/graph.json').read_text())
            G = json_graph.node_link_graph(old_data, edges='links')
            G_new = __import__('graphify.build', fromlist=['build_from_json']).build_from_json(ast_result)
            G.update(G_new)
            Path('graphify-out/graph.json').write_text(json.dumps(json_graph.node_link_data(G, edges='links'), indent=2))
            save_manifest(result['files'])
            print(f'[graphify] AST-only update: {G.number_of_nodes()} nodes, {G.number_of_edges()} edges')
        else:
            save_manifest(result['files'])
    else:
        print(f'[graphify] {new_total} file(s) changed (non-code). Run /graphify . --update for semantic refresh.', file=sys.stderr)
except Exception as e:
    pass  # silent fail — graphify update is best-effort
" 2>/dev/null || true
    fi
fi

# ── Mnemo sync push ──────────────────────────────────────────────────
if command -v mnemo &>/dev/null; then
    PUSH_OUTPUT=$(mnemo sync push 2>/dev/null || echo "")
    if [[ -n "$PUSH_OUTPUT" && "$PUSH_OUTPUT" != *"no sync.remote"* ]]; then
        echo "[mnemo] pushed to remote" >&2
    fi
fi

echo "{\"systemMessage\": \"Session ended. HDU progress saved. Run gingx-sdd hdu summary to see all.\"}"
exit 0
