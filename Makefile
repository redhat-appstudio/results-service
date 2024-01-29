SHELL := /bin/bash

default: build

fmt: ## Run go fmt against code.
	go fmt ./cmd/... ./pkg/...

vet: ## Run go vet against code.
	go vet ./cmd/... ./pkg/...

test: fmt vet ## Run tests.
	go test -v ./pkg/... -coverprofile cover.out

build:
	env GOOS=linux GOARCH=amd64 go build -mod=vendor -o out/results-service ./cmd/service

clean:
	rm -rf out
