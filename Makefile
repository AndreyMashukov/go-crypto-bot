.PHONY: fmt vet lint ailint build test check hooks init-db-dev

# All Go targets delegate to server/. The Go module lives in server/go.mod.

fmt:
	$(MAKE) -C server fmt

vet:
	$(MAKE) -C server vet

lint:
	$(MAKE) -C server lint

ailint:
	$(MAKE) -C server ailint

build:
	$(MAKE) -C server build

test:
	$(MAKE) -C server test

check:
	$(MAKE) -C server check

hooks:
	git config core.hooksPath .githooks/
	@echo "pre-commit hook activated"

init-db-dev:
	$(MAKE) -C server init-db-dev
