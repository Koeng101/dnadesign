package clone

import (
	_ "embed"
	"strings"
	"testing"

	"github.com/koeng101/dnadesign/lib/bio"
	"github.com/koeng101/dnadesign/lib/transform"
)

//go:embed data/ligations.fastq
var ligationReads string

func TestKmer(t *testing.T) {
	parts := []Part{Part{Sequence: transform.ReverseComplement("CACTCGATAGGTACAACCGGGGTCTCTGTCTCAGCACCGGCAGCAGCAGCACGCGCAGGTGGTCCGGACATCTCGGAGGCAGGTGCACTGGGTACGCGCGCACTGGCACGCGCATGGCTGGCATACGGTATCACGGCAGAGGACGTGAAGCGCATCCTGGACACGCTGGGTCTGGGTTCGGGTGAGTTCGGTTTCCGCCAGCAGGTGCTGGAGGACACGATCCAGCCGGCAATCGCATCGGGTGCAGGTCTGCACTTCGTGGACCTGAGACCGCCATCCTCTTATCTCGTGGCATTGAGT"), Circular: false}, Part{Sequence: "CACTCGATAGGTACAACCGGGGTCTCTGACCCGGAGGCAGCAGCAGGTTACCTGGCAGAGCTGGTGAAGCGCTCGTACTCGAAGGCAGTGCCGGCAGCAGTGAAGGCAGAGGACTGGCTGAAGAAGATCGCACTGCTGGTGGCATCGGAGTCGACGTCGTCGTCGGGTCTGAAGATCCTGAAGTCGCCGCAGCCGGTGAAGTGGGTGTCGCCGGACGGTTTCCCGGTGATCCTGGACCCGAAGAAGCCGATCCAGATGAGACCGCCATCCTCTTATCTCGTGGCCCCGAGCATCAGCGGA", Circular: false}, Part{Sequence: "CACTCGATAGGTACAACCGGGGTCTCTCAGACGCGCCTGCAGCTGATGTTCCTGGGTCAGTTCACGCTGCAGCCGACGATCAACACGAACAAGGACTCGGAGATCGACAAGGAGAAGCTGGCATCGGGTATCGCACCGAACTTCGTGGCATCGCAGATCGCATCGCTGGTGCGCCGCACGGTGGTGCTGGCACACGAGAAGTACGGTATCAAGTCGTTCTGGATCGACAAGGACGAGTACGGTACGATCCCGGCACAGATGGACAATGAGACCGCCATCCTCTTATCTCGTGGTCTCAGC", Circular: false}, Part{Sequence: "CACTCGATAGGTACAACCGGGGTCTCTACAAGCTGAAGGCAGCAATCAAGGAGTCGCTGGTGGAGACGTACACGAACAACGACCTGCTGGAGAACCTGCGCGACCTGGTGGAGAAGGAGGGTTGCCCGGGTGACTGCTCGTCGGTGCCGGAGCTGCCGCCGAAGGGTGACCTGGACCTGAACGAGATCCTGAAGTCGGAGTACGGTGGTTTCTAAATCCCGAGTGAGACCGCCATCCTCTTATCTCGTGGGGGTTAATCTAGTCAGCGCCACGACTACTTCCGCTTCCCCACATAAGCAG", Circular: false}}
	pOpenV3Methylated := Part{"AGGGTAATGGTCTCTCGAGACcAAGTCGTCATAGCTGTTTCCTGAGAGCTTGGCAGGTGATGACACACATTAACAAATTTCGTGAGGAGTCTCCAGAAGAATGCCATTAATTTCCATAGGCTCCGCCCCCCTGACGAGCATCACAAAAATCGACGCTCAAGTCAGAGGTGGCGAAACCCGACAGGACTATAAAGATACCAGGCGTTTCCCCCTGGAAGCTCCCTCGTGCGCTCTCCTGTTCCGACCCTGCCGCTTACCGGATACCTGTCCGCCTTTCTCCCTTCGGGAAGCGTGGCGCTTTCTCATAGCTCACGCTGTAGGTATCTCAGTTCGGTGTAGGTCGTTCGCTCCAAGCTGGGCTGTGTGCACGAACCCCCCGTTCAGCCCGACCGCTGCGCCTTATCCGGTAACTATCGTCTTGAGTCCAACCCGGTAAGACACGACTTATCGCCACTGGCAGCAGCCACTGGTAACAGGATTAGCAGAGCGAGGTATGTAGGCGGTGCTACAGAGTTCTTGAAGTGGTGGCCTAACTACGGCTACACTAGAAGAACAGTATTTGGTATCTGCGCTCTGCTGAAGCCAGTTACCTTCGGAAAAAGAGTTGGTAGCTCTTGATCCGGCAAACAAACCACCGCTGGTAGCGGTGGTTTTTTTGTTTGCAAGCAGCAGATTACGCGCAGAAAAAAAGGATCTCAAGAAGGCCTACTATTAGCAACAACGATCCTTTGATCTTTTCTACGGGGTCTGACGCTCAGTGGAACGAAAACTCACGTTAAGGGATTTTGGTCATGAGATTATCAAAAAGGATCTTCACCTAGATCCTTTTAAATTAAAAATGAAGTTTTAAATCAATCTAAAGTATATATGAGTAAACTTGGTCTGACAGTTACCAATGCTTAATCAGTGAGGCACCTATCTCAGCGATCTGTCTATTTCGTTCATCCATAGTTGCCTGACTCCCCGTCGTGTAGATAACTACGATACGGGAGGGCTTACCATCTGGCCCCAGTGCTGCAATGATACCGCGAGAACCACGCTCACCGGCTCCAGATTTATCAGCAATAAACCAGCCAGCCGGAAGGGCCGAGCGCAGAAGTGGTCCTGCAACTTTATCCGCCTCCATCCAGTCTATTAATTGTTGCCGGGAAGCTAGAGTAAGTAGTTCGCCAGTTAATAGTTTGCGCAACGTTGTTGCCATTGCTACAGGCATCGTGGTGTCACGCTCGTCGTTTGGTATGGCTTCATTCAGCTCCGGTTCCCAACGATCAAGGCGAGTTACATGATCCCCCATGTTGTGCAAAAAAGCGGTTAGCTCCTTCGGTCCTCCGATCGTTGTCAGAAGTAAGTTGGCCGCAGTGTTATCACTCATGGTTATGGCAGCACTGCATAATTCTCTTACTGTCATGCCATCCGTAAGATGCTTTTCTGTGACTGGTGAGTACTCAACCAAGTCATTCTGAGAATAGTGTATGCGGCGACCGAGTTGCTCTTGCCCGGCGTCAATACGGGATAATACCGCGCCACATAGCAGAACTTTAAAAGTGCTCATCATTGGAAAACGTTCTTCGGGGCGAAAACTCTCAAGGATCTTACCGCTGTTGAGATCCAGTTCGATGTAACCCACTCGTGCACCCAACTGATCTTCAGCATCTTTTACTTTCACCAGCGTTTCTGGGTGAGCAAAAACAGGAAGGCAAAATGCCGCAAAAAAGGGAATAAGGGCGACACGGAAATGTTGAATACTCATACTCTTCCTTTTTCAATATTATTGAAGCATTTATCAGGGTTATTGTCTCATGAGCGGATACATATTTGAATGTATTTAGAAAAATAAACAAATAGGGGTTCCGCGCACCTGCACCAGTCAGTAAAACGACGGCCAGTGACTTgGTCTCGAGACCTAGGGATA", false}
	parts = append(parts, pOpenV3Methylated)
	var fragments []Fragment
	for _, part := range parts {
		newFragments := CutWithEnzyme(part, true, DefaultEnzymes["BsaI"], true)
		fragments = append(fragments, newFragments...)
	}

	// First, ligate the fragments
	ligation, ligationPattern, err := Ligate(fragments, true)
	if err != nil {
		t.Errorf("Failed to ligate: %s", err)
	}

	// Now, find kmer overlaps
	kmerOverlaps, err := FindKmerOverlaps(fragments, ligation, ligationPattern, 16)
	if err != nil {
		t.Errorf("Failed to find kmer overlaps: %s", err)
	}
	expectedOverlaps := []string{"GACTTGGTCTCAGCAC", "AAATCCCGAGACCAAG", "AGATGGACAAGCTGAA", "CCGATCCAGACGCGCC", "CCTCCGGGTCCACGAA"}
	for i := range kmerOverlaps {
		if kmerOverlaps[i].Kmer != expectedOverlaps[i] {
			t.Errorf("Expected %s on kmerOverlap %d, Got: %s", expectedOverlaps[i], i, kmerOverlaps[i].Kmer)
		}
	}

	// Now, find kmers in example sequences
	// These are from real reads from a real ligation
	//
	// Interestingly enough, read #2 (idx 1) is an example of where this fails.
	// It definitely has a ligation event with kmer "CCTCCGGGTCCACGAA" that
	// is mutated by a single base pair, so is not detected. It is actually
	// a triple ligation, in that it ligated to part of a mal-PCRed vector
	// in addition to the two fragments, and that ligation event is sequence
	// correct.
	parser := bio.NewFastqParser(strings.NewReader(ligationReads))
	reads, _ := parser.Parse()
	expectedOverlapResults := []string{"CCTCCGGGTCCACGAA", "GACTTGGTCTCAGCAC", "CCTCCGGGTCCACGAA"}
	for i, read := range reads {
		overlaps := FindKmers(kmerOverlaps, read)
		if overlaps[0].Kmer != expectedOverlapResults[i] {
			t.Errorf("Failed to find correct overlap in read %d. Expected: %s Got: %s", i, expectedOverlapResults[i], overlaps[0].Kmer)
		}
	}
}

func TestKmerFailure(t *testing.T) {
	// Tests failure cases
	_, err := FindKmerOverlaps([]Fragment{}, "", []int{}, 0)
	if err == nil {
		t.Errorf("Expected: need at least two fragments to find overlaps")
	}
	_, err = FindKmerOverlaps([]Fragment{Fragment{}, Fragment{}}, "", []int{}, 0)
	if err == nil {
		t.Errorf("Expected: need at least a kmer of 12")
	}
}
