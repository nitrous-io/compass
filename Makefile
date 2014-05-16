image:
	@docker build -t registry.nitrous.io/compass .
.PHONY: image

shell: image
	@./script/shell
.PHONY: shell

setup:
	@rm -rf _vendor
	@./script/setup
.PHONY: setup

run:
	@./script/install
	@compass
.PHONY: setup

build: image
	@./script/build
.PHONY: build
