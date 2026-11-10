// SPDX-License-Identifier: Apache-2.0 OR MIT

package pain001

import (
	"context"
	"net/http"
	"net/url"
)

// filePayload is the server's UiFilePayload.
type filePayload struct {
	Filename    string      `json:"filename"`
	Content     string      `json:"content"`
	MessageType MessageType `json:"message_type,omitempty"`
	Scheme      string      `json:"scheme,omitempty"`
}

// checkRequest is the server's CheckRequest.
type checkRequest struct {
	XML        string `json:"xml"`
	Scheme     string `json:"scheme,omitempty"`
	Schematron string `json:"schematron,omitempty"`
}

// faceRequest is the server's FaceRequest.
type faceRequest struct {
	XML      string         `json:"xml"`
	Version  string         `json:"version,omitempty"`
	Bindings map[string]any `json:"bindings,omitempty"`
}

// Health reports whether the server is up and which version it runs.
func (c *Client) Health(ctx context.Context) (*HealthResponse, error) {
	out := new(HealthResponse)
	if err := c.call(ctx, http.MethodGet, "/api/v1/health", nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

// Validate checks a payment file without generating XML. content is the
// file's text; filename's suffix (.csv, .json, .jsonl) picks the parser.
// An empty messageType leaves the server's default; an empty scheme runs
// no scheme rulebook (comma-separate names to run several).
func (c *Client) Validate(ctx context.Context, content, filename string, messageType MessageType, scheme string) (*ValidationResponse, error) {
	body := filePayload{Filename: filename, Content: content, MessageType: messageType, Scheme: scheme}
	out := new(ValidationResponse)
	if err := c.call(ctx, http.MethodPost, "/api/v1/ui/validate", body, out); err != nil {
		return nil, err
	}
	return out, nil
}

// Generate turns a payment file into ISO 20022 XML, returned inline in
// the response. The arguments are those of Validate.
func (c *Client) Generate(ctx context.Context, content, filename string, messageType MessageType, scheme string) (*GenerateResponse, error) {
	body := filePayload{Filename: filename, Content: content, MessageType: messageType, Scheme: scheme}
	out := new(GenerateResponse)
	if err := c.call(ctx, http.MethodPost, "/api/v1/ui/generate", body, out); err != nil {
		return nil, err
	}
	return out, nil
}

// Check runs the server's checks on a pain.001 or pain.008 document.
// scheme adds scheme profiles (comma-separated, for example "sepa-sct")
// and schematron a Schematron preset (sepa, fednow, cbpr); empty strings
// leave them out.
func (c *Client) Check(ctx context.Context, xml, scheme, schematron string) (*CheckResponse, error) {
	body := checkRequest{XML: xml, Scheme: scheme, Schematron: schematron}
	out := new(CheckResponse)
	if err := c.call(ctx, http.MethodPost, "/api/v1/check", body, out); err != nil {
		return nil, err
	}
	return out, nil
}

// Faces lists the faces the server can project pain.001 onto.
func (c *Client) Faces(ctx context.Context) ([]FaceInfo, error) {
	var out []FaceInfo
	if err := c.call(ctx, http.MethodGet, "/api/v1/faces", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ProjectFace projects a pain.001 document onto the named face. An empty
// version lets the server detect the edition from the namespace; bindings
// carries side inputs the message lacks (see FaceInfo.Bindings) and may
// be nil.
func (c *Client) ProjectFace(ctx context.Context, face, xml, version string, bindings map[string]any) (*FaceResponse, error) {
	body := faceRequest{XML: xml, Version: version, Bindings: bindings}
	out := new(FaceResponse)
	path := "/api/v1/faces/" + url.PathEscape(face)
	if err := c.call(ctx, http.MethodPost, path, body, out); err != nil {
		return nil, err
	}
	return out, nil
}
