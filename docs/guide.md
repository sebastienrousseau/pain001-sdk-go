<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->

# Using the client

## Creating a client

```go
client, err := pain001.NewClient("https://pain001.internal.example",
    pain001.WithAPIKey(os.Getenv("PAIN001_API_KEY")),
    pain001.WithTimeout(10*time.Second))
```

`NewClient` takes an absolute `http` or `https` base URL; a trailing
slash is dropped. A `Client` is safe for concurrent use; build one and
share it.

| Option | Effect | Default |
| :--- | :--- | :--- |
| `WithAPIKey(key)` | Sends `Authorization: Bearer <key>`, for a server started with `PAIN001_API_KEY`. An empty key sends no header | No header |
| `WithHTTPClient(hc)` | Sends requests through `hc`: a proxy, TLS settings, a custom transport. `nil` is ignored | A fresh `http.Client` |
| `WithTimeout(d)` | Bounds each call, reading the response included, on top of the context's deadline. Zero or negative turns it off | `DefaultTimeout`, 30 seconds |
| `WithRetry(policy)` | Retries GET calls on a failed connection, `429` or `503` (below) | One attempt |

Every request carries `Accept: application/json` and
`User-Agent: pain001-sdk-go/<Version>`.

## Methods

Every method takes a `context.Context` first.

| Method | Endpoint | Returns |
| :--- | :--- | :--- |
| `Health(ctx)` | `GET /api/v1/health` | `*HealthResponse`: status and server version |
| `Validate(ctx, content, filename, messageType, scheme)` | `POST /api/v1/ui/validate` | `*ValidationResponse`: row counts, row errors, scheme violations |
| `Generate(ctx, content, filename, messageType, scheme)` | `POST /api/v1/ui/generate` | `*GenerateResponse` with the XML, or a `*RefusalError` |
| `Check(ctx, xml, scheme, schematron)` | `POST /api/v1/check` | `*CheckResponse`: verdict, rungs, findings |
| `Faces(ctx)` | `GET /api/v1/faces` | `[]FaceInfo`: each face and its bindings |
| `ProjectFace(ctx, face, xml, version, bindings)` | `POST /api/v1/faces/{face}` | `*FaceResponse`: payload and loss report |
| `Request(ctx, method, path, body)` | any | status and JSON body, whatever the status |

`filename`'s suffix (`.csv`, `.json`, `.jsonl`) picks the parser. An
empty `messageType` leaves the server's default; an empty `scheme` runs
no rulebook, and a comma-separated list runs several (`sepa-sct,anti-duplicate`).
`MessageType` constants cover the 13 types the server supports; any
other string converts, so a newer server's types work without a new
client.

## Errors

```go
res, err := client.Generate(ctx, csv, "payments.csv", pain001.Pain00100109, "")
var refusal *pain001.RefusalError
var apiErr *pain001.APIError
switch {
case errors.As(err, &refusal): // 422: the server refused the data
    for _, row := range refusal.Response.ValidationErrors {
        fmt.Println(row.Field, row.Message)
    }
case errors.As(err, &apiErr): // any other non-2xx
    fmt.Println(apiErr.Status, apiErr.Error())
case errors.Is(err, pain001.ErrNoDocument): // a pre-0.0.87 server's refusal
case err != nil: // no answer: transport, timeout, cancellation
default:
    use(res.XML)
}
```

- `*APIError{Status, Detail}` is every non-2xx response. `DetailText()`
  returns a string detail (a 400 for a bad IBAN); `Issues()` returns a
  422's validation issues.
- `*RefusalError` is a refused `Generate` (422, `success: false`). It
  wraps the `*APIError`, so check it first; its `Response` carries the
  `ValidationErrors` and `SchemeViolations`.
- `ErrNoDocument` is a 2xx `Generate` without XML, which only a server
  older than pain001 0.0.87 sends. A nil error from `Generate` always
  comes with the XML.
- `ErrResponseTooLarge` is a response over 64 MiB.
- Error messages never include the rejected input or a refusal's
  findings: both can hold payment data.

## Timeouts and cancellation

The client's timeout (`WithTimeout`, 30 seconds by default) and the
context's deadline both apply; the shorter wins. An expired deadline
returns an error that `errors.Is(err, context.DeadlineExceeded)`
matches.

## Retries

```go
client, err := pain001.NewClient(base, pain001.WithRetry(pain001.RetryPolicy{
    MaxAttempts: 4,                      // the first attempt included
    BaseDelay:   200 * time.Millisecond, // doubled after each attempt
    MaxDelay:    5 * time.Second,        // cap on one wait
}))
```

Without `WithRetry` the client makes one attempt per call. With it,
GET calls (`Health`, `Faces`, `Request` with GET) are retried when the
connection fails or the server answers `429` or `503`, after an
exponential backoff with jitter, or after the response's `Retry-After`.
A `Retry-After` longer than `MaxDelay` ends the retries and returns that
response. POST calls are never retried: repeating `Generate` could
submit a file twice. The timeout and the context bound all attempts
together. See [ADR 0004](adr/0004-opt-in-retries.md).

## The raw call

`Request(ctx, method, path, body)` reaches any endpoint, including the
async job endpoints the typed methods do not wrap. It returns the
status and the body decoded as JSON with `json.Number` (so integers and
decimals survive exactly) whatever the status; only a transport failure
or a body that is not JSON is an error.

## Security

The client writes nothing to disk and logs nothing. The API key travels
only in the `Authorization` header, and Go's `http.Client` drops it on a
redirect to another host. Use HTTPS for any server that is not on
loopback.
