// SPDX-License-Identifier: Apache-2.0 OR MIT

// Command apikey shows WithAPIKey, which sends "Authorization: Bearer
// <key>" on every call. A server started with PAIN001_API_KEY answers
// 401 to a call without it. The example makes the same call without and
// with the key read from PAIN001_API_KEY, and prints what each got.
//
//	PAIN001_API_KEY=secret pain001 serve --host 127.0.0.1 --port 8000 &
//	PAIN001_API_KEY=secret PAIN001_BASE_URL=http://127.0.0.1:8000 go run ./examples/apikey
//
// Run it from the repository root so the fixture path resolves. With
// PAIN001_API_KEY set, it exits non-zero unless the keyless call is
// refused with 401 and the keyed call succeeds.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"

	pain001 "github.com/sebastienrousseau/pain001-sdk-go"
)

const fixture = "conformance/fixtures/credit-transfer.pain.001.001.09.csv"

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "apikey:", err)
		os.Exit(1)
	}
}

func run() error {
	base := envOr("PAIN001_BASE_URL", "http://127.0.0.1:8000")
	key := os.Getenv("PAIN001_API_KEY")
	content, err := os.ReadFile(fixture)
	if err != nil {
		return err
	}
	if key == "" {
		fmt.Println("PAIN001_API_KEY is unset: the server must be running without one")
	} else {
		status := validate(base, "", string(content))
		fmt.Println("without the key:", status)
		if status != http.StatusUnauthorized {
			return fmt.Errorf("a keyless call got %d, want 401: is the server running with PAIN001_API_KEY?", status)
		}
	}
	status := validate(base, key, string(content))
	fmt.Println("with WithAPIKey:", status)
	if status != http.StatusOK {
		return fmt.Errorf("the keyed call got %d, want 200", status)
	}
	return nil
}

// validate calls Validate with key (none when empty) and returns the
// HTTP status it got: 200 on success, or the *APIError's status.
func validate(base, key, content string) int {
	client, err := pain001.NewClient(base, pain001.WithAPIKey(key))
	if err != nil {
		return 0
	}
	_, err = client.Validate(context.Background(), content, "payments.csv", pain001.Pain00100109, "")
	var apiErr *pain001.APIError
	switch {
	case err == nil:
		return http.StatusOK
	case errors.As(err, &apiErr):
		return apiErr.Status
	default:
		return 0
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
