package presentation

import (
	"errors"
	"fmt"
	"time"

	"github.com/nezdemkovski/folio212/internal/domain/portfolio"
	"github.com/nezdemkovski/folio212/internal/infrastructure/trading212"
)

func HumanizeAccountError(err error) error {
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

func HumanizePortfolioError(err error) error {
	var httpErr *trading212.HTTPError
	if errors.As(err, &httpErr) && httpErr.StatusCode == 403 {
		return fmt.Errorf("%w (missing permission: enable \"Portfolio\" for your Trading212 API key)", err)
	}
	return err
}

func HumanizeDomainError(err error) string {
	switch {
	case errors.Is(err, portfolio.ErrConfigNotLoaded):
		return "config not loaded; please run 'folio212 init' first"
	case errors.Is(err, portfolio.ErrMissingAPIKey):
		return "missing trading212 api key; please run 'folio212 init'"
	case errors.Is(err, portfolio.ErrMissingAPISecret):
		return "missing trading212 api secret; please run 'folio212 init'"
	case errors.Is(err, portfolio.ErrInvalidPeriod):
		return "invalid period format"
	default:
		return err.Error()
	}
}
