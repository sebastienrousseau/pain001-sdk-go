// SPDX-License-Identifier: Apache-2.0 OR MIT

// Package kit loads and runs the pain001 conformance kit (conformance/
// at the repository root). It ports the semantics of core's reference
// runner, tests/test_conformance_kit.py: $file substitution, an exact
// status, then equals, types and contains checks on RFC 6901 pointers.
package kit

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Kit is a loaded cases.json.
type Kit struct {
	Dir        string
	KitVersion int    `json:"kit_version"`
	Cases      []Case `json:"cases"`
}

// Case is one request and what its response must satisfy.
type Case struct {
	ID     string `json:"id"`
	Method string `json:"method"`
	Path   string `json:"path"`
	Body   any    `json:"body"`
	Expect Expect `json:"expect"`
}

// Expect is a case's expectations; the map keys are JSON Pointers.
type Expect struct {
	Status   int                 `json:"status"`
	Equals   map[string]any      `json:"equals"`
	Types    map[string]string   `json:"types"`
	Contains map[string][]string `json:"contains"`
}

// Sender performs one call and returns the status and the decoded JSON
// body, with numbers as json.Number. Client.Request has this shape.
type Sender func(ctx context.Context, method, path string, body any) (int, any, error)

// Load reads dir/cases.json. Numbers decode as json.Number so a case's
// integers and decimals keep their JSON form.
func Load(dir string) (*Kit, error) {
	data, err := os.ReadFile(filepath.Join(dir, "cases.json"))
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	k := &Kit{Dir: dir}
	if err := dec.Decode(k); err != nil {
		return nil, fmt.Errorf("kit: cases.json: %w", err)
	}
	if len(k.Cases) == 0 {
		return nil, errors.New("kit: cases.json has no cases")
	}
	return k, nil
}

// Resolve returns body with every {"$file": "<path>"} object replaced
// by that file's text, read relative to dir. Paths must stay inside dir.
func Resolve(dir string, body any) (any, error) {
	switch v := body.(type) {
	case map[string]any:
		if ref, ok := v["$file"]; ok && len(v) == 1 {
			return readFixture(dir, ref)
		}
		return resolveObject(dir, v)
	case []any:
		return resolveArray(dir, v)
	default:
		return body, nil
	}
}

func resolveObject(dir string, v map[string]any) (any, error) {
	out := make(map[string]any, len(v))
	for key, item := range v {
		r, err := Resolve(dir, item)
		if err != nil {
			return nil, err
		}
		out[key] = r
	}
	return out, nil
}

func resolveArray(dir string, v []any) (any, error) {
	out := make([]any, len(v))
	for i, item := range v {
		r, err := Resolve(dir, item)
		if err != nil {
			return nil, err
		}
		out[i] = r
	}
	return out, nil
}

// readFixture reads one $file reference as UTF-8 text.
func readFixture(dir string, ref any) (string, error) {
	name, ok := ref.(string)
	if !ok || !filepath.IsLocal(filepath.FromSlash(name)) {
		return "", fmt.Errorf("kit: $file %v is not a path inside the kit", ref)
	}
	data, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(name)))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Run sends one case through send and returns the failed expectations;
// an empty result is a pass. An error means the case could not be sent.
func (k *Kit) Run(ctx context.Context, send Sender, c Case) ([]string, error) {
	body, err := Resolve(k.Dir, c.Body)
	if err != nil {
		return nil, err
	}
	status, doc, err := send(ctx, c.Method, c.Path, body)
	if err != nil {
		return nil, err
	}
	return Evaluate(c.Expect, status, doc), nil
}
