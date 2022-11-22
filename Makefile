SHELL=/usr/bin/env bash

all: install-sqlc gen build
.PHONY: all

unexport GOFLAGS

GOCC?=go

titan-explorer: $(BUILD_DEPS)
	rm -f titan-explorer
	$(GOCC) build $(GOFLAGS) -o titan-explorer .
.PHONY: titan-explorer

install-sqlc:
	go install github.com/kyleconroy/sqlc/cmd/sqlc@latest

gen:
	sqlc generate
.PHONY: gen

build: titan-explorer
.PHONY: build
