<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->

<!--
The hand-written Highlights of the next GitHub release, read by
scripts/release-notes.sh. Two to four bullets, each
"* **<Headline>**: <one or two plain sentences>". No em dashes.
Rewrite it in the release-preparation commit of every version.
-->

* **First release, in lockstep with pain001 0.0.88**: the typed Go client for `pain001 serve` is now on proxy.golang.org, held to core's conformance kit in CI against a live server.
* **Opt-in retries for idempotent calls**: `WithRetry` retries `Health`, `Faces` and GET requests on a failed connection, `429` or `503`, with backoff and `Retry-After`. POST calls are never retried.
* **An example and a benchmark for every method**: eleven runnable programs and client-side plus loopback benchmarks, all run in CI, with the numbers in `docs/BENCHMARKS.md`.
