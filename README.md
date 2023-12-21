# DnaDesign

[![GitHub license](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/koeng101/dnadesign/blob/main/LICENSE) 
![Tests](https://github.com/koeng101/dnadesign/workflows/Test/badge.svg)
![Test Coverage](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/koeng101/e8462880f920d70b182d5df3617b30f5/raw/coverage.json)

DnaDesign is a Go project creating tools for automated genetic design, spanning from the lowest levels of DNA sequence manipulation to higher level functionality.

* **Practical:** DnaDesign tooling is meant to be used by practitioners of synthetic biology. Let's make something actually useful!

* **Modern:** DnaDesign is at the bleeding edge of technology. We are happy to adopt the newest advancements in synthetic biology, wasm, LLMs, and more to get our tools in the hands of people who need it.

* **Ambitious:** DnaDesign's goal is to be the most complete, open, and well used collection of computational synthetic biology tools ever assembled. If you like our dream and want to support us please star this repo, request a feature, or open a pull request.

## Documentation

* **[Library](https://pkg.go.dev/github.com/koeng101/dnadesign)**

## Contributing

Write good, useful code. Open a pull request, and we'll see if it fits!

## License

* [MIT](LICENSE)

### Sources

There are a few pieces of "complete" software that we have directly integrated into the source tree (with their associated licenses). These projects don't receive updates anymore (or just bug fixes with years between). In particular, `lib` has most of these, since we intend to have as few dependencies as possible in `lib`. The integrated projects include the following.
- [svb](https://github.com/rleiwang/svb) in `lib/bio/slow5/svb`
- [intel-cpuid](https://github.com/aregm/cpuid) in `lib/bio/slow5/svb/intel-cpuid`
- [wordwrap](https://github.com/mitchellh/go-wordwrap) in `lib/bio/genbank`

### Other

DnaDesign is a fork of [Poly](https://github.com/TimothyStiles/poly) at commit f76bf05 with a different mission focus. 
