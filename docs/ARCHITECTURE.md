<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->

# Architecture

pain001-sdk-go is a thin, typed HTTP client for the pain001 REST API
(`pain001 serve`). The server does all the payment work; the client
encodes requests, sends them, and decodes the answers into Go types
and errors. This page is for contributors: where each concern lives
and how a call flows through it.

## Layout

| Path | What it holds |
| :--- | :--- |
| `client.go` | `Client`, `NewClient`, the `With*` options, `Request`, and the transport loop (`send`, `attempt`, `newRequest`) |
| `methods.go` | One typed method per endpoint: `Health`, `Validate`, `Generate`, `Check`, `Faces`, `ProjectFace`, and the request bodies they send |
| `types.go` | Response types and the `MessageType` constants |
| `errors.go` | `APIError`, `Issue`, `RefusalError` and the mapping from a non-2xx body |
| `retry.go` | `RetryPolicy`, `WithRetry`, backoff and `Retry-After` parsing |
| `version.go` | `Version` and the `User-Agent` it produces |
| `doc.go` | Package documentation, rendered on pkg.go.dev |
| `internal/kit/` | The conformance runner: loads `conformance/cases.json`, substitutes `$file` fixtures, checks status, `equals`, `types` and `contains` on JSON Pointers |
| `conformance/` | Core's conformance kit, vendored byte for byte and pinned by `SHA256SUMS` |
| `examples/` | One runnable program per method, run against a live server in CI |
| `*_test.go` | Unit tests against stub servers, live tests, fuzz targets, benchmarks and the size and Halstead gates |
| `testdata/` | Fuzz regression inputs and the recorded responses the benchmarks decode |
| `scripts/` | The release-version check, the coverage gate, the regression run, the API check and the release notes |

## A call, end to end

```text
caller ── Client.Generate(ctx, csv, ...) ─┐
                                          │ methods.go: build filePayload
                                          ▼
                             Client.call ── Client.send ── per-call timeout (WithTimeout)
                                                  │
                                       ┌──────────┴───────────┐
                                       │ attempt:             │
                                       │  newRequest: JSON,   │
                                       │  Accept, User-Agent, │
                                       │  Authorization       │
                                       │  http.Client.Do      │
                                       │  read <= 64 MiB      │
                                       └──────────┬───────────┘
                                                  │ retrier.wait: GET only,
                                                  │ 429/503/connection error,
                                                  │ backoff or Retry-After
                                                  ▼
                             Client.call: 2xx → json.Unmarshal into the type
                                          else → newAPIError(status, body)
                                                  │
                             Generate: asRefusal(err) → *RefusalError on a
                                       422 with success false;
                                       2xx without XML → ErrNoDocument
```

`Request` shares `send` but skips `call`: it returns the status and the
body decoded with `json.Number` whatever the status, which is what the
conformance runner needs to compare numbers exactly.

## Invariants

- **No payment data leaves through errors or logs.** The client logs
  nothing. `APIError.Error()` lists issue locations and messages, never
  the rejected input; `RefusalError.Error()` gives counts, never the
  findings. Two fuzz targets hold this.
- **Bounded resources.** Each call has a deadline (30 seconds by
  default) and reads at most 64 MiB of response.
- **One attempt unless asked.** Retries need `WithRetry` and apply to
  GET only (ADR 0004).
- **Standard library only.** `go.mod` has no requirements (ADR 0002).
- **Exported API is the breaking axis.** `scripts/api-check.sh` runs
  gorelease against the last tag.

## Tests and gates

- Unit tests run against `httptest` stubs and a function transport,
  with no server, and cover 100% of statements (`scripts/coverage.sh`).
- Live tests (`typed_test.go`, `conformance_test.go`) and the examples
  run when `PAIN001_BASE_URL` is set; CI sets it in the conformance job
  through `make regression`.
- `size_test.go` and `halstead_test.go` hold the function-size and
  Halstead ceilings; gocyclo and gocognit hold the cyclomatic and
  cognitive ones.
- Five fuzz targets feed arbitrary statuses and bodies through
  `Request`, the typed methods and the error mapping.

## Where the decisions are recorded

[`docs/adr/`](adr/README.md): the thin-client design, the dependency
policy, the error model and the retry policy.
