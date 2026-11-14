// SPDX-License-Identifier: Apache-2.0 OR MIT

package pain001

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// recorded is what the stub server saw of one request.
type recorded struct {
	method, path, auth, contentType, accept string
	body                                    map[string]any
}

// stub serves status and reply to every request and records it.
func stub(t *testing.T, status int, reply string) (*Client, *recorded) {
	t.Helper()
	rec := &recorded{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec.method, rec.path = r.Method, r.URL.EscapedPath()
		rec.auth = r.Header.Get("Authorization")
		rec.contentType, rec.accept = r.Header.Get("Content-Type"), r.Header.Get("Accept")
		data, _ := io.ReadAll(r.Body)
		rec.body = nil
		if len(data) > 0 {
			if err := json.Unmarshal(data, &rec.body); err != nil {
				t.Errorf("request body is not a JSON object: %v", err)
			}
		}
		w.WriteHeader(status)
		_, _ = io.WriteString(w, reply)
	}))
	t.Cleanup(srv.Close)
	c, err := NewClient(srv.URL+"/", WithAPIKey("k3y"))
	if err != nil {
		t.Fatal(err)
	}
	return c, rec
}

func TestNewClientRejectsUnusableURLs(t *testing.T) {
	for _, u := range []string{"", "127.0.0.1:8000", "ftp://h", "http://", "http://h\x7f"} {
		if _, err := NewClient(u); err == nil {
			t.Errorf("NewClient(%q) must fail", u)
		}
	}
	c, err := NewClient("https://example.test//", WithHTTPClient(nil), WithTimeout(0))
	if err != nil || c.baseURL != "https://example.test" || c.httpClient == nil || c.timeout != 0 {
		t.Errorf("client %+v, err %v", c, err)
	}
}

func TestTypedMethodsSendTheDocumentedRequests(t *testing.T) {
	ctx := context.Background()
	c, rec := stub(t, 200, `{}`)
	calls := []struct {
		name, method, path string
		run                func() error
		body               map[string]any
	}{
		{"health", "GET", "/api/v1/health", func() error { _, err := c.Health(ctx); return err }, nil},
		{"validate", "POST", "/api/v1/ui/validate", func() error {
			_, err := c.Validate(ctx, "a,b", "p.csv", Pain00100109, "sepa-sct")
			return err
		}, map[string]any{"content": "a,b", "filename": "p.csv", "message_type": "pain.001.001.09", "scheme": "sepa-sct"}},
		{"generate", "POST", "/api/v1/ui/generate", func() error {
			_, err := c.Generate(ctx, "a,b", "p.csv", "", "")
			if errors.Is(err, ErrNoDocument) {
				return nil // the stub's {} carries no XML; the request is what is checked
			}
			return err
		}, map[string]any{"content": "a,b", "filename": "p.csv"}},
		{"check", "POST", "/api/v1/check", func() error {
			_, err := c.Check(ctx, "<x/>", "sepa-sct", "sepa")
			return err
		}, map[string]any{"xml": "<x/>", "scheme": "sepa-sct", "schematron": "sepa"}},
		{"face", "POST", "/api/v1/faces/a%2Fb%20c", func() error {
			_, err := c.ProjectFace(ctx, "a/b c", "<x/>", "pain.001.001.09", map[string]any{"ConsentId": "c1"})
			return err
		}, map[string]any{"xml": "<x/>", "version": "pain.001.001.09", "bindings": map[string]any{"ConsentId": "c1"}}},
	}
	for _, call := range calls {
		if err := call.run(); err != nil {
			t.Fatalf("%s: %v", call.name, err)
		}
		if rec.method != call.method || rec.path != call.path {
			t.Errorf("%s: sent %s %s", call.name, rec.method, rec.path)
		}
		if rec.auth != "Bearer k3y" || rec.accept != "application/json" {
			t.Errorf("%s: auth %q accept %q", call.name, rec.auth, rec.accept)
		}
		if got, _ := json.Marshal(rec.body); string(got) != mustJSON(t, call.body) {
			t.Errorf("%s: body %s", call.name, got)
		}
		if (call.body != nil) != (rec.contentType == "application/json") {
			t.Errorf("%s: content type %q", call.name, rec.contentType)
		}
	}
}

