#!/usr/bin/env pwsh
Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"
function exitIfFailed { if ($LASTEXITCODE -ne 0) { exit } }

Remove-Item -Force -ErrorAction SilentlyContinue ./gen.exe

go build -o ./gen.exe
exitIfFailed

# https://www.notion.so/essentialbooks/Essential-Batch-ea84bde7ed4e4353bdc6ae44125abc08
./gen.exe $args  ea84bde7ed4e4353bdc6ae44125abc08
Remove-Item -Force -ErrorAction SilentlyContinue ./gen.exe

