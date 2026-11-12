// SPDX-License-Identifier: Apache-2.0 OR MIT

package pain001_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	pain001 "github.com/sebastienrousseau/pain001-sdk-go"
)

const (
	creditCSV = "credit-transfer.pain.001.001.09.csv"
	creditXML = "credit-transfer.pain.001.001.09.xml"
	badIBAN   = "bad-iban.pain.001.001.09.csv"
	debitCSV  = "direct-debit.pain.008.001.08.csv"
	badCcy    = "bad-currency.pain.001.001.09.csv"
	chapsCSV  = "chaps-no-purpose.pain.001.001.09.csv"
)

func TestLiveHealth(t *testing.T) {
	res, err := liveClient(t).Health(context.Background())
	if err != nil || res.Status != "healthy" || res.Version == "" {
		t.Fatalf("health %+v, err %v", res, err)
	}
}

func TestLiveValidate(t *testing.T) {
	res, err := liveClient(t).Validate(context.Background(), fixture(t, creditCSV), "payments.csv", pain001.Pain00100109, "")
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsValid || res.TotalRows != 5 || res.ValidRows != 5 || res.InvalidRows != 0 || len(res.Errors) != 0 {
		t.Errorf("validate %+v", res)
	}
}

func TestLiveValidateBadIBANIsA400(t *testing.T) {
	_, err := liveClient(t).Validate(context.Background(), fixture(t, badIBAN), "payments.csv", pain001.Pain00100109, "")
	var apiErr *pain001.APIError
	if !errors.As(err, &apiErr) || apiErr.Status != 400 {
		t.Fatalf("err %v, want an APIError with status 400", err)
	}
	if text, ok := apiErr.DetailText(); !ok || !strings.Contains(text, "not a valid IBAN") {
		t.Errorf("detail %s", apiErr.Detail)
	}
}

func TestLiveUnknownMessageTypeIsA422(t *testing.T) {
	_, err := liveClient(t).Validate(context.Background(), fixture(t, creditCSV), "payments.csv", "pain.999.001.01", "")
	var apiErr *pain001.APIError
	if !errors.As(err, &apiErr) || apiErr.Status != 422 || len(apiErr.Issues()) == 0 {
		t.Fatalf("err %v, want a 422 with issues", err)
	}
	if strings.Contains(err.Error(), "creditor") {
		t.Error("the error message must not echo the payment file")
	}
}

func TestLiveGenerate(t *testing.T) {
	ctx := context.Background()
	c := liveClient(t)
	res, err := c.Generate(ctx, fixture(t, creditCSV), "payments.csv", pain001.Pain00100109, "")
	if err != nil || !res.Success || !strings.Contains(res.XML, "<NbOfTxs>5</NbOfTxs>") {
		t.Fatalf("generate success %v, err %v", res != nil && res.Success, err)
	}
	dd, err := c.Generate(ctx, fixture(t, debitCSV), "collections.csv", pain001.Pain00800108, "")
	if err != nil || !strings.Contains(dd.XML, "urn:iso:std:iso:20022:tech:xsd:pain.008.001.08") {
		t.Fatalf("direct debit generate failed: %v", err)
	}
}

func TestLiveGenerateRefusedRowsIsATypedRefusal(t *testing.T) {
	_, err := liveClient(t).Generate(context.Background(), fixture(t, badCcy), "payments.csv", pain001.Pain00100109, "")
	var refusal *pain001.RefusalError
	if !errors.As(err, &refusal) || refusal.Status != 422 || refusal.Response.Success {
		t.Fatalf("err %v, want a 422 *RefusalError", err)
	}
	rows := refusal.Response.ValidationErrors
	if len(rows) == 0 || rows[0].Field != "$.currency" || string(rows[0].Value) != `"EURO"` {
		t.Errorf("validation errors %+v", rows)
	}
	if len(refusal.Response.SchemeViolations) != 0 || refusal.Response.XML != "" {
		t.Errorf("refusal %+v", refusal.Response)
	}
	if strings.Contains(err.Error(), "EURO") {
		t.Error("the error message must not echo the rejected value")
	}
}

func TestLiveGenerateRefusedSchemeIsATypedRefusal(t *testing.T) {
	_, err := liveClient(t).Generate(context.Background(), fixture(t, chapsCSV), "payments.csv", pain001.Pain00100109, "uk-chaps")
	var refusal *pain001.RefusalError
	if !errors.As(err, &refusal) || refusal.Status != 422 {
		t.Fatalf("err %v, want a 422 *RefusalError", err)
	}
	sv := refusal.Response.SchemeViolations
	if len(refusal.Response.ValidationErrors) != 0 || len(sv) == 0 || sv[0]["rule"] != "UK-CHAPS-PURP" || sv[0]["field"] != "purpose_code" {
		t.Errorf("refusal %+v", refusal.Response)
	}
}

func TestLiveCheck(t *testing.T) {
	res, err := liveClient(t).Check(context.Background(), fixture(t, creditXML), "", "")
	if err != nil {
		t.Fatal(err)
	}
	if !res.Passed || res.Errors != 0 || res.MessageType != "pain.001.001.09" || len(res.Rungs) == 0 {
		t.Errorf("check %+v", res)
	}
}

func TestLiveFaces(t *testing.T) {
	faces, err := liveClient(t).Faces(context.Background())
	if err != nil || len(faces) < 3 || faces[0].Name != "berlin_group" || faces[0].Description == "" {
		t.Fatalf("faces %+v, err %v", faces, err)
	}
}

func TestLiveProjectFace(t *testing.T) {
	ctx := context.Background()
	c := liveClient(t)
	res, err := c.ProjectFace(ctx, "berlin_group", fixture(t, creditXML), "", nil)
	if err != nil {
		t.Fatal(err)
	}
	var payload struct {
		DebtorAccount struct {
			IBAN string `json:"iban"`
		} `json:"debtorAccount"`
	}
	if err := json.Unmarshal(res.Payload, &payload); err != nil {
		t.Fatal(err)
	}
	if res.Face != "berlin_group" || res.Version != "pain.001.001.09" || payload.DebtorAccount.IBAN != "DE07512108001245126162" {
		t.Errorf("face %s version %s iban %s", res.Face, res.Version, payload.DebtorAccount.IBAN)
	}
	if res.LossReport.Face != "berlin_group" || res.LossReport.TotalSourceElements == 0 {
		t.Errorf("loss report %+v", res.LossReport)
	}
	var apiErr *pain001.APIError
	if _, err := c.ProjectFace(ctx, "no_such_face", fixture(t, creditXML), "", nil); !errors.As(err, &apiErr) || apiErr.Status != 404 {
		t.Errorf("unknown face: %v", err)
	}
}

func TestLiveRequest(t *testing.T) {
	status, doc, err := liveClient(t).Request(context.Background(), "GET", "/api/v1/health", nil)
	if err != nil || status != 200 || doc.(map[string]any)["status"] != "healthy" {
		t.Fatalf("status %d doc %v err %v", status, doc, err)
	}
}
