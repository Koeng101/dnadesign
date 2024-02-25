package megamash

import (
	"testing"

	"github.com/koeng101/dnadesign/lib/bio/fasta"
)

func TestMegamash(t *testing.T) {
	oligo1 := "CCGTGCGACAAGATTTCAAGGGTCTCTGTCTCAATGACCAAACCAACGCAAGTCTTAGTTCGTTCAGTCTCTATTTTATTCTTCATCACACTGTTGCACTTGGTTGTTGCAATGAGATTTCCTAGTATTTTCACTGCTGTGCTGAGACCCGGATCGAACTTAGGTAGCCT"
	oligo2 := "CCGTGCGACAAGATTTCAAGGGTCTCTGTGCTATTTGCCGCTAGTTCCGCTCTAGCTGCTCCAGTTAATACTACTACTGAAGATGAATTGGAGGGTGACTTCGATGTTGCTGTTCTGCCTTTTTCCGCTTCTGAGACCCGGATCGAACTTAGGTAGCCACTAGTCATAAT"
	oligo3 := "CCGTGCGACAAGATTTCAAGGGTCTCTCTTCTATCGCAGCCAAGGAAGAAGGTGTATCTCTAGAGAAGCGTCGAGTGAGACCCGGATCGAACTTAGGTAGCCCCCTTCGAAGTGGCTCTGTCTGATCCTCCGCGGATGGCGACACCATCGGACTGAGGATATTGGCCACA"

	samples := []string{"TTTTGTCTACTTCGTTCCGTTGCGTATTGCTAAGGTTAAGACTACTTTCTGCCTTTGCGAGACGGCGCCTCCGTGCGACGAGATTTCAAGGGTCTCTGTGCTATATTGCCGCTAGTTCCGCTCTAGCTGCTCCAGTTAATACTACTACTGAAGATGAATTGGAGGGTGACTTCGATGTTGCTGTTCTGCCTTTTTCCGCTTCTGAGACCCAGATCGACTTTTAGATTCCTCAGGTGCTGTTCTCGCAAAGGCAGAAAGTAGTCTTAACCTTAGCAATACGTGG", "TGTCCTTTACTTCGTTCAGTTACGTATTGCTAAGGTTAAGACTACTTTCTGCCTTTGCGAGAACAGCACCTCTGCTAGGGGCTACTTATCGGGTCTCTAGTTCCGCTCTAGCTGCTCCAGTTAATACTACTACTGAAGATGAATTGGAGGGTGACTTCGATGTTGCTGTTCTGCCTTTTTCCGCTTCTATCTGAGACCGAAGTGGTTTGCCTAAACGCAGGTGCTGTTGGCAAAGGCAGAAAGTAGTCTTAACCTTGACAATGAGTGGTA", "GTTATTGTCGTCTCCTTTGACTCAGCGTATTGCTAAGGTTAAGACTACTTTCTGCCTTTGCGAGAACAGCACCTCTGCTAGGGGCTGCTGGGTCTCTAGTTCCGCTCTAGCTGCTCCAGTTAATACTACTACTGAAGATGAATTGGAGGGTGACTTCGATGTTGCTGTTCTGCCTTTTCCGCTTCTATCTGAGACCGAAGTGGTTAT", "TGTTCTGTACTTCGTTCAGTTACGTATTGCTAAGGTTAAGACTACTTCTGCCTTAGAGACCACGCCTCCGTGCGACAAGATTCAAGGGTCTCTGTGCTCTGCCGCTAGTTCCGCTCTAGCTGCTCCGGTATGCATCTACTACTGAAGATGAATTGGAGGGTGACTTCGATGTTGCTGCTGTTCTGCCTTTTTCCGCTTCTGAGACCCGGATCGAACTTAGGTAGCCAGGTGCTGTTCTCGCAAAGGCAGAAAGTAGTCTTAACCTTAGCAACTGTTGGTT"}
	m, err := NewMegamashMap([]fasta.Record{{Sequence: oligo1, Identifier: "oligo1"}, {Sequence: oligo2, Identifier: "oligo2"}, {Sequence: oligo3, Identifier: "oligo3"}}, DefaultKmerSize, DefaultMinimalKmerCount, DefaultScoreThreshold)
	if err != nil {
		t.Errorf("Failed to make NewMegamashMap: %s", err)
	}
	for _, sample := range samples {
		scores := m.Match(sample)
		if scores[0].Identifier != "oligo2" {
			t.Errorf("Should have gotten oligo2. Got: %s", scores[0].Identifier)
		}
	}
}

