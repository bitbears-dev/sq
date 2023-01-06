subdirs = cli cmd/sq

.PHONY: all
all: lint test $(subdirs)

.PHONY: build-for-release
build-for-release: lint test $(subdirs) e2e-test

.PHONY: lint
lint:
	staticcheck ./...

.PHONY: clean test package release
clean test package release: $(subdirs)

$(subdirs): force
	$(MAKE) -C $@ $(MAKECMDGOALS)

.PHONY: force
force:

.PHONY: update-deps
update-deps:
	go get -u ./...
