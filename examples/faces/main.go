// SPDX-License-Identifier: Apache-2.0 OR MIT

// Command faces lists the faces the server can project pain.001 onto,
// with the side inputs (bindings) each one needs.
//
//	PAIN001_BASE_URL=http://127.0.0.1:8000 go run ./examples/faces
package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	pain001 "github.com/sebastienrousseau/pain001-sdk-go"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "faces:", err)
		os.Exit(1)
	}
}

func run() error {
	client, err := pain001.NewClient(envOr("PAIN001_BASE_URL", "http://127.0.0.1:8000"),
		pain001.WithAPIKey(os.Getenv("PAIN001_API_KEY")))
	if err != nil {
		return err
	}
	faces, err := client.Faces(context.Background())
	if err != nil {
		return err
	}
	for _, f := range faces {
		bindings := "none"
		if len(f.Bindings) > 0 {
			bindings = strings.Join(f.Bindings, ", ")
		}
		fmt.Printf("%-16s bindings: %s\n", f.Name, bindings)
	}
	return nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