func mustJSON(t *testing.T, v any) string {
	t.Helper()
	if m, ok := v.(map[string]any); ok && m == nil {
		return "null"
	}
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestNoAuthorizationHeaderWithoutAKey(t *testing.T) {
	c, rec := stub(t, 200, `{"status":"healthy"}`)
	c.apiKey = ""
	if _, err := c.Health(context.Background()); err != nil || rec.auth != "" {
		t.Errorf("auth %q, err %v", rec.auth, err)
	}
}

func TestFacesDecodesTheList(t *testing.T) {
	c, _ := stub(t, 200, `[{"name":"uk_open_banking","description":"d","bindings":["ConsentId","Risk"]}]`)
	faces, err := c.Faces(context.Background())
	if err != nil || len(faces) != 1 || faces[0].Bindings[1] != "Risk" {
		t.Errorf("faces %+v, err %v", faces, err)
	}
}

func TestTypedMethodsReturnAPIErrors(t *testing.T) {
	ctx := context.Background()
	c, _ := stub(t, 400, `{"detail":"bad IBAN"}`)
	runs := map[string]func() error{
		"health":   func() error { _, err := c.Health(ctx); return err },
		"validate": func() error { _, err := c.Validate(ctx, "", "p.csv", "", ""); return err },
		"generate": func() error { _, err := c.Generate(ctx, "", "p.csv", "", ""); return err },
		"check":    func() error { _, err := c.Check(ctx, "<x/>", "", ""); return err },
		"faces":    func() error { _, err := c.Faces(ctx); return err },
		"face":     func() error { _, err := c.ProjectFace(ctx, "f", "<x/>", "", nil); return err },
	}
	for name, run := range runs {
		var apiErr *APIError
		if err := run(); !errors.As(err, &apiErr) || apiErr.Status != 400 {
			t.Errorf("%s: err %v", name, err)
		}
	}
}

func TestUndecodableSuccessIsAnError(t *testing.T) {
	c, _ := stub(t, 200, `not json`)
	if _, err := c.Health(context.Background()); err == nil || !strings.Contains(err.Error(), "decoding response") {
		t.Errorf("err %v", err)
	}
}

func TestRequestReturnsStatusAndJSONForAnyStatus(t *testing.T) {
	c, _ := stub(t, 422, `{"detail":[{"loc":["body"],"msg":"m","type":"t"}],"n":5}`)
	status, doc, err := c.Request(context.Background(), "POST", "/api/v1/check", map[string]any{"xml": ""})
	if err != nil || status != 422 {
		t.Fatalf("status %d, err %v", status, err)
	}
	if n := doc.(map[string]any)["n"]; n != json.Number("5") {
		t.Errorf("numbers must decode as json.Number, got %T %v", n, n)
	}
}

func TestRequestEdgeCases(t *testing.T) {
	ctx := context.Background()
	c, _ := stub(t, 204, ``)
	if status, doc, err := c.Request(ctx, "GET", "/x", nil); status != 204 || doc != nil || err != nil {
		t.Errorf("empty body: %d %v %v", status, doc, err)
	}
	if _, _, err := c.Request(ctx, "GET", "x", nil); err == nil {
		t.Error("a path without a leading slash must fail")
	}
	if _, _, err := c.Request(ctx, "GET", "/x", make(chan int)); err == nil {
		t.Error("an unencodable body must fail")
	}
	if _, _, err := c.Request(ctx, "BAD METHOD", "/x", nil); err == nil {
		t.Error("an invalid method must fail")
	}
	c2, _ := stub(t, 200, `{"a":`)
	if status, _, err := c2.Request(ctx, "GET", "/x", nil); err == nil || status != 200 {
		t.Errorf("truncated JSON: status %d err %v", status, err)
	}
}

func TestRequestRejectsDataAfterTheJSONValue(t *testing.T) {
	for _, body := range []string{`000`, `{"status":"healthy"} trailing`, `{}{}`} {
		c, _ := stub(t, 200, body)
		if _, doc, err := c.Request(context.Background(), "GET", "/x", nil); !errors.Is(err, errTrailingData) {
			t.Errorf("%s: doc %v, err %v, want trailing-data error", body, doc, err)
		}
	}
	c, _ := stub(t, 200, " {\"a\":1} \n")
	if _, doc, err := c.Request(context.Background(), "GET", "/x", nil); err != nil || doc == nil {
		t.Errorf("trailing whitespace: doc %v, err %v", doc, err)
	}
}

func TestResponseSizeIsCapped(t *testing.T) {
	old := maxResponseBytes
	maxResponseBytes = 8
	t.Cleanup(func() { maxResponseBytes = old })
	c, _ := stub(t, 200, `{"status":"healthy"}`)
	if _, _, err := c.Request(context.Background(), "GET", "/x", nil); !errors.Is(err, ErrResponseTooLarge) {
		t.Errorf("err %v", err)
	}
}

func TestTimeoutBoundsTheCall(t *testing.T) {
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-release:
		case <-r.Context().Done():
		}
	}))
	t.Cleanup(func() { close(release); srv.Close() })
	c, err := NewClient(srv.URL, WithTimeout(50*time.Millisecond))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.Health(context.Background()); !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("err %v, want a deadline error", err)
	}
}

