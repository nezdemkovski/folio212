package trading212

import (
	"fmt"
	"time"
)

type HTTPError struct {
	Method     string
	URL        string
	StatusCode int
	Body       string

	// Rate limiting (may be zero if not provided by server).
	RateLimitResetUnix int64 // unix timestamp (seconds)
	RetryAfterSeconds  int   // Retry-After header (seconds)
}

func (e *HTTPError) Error() string {
	if e.Body == "" {
		return fmt.Sprintf("%s %s failed: HTTP %d", e.Method, e.URL, e.StatusCode)
	}
	return fmt.Sprintf("%s %s failed: HTTP %d: %s", e.Method, e.URL, e.StatusCode, e.Body)
}

// SuggestedRetryDelay returns the server-advertised delay to wait before retrying.
// ok=false if there is no useful server hint.
func (e *HTTPError) SuggestedRetryDelay(now time.Time) (d time.Duration, ok bool) {
	if e == nil {
		return 0, false
	}
	if e.RetryAfterSeconds > 0 {
		return time.Duration(e.RetryAfterSeconds) * time.Second, true
	}
	if e.RateLimitResetUnix > 0 {
		resetAt := time.Unix(e.RateLimitResetUnix, 0)
		return time.Until(resetAt), true
	}
	return 0, false
}
