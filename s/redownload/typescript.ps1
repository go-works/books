#!/usr/bin/env pwsh
Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"
function exitIfFailed { if ($LASTEXITCODE -ne 0) { exit } }

Remove-Item -Force -ErrorAction SilentlyContinue ./gen.exe

go build -o ./gen.exe
exitIfFailed

./gen.exe -redownload-book 9f3a0df9855747b1ab85b76637971d62
Remove-Item -Force -ErrorAction SilentlyContinue ./gen.exe

