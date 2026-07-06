package pulsight

import (
	"net/http"
	"time"
)

// DefaultBaseURL is the production Pulsight API root.
const DefaultBaseURL = "https://pulsight.xyz"

// Client carries the configuration the generated client needs: an
// authenticated *http.Client and the API base URL. Build it with New, then
// hand HTTPClient() and BaseURL() to the generated constructor (see
// README). Keeping this layer standalone means the handwritten code
// compiles without the generated core present.
type Client struct {
	baseURL    string
	httpClient *http.Client
	transport  *tokenTransport
}

// Option configures a Client.
type Option func(*Client)

// WithBaseURL overrides the API root (e.g. a staging URL).
func WithBaseURL(u string) Option { return func(c *Client) { c.baseURL = u } }

// WithRetries sets how many times an idempotent request is retried on 429
// (default 2). 0 disables retries.
func WithRetries(n int) Option {
	return func(c *Client) {
		if n >= 0 {
			c.transport.retry = n
		}
	}
}

// WithTimeout sets the per-request timeout (default 30s).
func WithTimeout(d time.Duration) Option {
	return func(c *Client) { c.httpClient.Timeout = d }
}

// New builds an authenticated client for the given api token (pk_live_…).
func New(apiToken string, opts ...Option) *Client {
	tr := &tokenTransport{token: apiToken, retry: 2}
	c := &Client{
		baseURL:    DefaultBaseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second, Transport: tr},
		transport:  tr,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// HTTPClient returns the authenticated *http.Client to hand to the
// generated API constructor.
func (c *Client) HTTPClient() *http.Client { return c.httpClient }

// BaseURL returns the configured API root.
func (c *Client) BaseURL() string { return c.baseURL }
