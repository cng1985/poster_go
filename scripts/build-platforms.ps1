$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

if (-not (Get-Command npm -ErrorAction SilentlyContinue)) {
    throw "未找到 npm"
}
if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    throw "未找到 go"
}

Push-Location (Join-Path $root "web")
if (Test-Path "package-lock.json") {
    npm ci
} else {
    npm install
}
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
npm run build
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
Pop-Location

$version = "1.0.0"
$commit = "none"
try { $commit = (git rev-parse --short HEAD).Trim() } catch { }
$built = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
$ldflags = "-s -w -X poster.app/studio/internal/app.Version=$version -X poster.app/studio/internal/app.Commit=$commit -X poster.app/studio/internal/app.BuiltAt=$built"

$targets = @(
    @{ GOOS = "windows"; GOARCH = "amd64"; Name = "poster.exe" },
    @{ GOOS = "linux"; GOARCH = "amd64"; Name = "poster" },
    @{ GOOS = "linux"; GOARCH = "arm64"; Name = "poster" }
)

foreach ($target in $targets) {
    $dir = Join-Path $root ("release/" + $target.GOOS + "-" + $target.GOARCH)
    New-Item -ItemType Directory -Force -Path $dir | Out-Null
    $env:GOOS = $target.GOOS
    $env:GOARCH = $target.GOARCH
    $env:CGO_ENABLED = "0"
    & go build -trimpath -ldflags $ldflags -o (Join-Path $dir $target.Name) ./cmd/server
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
    Copy-Item -Recurse -Force (Join-Path $root "configs") (Join-Path $dir "configs")
    Copy-Item -Force (Join-Path $root "scripts/install-linux.sh") (Join-Path $dir "install-linux.sh")
    @"
poster $version
commit $commit
built $built
"@ | Set-Content -Encoding utf8 (Join-Path $dir "VERSION.txt")
}

Write-Host "release complete"
