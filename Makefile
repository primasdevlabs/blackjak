.PHONY: all build test run ui extension dev clean

all: build

build: build-backend build-ui build-extension

build-backend:
	@echo "Building Go Agent binary..."
	go build -o bin/agent.exe ./cmd/agent

build-ui:
	@echo "Building React UI..."
	cd ui && npm run build

build-extension:
	@echo "Building VS Code Extension..."
	cd extension/vscode && npm run build

test: test-backend test-ui

test-backend:
	@echo "Running Go tests..."
	go test ./... -v

test-ui:
	@echo "Typechecking React UI..."
	cd ui && npx tsc --noEmit

run:
	@echo "Starting Go Agent Server on port 8080..."
	go run cmd/agent/main.go --server --port 8080

ui:
	@echo "Starting React UI development server..."
	cd ui && npm run dev

extension: build-extension

dev:
	@echo "Starting Go Agent Server and UI Dev Server..."
	go run cmd/agent/main.go --server --port 8080

clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/ ui/dist extension/vscode/out