func BenchmarkMegamash(b *testing.B) {
	for i := 0; i < b.N; i++ {
		oligo1 := "CCGTGCGACAAGATTTCAAGGGTCTCTGTCTCAATGACCAAACCAACGCAAGTCTTAGTTCGTTCAGTCTCTATTTTATTCTTCATCACACTGTTGCACTTGGTTGTTGCAATGAGATTTCCTAGTATTTTCACTGCTGTGCTGAGACCCGGATCGAACTTAGGTAGCCT"
		oligo2 := "CCGTGCGACAAGATTTCAAGGGTCTCTGTGCTATTTGCCGCTAGTTCCGCTCTAGCTGCTCCAGTTAATACTACTACTGAAGATGAATTGGAGGGTGACTTCGATGTTGCTGTTCTGCCTTTTTCCGCTTCTGAGACCCGGATCGAACTTAGGTAGCCACTAGTCATAAT"
		oligo3 := "CCGTGCGACAAGATTTCAAGGGTCTCTCTTCTATCGCAGCCAAGGAAGAAGGTGTATCTCTAGAGAAGCGTCGAGTGAGACCCGGATCGAACTTAGGTAGCCCCCTTCGAAGTGGCTCTGTCTGATCCTCCGCGGATGGCGACACCATCGGACTGAGGATATTGGCCACA"

		samples := []string{"TTTTGTCTACTTCGTTCCGTTGCGTATTGCTAAGGTTAAGACTACTTTCTGCCTTTGCGAGACGGCGCCTCCGTGCGACGAGATTTCAAGGGTCTCTGTGCTATATTGCCGCTAGTTCCGCTCTAGCTGCTCCAGTTAATACTACTACTGAAGATGAATTGGAGGGTGACTTCGATGTTGCTGTTCTGCCTTTTTCCGCTTCTGAGACCCAGATCGACTTTTAGATTCCTCAGGTGCTGTTCTCGCAAAGGCAGAAAGTAGTCTTAACCTTAGCAATACGTGG", "TGTCCTTTACTTCGTTCAGTTACGTATTGCTAAGGTTAAGACTACTTTCTGCCTTTGCGAGAACAGCACCTCTGCTAGGGGCTACTTATCGGGTCTCTAGTTCCGCTCTAGCTGCTCCAGTTAATACTACTACTGAAGATGAATTGGAGGGTGACTTCGATGTTGCTGTTCTGCCTTTTTCCGCTTCTATCTGAGACCGAAGTGGTTTGCCTAAACGCAGGTGCTGTTGGCAAAGGCAGAAAGTAGTCTTAACCTTGACAATGAGTGGTA", "GTTATTGTCGTCTCCTTTGACTCAGCGTATTGCTAAGGTTAAGACTACTTTCTGCCTTTGCGAGAACAGCACCTCTGCTAGGGGCTGCTGGGTCTCTAGTTCCGCTCTAGCTGCTCCAGTTAATACTACTACTGAAGATGAATTGGAGGGTGACTTCGATGTTGCTGTTCTGCCTTTTCCGCTTCTATCTGAGACCGAAGTGGTTAT", "TGTTCTGTACTTCGTTCAGTTACGTATTGCTAAGGTTAAGACTACTTCTGCCTTAGAGACCACGCCTCCGTGCGACAAGATTCAAGGGTCTCTGTGCTCTGCCGCTAGTTCCGCTCTAGCTGCTCCGGTATGCATCTACTACTGAAGATGAATTGGAGGGTGACTTCGATGTTGCTGCTGTTCTGCCTTTTTCCGCTTCTGAGACCCGGATCGAACTTAGGTAGCCAGGTGCTGTTCTCGCAAAGGCAGAAAGTAGTCTTAACCTTAGCAACTGTTGGTT"}
		m, _ := NewMegamashMap([]fasta.Record{{Sequence: oligo1, Identifier: "oligo1"}, {Sequence: oligo2, Identifier: "oligo2"}, {Sequence: oligo3, Identifier: "oligo3"}},
			DefaultKmerSize, DefaultMinimalKmerCount, DefaultScoreThreshold)
		for _, sample := range samples {
			_ = m.Match(sample)
		}
	}
}

func TestMatchesConversion(t *testing.T) {
	// Initial slice of Match structs
	matches := []Match{
		{"match1", 90.1},
		{"match2", 85.5},
	}
	// Convert matches to JSON string
	jsonStr, err := MatchesToJSON(matches)
	if err != nil {
		t.Fatalf("MatchesToJSON failed with error: %v", err)
	}

	// Convert JSON string back to slice of Match structs
	convertedMatches, err := JSONToMatches(jsonStr)
	if err != nil {
		t.Fatalf("JSONToMatches failed with error: %v", err)
	}

	// Convert the convertedMatches back to JSON to compare strings
	convertedJSONStr, err := MatchesToJSON(convertedMatches)
	if err != nil {
		t.Fatalf("MatchesToJSON failed with error: %v", err)
	}

	// Compare the original JSON string with the converted JSON string
	if jsonStr != convertedJSONStr {
		t.Errorf("Conversion mismatch. Original JSON: %v, After Conversion: %v", jsonStr, convertedJSONStr)
	}
}
