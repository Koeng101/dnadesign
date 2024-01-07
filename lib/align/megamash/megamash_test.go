package megamash

import (
	"testing"

	"github.com/koeng101/dnadesign/lib/random"
)

func TestCompressDNA(t *testing.T) {
	// Define test cases
	longDna, _ := random.DNASequence(300, 0)
	longerDna, _ := random.DNASequence(66000, 0)
	tests := []struct {
		name         string
		dna          string
		expectedLen  int  // Expected length of the compressed data
		expectedFlag byte // Expected flag byte
	}{
		{"Empty", "", 2, 0x00},
		{"Short", "ATGC", 3, 0x00},
		{"Medium", "ATGCGTATGCCGTAGC", 6, 0x00},
		{"Long", longDna, 78, 0x01},
		{"Longest", longerDna, 16505, 0x02},
		// Add more test cases for longer sequences and edge cases
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			compressed := CompressDNA(tc.dna)
			if len(compressed) != tc.expectedLen {
				t.Errorf("CompressDNA() with input %s, expected length %d, got %d", "", tc.expectedLen, len(compressed))
			}
			if compressed[0] != tc.expectedFlag {
				t.Errorf("CompressDNA() with input %s, expected flag %b, got %b", tc.dna, tc.expectedFlag, compressed[0])
			}
		})
	}
}

func TestDecompressDNA(t *testing.T) {
	longDna, _ := random.DNASequence(300, 0)
	longerDna, _ := random.DNASequence(66000, 0)
	// Define test cases
	tests := []struct {
		name     string
		dna      string
		expected string
	}{
		{"Empty", "", ""},
		{"Short", "ATGC", "ATGC"},
		{"Medium", "ATGCGTATGCCGTAGC", "ATGCGTATGCCGTAGC"},
		{"Long", longDna, longDna},
		{"Longest", longerDna, longerDna},
		// Add more test cases as needed
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			compressed := CompressDNA(tc.dna)
			decompressed := DecompressDNA(compressed)
			if decompressed != tc.expected {
				t.Errorf("DecompressDNA() with input %v, expected %s, got %s", compressed, tc.expected, decompressed)
			}
		})
	}
}

func TestMegamash(t *testing.T) {
	oligo1 := "CCGTGCGACAAGATTTCAAGGGTCTCTGTCTCAATGACCAAACCAACGCAAGTCTTAGTTCGTTCAGTCTCTATTTTATTCTTCATCACACTGTTGCACTTGGTTGTTGCAATGAGATTTCCTAGTATTTTCACTGCTGTGCTGAGACCCGGATCGAACTTAGGTAGCCT"
	oligo2 := "CCGTGCGACAAGATTTCAAGGGTCTCTGTGCTATTTGCCGCTAGTTCCGCTCTAGCTGCTCCAGTTAATACTACTACTGAAGATGAATTGGAGGGTGACTTCGATGTTGCTGTTCTGCCTTTTTCCGCTTCTGAGACCCGGATCGAACTTAGGTAGCCACTAGTCATAAT"
	oligo3 := "CCGTGCGACAAGATTTCAAGGGTCTCTCTTCTATCGCAGCCAAGGAAGAAGGTGTATCTCTAGAGAAGCGTCGAGTGAGACCCGGATCGAACTTAGGTAGCCCCCTTCGAAGTGGCTCTGTCTGATCCTCCGCGGATGGCGACACCATCGGACTGAGGATATTGGCCACA"

	samples := []string{"TTTTGTCTACTTCGTTCCGTTGCGTATTGCTAAGGTTAAGACTACTTTCTGCCTTTGCGAGACGGCGCCTCCGTGCGACGAGATTTCAAGGGTCTCTGTGCTATATTGCCGCTAGTTCCGCTCTAGCTGCTCCAGTTAATACTACTACTGAAGATGAATTGGAGGGTGACTTCGATGTTGCTGTTCTGCCTTTTTCCGCTTCTGAGACCCAGATCGACTTTTAGATTCCTCAGGTGCTGTTCTCGCAAAGGCAGAAAGTAGTCTTAACCTTAGCAATACGTGG", "TGTCCTTTACTTCGTTCAGTTACGTATTGCTAAGGTTAAGACTACTTTCTGCCTTTGCGAGAACAGCACCTCTGCTAGGGGCTACTTATCGGGTCTCTAGTTCCGCTCTAGCTGCTCCAGTTAATACTACTACTGAAGATGAATTGGAGGGTGACTTCGATGTTGCTGTTCTGCCTTTTTCCGCTTCTATCTGAGACCGAAGTGGTTTGCCTAAACGCAGGTGCTGTTGGCAAAGGCAGAAAGTAGTCTTAACCTTGACAATGAGTGGTA", "GTTATTGTCGTCTCCTTTGACTCAGCGTATTGCTAAGGTTAAGACTACTTTCTGCCTTTGCGAGAACAGCACCTCTGCTAGGGGCTGCTGGGTCTCTAGTTCCGCTCTAGCTGCTCCAGTTAATACTACTACTGAAGATGAATTGGAGGGTGACTTCGATGTTGCTGTTCTGCCTTTTCCGCTTCTATCTGAGACCGAAGTGGTTAT", "TGTTCTGTACTTCGTTCAGTTACGTATTGCTAAGGTTAAGACTACTTCTGCCTTAGAGACCACGCCTCCGTGCGACAAGATTCAAGGGTCTCTGTGCTCTGCCGCTAGTTCCGCTCTAGCTGCTCCGGTATGCATCTACTACTGAAGATGAATTGGAGGGTGACTTCGATGTTGCTGCTGTTCTGCCTTTTTCCGCTTCTGAGACCCGGATCGAACTTAGGTAGCCAGGTGCTGTTCTCGCAAAGGCAGAAAGTAGTCTTAACCTTAGCAACTGTTGGTT"}
	m := MakeMegamashMap([]string{oligo1, oligo2, oligo3}, 16)
	for _, sample := range samples {
		scores := m.Score(sample)
		if scores[1] < 0.5 {
			t.Errorf("Score for oligo2 should be above 0.5")
		}
	}
}
