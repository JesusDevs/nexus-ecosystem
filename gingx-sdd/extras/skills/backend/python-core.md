---
name: python-core
description: Core Python patterns — typing, async, context managers, dataclasses, protocols
category: backend
stack: [python, asyncio, typing]
triggers: [python, backend, api, cli]
---

# Python Core Patterns

## Rules
- Use type hints everywhere: function signatures, class attributes, return types
- Prefer `async/await` for I/O-bound operations, avoid for CPU-bound
- Use `with` (context managers) for resource management
- `dataclasses` for data containers, `Protocol` for structural subtyping
- `pathlib.Path` over `os.path` — always
- f-strings over `.format()` or `%` formatting
- `match/case` (Python 3.10+) for structural pattern matching
- List/dict comprehensions for simple transforms (max 2 levels deep)

## Do's
- Single-source imports: `from module import X`
- Handle exceptions at the right level — don't swallow silently
- Use `logging` (not `print`) with module-level loggers
- Use `.env` + `os.getenv` for configuration
- Write docstrings for public APIs (one line, not novels)

## Don'ts
- No `from module import *` — pollutes namespace
- No mutable default arguments
- No bare `except:` — always specify exception type
- No `os.system()` — use `subprocess.run()`
- No global mutable state — use dependency injection

## Recommended Commands
```bash
python3 -m pytest -xvs
python3 -m ruff check .
python3 -m mypy .
python3 -m pip install -e .
```
