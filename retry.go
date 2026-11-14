// SPDX-License-Identifier: Apache-2.0 OR MIT

package pain001

import (
	"context"
	"math/rand/v2"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// The RetryPolicy defaults for a zero BaseDelay or MaxDelay.
const (
	DefaultRetryBaseDelay = 200 * time.Millisecond
	DefaultRetryMaxDelay  = 5 * time.Second
)

// RetryPolicy configures WithRetry. The zero value retries nothing.
type RetryPolicy struct {
	// MaxAttempts is the number of attempts, the first one included.
	// Below 2, the client does not retry.
	MaxAttempts int
	// BaseDelay is the wait before the second attempt; each later wait
	// doubles it, with jitter. Zero means DefaultRetryBaseDelay.
	BaseDelay time.Duration
	// MaxDelay caps one wait. A Retry-After longer than MaxDelay ends
	// the retries and returns that response. Zero means
	// DefaultRetryMaxDelay.
	MaxDelay time.Duration
}

// WithRetry retries idempotent calls (Health, Faces, and Request with
// GET) when the connection fails or the server answers 429 or 503,
// waiting with exponential backoff and jitter, or for the response's
// Retry-After when it sends one. POST calls (Validate, Generate, Check,
// ProjectFace and Request with any other method) are never retried. The
// client's timeout and the caller's context bound all attempts together;
// when either ends during a wait, the last attempt's result is returned.
// Without this option the client makes exactly one attempt per call.
func WithRetry(p RetryPolicy) Option {
	return func(c *Client) {
		if p.MaxAttempts < 2 {
			c.retry = nil
			return
		}
		c.retry = &retrier{
			maxAttempts: p.MaxAttempts,
			base:        orDefault(p.BaseDelay, DefaultRetryBaseDelay),
			max:         orDefault(p.MaxDelay, DefaultRetryMaxDelay),
			jitter:      rand.Int64N,
			now:         time.Now,
			sleep:       sleep,
		}
	}
}

// orDefault returns d, or def when d is not positive.
func orDefault(d, def time.Duration) time.Duration {
	if d <= 0 {
		return def
	}
	return d
}

// retrier is a resolved RetryPolicy. The function fields are seams for
// tests; WithRetry sets them to the real clock, sleep and randomness.
type retrier struct {
	maxAttempts int
	base, max   time.Duration
	jitter      func(n int64) int64
	now         func() time.Time
	sleep       func(ctx context.Context, d time.Duration) error
}

// outcome is the result of one attempt.
type outcome struct {
	status     int
	raw        []byte
	retryAfter string
	err        error
	// transient marks a connection failure that a retry may cure: the
	// request did not get a response and the context is still live.
	transient bool
}

// wait reports whether attempt n, which produced o, is retried, and
// after how long. A nil retrier never retries.
func (r *retrier) wait(method string, n int, o outcome) (time.Duration, bool) {
	if r == nil || method != http.MethodGet || n >= r.maxAttempts {
		return 0, false
	}
	if o.transient {
		return r.backoff(n), true
	}
	if o.err != nil || (o.status != http.StatusTooManyRequests && o.status != http.StatusServiceUnavailable) {
		return 0, false
	}
	if d, ok := retryAfter(o.retryAfter, r.now()); ok {
		return d, d <= r.max
	}
	return r.backoff(n), true
}

// backoff is the wait after attempt n: base doubled n-1 times, capped
// at max, then drawn from its upper half so waits never collapse to 0.
func (r *retrier) backoff(n int) time.Duration {
	d := min(r.base, r.max)
	for i := 1; i < n && d < r.max; i++ {
		d = min(d, r.max/2) * 2
	}
	half := int64(d / 2)
	return time.Duration(half + r.jitter(int64(d)-half+1))
}

// retryAfter parses a Retry-After header: delay seconds or an HTTP
// date. A date in the past is a zero wait.
func retryAfter(value string, now time.Time) (time.Duration, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, false
	}
	if secs, err := strconv.ParseInt(value, 10, 64); err == nil {
		if secs < 0 || secs > int64(time.Duration(1<<62)/time.Second) {
			return 0, false
		}
		return time.Duration(secs) * time.Second, true
	}
	at, err := http.ParseTime(value)
	if err != nil {
		return 0, false
	}
	return max(at.Sub(now), 0), true
}

// sleep waits for d or until ctx ends, whichever is first.
func sleep(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
