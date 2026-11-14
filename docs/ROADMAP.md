<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->

# Roadmap

What is planned, in the order it is expected, and what is deliberately
out of scope. Versions move in `0.0.1` steps in lockstep with the
pain001 suite; the maintainer opens each release, so nothing here is a
date.

## 0.0.88 (this release)

- First tagged release and the first one on proxy.golang.org.
- Opt-in retries for idempotent calls (`WithRetry`, ADR 0004).
- `User-Agent: pain001-sdk-go/<version>` on every request.
- One example per method, run against a live server in CI.
- Benchmarks for every method, a 100% statement-coverage gate, a
  Halstead gate, an API-breakage check, the user manual, and a
  tag-triggered release workflow with a dry run.

## Next

- **Typed async jobs.** `/api/v1/generate/async`, `/status`,
  `/download` and `/jobs` are reachable only through `Request` today.
  Typed methods follow once the suite agrees one surface for them, so
  all five SDKs add them together.
- **Kit updates.** Each core release that changes the conformance kit
  is vendored here in the same suite release, and the client changes
  with it.
- **Go floor.** Moves to 1.27 when Go 1.26 leaves support (on the
  release of Go 1.28), in its own release with a changelog line.

## Not planned

- **Building XML in the client.** The server is the one engine
  (ADR 0001).
- **Third-party dependencies** (ADR 0002).
- **Retrying POST calls.** A lost response to `Generate` must not
  resubmit a payment file (ADR 0004).
- **Logging.** Request bodies hold payment data.
