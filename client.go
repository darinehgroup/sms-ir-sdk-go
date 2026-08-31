// Package smsir provides a Go client for the SMS.ir web service API
// (https://api.sms.ir).
//
// Create a client with New and an API key from the SMS.ir developers panel.
// Sandbox keys work with the same client and base URL; only the key type
// differs. All methods take a context.Context and return typed results or a
// *APIError describing the failure reported by the server.
package smsir

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Version is the SDK version, reported in the default User-Agent header.
const Version = "0.1.0"

// DefaultBaseURL is the SMS.ir API base URL, shared by the production and
// sandbox environments.
const DefaultBaseURL = "https://api.sms.ir"

// Client is an SMS.ir API client. It is immutable after construction and safe
// for concurrent use.
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	userAgent  string
}

// Option configures a Client created by New.
type Option func(*Client)

// WithBaseURL overrides the default base URL (https://api.sms.ir). Trailing
// slashes are trimmed.
func WithBaseURL(u string) Option {
	return func(c *Client) { c.baseURL = strings.TrimRight(u, "/") }
}

// WithHTTPClient sets a custom *http.Client, e.g. to configure proxies,
// transports, or timeouts.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.httpClient = hc }
}

// WithUserAgent overrides the default User-Agent header
// ("smsir-go/<version>").
func WithUserAgent(ua string) Option {
	return func(c *Client) { c.userAgent = ua }
}

// New creates an SMS.ir API client. apiKey may be a production or sandbox key
// generated in the SMS.ir developers panel.
func New(apiKey string, opts ...Option) *Client {
	c := &Client{
		apiKey:     apiKey,
		baseURL:    DefaultBaseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		userAgent:  "smsir-go/" + Version,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// apiEnvelope is the uniform response wrapper returned by every endpoint.
type apiEnvelope struct {
	Status  int             `json:"status"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// do executes an authenticated API request and decodes the envelope's data
// field into out (unless out is nil).
func (c *Client) do(ctx context.Context, method, path string, query url.Values, body, out any) error {
	return c.doRequest(ctx, method, path, query, body, out, true)
}

// doRequest is the single request pipeline behind every public method.
// withAuth controls the X-API-KEY header; only SendViaURL disables it because
// that endpoint authenticates through query-string credentials instead.
func (c *Client) doRequest(ctx context.Context, method, path string, query url.Values, body, out any, withAuth bool) error {
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("smsir: encode request: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, u, reqBody)
	if err != nil {
		return fmt.Errorf("smsir: build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	if withAuth {
		req.Header.Set("X-API-KEY", c.apiKey)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("smsir: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("smsir: read response: %w", err)
	}

	var env apiEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return &APIError{
			HTTPStatus: resp.StatusCode,
			Status:     -1,
			Message:    truncate(strings.TrimSpace(string(raw)), 512),
		}
	}
	if env.Status != StatusSuccess {
		return &APIError{
			HTTPStatus: resp.StatusCode,
			Status:     env.Status,
			Message:    env.Message,
		}
	}
	if out != nil && len(env.Data) > 0 && string(env.Data) != "null" {
		if err := json.Unmarshal(env.Data, out); err != nil {
			return fmt.Errorf("smsir: decode response data: %w", err)
		}
	}
	return nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
