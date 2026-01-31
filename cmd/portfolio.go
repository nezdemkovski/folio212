package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/nezdemkovski/folio212/internal/infrastructure/secrets"
	"github.com/nezdemkovski/folio212/internal/infrastructure/trading212"
	"github.com/spf13/cobra"
)

const portfolioSchemaVersion = 1

type periodRange struct {
	From *string `json:"from"` // YYYY-MM-DD or null
	To   *string `json:"to"`   // YYYY-MM-DD or null
}

type portfolioReport struct {
	ReportDate  string      `json:"reportDate"`  // YYYY-MM-DD (local)
	GeneratedAt string      `json:"generatedAt"` // RFC3339 (local time, with timezone)
	Period      periodRange `json:"period"`
}

type portfolioDerived struct {
	// Explicit capital buckets (to avoid "missing money" confusion).
	HoldingsValue float64 `json:"holdingsValue"` // executed holdings only (market value)
	PieCash       float64 `json:"pieCash"`       // cash inside pies, not yet invested
	Allocated     float64 `json:"allocated"`     // holdingsValue + pieCash
	FreeCash      float64 `json:"freeCash"`      // availableToTrade + reservedForOrders
	AccountTotal  float64 `json:"accountTotal"`  // should equal freeCash + allocated

	// Holdings-only performance (executed positions only).
	HoldingsCost      float64  `json:"holdingsCost"` // cost basis for executed holdings only
	HoldingsPnL       float64  `json:"holdingsPnL"`  // unrealized PnL for executed holdings only
	HoldingsFXImpact  *float64 `json:"holdingsFxImpact,omitempty"`
	HoldingsPnLExclFX *float64 `json:"holdingsPnLExclFx,omitempty"`

	HoldingsReturnPct float64 `json:"holdingsReturnPct"` // rounded
	HoldingsReturnBps int     `json:"holdingsReturnBps"`
	TWRPctEst         float64 `json:"twrPctEst"` // rounded
	TWRBpsEst         int     `json:"twrBpsEst"`
	TWRMethod         string  `json:"twrMethod"`
	TWRDescription    string  `json:"twrDescription,omitempty"`
}

type portfolioSnapshot struct {
	APIInvestmentsValue float64 `json:"apiInvestmentsValue"`
	APICashInPies       float64 `json:"apiCashInPies"`
	APICashAvailable    float64 `json:"apiCashAvailable"`
	APICashReserved     float64 `json:"apiCashReserved"`
	APIRealizedPnL      float64 `json:"apiRealizedPnL"`
	APITotalCost        float64 `json:"apiTotalCost"`
	APITotalValue       float64 `json:"apiTotalValue"`
}

type portfolioReconcile struct {
	AllocatedDiff    float64  `json:"allocatedDiff"`
	AccountTotalDiff float64  `json:"accountTotalDiff"`
	Warnings         []string `json:"warnings,omitempty"`
}

type portfolioSummary struct {
	Currency  string             `json:"currency"`
	Derived   portfolioDerived   `json:"derived"`
	Snapshot  portfolioSnapshot  `json:"snapshot"`
	Reconcile portfolioReconcile `json:"reconcile"`
}

type portfolioAllocationRow struct {
	Ticker      string  `json:"ticker"`
	MarketValue float64 `json:"marketValue"`
	HoldingsPct float64 `json:"holdingsPct"`
	HoldingsBps int     `json:"holdingsBps"`
}

type portfolioHoldingRow struct {
	Ticker      string  `json:"ticker"`
	Name        string  `json:"name"`
	ISIN        string  `json:"isin,omitempty"`
	OpenedAt    string  `json:"openedAt,omitempty"`
	Qty         float64 `json:"qty"`
	TradableQty float64 `json:"tradableQty"`
	QtyInPies   float64 `json:"qtyInPies"`

	InstrumentCurrency string  `json:"instrumentCurrency"`
	AvgPricePaid       float64 `json:"avgPricePaid"`
	CurrentPrice       float64 `json:"currentPrice"`

	AccountCurrency string   `json:"accountCurrency"`
	Invested        float64  `json:"invested"`
	MarketValue     float64  `json:"marketValue"`
	UnrealizedPnL   float64  `json:"unrealizedPnL"`
	FXImpact        *float64 `json:"fxImpact,omitempty"`
	FXPair          string   `json:"fxPair,omitempty"` // e.g. "USD/EUR"
	HoldingsPct     float64  `json:"holdingsPct"`
	HoldingsBps     int      `json:"holdingsBps"`
}

