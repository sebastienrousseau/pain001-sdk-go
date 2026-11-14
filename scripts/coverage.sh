#!/bin/sh
# SPDX-License-Identifier: Apache-2.0 OR MIT
#
# Runs the library's tests with the race detector and fails unless they
# cover 100.0% of statements. The measured packages are the module root
# (the client) and internal/kit (the conformance runner). examples/ are
# programs, not library code: CI builds each one and runs it against a
# live `pain001 serve` instead (see DEVELOPMENT.md).
#
#   scripts/coverage.sh                    # writes coverage.out
#   COVERPROFILE=/tmp/c.out scripts/coverage.sh
set -eu

cd "$(dirname "$0")/.."

profile="${COVERPROFILE:-coverage.out}"
go test -race -count=1 -covermode=atomic -coverprofile="$profile" . ./internal/...

total="$(go tool cover -func="$profile" | awk '/^total:/ { sub("%", "", $3); print $3 }')"
if [ "$total" != "100.0" ]; then
  echo "coverage: $total% of statements, the floor is 100.0%; uncovered:" >&2
  go tool cover -func="$profile" | grep -v '100.0%$' >&2
  exit 1
fi
echo "coverage: 100.0% of statements"
