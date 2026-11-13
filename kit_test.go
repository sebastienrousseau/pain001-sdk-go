// SPDX-License-Identifier: Apache-2.0 OR MIT

package pain001_test

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	pain001 "github.com/sebastienrousseau/pain001-sdk-go"
)

const kitDir = "conformance"

// TestVendoredKitMatchesItsPinnedSums fails when any kit file changes,
// appears or disappears without conformance/SHA256SUMS being updated, so
// the kit is only ever replaced by re-vendoring it from core.
func TestVendoredKitMatchesItsPinnedSums(t *testing.T) {
	pinned := readSums(t)
	var onDisk []string
	err := filepath.WalkDir(kitDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(kitDir, path)
		if rel = filepath.ToSlash(rel); rel != "SHA256SUMS" {
			onDisk = append(onDisk, rel)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, len(pinned))
	for name := range pinned {
		names = append(names, name)
	}
	slices.Sort(names)
	slices.Sort(onDisk)
	if !slices.Equal(names, onDisk) {
		t.Fatalf("kit files %q, SHA256SUMS lists %q", onDisk, names)
	}
	for name, want := range pinned {
		data, err := os.ReadFile(filepath.Join(kitDir, filepath.FromSlash(name)))
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(data)
		if got := hex.EncodeToString(sum[:]); got != want {
			t.Errorf("%s: sha256 %s, pinned %s", name, got, want)
		}
	}
}

// readSums parses conformance/SHA256SUMS ("<hex>  <path>", sorted).
func readSums(t *testing.T) map[string]string {
	t.Helper()
	f, err := os.Open(filepath.Join(kitDir, "SHA256SUMS"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	sums := map[string]string{}
	var order []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		sum, name, ok := strings.Cut(scanner.Text(), "  ")
		if !ok || len(sum) != 64 {
			t.Fatalf("malformed SHA256SUMS line %q", scanner.Text())
		}
		sums[name] = sum
		order = append(order, name)
	}
	if !slices.IsSorted(order) {
		t.Error("SHA256SUMS must be sorted by path")
	}
	return sums
}

// TestMessageTypesMatchTheKitsOpenAPI keeps the constants in step with
// the server's MessageType enum in the vendored openapi.json.
func TestMessageTypesMatchTheKitsOpenAPI(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(kitDir, "openapi.json"))
	if err != nil {
		t.Fatal(err)
	}
	var spec struct {
		Components struct {
			Schemas struct {
				MessageType struct {
					Enum []string `json:"enum"`
				} `json:"MessageType"`
			} `json:"schemas"`
		} `json:"components"`
	}
	if err := json.Unmarshal(data, &spec); err != nil {
		t.Fatal(err)
	}
	var ours []string
	for _, mt := range pain001.MessageTypes() {
		ours = append(ours, string(mt))
	}
	if want := spec.Components.Schemas.MessageType.Enum; !slices.Equal(ours, want) {
		t.Errorf("MessageTypes() = %q, openapi enum %q", ours, want)
	}
}
