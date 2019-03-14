#!/usr/bin/env pwsh
Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"
function exitIfFailed { if ($LASTEXITCODE -ne 0) { exit } }

Remove-Item -Force -ErrorAction SilentlyContinue ./gen.exe

go build -o ./gen.exe
exitIfFailed

# https://www.notion.so/kjkpublic/Essential-MySQL-4489ab73989f4ae9912486561e165deb
./gen.exe -redownload-book 4489ab73989f4ae9912486561e165deb
Remove-Item -Force -ErrorAction SilentlyContinue ./gen.exe
