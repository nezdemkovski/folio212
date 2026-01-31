# CLAUDE.md

This file provides guidance for Claude Code when working with this repository.

## Project Overview

A minimal, TUI-ready Go CLI template using Cobra for commands and Bubble Tea for terminal UI. Based on clean architecture patterns from `ndev`.

**Module**: `github.com/nezdemkovski/cli-tool-template`
**Go version**: 1.25.5

## Architecture

```
cli-tool-template/
├── cmd/                          # Thin orchestration layer (Cobra commands)
│   ├── root.go                   # Global flags, config loading via PersistentPreRunE
│   ├── init.go                   # Launches init TUI
│   └── run.go                    # TTY detection, launches run TUI
├── internal/
│   ├── domain/                   # Pure business logic (NO UI dependencies)
│   │   └── run/manager.go        # Context-aware operations, returns structured results
│   ├── infrastructure/           # External systems
│   │   ├── config/config.go      # Viper + YAML persistence
│   │   └── secrets/secrets.go    # 3-tier: env → keyring → file fallback
│   ├── presentation/             # UI layer (Bubble Tea models)
│   │   ├── init.go               # Init command TUI
│   │   └── run.go                # Run command TUI
│   └── shared/                   # Pure utilities (no internal dependencies)
│       ├── constants/app.go      # App name, config dir
│       ├── ui/                   # Centralized Lipgloss styling
│       │   ├── style.go          # All style definitions
│       │   ├── layout.go         # Layout helpers
│       │   ├── logo.go           # ASCII logo
│       │   └── ui.go             # Print helpers
│       └── validation/           # Generic validation
└── main.go                       # Entry point only
```

### Layer Rules (Strict)

1. **cmd/** - Orchestration only. No business logic, no UI rendering, no direct styling.
2. **domain/** - Pure business logic. No Bubble Tea imports, no UI dependencies.
3. **presentation/** - UI rendering. Delegates to domain for logic, uses `shared/ui` for styling.
4. **infrastructure/** - External systems only (config files, keyring, network).
5. **shared/** - Zero dependencies on other internal packages.

### Dependency Direction

```
cmd → presentation → domain → infrastructure → shared
                ↘            ↗
                  infrastructure
```

## Build & Run

```bash
go build -o app .
./app --help
./app init      # Interactive setup (TUI)
./app run       # Main operation (TUI)
```

## Secrets Management (3-tier)

Priority order for **retrieval**:
1. Environment variables: `APP_<KEY>` (e.g., `APP_API_TOKEN`)
2. OS keyring (macOS Keychain, Windows Credential Manager, Linux Secret Service)
3. Config file fallback: `~/.cli-tool-template/secrets.yml` (insecure, with warnings)

Keyring operations have a 3-second timeout to prevent hanging.

## UI Theme (Professional)

**Colors** (ANSI 256, muted):
- Primary: 111 (soft cyan) - titles, active focus
- Secondary: 250 (light gray) - labels, headers
- Neutral: 245 (muted gray) - metadata
- Accent: 109 (soft green) - success
- Warning: 180 (soft amber)
- Error: 167 (soft red)
- Highlight: 147 (soft purple) - selection

**Status symbols** (no emojis):
- Done: `•`  Active: `→`  Running: `◉`  Pending: `○`  Warning: `!`  Error: `✗`

**Rules**:
- No emojis in UI output
- All styles centralized in `internal/shared/ui/style.go`
- No inline styling in rendering code
- Tables use Charm components (bubbles/table or lipgloss.Table), not manual padding

## Adding a New Command

1. Add `cmd/mycmd.go` - argument parsing + orchestration
2. Add `internal/domain/myfeature/` - business logic + result structs
3. Add `internal/presentation/mycmd.go` - Bubble Tea model + `Render*Completion`

## Renaming for New Project

1. Update module path in `go.mod`
2. Rename `rootCmd.Use` in `cmd/root.go` (currently `app`)
3. Update `internal/shared/constants` (app name + config dir)
4. Update `.goreleaser.yaml` and `.github/workflows/release.yml`

## Key Dependencies

- **github.com/spf13/cobra** - Command framework
- **github.com/charmbracelet/bubbletea** - TUI framework
- **github.com/charmbracelet/bubbles** - TUI components
- **github.com/charmbracelet/huh** - Form components
- **github.com/charmbracelet/lipgloss** - Styling
- **github.com/spf13/viper** - Config management
- **github.com/zalando/go-keyring** - OS keyring access
