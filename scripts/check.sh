#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

export GOCACHE="${GOCACHE:-$(pwd)/.cache/go-build}"
export GOPATH="${GOPATH:-$(pwd)/.cache/gopath}"
mkdir -p "$GOCACHE" "$GOPATH"

echo '== gofmt =='
unformatted="$(gofmt -l .)"
if [ -n "$unformatted" ]; then
  echo "$unformatted"
  exit 1
fi

echo '== go vet (host) =='
go vet ./...

echo '== tests =='
go test ./...

echo '== Windows cross-build =='
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build ./...

echo 'All checks passed.'

