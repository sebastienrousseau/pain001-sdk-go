// SPDX-License-Identifier: Apache-2.0 OR MIT

package pain001

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"unicode/utf8"
)

// The fuzz targets treat the server as untrusted: every status and body
// it can send must come back as a value or one of the documented errors,
// never a panic, and an error message must never carry the rejected
// input or a refusal's findings. `go test` runs the seeds; CI fuzzes
// each target for a bounded time (see .github/workflows/ci.yml).

// fuzzSeed is one response the server can send.
type fuzzSeed struct {
	status uint16
	body   string
}

// responseSeeds are the responses the conformance kit's cases expect,
// plus the shapes the client must survive from a misbehaving server.
func responseSeeds(f *testing.F) []fuzzSeed {
	f.Helper()
	xml, err := os.ReadFile("conformance/fixtures/credit-transfer.pain.001.001.09.xml")
	if err != nil {
		f.Fatal(err)
	}
	generated, _ := json.Marshal(map[string]any{"success": true, "message": "Generated", "filename": "payments.xml",
		"xml": string(xml), "validation_errors": []any{}, "scheme_violations": []any{}})
	return []fuzzSeed{
		{200, `{"status":"healthy","version":"0.0.88","message":"ok"}`},
		{200, `{"is_valid":true,"total_rows":5,"valid_rows":5,"invalid_rows":0,"errors":[],"scheme_violations":[]}`},
		{200, `{"is_valid":false,"total_rows":1,"valid_rows":1,"invalid_rows":0,"errors":[],` +
			`"scheme_violations":[{"rule":"UK-CHAPS-PURP","field":"purpose_code","severity":"error"}]}`},
		{400, `{"detail":"Row 1: creditor_account_IBAN 'DE00' is not a valid IBAN"}`},
		{400, `{"detail":"Row 1: debtor_lei is not a valid LEI (P001-E022)"}`},
		{422, `{"detail":[{"loc":["body","message_type"],"msg":"Input should be 'pain.001.001.03'","type":"enum","input":"pain.999"}]}`},
		{200, string(generated)},
		{422, refusedRows},
		{422, `{"success":false,"message":"Scheme 'uk-chaps' validation failed","xml":null,"validation_errors":[],` +
			`"scheme_violations":[{"rule":"UK-CHAPS-PURP","field":"purpose_code","severity":"error"}]}`},
		{200, `{"message_type":"pain.001.001.09","passed":true,"errors":0,"warnings":0,"rungs":["xsd"],"findings":[]}`},
		{422, `{"detail":"not a pain.001 or pain.008 Document"}`},
		{200, `[{"name":"berlin_group","description":"d","bindings":[]},{"name":"uk_open_banking","bindings":["ConsentId"]}]`},
		{200, `{"face":"berlin_group","version":"pain.001.001.09","payload":{"debtorAccount":{"iban":"DE07512108001245126162"}},` +
			`"loss_report":{"face":"berlin_group","total_source_elements":3,"mapped_count":3,"is_lossless":true,"loss_ratio":0}}`},
		{502, `Bad Gateway`}, {500, "a\xffb"}, {204, ``}, {200, `{"a":`}, {200, `null`}, {200, `1e999`},
		{302, `{}`}, {422, `{"success":false,"validation_errors":"not a list"}`}, {200, `{"success":false}`},
	}
}

// addSeeds registers every response seed with f.
func addSeeds(f *testing.F) {
	for _, s := range responseSeeds(f) {
		f.Add(s.status, []byte(s.body))
	}
}

// fuzzStatus maps any fuzzed number onto a three-digit HTTP status,
// keeping one that already is.
func fuzzStatus(n uint16) int {
	if s := int(n) % 1000; s >= 100 {
		return s
	}
	return 100 + int(n)%100
}

// fuzzClient returns a client whose transport answers every request
// with status and body, without a network.
func fuzzClient(status int, body []byte) *Client {
	hc := &http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: status, Header: http.Header{}, Request: r,
			Body: io.NopCloser(bytes.NewReader(body))}, nil
	})}
	c, err := NewClient("http://fuzz.test", WithHTTPClient(hc), WithTimeout(0))
	if err != nil {
		panic(err)
	}
	return c
}

// FuzzRequest feeds arbitrary responses to Request, which must return
// the server's status and its JSON, and an error exactly when the body
// is neither empty nor JSON.
func FuzzRequest(f *testing.F) {
	addSeeds(f)
	f.Fuzz(func(t *testing.T, n uint16, body []byte) {
		status := fuzzStatus(n)
		got, doc, err := fuzzClient(status, body).Request(context.Background(), "GET", "/api/v1/health", nil)
		trimmed := bytes.TrimSpace(body)
		blank := len(trimmed) == 0
		switch {
		case (err != nil) != (!blank && !json.Valid(body)):
			t.Fatalf("HTTP %d %q: err %v", status, body, err)
		case err == nil && got != status:
			t.Fatalf("status %d, sent %d", got, status)
		case err == nil && doc == nil && !blank && string(trimmed) != "null":
			t.Fatalf("HTTP %d %q: no document", status, body)
		}
	})
}

