# Windows stand-in for GNU make — run:  .\make.ps1 build   or   .\make build
param(
  [Parameter(Position = 0)]
  [string]$Target = "all"
)

$ErrorActionPreference = "Stop"
Set-Location $PSScriptRoot

function Ensure-Dir([string]$Path) {
  if (-not (Test-Path $Path)) { New-Item -ItemType Directory -Path $Path | Out-Null }
}

function Invoke-BuildBackend {
  Write-Host "Building Go Agent binary..."
  Ensure-Dir "bin"
  Ensure-Dir "extension\bin"
  go build -o bin/agent.exe ./cmd/agent
  if ($LASTEXITCODE -ne 0) { throw "go build failed" }
  Copy-Item -Force "bin\agent.exe" "extension\bin\agent-win32-x64.exe"
}

function Invoke-BuildUi {
  Write-Host "Building React Webview UI..."
  Push-Location "extension\webview\react"
  try {
    npm run build
    if ($LASTEXITCODE -ne 0) { throw "UI build failed" }
  } finally { Pop-Location }
}

function Invoke-BuildExtension {
  Write-Host "Building VS Code Extension..."
  Push-Location "extension"
  try {
    npm run build
    if ($LASTEXITCODE -ne 0) { throw "extension build failed" }
  } finally { Pop-Location }
}

function Invoke-TestBackend {
  Write-Host "Running Go tests..."
  go test ./... -v
  if ($LASTEXITCODE -ne 0) { throw "go test failed" }
}

function Invoke-TestUi {
  Write-Host "Typechecking React UI..."
  Push-Location "extension\webview\react"
  try {
    npx tsc --noEmit
    if ($LASTEXITCODE -ne 0) { throw "UI typecheck failed" }
    Write-Host "Running UI smoke tests..."
    npm test --if-present
    if ($LASTEXITCODE -ne 0) { throw "UI tests failed" }
  } finally { Pop-Location }
}

