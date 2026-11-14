// SPDX-License-Identifier: Apache-2.0 OR MIT

// Command projectface projects the conformance kit's pain.001.001.09
// document onto the Berlin Group face and prints the target payload's
// size and the loss report: how much of the source the face carries.
//
//	PAIN001_BASE_URL=http://127.0.0.1:8000 go run ./examples/projectface
//
// Run it from the repository root so the fixture path resolves.
package main

import (
	"context"
	"fmt"
	"os"

	pain001 "github.com/sebastienrousseau/pain001-sdk-go"
)

const fixture = "conformance/fixtures/credit-transfer.pain.001.001.09.xml"

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "projectface:", err)
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
	// An empty version lets the server read the edition from the
	// namespace; Berlin Group needs no bindings, so they are nil.
	res, err := client.ProjectFace(context.Background(), "berlin_group", string(xml), "", nil)
	if err != nil {
		return err
	}
	lr := res.LossReport
	fmt.Printf("%s from %s: %d payload bytes\n", res.Face, res.Version, len(res.Payload))
	fmt.Printf("mapped %d of %d source elements, lossless=%t, loss ratio %.3f\n",
		lr.MappedCount, lr.TotalSourceElements, lr.IsLossless, lr.LossRatio)
	return nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
