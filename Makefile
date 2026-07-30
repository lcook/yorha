.DEFAULT_GOAL = help
.DELETE_ON_ERROR:
.ONESHELL:

.PHONY: build build-thin build-images build-images-extra clean help

VER = 0.1.0
HASH := $(shell git rev-parse --short HEAD 2>/dev/null || echo "devel")

PROGS = yorha yorha-inst

SUFFIX =
TAGS_SUFFIX =

IMAGES = base mainline

GO_MODULE = github.com/lcook/yorha
GO_FLAGS = -v \
	-ldflags "-s -w -X $(GO_MODULE)/internal/version.Build=$(VER)-$(HASH)$(SUFFIX)" \
	-tags exclude_graphdriver_btrfs$(TAGS_SUFFIX)

build: SUFFIX = -full
build: TAGS_SUFFIX =
build: $(PROGS)

build-thin: SUFFIX = -thin
build-thin: TAGS_SUFFIX = ,thin
build-thin: $(PROGS)

$(PROGS):
	go build $(GO_FLAGS) -o $@ cmd/$@/main.go && strip -s $@

build-images: yorha
	for image in $(IMAGES); do ./yorha build -c config/image-$$image.yaml; done

build-images-extra: IMAGES += nvidia intel
build-images-extra: build-images

clean:
	rm -f $(PROGS) Containerfile*

help:
	@echo "build              | Build binaries with all features enabled"
	@echo "build-thin         | Build thin-client binaries"
	@echo "build-images       | Build container images (base, mainline)"
	@echo "build-images-extra | Build container images including nvidia and intel"
	@echo "clean              | Remove built binaries and Containerfiles"
	@echo "help               | Show this help message"
