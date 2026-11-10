// SPDX-License-Identifier: Apache-2.0 OR MIT

package kit

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

const kitDir = "../../conformance"

func decode(t *testing.T, text string) any {
	t.Helper()
	dec := json.NewDecoder(strings.NewReader(text))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		t.Fatal(err)
	}
	return v
}

func TestPointerFollowsRFC6901(t *testing.T) {
	doc := decode(t, `{"a/b": {"~": 1}, "list": [10, 20], "": "empty"}`)
	cases := []struct {
		ptr  string
		want string
		ok   bool
	}{
		{"/a~1b/~0", "1", true},
		{"/list/1", "20", true},
		{"/list/01", "20", true},
		{"/list/2", "", false},
		{"/list/-", "", false},
		{"/list/", "", false},
		{"/list/99999999999999999999999", "", false},
		{"/nope", "", false},
		{"/list/0/deeper", "", false},
		{"/", `"empty"`, true},
		{"no-slash", "", false},
	}
	for _, c := range cases {
		got, ok := Pointer(doc, c.ptr)
		if ok != c.ok {
			t.Errorf("%q: ok = %v, want %v", c.ptr, ok, c.ok)
			continue
		}
		if ok && show(got) != c.want {
			t.Errorf("%q = %s, want %s", c.ptr, show(got), c.want)
		}
	}
	whole, ok := Pointer(doc, "")
	if !ok || !Equal(whole, doc) {
		t.Error(`"" must be the whole document`)
	}
}

func TestResolveSubstitutesFilesRecursively(t *testing.T) {
	body := decode(t, `[{"$file": "fixtures/bad-iban.pain.001.001.09.csv"},
		{"nested": {"$file": "fixtures/direct-debit.pain.008.001.08.csv"}, "n": 1},
		{"$file": "x", "other": 1}]`)
	got, err := Resolve(kitDir, body)
	if err != nil {
		t.Fatal(err)
	}
	list := got.([]any)
	if !strings.HasPrefix(list[0].(string), "id,") {
		t.Errorf("first item not substituted: %.20q", list[0])
	}
	nested := list[1].(map[string]any)["nested"].(string)
	want, _ := os.ReadFile(filepath.Join(kitDir, "fixtures", "direct-debit.pain.008.001.08.csv"))
	if nested != string(want) {
		t.Error("nested $file not substituted with the file text")
	}
	if _, ok := list[2].(map[string]any)["$file"]; !ok {
		t.Error("an object with keys besides $file must be left as is")
	}
}

func TestResolveRejectsBadReferences(t *testing.T) {
	for _, body := range []string{
		`{"$file": "../cases.json"}`,
		`{"$file": "/etc/passwd"}`,
		`{"$file": 3}`,
		`{"$file": "fixtures/missing.csv"}`,
		`{"k": [{"$file": "../x"}]}`,
		`[{"$file": "../x"}]`,
	} {
		if _, err := Resolve(kitDir, decode(t, body)); err == nil {
			t.Errorf("%s: want an error", body)
		}
	}
}

func TestIsTypeMatchesPythonsReading(t *testing.T) {
	doc := decode(t, `{"s":"x","i":5,"f":5.0,"e":1e3,"b":true,"a":[],"o":{},"n":null}`).(map[string]any)
	yes := map[string][]string{
		"string": {"s"}, "number": {"i", "f", "e"}, "integer": {"i"},
		"boolean": {"b"}, "array": {"a"}, "object": {"o"}, "null": {"n"},
	}
	for kind, keys := range yes {
		for key, v := range doc {
			is, known := IsType(v, kind)
			if !known || is != slices.Contains(keys, key) {
				t.Errorf("IsType(%s, %s) = %v, %v", key, kind, is, known)
			}
		}
	}
	if _, known := IsType("x", "float"); known {
		t.Error("an unknown kind must be reported")
	}
}

func TestEqualComparesNumbersByValue(t *testing.T) {
	pairs := []struct {
		a, b string
		want bool
	}{
		{`5`, `5.0`, true},
		{`5`, `6`, false},
		{`{"a":[1,"x",null,true]}`, `{"a":[1.00,"x",null,true]}`, true},
		{`{"a":1}`, `{"b":1}`, false},
		{`{"a":1}`, `{"a":1,"b":2}`, false},
		{`[1]`, `[1,2]`, false},
		{`"5"`, `5`, false},
		{`5`, `"5"`, false},
		{`[1]`, `{"0":1}`, false},
		{`{"a":1}`, `[1]`, false},
	}
	for _, p := range pairs {
		if got := Equal(decode(t, p.a), decode(t, p.b)); got != p.want {
			t.Errorf("Equal(%s, %s) = %v", p.a, p.b, got)
		}
	}
	if Equal(json.Number("x"), json.Number("y")) || !Equal(json.Number("x"), json.Number("x")) {
		t.Error("unparsable numbers must compare as text")
	}
}