// failingBody errors on read, after the headers have arrived.
type failingBody struct{}

func (failingBody) Read([]byte) (int, error) { return 0, errors.New("read failed") }
func (failingBody) Close() error             { return nil }

// roundTrip lets a function act as an http.RoundTripper.
type roundTrip func(*http.Request) (*http.Response, error)

func (f roundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestWithHTTPClientIsUsed(t *testing.T) {
	used := false
	hc := &http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
		used = true
		return &http.Response{StatusCode: 200, Body: failingBody{}, Header: http.Header{}}, nil
	})}
	c, err := NewClient("http://example.test", WithHTTPClient(hc))
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := c.Request(context.Background(), "GET", "/x", nil); err == nil || !used {
		t.Errorf("used %v, err %v", used, err)
	}
}

func TestTransportErrorsAreWrapped(t *testing.T) {
	boom := errors.New("dial failed")
	hc := &http.Client{Transport: roundTrip(func(*http.Request) (*http.Response, error) { return nil, boom })}
	c, _ := NewClient("http://example.test", WithHTTPClient(hc))
	if _, err := c.Health(context.Background()); !errors.Is(err, boom) {
		t.Errorf("err %v", err)
	}
}

func TestMessageTypesAreDistinct(t *testing.T) {
	seen := map[MessageType]bool{}
	for _, mt := range MessageTypes() {
		if seen[mt] || !strings.HasPrefix(string(mt), "pain.00") {
			t.Errorf("bad or duplicate %q", mt)
		}
		seen[mt] = true
	}
}

// refusedRows is core 0.0.87's 422 body for rows that fail validation.
const refusedRows = `{"success":false,"message":"Validation failed: 4/5 rows valid","filename":null,"xml":null,` +
	`"validation_errors":[{"field":"$.currency","message":"'EURO' does not match '^[A-Z]{3}$'","value":"EURO"}],"scheme_violations":[]}`

func TestGenerateRefusalIsATypedError(t *testing.T) {
	c, _ := stub(t, 422, refusedRows)
	res, err := c.Generate(context.Background(), "a,b", "p.csv", Pain00100109, "")
	var refusal *RefusalError
	if res != nil || !errors.As(err, &refusal) {
		t.Fatalf("res %+v, err %T %v, want a *RefusalError", res, err, err)
	}
	checkRowsRefusal(t, refusal)
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Status != 422 {
		t.Errorf("a refusal must still be an *APIError: %v", err)
	}
	want := "pain001: HTTP 422: Validation failed: 4/5 rows valid (1 validation error, 0 scheme violations)"
	if err.Error() != want {
		t.Errorf("message %q\nwant    %q", err.Error(), want)
	}
	if strings.Contains(err.Error(), "EURO") {
		t.Error("the error message must not echo the rejected value")
	}
}

