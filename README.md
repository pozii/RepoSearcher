<div align="center">

# RepoSearcher

**A fast, feature-rich CLI tool for searching code across repositories**

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg?style=for-the-badge&logo=opensourceinitiative&logoColor=white)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20macOS%20%7C%20Linux-blue?style=for-the-badge&logo=WindowsTerminal&logoColor=white)](#installation)

</div>

---

> SIMD-accelerated code search with .gitignore support, fuzzy matching, GitHub integration, and an interactive TUI — all in a single binary.

## Features

### Core Search
- **Keyword Search** — Fast literal matching (SIMD-accelerated via `bytes.Index`)
- **Regex Support** — Two-phase search with literal prefix extraction
- **Fuzzy Search** — Tolerates typos (e.g. "fucntion" finds "function") using Levenshtein + Jaro-Winkler
- **Smart Suggestions** — Auto-complete search terms from codebase identifiers
- **Case-Insensitive** — `--ignore-case` flag
- **Context Lines** — `--context N` lines around matches

### Performance
- **Literal-First Engine** — `bytes.Index` (AVX2) scans files before regex; 4.6x faster than naive approach
- **[]byte-Based Search** — 99% fewer memory allocations vs line-by-line `bufio.Scanner`
- **Parallel Processing** — Multi-core worker pool with parallel directory walking
- **Pattern Caching** — Compiled regex patterns cached across search calls
- **Binary File Skip** — Instantly ignores non-code files

### Filtering
- **`.gitignore` Support** — Automatically respects `.gitignore` rules
- **`--exclude` Glob** — Exclude files by pattern (`--exclude "vendor/**,test_*"`)
- **`--include` Glob** — Include only matching files (`--include "src/*.go"`)
- **`--extensions`** — Filter by file type (`.go,.py,.js`)

### Git Integration
- **`--since`** — Search files changed since date (e.g. `1 week ago`)
- **`--author`** — Filter by commit author
- **`--changed-in`** — Search files in commit range (e.g. `HEAD~5`)
- **`--commit-message`** — Search commit messages instead of file content

### GitHub Integration
- **Remote Search** — Search GitHub repos via Codesearch API
- **Token Support** — `--github-token` or `GITHUB_TOKEN` env var
- **Pagination** — Fetches up to 5 pages (150 results)

### Output
- **Streaming** — Results appear as found (no waiting for full search)
- **Colored Output** — File paths (cyan), line numbers (green), matches (red)
- **JSON Export** — `--json results.json`
- **CSV Export** — `--csv results.csv`

### Symbol Search (No LSP Required)
- **Symbol Finder** — Find functions, structs, variables
- **Definition Finder** — Locate where a symbol is defined
- **Reference Finder** — Find all uses (word-boundary matching)
- **Multi-Language** — Go, Python, JavaScript/TypeScript

### Interactive TUI
- **Full-Screen Search** — Bubbletea-powered terminal UI
- **Vim Navigation** — `j/k`, `g/G`, Enter, Esc

### Other
- **Auto-Update** — Daily cached checks (no network on every run)
- **Cross-Platform** — Windows, macOS, Linux

---

## Installation

### From Source (Recommended — No AV Issues)

```bash
go install github.com/pozii/RepoSearcher@v1.6.0
```

This compiles from source on your machine — antivirus won't flag it.

### Quick Install

**Windows (PowerShell):**
```powershell
irm https://raw.githubusercontent.com/pozii/RepoSearcher/master/install.ps1 | iex
```

**macOS/Linux:**
```bash
curl -sSL https://raw.githubusercontent.com/pozii/RepoSearcher/master/install.sh | bash
```

> **Note:** Antivirus software (Avast, Windows Defender, etc.) may flag the downloaded binary because it's unsigned. If this happens, use `go install` instead, or add an exception after verifying the SHA256 checksum from the release page.

### Download Binary

Download the latest release from [Releases](https://github.com/pozii/RepoSearcher/releases).

| Platform | File |
|----------|------|
| Windows AMD64 | `repo-searcher-windows-amd64.exe` |
| Windows ARM64 | `repo-searcher-windows-arm64.exe` |
| macOS Intel | `repo-searcher-darwin-amd64` |
| macOS Apple Silicon | `repo-searcher-darwin-arm64` |
| Linux AMD64 | `repo-searcher-linux-amd64` |
| Linux ARM64 | `repo-searcher-linux-arm64` |

### Build from Source

```bash
git clone https://github.com/pozii/RepoSearcher.git
cd RepoSearcher
go build -o repo-searcher .
```

### Add to PATH

After downloading or building, run:

```bash
repo-searcher install
```

This will automatically add repo-searcher to your PATH.

---

## Usage

### Basic Search

```bash
# Search in local directory
repo-searcher search "func" ./project

# Search with regex
repo-searcher search "func\s+\w+\(" ./src --regex

# Case-insensitive search
repo-searcher search "TODO" ./project --ignore-case

# Lines of context around matches
repo-searcher search "error" ./src --context 2
```

### Exclude / Include Files

```bash
# Exclude vendor and test files
repo-searcher search "func" ./project --exclude "vendor/**,test_*"

# Only search src directory
repo-searcher search "import" . --include "src/*"

# Combine with extensions
repo-searcher search "handler" ./src --include "*.go" --extensions .go
```

### Fuzzy Search (Tolerates Typos)

```bash
# Search with typo tolerance
repo-searcher search "fucntion" ./src --fuzzy
# Results: "function", "functions", "funct"

# The algorithm uses:
# - Levenshtein Distance (edit distance)
# - Jaro-Winkler Similarity (optimized for code names)
```

### Smart Suggestions

```bash
# Get AI-powered search suggestions (no API required)
repo-searcher search "pars" ./src --suggest
# Results: "parse", "parseJSON", "parseInt", "parseError"
```

### Multi-Directory Search

```bash
# Search across multiple directories
repo-searcher search "error" ./frontend ./backend ./shared
```

### File Extension Filter

```bash
# Search only Go files
repo-searcher search "import" ./project --extensions .go

# Search multiple file types
repo-searcher search "function" ./src --extensions .go,.py,.js
```

### GitHub Search

```bash
# Search public repository
repo-searcher search "TODO" octocat/Hello-World --github

# Search private repository (requires token)
repo-searcher search "api" owner/repo --github --github-token ghp_xxx

# Or set environment variable
export GITHUB_TOKEN=ghp_xxx
repo-searcher search "api" owner/repo --github
```

### Git History Search

```bash
# Search files changed in last week
repo-searcher search "error" ./project --since "1 week ago"

# Search files by specific author
repo-searcher search "TODO" ./project --author "john"

# Search files changed in last 5 commits
repo-searcher search "bug" ./project --changed-in HEAD~5

# Search commit messages
repo-searcher search "refactor" ./project --commit-message
```

### Export Results

```bash
# Export to JSON
repo-searcher search "func" ./project --json results.json

# Export to CSV
repo-searcher search "error" ./project --csv results.csv
```

### Interactive TUI

```bash
# Launch in current directory
repo-searcher i .

# Launch with regex mode
repo-searcher i ./src --regex

# Launch with file extensions
repo-searcher i ./project --extensions .go,.py
```

Keyboard shortcuts in TUI mode:
- `j/k` or `Up/Down` - Navigate results
- `Enter` - Execute search
- `g/G` - Jump to top/bottom
- `e` - Export results to JSON
- `q/Esc` - Quit

### Symbol Search (No LSP Required)

```bash
# Find symbols matching a name
repo-searcher find-symbol "Engine" .

# Find in Go files only
repo-searcher find-symbol "ParseJSON" ./src --ext .go

# Find definition of a symbol
repo-searcher find-definition "SearchEngine" .

# Find all references to a symbol
repo-searcher find-references "config" ./src

# Find across multiple directories
repo-searcher find-references "NewEngine" ./src ./lib
```

---

## Auto-Update

RepoSearcher checks for updates daily (cached in `~/.repo-searcher/update-check.json`). When a new version is available:

```
New version available: v1.6.1 (current: v1.6.0)
Run 'repo-searcher update' to update.
```

### Update Commands

```bash
# Interactive update
repo-searcher update

# Auto-confirm update
repo-searcher update --yes

# Skip update check
repo-searcher search "func" ./project --no-update-check
```

### Uninstall

```bash
repo-searcher uninstall
```

---

## All Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--regex` | Enable regex mode | `false` |
| `--fuzzy` | Fuzzy search (tolerates typos) | `false` |
| `--suggest` | Show smart suggestions | `false` |
| `--ignore-case` | Case-insensitive search | `false` |
| `--github` | Search GitHub repositories | `false` |
| `--github-token` | GitHub API token (or `GITHUB_TOKEN` env) | - |
| `--json <file>` | Export to JSON | - |
| `--csv <file>` | Export to CSV | - |
| `--extensions` | Filter by extensions (`.go,.py`) | all |
| `--exclude` | Exclude files by glob (`vendor/**,test_*`) | - |
| `--include` | Include only matching globs (`src/*.go`) | - |
| `--context` | Lines of context around matches | `0` |
| `--since` | Files changed since (e.g. `1 week ago`) | - |
| `--author` | Filter by commit author | - |
| `--changed-in` | Files in commit range (e.g. `HEAD~5`) | - |
| `--commit-message` | Search commit messages instead of files | `false` |
| `--no-update-check` | Skip automatic update check | `false` |

---

## Performance

RepoSearcher uses SIMD-accelerated literal search for maximum speed:

| Search Type | Technique | vs Naive Approach |
|-------------|-----------|-------------------|
| Literal query | `bytes.Index` (AVX2) | **1.9x faster, 99% less allocs** |
| Regex with prefix | Two-phase (literal scan + regex) | **4.6x faster, 99% less allocs** |
| File traversal | Parallel directory walking | **~4x faster** on multi-core |
| Pattern matching | `sync.Map` cache | **0ns** for cached patterns |

---

## Project Structure

```
repo-searcher/
├── cmd/
│   ├── root.go           # Root command + auto-update
│   ├── search.go         # Search subcommand
│   ├── version.go        # Version command
│   ├── install.go        # PATH installation
│   ├── update.go         # Manual update
│   ├── uninstall.go      # Remove from PATH
│   ├── interactive.go    # Interactive TUI command
│   ├── finddef.go        # Find definition
│   ├── findref.go        # Find references
│   ├── findsymbol.go     # Find symbols
│   └── completions.go    # Shell completions
├── internal/
│   ├── search/
│   │   ├── fast.go       # FastPattern, literal-first engine
│   │   ├── local.go      # Local filesystem search
│   │   ├── github.go     # GitHub Codesearch API
│   │   ├── git.go        # Git history search
│   │   ├── matcher.go    # Regex/keyword/glob matching
│   │   ├── fuzzy.go      # Fuzzy search (Levenshtein, Jaro-Winkler)
│   │   ├── suggest.go    # Smart suggestion engine
│   │   └── performance.go# Parallel engine + file collection
│   ├── fileutil/
│   │   ├── fileutil.go   # ShouldSkipDir (shared)
│   │   ├── gitignore.go  # .gitignore parsing
│   │   └── parallel_walk.go # Parallel directory walking
│   ├── export/
│   │   ├── json.go       # JSON export
│   │   └── csv.go        # CSV export
│   ├── output/
│   │   └── color.go      # Colored terminal output + streaming
│   ├── installer/
│   │   └── installer.go  # Platform-specific PATH logic
│   ├── updater/
│   │   ├── updater.go    # GitHub release check + download
│   │   └── cache.go      # Daily update check cache
│   ├── tui/
│   │   ├── tui.go        # Interactive TUI model
│   │   └── styles.go     # TUI styling
│   └── lsp/
│       └── symbol.go     # Symbol extraction and indexing
├── pkg/
│   └── models/
│       ├── result.go     # SearchResult, SearchConfig
│       └── git.go        # GitSearchConfig
├── .github/
│   └── workflows/
│       ├── go.yml        # CI (build + test + lint)
│       └── release.yml   # Auto-release on tag
├── install.sh            # macOS/Linux installer
├── install.ps1           # Windows installer
├── build.sh              # Cross-compile (Unix)
├── build.ps1             # Cross-compile (Windows)
├── go.mod
└── main.go
```

---

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

<div align="center">

Made with ❤️ and Go

</div>
