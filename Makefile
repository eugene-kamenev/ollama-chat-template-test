# Makefile for building multiple Go executables

.PHONY: all native wasip1 wasm-js clean

# Build all targets
all: test native wasip1 wasm-js

# Build the native executable
native:
	@echo "Building native executable..."
	go build -o bin/ollama-template-test-native ./cmd/native
	@echo "Done"

# Build the WASI/wasip1 executable
wasip1:
	@echo "Building WASI/wasip1 executable..."
	GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared -o bin/ollama-template-test-wasip1.wasm ./cmd/wasip1
	@echo "Done"

# Build the WebAssembly module for JavaScript
wasm-js:
	@echo "Building WebAssembly (JS) module..."
	GOOS=js GOARCH=wasm go build -o bin/ollama-template-test-wasm-js.wasm ./cmd/wasm-js
	cp ./bin/ollama-template-test-wasm-js.wasm ./public_html/ollama-template-test-wasm-js.wasm
	@echo "Done"

test:
	@echo "Starting test suite"
	go list ./... | grep -v "cmd/wasm-js" | xargs go test
	@echo "Done"	

# Clean up built files
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/*
	rm -f ./public_html/ollama-template-test-wasm-js.wasm
