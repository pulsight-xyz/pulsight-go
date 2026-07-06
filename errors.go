package pulsight

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// Credits carries the remaining api-credit balance surfaced on every
// response via the X-Credits-Remaining header. Present is false when the
// header is absent (unmetered route or session auth).
type Credits struct {
	Remaining int64
	Present   bool
}

// CreditsFromHeader reads the X-Credits-Remaining response header.
func CreditsFromHeader(h http.Header) Credits {
	v := h.Get("X-Credits-Remaining")
	if v == "" {
		return Credits{}
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return Credits{}
	}
	return Credits{Remaining: n, Present: true}
}

// CreditExhaustedError maps the 402 CREDIT_EXHAUSTED response: the api
// credit pool is empty for the current billing cycle.
type CreditExhaustedError struct {
	Pool string
}

func (e *CreditExhaustedError) Error() string {
	return fmt.Sprintf("pulsight: credit pool %q exhausted (HTTP 402)", e.Pool)
}

// RateLimitedError maps a 429 response. RetryAfter is the server's
// Retry-After hint (0 when absent).
type RateLimitedError struct {
	RetryAfter time.Duration
}

func (e *RateLimitedError) Error() string {
	return fmt.Sprintf("pulsight: rate limited, retry after %s (HTTP 429)", e.RetryAfter)
}

// MissingScopeError maps a 403 caused by an api token lacking a scope.
type MissingScopeError struct {
	Message string
}

func (e *MissingScopeError) Error() string {
	return fmt.Sprintf("pulsight: %s (HTTP 403)", e.Message)
}

// APIError is the fallback for any other non-2xx response.
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("pulsight: unexpected status %d: %s", e.StatusCode, e.Body)
}

// ErrorFromResponse maps a non-2xx *http.Response to a typed error and
// returns nil for 2xx. It consumes (and bounds) the response body, so call
// it before reading the body yourself. The generated client returns the
// raw *http.Response; pass it here to get first-class error types.
func ErrorFromResponse(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
	switch resp.StatusCode {
	case http.StatusPaymentRequired:
		var payload struct {
			Code string `json:"code"`
			Pool string `json:"pool"`
		}
		_ = json.Unmarshal(body, &payload)
		return &CreditExhaustedError{Pool: payload.Pool}
	case http.StatusTooManyRequests:
		return &RateLimitedError{RetryAfter: parseRetryAfter(resp.Header.Get("Retry-After"))}
	case http.StatusForbidden:
		return &MissingScopeError{Message: messageFromBody(body)}
	default:
		return &APIError{StatusCode: resp.StatusCode, Body: string(body)}
	}
}

func messageFromBody(body []byte) string {
	var payload struct {
		Error string `json:"error"`
	}
	if json.Unmarshal(body, &payload) == nil && payload.Error != "" {
		return payload.Error
	}
	return "forbidden"
}

func parseRetryAfter(v string) time.Duration {
	if v == "" {
		return 0
	}
	if secs, err := strconv.Atoi(v); err == nil {
		return time.Duration(secs) * time.Second
	}
	return 0
}
