#!/bin/sh
# Runs from the tester's root folder. The Go program is compiled once and
# cached; it is rebuilt only when the source is newer than the binary.
BIN=/tmp/guess_student
SRC=./student/main.go
if [ ! -f "$BIN" ] || [ "$SRC" -nt "$BIN" ]; then
    go build -o "$BIN" "$SRC" || exit 1
fi
exec "$BIN"
