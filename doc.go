// SPDX-License-Identifier: Apache-2.0 OR MIT

// Package pain001 is a typed client for the pain001 REST API
// (`pain001 serve`), which validates payment files and generates
// ISO 20022 pain.001 (credit transfer) and pain.008 (direct debit) XML.
//
// The server does the work; this package sends JSON requests and decodes
// the responses into Go types. Every method takes a context.Context
// first. A non-2xx response comes back as an *APIError carrying the HTTP
// status and the server's detail, for use with errors.As:
//
//	client, err := pain001.NewClient("http://127.0.0.1:8000",
//		pain001.WithAPIKey(os.Getenv("PAIN001_API_KEY")))
//	if err != nil {
//		return err
//	}
//	res, err := client.Generate(ctx, csvText, "payments.csv",
//		pain001.Pain00100109, "")
//	var refusal *pain001.RefusalError
//	var apiErr *pain001.APIError
//	switch {
//	case errors.As(err, &refusal):
//		// refused data: refusal.Response.ValidationErrors,
//		// refusal.Response.SchemeViolations
//	case errors.As(err, &apiErr):
//		// apiErr.Status, apiErr.Detail
//	}
//
// The client never retries and never logs: request bodies hold payment
// data, so nothing here writes them anywhere.
package pain001
