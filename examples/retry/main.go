// SPDX-License-Identifier: Apache-2.0 OR MIT

// Command retry shows WithRetry, the opt-in retry policy for idempotent
// calls. It reads Health and Faces through a retrying client, then
// points the same policy at a port nothing listens on and counts the
// attempts a failed connection gets, through a WithHTTPClient transport
// that counts requests. POST calls are never retried.
//
//	PAIN001_BASE_URL=http://127.0.0.1:8000 go run ./examples/retry
//
// It exits non-zero unless the unreachable server got every attempt.
package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	pain001 "github.com/sebastienrousseau/pain001-sdk-go"
)

var policy = pain001.RetryPolicy{MaxAttempts: 3, BaseDelay: 50 * time.Millisecond, MaxDelay: time.Second}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "retry:", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()
	client, err := pain001.NewClient(envOr("PAIN001_BASE_URL", "http://127.0.0.1:8000"),
		pain001.WithAPIKey(os.Getenv("PAIN001_API_KEY")), pain001.WithRetry(policy))
	if err != nil {
		return err
	}
	health, err := client.Health(ctx)
	if err != nil {
		return err
	}
	faces, err := client.Faces(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("health %s, %d faces, read with up to %d attempts each\n", health.Status, len(faces), policy.MaxAttempts)

	attempts, err := unreachable(ctx)
	fmt.Printf("unreachable server: %d attempts, then: %v\n", attempts, err)
	if attempts != int32(policy.MaxAttempts) {
		return fmt.Errorf("%d attempts, want %d", attempts, policy.MaxAttempts)
	}
	return nil
}

// unreachable calls Health on a loopback port that was just closed and
// returns how many requests the transport saw.
func unreachable(ctx context.Context) (int32, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	addr := ln.Addr().String()
	_ = ln.Close()
	var count atomic.Int32
	counting := &http.Client{Transport: countingTransport{&count}}
	client, err := pain001.NewClient("http://"+addr, pain001.WithRetry(policy), pain001.WithHTTPClient(counting))
	if err != nil {
		return 0, err
	}
	_, err = client.Health(ctx)
	return count.Load(), err
}

// countingTransport counts requests before handing them to the default
// transport.
type countingTransport struct{ n *atomic.Int32 }

func (t countingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	t.n.Add(1)
	return http.DefaultTransport.RoundTrip(r)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
