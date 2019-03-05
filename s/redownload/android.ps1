#!/usr/bin/env pwsh
Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"
function exitIfFailed { if ($LASTEXITCODE -ne 0) { exit } }

Remove-Item -Force -ErrorAction SilentlyContinue ./gen.exe

go build -o ./gen.exe ./cmd/gen-books
exitIfFailed

./gen.exe -redownload-book f90b0a6b648343e28dc5ed6e8f5c0780
Remove-Item -Force -ErrorAction SilentlyContinue ./gen.exe

