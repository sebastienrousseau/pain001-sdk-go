// SPDX-License-Identifier: Apache-2.0 OR MIT

package kit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

// Evaluate checks a response against expect and returns the failures.
// A status mismatch is the only failure reported, as in the reference
// runner. Pointers are checked in sorted order, so the list is stable.
func Evaluate(expect Expect, status int, doc any) []string {
	if status != expect.Status {
		return []string{fmt.Sprintf("status %d != %d", status, expect.Status)}
	}
	var failures []string
	for _, ptr := range sortedKeys(expect.Equals) {
		failures = append(failures, checkEquals(doc, ptr, expect.Equals[ptr])...)
	}
	for _, ptr := range sortedKeys(expect.Types) {
		failures = append(failures, checkType(doc, ptr, expect.Types[ptr])...)
	}
	for _, ptr := range sortedKeys(expect.Contains) {
		failures = append(failures, checkContains(doc, ptr, expect.Contains[ptr])...)
	}
	return failures
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

func checkEquals(doc any, ptr string, want any) []string {
	got, ok := Pointer(doc, ptr)
	if !ok {
		return []string{ptr + " missing"}
	}
	if !Equal(got, want) {
		return []string{fmt.Sprintf("%s != %s", ptr, show(want))}
	}
	return nil
}

func checkType(doc any, ptr, kind string) []string {
	got, ok := Pointer(doc, ptr)
	if !ok {
		return []string{ptr + " missing"}
	}
	is, known := IsType(got, kind)
	if !known {
		return []string{fmt.Sprintf("%s: unknown type %q", ptr, kind)}
	}
	if !is {
		return []string{ptr + " is not " + kind}
	}
	return nil
}

func checkContains(doc any, ptr string, needles []string) []string {
	got, ok := Pointer(doc, ptr)
	if !ok {
		return []string{ptr + " missing"}
	}
	text, ok := got.(string)
	if !ok {
		return []string{ptr + " is not a string"}
	}
	var failures []string
	for _, n := range needles {
		if !strings.Contains(text, n) {
			failures = append(failures, fmt.Sprintf("%s lacks %s", ptr, show(n)))
		}
	}
	return failures
}

// show renders a value for a failure message as JSON, leaving <, > and
// & unescaped so XML needles read as written.
func show(v any) string {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return fmt.Sprint(v)
	}
	return strings.TrimSuffix(buf.String(), "\n")
}

// Pointer looks up an RFC 6901 JSON Pointer in a decoded document. ""
// is the whole document. Like the reference runner, an array index is a
// run of ASCII digits below the array's length.
func Pointer(doc any, ptr string) (any, bool) {
	if ptr == "" {
		return doc, true
	}
	if !strings.HasPrefix(ptr, "/") {
		return nil, false
	}
	value := doc
	for _, raw := range strings.Split(ptr, "/")[1:] {
		token := strings.ReplaceAll(strings.ReplaceAll(raw, "~1", "/"), "~0", "~")
		next, ok := step(value, token)
		if !ok {
			return nil, false
		}
		value = next
	}
	return value, true
}

// step descends one pointer token into an array or object.
func step(value any, token string) (any, bool) {
	switch v := value.(type) {
	case []any:
		if !isDigits(token) {
			return nil, false
		}
		i, err := strconv.Atoi(token)
		if err != nil || i >= len(v) {
			return nil, false
		}
		return v[i], true
	case map[string]any:
		item, ok := v[token]
		return item, ok
	default:
		return nil, false
	}
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// IsType reports whether v is of the kit type kind; known is false for a
// kind the kit does not define. An integer is a JSON number written with
// no fraction or exponent, as Python's json module reads it.
func IsType(v any, kind string) (is, known bool) {
	switch kind {
	case "string":
		_, is = v.(string)
	case "number":
		_, is = v.(json.Number)
	case "integer":
		n, ok := v.(json.Number)
		is = ok && !strings.ContainsAny(n.String(), ".eE")
	case "boolean":
		_, is = v.(bool)
	case "array":
		_, is = v.([]any)
	case "object":
		_, is = v.(map[string]any)
	case "null":
		is = v == nil
	default:
		return false, false
	}
	return is, true
}

// Equal is deep equality on decoded JSON, comparing numbers by value so
// 5 equals 5.0, as in Python.
func Equal(a, b any) bool {
	switch x := a.(type) {
	case json.Number:
		y, ok := b.(json.Number)
		return ok && numbersEqual(x, y)
	case []any:
		y, ok := b.([]any)
		return ok && slices.EqualFunc(x, y, Equal)
	case map[string]any:
		y, ok := b.(map[string]any)
		return ok && mapsEqual(x, y)
	default:
		return reflect.DeepEqual(a, b)
	}
}

func numbersEqual(a, b json.Number) bool {
	x, okx := new(big.Rat).SetString(a.String())
	y, oky := new(big.Rat).SetString(b.String())
	if !okx || !oky {
		return a == b
	}
	return x.Cmp(y) == 0
}

func mapsEqual(a, b map[string]any) bool {
	if len(a) != len(b) {
		return false
	}
	for k, va := range a {
		vb, ok := b[k]
		if !ok || !Equal(va, vb) {
			return false
		}
	}
	return true
}
