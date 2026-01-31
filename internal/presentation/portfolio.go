package presentation

import (
	"fmt"
	"io"
	"strings"

	"github.com/nezdemkovski/folio212/internal/domain/portfolio"
)

func RenderPortfolioText(output *portfolio.Output, w io.Writer) error {
	var s strings.Builder

	s.WriteString(fmt.Sprintf("Report date: %s\n", output.Report.ReportDate))
	s.WriteString(fmt.Sprintf("Reporting period: %s\n", formatPeriodLabel(output.Report.Period)))
	if !isAllTime(output.Report.Period) {
		s.WriteString("Note: Holdings metrics reflect executed positions; pie cash is a snapshot at period end (uninvested).\n")
	}
	s.WriteString("\n")

	s.WriteString(fmt.Sprintf("Investments (as of %s, %s)\n", output.Report.ReportDate, output.Summary.Currency))
	s.WriteString(fmt.Sprintf("  holdings value: %.2f\n", output.Summary.Derived.HoldingsValue))
	s.WriteString(fmt.Sprintf("  pie cash (uninvested): %.2f\n", output.Summary.Derived.PieCash))
	s.WriteString(fmt.Sprintf("  total allocated to investments: %.2f\n\n", output.Summary.Derived.Allocated))

	for _, warning := range output.Summary.Reconciliation.Warnings {
		if strings.Contains(warning, "investments allocated") {
			s.WriteString(fmt.Sprintf("WARNING: %s\n\n", warning))
		}
	}

	s.WriteString(fmt.Sprintf("Holdings performance (%s)\n", output.Summary.Currency))
	s.WriteString(fmt.Sprintf("  cost basis: %.2f\n", output.Summary.Derived.HoldingsCost))
	s.WriteString(fmt.Sprintf("  uPnL: %.2f\n", output.Summary.Derived.HoldingsPnL))
	if output.Summary.Derived.HoldingsFXImpact != nil && output.Summary.Derived.HoldingsPnLExclFX != nil {
		s.WriteString(fmt.Sprintf("  fx impact: %.2f\n", *output.Summary.Derived.HoldingsFXImpact))
		s.WriteString(fmt.Sprintf("  uPnL excl. FX: %.2f\n", *output.Summary.Derived.HoldingsPnLExclFX))
	} else {
		s.WriteString("  fx impact: n/a\n")
	}
	s.WriteString(fmt.Sprintf("  return: %.2f%%\n", output.Summary.Derived.HoldingsReturnPct))
	s.WriteString(fmt.Sprintf("  twr (est.): %.2f%%\n\n", output.Summary.Derived.TWRPctEst))

	s.WriteString(fmt.Sprintf("Account total (as of %s, %s)\n", output.Report.ReportDate, output.Summary.Currency))
	s.WriteString(fmt.Sprintf("  free cash: %.2f\n", output.Summary.Derived.FreeCash))
	s.WriteString(fmt.Sprintf("  investments allocated: %.2f\n", output.Summary.Derived.Allocated))
	s.WriteString(fmt.Sprintf("  account total: %.2f\n", output.Summary.Derived.AccountTotal))
	for _, warning := range output.Summary.Reconciliation.Warnings {
		if strings.Contains(warning, "account total") {
			s.WriteString(fmt.Sprintf("  WARNING: %s\n", warning))
		}
	}
	s.WriteString("\n")

	s.WriteString(fmt.Sprintf("Allocation (holdings only, as of %s):\n", output.Report.ReportDate))
	if output.Summary.Derived.HoldingsValue <= 0 {
		s.WriteString("  n/a (no holdings)\n")
	} else {
		for _, row := range output.Allocation {
			s.WriteString(fmt.Sprintf("  %-10s %7.2f%%  (%.2f %s)\n",
				row.Ticker, row.HoldingsPct, row.MarketValue, output.Summary.Currency))
		}
	}
	s.WriteString("\n")

	if !isAllTime(output.Report.Period) {
		s.WriteString(fmt.Sprintf("Period flows (executed trades, %s)\n", output.Summary.Currency))
		s.WriteString("  buys: 0.00\n")
		s.WriteString("  sells: 0.00\n")
		s.WriteString("  net: 0.00\n")
		s.WriteString("  Note: This is not implemented yet (requires History - Orders permission).\n\n")
	}

	if len(output.Holdings) == 0 {
		s.WriteString("No open positions.\n")
	} else {
		for _, h := range output.Holdings {
			s.WriteString(renderHolding(h, output.Summary.Currency))
		}
	}

	_, err := w.Write([]byte(s.String()))
	return err
}

func renderHolding(h portfolio.HoldingRow, currency string) string {
	fxImpactStr := "n/a"
	if h.FXImpact != nil {
		fxImpactStr = fmt.Sprintf("%.2f", *h.FXImpact)
	}

	return fmt.Sprintf(
		"%s (%s)\n  market value: %.2f %s (%.2f%% of holdings)\n  isin: %s | opened: %s\n  shares: %.6g | tradable: %.6g | in pies: %.6g\n  avg price: %.6g %s | current price: %.6g %s\n  invested: %.2f %s | uPnL: %.2f %s\n  fx impact (%s): %s %s\n\n",
		h.Name,
		h.Ticker,
		h.MarketValue,
		currency,
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
		currency,
		h.UnrealizedPnL,
		currency,
		portfolio.ChooseFXPair(h.FXPair, h.InstrumentCurrency, currency),
		fxImpactStr,
		currency,
	)
}

func formatPeriodLabel(period portfolio.PeriodRange) string {
	if period.From == nil || period.To == nil {
		return "all-time"
	}
	return *period.From + " -> " + *period.To
}

func isAllTime(period portfolio.PeriodRange) bool {
	return period.From == nil && period.To == nil
}
