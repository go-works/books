#!/usr/bin/env pwsh
Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"
function exitIfFailed { if ($LASTEXITCODE -ne 0) { exit } }

Remove-Item -Force -ErrorAction SilentlyContinue ./gen.exe

go build -o ./gen.exe
exitIfFailed

# https://www.notion.so/essentialbooks/Essential-Bash-77d28932012b489db9a6d0b349cea865
./gen.exe $args  77d28932012b489db9a6d0b349cea865
Remove-Item -Force -ErrorAction SilentlyContinue ./gen.exe

