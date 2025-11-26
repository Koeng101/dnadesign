# genolib

This is a library of genetic parts in plasmids from the [GenoLIB](https://doi.org/10.1093/nar/gkv272) paper. In particular, these were derived from plannotate's BLAST database. Instead of using BLAST, we use mash containment to filter most the sequences, then go back and clean up with a smith-waterman alignment.
