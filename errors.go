// SPDX-License-Identifier: Apache-2.0 OR MIT

package pain001

import (
	"encoding/json"
	"fmt"
	"strings"
)

// maxDetailBytes caps a non-JSON error body kept as Detail.
const maxDetailBytes = 1024

// APIError is a non-2xx response from the server. Detail is the
// response's "detail" member as raw JSON: a string for most errors, an
// array of validation issues for a 422 request-shape error. When the body
// has no "detail", Detail holds the whole JSON body, or the start of a
// non-JSON body as a JSON string.
type APIError struct {
	Status int
	Detail json.RawMessage
}

// Issue is one entry of a 422 detail array.
type Issue struct {
	Loc  []any  `json:"loc"`
	Msg  string `json:"msg"`
	Type string `json:"type"`
}

// Error describes the failure. For a 422 array it lists each issue's
// location and message, and never the rejected input, which can hold
// payment data.
func (e *APIError) Error() string {
	if text, ok := e.DetailText(); ok {
		return fmt.Sprintf("pain001: HTTP %d: %s", e.Status, text)
	}
	issues := e.Issues()
	if len(issues) == 0 {
		return fmt.Sprintf("pain001: HTTP %d", e.Status)
	}
	parts := make([]string, len(issues))
	for i, is := range issues {
		parts[i] = formatLoc(is.Loc) + ": " + is.Msg
	}
	return fmt.Sprintf("pain001: HTTP %d: %s", e.Status, strings.Join(parts, "; "))
}

// DetailText returns Detail when it is a JSON string.
func (e *APIError) DetailText() (string, bool) {
	var s *string
	if err := json.Unmarshal(e.Detail, &s); err != nil || s == nil {
		return "", false
	}
	return *s, true
}

// Issues returns Detail decoded as a validation-issue array, or nil when
// it is not one.
func (e *APIError) Issues() []Issue {
	var issues []Issue
	if err := json.Unmarshal(e.Detail, &issues); err != nil {
		return nil
	}
	return issues
}

// formatLoc joins a location such as ["body", "message_type"] with dots.
func formatLoc(loc []any) string {
	parts := make([]string, len(loc))
	for i, p := range loc {
		parts[i] = fmt.Sprint(p)
	}
	return strings.Join(parts, ".")
}

// newAPIError builds the error for a non-2xx status from its raw body.
func newAPIError(status int, raw []byte) *APIError {
	e := &APIError{Status: status}
	var env struct {
		Detail json.RawMessage `json:"detail"`
	}
	switch {
	case json.Unmarshal(raw, &env) == nil && len(env.Detail) > 0:
		e.Detail = env.Detail
	case json.Valid(raw):
		e.Detail = json.RawMessage(raw)
	case len(raw) > 0:
		if len(raw) > maxDetailBytes {
			raw = raw[:maxDetailBytes]
		}
		e.Detail, _ = json.Marshal(strings.ToValidUTF8(string(raw), "�"))
	}
	return e
}
