GOCACHE := $(CURDIR)/.test-output/go-cache
export GOCACHE

.PHONY: build test check
build:
	go build -o bin/html-ui-to-vue-vapor ./cmd/html-ui-to-vue-vapor

test:
	go test ./...
	bun run test

check:
	go test ./...
	go test -race ./...
	go vet ./...
	bun run test:compiler
	bun run test:types
	bun run test:browser
	sh scripts/check-standalone.sh