// checkRowsRefusal checks the findings decoded from refusedRows.
func checkRowsRefusal(t *testing.T, refusal *RefusalError) {
	t.Helper()
	if refusal.Status != 422 || refusal.Response.Success || refusal.Response.Message != "Validation failed: 4/5 rows valid" {
		t.Errorf("refusal %+v", refusal.Response)
	}
	rows := refusal.Response.ValidationErrors
	if len(rows) != 1 || rows[0].Field != "$.currency" || string(rows[0].Value) != `"EURO"` {
		t.Errorf("validation errors %+v", rows)
	}
	if refusal.Response.SchemeViolations == nil || len(refusal.Response.SchemeViolations) != 0 {
		t.Errorf("scheme violations %+v", refusal.Response.SchemeViolations)
	}
}

func TestGenerateSchemeRefusalCarriesTheViolations(t *testing.T) {
	c, _ := stub(t, 422, `{"success":false,"message":"Scheme 'uk-chaps' validation failed","xml":null,"validation_errors":[],`+
		`"scheme_violations":[{"rule":"UK-CHAPS-PURP","field":"purpose_code","severity":"error"},{"rule":"R2"}]}`)
	_, err := c.Generate(context.Background(), "a,b", "p.csv", "", "uk-chaps")
	var refusal *RefusalError
	if !errors.As(err, &refusal) || len(refusal.Response.SchemeViolations) != 2 ||
		refusal.Response.SchemeViolations[0]["rule"] != "UK-CHAPS-PURP" {
		t.Fatalf("err %v", err)
	}
	if want := "pain001: HTTP 422: Scheme 'uk-chaps' validation failed (0 validation errors, 2 scheme violations)"; err.Error() != want {
		t.Errorf("message %q", err.Error())
	}
}

func TestGenerateOtherErrorsAreNotRefusals(t *testing.T) {
	for _, reply := range []struct {
		status int
		body   string
	}{
		{422, `{"detail":[{"loc":["body","message_type"],"msg":"m","type":"enum"}]}`},
		{422, `{"success":true,"message":"odd"}`},
		{422, `{"message":"no success member"}`},
		{422, `not json`},
		{422, `{"success":false,"message":"m","validation_errors":"not a list"}`},
		{400, `{"success":false,"message":"refused upload"}`},
		{500, `{"detail":"boom"}`},
	} {
		c, _ := stub(t, reply.status, reply.body)
		_, err := c.Generate(context.Background(), "a,b", "p.csv", "", "")
		var refusal *RefusalError
		var apiErr *APIError
		if errors.As(err, &refusal) || !errors.As(err, &apiErr) || apiErr.Status != reply.status {
			t.Errorf("%d %s: err %T %v", reply.status, reply.body, err, err)
		}
	}
}

func TestGenerateWithoutTheXMLIsAnError(t *testing.T) {
	for _, body := range []string{
		`{"success":false,"message":"Validation failed: 4/5 rows valid","xml":null,"validation_errors":[{"field":"$.currency","value":"EURO"}]}`,
		`{"success":true,"message":"Generated","xml":""}`,
		`{}`,
	} {
		c, _ := stub(t, 200, body)
		res, err := c.Generate(context.Background(), "a,b", "p.csv", "", "")
		var apiErr *APIError
		if res != nil || !errors.Is(err, ErrNoDocument) || errors.As(err, &apiErr) {
			t.Errorf("%s: res %+v, err %T %v", body, res, err, err)
		}
		if err != nil && strings.Contains(err.Error(), "EURO") {
			t.Errorf("the error message must not echo the rejected value: %v", err)
		}
	}
}
