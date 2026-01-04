# Project
MODULE := github.com/samarthasthan/snowflake
PKG := ./...

# Go commands
GO := go

.PHONY: help test race bench perf run clean

## Show available commands
help:
	@echo "Available commands:"
	@echo "  make test    - Run all tests"
	@echo "  make race    - Run tests with race detector"
	@echo "  make bench   - Run benchmarks"
	@echo "  make perf    - Run performance tests"
	@echo "  make run     - Run example main"
	@echo "  make clean   - Clean test cache"

## Run all tests
test:
	$(GO) test -v $(PKG)

## Run tests with race detector
race:
	$(GO) test -race -v $(PKG)

## Run benchmarks
bench:
	$(GO) test -bench=. -benchmem $(PKG)

## Run performance tests only
perf:
	$(GO) test -v -run=Performance $(PKG)

## Run example / demo
run:
	$(GO) run cmd/demo/main.go

## Clean test cache
clean:
	$(GO) clean -testcache