type portfolioJSON struct {
	SchemaVersion int                      `json:"schemaVersion"`
	Report        portfolioReport          `json:"report"`
	Summary       portfolioSummary         `json:"summary"`
	Allocation    []portfolioAllocationRow `json:"allocation"`
	Holdings      []portfolioHoldingRow    `json:"holdings"`
	Raw           *struct {
		AccountSummary *trading212.AccountSummary `json:"accountSummary,omitempty"`
		Positions      []trading212.Position      `json:"positions"`
	} `json:"raw,omitempty"`
}

var portfolioCmd = &cobra.Command{
	Use:     "portfolio",
	Aliases: []string{"positions"},
	Short:   "Show current holdings",
	Long:    "Fetches open positions from Trading212 and prints holdings.",
	RunE: func(cmd *cobra.Command, args []string) error {
		asJSON, _ := cmd.Flags().GetBool("json")
		includeRaw, _ := cmd.Flags().GetBool("include-raw")
		fromStr, _ := cmd.Flags().GetString("from")
		toStr, _ := cmd.Flags().GetString("to")

		cfg := GetConfig()
		if cfg == nil {
			return fmt.Errorf("config not loaded; please run 'folio212 init' first")
		}
		if strings.TrimSpace(cfg.Trading212APIKey) == "" {
			return fmt.Errorf("missing trading212 api key; please run 'folio212 init'")
		}

		secret, _, err := secrets.Get(secrets.KeyTrading212APISecret)
		if err != nil {
			return err
		}
		secret = strings.TrimSpace(secret)
		if secret == "" {
			return fmt.Errorf("missing trading212 api secret; please run 'folio212 init'")
		}

		baseURL := trading212.BaseURLDemo
		if strings.EqualFold(strings.TrimSpace(cfg.Trading212Env), "live") {
			baseURL = trading212.BaseURLLive
		}

		periodLabel, period, err := formatPeriod(fromStr, toStr)
		if err != nil {
			return err
		}

		client, err := trading212.NewClient(baseURL, cfg.Trading212APIKey, secret)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		summary, err := client.GetAccountSummary(ctx)
		if err != nil {
			return humanizeAccountDataError(err)
		}

		positions, err := client.GetPositions(ctx, "")
		if err != nil {
			return humanizePortfolioError(err)
		}

		now := time.Now()
		reportDate := now.Format("2006-01-02")
		generatedAt := now.Format(time.RFC3339)
		holdingsValue := sumPositionsValue(positions)
		holdingsCost := sumPositionsCost(positions)
		holdingsPnL := sumPositionsPnL(positions)
		fxImpactSum, fxImpactOK := sumPositionsFXImpact(positions)

		pieCash := summary.Cash.InPies
		allocated := holdingsValue + pieCash
		freeCash := summary.Cash.AvailableToTrade + summary.Cash.ReservedForOrders
		accountTotal := summary.TotalValue

		warnings := make([]string, 0, 3)
		var warnInvestmentsAlloc string
		var warnAccountTotal string

		// Reconcile: account total should equal free cash + investments allocated.
		if diff := accountTotal - (freeCash + allocated); abs(diff) > 0.01 {
			warnAccountTotal = fmt.Sprintf("WARNING: account total does not reconcile (diff: %.2f %s)", diff, summary.Currency)
			warnings = append(warnings, warnAccountTotal)
		}
		// Sanity: API investments.currentValue should match allocated investments (holdings + pie cash).
		if diff := summary.Investments.CurrentValue - allocated; abs(diff) > 0.01 {
			warnInvestmentsAlloc = fmt.Sprintf("WARNING: investments allocated does not reconcile (diff: %.2f %s)", diff, summary.Currency)
			warnings = append(warnings, warnInvestmentsAlloc)
		}

		holdingsReturn := 0.0
		if holdingsCost > 0 {
			holdingsReturn = (holdingsPnL / holdingsCost) * 100
		}
		// TWR approximation: without cash-flow history, we approximate with holdings-only return.
		twrPct := holdingsReturn

		var holdingsFXImpact *float64
		var holdingsPnLExclFX *float64
		if fxImpactOK {
			v := fxImpactSum
			holdingsFXImpact = &v
			ex := holdingsPnL - fxImpactSum
			holdingsPnLExclFX = &ex
		}

		allocation := make([]portfolioAllocationRow, 0, len(positions))
		holdings := make([]portfolioHoldingRow, 0, len(positions))
		for _, p := range positions {
			pct := 0.0
			// Allocation percentages are based on holdings only (exclude pie cash).
			if holdingsValue > 0 {
				pct = (p.CurrentValue() / holdingsValue) * 100
			}

			fxPair := ""
			if p.Instrument.Currency != "" && summary.Currency != "" && p.Instrument.Currency != summary.Currency {
				fxPair = p.Instrument.Currency + "/" + summary.Currency
			}

			allocation = append(allocation, portfolioAllocationRow{
				Ticker:      p.Instrument.Ticker,
				MarketValue: p.CurrentValue(),
				HoldingsPct: round(pct, 2),
				HoldingsBps: pctToBps(pct),
			})

			opened := ""
			if !p.CreatedAt.IsZero() {
				opened = p.CreatedAt.Format(time.RFC3339)
			}

			holdings = append(holdings, portfolioHoldingRow{
				Ticker:      p.Instrument.Ticker,
				Name:        p.Instrument.Name,
				ISIN:        p.Instrument.ISIN,
				OpenedAt:    opened,
				Qty:         p.Quantity,
				TradableQty: p.QuantityAvailableForTrading,
				QtyInPies:   p.QuantityInPies,

				InstrumentCurrency: p.Instrument.Currency,
				AvgPricePaid:       p.AveragePricePaid,
				CurrentPrice:       p.CurrentPrice,

				AccountCurrency: summary.Currency,
				Invested:        p.Invested(),
				MarketValue:     p.CurrentValue(),
				UnrealizedPnL:   p.WalletImpact.UnrealizedProfitLoss,
				FXImpact:        p.WalletImpact.FXImpact,
				FXPair:          fxPair,
				HoldingsPct:     round(pct, 2),
				HoldingsBps:     pctToBps(pct),
			})
		}

		sort.SliceStable(allocation, func(i, j int) bool {
			return allocation[i].MarketValue > allocation[j].MarketValue
		})
		sort.SliceStable(holdings, func(i, j int) bool {
			return holdings[i].MarketValue > holdings[j].MarketValue
		})

		if asJSON {
			enc := json.NewEncoder(os.Stdout)
			var out portfolioJSON
			out.SchemaVersion = portfolioSchemaVersion
			out.Report = portfolioReport{
				ReportDate:  reportDate,
				GeneratedAt: generatedAt,
				Period:      period,
			}
			out.Summary = portfolioSummary{
				Currency: summary.Currency,
				Derived: portfolioDerived{
					HoldingsValue:     holdingsValue,
					PieCash:           pieCash,
					Allocated:         allocated,
					FreeCash:          freeCash,
					AccountTotal:      accountTotal,
					HoldingsCost:      holdingsCost,
					HoldingsPnL:       holdingsPnL,
					HoldingsFXImpact:  holdingsFXImpact,
					HoldingsPnLExclFX: holdingsPnLExclFX,
					HoldingsReturnPct: round(holdingsReturn, 4),
					HoldingsReturnBps: pctToBps(holdingsReturn),
					TWRPctEst:         round(twrPct, 4),
					TWRBpsEst:         pctToBps(twrPct),
					TWRMethod:         "holdings-only-no-flows",
					TWRDescription:    "Estimated TWR based on holdings only; excludes cash flows and pie allocations.",
				},
				Snapshot: portfolioSnapshot{
					APIInvestmentsValue: summary.Investments.CurrentValue,
					APICashInPies:       summary.Cash.InPies,
					APICashAvailable:    summary.Cash.AvailableToTrade,
					APICashReserved:     summary.Cash.ReservedForOrders,
					APIRealizedPnL:      summary.Investments.RealizedProfitLoss,
					APITotalCost:        summary.Investments.TotalCost,
					APITotalValue:       summary.TotalValue,
				},
				Reconcile: portfolioReconcile{
					AllocatedDiff:    round(summary.Investments.CurrentValue-allocated, 2),
					AccountTotalDiff: round(summary.TotalValue-(freeCash+allocated), 2),
					Warnings:         warnings,
				},
			}
			out.Allocation = allocation
			out.Holdings = holdings
			if includeRaw {
				out.Raw = &struct {
					AccountSummary *trading212.AccountSummary `json:"accountSummary,omitempty"`
					Positions      []trading212.Position      `json:"positions"`
				}{
					AccountSummary: summary,
					Positions:      positions,
				}
			}
			return enc.Encode(out)
		}

		fmt.Printf("Report date: %s\n", reportDate)
		fmt.Printf("Reporting period: %s\n", periodLabel)
		if periodLabel != "all-time" {
			fmt.Println("Note: Holdings metrics reflect executed positions; pie cash is a snapshot at period end (uninvested).")
		}
		fmt.Println()

		fmt.Printf("Investments (as of %s, %s)\n", reportDate, summary.Currency)
		fmt.Printf("  holdings value: %.2f\n", holdingsValue)
		fmt.Printf("  pie cash (uninvested): %.2f\n", pieCash)
		fmt.Printf("  total allocated to investments: %.2f\n\n", allocated)
		if warnInvestmentsAlloc != "" {
			fmt.Printf("%s\n\n", warnInvestmentsAlloc)
		}

		fmt.Printf("Holdings performance (%s)\n", summary.Currency)
		fmt.Printf("  cost basis: %.2f\n", holdingsCost)
		fmt.Printf("  uPnL: %.2f\n", holdingsPnL)
		if holdingsFXImpact != nil && holdingsPnLExclFX != nil {
			fmt.Printf("  fx impact: %.2f\n", *holdingsFXImpact)
			fmt.Printf("  uPnL excl. FX: %.2f\n", *holdingsPnLExclFX)
		} else {
			fmt.Printf("  fx impact: n/a\n")
		}
		fmt.Printf("  return: %.2f%%\n", holdingsReturn)
		fmt.Printf("  twr (est.): %.2f%%\n\n", twrPct)

		fmt.Printf("Account total (as of %s, %s)\n", reportDate, summary.Currency)
		fmt.Printf("  free cash: %.2f\n", freeCash)
		fmt.Printf("  investments allocated: %.2f\n", allocated)
		fmt.Printf("  account total: %.2f\n", accountTotal)
		if warnAccountTotal != "" {
			fmt.Printf("  %s\n", warnAccountTotal)
		}
		fmt.Println()

		fmt.Printf("Allocation (holdings only, as of %s):\n", reportDate)
		if holdingsValue <= 0 {
			fmt.Println("  n/a (no holdings)")
		} else {
			for _, row := range allocation {
				fmt.Printf("  %-10s %7.2f%%  (%.2f %s)\n", row.Ticker, row.HoldingsPct, row.MarketValue, summary.Currency)
			}
		}
		fmt.Println()

		if periodLabel != "all-time" {
			fmt.Printf("Period flows (executed trades, %s)\n", summary.Currency)
			fmt.Printf("  buys: 0.00\n")
			fmt.Printf("  sells: 0.00\n")
			fmt.Printf("  net: 0.00\n")
			fmt.Printf("  Note: This is not implemented yet (requires History - Orders permission).\n\n")
		}

		if len(positions) == 0 {
			fmt.Println("No open positions.")
			return nil
		}

		for _, h := range holdings {
			fxImpactStr := "n/a"
			if h.FXImpact != nil {
				fxImpactStr = fmt.Sprintf("%.2f", *h.FXImpact)
			}
			fmt.Printf(
				"%s (%s)\n  market value: %.2f %s (%.2f%% of holdings)\n  isin: %s | opened: %s\n  shares: %.6g | tradable: %.6g | in pies: %.6g\n  avg price: %.6g %s | current price: %.6g %s\n  invested: %.2f %s | uPnL: %.2f %s\n  fx impact (%s): %s %s\n\n",
				h.Name,
				h.Ticker,
				h.MarketValue,
				h.AccountCurrency,
				h.HoldingsPct,
				h.ISIN,
				h.OpenedAt,
				h.Qty,
				h.TradableQty,
				h.QtyInPies,
				h.AvgPricePaid,
				h.InstrumentCurrency,
				h.CurrentPrice,
				h.InstrumentCurrency,
				h.Invested,
				h.AccountCurrency,
				h.UnrealizedPnL,
				h.AccountCurrency,
				chooseFXPair(h.FXPair, h.InstrumentCurrency, h.AccountCurrency),
				fxImpactStr,
				h.AccountCurrency,
			)
		}

		return nil
	},
}

