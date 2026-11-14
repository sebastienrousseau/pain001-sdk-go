<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->

# Contributing

Thanks for looking. This module is a thin client on purpose: the pain001
server does the payment work, and the client's job is to call it
correctly and return typed results.

## Before you open a pull request

Everything CI checks, you can run locally with Go 1.26 or newer:

```sh
make check
```

That runs `gofmt -l`, `go vet`, staticcheck, the gocyclo and gocognit
ceilings, `go build` and `go test -race`. To add the conformance kit,
the typed tests and the example, start a server and point the tests at
it:

```sh
pip install 'pain001[api]==0.0.87'
pain001 serve --host 127.0.0.1 --port 8000 &
PAIN001_BASE_URL=http://127.0.0.1:8000 make check
```

Use synthetic data only. The kit's fixtures are synthetic; never add
real account numbers or personal data to a test.

## What the tests are for

- `client_test.go`, `errors_test.go`: requests, headers, timeouts and
  error mapping against a stub server; no pain001 needed.
- `internal/kit/`: the conformance runner (JSON Pointer, `$file`
  substitution, expectations), unit-tested without a server.
- `conformance_test.go`: every case in `conformance/cases.json` through
  `Client.Request`; `typed_test.go`: one test per typed method. Both
  skip when `PAIN001_BASE_URL` is unset.
- `kit_test.go`: the vendored kit matches `conformance/SHA256SUMS`, and
  `MessageTypes()` matches the kit's `openapi.json`.
- `size_test.go`: every function within 60 lines and every file within
  500.
- `fuzz_test.go`: native Go fuzz targets that feed arbitrary statuses
  and bodies to `Request`, the typed methods and the error mapping, and
  check that the client never panics, returns a value or a documented
  error, and never puts the rejected input or a refusal's findings in an
  error message. `go test` runs their seeds and `testdata/fuzz`;
  `make fuzz` fuzzes each for `FUZZTIME` (default 60s). Commit any
  crashing input it writes to `testdata/fuzz` with the fix.

## The conformance kit

`conformance/` is vendored from pain001 core byte for byte. **Do not
edit it here.** A change to a case belongs in core; this repository
takes the new kit by copying core's `conformance/` for the matching
version and regenerating `conformance/SHA256SUMS`.

## Commits

Conventional commits, each signed off (`git commit -s`) under the
[Developer Certificate of Origin](DCO.txt); the DCO workflow checks
every commit in a pull request.

## Versioning

Versions increment by 0.0.1 in lockstep with the pain001 suite;
`0.1.0` follows `0.0.999`. The maintainer decides when a release opens.

## Licence

Apache-2.0 OR MIT, at your option. By contributing you agree your work is
released under the same dual grant.
