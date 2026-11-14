// SPDX-License-Identifier: Apache-2.0 OR MIT

// Command health asks a running `pain001 serve` whether it is up and
// which version it runs.
//
//	PAIN001_BASE_URL=http://127.0.0.1:8000 go run ./examples/health
package main

import (
	"context"
	"fmt"
	"os"

	pain001 "github.com/sebastienrousseau/pain001-sdk-go"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "health:", err)
		os.Exit(1)
	}
}

func run() error {
	client, err := pain001.NewClient(envOr("PAIN001_BASE_URL", "http://127.0.0.1:8000"))
	if err != nil {
		return err
	}
	res, err := client.Health(context.Background())
	if err != nil {
		return err
	}
	fmt.Printf("server %s, pain001 %s (client %s)\n", res.Status, res.Version, pain001.Version)
	return nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
