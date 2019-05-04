#!/usr/bin/env pwsh
Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"
function exitIfFailed { if ($LASTEXITCODE -ne 0) { exit } }

Remove-Item -Force -ErrorAction SilentlyContinue ./gen.exe

go build -o ./gen.exe
exitIfFailed

./gen.exe 6289ba4610874393844dfbae890ed035
Remove-Item -Force -ErrorAction SilentlyContinue ./gen.exe

