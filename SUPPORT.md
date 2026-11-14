<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->

# Getting support

pain001-sdk-go is maintained by one person in a personal capacity, with
no SLA. This page says where to ask, and which versions get fixes.

## Supported versions

The module releases in lockstep with the
[pain001 suite](https://github.com/sebastienrousseau/pain001). While
pre-`1.0`, only the latest `0.0.x` receives fixes.

| Version | Status | Receives fixes? |
| :--- | :--- | :--- |
| 0.0.88 (unreleased, the first release) | In development | Yes, before release |

The client is tested against the pain001 server pinned in
`ci/server-requirements.in` (0.0.87, the latest released server).

## Read first

- [README.md](https://github.com/sebastienrousseau/pain001-sdk-go/blob/main/README.md): install, quick start, every option and the
  example index.
- The [user manual](https://sebastienrousseau.github.io/pain001-sdk-go/)
  and the [API reference](https://pkg.go.dev/github.com/sebastienrousseau/pain001-sdk-go).
- [`examples/`](https://github.com/sebastienrousseau/pain001-sdk-go/tree/main/examples): one runnable program per method, each run
  against a live server in CI.

## Questions and bugs

Open an issue at
<https://github.com/sebastienrousseau/pain001-sdk-go/issues/new/choose>
with:

- the module version and `go version`;
- the pain001 server version (`client.Health` returns it);
- a minimal program that reproduces the problem, using synthetic data
  only (the conformance kit's fixtures are a good start);
- the full error text.

A problem in what the server accepts, refuses or generates belongs in
[pain001](https://github.com/sebastienrousseau/pain001/issues); this
client only carries requests and responses.

## Security reports

Never in a public issue. Follow [SECURITY.md](https://github.com/sebastienrousseau/pain001-sdk-go/blob/main/SECURITY.md).
