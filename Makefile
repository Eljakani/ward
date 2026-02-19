VERSION  ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT   ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE     ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo unknown)
MODULE    = github.com/eljakani/ward
LDFLAGS   = -s -w -X $(MODULE)/cmd.Version=$(VERSION) -X $(MODULE)/cmd.Commit=$(COMMIT) -X $(MODULE)/cmd.Date=$(DATE)
BINARY    = ward

.PHONY: build install test lint clean

## build: Compile ward with version info
build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) .

## install: Install ward to $GOPATH/bin
install:
	go install -ldflags "$(LDFLAGS)" .

## test: Run all tests
test:
	go test ./... -v

## lint: Run go vet
lint:
	go vet ./...

## clean: Remove build artifacts
clean:
	$(RM) $(BINARY) $(BINARY).exe
