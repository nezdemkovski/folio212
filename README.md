# folio212

A Trading212 portfolio checker CLI (TUI-ready) based on the architecture and terminal styling patterns from `ndev`.

## What you get

- **Cobra** command layer (`cmd/`) that stays intentionally thin.
- **Clean layers** under `internal/`:
  - `internal/domain/` business logic (no UI dependencies)
  - `internal/presentation/` Bubble Tea models + rendering
  - `internal/infrastructure/` config persistence (Viper + YAML) + secrets (OS keyring)
  - `internal/shared/` constants, validation, and centralized `ui` styling

## Quick start

```bash
go run . --help
go run . init
go run . portfolio
```

## Install

### Homebrew

If you use Homebrew, you can install `folio212` from your tap:

```bash
brew tap nezdemkovski/homebrew-tap
brew install folio212
```

Upgrade:

```bash
brew upgrade folio212
```

## Trading212 API key permissions

When creating the Trading212 API key, you must enable at least:

- **Account data**: used by `folio212 init` to validate credentials via `GET /equity/account/summary`
- **Portfolio**: used by `folio212 portfolio` / `folio212 positions` via `GET /equity/positions`

Optional (recommended if you want richer instrument info):

- **Metadata**: used to fetch instrument metadata (type: stock vs ETF, names, etc.) via `GET /equity/metadata/instruments`

## Generate Trading212 API keys

The Trading212 Public API uses an **API Key + API Secret** (HTTP Basic auth). The secret is shown **only once** when you create the key, so store it safely.

Steps (web or mobile):

1. Open Trading212 → **Settings** → **API (Beta)**
2. Accept the risk warning
3. Tap **Generate API key**
4. Choose permissions (at minimum: **Account data** + **Portfolio**)
5. Save:
   - **API Key** (we store the key id in `config.yaml`)
   - **API Secret** (we store it in your OS keyring)

Official guide: [Trading 212 API key](https://helpcentre.trading212.com/hc/en-us/articles/14584770928157-Trading-212-API-key)

## Secrets management (3-tier storage)

Sensitive data (like API tokens) should **never** be stored in plain YAML config files. This template includes `internal/infrastructure/secrets` with a **3-tier fallback strategy** inspired by [GitHub CLI](https://github.com/cli/cli):

### Storage priority (Get)

When retrieving secrets, the following order is used:

1. **Environment variables** (highest priority)
   - Format: `FOLIO212_<KEY>` (e.g., `FOLIO212_API_TOKEN`)
   - Works everywhere (desktop, Docker, CI/CD)
   - Explicit override mechanism
2. **OS keyring** (secure desktop storage)
   - **macOS**: Keychain
   - **Windows**: Credential Manager
   - **Linux**: Secret Service (GNOME Keyring)
   - 3-second timeout to prevent hanging
3. **Config file** (insecure fallback for headless environments)
   - Stored in `~/.folio212/secrets.yml` with `0600` permissions
   - Used when keyring is unavailable (Docker, headless servers)
   - A warning is shown when this fallback is used

### Storage priority (Set)

When storing secrets:

1. **OS keyring** (attempted first on desktop environments)
2. **Config file fallback** (used when keyring unavailable/times out)
   - Returns `insecure=true` flag when this happens
   - UI shows warning: "⚠ API token stored in config file (insecure)"

### Example usage

```go
import "github.com/nezdemkovski/folio212/internal/infrastructure/secrets"

// Store a secret (returns source + insecure flag)
source, insecure, err := secrets.Set(secrets.KeyAPIToken, "sk-abc123")
if err != nil {
    // handle error
}
if insecure {
    fmt.Println("Warning: Using insecure file storage (keyring unavailable)")
}

// Retrieve a secret (returns value + source + error)
token, source, err := secrets.Get(secrets.KeyAPIToken)
if err != nil {
    // handle error
}
fmt.Printf("Token loaded from: %s\n", source) // "environment", "keyring", "config_file"

// Delete a secret (removes from all locations)
if err := secrets.Delete(secrets.KeyAPIToken); err != nil {
    // handle error
}
```

### Server/Docker usage

For headless environments where the OS keyring is unavailable:

```bash
# Preferred: Use environment variables
export FOLIO212_API_TOKEN="your-token-here"
./folio212 portfolio

# Alternative: Let it fall back to file storage (insecure)
./folio212 init  # Will warn about insecure storage
```

### Why this matters

- **Desktop**: Secrets stored securely in OS keyring (protected by system authentication)
- **Servers/Docker**: Explicit environment variables (standard practice in 2026)
- **Emergency fallback**: File storage ensures CLI works everywhere, with clear warnings

## Homebrew publishing (optional)

This template supports publishing a formula to a tap repo via GoReleaser (`brews:` in `.goreleaser.yaml`).

- Create a tap repository: `homebrew-tap`
- Add a GitHub Actions secret: `RELEASE_GITHUB_TOKEN` (PAT with repo access to the tap repo)
- Update `.goreleaser.yaml` placeholders:
  - `brews[0].name`
  - `brews[0].repository.owner` / `brews[0].repository.name`
  - `homepage`, `description`, `license`

## Adding a new command (pattern)

1. Add `cmd/mycmd.go` (argument parsing + orchestration)
2. Add `internal/domain/myfeature/` (business logic + result structs)
3. Add `internal/presentation/mycmd.go` (Bubble Tea model + `Render*Completion`)

No business logic should live in `cmd/` or `internal/presentation/`.
