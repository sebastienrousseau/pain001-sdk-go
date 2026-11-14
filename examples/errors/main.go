// SPDX-License-Identifier: Apache-2.0 OR MIT

// Command errors provokes each error the client returns and shows how
// to tell them apart with errors.As and errors.Is: an *APIError with a
// string detail (a bad IBAN, 400), an *APIError with validation issues
// (an unknown message type, 422), and a *RefusalError carrying a refused
// Generate's findings (a bad currency, 422). ErrNoDocument, which only
// a server older than pain001 0.0.87 can cause, is handled the same way.
//
//	PAIN001_BASE_URL=http://127.0.0.1:8000 go run ./examples/errors
//
// Run it from the repository root so the fixture paths resolve. It exits
// non-zero when a call does not fail the way it should.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	pain001 "github.com/sebastienrousseau/pain001-sdk-go"
)

const fixtures = "conformance/fixtures/"

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "errors:", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()
	client, err := pain001.NewClient(envOr("PAIN001_BASE_URL", "http://127.0.0.1:8000"),
		pain001.WithAPIKey(os.Getenv("PAIN001_API_KEY")))
	if err != nil {
		return err
	}
	badIBAN, err := os.ReadFile(fixtures + "bad-iban.pain.001.001.09.csv")
	if err != nil {
		return err
	}
	badCurrency, err := os.ReadFile(fixtures + "bad-currency.pain.001.001.09.csv")
	if err != nil {
		return err
	}
	_, err = client.Validate(ctx, string(badIBAN), "payments.csv", pain001.Pain00100109, "")
	if got := describe(err); got != "api-error" {
		return fmt.Errorf("bad IBAN: got %s", got)
	}
	_, err = client.Validate(ctx, string(badIBAN), "payments.csv", "pain.999.001.01", "")
	if got := describe(err); got != "api-error" {
		return fmt.Errorf("unknown message type: got %s", got)
	}
	_, err = client.Generate(ctx, string(badCurrency), "payments.csv", pain001.Pain00100109, "")
	if got := describe(err); got != "refusal" {
		return fmt.Errorf("bad currency: got %s", got)
	}
	return nil
}

// describe prints what err is and returns its kind. Check RefusalError
// before APIError: a refusal wraps an *APIError, so both match it.
func describe(err error) string {
	var refusal *pain001.RefusalError
	var apiErr *pain001.APIError
	switch {
	case err == nil:
		fmt.Println("no error")
		return "none"
	case errors.As(err, &refusal):
		fmt.Printf("refused (HTTP %d): %d row error(s), %d scheme violation(s)\n", refusal.Status,
			len(refusal.Response.ValidationErrors), len(refusal.Response.SchemeViolations))
		for _, row := range refusal.Response.ValidationErrors {
			fmt.Printf("  %s: %s\n", row.Field, row.Message)
		}
		return "refusal"
	case errors.As(err, &apiErr):
		describeAPIError(apiErr)
		return "api-error"
	case errors.Is(err, pain001.ErrNoDocument):
		fmt.Println("a 2xx without the XML: the server predates pain001 0.0.87")
		return "no-document"
	default:
		fmt.Println("no answer from the server:", err)
		return "transport"
	}
}

// describeAPIError prints a string detail or each validation issue.
func describeAPIError(e *pain001.APIError) {
	if text, ok := e.DetailText(); ok {
		fmt.Printf("HTTP %d: %s\n", e.Status, text)
		return
	}
	fmt.Printf("HTTP %d with %d issue(s)\n", e.Status, len(e.Issues()))
	for _, is := range e.Issues() {
		fmt.Printf("  %v: %s\n", is.Loc, is.Type)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
