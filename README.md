# folio212

Check your Trading212 portfolio from the terminal.

## What it does

- **Quick portfolio overview** - See all your holdings, allocations, and performance in one view
- **Secure by default** - API secrets stored in your OS keyring (macOS Keychain, Windows Credential Manager, Linux Secret Service)
- **JSON export** - Get your data in JSON format for further analysis
- **Works everywhere** - Desktop, servers, Docker with environment variable fallback

## Quick start

```bash
# Install
go install github.com/nezdemkovski/folio212@latest

# Or use Homebrew
brew tap nezdemkovski/homebrew-tap
brew install folio212

# Setup (interactive)
folio212 init

# Check your portfolio
folio212 portfolio

# Export as JSON
folio212 portfolio --json
```

## Install

### Homebrew

```bash
brew tap nezdemkovski/homebrew-tap
brew install folio212
```

Upgrade: `brew upgrade folio212`

### From source

```bash
go install github.com/nezdemkovski/folio212@latest
```

## Setup

You need a Trading212 API key to use this tool.

### 1. Generate API keys

1. Open Trading212 → **Settings** → **API (Beta)**
2. Accept the risk warning
3. Tap **Generate API key**
4. Enable permissions: **Account data** + **Portfolio**
5. Save both the **API Key** and **API Secret** (secret is shown only once!)

[Official guide](https://helpcentre.trading212.com/hc/en-us/articles/14584770928157-Trading-212-API-key)

### 2. Configure folio212

Run the interactive setup:

```bash
folio212 init
```

This will:
- Ask for your API key and secret
- Validate the credentials
- Store config in `~/.folio212/config.yaml`
- Store secret securely in your OS keyring

## Usage

### Check portfolio

```bash
folio212 portfolio
```

Output includes:
- Holdings value and pie cash
- Performance metrics (return, PnL, FX impact)
- Account total breakdown
- Allocation percentages
- Individual position details

### JSON export

```bash
folio212 portfolio --json
folio212 portfolio --json --include-raw  # Include raw API data
```

### Period filtering

```bash
folio212 portfolio --from 2024-01-01 --to 2024-12-31
```

## Security

### How secrets are stored

Your API secret is stored using a 3-tier approach:

1. **OS keyring** (default on desktop)
   - macOS: Keychain
   - Windows: Credential Manager
   - Linux: Secret Service

2. **Environment variables** (for servers/Docker)
   ```bash
   export FOLIO212_T212_API_SECRET="your-secret"
   ```

3. **Config file** (fallback only)
   - Stored in `~/.folio212/secrets.yml` with `0600` permissions
   - Shows warning when used

### Server/Docker usage

For headless environments without a keyring:

```bash
# Recommended: Use environment variable
export FOLIO212_T212_API_SECRET="your-secret"
folio212 portfolio

# Or use the --env flag pattern
FOLIO212_T212_API_SECRET="your-secret" folio212 portfolio
```

## API Permissions Required

- **Account data**: Required for `folio212 init` to validate credentials
- **Portfolio**: Required for `folio212 portfolio` to fetch positions
- **Metadata** (optional): For richer instrument information

## For Developers

See [CLAUDE.md](CLAUDE.md) for:
- Architecture overview
- Code organization
- Adding new commands
- Layer rules and dependency directions

## License

MIT
