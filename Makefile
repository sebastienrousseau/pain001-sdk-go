# SPDX-License-Identifier: Apache-2.0 OR MIT
#
# `make check` runs every local gate CI runs. The analysers are fetched
# as Go modules at pinned versions by `go run`; nothing is installed
# globally. Set PAIN001_BASE_URL to a running `pain001 serve` to add the
# conformance and typed tests and the example (they skip otherwise).

STATICCHECK := honnef.co/go/tools/cmd/staticcheck@v0.8.2
GOCYCLO     := github.com/fzipp/gocyclo/cmd/gocyclo@v0.6.0
GOCOGNIT    := github.com/uudashr/gocognit/cmd/gocognit@v1.3.0

.PHONY: check fmt-check fmt vet staticcheck complexity build test conformance example

check: fmt-check vet staticcheck complexity build test conformance

fmt-check:
	@out="$$(gofmt -l .)"; if [ -n "$$out" ]; then echo "gofmt needed:"; echo "$$out"; exit 1; fi

fmt:
	gofmt -w .

vet:
	go vet ./...

staticcheck:
	go run $(STATICCHECK) ./...

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

# Writes into bin/, which .gitignore excludes.
example:
	mkdir -p bin
	go run ./examples/basic -out bin/payments.xml
