<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->

<p align="center">
  <img src="https://cloudcdn.pro/pain001/v1/logos/pain001.svg" alt="pain001-sdk-go logo" width="128" />
</p>

<h1 align="center">pain001-sdk-go</h1>

<p align="center">
  A typed Go client for the pain001 REST API: validate payment files and generate ISO 20022 pain.001 and pain.008 XML from Go.
</p>

<p align="center">
  <a href="https://github.com/sebastienrousseau/pain001-sdk-go/actions"><img src="https://img.shields.io/github/actions/workflow/status/sebastienrousseau/pain001-sdk-go/ci.yml?style=for-the-badge&logo=github" alt="Build" height="28" /></a>
  <a href="https://pkg.go.dev/github.com/sebastienrousseau/pain001-sdk-go"><img src="https://img.shields.io/badge/pkg.go.dev-not%20yet%20published-lightgrey?style=for-the-badge&color=fc8d62&logo=go" alt="Registry" height="28" /></a>
  <a href="https://pkg.go.dev/github.com/sebastienrousseau/pain001-sdk-go"><img src="https://img.shields.io/badge/docs-pkg.go.dev-blue?style=for-the-badge&labelColor=555555&logo=go" alt="Docs" height="28" /></a>
  <a href="https://scorecard.dev/viewer/?uri=github.com/sebastienrousseau/pain001-sdk-go"><img src="https://img.shields.io/ossf-scorecard/github.com/sebastienrousseau/pain001-sdk-go?style=for-the-badge&label=OpenSSF%20Scorecard&logo=openssf" alt="OpenSSF Scorecard" height="28" /></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-Apache--2.0%20OR%20MIT-blue.svg?style=for-the-badge" alt="License: Apache-2.0 OR MIT" height="28" /></a>
  <a href="#requirements"><img src="https://img.shields.io/badge/Go-1.26%2B-93450a.svg?style=for-the-badge&logo=go" alt="Go 1.26 or newer" height="28" /></a>
</p>

<p align="center">
  No demo GIF: this is a library with no terminal interface of its own,
  and its one example needs a running server. The
  <a href="#quick-start">Quick Start</a> shows the call and its output.
</p>

---

## Contents

**Getting started**

