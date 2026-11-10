// SPDX-License-Identifier: Apache-2.0 OR MIT

package pain001

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// DefaultTimeout bounds each call when WithTimeout is not given.
const DefaultTimeout = 30 * time.Second

// maxResponseBytes caps how much of a response body the client reads.
// The server accepts documents up to 10 MiB, so 64 MiB leaves room for
// generated XML and JSON escaping. It is a variable so tests can lower it.
var maxResponseBytes int64 = 64 << 20

// Client calls one pain001 server. It is safe for concurrent use.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	timeout    time.Duration
}

// Option configures a Client in NewClient.
type Option func(*Client)

// WithAPIKey sends key as "Authorization: Bearer <key>" on every call.
// An empty key sends no header, for a server run without PAIN001_API_KEY.
func WithAPIKey(key string) Option {
	return func(c *Client) { c.apiKey = key }
}

// WithHTTPClient makes the client send its requests through hc, for a
// custom transport, proxy or TLS setting. A nil hc is ignored.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		if hc != nil {
			c.httpClient = hc
		}
	}
}

// WithTimeout bounds each call, including reading the response, to d.
// It applies on top of any deadline the caller's context carries. Zero
// or a negative d turns the client's own timeout off.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) { c.timeout = d }
}

// NewClient returns a client for the server at baseURL, for example
// "http://127.0.0.1:8000". The URL must be absolute http or https; a
// trailing slash is dropped.
func NewClient(baseURL string, opts ...Option) (*Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("pain001: invalid base URL: %w", err)
	}
	if (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return nil, fmt.Errorf("pain001: base URL %q must be absolute http or https", baseURL)
	}
	c := &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{},
		timeout:    DefaultTimeout,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c, nil
}

// Request sends one call and returns the HTTP status and the response
// body decoded as JSON, whatever the status. It is the low-level call the
// typed methods and the conformance runner share: a non-2xx status is not
// an error here. path starts with "/", for example "/api/v1/health".
// body, when not nil, is encoded as JSON. Numbers in the result are
// json.Number, so integers and decimals survive exactly; an empty
// response body decodes to nil.
func (c *Client) Request(ctx context.Context, method, path string, body any) (int, any, error) {
	status, raw, err := c.send(ctx, method, path, body)
	if err != nil {
		return 0, nil, err
	}
	if len(bytes.TrimSpace(raw)) == 0 {
		return status, nil, nil
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var doc any
	if err := dec.Decode(&doc); err != nil {
		return status, nil, fmt.Errorf("pain001: %s %s: HTTP %d response is not JSON: %w", method, path, status, err)
	}
	return status, doc, nil
}

// call sends one call and decodes a 2xx body into out; any other status
// becomes an *APIError.
func (c *Client) call(ctx context.Context, method, path string, body, out any) error {
	status, raw, err := c.send(ctx, method, path, body)
	if err != nil {
		return err
	}
	if status < 200 || status > 299 {
		return newAPIError(status, raw)
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("pain001: %s %s: decoding response: %w", method, path, err)
	}
	return nil
}

// send performs the HTTP exchange and returns the status and raw body.
func (c *Client) send(ctx context.Context, method, path string, body any) (int, []byte, error) {
	if !strings.HasPrefix(path, "/") {
		return 0, nil, fmt.Errorf("pain001: path %q must start with /", path)
	}
	if c.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.timeout)
		defer cancel()
	}
	req, err := c.newRequest(ctx, method, path, body)
	if err != nil {
		return 0, nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("pain001: %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes+1))
	if err != nil {
		return 0, nil, fmt.Errorf("pain001: %s %s: reading response: %w", method, path, err)
	}
	if int64(len(raw)) > maxResponseBytes {
		return 0, nil, fmt.Errorf("pain001: %s %s: %w", method, path, ErrResponseTooLarge)
	}
	return resp.StatusCode, raw, nil
}

// ErrResponseTooLarge reports a response body over the client's cap.
var ErrResponseTooLarge = errors.New("response body exceeds 64 MiB")

// newRequest builds the HTTP request with the JSON and auth headers.
func (c *Client) newRequest(ctx context.Context, method, path string, body any) (*http.Request, error) {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("pain001: encoding request body: %w", err)
		}
		reader = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return nil, fmt.Errorf("pain001: %s %s: %w", method, path, err)
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	return req, nil
}
