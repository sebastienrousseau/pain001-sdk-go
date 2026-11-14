// SPDX-License-Identifier: Apache-2.0 OR MIT

package pain001

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// script serves the given statuses in turn, then 200, and counts calls.
// A status of 0 drops the connection instead of answering.
func script(t *testing.T, retryAfter string, statuses ...int) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := int(calls.Add(1))
		if n > len(statuses) {
			_, _ = io.WriteString(w, `{"status":"healthy"}`)
			return
		}
		if statuses[n-1] == 0 {
			conn, _, _ := w.(http.Hijacker).Hijack()
			_ = conn.Close()
			return
		}
		if retryAfter != "" {
			w.Header().Set("Retry-After", retryAfter)
		}
		w.WriteHeader(statuses[n-1])
		_, _ = io.WriteString(w, `{"detail":"busy"}`)
	}))
	t.Cleanup(srv.Close)
	return srv, &calls
}

// retrying returns a client for srv with policy p whose waits are
// recorded instead of slept.
func retrying(t *testing.T, srv *httptest.Server, p RetryPolicy) (*Client, *[]time.Duration) {
	t.Helper()
	c, err := NewClient(srv.URL, WithRetry(p))
	if err != nil {
		t.Fatal(err)
	}
	waits := &[]time.Duration{}
	c.retry.jitter = func(n int64) int64 { return n - 1 }
	c.retry.sleep = func(_ context.Context, d time.Duration) error {
		*waits = append(*waits, d)
		return nil
	}
	return c, waits
}

func TestNoRetryWithoutTheOption(t *testing.T) {
	srv, calls := script(t, "", 503)
	c, err := NewClient(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	var apiErr *APIError
	if _, err := c.Health(context.Background()); !errors.As(err, &apiErr) || apiErr.Status != 503 {
		t.Errorf("err %v, want the 503", err)
	}
	if calls.Load() != 1 {
		t.Errorf("%d calls, want 1", calls.Load())
	}
}

func TestGetIsRetriedWithExponentialBackoff(t *testing.T) {
	srv, calls := script(t, "", 503, 429, 503)
	c, waits := retrying(t, srv, RetryPolicy{MaxAttempts: 4, BaseDelay: 100 * time.Millisecond, MaxDelay: 300 * time.Millisecond})
	if _, err := c.Health(context.Background()); err != nil {
		t.Fatal(err)
	}
	want := []time.Duration{100 * time.Millisecond, 200 * time.Millisecond, 300 * time.Millisecond}
	if calls.Load() != 4 || !equalDurations(*waits, want) {
		t.Errorf("%d calls, waits %v, want 4 and %v", calls.Load(), *waits, want)
	}
}

func equalDurations(a, b []time.Duration) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestRetryStopsAtMaxAttempts(t *testing.T) {
	srv, calls := script(t, "", 503, 503, 503)
	c, _ := retrying(t, srv, RetryPolicy{MaxAttempts: 2})
	var apiErr *APIError
	if _, err := c.Faces(context.Background()); !errors.As(err, &apiErr) || apiErr.Status != 503 {
		t.Errorf("err %v, want the last 503", err)
	}
	if calls.Load() != 2 {
		t.Errorf("%d calls, want 2", calls.Load())
	}
}

func TestRetryAfterIsHonoured(t *testing.T) {
	srv, _ := script(t, "2", 429)
	c, waits := retrying(t, srv, RetryPolicy{MaxAttempts: 3})
	if _, err := c.Health(context.Background()); err != nil || !equalDurations(*waits, []time.Duration{2 * time.Second}) {
		t.Errorf("err %v, waits %v", err, *waits)
	}
	date := time.Now().Add(-time.Hour).UTC().Format(http.TimeFormat)
	srv, _ = script(t, date, 503)
	c, waits = retrying(t, srv, RetryPolicy{MaxAttempts: 3})
	if _, err := c.Health(context.Background()); err != nil || !equalDurations(*waits, []time.Duration{0}) {
		t.Errorf("past date: err %v, waits %v", err, *waits)
	}
}

func TestRetryAfterBeyondMaxDelayEndsTheRetries(t *testing.T) {
	srv, calls := script(t, "60", 503)
	c, waits := retrying(t, srv, RetryPolicy{MaxAttempts: 3, MaxDelay: time.Second})
	var apiErr *APIError
	if _, err := c.Health(context.Background()); !errors.As(err, &apiErr) || calls.Load() != 1 || len(*waits) != 0 {
		t.Errorf("err %v, %d calls, waits %v", err, calls.Load(), *waits)
	}
}

func TestPostIsNeverRetried(t *testing.T) {
	ctx := context.Background()
	for name, run := range map[string]func(c *Client) error{
		"validate": func(c *Client) error { _, err := c.Validate(ctx, "", "p.csv", "", ""); return err },
		"generate": func(c *Client) error { _, err := c.Generate(ctx, "", "p.csv", "", ""); return err },
		"check":    func(c *Client) error { _, err := c.Check(ctx, "<x/>", "", ""); return err },
		"face":     func(c *Client) error { _, err := c.ProjectFace(ctx, "f", "<x/>", "", nil); return err },
		"request":  func(c *Client) error { _, _, err := c.Request(ctx, "POST", "/x", map[string]any{}); return err },
	} {
		srv, calls := script(t, "", 503, 0)
		c, _ := retrying(t, srv, RetryPolicy{MaxAttempts: 5})
		_ = run(c)
		if calls.Load() != 1 {
			t.Errorf("%s: %d calls, want 1", name, calls.Load())
		}
	}
}

func TestOtherStatusesAreNotRetried(t *testing.T) {
	srv, calls := script(t, "", 500)
	c, _ := retrying(t, srv, RetryPolicy{MaxAttempts: 5})
	if status, _, err := c.Request(context.Background(), "GET", "/x", nil); err != nil || status != 500 || calls.Load() != 1 {
		t.Errorf("status %d, err %v, %d calls", status, err, calls.Load())
	}
}

func TestConnectionFailuresAreRetried(t *testing.T) {
	srv, calls := script(t, "", 0, 0)
	c, waits := retrying(t, srv, RetryPolicy{MaxAttempts: 3})
	if status, _, err := c.Request(context.Background(), "GET", "/x", nil); err != nil || status != 200 {
		t.Fatalf("status %d, err %v", status, err)
	}
	if calls.Load() != 3 || len(*waits) != 2 {
		t.Errorf("%d calls, waits %v", calls.Load(), *waits)
	}
}

func TestErrorsThatARetryCannotCureAreReturnedAtOnce(t *testing.T) {
	srv, calls := script(t, "")
	c, _ := retrying(t, srv, RetryPolicy{MaxAttempts: 3})
	if _, _, err := c.Request(context.Background(), "GET", "/x", make(chan int)); err == nil || calls.Load() != 0 {
		t.Errorf("unencodable body: err %v, %d calls", err, calls.Load())
	}
	reads := 0
	c.httpClient = &http.Client{Transport: roundTrip(func(*http.Request) (*http.Response, error) {
		reads++
		return &http.Response{StatusCode: 503, Body: failingBody{}, Header: http.Header{}}, nil
	})}
	if _, err := c.Health(context.Background()); err == nil || reads != 1 {
		t.Errorf("read error: err %v, %d reads", err, reads)
	}
}

func TestACancelledContextEndsTheRetries(t *testing.T) {
	srv, calls := script(t, "", 503, 503)
	c, err := NewClient(srv.URL, WithRetry(RetryPolicy{MaxAttempts: 3, BaseDelay: time.Hour, MaxDelay: time.Hour}))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	var apiErr *APIError
	if _, err := c.Health(ctx); !errors.As(err, &apiErr) || apiErr.Status != 503 || calls.Load() != 1 {
		t.Errorf("err %v, %d calls, want the first 503", err, calls.Load())
	}
	failing := &http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
		cancel()
		return nil, errors.New("connection refused")
	})}
	c.httpClient = failing
	if _, err := c.Health(ctx); err == nil || !strings.Contains(err.Error(), "connection refused") {
		t.Errorf("err %v", err)
	}
}

