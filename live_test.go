// SPDX-License-Identifier: Apache-2.0 OR MIT

package pain001_test

import (
	"os"
	"path/filepath"
	"testing"

	pain001 "github.com/sebastienrousseau/pain001-sdk-go"
)

// liveClient returns a client for PAIN001_BASE_URL, or skips the test
// when it is unset. Start a server with `pain001 serve` to run these.
func liveClient(t *testing.T) *pain001.Client {
	t.Helper()
	base := os.Getenv("PAIN001_BASE_URL")
	if base == "" {
		t.Skip("PAIN001_BASE_URL is unset; start `pain001 serve` and set it to run the live tests")
	}
	c, err := pain001.NewClient(base, pain001.WithAPIKey(os.Getenv("PAIN001_API_KEY")))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// fixture returns the text of a kit fixture.
func fixture(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(kitDir, "fixtures", name))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
