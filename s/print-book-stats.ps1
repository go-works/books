#!/usr/bin/env pwsh
Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"
function exitIfFailed { if ($LASTEXITCODE -ne 0) { exit } }

Remove-Item -Force -ErrorAction SilentlyContinue ./so-to-md.exe
Remove-Item -Force -ErrorAction SilentlyContinue ./so-to-md

go build -o so-to-md.exe ./cmd/stack-overflow-to-md
exitIfFailed

./so-to-md -stats

Remove-Item -Force -ErrorAction SilentlyContinue ./so-to-md.exe
Remove-Item -Force -ErrorAction SilentlyContinue ./so-to-md
