<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->

<!--
The hand-written Highlights of the next GitHub release, read by
scripts/release-notes.sh. Two to four bullets, each
"* **<Headline>**: <one or two plain sentences>". No em dashes.
Rewrite it in the release-preparation commit of every version.
-->

* **In lockstep with pain001 0.0.89**: CI runs the client against a live pain001 0.0.89 server, and the vendored conformance kit matches core's 0.0.89 kit byte for byte.
* **Every release asset in the checksums**: the release notes' Checksums block now lists `SHA256SUMS` itself, which 0.0.88's notes left out.
