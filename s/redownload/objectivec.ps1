#!/usr/bin/env pwsh
Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"
function exitIfFailed { if ($LASTEXITCODE -ne 0) { exit } }

Remove-Item -Force -ErrorAction SilentlyContinue ./gen.exe

go build -o ./gen.exe
exitIfFailed

./gen.exe d16f98c8a3d641cd88a65fb67e6b0081
Remove-Item -Force -ErrorAction SilentlyContinue ./gen.exe