- [Install](#install): `go get`, standard library only
- [Requirements](#requirements): toolchain floor, platforms
- [Quick Start](#quick-start): one call from a CSV file to pain.001 XML

**The pain001-sdk-go ecosystem**

- [The pain001-sdk-go ecosystem](#the-pain001-sdk-go-ecosystem): the pain001 core server, its MCP and LSP servers, its loaders and the conformance kit

**Library reference**

- [Capabilities at a glance](#capabilities-at-a-glance): the current surface by theme
- [Ecosystem comparison](#ecosystem-comparison): how this differs from building XML in Go
- [Benchmarks](#benchmarks): why there are none
- [Features](#features): module-level capability list
- [Configuration](#configuration): core options
- [Examples](#examples): runnable example index

**Operational**

- [When not to use pain001-sdk-go](#when-not-to-use-pain001-sdk-go): limitations
- [Development](#development): make targets, CI
- [Security](#security): guarantees and compliance
- [Documentation](#documentation): all reference docs
- [Stability guarantees](#stability-guarantees): SemVer axis, output stability, minimum toolchain discipline
- [License](#license)

---

## Install

### As a Go library

```sh
go get github.com/sebastienrousseau/pain001-sdk-go
```

> **Not yet published.** The module has no tagged release yet; the first
> will be 0.0.87, in lockstep with the pain001 suite. Until then the
> command above does not resolve.

The module uses the Go standard library only. It needs a running
pain001 server (`pain001 serve`, from `pip install 'pain001[api]'`) to
talk to.

---

## Requirements

- Go 1.26 or newer. CI tests Go 1.26 and 1.27, the two releases the Go
  team supports; the floor moves when 1.26 leaves support.
- A pain001 server, version 0.0.85 or newer, reachable over HTTP or
  HTTPS. The vendored conformance kit (kit 2) is core 0.0.87's, and its
  refusal cases need a 0.0.87 server. Before 0.0.87 the server answers
  a refused `Generate` with `200` and `success: false`, which comes back
  as an error wrapping `ErrNoDocument` rather than a `*RefusalError`,
  so a nil error from `Generate` always comes with the XML.
- Pure Go with no cgo, so any platform Go builds for; CI exercises
  Linux, and the tests have also run on macOS (arm64).

---

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    pain001 "github.com/sebastienrousseau/pain001-sdk-go"
)

func main() {
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
    fmt.Println(res.Success, len(res.XML)) // true 3466
}
```

`Generate` posts the file to `/api/v1/ui/generate`; the server parses
it, validates every row and returns the XML inline. The fixture is the
conformance kit's five-transaction credit transfer, so the document
carries `<NbOfTxs>5</NbOfTxs>`.

When the server refuses the data (rows that fail validation or a scheme
rule) it answers `422`, and `Generate` returns a `*RefusalError` whose
`Response` holds the findings:

```go
var refusal *pain001.RefusalError
if errors.As(err, &refusal) {
    for _, row := range refusal.Response.ValidationErrors {
        fmt.Println(row.Field, row.Message) // $.currency 'EURO' does not match ...
    }
    for _, v := range refusal.Response.SchemeViolations {
        fmt.Println(v["rule"], v["message"])
    }
}
```

---

## The pain001-sdk-go ecosystem

This module is one of the pain001 suite, which releases at one version,
core first. The design is one engine (the Python server) and thin typed
clients per language, all held to one conformance kit that core
publishes and every client vendors. New to the format?
[What is pain.001?](https://pain001.com/pain-001/) explains it.

| Component | Purpose | Use case |
| :--- | :--- | :--- |
| [`pain001`][core] | Generates and validates ISO 20022 payment files; serves the REST API this client calls | The engine behind every client |
| **`pain001-sdk-go`** | Typed Go client for the REST API | Go services that produce or check payment files |
| [`pain001-mcp`][mcp] | MCP server exposing pain001 validation, generation and inspection | Agent workflows over payment files |
| [`pain001-lsp`][lsp] | Language server for pain001 input files | Editor support |
| [`pain001-loader-mt101`][mt101] | SWIFT MT101 text to pain.001 records | A treasury or ERP system still emits MT101 |
| [`pain001-loader-xlsx`][xlsx] | Excel loader plugin for `.xlsx` / `.xlsm` | Payment batches kept in spreadsheets |

---

## Capabilities at a glance

| Area | Capability | Status |
| :--- | :--- | :--- |
| Health | `Health`: server status and version | Supported |
| Files | `Validate` and `Generate` for CSV, JSON and JSONL payment files, with optional scheme rulebooks | Supported |
| Documents | `Check`: XSD, MDR, scheme and Schematron rungs on existing XML | Supported |
| Faces | `Faces` and `ProjectFace`: pain.001 projected onto Berlin Group, UK Open Banking or ISO 2018 JSON, with a loss report | Supported |
| Escape hatch | `Request`: any method and path, status and decoded JSON | Supported |
| Errors | `*APIError` with status and server detail, and `*RefusalError` with a refused `Generate`'s findings, for `errors.As` | Supported |
| Async jobs | `/api/v1/generate/async`, `/status`, `/download`, `/jobs` | Not wrapped; reachable through `Request` |
| Retries | Automatic retry on failure | Not provided, by design |

---

## Ecosystem comparison

This package has not been benchmarked or feature-compared with other Go
ISO 20022 libraries, so no comparison is claimed. It differs in kind:
it builds no XML itself and sends the work to the pain001 server, so Go
code gets the same templates, schema checks and rulebooks as the Python
CLI and every other pain001 client.

| Project | Builds XML in-process | Checks done by | Licence |
| :--- | :---: | :---: | :---: |
| **pain001-sdk-go** | No, the server does | pain001 server (XSD, MDR, scheme rules) | Apache-2.0 OR MIT |

---

## Benchmarks

None. A call's cost is one HTTP round trip plus the server's work, and
both depend on the deployment, not on this client; a number measured
here would describe the machine, not the package. The server's own
benchmarks live in [`pain001`][core].

---

## Features

- **One method per endpoint.** `Health`, `Validate`, `Generate`,
  `Check`, `Faces` and `ProjectFace` map to the REST API's documented
  endpoints and decode into Go structs (`HealthResponse`,
  `ValidationResponse`, `GenerateResponse`, `CheckResponse`,
  `FaceInfo`, `FaceResponse`).
- **Context first.** Every method takes a `context.Context`, so callers
  cancel or bound calls the usual way.
- **Typed errors.** A non-2xx response returns `*APIError{Status,
  Detail}`. `DetailText()` gives a string detail (a 400 for a bad IBAN);
  `Issues()` gives the 422 validation array. When `Generate` is refused
  (`422` with `success: false`) the error is a `*RefusalError`: it wraps
  the `*APIError` (so `Status` stays 422) and its `Response` carries the
  decoded `ValidationErrors` and `SchemeViolations`. The error message
  never includes the rejected input.
- **Message types as constants, not a closed set.** `Pain00100103` to
  `Pain00100113`, `Pain00800102` and `Pain00800108` match the server's
  enum; any other string converts to `MessageType`, so a newer server's
  types work without a new client.
- **Raw access.** `Request(ctx, method, path, body)` returns the status
  and the body decoded as JSON with `json.Number`, whatever the status.
  The conformance runner uses it.
- **Bounded reads.** A response body over 64 MiB is refused with
  `ErrResponseTooLarge`.

---

## Configuration

`NewClient(baseURL, opts...)` takes an absolute `http` or `https` base
URL and these options:

| Option | Effect | Default |
| :--- | :--- | :--- |
| `WithAPIKey(key)` | Sends `Authorization: Bearer <key>`, for a server started with `PAIN001_API_KEY` | No header |
| `WithHTTPClient(hc)` | Sends requests through `hc` (proxy, TLS, transport) | A fresh `http.Client` |
| `WithTimeout(d)` | Bounds each call, including reading the body; zero turns it off | `DefaultTimeout`, 30 seconds |

---

## Examples

- [`examples/basic`](examples/basic/main.go): generate a pain.001.001.09
  file from the kit's credit-transfer fixture and write it to disk, or
  print each finding when the server refuses the file.
  Run it from the repository root with a server up:
  `PAIN001_BASE_URL=http://127.0.0.1:8000 go run ./examples/basic -out payments.xml`.
- [`example_test.go`](example_test.go): validating a file and telling
  a refused file apart from a transport error.

CI runs `examples/basic` against a live server on every pull request.

---

## When not to use pain001-sdk-go

- **You cannot run the pain001 server.** This client does nothing
  without it; there is no offline mode.
- **You need XML built inside your Go process**, for example in a
  sandbox with no network. Use a Go XML library and the ISO 20022
  schemas directly.
- **You need the asynchronous job endpoints with typed results.** They
  are reachable only through `Request` today.
- **You want automatic retries.** The client never retries, because it
  cannot know which calls are safe to repeat against your deployment;
  wrap calls yourself where they are.

---

## Development

```bash
git clone https://github.com/sebastienrousseau/pain001-sdk-go
cd pain001-sdk-go
make check
# with a server, to add the conformance kit, typed tests and example:
PAIN001_BASE_URL=http://127.0.0.1:8000 make check
```

`make check` runs `gofmt -l`, `go vet`, staticcheck, the gocyclo
(cyclomatic 10) and gocognit (cognitive 15) ceilings, `go build` and
`go test -race` with coverage; a test holds functions to 60 lines and
files to 500. The analysers run through `go run` at pinned versions, so
nothing is installed globally. `go test` also runs the fuzz targets'
seeds; `make fuzz` fuzzes each target for a minute (`FUZZTIME=30s` to
change it). CI adds a Go 1.26 and 1.27 matrix, 30 seconds of fuzzing
per target, a conformance job against `pain001 serve` installed from
hash-pinned requirements, markdownlint and the DCO check. See
[`CONTRIBUTING.md`](CONTRIBUTING.md).

---

## Security

The client sends JSON to the server you configure and decodes the
reply. It writes nothing to disk, logs nothing (request bodies hold
payment data) and keeps rejected input out of error messages. The API
key travels only in the `Authorization` header. Response bodies are
capped at 64 MiB. Use HTTPS for any server that is not on loopback. The
module has no dependencies outside the Go standard library.

Report vulnerabilities according to [`SECURITY.md`](SECURITY.md).

---

## Documentation

[API reference](https://pkg.go.dev/github.com/sebastienrousseau/pain001-sdk-go) ·
[Developer guide](CONTRIBUTING.md) ·
[Conformance kit](conformance/README.md) ·
[Family map](#the-pain001-sdk-go-ecosystem) ·
[Changelog](CHANGELOG.md) · [Security](SECURITY.md) ·
[Agent rules](AGENTS.md)

---

## Stability guarantees

The project is pre-1.0 and versions advance in `0.0.1` steps, in
lockstep with the pain001 suite: client 0.0.N vendors the conformance
kit from core 0.0.N. The breaking axis is the exported Go API (function
signatures, struct fields and option names) and the requests it sends;
the changelog calls out any change to either. The Go floor moves only
when the Go team stops supporting a release.

---

## License

This project is dual-licensed under either:

- Apache License, Version 2.0 ([`LICENSE-APACHE`](LICENSE-APACHE) or <https://www.apache.org/licenses/LICENSE-2.0>)
- MIT License ([`LICENSE-MIT`](LICENSE-MIT) or <https://opensource.org/licenses/MIT>)

at your option.

[core]: https://github.com/sebastienrousseau/pain001
[mcp]: https://github.com/sebastienrousseau/pain001-mcp
[lsp]: https://github.com/sebastienrousseau/pain001-lsp
[mt101]: https://github.com/sebastienrousseau/pain001-loader-mt101
[xlsx]: https://github.com/sebastienrousseau/pain001-loader-xlsx
