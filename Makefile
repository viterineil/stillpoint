BINARY := bin/stillpoint
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w \
	-X github.com/viterineil/stillpoint/internal/version.Version=$(VERSION) \
	-X github.com/viterineil/stillpoint/internal/version.Commit=$(COMMIT) \
	-X github.com/viterineil/stillpoint/internal/version.Date=$(BUILD_DATE)

.PHONY: all build test vet fmt fmt-check check clean

all: check build

build:
	mkdir -p bin
	CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/stillpoint

test:
	go test -race ./...

vet:
	go vet ./...

fmt:
	gofmt -w $$(find cmd internal -name '*.go' -type f)

fmt-check:
	@test -z "$$(gofmt -l $$(find cmd internal -name '*.go' -type f))" || \
		(echo "Go files need formatting; run 'make fmt'" && exit 1)

check: fmt-check vet test

clean:
	rm -f $(BINARY) coverage.out
