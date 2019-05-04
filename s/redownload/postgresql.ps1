#!/usr/bin/env pwsh
Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"
function exitIfFailed { if ($LASTEXITCODE -ne 0) { exit } }

Remove-Item -Force -ErrorAction SilentlyContinue ./gen.exe

go build -o ./gen.exe
exitIfFailed

./gen.exe 799304340f2c4081b6c4b7eb28df368e
Remove-Item -Force -ErrorAction SilentlyContinue ./gen.exe
