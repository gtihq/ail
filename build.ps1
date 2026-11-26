$ErrorActionPreference = 'Stop'

New-Item -ItemType Directory -Force dist | Out-Null

$targets = @(
  @{ Goos = 'windows'; Goarch = 'amd64'; Name = 'ail-windows-amd64.exe' }
  @{ Goos = 'darwin'; Goarch = 'arm64'; Name = 'ail-darwin-arm64' }
  @{ Goos = 'darwin'; Goarch = 'amd64'; Name = 'ail-darwin-amd64' }
  @{ Goos = 'linux'; Goarch = 'amd64'; Name = 'ail-linux-amd64' }
)

foreach ($target in $targets) {
  $env:GOOS = $target.Goos
  $env:GOARCH = $target.Goarch
  $env:CGO_ENABLED = '0'
  $output = Join-Path 'dist' $target.Name
  go build -o $output ./cmd/ail
  if ($LASTEXITCODE -ne 0) {
    throw "build failed for $($target.Name)"
  }
  Write-Output "built $output"
}

Remove-Item Env:GOOS, Env:GOARCH, Env:CGO_ENABLED -ErrorAction SilentlyContinue
Get-ChildItem dist | Select-Object Name, Length
