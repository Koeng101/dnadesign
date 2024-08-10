To deploy your package:
1. Run `make` to build the shared libraries for all supported platforms.
2. Run `python3 setup.py sdist bdist_wheel` to create the distribution files.
3. Use `twine` to upload the distribution files to PyPI.

Simple compilation: `go build -o dnadesign/libdnadesign.so -buildmode=c-shared lib.go`
