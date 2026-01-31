# CLAUDE.md

This file provides guidance for Claude Code when working with this repository.

## Project Overview

A Trading212 portfolio checker CLI using Cobra for commands and Bubble Tea for terminal UI. Based on clean architecture patterns from `ndev`.

**Module**: `github.com/nezdemkovski/folio212`
**Go version**: 1.25.5

## Architecture

```
folio212/
├── cmd/                          # Thin orchestration layer (Cobra commands)
│   ├── root.go                   # Global flags, config loading via PersistentPreRunE
│   ├── init.go                   # Launches init TUI (orchestration only)
│   ├── portfolio.go              # Portfolio command orchestration (~60 lines)
│   └── skill.go                  # Skill command
├── internal/
│   ├── domain/                   # Pure business logic (NO UI dependencies)
│   │   └── portfolio/            # Portfolio domain
│   │       ├── types.go          # Data structures (JSON schema)
│   │       ├── calculator.go     # Pure math functions (testable)
│   │       ├── service.go        # Business orchestration
│   │       └── errors.go         # Domain error types
│   ├── infrastructure/           # External systems
│   │   ├── config/config.go      # Viper + YAML persistence
│   │   ├── secrets/secrets.go    # 3-tier: env → keyring → file fallback
│   │   └── trading212/           # API client
│   │       ├── client.go         # HTTP client
│   │       ├── types.go          # API response types
│   │       └── errors.go         # HTTP error handling
│   ├── presentation/             # UI layer (all user-facing output)
│   │   ├── init.go               # Init command TUI (Bubble Tea)
│   │   ├── portfolio.go          # Portfolio text rendering
│   │   └── errors.go             # Error humanization (UI messages)
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

1. **cmd/** - **Orchestration only** (~60-100 lines per command)
   - Parse flags and arguments
   - Load config and secrets
   - Delegate to domain layer for business logic
   - Delegate to presentation layer for UI rendering
   - **NO business logic, NO calculations, NO data transformations**

2. **domain/** - **Pure business logic** (NO UI dependencies)
   - **types.go**: Data structures and JSON schemas
   - **calculator.go**: Pure functions (math, aggregations) - easily testable
   - **service.go**: Business orchestration, reconciliations, workflows
   - **errors.go**: Domain-specific error types
   - **NO Bubble Tea imports, NO fmt.Print, NO direct infrastructure calls**

3. **presentation/** - **All user-facing output**
   - Bubble Tea models and TUI components
   - Text rendering and formatting
   - Error humanization (converting errors to user-friendly messages)
   - Uses `shared/ui` for styling
   - **NO business logic, delegates to domain layer**

4. **infrastructure/** - **External systems only**
   - Config persistence (Viper + YAML)
   - Secrets management (OS keyring + file fallback)
   - API clients (HTTP clients with timeout and retry logic)
   - **NO business logic, NO UI code**

5. **shared/** - **Zero dependencies on other internal packages**
   - Constants, validation functions
   - UI styling primitives (Lipgloss)
   - Generic utilities
   - **Only depends on stdlib and external libraries**

### Dependency Direction

```
cmd → domain → infrastructure → shared
  ↓      ↘
presentation
  ↓
shared
```

**Rules:**
- `cmd` imports: `domain`, `presentation`, `infrastructure`
- `presentation` imports: `domain`, `shared`
- `domain` imports: `infrastructure` (for API clients), `shared`
- `infrastructure` imports: `shared`
- `shared` imports: **nothing internal**

## Build & Run

```bash
go build -o folio212 .
./folio212 --help
./folio212 init      # Interactive setup (TUI)
./folio212 portfolio # Show current holdings
```

## Secrets Management (3-tier)

Priority order for **retrieval**:

1. Environment variables: `FOLIO212_<KEY>` (e.g., `FOLIO212_API_TOKEN`)
2. OS keyring (macOS Keychain, Windows Credential Manager, Linux Secret Service)
3. Config file fallback: `~/.folio212/secrets.yml` (insecure, with warnings)

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

- Done: `•` Active: `→` Running: `◉` Pending: `○` Warning: `!` Error: `✗`

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
2. Rename `rootCmd.Use` in `cmd/root.go` (currently `folio212`)
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