func formatPeriod(fromStr, toStr string) (label string, period periodRange, err error) {
	fromStr = strings.TrimSpace(fromStr)
	toStr = strings.TrimSpace(toStr)
	period = periodRange{From: nil, To: nil}

	if fromStr == "" && toStr == "" {
		return "all-time", period, nil
	}
	if fromStr == "" || toStr == "" {
		return "", period, fmt.Errorf("reporting period requires both --from and --to (YYYY-MM-DD)")
	}
	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		return "", period, fmt.Errorf("invalid --from (expected YYYY-MM-DD): %w", err)
	}
	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		return "", period, fmt.Errorf("invalid --to (expected YYYY-MM-DD): %w", err)
	}
	if to.Before(from) {
		return "", period, fmt.Errorf("invalid period: --to is before --from")
	}
	period.From = &fromStr
	period.To = &toStr
	return fromStr + " -> " + toStr, period, nil
}

func sumPositionsValue(positions []trading212.Position) float64 {
	var sum float64
	for _, p := range positions {
		sum += p.CurrentValue()
	}
	return sum
}

func sumPositionsCost(positions []trading212.Position) float64 {
	var sum float64
	for _, p := range positions {
		sum += p.Invested()
	}
	return sum
}

func sumPositionsPnL(positions []trading212.Position) float64 {
	var sum float64
	for _, p := range positions {
		sum += p.WalletImpact.UnrealizedProfitLoss
	}
	return sum
}

