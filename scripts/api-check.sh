#!/bin/sh
# SPDX-License-Identifier: Apache-2.0 OR MIT
#
# API-breakage check: gorelease compares the exported API of HEAD with
# the last release tag before this version and fails on an incompatible
# change. Versions move in 0.0.1 steps (AGENTS.md), so an intentional
# break ships in a patch release; it passes only when the CHANGELOG's
# top section says so in a line containing "Breaking:". Until the first
# tag exists there is nothing to compare against, so it says so and
# passes.
#
#   scripts/api-check.sh
#
# Needs a clean work tree and the tags (CI checks out with
# fetch-depth: 0). gorelease runs through `go run` at a pinned version.
set -eu

cd "$(dirname "$0")/.."

GORELEASE="golang.org/x/exp/cmd/gorelease@v0.0.0-20261113113407-e3970619a1a4"

version="v$(sed -n 's/^const Version = "\(.*\)"$/\1/p' version.go)"
base="$(git tag --list 'v*' --sort=-version:refname | grep -vx "$version" | head -n 1 || true)"
if [ -z "$base" ]; then
  echo "api-check: no release tag before $version; nothing to compare"
  exit 0
fi
echo "api-check: comparing HEAD with $base"
if report="$(go run "$GORELEASE" -base="$base" 2>&1)"; then
  echo "$report"
  exit 0
fi
echo "$report"
if ! printf '%s\n' "$report" | grep -q '^## incompatible changes'; then
  echo "api-check: gorelease failed (see above)" >&2
  exit 1
fi
# The CHANGELOG's top section: from its first "## [" heading to the next.
if awk '/^## \[/ { n++ } n == 1' CHANGELOG.md | grep -q 'Breaking:'; then
  echo "api-check: incompatible change acknowledged by a 'Breaking:' line in CHANGELOG.md"
  exit 0
fi
echo "api-check: incompatible API change since $base; record it in CHANGELOG.md with a 'Breaking:' line, or restore compatibility" >&2
exit 1
