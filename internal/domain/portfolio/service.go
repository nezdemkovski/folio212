package portfolio

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/nezdemkovski/folio212/internal/infrastructure/trading212"
)

type Service struct {
	client *trading212.Client
}

func NewService(client *trading212.Client) *Service {
	return &Service{client: client}
}

func (s *Service) GetPortfolio(ctx context.Context, period PeriodRange, includeRaw bool) (*Output, error) {
	summary, err := s.client.GetAccountSummary(ctx)
	if err != nil {
		return nil, classifyAccountError(err)
	}

	positions, err := s.client.GetPositions(ctx, "")
	if err != nil {
		return nil, classifyPortfolioError(err)
	}

	now := time.Now()

	holdingsValue := SumPositionsValue(positions)
	holdingsCost := SumPositionsCost(positions)
	holdingsPnL := SumPositionsPnL(positions)
	fxImpactSum, fxImpactOK := SumPositionsFXImpact(positions)

	pieCash := summary.Cash.InPies
	allocated := holdingsValue + pieCash
	freeCash := summary.Cash.AvailableToTrade + summary.Cash.ReservedForOrders

	holdingsReturn := CalculateHoldingsReturn(holdingsPnL, holdingsCost)
	twrPct := holdingsReturn // TWR approximation

	var holdingsFXImpact *float64
	var holdingsPnLExclFX *float64
	if fxImpactOK {
		v := fxImpactSum
		holdingsFXImpact = &v
		ex := holdingsPnL - fxImpactSum
		holdingsPnLExclFX = &ex
	}

	reconciliation := s.reconcile(summary, pieCash, freeCash, allocated)

	allocation := make([]AllocationRow, 0, len(positions))
	holdings := make([]HoldingRow, 0, len(positions))

	for _, p := range positions {
		pct := CalculateAllocationPercentage(p.CurrentValue(), holdingsValue)

		fxPair := ""
		if p.Instrument.Currency != "" && summary.Currency != "" && p.Instrument.Currency != summary.Currency {
			fxPair = p.Instrument.Currency + "/" + summary.Currency
		}

		allocation = append(allocation, AllocationRow{
			Ticker:      p.Instrument.Ticker,
			MarketValue: p.CurrentValue(),
			HoldingsPct: Round(pct, 2),
			HoldingsBps: PctToBps(pct),
		})

		opened := ""
		if !p.CreatedAt.IsZero() {
			opened = p.CreatedAt.Format(time.RFC3339)
		}

		holdings = append(holdings, HoldingRow{
			Ticker:             p.Instrument.Ticker,
			Name:               p.Instrument.Name,
			ISIN:               p.Instrument.ISIN,
			OpenedAt:           opened,
			Qty:                p.Quantity,
			TradableQty:        p.QuantityAvailableForTrading,
			QtyInPies:          p.QuantityInPies,
			InstrumentCurrency: p.Instrument.Currency,
			AvgPricePaid:       p.AveragePricePaid,
			CurrentPrice:       p.CurrentPrice,
			AccountCurrency:    summary.Currency,
			Invested:           p.Invested(),
			MarketValue:        p.CurrentValue(),
			UnrealizedPnL:      p.WalletImpact.UnrealizedProfitLoss,
			FXImpact:           p.WalletImpact.FXImpact,
			FXPair:             fxPair,
			HoldingsPct:        Round(pct, 2),
			HoldingsBps:        PctToBps(pct),
		})
	}

	sort.SliceStable(allocation, func(i, j int) bool {
		return allocation[i].MarketValue > allocation[j].MarketValue
	})
	sort.SliceStable(holdings, func(i, j int) bool {
		return holdings[i].MarketValue > holdings[j].MarketValue
	})

	output := &Output{
		SchemaVersion: SchemaVersion,
		Report: Report{
			ReportDate:  now.Format("2006-01-02"),
			GeneratedAt: now.Format(time.RFC3339),
			Period:      period,
		},
		Summary: Summary{
			Currency: summary.Currency,
			Derived: DerivedMetrics{
				HoldingsValue:     holdingsValue,
				PieCash:           pieCash,
				Allocated:         allocated,
				FreeCash:          freeCash,
				AccountTotal:      summary.TotalValue,
				HoldingsCost:      holdingsCost,
				HoldingsPnL:       holdingsPnL,
				HoldingsFXImpact:  holdingsFXImpact,
				HoldingsPnLExclFX: holdingsPnLExclFX,
				HoldingsReturnPct: Round(holdingsReturn, 4),
				HoldingsReturnBps: PctToBps(holdingsReturn),
				TWRPctEst:         Round(twrPct, 4),
				TWRBpsEst:         PctToBps(twrPct),
				TWRMethod:         "holdings-only-no-flows",
				TWRDescription:    "Estimated TWR based on holdings only; excludes cash flows and pie allocations.",
			},
			Snapshot: APISnapshot{
				APIInvestmentsValue: summary.Investments.CurrentValue,
				APICashInPies:       summary.Cash.InPies,
				APICashAvailable:    summary.Cash.AvailableToTrade,
				APICashReserved:     summary.Cash.ReservedForOrders,
				APIRealizedPnL:      summary.Investments.RealizedProfitLoss,
				APITotalCost:        summary.Investments.TotalCost,
				APITotalValue:       summary.TotalValue,
			},
			Reconciliation: reconciliation,
		},
		Allocation: allocation,
		Holdings:   holdings,
	}

	if includeRaw {
		output.Raw = &RawData{
			AccountSummary: summary,
			Positions:      positions,
		}
	}

	return output, nil
}

func (s *Service) reconcile(summary *trading212.AccountSummary, pieCash, freeCash, allocated float64) Reconciliation {
	var warnings []string

	accountTotal := summary.TotalValue

	if diff := accountTotal - (freeCash + allocated); Abs(diff) > 0.01 {
		warnings = append(warnings, fmt.Sprintf("account total does not reconcile (diff: %.2f %s)", diff, summary.Currency))
	}

	if diff := summary.Investments.CurrentValue - allocated; Abs(diff) > 0.01 {
		warnings = append(warnings, fmt.Sprintf("investments allocated does not reconcile (diff: %.2f %s)", diff, summary.Currency))
	}

	return Reconciliation{
		AllocatedDiff:    Round(summary.Investments.CurrentValue-allocated, 2),
		AccountTotalDiff: Round(summary.TotalValue-(freeCash+allocated), 2),
		Warnings:         warnings,
	}
}

func classifyAccountError(err error) error {
	if httpErr, ok := err.(*trading212.HTTPError); ok && httpErr != nil {
		if httpErr.StatusCode == 403 {
			return fmt.Errorf("%w: %v", ErrMissingAccountDataPermission, err)
		}
		if httpErr.StatusCode == 429 {
			return fmt.Errorf("%w: %v", ErrRateLimited, err)
		}
	}
	return err
}

func classifyPortfolioError(err error) error {
	if httpErr, ok := err.(*trading212.HTTPError); ok && httpErr != nil {
		if httpErr.StatusCode == 403 {
			return fmt.Errorf("%w: %v", ErrMissingPortfolioPermission, err)
		}
		if httpErr.StatusCode == 429 {
			return fmt.Errorf("%w: %v", ErrRateLimited, err)
		}
	}
	return err
}
