remove-item -ErrorAction Ignore main_alt.exe
go build -tags alt -o main_alt.exe .
./main_alt.exe
remove-item -ErrorAction Ignore main_alt.exe
