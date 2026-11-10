<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->

# AGENTS.md

`pain001-sdk-go` is the Go client for the pain001 REST API
(`pain001 serve`). The server generates and validates ISO 20022 pain.001
and pain.008 files; this module calls it and returns typed results. It
exists so Go services use the same engine as every other pain001
client, held to the same conformance kit. Read
[CONTRIBUTING.md](CONTRIBUTING.md) as well.

## Rules

- **Versions** are `0.0.x` and move by `0.0.1`, in lockstep with the
  pain001 suite. Never bump one as a side effect; the maintainer opens
  each release. Client 0.0.N vendors the kit from core 0.0.N, and
  `ci/server-requirements.in` pins the server that CI tests against.
- **`conformance/` is vendored from core, never edit it here.** Replace
  it whole from core's `conformance/` and regenerate
  `conformance/SHA256SUMS` (sorted `shasum -a 256` of every other file
  in it); `kit_test.go` fails otherwise.
- **Standard library only.** No third-party module in `go.mod`;
  analysers run through `go run <module>@<pinned version>`.
- **Every method takes a `context.Context` first**, and a non-2xx
  response is an `*APIError`. No retries, no logging, and no request
  body or rejected input in an error message.
- **Complexity**: cyclomatic 10 and cognitive 15 per function (gocyclo,
  gocognit), 60 lines per function and 500 per file (`size_test.go`).
- **Data**: synthetic only. No real account numbers or personal data.
- **Headers**: every source file opens with the SPDX identifier
  `Apache-2.0 OR MIT`; Markdown carries it as an HTML comment on line 1.
- **Commits**: conventional, signed off (`git commit -s`) and signed.
  Never rewrite published history.

## The gate

```sh
make check                                          # without a server
PAIN001_BASE_URL=http://127.0.0.1:8000 make check   # with one
```

A change is done when both pass and add no warnings.

## Leave alone

`conformance/` (vendored), `LICENSE`, `LICENSE-APACHE`, `LICENSE-MIT`
and `DCO.txt` (legal texts), and the server pin outside a release.
