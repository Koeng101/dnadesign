package fragment_test

import (
	_ "embed"
	"strings"
	"testing"

	"github.com/koeng101/dnadesign/lib/bio"
	"github.com/koeng101/dnadesign/lib/synthesis/fragment"
)

//go:embed data/blue1.fasta
var blue1 string

func TestFragment(t *testing.T) {
	gene := "atgaaaaaatttaactggaagaaaatagtcgcgccaattgcaatgctaattattggcttactaggtggtttacttggtgcctttatcctactaacagcagccggggtatcttttaccaatacaacagatactggagtaaaaacggctaagaccgtctacaccaatataacagatacaactaaggctgttaagaaagtacaaaatgccgttgtttctgtcatcaattatcaagaaggttcatcttcagattctctaaatgacctttatggccgtatctttggcggaggggacagttctgattctagccaagaaaattcaaaagattcagatggtctacaggtcgctggtgaaggttctggagtcatctataaaaaagatggcaaagaagcctacatcgtaaccaataaccatgttgtcgatggggctaaaaaacttgaaatcatgctttcggatggttcgaaaattactggtgaacttgttggtaaagacacttactctgacctagcagttgtcaaagtatcttcagataaaataacaactgttgcagaatttgcagactcaaactcccttactgttggtgaaaaagcaattgctatcggtagcccacttggtaccgaatacgccaactcagtaacagaaggaatcgtttctagccttagccgtactataacgatgcaaaacgataatggtgaaactgtatcaacaaacgctatccaaacagatgcagccattaaccctggtaactctggtggtgccctagtcaatattgaaggacaagttatcggtattaattcaagtaaaatttcatcaacgtctgcagtcgctggtagtgctgttgaaggtatggggtttgccattccatcaaacgatgttgttgaaatcatcaatcaattagaaaaagatggtaaagttacacgaccagcactaggaatctcaatagcagatcttaatagcctttctagcagcgcaacttctaaattagatttaccagatgaggtcaaatccggtgttgttgtcggtagtgttcagaaaggtatgccagctgacggtaaacttcaagaatatgatgttatcactgagattgatggtaagaaaatcagctcaaaaactgatattcaaaccaatctttacagccatagtatcggagatactatcaaggtaaccttctatcgtggtaaagataagaaaactgtagatcttaaattaacaaaatctacagaagacatatctgattaa"

	_, _, err := fragment.Fragment(gene, 90, 110, []string{})
	if err != nil {
		t.Error(err.Error())
	}
}

func TestUnfragmentable(t *testing.T) {
	// One should not be able to fragment this
	polyA := "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	_, _, err := fragment.Fragment(polyA, 40, 80, []string{})
	if err == nil {
		t.Errorf("polyA should fail to fragment")
	}
}

func TestFragmentSizes(t *testing.T) {
	// This tests if minSize > maxSize
	lacZ := "ATGACCATGATTACGCCAAGCTTGCATGCCTGCAGGTCGACTCTAGAGGATCCCCGGGTACCGAGCTCGAATTCACTGGCCGTCGTTTTACAACGTCGTGACTGGGAAAACCCTGGCGTTACCCAACTTAATCGCCTTGCAGCACATCCCCCTTTCGCCAGCTGGCGTAATAGCGAAGAGGCCCGCACCGATCGCCCTTCCCAACAGTTGCGCAGCCTGAATGGCGAATGGCGCCTGATGCGGTATTTTCTCCTTACGCATCTGTGCGGTATTTCACACCGCATATGGTGCACTCTCAGTACAATCTGCTCTGATGCCGCATAG"
	_, _, err := fragment.Fragment(lacZ, 105, 95, []string{})
	if err == nil {
		t.Errorf("Fragment should fail when minFragmentSize > maxFragmentSize")
	}

	_, _, err = fragment.Fragment(lacZ, 7, 95, []string{})
	if err == nil {
		t.Errorf("Fragment should fail when minFragmentSize < 8")
	}
}

func TestSmallFragmentSize(t *testing.T) {
	// The following should succeed, but will require setting minFragmentSize = 12
	lacZ := "ATGACCATGATTACGCCAAGCTTGCATGCCTGCAGGTCGACTCTAGAGGATCCCCGGGTACCGAGCTCGAATTCACTGGCCGTCGTTTTACAACGTCGTGACTGGGAAAACCCTGGCGTTACCCAACTTAATCGCCTTGCAGCACATCCCCCTTTCGCCAGCTGGCGTAATAGCGAAGAGGCCCGCACCGATCGCCCTTCCCAACAGTTGCGCAGCCTGAATGGCGAATGGCGCCTGATGCGGTATTTTCTCCTTACGCATCTGTGCGGTATTTCACACCGCATATGGTGCACTCTCAGTACAATCTGCTCTGATGCCGCATAG"
	_, _, err := fragment.Fragment(lacZ, 12, 30, []string{})
	if err != nil {
		t.Errorf("Got error in small fragmentation: %s", err)
	}
}

