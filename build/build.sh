#!/bin/sh
kill $(pidof 41_debug) 2>/dev/null
rm -f 41_debug
go build -o 41_debug cmd/41/main.go
sudo ./41_debug -i en0 -p 80 --protocol http1 