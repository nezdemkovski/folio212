package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/nezdemkovski/folio212/internal/domain/portfolio"
	"github.com/nezdemkovski/folio212/internal/infrastructure/secrets"
	"github.com/nezdemkovski/folio212/internal/infrastructure/trading212"
	"github.com/nezdemkovski/folio212/internal/presentation"
	"github.com/spf13/cobra"
)

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
			return fmt.Errorf("%s", presentation.HumanizeDomainError(portfolio.ErrConfigNotLoaded))
		}
		if strings.TrimSpace(cfg.Trading212APIKey) == "" {
			return fmt.Errorf("%s", presentation.HumanizeDomainError(portfolio.ErrMissingAPIKey))
		}

		secret, _, err := secrets.Get(secrets.KeyTrading212APISecret)
		if err != nil {
			return err
		}
		secret = strings.TrimSpace(secret)
		if secret == "" {
			return fmt.Errorf("%s", presentation.HumanizeDomainError(portfolio.ErrMissingAPISecret))
		}

		baseURL := trading212.BaseURLDemo
		if strings.EqualFold(strings.TrimSpace(cfg.Trading212Env), "live") {
			baseURL = trading212.BaseURLLive
		}

		period, err := parsePeriod(fromStr, toStr)
		if err != nil {
			return fmt.Errorf("%s: %w", presentation.HumanizeDomainError(portfolio.ErrInvalidPeriod), err)
		}

		client, err := trading212.NewClient(baseURL, cfg.Trading212APIKey, secret)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		svc := portfolio.NewService(client)
		output, err := svc.GetPortfolio(ctx, period, includeRaw)
		if err != nil {
			return presentation.HumanizeAccountError(err)
		}

		if asJSON {
			enc := json.NewEncoder(os.Stdout)
			return enc.Encode(output)
		}

		return presentation.RenderPortfolioText(output, os.Stdout)
	},
}

func parsePeriod(fromStr, toStr string) (portfolio.PeriodRange, error) {
	fromStr = strings.TrimSpace(fromStr)
	toStr = strings.TrimSpace(toStr)
	period := portfolio.PeriodRange{From: nil, To: nil}

	if fromStr == "" && toStr == "" {
		return period, nil
	}
	if fromStr == "" || toStr == "" {
		return period, fmt.Errorf("reporting period requires both --from and --to (YYYY-MM-DD)")
	}
	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		return period, fmt.Errorf("invalid --from (expected YYYY-MM-DD): %w", err)
	}
	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		return period, fmt.Errorf("invalid --to (expected YYYY-MM-DD): %w", err)
	}
	if to.Before(from) {
		return period, fmt.Errorf("invalid period: --to is before --from")
	}
	period.From = &fromStr
	period.To = &toStr
	return period, nil
}

func init() {
	portfolioCmd.Flags().Bool("json", false, "Output raw JSON")
	portfolioCmd.Flags().Bool("include-raw", false, "Include raw API payloads in JSON output")
	portfolioCmd.Flags().String("from", "", "Reporting period start (YYYY-MM-DD)")
	portfolioCmd.Flags().String("to", "", "Reporting period end (YYYY-MM-DD)")
}
