#!/usr/bin/env bash
# SPDX-License-Identifier: Apache-2.0 OR MIT
#
# Writes the GitHub release notes for the version in version.go to
# stdout, in the suite's release format:
#
#   ## Highlights ⭐️       from .github/release-highlights.md (by hand)
#   ## What's Changed      the pull requests merged since the last tag
#   ## New Contributors    authors of their first merged pull request
#   ## Checksums           dist/SHA256SUMS
#   **Full Changelog**     compare link (commits link for a first tag)
#
#   GH_TOKEN=... scripts/release-notes.sh > notes.md
#
# Needs gh, the full history with tags, and dist/ from
# scripts/release-artifacts.sh. Fails on an em dash anywhere in the
# notes, or on a Highlights section without 2 to 4 bullets.
set -euo pipefail

cd "$(dirname "$0")/.."

repo="${GITHUB_REPOSITORY:-sebastienrousseau/pain001-sdk-go}"
version="$(sed -n 's/^const Version = "\(.*\)"$/\1/p' version.go)"
tag="v$version"
highlights=.github/release-highlights.md

# The range: from the previous release tag to the tag (HEAD in a dry
# run, when the tag does not exist yet).
head_ref="$tag"
git rev-parse -q --verify "refs/tags/$tag" >/dev/null || head_ref=HEAD
prev="$(git tag --list 'v*' --sort=-version:refname --merged "$head_ref" | grep -vx "$tag" | head -n 1 || true)"
range="$head_ref"
[ -z "$prev" ] || range="$prev..$head_ref"

bullets="$(grep -c '^\* \*\*' "$highlights" || true)"
if [ "$bullets" -lt 2 ] || [ "$bullets" -gt 4 ]; then
  echo "release-notes: $highlights has $bullets bullets, want 2 to 4 '* **Headline**: ...' lines" >&2
  exit 1
fi

# pr_line <number> prints "* <title> by @<author> in <url>".
pr_line() {
  gh pr view "$1" --repo "$repo" --json title,author,url \
    --jq '"* \(.title) by @\(.author.login) in \(.url)"'
}

whats_changed() {
  local prs
  prs="$(git log --merges --format=%s "$range" | sed -n 's/^Merge pull request #\([0-9]*\) .*/\1/p' | sort -n)"
  if [ -n "$prs" ]; then
    for n in $prs; do pr_line "$n"; done
    return
  fi
  # No pull request in the range: list its commits in the same shape.
  for sha in $(git log --reverse --no-merges --format=%H "$range"); do
    gh api "repos/$repo/commits/$sha" \
      --jq '"* \(.commit.message | split("\n")[0]) by @\(.author.login // .commit.author.name) in \(.html_url)"'
  done
}

# new_contributors lists the authors whose first merged pull request is
# in the range. A first release has no earlier work to compare with, so
# the section is left out.
new_contributors() {
  [ -n "$prev" ] || return 0
  local before
  before="$(git log --merges --format=%s "$prev" | sed -n 's/^Merge pull request #\([0-9]*\) .*/\1/p')"
  local seen=" "
  for n in $before; do
    seen="$seen$(gh pr view "$n" --repo "$repo" --json author --jq .author.login) "
  done
  for n in $(git log --merges --format=%s "$range" | sed -n 's/^Merge pull request #\([0-9]*\) .*/\1/p' | sort -n); do
    login="$(gh pr view "$n" --repo "$repo" --json author --jq .author.login)"
    case "$seen" in *" $login "*) continue ;; esac
    seen="$seen$login "
    echo "* @$login made their first contribution in https://github.com/$repo/pull/$n"
  done
}

notes="$(
  echo "## Highlights ⭐️"
  echo
  grep '^\* \*\*' "$highlights"
  echo
  echo "## What's Changed"
  echo
  whats_changed
  contributors="$(new_contributors)"
  if [ -n "$contributors" ]; then
    echo
    echo "## New Contributors"
    echo
    echo "$contributors"
  fi
  echo
  echo "## Checksums"
  echo
  echo "SHA-256 of every asset attached to this release:"
  echo
  echo '```text'
  cat dist/SHA256SUMS
  echo '```'
  echo
  if [ -n "$prev" ]; then
    echo "**Full Changelog**: https://github.com/$repo/compare/$prev...$tag"
  else
    echo "**Full Changelog**: https://github.com/$repo/commits/$tag"
  fi
)"

if printf '%s' "$notes" | grep -q '—'; then
  echo "release-notes: the notes contain an em dash; use a comma, colon or parentheses" >&2
  exit 1
fi
printf '%s\n' "$notes"