func TestLongFragment(t *testing.T) {
	// A regression test for a bug that sometimes fragmented a sequence to
	// be longer than its max length
	gene := "GGAGGGTCTCAATGCTGGACGATCGCAAATTCAGCGAACAGGAGCTGGTCCGTCGCAACAAATACAAAACGCTGGTCGAGCAAAACAAAGACCCGTACAAGATTACGAACTGGAAACGCAATACCACCCTGCTGAAACTGAATGAGAAATACAAAGACTATAGCAAGGAGGACCTGTTGAACCTGAATCAAGAACTGGTCGTTGTTGCAGGTCGTATCAAACTGTATCGTGAAGCCGGTAAAAAAGCTGCCTTTGTGAACATTGATGATCAAGACTCCTCTATTCAGTTGTACGTGCGCCTGGATGAGATCGGTGATCAGAGCTTCGAGGATTTCCGCAATTTCGACCTGGGTGACATCATTGGTGTTAAAGGTATCATGATGCGCACCGACCACGGCGAGTTGAGCATCCGTTGTAAGGAAGTCGTGCTGCTGAGCAAGGCCCTGCGTCCGCTGCCGGATAAACACGCGGGCATTCAGGATATTGAGGAAAAGTACCGCCGTCGCTATGTGGACCTGATTATGAATCACGACGTGCGCAAGACGTTCCAGGCGCGTACCAAGATCATTCGTACCTTGCAAAACTTTCTGGATAATAAGGGTTACATGGAGGTCGAAACCCCGATCCTGCATAGCCTGAAGGGTGGCGCGAGCGCGAAACCGTTTATTACCCACTACAATGTGCTGAATACGGATGTGTATCTGCGTATCGCGACCGAGCTGCACCTGAAACGCCTGATTGTTGGCGGTTTCGAGGGTGTGTATGAGATCGGTCGCATCTTTCGCAATGAAGGTATGTCCACGCGTCACAATCCGGAATTCACGTCTATCGAACTGTATGTCGCCTATGAGGACATGTTCTTTTTGATGGATCTGACCGAAGAGATTTTTCGCGTTTGTAATGCCGCAGTCAACAGCTCCAGCATCATTGAGTATAACAACGTGAAAATTGACCTGAGCAAGCCGTTTAAGCGCCTGCATATGGTTGACGGTATTAAACAGGTGACCGGCGTCGACTTCTGGCAGGAGATGACGGTCCAACAGGCTCTGGAGCTGGCCAAAAAGCATAAAGTGCACGTTGAAAAACATCAAGAGTCTGTTGGTCACATTATCAATTTGTTCTATGAGGAGTTCGTGGAGTCCACGATTGTTGAGCCGACGTTCGTGTACGGTCACCCGAAGGAAATCTCTCCGCTGGCTAAGAGCAATCCGTCTGACCCGCGTTTCACGGACCGTTTCGAGCTGTTCATTCTGGGTCGTGAGTATGCGAATGCGTTTAGCGAGCTGAATGACCCGATTGACCAGTACGAACGCTTCAAGGCTCAGATTGAGGAGGAAAGCAAGGGCAACGATGAAGCCAACGACATGGACATTGATTTCATCGAGGCTCTGGAACACGCCATGCCGCCGACCGCGGGTATTGGTATCGGCATTGATCGCTTGGTTATGCTGCTGACGAATAGCGAATCCATCAAAGACGTGCTGTTGTTCCCGCAAATGAAGCCGCGCGAATGAAGAGCTTAGAGACCCGCT"
	frags, _, err := fragment.Fragment(gene, 79, 94, []string{})
	if err != nil {
		t.Error(err.Error())
	}
	for _, frag := range frags {
		if len(frag) > 94 {
			t.Errorf("Fragment too long. Expected <94, got %d", len(frag))
		}
	}
}

func TestCheckLongRegresssion(t *testing.T) {
	// This makes sure that reverse complements are simply skipped when
	// checking efficiency of overhangs
	overhangs := []string{"AGAC"}
	newOverhangs, _ := fragment.NextOverhangs(overhangs)
	foundGTCT := false
	for _, overhang := range newOverhangs {
		if overhang == "GTCT" {
			foundGTCT = true
		}
	}
	if foundGTCT {
		t.Errorf("Should not have found GTCT since it is the reverse complement of AGAC")
	}
}

