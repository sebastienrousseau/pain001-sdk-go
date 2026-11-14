<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->

# pain001-sdk-go

A typed Go client for the pain001 REST API: validate payment files and
generate ISO 20022 pain.001 and pain.008 XML from Go.

The pain001 server (`pain001 serve`) does the payment work: it parses
CSV, JSON and JSONL files, validates every row, applies scheme
rulebooks and renders the XML. This module sends it JSON and decodes
the answers into Go types and errors. It uses the Go standard library
only, and CI holds it to the conformance kit that every pain001 client
shares, against a live server.

## Install

```sh
go get github.com/sebastienrousseau/pain001-sdk-go@v0.0.88
```

Go 1.26 or newer. The client needs a reachable pain001 server; install
one with `pip install 'pain001[api]'` and start it with
`pain001 serve --host 127.0.0.1 --port 8000`.

## Quick start

```go
client, err := pain001.NewClient("http://127.0.0.1:8000")
if err != nil {
    log.Fatal(err)
}
csv, err := os.ReadFile("conformance/fixtures/credit-transfer.pain.001.001.09.csv")
if err != nil {
    log.Fatal(err)
}
res, err := client.Generate(context.Background(), string(csv),
    "payments.csv", pain001.Pain00100109, "")
if err != nil {
    log.Fatal(err)
}
fmt.Println(len(res.XML)) // 3466
```

## Where to go next

- [Using the client](guide.md): every method, option and error.
- [Examples](examples.md): one runnable program per method.
- [API reference](https://pkg.go.dev/github.com/sebastienrousseau/pain001-sdk-go)
  on pkg.go.dev.
- [Architecture](ARCHITECTURE.md) and the [decision records](adr/README.md),
  for contributors.
- [Ecosystem comparison](COMPARISON.md) and [benchmarks](BENCHMARKS.md).
