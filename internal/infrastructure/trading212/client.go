package trading212

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	BaseURLDemo = "https://demo.trading212.com"
	BaseURLLive = "https://live.trading212.com"
)

type Client struct {
	baseURL   string
	apiKey    string
	apiSecret string
	userAgent string
	http      *http.Client
}

type Option func(*Client)

func WithHTTPClient(h *http.Client) Option {
	return func(c *Client) {
		if h != nil {
			c.http = h
		}
	}
}

func WithUserAgent(ua string) Option {
	return func(c *Client) {
		c.userAgent = strings.TrimSpace(ua)
	}
}

func NewClient(baseURL, apiKey, apiSecret string, opts ...Option) (*Client, error) {
	if strings.TrimSpace(baseURL) == "" {
		return nil, fmt.Errorf("baseURL is required")
	}
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("apiKey is required")
	}
	if strings.TrimSpace(apiSecret) == "" {
		return nil, fmt.Errorf("apiSecret is required")
	}

	c := &Client{
		baseURL:   strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		apiKey:    apiKey,
		apiSecret: apiSecret,
		userAgent: "folio212",
		http: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
	for _, opt := range opts {
		if opt != nil {
			opt(c)
		}
	}
	return c, nil
}

func (c *Client) GetAccountSummary(ctx context.Context) (*AccountSummary, error) {
	var out AccountSummary
	if err := c.doJSON(ctx, http.MethodGet, "/api/v0/equity/account/summary", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetPositions returns your open positions. If ticker is non-empty, the API filters to that ticker.
func (c *Client) GetPositions(ctx context.Context, ticker string) ([]Position, error) {
	var q url.Values
	if strings.TrimSpace(ticker) != "" {
		q = url.Values{}
		q.Set("ticker", strings.TrimSpace(ticker))
	}

	var out []Position
	if err := c.doJSON(ctx, http.MethodGet, "/api/v0/equity/positions", q, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetInstruments returns all tradable instruments (stocks, ETFs, etc.). This can be large.
func (c *Client) GetInstruments(ctx context.Context) ([]TradableInstrument, error) {
	var out []TradableInstrument
	if err := c.doJSON(ctx, http.MethodGet, "/api/v0/equity/metadata/instruments", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) doJSON(ctx context.Context, method, path string, query url.Values, out any) error {
	if ctx == nil {
		ctx = context.Background()
	}

	u, err := url.Parse(c.baseURL)
	if err != nil {
		return fmt.Errorf("invalid baseURL %q: %w", c.baseURL, err)
	}
	u.Path = strings.TrimRight(u.Path, "/") + path
	if query != nil {
		u.RawQuery = query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(c.apiKey, c.apiSecret)
	req.Header.Set("Accept", "application/json")
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	for attempt := 0; attempt < 2; attempt++ {
		resp, err := c.http.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			b, _ := io.ReadAll(io.LimitReader(resp.Body, 8*1024))

			httpErr := &HTTPError{
				Method:     method,
				URL:        u.String(),
				StatusCode: resp.StatusCode,
				Body:       strings.TrimSpace(string(b)),
			}

			// Capture rate-limit headers (if present).
			if v := strings.TrimSpace(resp.Header.Get("Retry-After")); v != "" {
				if n, perr := strconv.Atoi(v); perr == nil && n > 0 {
					httpErr.RetryAfterSeconds = n
				}
			}
			if v := strings.TrimSpace(resp.Header.Get("x-ratelimit-reset")); v != "" {
				if n, perr := strconv.ParseInt(v, 10, 64); perr == nil && n > 0 {
					httpErr.RateLimitResetUnix = n
				}
			}

			// 429: retry once with server-advertised delay (bounded).
			if resp.StatusCode == http.StatusTooManyRequests && attempt == 0 {
				if d, ok := httpErr.SuggestedRetryDelay(time.Now()); ok {
					if d < 0 {
						d = 0
					}
					if d > 8*time.Second {
						// Too long for a friendly retry; return error with hints instead.
						return httpErr
					}
					timer := time.NewTimer(d)
					select {
					case <-ctx.Done():
						timer.Stop()
						return ctx.Err()
					case <-timer.C:
					}
					continue
				}
				// Fallback minimal wait for typical per-endpoint limits (e.g. 1 req / 5s).
				timer := time.NewTimer(5 * time.Second)
				select {
				case <-ctx.Done():
					timer.Stop()
					return ctx.Err()
				case <-timer.C:
				}
				continue
			}

			return httpErr
		}

		if out == nil {
			io.Copy(io.Discard, resp.Body)
			return nil
		}

		dec := json.NewDecoder(resp.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(out); err != nil {
			var se *json.SyntaxError
			if errors.As(err, &se) {
				return fmt.Errorf("failed to decode JSON response (syntax error at byte %d): %w", se.Offset, err)
			}
			return fmt.Errorf("failed to decode JSON response: %w", err)
		}
		return nil
	}

	// Should never happen (loop returns on success or error).
	return fmt.Errorf("request retry loop fell through unexpectedly")

}
