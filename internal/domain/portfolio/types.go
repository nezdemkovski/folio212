package portfolio

import "github.com/nezdemkovski/folio212/internal/infrastructure/trading212"

const SchemaVersion = 1

type PeriodRange struct {
	From *string `json:"from"` // YYYY-MM-DD or null
	To   *string `json:"to"`   // YYYY-MM-DD or null
}

type Report struct {
	ReportDate  string      `json:"reportDate"`  // YYYY-MM-DD (local)
	GeneratedAt string      `json:"generatedAt"` // RFC3339 (local time, with timezone)
	Period      PeriodRange `json:"period"`
}

type DerivedMetrics struct {
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

type APISnapshot struct {
	APIInvestmentsValue float64 `json:"apiInvestmentsValue"`
	APICashInPies       float64 `json:"apiCashInPies"`
	APICashAvailable    float64 `json:"apiCashAvailable"`
	APICashReserved     float64 `json:"apiCashReserved"`
	APIRealizedPnL      float64 `json:"apiRealizedPnL"`
	APITotalCost        float64 `json:"apiTotalCost"`
	APITotalValue       float64 `json:"apiTotalValue"`
}

type Reconciliation struct {
	AllocatedDiff    float64  `json:"allocatedDiff"`
	AccountTotalDiff float64  `json:"accountTotalDiff"`
	Warnings         []string `json:"warnings,omitempty"`
}

type Summary struct {
	Currency       string         `json:"currency"`
	Derived        DerivedMetrics `json:"derived"`
	Snapshot       APISnapshot    `json:"snapshot"`
	Reconciliation Reconciliation `json:"reconcile"`
}

type AllocationRow struct {
	Ticker      string  `json:"ticker"`
	MarketValue float64 `json:"marketValue"`
	HoldingsPct float64 `json:"holdingsPct"`
	HoldingsBps int     `json:"holdingsBps"`
}

type HoldingRow struct {
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

type Output struct {
	SchemaVersion int             `json:"schemaVersion"`
	Report        Report          `json:"report"`
	Summary       Summary         `json:"summary"`
	Allocation    []AllocationRow `json:"allocation"`
	Holdings      []HoldingRow    `json:"holdings"`
	Raw           *RawData        `json:"raw,omitempty"`
}

type RawData struct {
	AccountSummary *trading212.AccountSummary `json:"accountSummary,omitempty"`
	Positions      []trading212.Position      `json:"positions"`
}
