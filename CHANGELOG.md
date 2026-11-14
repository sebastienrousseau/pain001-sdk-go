<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->

# Changelog

All notable changes to this project are documented in this file. The
format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and versions advance in `0.0.1` steps in lockstep with the pain001 suite.

## [Unreleased]

The first release will be 0.0.88, in lockstep with the pain001 suite.
CI tests against pain001 0.0.87, the latest released server.

### Added

- `Version`, the module's release, and a `User-Agent:
  pain001-sdk-go/<Version>` header on every request, so a server's
  access log shows which SDK and release sent a call.
- `WithRetry(RetryPolicy{...})`: opt-in retries for idempotent calls
  only (`Health`, `Faces`, and `Request` with GET). It retries a failed
  connection and a `429` or `503`, waits with exponential backoff and
  jitter or for the response's `Retry-After`, and never retries a POST.
  Off by default, so a client without it behaves as before.
- `pain001.NewClient` with `WithAPIKey`, `WithHTTPClient` and
  `WithTimeout`; typed `Health`, `Validate`, `Generate`, `Check`,
  `Faces` and `ProjectFace` methods, each taking a `context.Context`;
  and `Request` for any endpoint.
- `*APIError` for non-2xx responses, carrying the status and the
  server's `detail`, with `DetailText` and `Issues` accessors.
- `MessageType` constants for the 13 message types pain001 0.0.85
  serves.
- The conformance kit (kit 2, from pain001 0.0.87: 14 cases, including
  the `422` generate refusals, a bad LEI and a CHAPS purpose code),
  vendored from core with `SHA256SUMS`, a runner for its cases, and
  typed tests against a live server.
- `*RefusalError`, returned by `Generate` when the server refuses the
  data with `422` and `success: false` (pain001 0.0.87 and later). It
  wraps the `*APIError`, so `Status` stays 422 and `errors.As` finds
  either type, and its `Response` carries the decoded
  `ValidationErrors` and `SchemeViolations`. A nil error from
  `Generate` now always comes with the XML.
- `examples/basic`, which generates pain.001.001.09 from the kit's
  credit-transfer fixture and prints each finding of a refusal.
- One runnable program per method under `examples/`: `health`,
  `validate`, `basic` (generate), `check`, `faces`, `projectface`,
  `request`, `errors` (`*APIError`, `*RefusalError`, `ErrNoDocument`),
  `timeout`, `apikey` and `retry`. `make examples` builds them all and
  runs each against `PAIN001_BASE_URL`; each exits non-zero when the
  server does not answer as it expects. `make conformance` runs them.
- Fuzz targets for the client's handling of untrusted responses:
  `FuzzRequest`, `FuzzTypedMethods`, `FuzzAPIError`,
  `FuzzErrorOmitsRejectedInput` and `FuzzRefusalOmitsFindings`, seeded
  from the conformance kit's responses. `go test` runs their seeds,
  `make fuzz` fuzzes each for a minute, and a CI job fuzzes each for 30
  seconds on every push and pull request.
- `ErrNoDocument`: `Generate` now returns an error wrapping it when a
  `2xx` response carries no XML (`success: false` or an empty `xml`),
  as a server older than pain001 0.0.87 sends for a refusal. It used to
  return that body with a nil error, against its documented promise
  that a nil error always comes with the XML. Found by
  `FuzzTypedMethods`.
- CI: gofmt, vet, staticcheck, gocyclo, gocognit, race-enabled tests on
  Go 1.26 and 1.27, and the conformance job against `pain001 serve`
  0.0.87 installed from hash-pinned requirements.
- `scripts/verify-release-versions`, run by `make check` and CI: fails
  when `version.go`, `CITATION.cff`, the `SECURITY.md` and `SUPPORT.md`
  support tables, a Markdown install snippet or the `CHANGELOG.md`
  heading disagree, or when the server pin and its lock name different
  pain001 releases. `--release` also requires the dated release heading.
- A 100.0% statement-coverage floor: `make test` runs
  `scripts/coverage.sh`, which fails below it, on both Go versions in
  CI. `halstead_test.go` holds every function to a Halstead difficulty
  of 30, the last of the workspace's complexity ceilings without a gate.
- `SUPPORT.md`: where to ask, what to include, and which versions get
  fixes.
- The README demo, `.github/demo.gif`, rendered by `make demo` from
  `.github/demo.tape` with vhs: the typed live tests and
  `examples/basic` against a `pain001 serve` the tape starts on
  loopback. CONTRIBUTING.md shows how to install the pinned server for
  it.

### Fixed

- `Request` reported a body with data after its first JSON value (`000`,
  `{"a":1} trailing`, `0` followed by a form feed) as that value with no
  error, so a garbled response passed the conformance runner while the
  typed methods rejected it. It now returns an error for anything but
  JSON whitespace after the value. Found by `FuzzRequest`.