switch ($Target.ToLowerInvariant()) {
  { $_ -in @("all", "build") } {
    Invoke-BuildBackend
    Invoke-BuildUi
    Invoke-BuildExtension
  }
  "build-backend" { Invoke-BuildBackend }
  "build-ui" { Invoke-BuildUi }
  "build-extension" { Invoke-BuildExtension }
  "test" {
    Invoke-TestBackend
    Invoke-TestUi
  }
  "test-backend" { Invoke-TestBackend }
  "test-ui" { Invoke-TestUi }
  "run" {
    Invoke-BuildBackend
    Write-Host "Starting Go Agent Server on port 47811..."
    & ".\bin\agent.exe" --server --port 47811 --workspace .
  }
  "ui" {
    Write-Host "Starting React UI development server..."
    Push-Location "extension\webview\react"
    try { npm run dev } finally { Pop-Location }
  }
  "extension" { Invoke-BuildExtension }
  "dev" {
    Write-Host "Starting Go Agent Server..."
    go run cmd/agent/main.go --server --port 47811
  }
  "desktop-dev" {
    Invoke-BuildBackend
    Invoke-BuildUi
    Write-Host "Starting Tauri desktop (dev)..."
    Push-Location "desktop"
    try { npm run dev } finally { Pop-Location }
  }
  "desktop-build" {
    Invoke-BuildBackend
    Invoke-BuildUi
    Write-Host "Building Tauri desktop bundle..."
    Push-Location "desktop"
    try { npm run build } finally { Pop-Location }
  }
  "clean" {
    Write-Host "Cleaning build artifacts..."
    Remove-Item -Recurse -Force -ErrorAction SilentlyContinue @(
      "bin",
      "extension\webview\react\dist",
      "extension\out"
    )
  }
  "vsix" {
    Invoke-BuildBackend
    Invoke-BuildUi
    Invoke-BuildExtension
    Write-Host "Packaging VSIX..."
    Push-Location "extension"
    try {
      npx --yes @vscode/vsce package --allow-missing-repository
      if ($LASTEXITCODE -ne 0) { throw "vsce package failed" }
    } finally { Pop-Location }
  }
  { $_ -in @("install-extension", "install-ides", "install") } {
    # Build + package, then install into every VS Code-family IDE we can find.
    Invoke-BuildBackend
    Invoke-BuildUi
    Invoke-BuildExtension
    Write-Host "Packaging VSIX..."
    Push-Location "extension"
    try {
      npx --yes @vscode/vsce package --allow-missing-repository
      if ($LASTEXITCODE -ne 0) { throw "vsce package failed" }
    } finally { Pop-Location }

    $vsix = Get-ChildItem "extension\*.vsix" | Sort-Object LastWriteTime -Descending | Select-Object -First 1
    if (-not $vsix) { throw "No .vsix produced under extension/" }
    Write-Host "VSIX: $($vsix.FullName)"

    $installs = @()
    $pf86 = ${env:ProgramFiles(x86)}
    $cursor = Join-Path $env:LOCALAPPDATA "Programs\cursor\resources\app\bin\cursor.cmd"
    $codeCandidates = @(
      (Join-Path $env:LOCALAPPDATA "Programs\Microsoft VS Code\bin\code.cmd"),
      (Join-Path $env:ProgramFiles "Microsoft VS Code\bin\code.cmd")
    )
    if ($pf86) { $codeCandidates += (Join-Path $pf86 "Microsoft VS Code\bin\code.cmd") }
    $code = $codeCandidates | Where-Object { Test-Path $_ } | Select-Object -First 1
    $antigravity = Join-Path $env:LOCALAPPDATA "Programs\Antigravity IDE\bin\antigravity-ide.cmd"
    $windsurf = Join-Path $env:LOCALAPPDATA "Programs\Windsurf\bin\windsurf.cmd"

    if (Test-Path $cursor) {
      Write-Host "Installing into Cursor..."
      & $cursor --install-extension $vsix.FullName --force
      if ($LASTEXITCODE -ne 0) { Write-Warning "Cursor install returned $LASTEXITCODE" }
      $installs += "Cursor"
    }
    if ($code) {
      Write-Host "Installing into VS Code..."
      & $code --install-extension $vsix.FullName --force
      if ($LASTEXITCODE -ne 0) { Write-Warning "VS Code install returned $LASTEXITCODE" }
      $installs += "VS Code"
    } else {
      Write-Host "VS Code CLI not found - skip (Install from VSIX manually if needed)."
    }
    if (Test-Path $antigravity) {
      Write-Host "Installing into Antigravity IDE..."
      $extDir = Join-Path $env:USERPROFILE ".antigravity-ide\extensions"
      Ensure-Dir $extDir
      & $antigravity --extensions-dir $extDir --install-extension $vsix.FullName --force
      if ($LASTEXITCODE -ne 0) { Write-Warning "Antigravity install returned $LASTEXITCODE" }
      $installs += "Antigravity"
    }
    if (Test-Path $windsurf) {
      Write-Host "Installing into Windsurf..."
      & $windsurf --install-extension $vsix.FullName --force
      $installs += "Windsurf"
    }

    if ($installs.Count -eq 0) {
      throw "No IDE CLIs found. Install manually via Extensions > Install from VSIX: $($vsix.Name)"
    }
    Write-Host ""
    Write-Host "Installed into: $($installs -join ', ')"
    Write-Host "Reload each IDE window (Ctrl+Shift+P then Developer: Reload Window)."
  }
  { $_ -in @("help", "-h", "--help", "/?") } {
    @"
BlackJak make targets (Windows):

  .\make build            Build backend + UI + extension
  .\make build-backend    Go agent binary only
  .\make build-ui         React webview only
  .\make build-extension  Extension host only
  .\make test             Go tests + UI typecheck/smoke
  .\make run              Standalone agent on :47811
  .\make ui               Vite UI dev server
  .\make desktop-dev      Tauri desktop (needs Rust)
  .\make desktop-build    Tauri production bundle
  .\make vsix             Package VSIX
  .\make install-extension Build, package, install into Cursor / VS Code / Antigravity
  .\make clean            Remove build artifacts

Tip: GNU make is optional. This script mirrors the Makefile on Windows.
If you prefer GNU make:  winget install ezwinports.make
"@
  }
  default {
    Write-Error "Unknown target '$Target'. Run: .\make help"
    exit 1
  }
}
