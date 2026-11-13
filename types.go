// SPDX-License-Identifier: Apache-2.0 OR MIT

package pain001

import "encoding/json"

// MessageType names an ISO 20022 message the server generates. The
// constants are the server's MessageType enum at kit version 2; any other
// string converts too, so a newer server's types work without a new
// release of this package.
type MessageType string

// The message types pain001 0.0.87 supports.
const (
	Pain00100103 MessageType = "pain.001.001.03"
	Pain00100104 MessageType = "pain.001.001.04"
	Pain00100105 MessageType = "pain.001.001.05"
	Pain00100106 MessageType = "pain.001.001.06"
	Pain00100107 MessageType = "pain.001.001.07"
	Pain00100108 MessageType = "pain.001.001.08"
	Pain00100109 MessageType = "pain.001.001.09"
	Pain00100110 MessageType = "pain.001.001.10"
	Pain00100111 MessageType = "pain.001.001.11"
	Pain00100112 MessageType = "pain.001.001.12"
	Pain00100113 MessageType = "pain.001.001.13"
	Pain00800102 MessageType = "pain.008.001.02"
	Pain00800108 MessageType = "pain.008.001.08"
)

// MessageTypes returns every MessageType constant, in the server's order.
func MessageTypes() []MessageType {
	return []MessageType{
		Pain00100103, Pain00100104, Pain00100105, Pain00100106,
		Pain00100107, Pain00100108, Pain00100109, Pain00100110,
		Pain00100111, Pain00100112, Pain00100113, Pain00800102,
		Pain00800108,
	}
}

// HealthResponse is the body of GET /api/v1/health.
type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
	Message string `json:"message"`
}

// RowError is one row-level problem in a validated or generated file.
// Value is the offending value as raw JSON, and may be absent.
type RowError struct {
	Field   string          `json:"field"`
	Message string          `json:"message"`
	Value   json.RawMessage `json:"value,omitempty"`
}

// ValidationResponse is the body of POST /api/v1/ui/validate.
type ValidationResponse struct {
	IsValid          bool             `json:"is_valid"`
	TotalRows        int              `json:"total_rows"`
	ValidRows        int              `json:"valid_rows"`
	InvalidRows      int              `json:"invalid_rows"`
	Errors           []RowError       `json:"errors"`
	SchemeViolations []map[string]any `json:"scheme_violations"`
}

// GenerateResponse is the body of POST /api/v1/ui/generate. Generate
// returns it with Success true and the document in XML; a refusal's body
// (Success false, XML empty, the findings in ValidationErrors or
// SchemeViolations) comes as RefusalError.Response.
type GenerateResponse struct {
	Success          bool             `json:"success"`
	Message          string           `json:"message"`
	Filename         string           `json:"filename"`
	XML              string           `json:"xml"`
	ValidationErrors []RowError       `json:"validation_errors"`
	SchemeViolations []map[string]any `json:"scheme_violations"`
}

// CheckFinding is one problem one rung of the check found.
type CheckFinding struct {
	Rung     string `json:"rung"`
	Rule     string `json:"rule"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Location string `json:"location"`
}

// CheckResponse is the body of POST /api/v1/check: the verdict, the
// rungs that ran and every finding.
type CheckResponse struct {
	MessageType string         `json:"message_type"`
	Passed      bool           `json:"passed"`
	Errors      int            `json:"errors"`
	Warnings    int            `json:"warnings"`
	Rungs       []string       `json:"rungs"`
	Findings    []CheckFinding `json:"findings"`
}

// FaceInfo describes one face: a projection of pain.001 onto another
// standard's JSON shape.
type FaceInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Bindings    []string `json:"bindings"`
}

// LossReport says what a projection carried and what it did not, by
// indexed source path.
type LossReport struct {
	Face                string   `json:"face"`
	TotalSourceElements int      `json:"total_source_elements"`
	MappedCount         int      `json:"mapped_count"`
	UnmappedCount       int      `json:"unmapped_count"`
	IsLossless          bool     `json:"is_lossless"`
	LossRatio           float64  `json:"loss_ratio"`
	MappedPaths         []string `json:"mapped_paths"`
	UnmappedPaths       []string `json:"unmapped_paths"`
}

// FaceResponse is the body of POST /api/v1/faces/{face}. Payload is the
// target's JSON, left raw because its shape belongs to the target.
type FaceResponse struct {
	Face       string          `json:"face"`
	Version    string          `json:"version"`
	Payload    json.RawMessage `json:"payload"`
	LossReport LossReport      `json:"loss_report"`
}
