#!/usr/bin/env pwsh

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"
function exitIfFailed { if ($LASTEXITCODE -ne 0) { exit } }

Remove-Item -Force -ErrorAction SilentlyContinue ./gen.exe
Remove-Item -Force -ErrorAction SilentlyContinue ./sotomd.exe

go build -o ./gen.exe
exitIfFailed

# TODO: make this work
# go build cmd/stack-overflow-to-md -o ./sotomd.exe
# exitIfFailed

Remove-Item -Force -ErrorAction SilentlyContinue ./gen.exe
Remove-Item -Force -ErrorAction SilentlyContinue ./sotomd.exe
