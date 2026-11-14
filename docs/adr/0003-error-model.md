<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->

# 0003. The error model

- **Status:** Accepted
- **Date:** 2026-11-10 (`*APIError`); 2026-11-12 (`*RefusalError`,
  `214771e`); 2026-11-14 (`ErrNoDocument`, `6cb5b1a`); recorded
  2026-11-14
- **Deciders:** maintainer

## Context

The server answers a bad request in three shapes: a `detail` string
(a 400 for a bad IBAN), a `detail` array of validation issues (a 422
for a malformed request), and, from pain001 0.0.87 (kit 2), a 422 whose
body is a generate response with `success: false` and the findings. A
server older than 0.0.87 sends that refusal with `200`. Error messages
must never carry the rejected input, which is payment data.

## Options considered

1. Return `(status, body, error)` from every method and let the caller
   inspect the body. Flexible, but every caller re-implements the
   mapping.
2. One error type per status. Many types, and a 422 can mean two
   different things.
3. One `*APIError` for any non-2xx, a `*RefusalError` that wraps it for
   a refused Generate, and a sentinel for a 2xx without the document.

## Decision

Option 3:

- `*APIError{Status, Detail}` for every non-2xx, with `DetailText()`
  and `Issues()` for the two `detail` shapes.
- `*RefusalError` wraps the `*APIError` (so `Status` stays 422 and
  `errors.As` finds either) and carries the decoded `GenerateResponse`
  with `ValidationErrors` and `SchemeViolations`.
- `ErrNoDocument` for a 2xx Generate without XML, so a nil error from
  `Generate` always comes with the document.
- `Error()` strings list issue locations and counts, never values.
  Transport errors are wrapped with the method and path only.
- `Request` returns any status as data, not as an error, because the
  conformance runner must see every response.

## Consequences

- Callers check `*RefusalError` before `*APIError`; the examples and
  the package documentation show that order.
- The fuzz targets (`FuzzErrorOmitsRejectedInput`,
  `FuzzRefusalOmitsFindings`) hold the no-payment-data rule.
- A new server error shape lands in `*APIError.Detail` as raw JSON, so
  it is never lost even before the client learns to decode it.
