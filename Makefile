.PHONY: build test test-short lint vet fmt tidy clean release

BINARY := stage-bin
PKG := ./...

build:
	go build -o $(BINARY) ./cmd/stage

test:
	go test -race $(PKG)

# US4 / FR-006: unit suite must pass without a Docker daemon.
test-short:
	go test -short $(PKG)

vet:
	go vet $(PKG)

lint:
	@command -v staticcheck >/dev/null 2>&1 || { echo "staticcheck not installed; run: go install honnef.co/go/tools/cmd/staticcheck@latest"; exit 1; }
	staticcheck $(PKG)

fmt:
	gofmt -s -w .

tidy:
	go mod tidy

clean:
	rm -f $(BINARY) coverage.out

# Release: deferred to follow-up spec; placeholder so the target exists.
release:
	@echo "Release pipeline implemented as part of T053 in a follow-up spec; see specs/003-rewrite-language-choices/tasks.md."
	@exit 1
