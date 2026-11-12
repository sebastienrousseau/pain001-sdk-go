// SPDX-License-Identifier: Apache-2.0 OR MIT

// Command basic generates a pain.001.001.09 file from the conformance
// kit's credit-transfer fixture through a running `pain001 serve`.
//
//	go run ./examples/basic -base http://127.0.0.1:8000 -out payments.xml
//
// It reads PAIN001_BASE_URL and PAIN001_API_KEY when set. Run it from the
// repository root so the default -in path resolves.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	pain001 "github.com/sebastienrousseau/pain001-sdk-go"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "basic:", err)
		os.Exit(1)
	}
}

func run() error {
	base := flag.String("base", envOr("PAIN001_BASE_URL", "http://127.0.0.1:8000"), "server base URL")
	in := flag.String("in", "conformance/fixtures/credit-transfer.pain.001.001.09.csv", "payment file (CSV)")
	out := flag.String("out", "payments.xml", "where to write the XML")
	flag.Parse()

	content, err := os.ReadFile(*in)
	if err != nil {
		return err
	}
	client, err := pain001.NewClient(*base,
		pain001.WithAPIKey(os.Getenv("PAIN001_API_KEY")),
		pain001.WithTimeout(30*time.Second))
	if err != nil {
		return err
	}
	res, err := client.Generate(context.Background(), string(content), "payments.csv", pain001.Pain00100109, "")
	var refusal *pain001.RefusalError
	var apiErr *pain001.APIError
	switch {
	case errors.As(err, &refusal):
		for _, row := range refusal.Response.ValidationErrors {
			fmt.Fprintf(os.Stderr, "row error: %s: %s\n", row.Field, row.Message)
		}
		for _, v := range refusal.Response.SchemeViolations {
			fmt.Fprintf(os.Stderr, "scheme violation: %v: %v\n", v["rule"], v["message"])
		}
		return err
	case errors.As(err, &apiErr):
		return fmt.Errorf("the server refused the request (HTTP %d): %w", apiErr.Status, err)
	case err != nil:
		return err
	}
	if err := os.WriteFile(*out, []byte(res.XML), 0o600); err != nil {
		return err
	}
	fmt.Printf("wrote %s: %d bytes, %d transaction(s)\n", *out, len(res.XML), strings.Count(res.XML, "<CdtTrfTxInf>"))
	return nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
