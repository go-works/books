#!/usr/bin/env pwsh
Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"
function exitIfFailed { if ($LASTEXITCODE -ne 0) { exit } }

Remove-Item -Force -ErrorAction SilentlyContinue ./gen.exe

go build -o ./gen.exe
exitIfFailed

./gen.exe b64d2e8d06e04dbea37e6e7d0e06bb48
Remove-Item -Force -ErrorAction SilentlyContinue ./gen.exe

