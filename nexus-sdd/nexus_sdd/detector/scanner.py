"""
Project Type Detector — scans for lock files, configs, framework markers.

Referenced by install.sh: from nexus_sdd.detector.scanner import detect_project_type
"""

from pathlib import Path
from dataclasses import dataclass, field


@dataclass
class ProjectType:
    type: str              # web, backend, mobile, ai, cli, library, infra
    languages: list[str] = field(default_factory=list)
    frameworks: list[str] = field(default_factory=list)
    testing: list[str] = field(default_factory=list)
    databases: list[str] = field(default_factory=list)
    recommended_profile: str = "fullstack"
    recommended_skills: list[str] = field(default_factory=list)


# Markers: (file_pattern, (language, framework))
MARKERS = {
    # Python
    "pyproject.toml": ("python", "setuptools"),
    "requirements.txt": ("python", "pip"),
    "Pipfile": ("python", "pipenv"),
    "poetry.lock": ("python", "poetry"),
    # Node/TypeScript
    "package.json": ("typescript", "node"),
    "tsconfig.json": ("typescript", "typescript"),
    "next.config.js": ("typescript", "nextjs"),
    "next.config.ts": ("typescript", "nextjs"),
    "vite.config.js": ("typescript", "vite"),
    "vite.config.ts": ("typescript", "vite"),
    "tailwind.config.js": ("typescript", "tailwind"),
    "tailwind.config.ts": ("typescript", "tailwind"),
    # Go
    "go.mod": ("go", "go"),
    "go.sum": ("go", "go"),
    # Rust
    "Cargo.toml": ("rust", "rust"),
    # Mobile
    "Podfile": ("swift", "ios"),
    "build.gradle": ("kotlin", "android"),
    "build.gradle.kts": ("kotlin", "android"),
    "app.json": ("typescript", "react-native"),
    # Infra
    "Dockerfile": ("docker", "docker"),
    "docker-compose.yml": ("docker", "docker"),
    "docker-compose.yaml": ("docker", "docker"),
    ".github/workflows": ("yaml", "github-actions"),
    # AI/ML
    "langgraph.json": ("python", "langgraph"),
    "langgraph.config.json": ("python", "langgraph"),
    # DB
    "schema.prisma": ("prisma", "prisma"),
    "migrations": ("sql", "database"),
}

# Framework deep detection (check file contents)
FRAMEWORK_DEEP = {
    "fastapi": {"pyproject.toml": "fastapi", "requirements.txt": "fastapi", "main.py": "from fastapi"},
    "django": {"pyproject.toml": "django", "requirements.txt": "django", "manage.py": "django"},
    "flask": {"pyproject.toml": "flask", "requirements.txt": "flask"},
    "react": {"package.json": '"react"'},
    "vue": {"package.json": '"vue"'},
    "svelte": {"package.json": '"svelte"'},
    "langgraph": {"pyproject.toml": "langgraph", "requirements.txt": "langgraph"},
    "pytest": {"pyproject.toml": "pytest", "requirements.txt": "pytest"},
    "playwright": {"package.json": '"playwright"', "pyproject.toml": "playwright"},
}


def detect_project_type(path: Path | None = None) -> ProjectType:
    """
    Detect project type by scanning for lock files, configs, framework markers.
    Returns a ProjectType with recommended profile and skills.
    """
    root = Path(path) if path else Path.cwd()
    result = ProjectType(type="cli")

    if not root.exists():
        return result

    all_files = set()
    for f in root.rglob("*"):
        if f.is_file():
            all_files.add(f.name)
        # Also check for directory markers
        if f.is_dir() and f.relative_to(root).as_posix() in MARKERS:
            all_files.add(f.relative_to(root).as_posix())

    # Only check top-level files for framework markers
    top_files = {f.name for f in root.iterdir() if f.is_file()}
    top_dirs = {f.name for f in root.iterdir() if f.is_dir()}

    languages: set[str] = set()
    frameworks: set[str] = set()
    databases: set[str] = set()

    # Scan markers
    for marker, (lang, fw) in MARKERS.items():
        if marker in top_files or marker in top_dirs:
            languages.add(lang)
            if fw not in ("node", "typescript", "go", "rust", "ios", "android"):
                frameworks.add(fw)

    # Deep detection
    for fw_name, checks in FRAMEWORK_DEEP.items():
        for check_file, check_pattern in checks.items():
            check_path = root / check_file
            if check_path.exists():
                try:
                    content = check_path.read_text()
                    if check_pattern in content:
                        frameworks.add(fw_name)
                except Exception:
                    pass

    result.languages = sorted(languages)
    result.frameworks = sorted(frameworks)

    # Determine project type
    if "react" in frameworks or "nextjs" in frameworks or "vue" in frameworks or "svelte" in frameworks:
        result.type = "web"
    elif "fastapi" in frameworks or "django" in frameworks or "flask" in frameworks or "go" in frameworks:
        result.type = "backend"
    elif "react-native" in frameworks or "ios" in result.languages or "android" in result.languages:
        result.type = "mobile"
    elif "langgraph" in frameworks:
        result.type = "ai"
    elif "docker" in frameworks or "github-actions" in frameworks:
        result.type = "infra"
    elif len(result.languages) >= 3:
        result.type = "library"
    elif len(frameworks) >= 2:
        result.type = "fullstack"

    # Recommend profile and skills
    if result.type == "ai" or "langgraph" in frameworks:
        result.recommended_profile = "fullstack-python-langgraph"
        result.recommended_skills = ["python-core", "langgraph-python", "langgraph-aws", "pytest"]
    elif result.type == "web" or "react" in frameworks or "nextjs" in frameworks:
        result.recommended_profile = "react-nextjs"
        result.recommended_skills = ["react", "nextjs", "tailwind", "typescript", "playwright"]
    elif "go" in result.languages:
        result.recommended_profile = "fullstack-go"
        result.recommended_skills = ["go-core", "go-fiber", "sql-database", "docker-kubernetes"]
    elif "python" in result.languages:
        result.recommended_profile = "fullstack"
        result.recommended_skills = ["python-core", "fastapi", "pytest"]
    else:
        result.recommended_profile = "minimal"
        result.recommended_skills = []

    # Detect testing frameworks
    if "pytest" in frameworks or ("pyproject.toml" in top_files and "pytest" in (root / "pyproject.toml").read_text()):
        result.testing.append("pytest")
    if "playwright" in frameworks:
        result.testing.append("playwright")
    if "go" in languages:
        result.testing.append("go-test")

    # Detect databases
    if "schema.prisma" in top_files:
        databases.add("prisma")
    if any(d in top_dirs for d in ["migrations", "alembic"]):
        databases.add("sql")
    result.databases = sorted(databases)

    return result
