<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->

# Ecosystem comparison

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

## Evidence

Each row's claims come from that repository's manifest (language floor
and runtime dependencies) and its CI workflow (the conformance job).

| Project | Conformance job | Manifest |
| :--- | :--- | :--- |
| pain001-sdk-go | [`ci.yml`, job `conformance`](https://github.com/sebastienrousseau/pain001-sdk-go/blob/main/.github/workflows/ci.yml) | [`go.mod`](https://github.com/sebastienrousseau/pain001-sdk-go/blob/main/go.mod) |
| pain001-sdk-java | [`ci.yml`, job `conformance`](https://github.com/sebastienrousseau/pain001-sdk-java/blob/main/.github/workflows/ci.yml) | [`pom.xml`](https://github.com/sebastienrousseau/pain001-sdk-java/blob/main/pom.xml) |
| pain001-sdk-rust | [`ci.yml`, job `conformance`](https://github.com/sebastienrousseau/pain001-sdk-rust/blob/main/.github/workflows/ci.yml) | [`Cargo.toml`](https://github.com/sebastienrousseau/pain001-sdk-rust/blob/main/Cargo.toml) |
| pain001-sdk-dotnet | [`ci.yml`, job `conformance`](https://github.com/sebastienrousseau/pain001-sdk-dotnet/blob/main/.github/workflows/ci.yml) | [`Pain001.Sdk.csproj`](https://github.com/sebastienrousseau/pain001-sdk-dotnet/blob/main/src/Pain001.Sdk/Pain001.Sdk.csproj) |
| pain001-sdk-cpp | [`ci.yml`, job `conformance`](https://github.com/sebastienrousseau/pain001-sdk-cpp/blob/main/.github/workflows/ci.yml) | [`CMakeLists.txt`](https://github.com/sebastienrousseau/pain001-sdk-cpp/blob/main/CMakeLists.txt) |
| openapi-generator client | [core `sdk.yml`](https://github.com/sebastienrousseau/pain001/blob/main/.github/workflows/sdk.yml) | Generated from core's OpenAPI document |

For this repository specifically:

- **Go 1.26**: the `go 1.26` directive in `go.mod`, and the CI test
  matrix (Go 1.26 and 1.27).
- **Conformance kit in CI**: the `conformance` job runs
  `make regression`, which starts the pinned `pain001 serve` and runs
  every case in `conformance/cases.json` through `Client.Request`
  (`conformance_test.go`); `kit_test.go` pins the vendored kit to
  `conformance/SHA256SUMS`.
- **No runtime dependencies**: `go.mod` has no `require` line and the
  module has no `go.sum`.
- **Blocking, `context.Context`**: every method takes a
  `context.Context` first and returns when the call ends.

## Against building XML in Go

There is no Go library compared here, because this package builds no
XML: it sends the work to the pain001 server, so Go code gets the same
templates, schema checks and rulebooks as the Python CLI and every
other client. If you need XML built in-process with no server, a Go
XML library and the ISO 20022 schemas are the alternative; see
"When not to use" in the README.
