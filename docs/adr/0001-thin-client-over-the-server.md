<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->

# 0001. A thin client over `pain001 serve`

- **Status:** Accepted
- **Date:** 2026-11-10 (first commit, `1f60a17`); recorded 2026-11-14
- **Deciders:** maintainer

## Context

pain001's engine (templates for 13 message types, JSON Schema and XSD
validation, the MDR and scheme rulebooks, Schematron presets and the
faces) is Python, and it changes every release. Go services need the
same results: a file one client accepts must not be refused by another.

## Options considered

1. Port the engine to Go. No server to run, but two implementations of
   every rule, which drift; the conformance kit could only catch drift
   in the cases it lists.
2. Generate a client from the server's OpenAPI document. Fast to
   produce, but generated types are loose (every field optional),
   errors are untyped, and the generator becomes a dependency.
3. Write a small typed client by hand over the REST API, and hold it to
   core's conformance kit in CI against a real server.

## Decision

Option 3. The client builds no XML and applies no rule; it sends JSON
to `pain001 serve` and decodes the reply into Go types. Every method
maps to one documented endpoint, and `Request` reaches any other.

## Consequences

- The client is useless without a reachable server; the README's "When
  not to use" section says so.
- Every rule, template and scheme comes from the server, so a new
  server release changes behaviour without a new client release. New
  message types work as plain `MessageType` strings.
- A call costs an HTTP round trip plus the server's work; the client's
  own cost is small by comparison (`docs/BENCHMARKS.md`).
- The kit, run in CI against the pinned server, is the contract. A
  client release vendors the kit of the matching core release.
