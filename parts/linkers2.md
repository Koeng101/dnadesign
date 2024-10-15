# DnaDesign Assembly

DnaDesign Assembly (shortened as dd assembly) is a GoldenGate DNA assembly method similar to MoClo. The overhang set was changed from MoClo in order to accomodate the overhangs `GTCT` and `CGAG`, which are used in recursive DNA assemblies. The BsaI overhangs were optimized using [empirical data](https://doi.org/10.1371/journal.pone.0238592) from the paper "Enabling one-pot Golden Gate assemblies of unprecedented complexity using data-optimized assembly design".

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
17. TACA
18. TTAC
19. TTGA
20. TGGA
21. GAAG
22. GACC
23. GCCG
24. AAAA
25. GGGG

## Linkers
The most basic dd assembly will create a transcriptional unit. Oftentimes, however, you will want to be able to combine different transcriptional units together. In order to do these multi-level assemblies, we use linkers. Linkers are genetic parts that are used within an assembly reaction that define the part overhangs of the assembled construct, for use in the next assembly reaction. For example, we could have 3 genetic parts:

* Promoter+RBS
* GFP
* Terminator

While we could construct a simple transcriptional unit of `["Promoter+RBS", "GFP", "Terminator"]`, we might want to use this whole transcriptional unit in a different construct. In order to do that, we will add 2 linkers. The prefix linker (A)
