<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->

# Policies

The rules this module keeps about versions, toolchains, dependencies
and supported servers, and where each is enforced.

## Versioning

- Versions are `0.0.x` and move by `0.0.1`, in lockstep with the
  pain001 suite; `0.1.0` follows `0.0.999`. The maintainer opens each
  release; nothing bumps a version as a side effect.
- `version.go` is the source of truth. `scripts/verify-release-versions`
  fails CI when `CITATION.cff`, the `SECURITY.md` and `SUPPORT.md`
  support tables, an install snippet or the CHANGELOG heading disagree
  with it.
- The breaking axis is the exported Go API (functions, methods, types,
  struct fields, options, constants) and the requests the client sends.
  While pre-1.0, a release may break either, and the CHANGELOG calls
  the break out with a `Breaking:` line; `scripts/api-check.sh`
  (gorelease) fails CI on an incompatible change without one.

## Minimum Go toolchain

- **The floor is the older of the two Go releases the Go team
  supports**: Go 1.26 today, in the `go` directive of `go.mod`.
- **It moves only when that release leaves support**, which happens
  when a new major Go release ships (Go 1.28 retires 1.26). The move is
  its own change in the next suite release, with a CHANGELOG line.
- **CI enforces it**: the test job runs on both supported releases
  (1.26 and 1.27), and the module uses nothing newer than the floor.
- No claim is made about distribution-packaged Go versions; a
  distribution whose Go is older than the floor can build with a
  toolchain from go.dev.

## Dependencies

- Runtime: the Go standard library only (ADR 0002). `go.mod` has no
  requirements and the module has no `go.sum`.
- Tools run through `go run <module>@<version>` at pinned versions, and
  Python tools install from hash-pinned lockfiles in `ci/`. CI actions
  are pinned by commit SHA.

## Supported servers

- CI tests against the pain001 server pinned in
  `ci/server-requirements.in`: the latest released core, 0.0.87 today.
- The vendored conformance kit is byte-identical to core's and pinned
  by `conformance/SHA256SUMS`.
- The client works with pain001 0.0.85 or newer. Before 0.0.87 a
  refused `Generate` arrives as `200` with `success: false`, which the
  client returns as `ErrNoDocument` rather than `*RefusalError`.

## Security fixes

Only the latest `0.0.x` receives fixes
([SECURITY.md](https://github.com/sebastienrousseau/pain001-sdk-go/blob/main/SECURITY.md)).
