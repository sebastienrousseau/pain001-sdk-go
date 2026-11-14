// SPDX-License-Identifier: Apache-2.0 OR MIT

// Command request uses the raw Request call, the escape hatch for any
// endpoint the typed methods do not wrap. It reads the health document,
// then posts a deliberately unknown message type and prints the 422's
// status and first issue: Request returns any status as data, not as an
// error.
//
//	PAIN001_BASE_URL=http://127.0.0.1:8000 go run ./examples/request
package main

import (
	"context"
	"fmt"
	"os"

	pain001 "github.com/sebastienrousseau/pain001-sdk-go"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "request:", err)
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
	status, doc, err := client.Request(ctx, "GET", "/api/v1/health", nil)
	if err != nil {
		return err
	}
	health, _ := doc.(map[string]any)
	fmt.Printf("GET /api/v1/health: %d, status %v\n", status, health["status"])

	body := map[string]any{"filename": "payments.csv", "content": "id\n", "message_type": "pain.999.001.01"}
	status, doc, err = client.Request(ctx, "POST", "/api/v1/ui/validate", body)
	if err != nil {
		return err
	}
	fmt.Printf("POST /api/v1/ui/validate: %d, first issue: %s\n", status, firstIssue(doc))
	return nil
}

// firstIssue digs the first message out of a 422 detail array.
func firstIssue(doc any) string {
	obj, _ := doc.(map[string]any)
	issues, _ := obj["detail"].([]any)
	if len(issues) == 0 {
		return "none"
	}
	issue, _ := issues[0].(map[string]any)
	return fmt.Sprint(issue["loc"], " ", issue["msg"])
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
