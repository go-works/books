Remove-Item -ErrorAction Ignore -Path app.exe,rsrc.syso
rsrc -manifest app.manifest -o rsrc.syso
go build -ldflags="-H windowsgui" -o app.exe .
./app.exe | Out-Null
Remove-Item -ErrorAction Ignore -Path app.exe,rsrc.syso
