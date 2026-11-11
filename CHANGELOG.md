<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->

# Changelog

All notable changes to this project are documented in this file. The
format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and versions advance in `0.0.1` steps in lockstep with the pain001 suite.

## [Unreleased]

The first release will be 0.0.86, with pain001 0.0.86.

### Added

- `pain001.NewClient` with `WithAPIKey`, `WithHTTPClient` and
  `WithTimeout`; typed `Health`, `Validate`, `Generate`, `Check`,
  `Faces` and `ProjectFace` methods, each taking a `context.Context`;
  and `Request` for any endpoint.
- `*APIError` for non-2xx responses, carrying the status and the
  server's `detail`, with `DetailText` and `Issues` accessors.
- `MessageType` constants for the 13 message types pain001 0.0.85
  serves.
- The conformance kit, vendored from core with `SHA256SUMS`, a runner
  for its cases, and typed tests against a live server.
- `examples/basic`, which generates pain.001.001.09 from the kit's
  credit-transfer fixture.
- CI: gofmt, vet, staticcheck, gocyclo, gocognit, race-enabled tests on
  Go 1.26 and 1.27, and the conformance job against `pain001 serve`
  0.0.86 installed from hash-pinned requirements.
