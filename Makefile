# Copyright (C) 2018 Betalo AB - All Rights Reserved

PROJECT_NAME := FORWARDING PROXY

.PHONY: help
help:
	@echo "------------------------------------------------------------------------"
	@echo "${PROJECT_NAME}"
	@echo "------------------------------------------------------------------------"
	@grep -E '^[a-zA-Z0-9_/%\-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build: ## build application binaries
	GOOS=darwin GOARCH=amd64 go build -o forwardingproxy-darwin-amd64 .
	GOOS=linux GOARCH=amd64 go build -o forwardingproxy-linux-amd64 .

.PHONY: image
image: ## build docker image
	docker build -t betalo/forwardingproxy .

.PHONY: lint
lint: ## check code for lint errors
	go vet ./...

.PHONY: test
test: ## run unit tests
	go test -race ./...