func TestRegressionTestMatching12(t *testing.T) {
	// This ensures that overhangs approximately match the efficiency generated
	// by NEB with their ligase fidelity viewer: https://ligasefidelity.neb.com/viewset/run.cgi
	overhangs := []string{"CGAG", "GTCT", "TACT", "AATG", "ATCC", "CGCT", "AAAA", "AAGT", "ATAG", "ATTA", "ACAA", "ACGC", "TATC", "TAGA", "TTAC", "TTCA", "TGTG", "TCGG", "TCCC", "GAAG", "GTGC", "GCCG", "CAGG", "TACG"}
	efficiency := fragment.SetEfficiency(overhangs)
	if (efficiency > 1) || (efficiency < 0.965) {
		t.Errorf("Expected efficiency of .99 - approximately matches NEB ligase fidelity viewer of .97. Got: %g", efficiency)
	}
}

func TestFragmentWithOverhangs(t *testing.T) {
	defaultOverhangs := []string{"CGAG", "GTCT", "GGGG", "AAAA", "AACT", "AATG", "ATCC", "CGCT", "TTCT", "AAGC", "ATAG", "ATTA", "ATGT", "ACTC", "ACGA", "TATC", "TAGG", "TACA", "TTAC", "TTGA", "TGGA", "GAAG", "GACC", "GCCG", "TCTG", "GTTG", "GTGC", "TGCC", "CTGG", "TAAA", "TGAG", "AAGA", "AGGT", "TTCG", "ACTA", "TTAG", "TCTC", "TCGG", "ATAA", "ATCA", "TTGC", "CACG", "AATA", "ACAA", "ATGG", "TATG", "AAAT", "TCAC"}
	gene := "atgaaaaaatttaactggaagaaaatagtcgcgccaattgcaatgctaattattggcttactaggtggtttacttggtgcctttatcctactaacagcagccggggtatcttttaccaatacaacagatactggagtaaaaacggctaagaccgtctacaccaatataacagatacaactaaggctgttaagaaagtacaaaatgccgttgtttctgtcatcaattatcaagaaggttcatcttcagattctctaaatgacctttatggccgtatctttggcggaggggacagttctgattctagccaagaaaattcaaaagattcagatggtctacaggtcgctggtgaaggttctggagtcatctataaaaaagatggcaaagaagcctacatcgtaaccaataaccatgttgtcgatggggctaaaaaacttgaaatcatgctttcggatggttcgaaaattactggtgaacttgttggtaaagacacttactctgacctagcagttgtcaaagtatcttcagataaaataacaactgttgcagaatttgcagactcaaactcccttactgttggtgaaaaagcaattgctatcggtagcccacttggtaccgaatacgccaactcagtaacagaaggaatcgtttctagccttagccgtactataacgatgcaaaacgataatggtgaaactgtatcaacaaacgctatccaaacagatgcagccattaaccctggtaactctggtggtgccctagtcaatattgaaggacaagttatcggtattaattcaagtaaaatttcatcaacgtctgcagtcgctggtagtgctgttgaaggtatggggtttgccattccatcaaacgatgttgttgaaatcatcaatcaattagaaaaagatggtaaagttacacgaccagcactaggaatctcaatagcagatcttaatagcctttctagcagcgcaacttctaaattagatttaccagatgaggtcaaatccggtgttgttgtcggtagtgttcagaaaggtatgccagctgacggtaaacttcaagaatatgatgttatcactgagattgatggtaagaaaatcagctcaaaaactgatattcaaaccaatctttacagccatagtatcggagatactatcaaggtaaccttctatcgtggtaaagataagaaaactgtagatcttaaattaacaaaatctacagaagacatatctgattaa"

	_, _, err := fragment.FragmentWithOverhangs(gene, 90, 110, []string{}, defaultOverhangs)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestRecursiveFragment(t *testing.T) {
	records, _ := bio.NewFastaParser(strings.NewReader(blue1)).Parse()
	// These are the 46 possible overhangs I personally use, plus the two identity overhangs CGAG+GTCT
	defaultOverhangs := []string{"GGGG", "AAAA", "AACT", "AATG", "ATCC", "CGCT", "TTCT", "AAGC", "ATAG", "ATTA", "ATGT", "ACTC", "ACGA", "TATC", "TAGG", "TACA", "TTAC", "TTGA", "TGGA", "GAAG", "GACC", "GCCG", "TCTG", "GTTG", "GTGC", "TGCC", "CTGG", "TAAA", "TGAG", "AAGA", "AGGT", "TTCG", "ACTA", "TTAG", "TCTC", "TCGG", "ATAA", "ATCA", "TTGC", "CACG", "AATA", "ACAA", "ATGG", "TATG", "AAAT", "TCAC"}
	excludeOverhangs := []string{"CGAG", "GTCT"} // These are the recursive BsaI definitions, and must be excluded from all builds.
	gene := records[0].Sequence
	maxOligoLen := 174                   // for Agilent oligo pools
	assemblyPattern := []int{5, 4, 4, 5} // seems reasonable enough
	_, err := fragment.RecursiveFragment(gene, maxOligoLen, assemblyPattern, excludeOverhangs, defaultOverhangs)
	if err != nil {
		t.Errorf("Failed to RecursiveFragment blue1. Got error: %s", err)
	}
}
