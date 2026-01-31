package trading212

import "time"

// Types are based on the official Trading 212 Public API OpenAPI spec.
// Source: `https://docs.trading212.com/_bundle/api.json?download`

type AccountSummary struct {
	Cash        Cash        `json:"cash"`
	Currency    string      `json:"currency"`
	ID          int64       `json:"id"`
	Investments Investments `json:"investments"`
	TotalValue  float64     `json:"totalValue"`
}

type Cash struct {
	AvailableToTrade  float64 `json:"availableToTrade"`
	InPies            float64 `json:"inPies"`
	ReservedForOrders float64 `json:"reservedForOrders"`
}

type Investments struct {
	CurrentValue         float64 `json:"currentValue"`
	RealizedProfitLoss   float64 `json:"realizedProfitLoss"`
	TotalCost            float64 `json:"totalCost"`
	UnrealizedProfitLoss float64 `json:"unrealizedProfitLoss"`
}

type Position struct {
	AveragePricePaid            float64              `json:"averagePricePaid"`
	CreatedAt                   time.Time            `json:"createdAt"`
	CurrentPrice                float64              `json:"currentPrice"`
	Instrument                  Instrument           `json:"instrument"`
	Quantity                    float64              `json:"quantity"`
	QuantityAvailableForTrading float64              `json:"quantityAvailableForTrading"`
	QuantityInPies              float64              `json:"quantityInPies"`
	WalletImpact                PositionWalletImpact `json:"walletImpact"`
}

// Invested is the total cost basis of the position in the account currency.
func (p Position) Invested() float64 {
	return p.WalletImpact.TotalCost
}

// CurrentValue is the current market value of the position in the account currency.
func (p Position) CurrentValue() float64 {
	return p.WalletImpact.CurrentValue
}

type PositionWalletImpact struct {
	Currency             string   `json:"currency"`
	CurrentValue         float64  `json:"currentValue"`
	FXImpact             *float64 `json:"fxImpact,omitempty"`
	TotalCost            float64  `json:"totalCost"`
	UnrealizedProfitLoss float64  `json:"unrealizedProfitLoss"`
}

type Instrument struct {
	Currency string `json:"currency"`
	ISIN     string `json:"isin"`
	Name     string `json:"name"`
	Ticker   string `json:"ticker"`
}

type TradableInstrument struct {
	AddedOn           time.Time `json:"addedOn"`
	CurrencyCode      string    `json:"currencyCode"`
	ExtendedHours     bool      `json:"extendedHours"`
	ISIN              string    `json:"isin"`
	MaxOpenQuantity   float64   `json:"maxOpenQuantity"`
	Name              string    `json:"name"`
	ShortName         string    `json:"shortName"`
	Ticker            string    `json:"ticker"`
	Type              string    `json:"type"` // e.g. "ETF", "STOCK"
	WorkingScheduleID int64     `json:"workingScheduleId"`
}
