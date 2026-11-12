<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->

# Security Policy

## Supported versions

This module follows the [`pain001`](https://github.com/sebastienrousseau/pain001)
suite cadence. While pre-`1.0`, the latest released `0.0.x` receives
security fixes; older `0.0.x` versions do not.

| Version | Status | Receives security fixes? |
| :--- | :--- | :--- |
| Unreleased (first release 0.0.86) | In development | Fixes land before release |

## Reporting a vulnerability

**Do not open a public issue for security reports.**

Use one of the following private channels, in this order of
preference:

1. **GitHub Private Vulnerability Reporting**
   <https://github.com/sebastienrousseau/pain001-sdk-go/security/advisories/new>
   (the maintainers are notified automatically; do this first)
2. **Email**: `security@pain001.com` (PGP available on request)

Please include:

- A description of the vulnerability and its potential impact
- A minimal reproducer (preferably a failing test case)
- The module version and the pain001 server version
- Any proposed remediation, if known

These are goals, not guarantees: pain001 is maintained in a personal
capacity, with no SLA and no out-of-hours cover.

**Acknowledgement**: the goal is within 5 working days.
**Triage**: a severity assessment and remediation plan after
acknowledgement.

## Security posture

### Scope

The module is an HTTP client. It sends JSON to the server the caller
configures and decodes the reply. It reads and writes no files, parses
no XML and runs no code from the server.

### Threat model

| Surface | How it's handled |
| :--- | :--- |
| **Payment data in logs** | The client logs nothing. `APIError.Error()` lists a 422's locations and messages and leaves out the rejected input; `RefusalError.Error()` gives the server's summary and the number of findings, not the findings, whose messages can quote the input. |
| **Credential leakage** | The API key goes only in the `Authorization` header, never in the URL or error text. Go's `http.Client` drops it on a redirect to another host. |
| **Resource exhaustion** | Response bodies are capped at 64 MiB; every call carries a timeout (30 seconds by default). |
| **Path injection** | The face name is path-escaped before it joins the URL. |
| **Transport** | Plain HTTP is accepted for loopback development; use HTTPS elsewhere. TLS settings come from the `http.Client` you pass. |
| **Dependency CVEs** | No dependencies outside the Go standard library; Dependabot watches `go.mod` and the CI actions. |

### Cryptography status

The module implements no cryptography of its own; TLS is Go's
`crypto/tls` through `net/http`.
