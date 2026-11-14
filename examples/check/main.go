// SPDX-License-Identifier: Apache-2.0 OR MIT

// Command check runs the server's checks on the conformance kit's
// pain.001.001.09 document, adding the SEPA credit transfer profile and
// the SEPA Schematron preset, and prints the verdict, the rungs that ran
// and every finding.
//
//	PAIN001_BASE_URL=http://127.0.0.1:8000 go run ./examples/check
//
// Run it from the repository root so the fixture path resolves.
package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	pain001 "github.com/sebastienrousseau/pain001-sdk-go"
)

const fixture = "conformance/fixtures/credit-transfer.pain.001.001.09.xml"

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "check:", err)
		os.Exit(1)
	}
}

func run() error {
	xml, err := os.ReadFile(fixture)
	if err != nil {
		return err
	}
	client, err := pain001.NewClient(envOr("PAIN001_BASE_URL", "http://127.0.0.1:8000"),
		pain001.WithAPIKey(os.Getenv("PAIN001_API_KEY")))
	if err != nil {
		return err
	}
	res, err := client.Check(context.Background(), string(xml), "sepa-sct", "sepa")
	if err != nil {
		return err
	}
	fmt.Printf("%s passed=%t errors=%d warnings=%d rungs=%s\n",
		res.MessageType, res.Passed, res.Errors, res.Warnings, strings.Join(res.Rungs, ","))
	for _, f := range res.Findings {
		fmt.Printf("  %s %s %s: %s\n", f.Severity, f.Rung, f.Rule, f.Message)
	}
	return nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
