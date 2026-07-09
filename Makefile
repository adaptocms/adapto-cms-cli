APP_NAME := adapto
VERSION ?= dev-$(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"

.PHONY: build test check-docs lint generate docs clean install release

build:
	go build $(LDFLAGS) -o $(APP_NAME) .

install:
	go install $(LDFLAGS) .

test:
	go build ./...
	go vet ./...
	go test -race ./...

check-docs:
	go run ./tools/gendocs -check

lint:
	golangci-lint run ./...

generate:
	./scripts/generate.sh

docs:
	go run ./tools/gendocs

clean:
	rm -f $(APP_NAME)

# Release: bump version, tag, and push to trigger GitHub Actions release workflow.
# Usage:
#   make release              # patch bump (0.1.0 -> 0.1.1)
#   make release BUMP=minor   # minor bump (0.1.1 -> 0.2.0)
#   make release BUMP=major   # major bump (0.2.0 -> 1.0.0)
BUMP ?= patch
release:
	@LATEST=$$(git tag --sort=-v:refname | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$$' | head -1); \
	if [ -z "$$LATEST" ]; then \
		LATEST="v0.0.0"; \
	fi; \
	MAJOR=$$(echo $$LATEST | sed 's/^v//' | cut -d. -f1); \
	MINOR=$$(echo $$LATEST | sed 's/^v//' | cut -d. -f2); \
	PATCH=$$(echo $$LATEST | sed 's/^v//' | cut -d. -f3); \
	case "$(BUMP)" in \
		major) MAJOR=$$((MAJOR + 1)); MINOR=0; PATCH=0 ;; \
		minor) MINOR=$$((MINOR + 1)); PATCH=0 ;; \
		patch) PATCH=$$((PATCH + 1)) ;; \
		*) echo "Invalid BUMP value: $(BUMP). Use major, minor, or patch."; exit 1 ;; \
	esac; \
	NEXT="v$$MAJOR.$$MINOR.$$PATCH"; \
	echo "Releasing $$NEXT (previous: $$LATEST)"; \
	git tag "$$NEXT" && \
	git push origin "$$NEXT" && \
	echo "Tag $$NEXT pushed. GitHub Actions will build and publish the release."