func TestWithRetryResolvesThePolicy(t *testing.T) {
	c, _ := NewClient("http://h", WithRetry(RetryPolicy{MaxAttempts: 3}), WithRetry(RetryPolicy{MaxAttempts: 1}))
	if c.retry != nil {
		t.Error("MaxAttempts below 2 must turn retrying off")
	}
	c, _ = NewClient("http://h", WithRetry(RetryPolicy{MaxAttempts: 2}))
	if c.retry.base != DefaultRetryBaseDelay || c.retry.max != DefaultRetryMaxDelay {
		t.Errorf("defaults %v %v", c.retry.base, c.retry.max)
	}
	if d := c.retry.backoff(1); d < DefaultRetryBaseDelay/2 || d > DefaultRetryBaseDelay {
		t.Errorf("real jitter gave %v", d)
	}
}

func TestBackoffIsCappedAndNeverOverflows(t *testing.T) {
	r := &retrier{base: time.Second, max: time.Duration(1<<63 - 1)}
	r.jitter = func(n int64) int64 { return n - 1 }
	for _, n := range []int{1, 2, 40, 100, 1000} {
		if d := r.backoff(n); d <= 0 || d > r.max {
			t.Errorf("backoff(%d) = %v", n, d)
		}
	}
	r = &retrier{base: time.Minute, max: time.Second, jitter: func(int64) int64 { return 0 }}
	if d := r.backoff(1); d != 500*time.Millisecond {
		t.Errorf("a base above max must be capped: %v", d)
	}
}

func TestRetryAfterParsing(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	cases := map[string]struct {
		d  time.Duration
		ok bool
	}{
		"":                              {0, false},
		" 3 ":                           {3 * time.Second, true},
		"-1":                            {0, false},
		"99999999999999999":             {0, false},
		"soon":                          {0, false},
		"Thu, 01 Jan 2026 00:00:10 GMT": {10 * time.Second, true},
		"Wed, 31 Dec 2025 00:00:00 GMT": {0, true},
	}
	for in, want := range cases {
		if d, ok := retryAfter(in, now); d != want.d || ok != want.ok {
			t.Errorf("retryAfter(%q) = %v, %v; want %v, %v", in, d, ok, want.d, want.ok)
		}
	}
}

func TestSleepWaitsOrStopsWithTheContext(t *testing.T) {
	if err := sleep(context.Background(), time.Millisecond); err != nil {
		t.Errorf("sleep: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := sleep(ctx, time.Hour); !errors.Is(err, context.Canceled) {
		t.Errorf("sleep on a cancelled context: %v", err)
	}
}