func sumPositionsFXImpact(positions []trading212.Position) (sum float64, ok bool) {
	ok = true
	for _, p := range positions {
		if p.WalletImpact.FXImpact == nil {
			ok = false
			continue
		}
		sum += *p.WalletImpact.FXImpact
	}
	return sum, ok
}

func chooseFXPair(explicit, instrumentCurrency, accountCurrency string) string {
	if explicit != "" {
		return explicit
	}
	if instrumentCurrency == "" || accountCurrency == "" || instrumentCurrency == accountCurrency {
		return "n/a"
	}
	return instrumentCurrency + "/" + accountCurrency
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func round(x float64, places int) float64 {
	p := math.Pow(10, float64(places))
	return math.Round(x*p) / p
}

// pctToBps converts a percent value (e.g. 0.8712%) to integer basis points (87).
// For allocation percentages, 100.00% becomes 10000 bps.
func pctToBps(pct float64) int {
	return int(math.Round(pct * 100))
}

func humanizeAccountDataError(err error) error {
	var httpErr *trading212.HTTPError
	if errors.As(err, &httpErr) && httpErr.StatusCode == 403 {
		return fmt.Errorf("%w (missing permission: enable \"Account data\" for your Trading212 API key)", err)
	}
	if errors.As(err, &httpErr) && httpErr.StatusCode == 429 {
		if d, ok := httpErr.SuggestedRetryDelay(time.Now()); ok {
			secs := int(d.Round(time.Second).Seconds())
			if secs < 1 {
				secs = 1
			}
			return fmt.Errorf("%w (rate limited: try again in ~%ds)", err, secs)
		}
		return fmt.Errorf("%w (rate limited: try again in a few seconds)", err)
	}
	return err
}

func humanizePortfolioError(err error) error {
	var httpErr *trading212.HTTPError
	if errors.As(err, &httpErr) && httpErr.StatusCode == 403 {
		return fmt.Errorf("%w (missing permission: enable \"Portfolio\" for your Trading212 API key)", err)
	}
	return err
}

func init() {
	portfolioCmd.Flags().Bool("json", false, "Output raw JSON")
	portfolioCmd.Flags().Bool("include-raw", false, "Include raw API payloads in JSON output")
	portfolioCmd.Flags().String("from", "", "Reporting period start (YYYY-MM-DD)")
	portfolioCmd.Flags().String("to", "", "Reporting period end (YYYY-MM-DD)")
}
