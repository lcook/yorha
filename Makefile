.DEFAULT_GOAL := help
.DELETE_ON_ERROR:
.ONESHELL:
.PHONY: build build-thin clean lint help

VER = 0.1.0
HASH != git rev-parse --short HEAD 2>/dev/null
PROGS = yorha yorha-inst

SUFFIX = -full
TAGS_SUFFIX =

GO_MODULE = github.com/lcook/yorha
GO_FLAGS = -v -ldflags "-s -w -X $(GO_MODULE)/internal/version.Build=$(VER)-$(HASH)$(SUFFIX)" \
						-tags exclude_graphdriver_btrfs$(TAGS_SUFFIX)

build-thin: SUFFIX = -thin
build-thin: TAGS_SUFFIX = ,thin
build build-thin:
	for prog in $(PROGS); do
		go build $(GO_FLAGS) -o $$prog cmd/$$prog/main.go && strip -s $$prog
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
