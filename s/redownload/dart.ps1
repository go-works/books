#!/usr/bin/env pwsh
Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"
function exitIfFailed { if ($LASTEXITCODE -ne 0) { exit } }

Remove-Item -Force -ErrorAction SilentlyContinue ./gen.exe

go build -o ./gen.exe
exitIfFailed

./gen.exe -redownload-book 0e2d248bf94b4aebaefbcf51ae435df0
Remove-Item -Force -ErrorAction SilentlyContinue ./gen.exe
