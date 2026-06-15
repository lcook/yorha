.DEFAULT_GOAL := help
.DELETE_ON_ERROR:
.ONESHELL:
.PHONY: build build-thin clean lint help

VER = 0.1.0
HASH != git rev-parse --short HEAD 2>/dev/null
PROGS = yorha yorha-inst
GO_MODULE = github.com/lcook/yorha

build: SUFFIX = -full
build-thin: SUFFIX = -thin
build build-thin:
	for prog in $(PROGS); do
		go build -v -ldflags "-s -w -X $(GO_MODULE)/internal/version.Build=$(VER)-$(HASH)$(SUFFIX)" \
			-tags exclude_graphdriver_btrfs$(if $(filter build-thin,$@),$(comma)thin) \
			-o $$prog cmd/$$prog/main.go && strip -s $$prog
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
