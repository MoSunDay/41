#!/bin/sh
kill $(pidof 41_debug) 2>/dev/null
rm -f 41_debug
go mod tidy
go mod download
go build -o 41_debug cmd/41/main.go
sudo ./41_debug -l 50000 -i lo -p 8001 --protocol http1
