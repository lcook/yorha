.PHONY: build build-thin build-images build-images-extra gen clean help
.DELETE_ON_ERROR:
.ONESHELL:

.DEFAULT_GOAL = help

include Makefile.qemu

VER = 0.1.0
HASH := $(shell git rev-parse --short HEAD 2>/dev/null || echo "devel")

PROGS = yorha yorha-inst
IMAGES = base mainline

SUFFIX = -full

GO_MODULE = github.com/lcook/yorha
GO_FLAGS = -v \
	-ldflags "-s -w -X $(GO_MODULE)/internal/version.Build=$(VER)-$(HASH)$(SUFFIX)" \
	-tags exclude_graphdriver_btrfs$(TAGS_SUFFIX)

build: $(PROGS)

build-thin: SUFFIX = -thin
build-thin: TAGS_SUFFIX = ,thin
build-thin: $(PROGS)

define foreach-image
	for image in $(IMAGES); do \
		./yorha $(1) -c config/image-$$image.yaml $(2); \
	done
endef

$(PROGS):
	go build $(GO_FLAGS) -o $@ cmd/$@/main.go

build-images: yorha
	$(call foreach-image,build)

build-images-extra: IMAGES += nvidia intel
build-images-extra: build-images

gen: IMAGES += nvidia intel
gen: yorha
	$(call foreach-image,gen,-o Containerfile.$$image)

clean:
	rm -rfv $(PROGS) Containerfile* $(ARCHISO_OUT) $(ARCHISO_TMP)

help:
	@echo "build              | Build binaries with all features enabled"
	@echo "build-thin         | Build thin-client binaries"
	@echo "build-images       | Build container images (base, mainline)"
	@echo "build-images-extra | Build container images including nvidia and intel"
	@echo "gen                | Generate Containerfiles for all image types"
	@echo "qemu-installer     | Build bootable installer ISO"
	@echo "qemu-installer-run | Create disk and launch QEMU with installer ISO"
	@echo "clean              | Remove build artifacts and generated files"
	@echo "help               | Show this help message"