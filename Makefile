.PHONY: all build test run ui extension dev clean

all: build

build: build-backend build-ui build-extension

build-backend:
	@echo "Building Go Agent binary..."
	go build -o bin/agent.exe ./cmd/agent
	copy bin\agent.exe extension\bin\agent-win32-x64.exe >nul

build-ui:
	@echo "Building React Webview UI..."
	cd extension/webview/react && npm run build

build-extension:
	@echo "Building VS Code Extension..."
	cd extension && npm run build

test: test-backend test-ui

test-backend:
	@echo "Running Go tests..."
	go test ./... -v

test-ui:
	@echo "Typechecking React UI..."
	cd extension/webview/react && npx tsc --noEmit

run:
	@echo "Starting Go Agent Server on port 47811..."
	go build -o bin/agent.exe ./cmd/agent && bin\agent.exe --server --port 47811 --workspace .

ui:
	@echo "Starting React UI development server..."
	cd extension/webview/react && npm run dev

extension: build-extension

dev:
	@echo "Starting Go Agent Server and UI Dev Server..."
	go run cmd/agent/main.go --server --port 47811

clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/ extension/webview/react/dist extension/out

vsix: build
	@echo "Packaging VSIX..."
	cd extension && npx @vscode/vsce package --allow-missing-repository
