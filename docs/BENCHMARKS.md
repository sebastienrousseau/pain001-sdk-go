<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->

# Benchmarks

Every public method is benchmarked two ways, with Go's `testing.B`
(`bench_test.go`). The numbers below were measured once, on the machine
named in [Environment](#environment); they describe that machine on
that day, not a guarantee.

## Method

- **`BenchmarkClient/<Method>`: the client's own cost, no network.** A
  canned `http.RoundTripper` answers every request with a response body
  recorded from pain001 0.0.87 (`testdata/bench/*.json`, recorded by
  `scripts/record-bench-responses.sh` with the conformance kit's
  credit-transfer fixtures). Each iteration builds the request, JSON
  encodes the kit fixture into it (the 5-row CSV for `Validate` and
  `Generate`, the 5-transaction pain.001.001.09 document for `Check`
  and `ProjectFace`), sends it through the transport, and reads and
  decodes the response into the method's Go type.
- **`BenchmarkLive/<Method>`: one loopback round trip**, the same calls
  against `pain001 serve` on `127.0.0.1`. This is dominated by the
  server's work (parsing, validating, rendering XML) and includes the
  client's cost above. Skipped when `PAIN001_BASE_URL` is unset.
- Each benchmark ran 10 times for 1 second (`-count 10 -benchtime 1s`),
  with `-benchmem`; the tables give benchstat's median and the spread
  of the 10 samples (`±`, a 95% confidence interval).

Reproduce:

```sh
pain001 serve --host 127.0.0.1 --port 8000 &
PAIN001_BASE_URL=http://127.0.0.1:8000 make bench BENCHCOUNT=10
go run golang.org/x/perf/cmd/benchstat@latest bin/bench.txt
```

CI smoke-runs every benchmark once per pull request
(`make bench BENCHCOUNT=1 BENCHTIME=1x` in `make regression`), to keep
them compiling and running; it asserts nothing about speed.

## Environment

| | |
| :--- | :--- |
| Machine | Apple A18 Pro, 6 cores (GOMAXPROCS 6), 8 GiB, macOS 27.2 (26B123) |
| Go | 1.27.4, darwin/arm64 |
| Server | pain001 0.0.87 (`ci/server-requirements.txt`), Python 3.12.14, `pain001 serve --host 127.0.0.1`, no API key |
| Date | 2026-11-14, commit of this release branch |
| Load | **Heavily shared**: the 1-minute load average was 83 at the start and 50 at the end of the run, from other builds on the same machine. The wide `±` spreads below come from that, not from the client. |

## Results

### Client-side cost (no network)

| Method | Time per call | Memory per call | Allocations |
| :--- | ---: | ---: | ---: |
| `Health` | 5.7 µs ± 55% | 2.97 KiB | 26 |
| `Validate` | 15.5 µs ± 186% | 5.02 KiB | 33 |
| `Generate` | 104.8 µs ± 21% | 24.22 KiB | 43 |
| `Check` | 36.0 µs ± 23% | 9.72 KiB | 35 |
| `Faces` | 9.2 µs ± 19% | 3.39 KiB | 31 |
| `ProjectFace` | 49.3 µs ± 25% | 26.91 KiB | 96 |
| `Request` (GET health) | 6.5 µs ± 30% | 4.12 KiB | 44 |

### Loopback round trip (`pain001 serve` on 127.0.0.1)

| Method | Time per call | Memory per call | Allocations |
| :--- | ---: | ---: | ---: |
| `Health` | 0.78 ms ± 87% | 5.59 KiB | 67 |
| `Validate` | 6.0 ms ± 92% | 8.01 KiB | 79 |
| `Generate` | 16.2 ms ± 43% | 27.23 KiB | 89 |
| `Check` | 10.6 ms ± 72% | 15.52 KiB | 83 |
| `Faces` | 1.7 ms ± 70% | 6.00 KiB | 72 |
| `ProjectFace` | 16.7 ms ± 43% | 34.09 KiB | 145 |
| `Request` (GET health) | 1.0 ms ± 91% | 6.73 KiB | 85 |

Memory and allocation counts are stable across samples (spread 0 to
6%); times are not, on this loaded machine.

## Reading the numbers

- The client's own work is microseconds per call; a round trip is
  milliseconds, so for every method the client is under 1% of the call.
  `Generate` and `ProjectFace` cost the most client-side because they
  decode the largest responses (the 3.5 KB XML document; the face
  payload and loss report).
- The round-trip numbers are the server's, on loopback, with no TLS and
  no network latency. They say nothing about a deployment; measure
  yours with `BenchmarkLive` against it.
- No number here is compared with another library: the suite has no other Go client
  for this API to compare with.
