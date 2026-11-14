// SPDX-License-Identifier: Apache-2.0 OR MIT

// Command timeout shows the two ways to bound a call: WithTimeout on the
// client, which applies to every call, and a deadline on the context of
// one call. Both end in an error that errors.Is matches against
// context.DeadlineExceeded. The shortest deadline wins.
//
//	PAIN001_BASE_URL=http://127.0.0.1:8000 go run ./examples/timeout
//
// It exits non-zero when a deadline does not end the call.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	pain001 "github.com/sebastienrousseau/pain001-sdk-go"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "timeout:", err)
		os.Exit(1)
	}
}

func run() error {
	base := envOr("PAIN001_BASE_URL", "http://127.0.0.1:8000")

	// A generous client timeout: the call completes.
	client, err := pain001.NewClient(base, pain001.WithTimeout(10*time.Second))
	if err != nil {
		return err
	}
	if _, err := client.Health(context.Background()); err != nil {
		return err
	}
	fmt.Println("WithTimeout(10s): the call completed")

	// A deadline on one call's context that has already passed.
	ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	defer cancel()
	if _, err := client.Health(ctx); !errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("context deadline: want DeadlineExceeded, got %v", err)
	}
	fmt.Println("context deadline: context.DeadlineExceeded")

	// A client timeout too short for any call.
	hasty, err := pain001.NewClient(base, pain001.WithTimeout(time.Nanosecond))
	if err != nil {
		return err
	}
	if _, err := hasty.Health(context.Background()); !errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("WithTimeout(1ns): want DeadlineExceeded, got %v", err)
	}
	fmt.Println("WithTimeout(1ns): context.DeadlineExceeded")
	return nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
