package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const skillText = `---
name: folio212
description: Trading212 portfolio checker CLI. Use when initializing Trading212 credentials, checking portfolio holdings/positions, exporting holdings as JSON, or troubleshooting common Trading212 permission/rate-limit errors.
homepage: https://github.com/nezdemkovski/folio212
---

# folio212

Use ` + "`folio212`" + ` to check a Trading212 account portfolio (holdings/positions) from the terminal.

Quick start

` + "```bash" + `
folio212 init
folio212 portfolio
folio212 portfolio --json
` + "```" + `

Critical rule (init)

- ` + "`folio212 init`" + ` is interactive and requires Trading212 API key + secret.
- The user must run ` + "`folio212 init`" + ` themselves.
- You (the agent) can explain permissions, where to click in Trading212, and how to fix errors — but do not try to automate secret entry.

Commands

` + "`folio212 skill`" + `

- Prints this skill text (for AI agents / onboarding).
- Usage: ` + "`folio212 skill`" + `

` + "`folio212 init`" + `

- Interactive setup (required before most other commands).
- Collects Trading212 API key + secret from the user and validates access.
- Usage: ` + "`folio212 init`" + `

` + "`folio212 portfolio`" + ` (alias: ` + "`positions`" + `)

- Fetches account summary + open positions and prints holdings.
- Usage:
  - ` + "`folio212 portfolio`" + `
  - ` + "`folio212 positions`" + `
- Flags:
  - ` + "`--json`" + `: output a single JSON object (schema versioned)
  - ` + "`--include-raw`" + `: include raw Trading212 payloads in JSON output (only meaningful with ` + "`--json`" + `)
  - ` + "`--from YYYY-MM-DD`" + ` and ` + "`--to YYYY-MM-DD`" + `: label a reporting period
    - Must provide both; format must be ` + "`YYYY-MM-DD`" + `
    - ` + "`--to`" + ` must be >= ` + "`--from`" + `

` + "`folio212 run`" + `

- Demo operation (TUI-ready). If not attached to a terminal (non-TTY), it prints a plain completion summary.
- Usage: ` + "`folio212 run`" + `

Trading212 API key permissions

- Required: ` + "**Account data**" + `, ` + "**Portfolio**" + `
- Optional (recommended): ` + "**Metadata**" + `

Troubleshooting (common)

- ` + "`403`" + ` on account summary: missing ` + "**Account data**" + ` permission
- ` + "`403`" + ` on positions: missing ` + "**Portfolio**" + ` permission
- ` + "`429`" + `: rate limited; retry in a bit

Example output (plain text)

` + "```text" + `
$ folio212 portfolio
Report date: YYYY-MM-DD
Reporting period: all-time

Investments (as of YYYY-MM-DD, <CURRENCY>)
  holdings value: <AMOUNT>
  pie cash (uninvested): <AMOUNT>
  total allocated to investments: <AMOUNT>

Holdings performance (<CURRENCY>)
  cost basis: <AMOUNT>
  uPnL: <AMOUNT>
  fx impact: <AMOUNT>
  uPnL excl. FX: <AMOUNT>
  return: <PCT>%
  twr (est.): <PCT>%

Account total (as of YYYY-MM-DD, <CURRENCY>)
  free cash: <AMOUNT>
  investments allocated: <AMOUNT>
  account total: <AMOUNT>

Allocation (holdings only, as of YYYY-MM-DD):
  <TICKER>     <PCT>%  (<AMOUNT> <CURRENCY>)

<InstrumentName> (<TICKER>)
  market value: <AMOUNT> <CURRENCY> (<PCT>% of holdings)
  isin: <ISIN> | opened: <RFC3339>
  shares: <QTY> | tradable: <QTY> | in pies: <QTY>
  avg price: <AMOUNT> <CCY> | current price: <AMOUNT> <CCY>
  invested: <AMOUNT> <CURRENCY> | uPnL: <AMOUNT> <CURRENCY>
  fx impact (<FXPAIR>): <AMOUNT> <CURRENCY>
` + "```" + `

Example output (period + JSON)

` + "```text" + `
$ folio212 portfolio --from 2026-01-01 --to 2026-01-31
Report date: YYYY-MM-DD
Reporting period: 2026-01-01 -> 2026-01-31
...
Period flows (executed trades, <CURRENCY>)
  buys: 0.00
  sells: 0.00
  net: 0.00
  Note: This is not implemented yet (requires History - Orders permission).
` + "```" + `

` + "```text" + `
$ folio212 portfolio --from 2026-01-01 --to 2026-01-31 --json
{"schemaVersion":1,"report":{"reportDate":"YYYY-MM-DD","generatedAt":"RFC3339","period":{"from":"2026-01-01","to":"2026-01-31"}},"summary":{"currency":"<CURRENCY>","derived":{"holdingsValue":<N>,"pieCash":<N>,"allocated":<N>,"freeCash":<N>,"accountTotal":<N>,"holdingsCost":<N>,"holdingsPnL":<N>,"holdingsFxImpact":<N>,"holdingsPnLExclFx":<N>,"holdingsReturnPct":<N>,"holdingsReturnBps":<N>,"twrPctEst":<N>,"twrBpsEst":<N>,"twrMethod":"holdings-only-no-flows"},"snapshot":{"apiInvestmentsValue":<N>,"apiCashInPies":<N>,"apiCashAvailable":<N>,"apiCashReserved":<N>,"apiRealizedPnL":<N>,"apiTotalCost":<N>,"apiTotalValue":<N>},"reconcile":{"allocatedDiff":<N>,"accountTotalDiff":<N>}},"allocation":[{"ticker":"<TICKER>","marketValue":<N>,"holdingsPct":<N>,"holdingsBps":<N>}],"holdings":[{"ticker":"<TICKER>","name":"<InstrumentName>","isin":"<ISIN>","openedAt":"RFC3339","qty":<N>,"tradableQty":<N>,"qtyInPies":<N>,"instrumentCurrency":"<CCY>","avgPricePaid":<N>,"currentPrice":<N>,"accountCurrency":"<CURRENCY>","invested":<N>,"marketValue":<N>,"unrealizedPnL":<N>,"fxImpact":<N>,"fxPair":"<FXPAIR>","holdingsPct":<N>,"holdingsBps":<N>}]}
` + "```" + `

Safety

- Prefer demo unless the user explicitly requests live (live = real money account).
`

var skillCmd = &cobra.Command{
	Use:   "skill",
	Short: "Print AI agent skill documentation",
	Long:  "Prints SKILL.md-style documentation for AI agents (OpenClaw-compatible) so agents can use folio212 correctly.",
	Run: func(cmd *cobra.Command, args []string) {
		_, _ = fmt.Fprint(cmd.OutOrStdout(), skillText)
	},
}