// TestEvaluateReportsEachKindOfFailure ports the reference runner's
// test_the_runner_reports_each_kind_of_failure.
func TestEvaluateReportsEachKindOfFailure(t *testing.T) {
	health := decode(t, `{"status":"healthy","version":"0.0.85","message":"up"}`)
	expect := Expect{
		Status:   200,
		Equals:   map[string]any{"/status": "down", "/nope": json.Number("1")},
		Types:    map[string]string{"/status": "integer", "/gone": "string", "/message": "float"},
		Contains: map[string][]string{"/status": {"zzz"}, "/version": {"0.0."}, "/none": {"a"}},
	}
	got := Evaluate(expect, 200, health)
	want := []string{
		"/nope missing", `/status != "down"`,
		"/gone missing", `/message: unknown type "float"`, "/status is not integer",
		"/none missing", `/status lacks "zzz"`,
	}
	if !slices.Equal(got, want) {
		t.Errorf("failures = %q\nwant      %q", got, want)
	}
	if got := Evaluate(Expect{Status: 404}, 200, health); !slices.Equal(got, []string{"status 200 != 404"}) {
		t.Errorf("status failure = %q", got)
	}
	faces := decode(t, `[{"name":"berlin_group"}]`)
	list := Expect{
		Status:   200,
		Equals:   map[string]any{"/9/name": "x", "/a": json.Number("1")},
		Contains: map[string][]string{"/0": {"x"}},
	}
	want = []string{"/9/name missing", "/a missing", "/0 is not a string"}
	if got := Evaluate(list, 200, faces); !slices.Equal(got, want) {
		t.Errorf("list failures = %q", got)
	}
	if got := Evaluate(Expect{Status: 200, Types: map[string]string{"/0/name": "string"}}, 200, faces); got != nil {
		t.Errorf("passing case reported %q", got)
	}
}

func TestShowKeepsXMLReadable(t *testing.T) {
	if got := show("<NbOfTxs>5</NbOfTxs>"); got != `"<NbOfTxs>5</NbOfTxs>"` {
		t.Errorf("show = %s", got)
	}
}

func TestShowFallsBackForUnencodableValues(t *testing.T) {
	if got := show(make(chan int)); !strings.HasPrefix(got, "0x") {
		t.Errorf("show(chan) = %q", got)
	}
}

func TestLoadReadsTheVendoredKit(t *testing.T) {
	k, err := Load(kitDir)
	if err != nil {
		t.Fatal(err)
	}
	if k.KitVersion != 1 || len(k.Cases) != 10 || k.Cases[0].ID != "health" {
		t.Errorf("kit_version %d, %d cases, first %q", k.KitVersion, len(k.Cases), k.Cases[0].ID)
	}
	for _, c := range k.Cases {
		if _, err := Resolve(k.Dir, c.Body); err != nil {
			t.Errorf("%s: %v", c.ID, err)
		}
	}
}

func TestLoadRejectsMissingAndEmptyKits(t *testing.T) {
	dir := t.TempDir()
	if _, err := Load(dir); err == nil {
		t.Error("missing cases.json must fail")
	}
	for _, text := range []string{`{"cases": []}`, `{"cases": 3}`} {
		if err := os.WriteFile(filepath.Join(dir, "cases.json"), []byte(text), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := Load(dir); err == nil {
			t.Errorf("%s must fail", text)
		}
	}
}

func TestRunSendsTheResolvedBody(t *testing.T) {
	k := &Kit{Dir: kitDir}
	c := Case{
		Method: "POST", Path: "/p",
		Body:   map[string]any{"xml": map[string]any{"$file": "fixtures/credit-transfer.pain.001.001.09.xml"}},
		Expect: Expect{Status: 200, Equals: map[string]any{"/ok": true}},
	}
	var sent any
	send := func(_ context.Context, method, path string, body any) (int, any, error) {
		sent = body
		return 200, map[string]any{"ok": true}, nil
	}
	failures, err := k.Run(context.Background(), send, c)
	if err != nil || failures != nil {
		t.Fatalf("failures %q, err %v", failures, err)
	}
	if !strings.Contains(sent.(map[string]any)["xml"].(string), "<NbOfTxs>5</NbOfTxs>") {
		t.Error("the fixture text was not sent")
	}
	boom := errors.New("boom")
	fail := func(context.Context, string, string, any) (int, any, error) { return 0, nil, boom }
	if _, err := k.Run(context.Background(), fail, c); !errors.Is(err, boom) {
		t.Errorf("send error not returned: %v", err)
	}
	c.Body = map[string]any{"$file": "../x"}
	if _, err := k.Run(context.Background(), send, c); err == nil {
		t.Error("a bad $file must fail before sending")
	}
}
