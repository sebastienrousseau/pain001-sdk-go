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
  <a href="https://pkg.go.dev/github.com/sebastienrousseau/pain001-sdk-go"><img src="https://img.shields.io/github/v/tag/sebastienrousseau/pain001-sdk-go?sort=semver&label=go%20module&style=for-the-badge&color=fc8d62&logo=go" alt="Registry" height="28" /></a>
  <a href="https://pkg.go.dev/github.com/sebastienrousseau/pain001-sdk-go"><img src="https://img.shields.io/badge/docs-pkg.go.dev-blue?style=for-the-badge&labelColor=555555&logo=go" alt="Docs" height="28" /></a>
  <a href="https://scorecard.dev/viewer/?uri=github.com/sebastienrousseau/pain001-sdk-go"><img src="https://img.shields.io/ossf-scorecard/github.com/sebastienrousseau/pain001-sdk-go?style=for-the-badge&label=OpenSSF%20Scorecard&logo=openssf" alt="OpenSSF Scorecard" height="28" /></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-Apache--2.0%20OR%20MIT-blue.svg?style=for-the-badge" alt="License: Apache-2.0 OR MIT" height="28" /></a>
  <a href="https://github.com/sebastienrousseau/pain001-sdk-go/blob/main/docs/POLICIES.md"><img src="https://img.shields.io/badge/Go-1.26%2B-93450a.svg?style=for-the-badge&logo=go" alt="Go 1.26 or newer" height="28" /></a>
</p>

<p align="center">
  <img src="https://raw.githubusercontent.com/sebastienrousseau/pain001-sdk-go/main/.github/demo.gif" alt="A terminal running the client's live tests for Health, Validate, Generate and Check against a local pain001 server, then examples/basic writing a five-transaction pain.001.001.09 file from the conformance fixture and printing its group header" width="100%" />
</p>

---

## Contents

**Getting started**

