.PHONY: all build test vet fmt clean test-cross dupl

BINARY_NAME := ait
BUILD_DIR   := bin
MAIN_PKG    := ./cmd/ait

all: build

build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PKG)

test:
	go test ./...

test-cross:
	GOOS=windows GOARCH=amd64 go build ./...
	GOOS=darwin GOARCH=amd64 go build ./...

vet:
	go vet ./...

fmt:
	go fmt ./...

dupl:
	go run github.com/mibk/dupl@latest -threshold 15 $$(find . -name '*.go' ! -name '*_test.go' ! -path './vendor/*')

clean:
	rm -rf $(BUILD_DIR)
