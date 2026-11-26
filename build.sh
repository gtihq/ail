#!/usr/bin/env bash
set -euo pipefail

mkdir -p dist

build() {
  local goos="$1"
  local goarch="$2"
  local name="$3"
  GOOS="$goos" GOARCH="$goarch" CGO_ENABLED=0 go build -o "dist/$name" ./cmd/ail
  echo "built dist/$name"
}

build windows amd64 ail-windows-amd64.exe
build darwin arm64 ail-darwin-arm64
build darwin amd64 ail-darwin-amd64
build linux amd64 ail-linux-amd64

ls -lh dist
