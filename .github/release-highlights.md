<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->

<!--
The hand-written Highlights of the next GitHub release, read by
scripts/release-notes.sh. Two to four bullets, each
"* **<Headline>**: <one or two plain sentences>". No em dashes.
Rewrite it in the release-preparation commit of every version.
-->

* **In lockstep with pain001 0.0.91**: CI runs the client against a live pain001 0.0.91 server, and the vendored conformance kit, unchanged since 0.0.90, matches core's byte for byte.
* **Manual built with mkdocs-material 9.8.2**: it fixes GHSA-wxw8-4jgj-q85j, a path traversal in the theme's privacy plugin, and `ci/docs-requirements.in` now refuses older versions.
