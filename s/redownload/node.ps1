#!/usr/bin/env pwsh
Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"
function exitIfFailed { if ($LASTEXITCODE -ne 0) { exit } }

Remove-Item -Force -ErrorAction SilentlyContinue ./gen.exe

go build -o ./gen.exe
exitIfFailed

./gen.exe a44f25f371164e69a70578bd98e71eb1
Remove-Item -Force -ErrorAction SilentlyContinue ./gen.exe

