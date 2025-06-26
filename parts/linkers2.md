# DnaDesign Assembly

DnaDesign Assembly (shortened as dd assembly) is a GoldenGate DNA assembly method similar to MoClo. It is designed from the bottom up for manufacturing with the intention of globally lowering the cost of useful synthetic DNA. The overhang set was changed from MoClo in order to accomodate the overhangs `GTCT` and `CGAG`, which are used in recursive DNA assemblies. The BsaI overhangs were optimized using [empirical data](https://doi.org/10.1371/journal.pone.0238592) from the paper "Enabling one-pot Golden Gate assemblies of unprecedented complexity using data-optimized assembly design".

## Simplest user perspective

From the simplest user perspective, dd assembly has genetic parts. These parts can be put together in a modular fashion into much larger genetic designs. Synthesis and assembly time is minimized, while allowing the user to do almost anything they want - from creating fusion proteins, transcriptional units, operons, and shuttle vectors. If they do not require novel synthesis, the goal of dd assembly is to allow time from ordering to recieving clonal DNA within 3 days (3 day plasmids), regardless of the size or complexity.

Here are the basic part definitions:

1. TACA - AACT: Promoter
2. AACT - AATG: RBS/kozak/ntag
3. AATG - ATCC: CDS
4. ATCC - CGCT: Terminator 

Some additional useful definitions:

1. AACT - [GG]AAGC: ntag1/ntag2

## Overhangs

overhangs: `CGAG,GTCT,GGGG,AAAA,AACT,AATG,ATCC,CGCT,TTCT,AAGC,ATAG,ATTA,ATGT,ACTC,ACGA,TATC,TAGG,TACA,TTAC,TTGA,TGGA,GAAG,GACC,GCCG`
overhangs with numbers:
X. GTCT
Y. CGAG

2. TACA
3. AACT
4. AATG
5. ATCC
6. CGCT
7. TACA
8. AAGC
9. ATAG
10. ATTA
11. TTCT
12. ATGT
13. ACTC
14. ACGA
15. TATC
16. TAGG
17. TACA
18. TTAC
19. TTGA
20. TGGA
21. GAAG
22. GACC
23. GCCG
24. AAAA
25. GGGG

This document is techincal reference material, not a how-to guide or tutorial. It contains the following sections:

1. Assembly
2. Parts
3. Vectors
4. Primers
5. Cache blocks

### Methylation redefinition

A novel approach of dd assembly is that of modular redefinition connectors. While the yeast toolkit (YTK) has the concept of connectors, dd assembly goes a step further by defining these as synthesized duplexes that are methylated, therefore making BsaI reusable, while also giving the assembly the interesting property of being able to recurse onto itself. Here is an example of two connectors you could add to your build:

```
Build
--------------
promoter
rbs
cds
terminator
f158_GTCT_TACA
f159_CGAG_AACT
==
f158_GTCT_TACA - promoter - rbs - cds - terminator - f159_CGAG_AACT
```

This would be considered a basic build. The connector sequences redefine what will be cut out next: in this case, you can cut this cassette out to be a `TACA`-`AACT` part - ie, taking the place of a promoter. This can become much more powerful when you do things like this:

```
Build
--------------
rbs
cds
terminator
f158_AACT_AACT
f159_CGCT_CGCT
==
f158_AACT_AACT - rbs - cds - terminator - f159_CGCT_CGCT
```

What you have done now is created a part, which to all other assemblies looks like a rbs-cds-terminator combination. This can then be directly used in another assembly with a promoter to do an efficient 2 part assembly.

# Assembly

Standard assemblies are with the following overhangs: `[GTCT] TACA AACT AATG ATCC CGCT [CGAG]`. A basic build is as follows:

```
GTCT    connector1  TACA
TACA    promoter    AACT
AACT    RBS         AATG
AATG    CDS         ATCC
ATCC    terminator  CGCT
CGCT    connector2  CGAG
```

These sites are considered special and privileged. Each site has a connector which, after assembly, redefines the part. Connectors are defined as such:

```
direction/primer - new overhang - old overhang
===
f158.AACT.TACA
r159.AATG.CGCT
```

## Special overhangs: GTCT and CGAG

Both `GTCT` and `CGAG` are special overhangs that make dd assembly different than alternative GoldenGate standards like MoClo. These two sites enable recursive GoldenGate assembly of genetic parts by containing part of the BsaI cut site, but not an edge base pair. This edge base pair can be methylated, preventing cutting during a GoldenGate assembly. This methylation is then deprotected during amplification so that BsaI can be used again for another assembly reaction. Since these overhangs are within the BsaI cut site, when BsaI is used in subsequent reactions, it cuts user-defined DNA. Let's take a detailed look at what a vector would look like.

```
>recursive vector
... g[GTCT](NGAGACC---GGTCTCN)[CGAG]ACc ...

>insert
GGTCTCN [GTCT]CA (NNNN --- NNNN) [CGAG] NGAGACC

>result
g[GTCT]CA (NNNN --- NNNN) [CGAG]ACc
```

To break the `recursive` vector down:
1. `GGTCT` is BsaI in the forward direction and `GAGACC` is BsaI in the reverse direction. It cuts `1,4`, or `GGTCTC N [NNNN]` where `NNNN` is the overhang.
2. The rest of the circular vector (ori and amp) are simplified with `...`
3. The lowercase letters are methylated cytosines (in the complement for `g`)
4. The two overhangs in brackets are our special overhangs `GTCT` and `CGAG`
5. The sequence within the parathesis is sequence that will get cut out during the GoldenGate reaction, and will be replaced with our sequence of interest.

To break the insert down:
1. `GGTCTC` and `GAGACC` is still BsaI 
2. The user insert is `(NNNN --- NNNN)`. The 4 NNNNs on both sides are the overhangs which will be exposed after methylation is removed. 
3. `CA` after `[GTCT]` is spacer needed to properly space the user DNA from the methylation-exposed BsaI cut sites.

In dd assembly, there are two kinds of vectors: `recursive` vectors and `base` vectors. `recursive` vectors are just like they are described above, while `base` vectors derive their overhangs from `linkers` or from a special kind of insert called a `replaceable`, creating a `replaceable` vector. Since `base` vectors by themselves do not contain more BsaI sites, they can also be used to create `shuttle` vectors. We will go over each before diving into specific overhangs for parts. Briefly:

1. `recursive` vectors are used for creating synthetic DNA, independent of partification.
2. `linkers` are used to create multigene constructs. **Does not require vector intermediates**
3. `replaceable` vectors are created from `base` vectors by inserting genetic parts and a replacement insert. Mostly just used for expression vectors, but can also be used for shuttle vectors.

## Recursive construction

## Linkers
The most basic dd assembly will create a transcriptional unit. Oftentimes, however, you will want to be able to combine different transcriptional units together. In order to do these multi-level assemblies, we use linkers. Linkers are genetic parts that are used within an assembly reaction that define the part overhangs of the assembled construct, for use in the next assembly reaction. For example, we could have 3 genetic parts:

* Promoter+RBS
* GFP
* Terminator

While we could construct a simple transcriptional unit of `["Promoter+RBS", "GFP", "Terminator"]`, we might want to use this whole transcriptional unit in a different construct. In order to do that, we will add 2 linkers. The prefix linker (A) and the suffix linker (B). These linkers have additional numbers,

## Replaceable vectors

# Parts

## Orthogonal primer binding sites

### CDS fusion using SapI
```
NNN TGA AGAGC ACTT <- clean C terminus, SapI compatible
NNN GGA AGAGC ACTT <- GRAS C tag, SapI compatible
NNN GG ACTT <- GS C tag, not SapI compatible
```

In dd assembly, there are three methods of creating proteins for protein fusion:
1. Clean C terminus + SapI fusions
2. GRAS C tag + SapI fusions
3. GS C tag

In dd assembly, CDSs have either their protein tags directly fused to them, or use SapI fusion. SapI fusions are enabled by the following observation: You can overlap SapI with a stop codon to specifically cut the last codon of a protein. By cutting the last codon, without cutting any other sequence, we can create seamless protein fusions for any protein. Proteins do not need to be specifically designed to have fusion tags.

Proteins which are commonly tagged may also use GGA instead of TGA. This creates a scar of `GRAS`, which should be innocous.

```
Lys	K	AAA
Leu	L	TTA
Arg	R	CGA
Phe	F	TTT
Gln	Q	CAA
Trp	W	TGG
His	H	CAC
Pro	P	CCG
Val	V	GTG
Cys	C	TGC
Thr	T	ACA
Ser	S	TCC
Tyr	Y	TAC
Ile	I	ATA
Gly	G	GGT
Asn	N	AAC
Ala	A	GCC
Met	M	ATG
Asp	D	GAC
Glu	E	GAA
```

Terminators with ACTT should always be ACTTTAA, encoding a stop codon in case the protein has a GRAS C tag. 

# Vectors

## Standard oriT
dd assembly takes is opinionated in how transfer to non-cloning organisms should be done.

## Backbones
Canonically, the backbones look like this:
```
GTCT    connector1  TACA
TACA    promoter    AACT
AACT    RBS         AATG
AATG    CDS         ATCC
ATCC    terminator  CGCT
CGCT    connector2  CGAG
```
With `AAGC` sometimes also being involved with ntags.

This, obviously, does not leave room for building backbones. Oh no! This is where `ATTA` comes into play:

```
GTCT    connector1  TACA
TACA    promoter    AACT
AACT    ntag1       AAGC
AAGC    ntag2       AATG
AATG    CDS         ATCC
ATCC    ctag        ATGT
ATGT    Cterminator CGCT
CGCT    connector2  CGAG

TTCT - reserved for GS tag

ATTA    Upstream homology
ACTC    Selection
ATGT    Downstream homology
```

So if we add these into the most complex assembly that one can do at one time, we get this:

```
GTCT    connector1  ATTA
ATTA    upstream    TACA
TACA    promoter    AACT
AACT    ntag1       AAGC
AAGC    ntag2       AATG
AATG    CDS         ATCC
ATCC    ctag        ATGT
ATGT    Cterminator CGCT
CGCT    selection   ACTC
ACTC    downstream  ATGT
ATGT    connector2  CGAG

+ vector = 12 part assembly
```

Due to the flexibility of redefinitions, you can theoretically fit a lot more into each spot - you could, for example, fit an entire operon + selection marker in the "selection" location.

# Cache blocks

Cache blocking is a concept unique to dd assembly, designed to give us a way to feasibly synthesize and test massive sequences. It stems from one fundamental observation: we are limited in our ability to synthesize correct DNA. Cache blocking aims to minimize the necessity of synthesis when creating and testing DNA.

In practical terms, you chunk a given sequence into `cache blocks`, which are clonally verified. Each `cache block` is defined by dd assembly overhangs, so can be used like any other construct, but unlike when doing classic dd assembly, these `cache blocks` are seamless. They are fragmented at dd assembly overhangs, but they do not have scar sequences. Subsequent cache block assemblies simply maintain whatever overhangs were on the edge `cache blocks`, and these assemblies can create new `cache blocks` that get sequence verified, or create a final desired sequence.

## Mutational limiting

When you use directly utilize synthetic DNA, you are typically limited by the mutation rate of the synthesis reaction. Synthesis blocks have a lot of mutations. This means you have to clone more intermediate fragments. One idea behind `cache blocks` is that, if you take the upfront cost of clonally verifying each block, the replacement of any given block within a target sequence becomes increasingly lower.

Let's say, for example, we have a 8kbp metabolic circuit that is `cache block`ed down to 250bp, or 32 `cache blocks`. If you use an AI system to mutate one of the genes, or perhaps some of the ribosomal binding sites, you would just swap the specific blocks you need to change. If you're only changing a few blocks, you only need to synthesize the 250bp from those blocks, lowering synthesis costs, but also lowering mutation rates to a screenable level, whereas it is very difficult to do that with 8kbp of purely synthetic DNA.

An important thing to note here is that the `cache blocks` essentially act as constants: so the same concept that works with 8kbp works with 50kbp, or perhaps even **genomes**. In this way, we make whole genome rearrangement and testing trivial: you only need to resynthesize the specific blocks you're changing. The entire system can work computationally in a consistent manner for testing any piece of DNA in a modular fashion.

## Assembly caching

An astute reader may still observe that we need to assemble an awful lot of `cache blocks`. There are two ways we handle this: assembly caching using `identity` linkers and clone-less assembly.

Linker based recursion is similar to the above linker section with one exception: all linkers are `identity` linkers. `identity` linkers do not redefine their overhangs. They simply take in `cache blocks` (or parts, since they appear the same to dd assembly), and spit out assemblies of those `cache blocks` with the edge overhangs exposed. This enables you to create compositions of `cache blocks` as `cache blocks` themselves - for example, in our 8bkp example, if we are only changing the first 2kbp, we can `cache block` the remaining 6kbp as a single block. This new block can be sequence verified and used further.

However, we can also use clone-less assembly, as described above. Basically, we do not pause for a cloning step between putting together a number of `cache blocks` together. We simply amplify the resultant GoldenGate and continue with the next step of assembly. The most important thing that this process does is limit our need to achieve equimolar GoldenGate ratios for efficient assembly - as this would take in intermediate quantification and normalization step which is, ironically, more expensive and more annoying than simply recursing on assemblies. This also allows reuse of overhangs - for example, if you want to assemble a transcriptional unit.

## Plasmid resynthesis and Genome testing

There are two killer applications of `cache blocks`: plasmid resynthesis and genome testing. Many plasmids use the same components - ampicillin resistance, pUC origin, etc - and these can be cached in such a way that users can synthesize arbitrary plasmids without even thinking about parts - in each case, from an end-user perspective, the only thing that gets synthesized is the user-specified DNA with some minimal flanking sequence to compensate on either side since the mutations are limited and we assembly cache most of the vector backbone. In scaled facilities that can handle large quantities of DNA, this eliminates the need for any vector onboarding.

Perhaps the most difficult part of creating new synthetic genomes is testing whether or not changes work. Once cached, however, this becomes much much easier: parts can be swapped out piecewise, with synthesis and assembly only occuring at locations with changes, in a similar way to plasmid resynthesis, except at scale. Testing any particular change becomes just a task of assembly.

### Genome replacement

Rather than focus on methods of genome replacement that are specific to certain organisms on the basis of their unique properties (for example, homologous recombination/integration into yeast, natural competence of Bacillus subtilis), dd assembly foucses
