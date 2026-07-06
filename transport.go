package pulsight

import (
	"math/rand"
	"net/http"
	"time"
)

// tokenTransport injects the api token on every request and retries
// idempotent (GET/HEAD) requests on 429, honouring Retry-After. It wraps
// any base RoundTripper, so timeouts/proxies on the underlying transport
// are preserved.
type tokenTransport struct {
	token string
	base  http.RoundTripper
	retry int
}

func (t *tokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// The wire uses the standard Authorization header; product docs call
	// the credential an "api token", never a "Bearer token".
	req.Header.Set("Authorization", "Bearer "+t.token)

	base := t.base
	if base == nil {
		base = http.DefaultTransport
	}

	attempts := 1
	if req.Method == http.MethodGet || req.Method == http.MethodHead {
		attempts += t.retry
	}

	var resp *http.Response
	var err error
	for i := 0; i < attempts; i++ {
		resp, err = base.RoundTrip(req)
		if err != nil || resp.StatusCode != http.StatusTooManyRequests || i == attempts-1 {
			return resp, err
		}
		wait := parseRetryAfter(resp.Header.Get("Retry-After"))
		if wait <= 0 {
			wait = backoff(i)
		}
		resp.Body.Close()
		time.Sleep(wait)
	}
	return resp, err
}

// backoff is exponential (200ms base) with up to 100ms of jitter.
func backoff(attempt int) time.Duration {
	base := time.Duration(1<<uint(attempt)) * 200 * time.Millisecond
	jitter := time.Duration(rand.Int63n(int64(100 * time.Millisecond)))
	return base + jitter
}
