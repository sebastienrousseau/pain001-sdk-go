// SPDX-License-Identifier: Apache-2.0 OR MIT

package pain001_test

import (
	"context"
	"errors"
	"fmt"

	pain001 "github.com/sebastienrousseau/pain001-sdk-go"
)

// Validate a file and tell a refused file apart from a transport error.
func ExampleClient_Validate() {
	client, err := pain001.NewClient("http://127.0.0.1:8000")
	if err != nil {
		panic(err)
	}
	csv := "id,date,nb_of_txs,..." // the payment file's text
	res, err := client.Validate(context.Background(), csv, "payments.csv", pain001.Pain00100109, "")
	var apiErr *pain001.APIError
	switch {
	case errors.As(err, &apiErr):
		fmt.Println("refused:", apiErr.Status, apiErr.Error())
	case err != nil:
		fmt.Println("no answer:", err)
	default:
		fmt.Println("valid:", res.IsValid, res.ValidRows, "of", res.TotalRows)
	}
}
