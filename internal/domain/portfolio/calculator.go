package portfolio

import (
	"math"

	"github.com/nezdemkovski/folio212/internal/infrastructure/trading212"
)

func SumPositionsValue(positions []trading212.Position) float64 {
	var sum float64
	for _, p := range positions {
		sum += p.CurrentValue()
	}
	return sum
}

func SumPositionsCost(positions []trading212.Position) float64 {
	var sum float64
	for _, p := range positions {
		sum += p.Invested()
	}
	return sum
}

func SumPositionsPnL(positions []trading212.Position) float64 {
	var sum float64
	for _, p := range positions {
		sum += p.WalletImpact.UnrealizedProfitLoss
	}
	return sum
}

func SumPositionsFXImpact(positions []trading212.Position) (float64, bool) {
	var sum float64
	ok := true
	for _, p := range positions {
		if p.WalletImpact.FXImpact == nil {
			ok = false
			continue
		}
		sum += *p.WalletImpact.FXImpact
	}
	return sum, ok
}

func Abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func Round(x float64, places int) float64 {
	p := math.Pow(10, float64(places))
	return math.Round(x*p) / p
}

func PctToBps(pct float64) int {
	return int(math.Round(pct * 100))
}

func ChooseFXPair(explicit, instrumentCurrency, accountCurrency string) string {
	if explicit != "" {
		return explicit
	}
	if instrumentCurrency == "" || accountCurrency == "" || instrumentCurrency == accountCurrency {
		return "n/a"
	}
	return instrumentCurrency + "/" + accountCurrency
}

func CalculateAllocationPercentage(positionValue, totalHoldingsValue float64) float64 {
	if totalHoldingsValue <= 0 {
		return 0
	}
	return (positionValue / totalHoldingsValue) * 100
}

func CalculateHoldingsReturn(holdingsPnL, holdingsCost float64) float64 {
	if holdingsCost <= 0 {
		return 0
	}
	return (holdingsPnL / holdingsCost) * 100
}
