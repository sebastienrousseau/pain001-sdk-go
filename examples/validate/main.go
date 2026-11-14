// SPDX-License-Identifier: Apache-2.0 OR MIT

// Command validate checks the conformance kit's credit-transfer file
// without generating XML, first alone and then against the SEPA credit
// transfer rulebook, and prints the row counts and any findings.
//
//	PAIN001_BASE_URL=http://127.0.0.1:8000 go run ./examples/validate
//
// Run it from the repository root so the fixture path resolves.
package main

import (
	"context"
	"fmt"
	"os"

	pain001 "github.com/sebastienrousseau/pain001-sdk-go"
)

const fixture = "conformance/fixtures/credit-transfer.pain.001.001.09.csv"

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "validate:", err)
		os.Exit(1)
	}
}

func run() error {
	content, err := os.ReadFile(fixture)
	if err != nil {
		return err
	}
	client, err := pain001.NewClient(envOr("PAIN001_BASE_URL", "http://127.0.0.1:8000"),
		pain001.WithAPIKey(os.Getenv("PAIN001_API_KEY")))
	if err != nil {
		return err
	}
	for _, scheme := range []string{"", "sepa-sct"} {
		res, err := client.Validate(context.Background(), string(content), "payments.csv", pain001.Pain00100109, scheme)
		if err != nil {
			return err
		}
		fmt.Printf("scheme %-10q valid=%t rows=%d valid=%d invalid=%d scheme violations=%d\n",
			scheme, res.IsValid, res.TotalRows, res.ValidRows, res.InvalidRows, len(res.SchemeViolations))
		for _, row := range res.Errors {
			fmt.Printf("  row error: %s: %s\n", row.Field, row.Message)
		}
	}
	return nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
