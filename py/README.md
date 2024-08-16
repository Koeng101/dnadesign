# DnaDesign (Python)
This directory contains code for allowing python users to use dnadesign through a shared C library.

This is a work-in-progress. Right now, we have only ported the fasta parser. 

### Other platforms
If you have interest in other platforms, like openbsd or freebsd, please add an issue! I'd be happy to add automatic packaging for these alternative platforms if I know someone will use them.

### Testing
```
CC="zig cc -target x86_64-linux-gnu" CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o dnadesign/libdnadesign.so -buildmode=c-shared lib.go
```
