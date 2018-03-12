# Copyright (C) 2018 Betalo AB - All Rights Reserved

.PHONY: all
all: build

.PHONY: build
build:
	GOOS=darwin GOARCH=amd64 go build -o forwardingproxy-darwin-amd64 .
	GOOS=linux GOARCH=amd64 go build -o forwardingproxy-linux-amd64 .
