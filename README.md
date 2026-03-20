<div align="center">

# RepoSearcher

**A powerful CLI tool for searching code across multiple repositories**

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg?style=for-the-badge&logo=opensourceinitiative&logoColor=white)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20macOS%20%7C%20Linux-blue?style=for-the-badge&logo=WindowsTerminal&logoColor=white)](#installation)

</div>

---

> Search with keywords or regex across local directories and GitHub repositories with beautiful colored output and export options.

## Demo

<!-- GIF 1: Main search demo with colored output -->
<!-- Content: repo-searcher search "func" ./project --extensions .go -->
<!-- Show: File paths (cyan), line numbers (green), matches (red bold) -->

![Demo GIF - Placeholder](./assets/demo-search.gif)

*Replace this with your GIF showing colored search results*

---

## Features

- **Keyword Search** - Fast text matching across files
- **Regex Support** - Powerful pattern matching with `--regex`
- **Colored Output** - Beautiful terminal results (Cyan, Green, Red)
- **Multi-Directory** - Search across multiple repos/directories
- **Export Options** - Save results as JSON or CSV
- **Extension Filter** - Search only specific file types (`.go`, `.py`, `.js`)
- **GitHub Integration** - Search public repos via Codesearch API
- **GitHub Token** - Private repo support with `--github-token`
- **Auto-Update** - Checks for updates on every run
- **Cross-Platform** - Works on Windows, macOS, Linux

---

## Installation

### Quick Install (Recommended)

**Windows (PowerShell):**
```powershell
irm https://raw.githubusercontent.com/pozii/RepoSearcher/main/install.ps1 | iex
```

**macOS/Linux:**
```bash
curl -sSL https://raw.githubusercontent.com/pozii/RepoSearcher/main/install.sh | bash
```

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
```

### Export Results

```bash
# Export to JSON
repo-searcher search "func" ./project --json results.json

# Export to CSV
repo-searcher search "error" ./project --csv results.csv
```

---

## Demos (GIFs)

<!-- GIF 2: Regex search demo -->
<!-- Content: repo-searcher search "func\s+\w+\(" ./src --regex -->
<!-- Show: Pattern matching in action -->

![Regex GIF - Placeholder](./assets/demo-regex.gif)

*Replace this with your GIF showing regex search*

<!-- GIF 3: JSON export demo -->
<!-- Content: repo-searcher search "error" ./project --json results.json -->
<!-- Show: Terminal output + JSON file creation -->

![Export GIF - Placeholder](./assets/demo-export.gif)

*Replace this with your GIF showing JSON export*

---

## Auto-Update

RepoSearcher automatically checks for updates on every run. When a new version is available, you'll see:

```
New version available: v1.0.1 (current: v1.0.0)
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
| `--ignore-case` | Case-insensitive search | `false` |
| `--github` | Search GitHub repositories | `false` |
| `--github-token` | GitHub API token | - |
| `--json <file>` | Export to JSON | - |
| `--csv <file>` | Export to CSV | - |
| `--extensions` | Filter by extensions (`.go,.py`) | all |
| `--context` | Lines of context around matches | `0` |
| `--no-update-check` | Skip automatic update check | `false` |

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
│   └── uninstall.go      # Remove from PATH
├── internal/
│   ├── search/
│   │   ├── engine.go     # Search interface
│   │   ├── local.go      # Local filesystem search
│   │   ├── github.go     # GitHub Codesearch API
│   │   └── matcher.go    # Regex/keyword matching
│   ├── export/
│   │   ├── json.go       # JSON export
│   │   └── csv.go        # CSV export
│   ├── output/
│   │   └── color.go      # Colored terminal output
│   ├── installer/
│   │   └── installer.go  # Platform-specific PATH logic
│   ├── updater/
│   │   └── updater.go    # GitHub release check + download
│   └── utils/
│       └── utils.go      # Helpers
├── pkg/
│   └── models/
│       └── result.go     # SearchResult struct
├── .github/
│   └── workflows/
│       ├── go.yml        # CI/CD
│       └── release.yml   # Auto-release on tag
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
