// SPDX-License-Identifier: Apache-2.0 OR MIT

package pain001

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEveryRequestCarriesTheUserAgent(t *testing.T) {
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("User-Agent")
		_, _ = w.Write([]byte(`{"status":"healthy"}`))
	}))
	t.Cleanup(srv.Close)
	c, err := NewClient(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.Health(context.Background()); err != nil {
		t.Fatal(err)
	}
	if want := "pain001-sdk-go/" + Version; got != want {
		t.Errorf("User-Agent %q, want %q", got, want)
	}
}
