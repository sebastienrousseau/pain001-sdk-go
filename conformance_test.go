// SPDX-License-Identifier: Apache-2.0 OR MIT

package pain001_test

import (
	"context"
	"testing"

	"github.com/sebastienrousseau/pain001-sdk-go/internal/kit"
)

// TestConformance runs every case in conformance/cases.json through
// Client.Request against the server at PAIN001_BASE_URL.
func TestConformance(t *testing.T) {
	k, err := kit.Load(kitDir)
	if err != nil {
		t.Fatal(err)
	}
	client := liveClient(t)
	for _, c := range k.Cases {
		t.Run(c.ID, func(t *testing.T) {
			failures, err := k.Run(context.Background(), client.Request, c)
			if err != nil {
				t.Fatalf("%s %s: %v", c.Method, c.Path, err)
			}
			for _, f := range failures {
				t.Error(f)
			}
		})
	}
}
