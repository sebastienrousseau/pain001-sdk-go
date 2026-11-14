// SPDX-License-Identifier: Apache-2.0 OR MIT

package pain001_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	pain001 "github.com/sebastienrousseau/pain001-sdk-go"
)

// Two kinds of benchmark, one sub-benchmark per public method:
//
//   - BenchmarkClient: the client's own cost with no network. A canned
//     transport answers with a response recorded from pain001 serve
//     (testdata/bench, scripts/record-bench-responses.sh), so each
//     iteration builds the request, encodes the kit fixture into it,
//     and reads and decodes the response.
//   - BenchmarkLive: one loopback round trip to the server at
//     PAIN001_BASE_URL, skipped when it is unset.
//
// `make bench` runs both; the output is benchstat's input format.

// benchCall is one method under benchmark.
type benchCall struct {
	name string
	run  func(ctx context.Context, c *pain001.Client) error
}

// benchCalls returns every public method with the kit's fixtures.
func benchCalls(b *testing.B) []benchCall {
	b.Helper()
	csv := readFile(b, filepath.Join(kitDir, "fixtures", "credit-transfer.pain.001.001.09.csv"))
	xml := readFile(b, filepath.Join(kitDir, "fixtures", "credit-transfer.pain.001.001.09.xml"))
	return []benchCall{
		{"Health", func(ctx context.Context, c *pain001.Client) error { _, err := c.Health(ctx); return err }},
		{"Validate", func(ctx context.Context, c *pain001.Client) error {
			_, err := c.Validate(ctx, csv, "payments.csv", pain001.Pain00100109, "")
			return err
		}},
		{"Generate", func(ctx context.Context, c *pain001.Client) error {
			_, err := c.Generate(ctx, csv, "payments.csv", pain001.Pain00100109, "")
			return err
		}},
		{"Check", func(ctx context.Context, c *pain001.Client) error { _, err := c.Check(ctx, xml, "", ""); return err }},
		{"Faces", func(ctx context.Context, c *pain001.Client) error { _, err := c.Faces(ctx); return err }},
		{"ProjectFace", func(ctx context.Context, c *pain001.Client) error {
			_, err := c.ProjectFace(ctx, "berlin_group", xml, "", nil)
			return err
		}},
		{"Request", func(ctx context.Context, c *pain001.Client) error {
			_, _, err := c.Request(ctx, http.MethodGet, "/api/v1/health", nil)
			return err
		}},
	}
}

// recordedFor names the recorded response each method decodes.
var recordedFor = map[string]string{
	"Health": "health", "Validate": "validate", "Generate": "generate", "Check": "check",
	"Faces": "faces", "ProjectFace": "projectface", "Request": "health",
}

func readFile(b *testing.B, path string) string {
	b.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		b.Fatal(err)
	}
	return string(data)
}

// canned answers every request with status 200 and body, draining the
// request body first as a server would.
type canned []byte

func (body canned) RoundTrip(r *http.Request) (*http.Response, error) {
	if r.Body != nil {
		_, _ = io.Copy(io.Discard, r.Body)
		_ = r.Body.Close()
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": {"application/json"}},
		Body:       io.NopCloser(bytes.NewReader(body)),
		Request:    r,
	}, nil
}

func BenchmarkClient(b *testing.B) {
	ctx := context.Background()
	for _, call := range benchCalls(b) {
		body := readFile(b, filepath.Join("testdata", "bench", recordedFor[call.name]+".json"))
		c, err := pain001.NewClient("http://127.0.0.1:8000", pain001.WithHTTPClient(&http.Client{Transport: canned(body)}))
		if err != nil {
			b.Fatal(err)
		}
		b.Run(call.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				if err := call.run(ctx, c); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkLive(b *testing.B) {
	base := os.Getenv("PAIN001_BASE_URL")
	if base == "" {
		b.Skip("PAIN001_BASE_URL is unset; start `pain001 serve` and set it to run the round-trip benchmarks")
	}
	c, err := pain001.NewClient(base, pain001.WithAPIKey(os.Getenv("PAIN001_API_KEY")))
	if err != nil {
		b.Fatal(err)
	}
	ctx := context.Background()
	for _, call := range benchCalls(b) {
		b.Run(call.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				if err := call.run(ctx, c); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
