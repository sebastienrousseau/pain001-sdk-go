<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->

<!--
The hand-written Highlights of the next GitHub release, read by
scripts/release-notes.sh. Two to four bullets, each
"* **<Headline>**: <one or two plain sentences>". No em dashes.
Rewrite it in the release-preparation commit of every version.
-->

* **In lockstep with pain001 0.0.90**: CI runs the client against a live pain001 0.0.90 server, and the vendored conformance kit matches core's kit byte for byte.
* **Tested against a server without the first-request race**: core 0.0.90 fixes a race in `lazy_module` that made a freshly started `pain001 serve` answer its first parallel requests with HTTP 500.
