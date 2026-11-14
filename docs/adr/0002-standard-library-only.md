<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->

# 0002. The Go standard library only

- **Status:** Accepted
- **Date:** 2026-11-10 (first commit, `1f60a17`); recorded 2026-11-14
- **Deciders:** maintainer

## Context

The client carries payment files to a server and decodes JSON. Banks
and payment service providers that adopt it review every dependency it
brings into their build, and each one is a supply-chain risk and an
upgrade obligation.

## Options considered

1. Use a third-party HTTP client (retries, middleware) and a faster
   JSON codec. Less code here, more review for every adopter.
2. Use `net/http` and `encoding/json` only.

## Decision

Option 2. `go.mod` lists no requirement and the module has no `go.sum`.
Analysers (staticcheck, gocyclo, gocognit, govulncheck, gorelease,
cyclonedx-gomod) run through `go run <module>@<pinned version>`, so
they never enter the module graph.

## Consequences

- govulncheck covers the standard library and toolchain only, and
  Dependabot watches the CI actions and the Python server pin.
- Retries (ADR 0004) and the response cap are implemented here, in a
  few dozen lines each, rather than taken from a library.
- JSON decoding speed is `encoding/json`'s; the benchmarks show it is
  not where a call's time goes.
