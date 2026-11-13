// SPDX-License-Identifier: Apache-2.0 OR MIT

package pain001

import (
	"encoding/json"
	"errors"
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

// RefusalError is the 422 Generate returns when the server refuses the
// data: rows that fail validation or a scheme rule. It wraps the
// *APIError, so errors.As finds either type and Status stays 422, and
// Response holds the decoded body with its ValidationErrors and
// SchemeViolations.
type RefusalError struct {
	*APIError
	Response GenerateResponse
}

// Error gives the server's message and how many findings it carries,
// never the findings themselves, whose messages can quote the input.
func (e *RefusalError) Error() string {
	return fmt.Sprintf("pain001: HTTP %d: %s (%s, %s)", e.Status, e.Response.Message,
		plural(len(e.Response.ValidationErrors), "validation error"),
		plural(len(e.Response.SchemeViolations), "scheme violation"))
}

// Unwrap returns the underlying *APIError.
func (e *RefusalError) Unwrap() error { return e.APIError }

// plural formats a count with its noun, adding "s" unless it is one.
func plural(n int, noun string) string {
	if n == 1 {
		return "1 " + noun
	}
	return fmt.Sprintf("%d %ss", n, noun)
}

// asRefusal turns a 422 whose body is a generate response with
// "success": false into a *RefusalError; any other error is returned
// as is.
func asRefusal(err error) error {
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Status != 422 {
		return err
	}
	var body struct {
		Success *bool `json:"success"`
	}
	if json.Unmarshal(apiErr.Detail, &body) != nil || body.Success == nil || *body.Success {
		return err
	}
	refusal := &RefusalError{APIError: apiErr}
	if json.Unmarshal(apiErr.Detail, &refusal.Response) != nil {
		return err
	}
	return refusal
}
