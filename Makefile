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

.PHONY: check fmt-check fmt vet staticcheck complexity vulncheck build test conformance example fuzz

check: fmt-check vet staticcheck complexity vulncheck build test conformance

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
vulncheck:
	go run $(GOVULNCHECK) ./...

# Cyclomatic <= 10 and cognitive <= 15 per function; size_test.go holds
# 60 lines per function and 500 per file.
complexity:
	go run $(GOCYCLO) -over 10 .
	go run $(GOCOGNIT) -over 15 .

build:
	go build ./...

test:
	go test -race -count=1 -cover ./...

conformance:
	@if [ -z "$$PAIN001_BASE_URL" ]; then \
		echo "conformance: PAIN001_BASE_URL is unset, skipping the live tests"; \
	else \
		go test -race -count=1 -v -run 'TestConformance|TestLive' . && $(MAKE) example; \
	fi

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

# Writes into bin/, which .gitignore excludes.
example:
	mkdir -p bin
	go run ./examples/basic -out bin/payments.xml
