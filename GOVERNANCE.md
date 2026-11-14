<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->

# Governance

How pain001-sdk-go is run, who decides, and what happens if its
maintainer becomes unavailable. It follows the
[pain001 suite's governance](https://github.com/sebastienrousseau/pain001/blob/main/GOVERNANCE.md);
this page records what is specific to this repository.

## Scope

A thin, typed Go client for the pain001 REST API, held to core's
conformance kit. Changes are weighed against that scope: correctness,
security and a small, stable API over breadth. Payment logic belongs in
[pain001](https://github.com/sebastienrousseau/pain001), not here.

## Roles and responsibilities

| Role | Who | Responsibilities |
| :--- | :--- | :--- |
| **Maintainer** | Sebastien Rousseau (sole maintainer) | Sets direction, reviews and merges pull requests, opens each suite release, signs tags, approves the `release` environment, triages issues and security reports |
| **Contributor** | Anyone with a merged pull request | Proposes changes with tests, signs off every commit (DCO), keeps CI green |
| **AI-assisted contributor** | Agents working under a human's direction | Follow [`AGENTS.md`](https://github.com/sebastienrousseau/pain001-sdk-go/blob/main/AGENTS.md); never merge, tag or publish; the human signs off and is accountable |
| **User** | Everyone | Files issues and questions ([`SUPPORT.md`](https://github.com/sebastienrousseau/pain001-sdk-go/blob/main/SUPPORT.md)) and reports vulnerabilities privately ([`SECURITY.md`](https://github.com/sebastienrousseau/pain001-sdk-go/blob/main/SECURITY.md)) |

## Decisions

- **Day-to-day changes** (fixes, tests, docs, additive API within
  scope) go through a pull request; the maintainer merges once CI is
  green.
- **Significant changes** (a breaking API change, a new dependency, a
  change to the retry or error model) get an ADR in `docs/adr/` in the
  same pull request.
- **Suite-wide changes** (a new method surface, a conformance-kit
  change) are agreed in core first and land in all five SDKs in the same
  suite release.

## Releases

Versions are `0.0.x`, move by `0.0.1` and follow the suite; the
maintainer opens each release. A release is a signed annotated tag; the
release workflow builds and attests the artifacts, writes the GitHub
release and warms proxy.golang.org, gated by the `release` environment
(see [`DEVELOPMENT.md`](https://github.com/sebastienrousseau/pain001-sdk-go/blob/main/DEVELOPMENT.md#releasing)).

## Continuity plan

The project has one maintainer. What happens to each asset if the
maintainer becomes unavailable; items marked **Not yet arranged** have
no arrangement today.

| Asset | Today | Continuity |
| :--- | :--- | :--- |
| Published versions | Tags on GitHub, cached by proxy.golang.org and recorded in sum.golang.org | The proxy keeps every published version, so builds that depend on one keep working even if the repository disappears. |
| Tag-signing keys | SSH keys in `KEYS.asc` | A new maintainer signs with their own key, added to `KEYS.asc` in a reviewed change; old tags stay verifiable from the file's history. **Not yet arranged:** a documented key backup. |
| Registry accounts | None: Go publishes from the tag, with no account or token | Nothing to hand over. |
| GitHub repository | Owned by the maintainer's account | Transfer to an organisation with at least two owners, keeping redirects. **Not yet arranged:** no second owner is named. |
| Module path | `github.com/sebastienrousseau/pain001-sdk-go` | A transfer keeps working through GitHub's redirect; a fork continues under its own path. |
| CI and release process | Documented in `DEVELOPMENT.md` and run by GitHub Actions | Anyone with write access and a key in `KEYS.asc` can release from the documented process. |
| The code | Apache-2.0 OR MIT, public history | Anyone can fork and continue it. |

## Conduct and security

Participation follows
[`CODE_OF_CONDUCT.md`](https://github.com/sebastienrousseau/pain001-sdk-go/blob/main/CODE_OF_CONDUCT.md).
Vulnerabilities go through the private process in
[`SECURITY.md`](https://github.com/sebastienrousseau/pain001-sdk-go/blob/main/SECURITY.md),
never a public issue.
