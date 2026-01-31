package run

import (
	"context"
	"time"

	"github.com/nezdemkovski/folio212/internal/infrastructure/config"
)

type Manager struct {
	cfg *config.Config
}

type Result struct {
	Environment string
	Completed   []string
	Duration    time.Duration
}

func NewManager(cfg *config.Config) *Manager {
	return &Manager{cfg: cfg}
}

func (m *Manager) Run(ctx context.Context) (*Result, error) {
	start := time.Now()
	completed := make([]string, 0, 3)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	time.Sleep(200 * time.Millisecond)
	completed = append(completed, "Configuration loaded")

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	time.Sleep(350 * time.Millisecond)
	completed = append(completed, "Work completed")

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	time.Sleep(150 * time.Millisecond)
	completed = append(completed, "Finalized")

	env := ""
	if m.cfg != nil {
		env = m.cfg.Environment
	}

	return &Result{
		Environment: env,
		Completed:   completed,
		Duration:    time.Since(start),
	}, nil
}