// FuzzTypedMethods feeds arbitrary responses to every typed method. Each
// returns a value or an error, never both; a non-2xx status is always an
// *APIError with that status; a 2xx is never one.
func FuzzTypedMethods(f *testing.F) {
	addSeeds(f)
	f.Fuzz(func(t *testing.T, n uint16, body []byte) {
		status := fuzzStatus(n)
		c, ctx := fuzzClient(status, body), context.Background()
		h, err := c.Health(ctx)
		checkOutcome(t, status, "Health", h != nil, err)
		v, err := c.Validate(ctx, "a,b", "p.csv", Pain00100109, "")
		checkOutcome(t, status, "Validate", v != nil, err)
		g, err := c.Generate(ctx, "a,b", "p.csv", Pain00100109, "")
		checkOutcome(t, status, "Generate", g != nil, err)
		checkGenerate(t, status, g, err)
		ch, err := c.Check(ctx, "<x/>", "", "")
		checkOutcome(t, status, "Check", ch != nil, err)
		_, err = c.Faces(ctx)
		checkOutcome(t, status, "Faces", err == nil, err)
		p, err := c.ProjectFace(ctx, "berlin_group", "<x/>", "", nil)
		checkOutcome(t, status, "ProjectFace", p != nil, err)
	})
}

// checkOutcome holds the contract every typed method shares.
func checkOutcome(t *testing.T, status int, name string, gotValue bool, err error) {
	t.Helper()
	if gotValue == (err != nil) {
		t.Fatalf("%s: value %v with error %v", name, gotValue, err)
	}
	var apiErr *APIError
	isAPI := errors.As(err, &apiErr)
	success := status >= 200 && status <= 299
	if success == isAPI || (isAPI && apiErr.Status != status) {
		t.Fatalf("%s: HTTP %d gave %T %v", name, status, err, err)
	}
	if err != nil && !strings.HasPrefix(err.Error(), "pain001: ") {
		t.Fatalf("%s: message %q lacks the pain001 prefix", name, err)
	}
}

// checkGenerate holds Generate's own promises: a refusal is a 422 with
// success false, and a nil error always comes with the XML.
func checkGenerate(t *testing.T, status int, res *GenerateResponse, err error) {
	t.Helper()
	var refusal *RefusalError
	if errors.As(err, &refusal) && (refusal.Status != 422 || status != 422 || refusal.Response.Success) {
		t.Fatalf("refusal from HTTP %d: %+v", status, refusal.Response)
	}
	if err == nil && (!res.Success || res.XML == "") {
		t.Fatalf("HTTP %d: nil error without the document: %+v", status, res)
	}
}

// FuzzAPIError feeds arbitrary error bodies to the error mapping. The
// message names the status and is valid UTF-8, Detail is empty or valid
// JSON, and a non-JSON body is kept to its first maxDetailBytes.
func FuzzAPIError(f *testing.F) {
	addSeeds(f)
	f.Fuzz(func(t *testing.T, n uint16, body []byte) {
		status := fuzzStatus(n)
		e := newAPIError(status, body)
		msg := e.Error()
		if !strings.HasPrefix(msg, fmt.Sprintf("pain001: HTTP %d", status)) || !utf8.ValidString(msg) {
			t.Fatalf("message %q", msg)
		}
		if len(e.Detail) > 0 && !json.Valid(e.Detail) {
			t.Fatalf("detail %q is not JSON", e.Detail)
		}
		if text, ok := e.DetailText(); ok && !json.Valid(body) && utf8.RuneCountInString(text) > maxDetailBytes {
			t.Fatalf("kept %d runes of a non-JSON body", utf8.RuneCountInString(text))
		}
		if wrapped := asRefusal(fmt.Errorf("w: %w", e)); !errors.Is(wrapped, e) {
			t.Fatalf("asRefusal lost the *APIError: %v", wrapped)
		}
	})
}

// FuzzErrorOmitsRejectedInput checks that a 422 validation array's
// message is the same with or without the rejected "input", which can
// hold payment data.
func FuzzErrorOmitsRejectedInput(f *testing.F) {
	f.Add("message_type", "Input should be 'pain.001.001.03'", "pain.999")
	f.Add("content", "Field required", "DE07512108001245126162")
	f.Add("", "", "")
	f.Fuzz(func(t *testing.T, loc, msg, input string) {
		issue := map[string]any{"loc": []any{"body", loc}, "msg": msg, "type": "t"}
		without, _ := json.Marshal(map[string]any{"detail": []any{issue}})
		issue["input"] = input
		with, _ := json.Marshal(map[string]any{"detail": []any{issue}})
		if got, want := newAPIError(422, with).Error(), newAPIError(422, without).Error(); got != want {
			t.Fatalf("the input changed the message: %q, want %q", got, want)
		}
	})
}

// FuzzRefusalOmitsFindings checks that a refusal's message is the
// server's message and the finding counts, never the findings, whose
// values and messages can quote the input.
func FuzzRefusalOmitsFindings(f *testing.F) {
	f.Add("Validation failed: 4/5 rows valid", "$.currency", "'EURO' does not match '^[A-Z]{3}$'", "EURO", "UK-CHAPS-PURP")
	f.Add("", "", "", "", "")
	f.Fuzz(func(t *testing.T, message, field, rowMsg, value, rule string) {
		body, _ := json.Marshal(map[string]any{"success": false, "message": message, "xml": nil,
			"validation_errors": []any{map[string]any{"field": field, "message": rowMsg, "value": value}},
			"scheme_violations": []any{map[string]any{"rule": rule, "message": rowMsg}}})
		_, err := fuzzClient(422, body).Generate(context.Background(), "a,b", "p.csv", "", "")
		var refusal *RefusalError
		if !errors.As(err, &refusal) {
			t.Fatalf("err %T %v, want a *RefusalError", err, err)
		}
		var sent struct{ Message string }
		_ = json.Unmarshal(body, &sent)
		want := "pain001: HTTP 422: " + sent.Message + " (1 validation error, 1 scheme violation)"
		if err.Error() != want {
			t.Fatalf("message %q\nwant    %q", err.Error(), want)
		}
	})
}
