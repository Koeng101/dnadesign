# KG Parts Library

This is the KG Genetic Parts library. It is written with basic yaml for data portability purposes.

The library is static and intended for embedded distribution in python or go, or online using json. The built files in json or yaml are therefore saved to the project itself.

## Purpose

The KG Genetic Parts Library is a complete genetic parts library, covering all major organisms, with consistent and high-quality documentation. 

## Rules
1. All genes must have a unique name, including genes used between different toolkits
2. Parts should be accessible from a URL appended with `.json`. No weird special characters or spaces.
3. A given sequence should only have 1 name.
4. In cases of proteins or tags encoded for a certain organism, add a parathesis tag to the end of the protein. For example, `SceI(Scerevisae)`.
5. Sequences are identified by their fragment seqhash. Any code interacting with these genetic parts should identify by seqhash, NOT name. The name is only for human readability sake.
6. No I-SceI sites or enzyme expression.

## Some more
- terminators have stop codons (`ATCCTAA` prefix) in the case you want to use a c-taggable protein with a GS on the c terminal. ctag terminators rely on ctags to have stop codons.
- Unlike most MoClo toolkits, we do not place the promoter next to the coding sequence in eukaryotic toolkits. There is space for kozak sequences or ntags, which occupy the same prefix/suffix space.

## Primers
- Dialout primers P1-P96 are reserved for synthesis and assembly usage. Do not use them.
- Dialout primers P97-P144 are reserved for pooling toolkits that are going to be distributed together, so try not to use them.
- Dialout primers 165-161 are used for standardized primers. Do not use them.
- If you need to use primers, it is recommended to use set P150,P151,P152,P153,P154,P155,P156,P158

## Organisms:
- Escherichia coli (Ec)
- Bacillus subtilis (Bs)
- Vibrio natriegens (Vn)
- Saccharomyces cerevisiae (Sc)
