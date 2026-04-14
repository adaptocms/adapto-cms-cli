APP_NAME := adapto
VERSION ?= dev
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"

.PHONY: build test lint generate clean install

build:
	go build $(LDFLAGS) -o $(APP_NAME) .

install:
	go install $(LDFLAGS) .

test:
	go test ./...

lint:
	golangci-lint run ./...

generate:
	./scripts/generate.sh

clean:
	rm -f $(APP_NAME)
