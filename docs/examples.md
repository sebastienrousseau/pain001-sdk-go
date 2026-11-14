<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->

# Examples

Each example is a small `package main` under
[`examples/`](https://github.com/sebastienrousseau/pain001-sdk-go/tree/main/examples).
Run them from the repository root, so the conformance kit's fixture
paths resolve, against a running server:

```sh
pain001 serve --host 127.0.0.1 --port 8000 &
PAIN001_BASE_URL=http://127.0.0.1:8000 go run ./examples/health
PAIN001_BASE_URL=http://127.0.0.1:8000 make examples   # all of them
```

`PAIN001_API_KEY`, when set, is sent with every call. CI builds every
example and runs it against `pain001 serve` (`make regression`); each
exits non-zero when the server does not answer as it expects.

| Example | Shows |
| :--- | :--- |
| `health` | `Health`: server status and version |
| `validate` | `Validate`, without and with the `sepa-sct` rulebook |
| `basic` | `Generate`: the kit's credit transfer to pain.001.001.09 XML on disk, or each finding of a refusal |
| `check` | `Check` with a scheme profile and a Schematron preset |
| `faces` | `Faces`: each face and its bindings |
| `projectface` | `ProjectFace` onto Berlin Group, with the loss report |
| `request` | `Request`, the raw call, for a 200 and a 422 |
| `errors` | `*APIError` (string detail and issues), `*RefusalError`, `ErrNoDocument` |
| `timeout` | `WithTimeout` and a context deadline |
| `apikey` | `WithAPIKey` against a server started with `PAIN001_API_KEY` |
| `retry` | `WithRetry` and `WithHTTPClient`, against a closed port |
