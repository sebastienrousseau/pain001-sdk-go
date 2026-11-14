#!/bin/sh
# SPDX-License-Identifier: Apache-2.0 OR MIT
#
# Records the responses the client-side benchmarks (bench_test.go)
# decode, from a running `pain001 serve` and the conformance kit's
# fixtures, into testdata/bench/. Re-run it when the server pin moves:
#
#   PAIN001_BASE_URL=http://127.0.0.1:8000 scripts/record-bench-responses.sh
#
# Needs curl and python3 (to JSON-encode the fixtures).
set -eu

cd "$(dirname "$0")/.."

base="${PAIN001_BASE_URL:?set PAIN001_BASE_URL to a running pain001 serve}"
out=testdata/bench
mkdir -p "$out"
fixtures=conformance/fixtures

# body <field> <fixture> [<extra JSON members>] prints a JSON request body
# whose <field> is the fixture's text.
body() {
  extra="${3:-}"
  [ -n "$extra" ] || extra='{}'
  python3 -I -c 'import json, sys
doc = {sys.argv[1]: open(sys.argv[2], encoding="utf-8").read()}
doc.update(json.loads(sys.argv[3]))
print(json.dumps(doc))' "$1" "$2" "$extra"
}

# record <name> <method> <path> [<body>] saves one response body.
record() {
  if [ -n "${4:-}" ]; then
    printf '%s' "$4" | curl -fsS -X "$2" -H 'Content-Type: application/json' \
      ${PAIN001_API_KEY:+-H "Authorization: Bearer $PAIN001_API_KEY"} \
      --data-binary @- "$base$3" -o "$out/$1.json"
  else
    curl -fsS -X "$2" ${PAIN001_API_KEY:+-H "Authorization: Bearer $PAIN001_API_KEY"} \
      "$base$3" -o "$out/$1.json"
  fi
  echo "recorded $out/$1.json"
}

csv="$fixtures/credit-transfer.pain.001.001.09.csv"
xml="$fixtures/credit-transfer.pain.001.001.09.xml"
file='{"filename": "payments.csv", "message_type": "pain.001.001.09"}'

record health GET /api/v1/health
record validate POST /api/v1/ui/validate "$(body content "$csv" "$file")"
record generate POST /api/v1/ui/generate "$(body content "$csv" "$file")"
record check POST /api/v1/check "$(body xml "$xml")"
record faces GET /api/v1/faces
record projectface POST /api/v1/faces/berlin_group "$(body xml "$xml")"
