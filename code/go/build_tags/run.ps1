remove-item -ErrorAction Ignore main.exe
go build -o main.exe .
./main.exe
remove-item -ErrorAction Ignore main.exe

