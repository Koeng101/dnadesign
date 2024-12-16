**NOTE! This is old documentation on linkers, which is still relevant for part overhangs, but the linkers themselves are still under design.**

# Linkers

There have been many standard methods for assemblying DNA, with the most notable being the [BioBrick assembly](https://en.wikipedia.org/wiki/BioBrick) method being developed in 2003. However, BioBrick assembly could only assemble 2 DNA parts at once, and therefore limited the amount of assembly that one could do in a short amount of time. In 2008, a new method called [GoldenGate assembly](https://dx.doi.org/10.1371%2Fjournal.pone.0003647) was developed that overcame limitations of BioBrick assembly, allowing many fragments to be put together at once in a single tube. In 2011, this GoldenGate assembly was standardized with the [MoClo assembly](https://doi.org/10.1371/journal.pone.0016765) method. 

The Linkers Collection is a collection of linkers ([here](https://doi.org/10.3389/fbioe.2019.00271) is a good introduction) for MoClo assembly. These linkers have 2 special attributes:

- The BsaI overhangs are optimized using [empirical data](https://doi.org/10.1371/journal.pone.0238592)
- The assembly is recursive - the same linkers are used at each level of assembly. This is accomplished using methyltransferases which [methylate at GACNNNGTC](https://doi.org/10.1093%2Fnar%2Fgky596) or [methylate at CCGG positions](http://www.greatlakesbiotech.org/news/2016/8/26/designing-a-low-cost-molecular-biology-platform)

### Changes
Linkers were recreated from my previous toolkits because `CGAG,GTCT` are required for the assembly method, and these greatly lowered the efficiency of traditional MoClo, mainly because GGAG is used as one of the classic overhangs. I added `AAAA` and `GGGG` as two standard overhangs - while they aren't used in in normal assemblies, they are available for if a fragment needs to build polyA/polyT/polyG/polyC.

overhangs: `CGAG,GTCT,GGGG,AAAA,AACT,AATG,ATCC,CGCT,TTCT,AAGC,ATAG,ATTA,ATGT,ACTC,ACGA,TATC,TAGG,TACA,TTAC,TTGA,TGGA,GAAG,GACC,GCCG`
overhangs with numbers:
1. CGAG
2. TACA
3. AACT
4. AATG
5. ATCC
6. CGCT
7. GTCT
8. AAGC
9. ATAG
10. ATTA
11. TTCT
12. ATGT
13. ACTC
14. ACGA
15. TATC
16. TAGG
17. TTAC
18. TTGA
19. TGGA
20. GAAG
21. GACC
22. GCCG
23. AAAA
24. GGGG

Once a plasmid cloned using a linker, its new definition is derived from the particular linkers used to construct it. For example, an assembly with A1(2) + B1(4) creates a new plasmid with a fragment defined with the overhangs 2 and 4.

Note: 11 is no longer defined as ntag.

# Normal build process
### Simple build
A simple build constructs an Escherichia coli vector with no added fluff.
* [A1]  CGAG - TACA (linker prefix)
* [P]   TACA - AACT (promoter)
* [R]   AACT - AATG (rbs)
* [C]   AATG - ATCC (cds)
* [T]   ATCC - CGCT (terminator)
* [B1]  CGCT - GTCT (linker suffix)
* [E1]  GTCT - CGAG (e coli vector)

Or, with numbered overhangs:

```
Simple build:
1 [linker prefix] 2 [promoter] 3 [RBS] 4 [CDS] 5 [terminator] 6 [linker suffix] 7 [vector1] ...
```

### Operon assembly
The first operon component, X number of mid operon components, and the last operon component can then be combined based off of their prefix and suffix linkers.

Operon assembly first:
* [A1]  CGAG - TACA (linker prefix)
* [P]   TACA - AACT (promoter)
* [R]   AACT - AATG (rbs)
* [C]   AATG - ATCC (cds)
* [B2]  ATCC - GTCT (linker suffix)
* [E1]  GTCT - CGAG (e coli vector)

Operon assembly mid:
* [A2]  CGAG - AACT (linker prefix)
* [R]   AACT - AATG (rbs)
* [C]   AATG - ATCC (cds)
* [B2]  ATCC - GTCT (linker suffix)
* [E1]  GTCT - CGAG (e coli vector)

Operon assembly last:
* [A2]  CGAG - AACT (linker prefix)
* [R]   AACT - AATG (rbs)
* [C]   AATG - ATCC (cds)
* [T]   ATCC - CGCT (terminator)
* [B1]  CGCT - GTCT (linker suffix)
* [E1]  GTCT - CGAG (e coli vector)

```
Operon assembly (first):
1 [linker prefix] 2 [promoter] 3 [RBS] 4 [CDS] 5 [linker suffix] 7 [vector1] ...

Operon assembly (mid):
1 [linker prefix] 3 [RBS] 4 [CDS] 5 [linker suffix] 7 [vector1] ...

Operon assembly (last):
1 [linker prefix] 3 [RBS] 4 [CDS] 5 [terminator] 6 [linker suffix] 7 [vector1] ...
```


### Shuttle vector
Often, users will want to move transcription units to new organisms of interest. The following is a simple shuttle vector:

* [A]   CGAG - TACA (linker prefix)
* [P]   TACA - AACT (promoter)
* [R]   AACT - AATG (rbs)
* [C]   AATG - ATCC (cds)
* [T]   ATCC - CGCT (terminator)
* [B]   CGCT - GTCT (linker suffix)
* [S]   GTCT - AAGC (target selective marker)
* [D]   AAGC - ATAG (target origin of replication)
* [E2]  ATAG - CGAG (e coli vector 2)

```
Simple shuttle:
1 [linker prefix] 2 [promoter] 3 [RBS] 4 [CDS] 5 [terminator] 6 [linker suffix] 7 [target ori] 8 [target marker] 9 [vector2] ...
```

### Integration vector
Instead of shuttle vectors, users will sometimes want to integrate sections of DNA into their organism of interest. The following is a build definition for an integration vector:

* [A]   CGAG - TACA (linker prefix)
* [P]   TACA - AACT (promoter)
* [R]   AACT - AATG (rbs)
* [C]   AATG - ATCC (cds)
* [T]   ATCC - CGCT (terminator)
* [B]   CGCT - GTCT (linker suffix)
* [S]   GTCT - AAGC (target selective marker)
* [D]   AAGC - ATAG (target upstream homology)
* [EC3] ATAG - ATTA (e coli vector 3)
* [U]   ATTA - CGAG (upstream homology)
```
Integration vector:
1 [linker prefix] 2 [promoter] 3 [RBS] 4 [CDS] 5 [terminator] 6 [linker suffix] 7 [downstream homology]  8 [target marker] 9 [vector3] 10 [upstream homology] ...
```

### Protein tags
It is common that one would want to add tags to a protein sequence. You can add to the N terminal or C terminal. The N tag is carried along with the ribosomal binding site.

* [A1]  CGAG - TACA (linker prefix)
* [P]   TACA - AACT (promoter)
* [R]   AACT - AATG (rbs/nterminal tag)
* [C]   AATG - ATCC (cds)
* [Cc]  ATCC - ATGT (c terminal tag)
* [Tt]  ATGT - CGCT (terminator with c terminal tag)
* [B1]  CGCT - GTCT (linker suffix)
* [E1]  GTCT - CGAG (e coli vector)

Or, with numbered overhangs:

```
Protein tag build:
1 [linker prefix] 2 [promoter] 3 [RBS/N tag] 4 [CDS] 5 [C tag] 12 [terminator] 6 [linker suffix] 7 [vector1] ...
```

### Conserved overhangs
The following overhangs are reserved for more complicated backbone assemblies when manipulating Escherichia coli backbones. These represent overhangs 13, 14, 15, 16.

* [EFX] NNNN - ACTC (e coli vector compatibilizer prefix)
* [M]   ACTC - ACGA (e coli marker) {always in R6K backbones}
* [O]   ACGA - TATC (e coli origin)
* [Z]   TATC - TAGG (package signal, usually oriT)
* [ERX] TAGG - NNNN (e coli vector compatibilizer suffix)

EFX and ERX can have the following definitions:
* EF1
* EF2
* ER1
* ER3

An E3 vector, for example, would use the flanks EF2 and ER3, while E1 would use EF1 and ER1.

Backbones parts are typically used in situations where users can't rely on ccdB counter-selection normally present in cached vectors - for example, when constructing a new vector that a single part, like a CDS, can be integrated into. 

In addition, the following overhangs are used for specialty vector construction:
* TTAC
* TTGA

# Special constructions
### Backbone part construction
Parts can be constructed to function in backbones, except type M parts.
```
1 [linker prefix (7)] 2 [promoter] 3 [RBS] 4 [Kanamycin resistance coding sequence] 5 [terminator] 6 [linker suffix (8)] 7 [negative selection marker] 8... [vector] ...
After assembly: 
7 [Kanamycin resistance] 13
```

### Reversing during higher level construction
Each linker is defined with a number, representing the overhang it switches its assembly to. These can be negative numbers, which are the reverse complements of that particular overhang. These can be used to flip sequences during construction. For example:
```
Simplified transcription unit now represented by `-->`

1( ----> )2 + -3( --> )-2 + 3( -> )4 = ( ----> <-- -> )
```

## Vectors
### Vector [E0] construction
E0 vectors, or vectors used in recurse builds or normal foundry synthesis orders, are always constructs by-hand in a non-modular fashion using BbsI. Users cannot construct E0 vectors in our foundry.

### Resistance marker construction (M)
Escherichia coli resistrction marker parts, or M parts, are always in R6K vectors. This allows switching to a non-R6K strain as a way to select out the original vector. The foundry does provide R6K transformation resources.

### Vector [E1,E2,E3] construction
Vector types E1,E2,E3 are used in constructing normal DNA. They use ccdB for negative selection. They are always constructed from (M) parts, so that the marker can be switched during the GoldenGate reaction.

### Specialty vector construction
Speciality vectors use special linkers to add a `ccdB-MOsp87` cassette into any site of a normal construction. For example, if you have an expression vector that you know works well, you can swap out your gene of interest with 2 linkers and `ccdB-MOsp87`. This creates a new vector that you can directly add new genes of interest into, without adding the corresponding promoters or terminators to the reaction.

### Recurse builds
Recurse builds are the exception case to the rule that assemblies are redefined with their given linkers. In recurse builds, nothing is ever redefined: whatever overhang the input fragments had, the output assembly will have.

A recursive build is designed to build DNA from blocks. The designed DNA can be built from any number of blocks at any given step: If there are 20 blocks to be added together, one could design the DNA to be built with 4-5, or 2-2-5, or anything else. For difficult sequences, this allows clone-time optimizations, without going back to the synthesis phase. 
* [A1]   CGAG - {} (recurse linker prefix)
* [ ]   {} - {} (fragment of interest)
* [B1]   {} - GTCT (recurse linker suffix)
* [E1]   GTCT - CGAG (e coli vector)

```
recurse build:
1 [recurse linker prefix] x [n] x [n+1...] x [recurse linker suffix] 7 [vector1] ...
```


# FAQ
## How were the linkers designed
The efficiency designer is based off of [this datasheet](https://doi.org/10.1371/journal.pone.0238592.s001) from "Enabling one-pot Golden Gate assemblies of unprecedented complexity using data-optimized assembly design". I am using [Poly](https://github.com/timothystiles/poly)'s fragment designer, which I coded.

## What are linkers?
When building a construct using GoldenGate, simply ligate linkers between your vector and your genes during a GoldenGate reaction to enable use of that gene in multi-gene constructs. 

Typically, you will do an assembly reaction (also known as a level 1 cloning reaction in MoClo lingo) to give context to your gene. For example, you may have a protein called GFP that you wish to express. In this case, you would do a level 1 cloning reaction to contextualize GFP with a proper promoter and terminator for your target organism to make a transcriptional unit (TU). In that reaction, you may have to add linkers to connect your construct into the vector it belongs in. 

Afterwards, you can combine the GFP transcriptional unit to up to 24 other constructs with clever usage of linkers. To answer specifically which ones to use and when, read below.

## What linkers are included in the Linkers Collection?

This collection has 384 linkers. There are 96 linkers for building independent transcription units, 96 linkers for building operons, 96 linkers for recursive builds, and 96 linkers for defining new speciality vectors. For each set of 96, linkers are split into 48 prefix and suffix linkers. Those 48 linkers are split into 24 positive and 24 negative linkers. The positive linkers are used for constructing genes in the forward direction, and the negative linkers are used to construct genes in the reverse direction.

Each overhang is assigned a number. The reverse complement of each overhang is represented as the negative version of its number, which is also how we are able to flip constructs.

Linkers are named with a 1 letter + 3 number scheme, separated by underscores ( _ ) in the format `Y_X_X_X`. Y describes the direction (F, or forward, for prefix, and R, or reverse, for suffix) of the linker. The first 2 numbers describe the two overhangs which the linker itself will be cut out with, the third number describes the overhang which the linker will introduce to the construct. After a GoldenGate assembly and transformation, anything between the prefix and suffix linkers can be cut out with BsaI and used in another assembly.
