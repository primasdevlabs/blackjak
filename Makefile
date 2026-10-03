# Prefer GNU make when installed. On Windows without make, use:
#   .\make build
# which runs make.cmd → make.ps1 (same targets).

ifeq ($(OS),Windows_NT)
  EXE := .exe
  COPY_AGENT := copy /Y bin\agent.exe extension\bin\agent-win32-x64.exe >nul
  RMDIR = if exist $(1) rmdir /s /q $(1)
  MKDIR_BIN = if not exist bin mkdir bin & if not exist extension\bin mkdir extension\bin
  RUN_AGENT = bin\agent.exe --server --port 47811 --workspace .
else
  EXE :=
  COPY_AGENT := cp -f bin/agent$(EXE) extension/bin/agent-win32-x64.exe 2>/dev/null || cp -f bin/agent extension/bin/agent || true
  RMDIR = rm -rf $(1)
  MKDIR_BIN = mkdir -p bin extension/bin
  RUN_AGENT = ./bin/agent --server --port 47811 --workspace .
endif

.PHONY: all build build-backend build-ui build-extension test test-backend test-ui \
	run ui extension dev clean desktop-dev desktop-build vsix help

all: build

help:
	@echo "Targets: build build-backend build-ui build-extension test run ui desktop-dev desktop-build vsix clean"
	@echo "Windows without GNU make:  .\\make build"

build: build-backend build-ui build-extension

build-backend:
	@echo "Building Go Agent binary..."
	@$(MKDIR_BIN)
	go build -o bin/agent$(EXE) ./cmd/agent
	@$(COPY_AGENT)

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
	@echo "Running UI smoke tests..."
	cd extension/webview/react && npm test --if-present

run: build-backend
	@echo "Starting Go Agent Server on port 47811..."
	$(RUN_AGENT)

ui:
	@echo "Starting React UI development server..."
	cd extension/webview/react && npm run dev

extension: build-extension

dev:
	@echo "Starting Go Agent Server..."
	go run cmd/agent/main.go --server --port 47811

desktop-dev: build-backend build-ui
	@echo "Starting Tauri desktop (dev) — requires Rust + tauri-cli..."
	cd desktop && npm run dev

desktop-build: build-backend build-ui
	@echo "Building Tauri desktop bundle — requires Rust + tauri-cli..."
	cd desktop && npm run build

clean:
	@echo "Cleaning build artifacts..."
ifeq ($(OS),Windows_NT)
	-if exist bin rmdir /s /q bin
	-if exist extension\webview\react\dist rmdir /s /q extension\webview\react\dist
	-if exist extension\out rmdir /s /q extension\out
else
	rm -rf bin/ extension/webview/react/dist extension/out
endif

vsix: build
	@echo "Packaging VSIX..."
	cd extension && npx @vscode/vsce package --allow-missing-repository
