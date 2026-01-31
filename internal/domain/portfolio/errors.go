package portfolio

import "errors"

var (
	ErrMissingAccountDataPermission = errors.New("missing account data permission")
	ErrMissingPortfolioPermission   = errors.New("missing portfolio permission")
	ErrRateLimited                  = errors.New("rate limited")
	ErrInvalidPeriod                = errors.New("invalid period")
	ErrConfigNotLoaded              = errors.New("config not loaded")
	ErrMissingAPIKey                = errors.New("missing api key")
	ErrMissingAPISecret             = errors.New("missing api secret")
)
