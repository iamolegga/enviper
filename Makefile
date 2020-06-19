PROJECTNAME=$(shell basename "$(PWD)")
MAKEFLAGS += --silent

all: help
.PHONY: all

## install: Install deps
install:
	@echo "	Installing deps..."
	@go mod download
.PHONY: install

## test: Run tests
test:
	@echo "	Running tests..."
	@go test ./...
.PHONY: test

## test-watch: Run tests in watch mode (rerun on change)
test-watch:
	@echo " Running tests in watch mode..."
	go test ./...
	watchexec -e go -r "go test ./..."
.PHONY: test-watch

## coverage: Run tests with coverage
coverage:
	@echo "	Running tests with coverage..."
	@go test -v -run=Test ./... -coverprofile c.out
.PHONY: coverage

## coverage-watch: Run tests with coverage in watch mode (rerun on change)
coverage-watch:
	@echo "	Running tests with coverage in watch mode..."
	watchexec -e go -r "go test -v -run=Test ./... -coverprofile c.out"
.PHONY: coverage-watch

## coverage-results: Open coverage results
coverage-results:
	@go tool cover -html=c.out
.PHONY: coverage-results

## doc: Run godoc in watch mode (rerun on change)
doc:
	@echo " Running godoc in watch mode..."
	watchexec -e go -r "godoc -http=:8080"
.PHONY: doc

.PHONY: help
help: Makefile
	@echo
	@echo "	Choose a command run in "$(PROJECTNAME)":"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo
