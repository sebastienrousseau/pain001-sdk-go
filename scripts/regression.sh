#!/usr/bin/env bash
# SPDX-License-Identifier: Apache-2.0 OR MIT
#
# The full regression run, as CI's conformance job runs it:
#
#   1. installs the pinned server (ci/server-requirements.txt, hash
#      checked) into a temporary venv, unless PAIN001_SERVER names a
#      `pain001` binary already;
#   2. starts `pain001 serve` on a free loopback port with a random
#      PAIN001_API_KEY, and waits for /api/v1/health;
#   3. runs `make check` against it: every static gate, the tests at
#      100% coverage, the conformance kit, the live typed tests and
#      every example;
#   4. smoke-runs every benchmark once (`make bench`);
#   5. stops the server and deletes the venv, also on failure.
#
# Idempotent: everything it creates lives in one temporary directory,
# apart from bin/ (gitignored build output). Set REGRESSION_PORT to pick
# the port, and VULNDB to point govulncheck at a local database copy.
#
#   make regression
#   PAIN001_SERVER=/path/to/venv/bin/pain001 make regression
set -euo pipefail

cd "$(dirname "$0")/.."

work="$(mktemp -d "${TMPDIR:-/tmp}/pain001-sdk-go-regression.XXXXXX")"
server_pid=""

cleanup() {
  status=$?
  if [ -n "$server_pid" ] && kill -0 "$server_pid" 2>/dev/null; then
    kill "$server_pid" 2>/dev/null || true
    wait "$server_pid" 2>/dev/null || true
  fi
  if [ "$status" -ne 0 ] && [ -f "$work/server.log" ]; then
    echo "regression: failed; server log follows" >&2
    tail -n 50 "$work/server.log" >&2
  fi
  rm -rf "$work"
  exit "$status"
}
trap cleanup EXIT
trap 'exit 130' INT TERM

install_server() {
  echo "regression: installing the pinned server into $work/venv"
  if command -v uv >/dev/null 2>&1; then
    uv venv --quiet --python 3.12 "$work/venv"
    uv pip install --quiet --python "$work/venv/bin/python" --require-hashes -r ci/server-requirements.txt
  else
    python3 -m venv "$work/venv"
    "$work/venv/bin/python" -m pip install --quiet --require-hashes -r ci/server-requirements.txt
  fi
  PAIN001_SERVER="$work/venv/bin/pain001"
}

free_port() {
  python3 -I -c 'import socket
s = socket.socket()
s.bind(("127.0.0.1", 0))
print(s.getsockname()[1])
s.close()'
}

start_server() {
  local port="$1"
  echo "regression: starting $("$PAIN001_SERVER" --version) on 127.0.0.1:$port"
  PAIN001_API_KEY="$PAIN001_API_KEY" "$PAIN001_SERVER" serve --host 127.0.0.1 --port "$port" \
    >"$work/server.log" 2>&1 &
  server_pid=$!
  for _ in $(seq 1 60); do
    if curl -fsS "http://127.0.0.1:$port/api/v1/health" >/dev/null 2>&1; then
      return 0
    fi
    if ! kill -0 "$server_pid" 2>/dev/null; then
      echo "regression: the server exited during start-up" >&2
      return 1
    fi
    sleep 1
  done
  echo "regression: the server did not answer /api/v1/health within 60s" >&2
  return 1
}

if [ -z "${PAIN001_SERVER:-}" ]; then
  install_server
fi
port="${REGRESSION_PORT:-$(free_port)}"
PAIN001_API_KEY="regression-$(od -An -N12 -tx1 /dev/urandom | tr -d ' \n')"
export PAIN001_API_KEY
start_server "$port"
export PAIN001_BASE_URL="http://127.0.0.1:$port"

echo "regression: make check against $PAIN001_BASE_URL"
make check
echo "regression: benchmark smoke run"
make bench BENCHCOUNT=1 BENCHTIME=1x
echo "regression: passed"
