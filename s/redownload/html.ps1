#!/usr/bin/env pwsh
Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"
function exitIfFailed { if ($LASTEXITCODE -ne 0) { exit } }

Remove-Item -Force -ErrorAction SilentlyContinue ./gen.exe

go build -o ./gen.exe
exitIfFailed

./gen.exe -redownload-book 1c13e594ccd5472fb58d4c56379e7540
Remove-Item -Force -ErrorAction SilentlyContinue ./gen.exe

