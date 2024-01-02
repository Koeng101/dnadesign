# DnaDesign

[![GitHub license](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/koeng101/dnadesign/blob/main/LICENSE) 
![Tests](https://github.com/koeng101/dnadesign/workflows/Test/badge.svg)
![Test Coverage](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/koeng101/e8462880f920d70b182d5df3617b30f5/raw/coverage.json)

DnaDesign is a Go project creating tools for automated genetic design, spanning from the lowest levels of DNA sequence manipulation to higher level functionality.

* **Practical:** DnaDesign tooling is meant to be used by practitioners of synthetic biology.

* **Modern:** DnaDesign is at the bleeding edge of technology. We are happy to adopt the newest advancements in synthetic biology, wasm, LLMs, and more to get our tools in the hands of people who need it.

* **Ambitious:** DnaDesign's goal is to be the most complete, open, and well used collection of computational synthetic biology tools ever assembled. If you like our dream and want to support us please star this repo, request a feature, or open a pull request.

## Documentation

* **[Library](https://pkg.go.dev/github.com/koeng101/dnadesign)**

## Repo organization

* `lib` contains the core DnaDesign library, with nearly all functionality, all in idiomatic Go with nearly no dependencies.
    * `lib/bio` contains biological parsers for file formats including genbank, fasta, uniprot, fastq, slow5, sam, and pileup files.
    * `lib/align` contains Needleman-Wunsch and Smith-Waterman alignment functions, as well as the mash similarity algorithm.
    * `lib/clone` contains functions for simulating DNA cloning, including restriction digestion, ligation, and GoldenGate assembly.
    * `lib/fold` contains DNA and RNA folding simulation software, including the Zuker and Linearfold folding algorithms.
    * `lib/primers` contains DNA primer design and PCR simulation functions.
    * `lib/seqhash` contains the Seqhash algorithm to create universal identifiers for DNA/RNA/protein.
    * `lib/synthesis` contains various functions for designing synthetic DNA.
        * `lib/synthesis/codon` contains functions for working with codon tables, translating genes, and optimizing codons for expression.
        * `lib/synthesis/fragment` contains functions for optimal GoldenGate fragmentation.
        * `lib/synthesis/fix` contains functions for fixing proteins in preparation for synthesis.
    * `lib/transform` contains basic utility functions for transforming DNA, like reverse complementation.
* `external` contains integrations with external bioinformatics software, usually operating on the command line.
    * `external/minimap2` contains a function for working with minimap2 with Go.
    * `external/samtools` contains a function for generating pileup files with Go.


## Contributing

Write good, useful code. Open a pull request, and we'll see if it fits!

## License

* [MIT](LICENSE)

## Sources

There are a few pieces of "complete" software that we have directly integrated into the source tree (with their associated licenses). These projects don't receive updates anymore (or just bug fixes with years between). In particular, `lib` has most of these, since we intend to have as few dependencies as possible in `lib`. The integrated projects include the following.
- [svb](https://github.com/rleiwang/svb) in `lib/bio/slow5/svb`
- [intel-cpuid](https://github.com/aregm/cpuid) in `lib/bio/slow5/svb/intel-cpuid`
- [wordwrap](https://github.com/mitchellh/go-wordwrap) in `lib/bio/genbank`

## Other

DnaDesign is a fork of [Poly](https://github.com/TimothyStiles/poly) at commit f76bf05 with a different mission focus. 

# Changelog

All notable changes to this project will be documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
- Added minimap2 and samtools(pileup) integrations in external (#46)
- Added sam parser (#5)
- Added the LinearFold folding algorithms (#38)
- Added Get function to uniprot for getting a single uniprot xml from online (#37)
- Removed murmur3 in favor of crc32 for mash (#33)
- Patch start codon problems (#32)
- Added tests for OpenBSD (#31)
- Removed a large number of unneeded dependencies (#28)
- Added full rebase text database into tree (27b41fb)
- Updated uniprot to be a standardized parser (#22)
- Purged unsupported gff parser (#23)
- Moved library to lib directory (#21)
- Fixed issue with JSON codon tables (#4)
- Added Seqhash v2 (#3)
- Added lowercase methylation options during cloning (#2)
- Standardized parsers with generics (#1)
