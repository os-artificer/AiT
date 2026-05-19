.PHONY: all build test vet fmt clean

BINARY_NAME := ait
BUILD_DIR   := bin
MAIN_PKG    := ./cmd/ait

all: build

build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PKG)

test:
	go test ./...

vet:
	go vet ./...

fmt:
	go fmt ./...

clean:
	rm -rf $(BUILD_DIR)
