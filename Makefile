# SPDX-License-Identifier: Apache-2.0 OR MIT
#
# `make check` runs every local gate CI runs. The analysers are fetched
# as Go modules at pinned versions by `go run`; nothing is installed
# globally. Set PAIN001_BASE_URL to a running `pain001 serve` to add the
# conformance and typed tests and the example (they skip otherwise).

STATICCHECK := honnef.co/go/tools/cmd/staticcheck@v0.8.2
GOCYCLO     := github.com/fzipp/gocyclo/cmd/gocyclo@v0.6.0
GOCOGNIT    := github.com/uudashr/gocognit/cmd/gocognit@v1.3.0
GOVULNCHECK := golang.org/x/vuln/cmd/govulncheck@v1.9.0

.PHONY: check versions fmt-check fmt vet staticcheck complexity vulncheck build test conformance examples bench regression docs-lint fuzz demo

check: versions fmt-check vet staticcheck complexity vulncheck build test conformance

# Every file that restates the version agrees with version.go, and the
# server pin agrees with its lock (scripts/verify-release-versions).
versions:
	scripts/verify-release-versions

fmt-check:
	@out="$$(gofmt -l .)"; if [ -n "$$out" ]; then echo "gofmt needed:"; echo "$$out"; exit 1; fi

fmt:
	gofmt -w .

vet:
	go vet ./...

staticcheck:
	go run $(STATICCHECK) ./...

# Known vulnerabilities reachable from this module. The client has no
# third-party dependencies (no go.sum), so this covers the Go standard
# library and toolchain in use; it needs network access to vuln.go.dev.
# VULNDB points it at a local copy of the database (file:///path), for a
# machine where Go cannot reach vuln.go.dev directly.
VULNDB ?=

vulncheck:
	go run $(GOVULNCHECK) $(if $(VULNDB),-db $(VULNDB)) ./...

# Cyclomatic <= 10 and cognitive <= 15 per function; size_test.go holds
# 60 lines per function and 500 per file, halstead_test.go a Halstead
# difficulty of 30.
complexity:
	go run $(GOCYCLO) -over 10 .
	go run $(GOCOGNIT) -over 15 .

build:
	go build ./...

# Race-enabled tests of the library; fails below 100.0% of statements
# (scripts/coverage.sh). size_test.go and halstead_test.go hold the size
# and Halstead ceilings.
test:
	scripts/coverage.sh

conformance:
	@if [ -z "$$PAIN001_BASE_URL" ]; then \
		echo "conformance: PAIN001_BASE_URL is unset, skipping the live tests"; \
	else \
		go test -race -count=1 -v -run 'TestConformance|TestLive' . && $(MAKE) examples; \
	fi

# Benchmarks every public method (bench_test.go): BenchmarkClient, the
# client's own cost against recorded responses with no network, and
# BenchmarkLive, a loopback round trip to PAIN001_BASE_URL (skipped when
# unset). The output, also written to bin/bench.txt, is benchstat input.
# CI smoke-runs it with BENCHCOUNT=1 BENCHTIME=1x.
BENCHCOUNT ?= 6
BENCHTIME ?= 1s

bench:
	mkdir -p bin
	go test -run '^$$' -bench . -benchmem -count $(BENCHCOUNT) -benchtime $(BENCHTIME) . | tee bin/bench.txt

# The full regression run CI's conformance job makes: installs the
# pinned server into a temporary venv (unless PAIN001_SERVER is set),
# starts it on a free loopback port with an API key, runs `make check`
# and the benchmark smoke against it, and stops it even on failure.
regression:
	scripts/regression.sh

# The Docs lint workflow: markdownlint, codespell and REUSE. Needs npx,
# and codespell and reuse from ci/lint-requirements.txt on PATH.
docs-lint:
	npx --yes markdownlint-cli2@0.23.3 "**/*.md"
	codespell
	reuse lint

# Fuzzes each Fuzz target the package lists for FUZZTIME (go test runs
# only the seeds and testdata/fuzz). A crash leaves its input in
# testdata/fuzz; commit it as a regression with the fix.
FUZZTIME ?= 60s

fuzz:
	@set -e; targets="$$(go test -list '^Fuzz' . | grep '^Fuzz')"; \
	test -n "$$targets"; \
	for target in $$targets; do \
		echo "fuzz: $$target for $(FUZZTIME)"; \
		go test -run='^$$' -fuzz="^$$target$$" -fuzztime=$(FUZZTIME) . ; \
	done

# Builds every program under examples/ into bin/ (which .gitignore
# excludes) and runs each against PAIN001_BASE_URL from the repository
# root, so the kit's fixture paths resolve. Each exits non-zero when the
# server does not answer as the example expects.
EXAMPLES := $(notdir $(patsubst %/,%,$(dir $(wildcard examples/*/main.go))))

examples:
	mkdir -p bin
	go build -o bin/ ./examples/...
	@set -e; for name in $(EXAMPLES); do \
		echo "example: $$name"; \
		if [ "$$name" = basic ]; then bin/basic -out bin/payments.xml; else bin/$$name; fi; \
	done

# The README demo, .github/demo.gif, rendered by vhs from .github/demo.tape.
# The tape starts `pain001 serve` on 127.0.0.1:$(DEMO_PORT) and stops it at
# the end; make stops it too if vhs fails. Set PAIN001_SERVER to the
# server's `pain001` binary, or put it on PATH; CONTRIBUTING.md shows how
# to install the pinned one.
PAIN001_SERVER ?= $(shell command -v pain001 2>/dev/null)
DEMO_PORT ?= 18765

demo:
	@if [ -z "$(PAIN001_SERVER)" ] || [ ! -x "$(abspath $(PAIN001_SERVER))" ]; then \
		echo "demo: no pain001 server; set PAIN001_SERVER=/path/to/bin/pain001 (see CONTRIBUTING.md)" >&2; exit 1; \
	fi
	@if curl -s -o /dev/null "http://127.0.0.1:$(DEMO_PORT)/"; then \
		echo "demo: 127.0.0.1:$(DEMO_PORT) is already in use; set DEMO_PORT to a free port" >&2; exit 1; \
	fi
	@server="$(abspath $(PAIN001_SERVER))"; \
	PAIN001_SERVER="$$server" DEMO_PORT="$(DEMO_PORT)" vhs .github/demo.tape; status=$$?; \
	pkill -f "$$server serve --host 127.0.0.1 --port $(DEMO_PORT)" || true; \
	exit $$status