- [Install](#install): `go get`, standard library only
- [Requirements](#requirements): toolchain floor, platforms
- [Quick Start](#quick-start): one call from a CSV file to pain.001 XML

**The pain001-sdk-go ecosystem**

- [The pain001-sdk-go ecosystem](#the-pain001-sdk-go-ecosystem): the pain001 server, the five SDKs, the MCP and LSP servers, the loaders and the conformance kit

**Library reference**

- [Capabilities at a glance](#capabilities-at-a-glance): the current surface by theme
- [Ecosystem comparison](#ecosystem-comparison): short matrix; full table at [`docs/COMPARISON.md`](docs/COMPARISON.md)
- [Benchmarks](#benchmarks): headline numbers; full table at [`docs/BENCHMARKS.md`](docs/BENCHMARKS.md)
- [Features](#features): module-level capability list
- [Configuration](#configuration): core options
- [Examples](#examples): runnable example index

**Operational**

- [When not to use pain001-sdk-go](#when-not-to-use-pain001-sdk-go): limitations
- [Development](#development): make targets, fuzzing, CI
- [Security](#security): guarantees and compliance
- [Documentation](#documentation): all reference docs
- [Stability guarantees](#stability-guarantees): SemVer axis, output stability, minimum toolchain discipline
- [License](#license)

---

## Install

### As a Go library

```sh
go get github.com/sebastienrousseau/pain001-sdk-go@v0.0.88
```

> **Not yet published.** The module has no tagged release yet; the first
> will be 0.0.88, in lockstep with the pain001 suite. Until then the
> command above does not resolve; use the repository directly with a
> `replace` directive or `go get github.com/sebastienrousseau/pain001-sdk-go@main`.

The module uses the Go standard library only: `go get` adds one line to
your `go.mod` and nothing to your `go.sum` beyond this module. It needs
a running pain001 server to talk to:

```sh
pip install 'pain001[api]'
pain001 serve --host 127.0.0.1 --port 8000
```

---

## Requirements

- **Go 1.26 or newer.** CI tests Go 1.26 and 1.27, the two releases the
  Go team supports; the floor moves when 1.26 leaves support
  ([policy](docs/POLICIES.md#minimum-go-toolchain)).
- **A pain001 server**, 0.0.85 or newer, reachable over HTTP or HTTPS.
  CI tests against pain001 0.0.87, pinned in `ci/server-requirements.in`.
  The vendored conformance kit (kit 2) needs 0.0.87 for its refusal
  cases: before it, a refused `Generate` comes back as an error wrapping
  `ErrNoDocument` rather than a `*RefusalError`.
- **Any platform Go builds for.** Pure Go, no cgo; CI runs on Linux, and
  the tests and regression run also pass on macOS (arm64).

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
        fmt.Println(row.Field, row.Message) // $.currency 'EURO' does not match '^[A-Z]{3}$'
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
| [`pain001-sdk-java`][java] | Typed Java client for the REST API | JVM services |
| [`pain001-sdk-rust`][rust] | Typed Rust client for the REST API | Rust services |
| [`pain001-sdk-dotnet`][dotnet] | Typed .NET client for the REST API | .NET services |
| [`pain001-sdk-cpp`][cpp] | Typed C++ client for the REST API | C and C++ systems |
| [`pain001-mcp`][mcp] | MCP server exposing pain001 validation, generation and inspection | Agent workflows over payment files |
| [`pain001-lsp`][lsp] | Language server for pain001 input files | Editor support |
| [`pain001-loader-mt101`][mt101] | SWIFT MT101 text to pain.001 records | A treasury or ERP system still emits MT101 |
| [`pain001-loader-xlsx`][xlsx] | Excel loader plugin for `.xlsx` / `.xlsm` | Payment batches kept in spreadsheets |
| [`pain001-mockbank`][mockbank] | Synthetic SFTP bank-response test service | pain.002 replies in integration tests |

---

## Capabilities at a glance

| Area | Capability | Status |
| :--- | :--- | :--- |
| Health | `Health`: server status and version | Supported |
| Files | `Validate` and `Generate` for CSV, JSON and JSONL payment files, with optional scheme rulebooks | Supported |
| Documents | `Check`: XSD, MDR, scheme and Schematron rungs on existing XML | Supported |
| Faces | `Faces` and `ProjectFace`: pain.001 projected onto Berlin Group, UK Open Banking or ISO 2018 JSON, with a loss report | Supported |
| Escape hatch | `Request`: any method and path, status and decoded JSON | Supported |
| Errors | `*APIError` with status and server detail, `*RefusalError` with a refused `Generate`'s findings, `ErrNoDocument`, for `errors.As` and `errors.Is` | Supported |
| Retries | `WithRetry`: GET calls only, on a failed connection, `429` or `503`, with backoff and `Retry-After` | Supported, opt-in |
| Async jobs | `/api/v1/generate/async`, `/status`, `/download`, `/jobs` | Not wrapped; reachable through `Request` |

---

## Ecosystem comparison

The suite's five SDKs share one method surface and run core's conformance kit against a live `pain001 serve` in CI. Core's CI also generates a Python client from the OpenAPI document, which it imports but does not run against the kit.

| Project | Language (minimum) | Conformance kit in CI | Types | Runtime dependencies | Call model |
| :--- | :---: | :---: | :---: | :--- | :--- |
| **pain001-sdk-go** | **Go 1.26** | **Yes** | **Hand-written** | **None (standard library)** | **Blocking, `context.Context`** |
| pain001-sdk-java | Java 17 | Yes | Hand-written | Jackson databind | Blocking |
| pain001-sdk-rust | Rust 1.85 | Yes | Hand-written (serde) | ureq, serde, serde_json, thiserror | Blocking |
| pain001-sdk-dotnet | .NET 10 | Yes | Hand-written | None | async/await, `CancellationToken` |
| pain001-sdk-cpp | C++17 | Yes | Hand-written | libcurl, nlohmann/json | Blocking |
| openapi-generator client (core CI artifact) | Python | No, CI imports it only | Generated | Generator's | Blocking |
| Raw HTTP | Any | No | None | Your HTTP client | Yours |

See [`docs/COMPARISON.md`](docs/COMPARISON.md) for the evidence and complete matrix.

---

## Benchmarks

Every public method is benchmarked twice with `testing.B`: the
client's own cost against recorded responses (no network), and one
loopback round trip to `pain001 serve` 0.0.87. The client's work is
microseconds; the round trip, which is the server's work, is
milliseconds. Medians of 10 one-second samples, measured on
2026-11-14 on a heavily shared machine (see the full page for the load
and spreads):

| Scenario | Result | Environment |
| :--- | ---: | :--- |
| `Health`, client-side | 5.7 µs, 26 allocs | Apple A18 Pro, macOS 27.2, Go 1.27.4 |
| `Generate` (5-row CSV in, 3.5 KB XML out), client-side | 104.8 µs, 43 allocs | same |
| `ProjectFace` (Berlin Group), client-side | 49.3 µs, 96 allocs | same |
| `Health`, loopback round trip | 0.78 ms | same, pain001 0.0.87 on 127.0.0.1 |
| `Generate`, loopback round trip | 16.2 ms | same |
| `Check`, loopback round trip | 10.6 ms | same |

See [`docs/BENCHMARKS.md`](docs/BENCHMARKS.md) for methodology and full results.

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
- **Opt-in retries for idempotent calls.** `WithRetry` retries GET
  calls on a failed connection, `429` or `503`, with exponential
  backoff and jitter, honouring `Retry-After`. POST calls are never
  retried.
- **Message types as constants, not a closed set.** `Pain00100103` to
  `Pain00100113`, `Pain00800102` and `Pain00800108` match the server's
  enum; any other string converts to `MessageType`, so a newer server's
  types work without a new client.
- **Raw access.** `Request(ctx, method, path, body)` returns the status
  and the body decoded as JSON with `json.Number`, whatever the status.
  The conformance runner uses it.
- **Bounded reads.** A response body over 64 MiB is refused with
  `ErrResponseTooLarge`.
- **Identifies itself.** Every request carries
  `User-Agent: pain001-sdk-go/<Version>`.

---

## Configuration

`NewClient(baseURL, opts...)` takes an absolute `http` or `https` base
URL and these options:

| Option | Effect | Default |
| :--- | :--- | :--- |
| `WithAPIKey(key)` | Sends `Authorization: Bearer <key>`, for a server started with `PAIN001_API_KEY` | No header |
| `WithHTTPClient(hc)` | Sends requests through `hc` (proxy, TLS, transport) | A fresh `http.Client` |
| `WithTimeout(d)` | Bounds each call, including reading the body, and all its retries; zero turns it off | `DefaultTimeout`, 30 seconds |
| `WithRetry(RetryPolicy{MaxAttempts, BaseDelay, MaxDelay})` | Retries GET calls on a failed connection, `429` or `503`; POST is never retried | One attempt; with the option, `BaseDelay` 200 ms and `MaxDelay` 5 s when zero |

---

## Examples

Every example is a runnable program under [`examples/`](examples/). Run
them from the repository root against a server, one at a time with
`PAIN001_BASE_URL=http://127.0.0.1:8000 go run ./examples/<name>` or all
together with `make examples`. CI builds each and runs it against
`pain001 serve` on every pull request.

| Example | Shows |
| :--- | :--- |
| [`health`](examples/health/main.go) | `Health`: server status and version |
| [`validate`](examples/validate/main.go) | `Validate`, without and with the `sepa-sct` rulebook |
| [`basic`](examples/basic/main.go) | `Generate`: the kit's credit transfer to pain.001.001.09 XML on disk, or each finding of a refusal |
| [`check`](examples/check/main.go) | `Check` with a scheme profile and a Schematron preset |
| [`faces`](examples/faces/main.go) | `Faces`: each face and its bindings |
| [`projectface`](examples/projectface/main.go) | `ProjectFace` onto Berlin Group, with the loss report |
| [`request`](examples/request/main.go) | `Request`, the raw call, for a 200 and a 422 |
| [`errors`](examples/errors/main.go) | `*APIError` (string detail and issues), `*RefusalError`, `ErrNoDocument` |
| [`timeout`](examples/timeout/main.go) | `WithTimeout` and a context deadline |
| [`apikey`](examples/apikey/main.go) | `WithAPIKey` against a server started with `PAIN001_API_KEY` |
| [`retry`](examples/retry/main.go) | `WithRetry` and `WithHTTPClient`, against a closed port |
| [`example_test.go`](example_test.go) | The pkg.go.dev example: validating a file and telling a refusal from a transport error |

---

## When not to use pain001-sdk-go

- **You cannot run the pain001 server.** This client does nothing
  without it; there is no offline mode.
- **You need XML built inside your Go process**, for example in a
  sandbox with no network. Use a Go XML library and the ISO 20022
  schemas directly.
- **You need the asynchronous job endpoints with typed results.** They
  are reachable only through `Request` today.
- **You need POST calls retried.** The client retries only GET; whether
  a lost `Generate` may be resent is your deployment's decision, made
  outside the client.

---

## Development

```bash
git clone https://github.com/sebastienrousseau/pain001-sdk-go
cd pain001-sdk-go
make check          # every static gate and the tests at 100% coverage
make regression     # make check against a pinned pain001 serve it starts, plus examples and benchmarks
make bench          # benchmarks (BenchmarkLive needs PAIN001_BASE_URL)
make fuzz           # each fuzz target for FUZZTIME (default 60s)
```

`make check` runs the version check, `gofmt -l`, `go vet`,
staticcheck, the gocyclo (cyclomatic 10) and gocognit (cognitive 15)
ceilings, govulncheck, `go build` and `go test -race`, which fails
below 100.0% statement coverage and holds every function to 60 lines,
every file to 500 and a Halstead difficulty of 30. With
`PAIN001_BASE_URL` set it adds the conformance kit, the live typed
tests and every example. `make regression` does that against a server
it installs and starts itself, exactly as CI does. CI adds a Go 1.26
and 1.27 matrix, 30 seconds of fuzzing per target, the gorelease API
check, markdownlint, codespell, REUSE, the strict manual build,
CodeQL, OpenSSF Scorecard and the DCO check. `make demo` re-renders the
GIF above with [vhs](https://github.com/charmbracelet/vhs). See
[`DEVELOPMENT.md`](DEVELOPMENT.md) for every gate and the release
process.

---

## Security

The client sends JSON to the server you configure and decodes the
reply. It writes nothing to disk, logs nothing (request bodies hold
payment data) and keeps rejected input and refusal findings out of
error messages; two fuzz targets hold that. The API key travels only
in the `Authorization` header. Response bodies are capped at 64 MiB and
every call has a deadline. Use HTTPS for any server that is not on
loopback. The module has no dependencies outside the Go standard
library; govulncheck, CodeQL and OpenSSF Scorecard run in CI, actions
are pinned by commit SHA, and every file is REUSE-compliant. Five fuzz
targets run for 30 seconds each on every push, and their crash
regressions replay in every `go test`.

Report vulnerabilities according to [`SECURITY.md`](SECURITY.md).

---

## Documentation

[User manual](https://sebastienrousseau.github.io/pain001-sdk-go/) ·
[API reference](https://pkg.go.dev/github.com/sebastienrousseau/pain001-sdk-go) ·
[Developer docs](DEVELOPMENT.md) ·
[Family map](#the-pain001-sdk-go-ecosystem)

Also: [Architecture](docs/ARCHITECTURE.md) ·
[Decisions](docs/adr/README.md) · [Policies](docs/POLICIES.md) ·
[Roadmap](docs/ROADMAP.md) · [Packaging](docs/packaging.md) ·
[Conformance kit](conformance/README.md) · [Changelog](CHANGELOG.md) ·
[Support](SUPPORT.md) · [Governance](GOVERNANCE.md) ·
[Security](SECURITY.md) · [Agent rules](AGENTS.md)

---

## Stability guarantees

The project is pre-1.0 and versions advance in `0.0.1` steps, in
lockstep with the pain001 suite. The breaking axis is the exported Go
API (function signatures, struct fields, option and constant names)
and the requests it sends; a release may break either while pre-1.0,
and the CHANGELOG calls out every break with a `Breaking:` line, which
the gorelease check in CI requires. Output stability: the client
returns what the server sends, so a change in generated XML or
findings is the server's and is versioned there. The Go floor is the
older of the two supported Go releases and moves only when it leaves
support ([policy](docs/POLICIES.md#minimum-go-toolchain)).

---

## License

This project is dual-licensed under either:

- Apache License, Version 2.0 ([`LICENSE-APACHE`](LICENSE-APACHE) or <https://www.apache.org/licenses/LICENSE-2.0>)
- MIT License ([`LICENSE-MIT`](LICENSE-MIT) or <https://opensource.org/licenses/MIT>)

at your option.

[core]: https://github.com/sebastienrousseau/pain001
[java]: https://github.com/sebastienrousseau/pain001-sdk-java
[rust]: https://github.com/sebastienrousseau/pain001-sdk-rust
[dotnet]: https://github.com/sebastienrousseau/pain001-sdk-dotnet
[cpp]: https://github.com/sebastienrousseau/pain001-sdk-cpp
[mcp]: https://github.com/sebastienrousseau/pain001-mcp
[lsp]: https://github.com/sebastienrousseau/pain001-lsp
[mt101]: https://github.com/sebastienrousseau/pain001-loader-mt101
[xlsx]: https://github.com/sebastienrousseau/pain001-loader-xlsx
[mockbank]: https://github.com/sebastienrousseau/pain001-mockbank
