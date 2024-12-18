# DnaDesign

[![GitHub license](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/koeng101/dnadesign/blob/main/LICENSE) 
![Tests](https://github.com/koeng101/dnadesign/workflows/Test/badge.svg)
![Test Coverage](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/koeng101/e8462880f920d70b182d5df3617b30f5/raw/coverage.json)

DnaDesign is a Go project creating tools for automated genetic design, spanning from the lowest levels of DNA sequence manipulation to higher level functionality.

* **Practical:** DnaDesign tooling is meant to be used by practitioners of synthetic biology.

* **Modern:** DnaDesign is at the bleeding edge of technology. We are happy to adopt the newest advancements in synthetic biology, wasm, LLMs, and more to get our tools in the hands of people who need it.

* **Ambitious:** DnaDesign's goal is to be the most complete, open, and well used collection of computational synthetic biology tools ever assembled. If you like our dream and want to support us please star this repo, request a feature, or open a pull request.

## Documentation

* **[Python Documentation](https://py.dnadesign.bio)**
* **[Go Library Documentation](https://pkg.go.dev/github.com/koeng101/dnadesign)**

## Repo organization

On the highest level:
* `lib` contains core functionality as a go library.
* `external` contains functions to work with external bioinformatics command-line interfaces.
* `py` contains code to use the dnadesign library in python using a C shared library.

### Detailed repo organization

* [lib](https://pkg.go.dev/github.com/koeng101/dnadesign/lib) contains the core DnaDesign library, with nearly all functionality, all in idiomatic Go with nearly no dependencies.
    * [lib/bio](https://pkg.go.dev/github.com/koeng101/dnadesign/lib/bio) contains biological parsers for file formats including [genbank](https://github.com/Koeng101/dnadesign/blob/main/lib/bio/genbank/genbank.go), [fasta](https://github.com/Koeng101/dnadesign/blob/main/lib/bio/fasta/fasta.go), [uniprot](https://github.com/Koeng101/dnadesign/blob/main/lib/bio/uniprot/uniprot.go), [fastq](https://github.com/Koeng101/dnadesign/blob/main/lib/bio/fastq/fastq.go), [slow5](https://github.com/Koeng101/dnadesign/blob/main/lib/bio/slow5/slow5.go), [sam](https://github.com/Koeng101/dnadesign/blob/main/lib/bio/sam/sam.go), and [pileup](https://github.com/Koeng101/dnadesign/blob/main/lib/bio/pileup/pileup.go) files.
    * [lib/align](https://pkg.go.dev/github.com/koeng101/dnadesign/lib/align) contains [Needleman-Wunsch](https://en.wikipedia.org/wiki/Needleman%E2%80%93Wunsch_algorithm) and [Smith-Waterman](https://en.wikipedia.org/wiki/Smith%E2%80%93Waterman_algorithm) alignment functions, as well as the [mash](https://doi.org/10.1186/s13059-016-0997-x) similarity algorithm.
    * [lib/clone](https://pkg.go.dev/github.com/koeng101/dnadesign/lib/clone) contains functions for simulating [DNA cloning](https://en.wikipedia.org/wiki/Molecular_cloning), including [restriction digestion](https://www.neb.com/en-us/applications/cloning-and-synthetic-biology/dna-preparation/restriction-enzyme-digestion), [ligation](https://en.wikipedia.org/wiki/Ligation_(molecular_biology)), and [GoldenGate assembly](https://en.wikipedia.org/wiki/Golden_Gate_Cloning).
    * [lib/fold](https://pkg.go.dev/github.com/koeng101/dnadesign/lib/fold) contains DNA and RNA folding simulation software, including the [Zuker](https://doi.org/10.1093/nar/9.1.133) and [LinearFold](https://doi.org/10.1093/bioinformatics/btz375) folding algorithms.
    * [lib/primers](https://pkg.go.dev/github.com/koeng101/dnadesign/lib/primers) contains [DNA primer](https://www.nature.com/scitable/definition/primer-305/) design functions.
        * [lib/primers/pcr](https://pkg.go.dev/github.com/koeng101/dnadesign/lib/primers/pcr) contains [PCR](https://www.ncbi.nlm.nih.gov/probe/docs/techpcr/) simulation functions.
    * [lib/seqhash](https://pkg.go.dev/github.com/koeng101/dnadesign/lib/seqhash) contains the Seqhash algorithm to create universal identifiers for DNA/RNA/protein.
    * [lib/synthesis](https://pkg.go.dev/github.com/koeng101/dnadesign/lib/synthesis) contains various functions for designing synthetic DNA.
        * [lib/synthesis/codon](https://pkg.go.dev/github.com/koeng101/dnadesign/lib/synthesis/codon) contains functions for working with [codon tables](https://en.wikipedia.org/wiki/DNA_and_RNA_codon_tables), [translating genes](https://en.wikipedia.org/wiki/Translation_(biology)), and [optimizing codons](https://doi.org/10.1073/pnas.0909910107) for expression.
        * [lib/synthesis/fragment](https://pkg.go.dev/github.com/koeng101/dnadesign/lib/synthesis/fragment) contains functions for [optimal GoldenGate fragmentation](https://doi.org/10.1371/journal.pone.0238592).
        * [lib/synthesis/fix](https://pkg.go.dev/github.com/koeng101/dnadesign/lib/synthesis/fix) contains functions for fixing proteins in preparation for synthesis.
    * [lib/transform](https://pkg.go.dev/github.com/koeng101/dnadesign/lib/transform) contains basic utility functions for transforming DNA, like reverse complementation.
* [external](https://pkg.go.dev/github.com/koeng101/dnadesign/external) contains integrations with external bioinformatics software, usually operating on the command line.
    * [external/minimap2](https://pkg.go.dev/github.com/koeng101/dnadesign/external/minimap2) contains a function for working with [minimap2](https://github.com/lh3/minimap2) with Go.
    * [external/samtools](https://pkg.go.dev/github.com/koeng101/dnadesign/external/samtools) contains a function for generating pileup files using [samtools](https://github.com/samtools/samtools) with Go.
    * [external/bcftools](https://pkg.go.dev/github.com/koeng101/dnadesign/external/bcftools) contains GenerateVCF to generate a VCF file from sam alignments using [bcftools](https://samtools.github.io/bcftools/) with Go.

## Python

We have python package, `dnadesign`, which allows python users to use dnadesign. This is a work-in-progress: more documentation coming soon!

## Contributing

Write good, useful code. Open a pull request, and we'll see if it fits!

## License

* [MIT](LICENSE)

## Sources

There are a few pieces of "complete" software that we have directly integrated into the source tree (with their associated licenses). These projects don't receive updates anymore (or just bug fixes with years between). In particular, `lib` has most of these, since we intend to have as few dependencies as possible in `lib`. The integrated projects include the following.
- [svb](https://github.com/rleiwang/svb) in `lib/bio/slow5/svb`
- [intel-cpuid](https://github.com/aregm/cpuid) in `lib/bio/slow5/svb/intel-cpuid`
- [wordwrap](https://github.com/mitchellh/go-wordwrap) in `lib/bio/genbank`
- [errgroup](https://cs.opensource.google/go/x/sync/+/master:errgroup/) in `lib/bio/`

## Other

DnaDesign is a fork of [Poly](https://github.com/TimothyStiles/poly) at commit f76bf05 with a different mission focus. 

# Changelog

All notable changes to this project will be documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
- Adds uniref parser [#107](https://github.com/Koeng101/dnadesign/pull/107)
- Fixes iso-8859-1 error in reading uniref data dumps [#106](https://github.com/Koeng101/dnadesign/pull/106)
- Updates uniprot parser to read IDs [#104](https://github.com/Koeng101/dnadesign/pull/104)
- Fixes RecursiveFragment to not add flanks to the initial input [#102](https://github.com/Koeng101/dnadesign/pull/102)
- Fixes add flank bug, releases new version of python lib [#101](https://github.com/Koeng101/dnadesign/pull/101)
- Adds feature for adding flanks to RecursiveFragment. [#100](https://github.com/Koeng101/dnadesign/pull/100)
- Adds cloning and recursion functions to python. [#96](https://github.com/Koeng101/dnadesign/pull/96)
- Adds recursive fragmentation. [#92](https://github.com/Koeng101/dnadesign/pull/92)
- Updated megamash documentation to be more specific. [#91](https://github.com/Koeng101/dnadesign/pull/91)
- Adds automatic python documentation generation. [#88](https://github.com/Koeng101/dnadesign/pull/88)
- Adds genbank parsing to python package. Release version 0.1.5 of dnadesign python. [#87](https://github.com/Koeng101/dnadesign/pull/87)
- Adds fastq parsing to python package. Releases version 0.1.4 of dnadesign python. [#86](https://github.com/Koeng101/dnadesign/pull/86)
- Integrated errgroup into source tree [#84](https://github.com/Koeng101/dnadesign/pull/84)
- Added kmer detection for ligation events in cloning and removed enzyme manager [#83](https://github.com/Koeng101/dnadesign/pull/83)
- Added option for linear ligations [#82](https://github.com/Koeng101/dnadesign/pull/82)
- Added minimal python packaging [#81](https://github.com/Koeng101/dnadesign/pull/81)
- Greatly simplified the Ligate function [#77](https://github.com/Koeng101/dnadesign/pull/77)
- Updated barcoding functions to handle edge case of hanging-edge barcodes [#74](https://github.com/Koeng101/dnadesign/pull/74)
- Updated megamash to use int instead of uint for minimal Kmer counts (so you can use -1) [#73](https://github.com/Koeng101/dnadesign/pull/73)
- Added bcftools to external [#72](https://github.com/Koeng101/dnadesign/pull/72)
- Fixed bug in PCR where certain matching primers wouldn't create any amplicons [#71](https://github.com/Koeng101/dnadesign/pull/71)
- Updated seqhash2 to use base58 rather than base64 [#69](https://github.com/Koeng101/dnadesign/pull/69)
- Updated dual barcodes primer sets to be created without csv files [#67](https://github.com/Koeng101/dnadesign/pull/67)
- Added workers to bio as a way to process data [#62](https://github.com/Koeng101/dnadesign/pull/62)
- Improved megamash efficiency and added []Match JSON conversion [#61](https://github.com/Koeng101/dnadesign/pull/61)
- Added barcoding functionality for sequencing reads [#59](https://github.com/Koeng101/dnadesign/pull/59)
- Added the megamash algorithm [#50](https://github.com/Koeng101/dnadesign/pull/50)
- Changed parsers to return values instead of pointers. Added some sequencing utils [#49](https://github.com/Koeng101/dnadesign/pull/49)
- Added minimap2 and samtools(pileup) integrations in external [#46](https://github.com/Koeng101/dnadesign/pull/46)
- Added sam parser [#5](https://github.com/Koeng101/dnadesign/pull/5)
- Added the LinearFold folding algorithms [#38](https://github.com/Koeng101/dnadesign/pull/38)
- Added Get function to uniprot for getting a single uniprot xml from online [#37](https://github.com/Koeng101/dnadesign/pull/37)
- Removed murmur3 in favor of crc32 for mash [#33](https://github.com/Koeng101/dnadesign/pull/33)
- Patch start codon problems [#32](https://github.com/Koeng101/dnadesign/pull/32)
- Added tests for OpenBSD [#31](https://github.com/Koeng101/dnadesign/pull/31)
- Removed a large number of unneeded dependencies [#28](https://github.com/Koeng101/dnadesign/pull/28)
- Added full rebase text database into tree [27b41fb](https://github.com/Koeng101/dnadesign/commit/27b41fb4fdb849d569278c965849e6f28fb2a7f6)
- Updated uniprot to be a standardized parser [#22](https://github.com/Koeng101/dnadesign/pull/22)
- Purged unsupported gff parser [#23](https://github.com/Koeng101/dnadesign/pull/23)
- Moved library to lib directory [#21](https://github.com/Koeng101/dnadesign/pull/21)
- Fixed issue with JSON codon tables [#4](https://github.com/Koeng101/dnadesign/pull/4)
- Added Seqhash v2 [#3](https://github.com/Koeng101/dnadesign/pull/3)
- Added lowercase methylation options during cloning [#2](https://github.com/Koeng101/dnadesign/pull/2)
- Standardized parsers with generics [#1](https://github.com/Koeng101/dnadesign/pull/1)

