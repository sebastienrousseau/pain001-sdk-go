<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->

# Development

The single entry point for working on pain001-sdk-go: the toolchain,
how to reproduce every CI gate locally, how the tests are laid out, and
how a release is made. Contribution rules (DCO, synthetic data, the
vendored kit) are in
[CONTRIBUTING.md](https://github.com/sebastienrousseau/pain001-sdk-go/blob/main/CONTRIBUTING.md);
agent rules in
[AGENTS.md](https://github.com/sebastienrousseau/pain001-sdk-go/blob/main/AGENTS.md).

## Toolchain

| Tool | Version | Used for |
| :--- | :--- | :--- |
| Go | 1.26 or newer (CI: 1.26 and 1.27) | everything; `mise use go@1.27` or the devcontainer |
| Python | 3.12 | the pinned `pain001 serve` for the live tests (`make regression` installs it) |
| uv | any recent | optional; `make regression` uses it when present, else `python3 -m venv` and pip |
| Node.js | 22 | markdownlint (`npx`) |
| vhs, ttyd, ffmpeg | any recent | only to re-render the README demo (`make demo`) |

The Go analysers (staticcheck, gocyclo, gocognit, govulncheck,
gorelease, cyclonedx-gomod) run through `go run <module>@<version>`
at the versions pinned in the `Makefile` and `scripts/`, so nothing is
installed globally and `go.mod` stays free of tool dependencies. The
Python tools (pain001, codespell, reuse, MkDocs Material) install from
hash-pinned lockfiles in `ci/`.

The `.devcontainer/` boots with Go, Python and Node.js installed and
runs `make check` once to warm the module cache.

## Every CI gate, locally

| CI job (workflow) | Local command |
| :--- | :--- |
| `gofmt, vet, staticcheck, complexity` (CI) | `make versions fmt-check vet staticcheck complexity` |
| `govulncheck` (CI) | `make vulncheck` (`VULNDB=file:///path` for an offline database copy) |
| `build and unit tests (Go 1.26, 1.27)` (CI) | `make build test` |
| `fuzz response decoding` (CI) | `make fuzz FUZZTIME=30s` |
| `API-breakage check` (CI) | `scripts/api-check.sh` (needs a clean tree and the tags) |
| `conformance kit against pain001 serve` (CI) | `make regression` |
| `markdownlint` (Docs lint) | `make docs-lint` |
| `REUSE lint` (Docs lint) | `make docs-lint` |
| `manual (mkdocs --strict)` (Docs) | `make docs` |
| `CodeQL`, `OpenSSF Scorecard`, `DCO` | GitHub only; sign off every commit with `git commit -s` |

`make check` is the single local gate: versions, gofmt, vet,
staticcheck, the complexity ceilings, govulncheck, build, the tests at
100% coverage, and, when `PAIN001_BASE_URL` is set, the conformance
kit, the live typed tests and every example. `make regression` runs
`make check` against a server it starts itself, plus the benchmark
smoke, exactly as CI's conformance job does:

```sh
make regression                                   # installs the pinned server
PAIN001_SERVER=/path/to/venv/bin/pain001 make regression
REGRESSION_PORT=18800 make regression             # pick the port
```

It installs `ci/server-requirements.txt` into a temporary venv (unless
`PAIN001_SERVER` is set), starts `pain001 serve` on a free loopback
port with a random `PAIN001_API_KEY`, runs the gates, and stops the
server and deletes the venv on any exit. To run against a server you
started yourself:

```sh
uv venv --python 3.12 /tmp/pain001-srv
uv pip install --python /tmp/pain001-srv/bin/python \
  --require-hashes -r ci/server-requirements.txt
/tmp/pain001-srv/bin/pain001 serve --host 127.0.0.1 --port 8000 &
PAIN001_BASE_URL=http://127.0.0.1:8000 make check
```

## Quality gates and their thresholds

- **Coverage: 100.0% of statements**, enforced by `scripts/coverage.sh`
  (`make test`) on both Go versions. It measures the module root (the
  client) and `internal/kit` (the conformance runner). Go's coverage
  tool measures statements, not branches.
- **Coverage exclusions: none in the library.** The one scope decision
  is `examples/`: they are `package main` programs, not library code,
  and are excluded from the measured packages as a whole; CI instead
  builds each and runs it against a live server, where it exits
  non-zero on any unexpected answer. Go has no inline exclusion marker,
  and none is used.
- **Complexity per function**: cyclomatic 10 (gocyclo), cognitive 15
  (gocognit), Halstead difficulty 30 (`halstead_test.go`), 60 lines and
  500 lines per file (`size_test.go`). No function is over a ceiling,
  so there is no baseline file.
- **Lint**: gofmt, `go vet` and staticcheck, with no suppressions.
- **Versions**: `scripts/verify-release-versions` (below).

## Test layout

| File | What it tests | Needs a server |
| :--- | :--- | :--- |
| `client_test.go`, `errors_test.go`, `version_test.go` | requests, headers, timeouts, error mapping, User-Agent | No (stubs) |
| `retry_test.go` | the retry policy, every branch | No |
| `fuzz_test.go` | five fuzz targets over untrusted responses; seeds run in `go test` | No |
| `internal/kit/kit_test.go` | the conformance runner | No |
| `kit_test.go` | the vendored kit matches `SHA256SUMS`; `MessageTypes()` matches its OpenAPI | No |
| `size_test.go`, `halstead_test.go` | the size and Halstead ceilings | No |
| `conformance_test.go` | every kit case through `Client.Request` | Yes |
| `typed_test.go` | one live test per typed method | Yes |
| `bench_test.go` | `BenchmarkClient` (no network), `BenchmarkLive` (loopback) | `BenchmarkLive` only |
| `example_test.go` | the godoc example | No (compiled, not run) |

Tests that need a server skip when `PAIN001_BASE_URL` is unset.

## Benchmarks

```sh
PAIN001_BASE_URL=http://127.0.0.1:8000 make bench   # 6 runs of 1s each
make bench BENCHCOUNT=10                            # more samples
go run golang.org/x/perf/cmd/benchstat@latest bin/bench.txt
```

`BenchmarkClient` decodes responses recorded in `testdata/bench`;
re-record them with `scripts/record-bench-responses.sh` when the server
pin moves. [`docs/BENCHMARKS.md`](https://github.com/sebastienrousseau/pain001-sdk-go/blob/main/docs/BENCHMARKS.md)
holds the method and the last measured numbers.

## The manual

```sh
uv venv --python 3.12 /tmp/pain001-docs
uv pip install --python /tmp/pain001-docs/bin/python --require-hashes -r ci/docs-requirements.txt
PATH="/tmp/pain001-docs/bin:$PATH" make docs        # strict build into site/
```

`docs/` is the only documentation root. Root files (this one,
`CHANGELOG.md`, `SECURITY.md`, `SUPPORT.md`, `GOVERNANCE.md`) are
included into the manual by the stubs in `docs/project/`, so they link
to other repository files by absolute URL. The Pages workflow
(`.github/workflows/docs.yml`) deploys from main once the maintainer
switches it on (below).

## Releasing

Versions are `0.0.x`, move by `0.0.1` in lockstep with the pain001
suite, and the maintainer opens each release. A release is prepared on
the `feat/v<version>` branch and published from a tag:

1. **Prepare** (in the release pull request): set `Version` in
   `version.go`; rename `## [Unreleased]` in `CHANGELOG.md` to
   `## [<version>] - <date>`; update `CITATION.cff`, the `SECURITY.md`
   and `SUPPORT.md` support tables and the install snippets; rewrite
   `.github/release-highlights.md` (2 to 4 bullets, no em dashes).
   `scripts/verify-release-versions --release` must pass.
2. **Dry run**: run the Release workflow by hand (Actions, Release,
   "Run workflow", `dry_run` on). It checks the versions, runs the
   tests and the API check, builds the artifacts and prints the release
   notes, and publishes nothing.
3. **Tag** the merged release commit on main with a signed annotated
   tag whose message is exactly `pain001-sdk-go v<version>`:
   `git tag -s -m "pain001-sdk-go v0.0.88" v0.0.88 && git push origin v0.0.88`.
4. **The workflow** verifies the tag (annotated, signed by a key in
   `KEYS.asc`, message, every version restatement), builds the source
   archive, the CycloneDX SBOM and `SHA256SUMS`, attests their build
   provenance, creates the GitHub release titled `pain001-sdk-go
   <version>`, audits the live release, and publishes the module by
   asking proxy.golang.org for the tag and checking sum.golang.org.
   Re-running it on the tag is safe: the release is edited and its
   assets replaced.

The API-breakage check (`scripts/api-check.sh`, gorelease) compares
HEAD with the last tag before this version. An intentional break must
be called out in the CHANGELOG's top section with a `Breaking:` line.
Until the first tag it is a no-op that says so.

### Settings the maintainer creates once

Nothing below is created by CI or by an agent.

| Setting | Where | Why |
| :--- | :--- | :--- |
| Environment `release`, with the maintainer as required reviewer | Settings, Environments | Gates the `release` and `publish` jobs of the Release workflow |
| Pages source "GitHub Actions" | Settings, Pages | Lets `docs.yml` deploy the manual (build_type `workflow`) |
| Repository variable `PAGES_ENABLED` = `true` | Settings, Secrets and variables, Actions, Variables | Turns on the deploy job of `docs.yml`; until then it is skipped |
| Required status checks for the new jobs (optional) | Settings, Branches, `main` | `API-breakage check (gorelease)`, `REUSE lint`, `manual (mkdocs --strict)` are new and not yet required |

No secret is needed: Go has no registry account or token, the release
uses the workflow's `GITHUB_TOKEN`, and attestations use its OIDC
token.
