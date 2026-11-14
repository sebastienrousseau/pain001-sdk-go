#!/bin/sh
# SPDX-License-Identifier: Apache-2.0 OR MIT
#
# Fails unless every HTML page of the built manual carries the author
# link on the copyright line (workspace standard, AGENTS.md 7.1.2):
#   © <year> <a href="https://sebastienrousseau.com/" rel="author">Sebastien Rousseau</a>
#
#   scripts/check-site-footer.sh site
set -eu

dir="${1:-site}"
link='<a href="https://sebastienrousseau.com/" rel="author">Sebastien Rousseau</a>'
pages=0
missing=0
for page in $(find "$dir" -name '*.html' | sort); do
  pages=$((pages + 1))
  if ! grep -Fq "$link" "$page" || ! grep -Eq '© [0-9]{4} <a href="https://sebastienrousseau.com/"' "$page"; then
    echo "check-site-footer: $page lacks the author link" >&2
    missing=$((missing + 1))
  fi
done
if [ "$pages" -eq 0 ]; then
  echo "check-site-footer: no HTML pages under $dir" >&2
  exit 1
fi
if [ "$missing" -gt 0 ]; then
  exit 1
fi
echo "check-site-footer: all $pages pages carry the author link"
