// SPDX-License-Identifier: Apache-2.0 OR MIT

package pain001

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestNewAPIErrorMapsEachBodyShape(t *testing.T) {
	long := strings.Repeat("x", 2000)
	cases := []struct {
		name, body, detail, msg string
	}{
		{"string detail", `{"detail":"Unknown face"}`, `"Unknown face"`, "pain001: HTTP 404: Unknown face"},
		{"array detail", `{"detail":[{"loc":["body","message_type"],"msg":"Input should be x","type":"enum","input":"SECRET"}]}`,
			`[{"loc":["body","message_type"],"msg":"Input should be x","type":"enum","input":"SECRET"}]`,
			"pain001: HTTP 404: body.message_type: Input should be x"},
		{"no detail", `{"error":"x"}`, `{"error":"x"}`, "pain001: HTTP 404"},
		{"null detail", `{"detail":null}`, `null`, "pain001: HTTP 404"},
		{"not JSON", `Bad Gateway`, `"Bad Gateway"`, "pain001: HTTP 404: Bad Gateway"},
		{"empty", ``, ``, "pain001: HTTP 404"},
		{"long text", long, `"` + long[:maxDetailBytes] + `"`, "pain001: HTTP 404: " + long[:maxDetailBytes]},
		{"invalid UTF-8", "a\xffb", `"a` + "�" + `b"`, "pain001: HTTP 404: a�b"},
	}
	for _, c := range cases {
		e := newAPIError(404, []byte(c.body))
		if string(e.Detail) != c.detail {
			t.Errorf("%s: detail %s", c.name, e.Detail)
		}
		if e.Error() != c.msg {
			t.Errorf("%s: message %q", c.name, e.Error())
		}
		if strings.Contains(e.Error(), "SECRET") {
			t.Errorf("%s: the rejected input leaked into the message", c.name)
		}
	}
}

func TestAPIErrorWorksWithErrorsAs(t *testing.T) {
	var err error = fmt.Errorf("wrapped: %w", newAPIError(422, []byte(`{"detail":[{"loc":["body",0],"msg":"m","type":"t"}]}`)))
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Status != 422 {
		t.Fatalf("errors.As failed on %v", err)
	}
	issues := apiErr.Issues()
	if len(issues) != 1 || formatLoc(issues[0].Loc) != "body.0" || issues[0].Type != "t" {
		t.Errorf("issues %+v", issues)
	}
	if _, ok := apiErr.DetailText(); ok {
		t.Error("an array detail is not text")
	}
	text := &APIError{Status: 400, Detail: json.RawMessage(`"x"`)}
	if s, ok := text.DetailText(); !ok || s != "x" || text.Issues() != nil {
		t.Errorf("text detail %q %v", s, ok)
	}
}
