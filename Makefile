.DEFAULT_GOAL := help
.DELETE_ON_ERROR:
.ONESHELL:
.PHONY: build build-thin clean lint run help
VER = 0.1.0
HASH != git rev-parse --short HEAD 2>/dev/null
ifdef HASH
PROG_VERSION := $(VER)-$(HASH)
else
PROG_VERSION = $(VER)
endif
PROGS = yorha yorha-inst

GH_ACCOUNT = lcook
GH_PROJECT = yorha

GO_MODULE = github.com/$(GH_ACCOUNT)/$(GH_PROJECT)
GO_FLAGS = -v -ldflags "-s -w -X $(GO_MODULE)/internal/version.Build=$(PROG_VERSION)"

build:
	for prog in $(PROGS); do
		go build $(GO_FLAGS) -tags exclude_graphdriver_btrfs -o $$prog cmd/$$prog/main.go && strip -s $$prog
	done

build-thin:
	for prog in $(PROGS); do
		go build $(GO_FLAGS) -tags exclude_graphdriver_btrfs,thin -o $$prog cmd/$$prog/main.go && strip -s $$prog
	done

clean:
	@for prog in $(PROGS); do rm -fv $$prog; done
	@rm -fv Containerfile*

lint:
	golangci-lint run

help:
	@echo "Available targets:"
	@echo "  build           - Build binaries with all features enabled"
	@echo "  build-thin      - Build thin-client binaries"
	@echo "  clean           - Remove built binaries and Containerfiles"
	@echo "  lint            - Run linter"
	@echo "  help            - Show this help message"
