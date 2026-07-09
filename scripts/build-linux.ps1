$ErrorActionPreference = "Stop"

$root = Resolve-Path (Join-Path $PSScriptRoot "..")
$outDir = Join-Path $root "bin"
$outFile = Join-Path $outDir "blog"

$oldGoos = $env:GOOS
$oldGoarch = $env:GOARCH

try {
    New-Item -ItemType Directory -Force -Path $outDir | Out-Null

    $env:GOOS = "linux"
    $env:GOARCH = "amd64"

    Push-Location $root
    try {
        go build -tags admin -o $outFile -buildvcs=false ./cmd/blog
    } finally {
        Pop-Location
    }

    Write-Host "Built Linux binary: $outFile"
} finally {
    if ($null -eq $oldGoos) {
        Remove-Item Env:GOOS -ErrorAction SilentlyContinue
    } else {
        $env:GOOS = $oldGoos
    }

    if ($null -eq $oldGoarch) {
        Remove-Item Env:GOARCH -ErrorAction SilentlyContinue
    } else {
        $env:GOARCH = $oldGoarch
    }
}
