<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->

# Architecture decision records

Each record states a decision that will be questioned later, the
context it was made in, the options weighed and what follows from it.
A record is never rewritten to change its decision; a later record
supersedes it and says so.

| Record | Decision | Status |
| :--- | :--- | :--- |
| [0001](0001-thin-client-over-the-server.md) | A thin client over `pain001 serve`, not a Go port of the engine | Accepted |
| [0002](0002-standard-library-only.md) | The Go standard library only, no third-party modules | Accepted |
| [0003](0003-error-model.md) | `*APIError`, `*RefusalError` and `ErrNoDocument` as the error model | Accepted |
| [0004](0004-opt-in-retries.md) | Retries are opt-in and limited to GET | Accepted |

New records take the next number and the same headings: Status, Date,
Deciders, Context, Options considered, Decision, Consequences.
