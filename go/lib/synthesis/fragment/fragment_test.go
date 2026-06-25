package fragment_test

import (
	_ "embed"
	"reflect"
	"strings"
	"testing"

	"github.com/koeng101/dnadesign/lib/bio"
	"github.com/koeng101/dnadesign/lib/synthesis/fragment"
	"github.com/koeng101/dnadesign/lib/transform"
)

// breakPlasmid is a plasmid where a kanR marker spans the window [1962, 2485]
// (see the FASTA header fragment_between_1962_and_2485). It is shared by the
// FragmentWithBreak tests and example.
const breakPlasmid = "TTTCGGGAGCGGATTATACACAAGCTTctagcataaccccttggggcctctaaacgggtcttgaggggttttttgGTCGTGGTTTGTCTGGTCAACCACCGCGGGCTCAGTGGTGTACGGTACAAACCCCGACTTTTAAATTTatgacacacattaacaaatttcgtgaggagtctccagaagaatgccattaaCGGGTATGCTTGCCTACTAATTAGGGATAACAGGGTAATCTCTCTTAAGGTAGCTTGAGGAGGTTTCTCTGTAAATAATAACTATAACGGTCCTAAGGTAGCGAggattttaacattttgcgttgttccaaaagttatcaacagcctagaacgtcataggaagcgattacagacactttatagctatcagcatgggaacataagggcaggatgaaatatgggtcttaaaacgcaaatggtgaggttttagaggtattttttgaagatgattaaggcggtttgtttttaaaatttttggcggctctcaggctgcttacattttaaccagttcagtgaaaagttctttttcagcaaatttctgtttagcaccatagctaaaacttgcgtggaacatattaagaattgaccgaaatgacacaatctcaattatattttttttgaaaagttttctttatcaaatattttaaatcattgatttatatataagtatgcattcattttaataattaatctttatttaacaatgatttatctatattcaattgtttaattattcttactaatattatctctatatcaatattttttatttaaaaacatatgtttagtagtgcttttgattaaagtaccagagggagggagcagagctgaatgggaaatactcaccctagagcgattcttaaaaatcaccctaaagtattcccattcgatgtaccgtcggtcggtcgctttcgcatcagggatgacatcactgtatcaagctgccactgttatgattacgattgatagcaccgcctgaacacgctcataaccgccaattaaatgactactcattgcgtccgctactctgttcagttcccctagtaatagcgtttttccgatgtgtgctagcgtcactgtacctcatcacccacacatggacagattgagttacagacattggctaaatttttggggtctatgcttgacaaagcgagttaaaaccttagaatttaagacaggtacattaagcccctgtggtgaaatcattagcggtcgtcaaaccttaatgctttcgattgccgatatgtcagtatcgaagatctgtaccccataaattacggtaaagccccaagcaattgcaaggggcttttatcttttttaacaaaaaaaatttataaatcaggattttataacactaaataatccaaagtacatgagtaagtcatgaccactcctgcgatgtgtgtagcactctgagtatccgtattatatcagtgatgtcatatacaaccacataatgcggatgtatcactaactccctagtatctttctgtctgcctgtgcgccccatcagtcgattatcaacaagtaaatctgtcttttcttcaattaaatcatcaatttcaatagctgctatagggttgcgttcttcaagataatcatagatatttctacgatcttggcTCTGCCCGGGattctcaccaataaaaaacgcccggcggcaaccgagcgttctgaacaaatccagatggagttctgaggtcattactggatctatcaacgggagtccaagcgagctcggtactaaaacaattcatccagtaaaatataatattttattttctcccaatcaggcttgatccccagtaagtcaaaaaatagctcgacatactgttcttccccgatatcctccctgatcgaccggacgcagaaggcaatgtcataccacttgtccgccctgccgcttctcccaagatcaataaagccacttactttgccatctttcacaaagatgttgctgtctcccaggtcgccgtgggaaaagacaagttcctcttcgggcttttccgtctttaaaaaatcatacagctcgcgcggatctttaaatggagtatcttcttcccagttttcgcaatccacatcggccagatcgttattcagtaagtaatccaattcggctaagcggccgtctaagctattcgtatagggacaatccgatatgtcgatggagtgaaagagcctgatgcactccgcatacagctcgatagtcttttcagggctttgttcatcttcatacccttccgagcaaaggacgccatcggcctcactcatgagcagattgctccagccatcatgccgttcaaagtgcaggacctttggaacaggcagctttccttccagccatagcatcatgtccttttcccgttccacatcataggtggtccctttataccggctgtccgtcatttttaaatataggatttcattttctcccaccagcttatataccttagcaggagacattccttccgtatcttttacgcagcggtattcttcgatcagttttttcaattccggtgatattctcattttagccatttattatttccttcctcttttctacagtatttaaagataccccaagaagctaattataacaagacgaactccaattcactgttccttgcattctaaaaccttaaatacagaaaacagccttttcaaagttgttttcaaagttggcgtataacatagtatcgacggagccgattttgaaaccacaattatgatagaatttgacgtccttttccgctgcataaccctgcttcggggtcattatagcgattttttcggtatatccatcctttttcgcacgatatacaggattttgccaaagggttcgtgtagactttccttggtgtatccaacggcgtcagccgggcaggataggtgaagtaggcccacccgcgagcgggtgttccttcttcactgtcccttattcgcacctggcggtgctcaacgggaatcctgctctgcgaggctggccgtaTTGACAGACAATCCGTAGGCACAATTTTCGAAAAAACCCGCTTCGGCGGGTTTTTTTATAGCTAAAAATGTTCCAGCGCTGGCACGCAACCTCTCATGCGCTACTTATCACGCCGCGCCAATTTATTACCGCTATGGCCAATTGATCGGCCGGCTTGTCGACGACGGCGGACTCCGTCGTCAGGATCATCCGGGCGAATTCCGTGTTATCCAGTCCCAGAA"

// reassembleFragments concatenates a fragment list, dropping the 4bp overhang
// each consecutive fragment shares with the previous one.
func reassembleFragments(frags []string) string {
	seq := frags[0]
	for i := 1; i < len(frags); i++ {
		seq += frags[i][4:]
	}
	return seq
}

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
	_, err := fragment.RecursiveFragment(gene, maxOligoLen, assemblyPattern, excludeOverhangs, defaultOverhangs, "GTCTCT", "CGAG")
	if err != nil {
		t.Errorf("Failed to RecursiveFragment blue1. Got error: %s", err)
	}
}

func TestRecursiveFragmentPy(t *testing.T) {
	// These are the 46 possible overhangs I personally use, plus the two identity overhangs CGAG+GTCT
	defaultOverhangs := []string{"GGGG", "AAAA", "AACT", "AATG", "ATCC", "CGCT", "TTCT", "AAGC", "ATAG", "ATTA", "ATGT", "ACTC", "ACGA", "TATC", "TAGG", "TACA", "TTAC", "TTGA", "TGGA", "GAAG", "GACC", "GCCG", "TCTG", "GTTG", "GTGC", "TGCC", "CTGG", "TAAA", "TGAG", "AAGA", "AGGT", "TTCG", "ACTA", "TTAG", "TCTC", "TCGG", "ATAA", "ATCA", "TTGC", "CACG", "AATA", "ACAA", "ATGG", "TATG", "AAAT", "TCAC"}
	excludeOverhangs := []string{"CGAG", "GTCT"} // These are the recursive BsaI definitions, and must be excluded from all builds.
	gene := "ATGACCATGATTACGCCAAGCTTGCATGCCTGCAGGTCGACTCTAGAGGATCCCCGGGTACCGAGCTCGAATTCACTGGCCGTCGTTTTACAACGTCGTGACTGGGAAAACCCTGGCGTTACCCAACTTAATCGCCTTGCAGCACATCCCCCTTTCGCCAGCTGGCGTAATAGCGAAGAGGCCCGCACCGATCGCCCTTCCCAACAGTTGCGCAGCCTGAATGGCGAATGGCGCCTGATGCGGTATTTTCTCCTTACGCATCTGTGCGGTATTTCACACCGCATATGGTGCACTCTCAGTACAATCTGCTCTGATGCCGCATAG"
	maxOligoLen := 174                   // for Agilent oligo pools
	assemblyPattern := []int{5, 4, 4, 5} // seems reasonable enough
	result, err := fragment.RecursiveFragment(gene, maxOligoLen, assemblyPattern, excludeOverhangs, defaultOverhangs, "GTCTCT", "CGAG")
	if err != nil {
		t.Errorf("Failed to RecursiveFragment blue1. Got error: %s", err)
	}

	// Add more specific assertions based on the expected structure of the result
	expectedFragments := []string{
		"ATGACCATGATTACGCCAAGCTTGCATGCCTGCAGGTCGACTCTAGAGGATCCCCGGGTACCGAGCTCGAATTCACTGGCCGTCGTTTTACAACGTCGTGACTGGGAAAACCCTGGCGTTACCCAACTTAATCGCCTTGCAGCACATCCCCCTTTCGCCAG",
		"CCAGCTGGCGTAATAGCGAAGAGGCCCGCACCGATCGCCCTTCCCAACAGTTGCGCAGCCTGAATGGCGAATGGCGCCTGATGCGGTATTTTCTCCTTACGCATCTGTGCGGTATTTCACACCGCATATGGTGCACTCTCAGTACAATCTGCTCTGATGCCGCATAG",
	}

	if !reflect.DeepEqual(result.Fragments, expectedFragments) {
		t.Errorf("Unexpected fragments. Got %v, want %v", result.Fragments, expectedFragments)
	}
}

func TestFragmentWithBreak(t *testing.T) {
	breakStart := 1962
	breakEnd := 2485
	before, after, efficiency, err := fragment.FragmentWithBreak(breakPlasmid, breakStart, breakEnd, 800, 1000, []string{})
	if err != nil {
		t.Fatal(err.Error())
	}
	if len(before) == 0 || len(after) == 0 {
		t.Fatalf("Expected non-empty halves, got before=%d after=%d", len(before), len(after))
	}

	// The break overhang is the junction shared by the last before-fragment and
	// the first after-fragment.
	lastBefore := before[len(before)-1]
	breakOverhang := lastBefore[len(lastBefore)-4:]
	if breakOverhang != after[0][:4] {
		t.Errorf("Break overhang mismatch: before ends %s, after starts %s", breakOverhang, after[0][:4])
	}

	// That junction must sit within the requested window. The before region is
	// plasmid[:breakPosition], so its length is the break position.
	breakPosition := len(reassembleFragments(before))
	if breakPosition < breakStart || breakPosition > breakEnd {
		t.Errorf("Break position %d outside window [%d, %d]", breakPosition, breakStart, breakEnd)
	}

	// No two junction overhangs may be identical or reverse complements,
	// otherwise the two halves would mis-assemble when combined.
	allFrags := append(append([]string{}, before...), after...)
	overhangs := []string{before[0][:4]}
	for _, frag := range allFrags {
		overhangs = append(overhangs, frag[len(frag)-4:])
	}
	for i := 0; i < len(overhangs); i++ {
		for j := i + 1; j < len(overhangs); j++ {
			if overhangs[i] == overhangs[j] || transform.ReverseComplement(overhangs[i]) == overhangs[j] {
				t.Errorf("Overhang collision between %s and %s", overhangs[i], overhangs[j])
			}
		}
	}

	// The combined overhang set should assemble efficiently.
	if efficiency <= 0.85 {
		t.Errorf("Expected efficiency > 0.85, got %g", efficiency)
	}

	// Round trip: concatenating all fragments (before then after) while dropping
	// each shared overhang must reproduce the original sequence.
	if got := reassembleFragments(allFrags); got != strings.ToUpper(breakPlasmid) {
		t.Errorf("Round trip did not reproduce the original sequence")
	}
}

func TestFragmentWithBreakRecursion(t *testing.T) {
	// Two-level synthesis. The top level breaks the plasmid into ~1kbp chunks
	// (800-1000bp) with a break in the kanR window. Each of those chunks is
	// itself assembled from smaller 151-176bp pieces, so we recurse down by
	// fragmenting each chunk again.
	before, after, _, err := fragment.FragmentWithBreak(breakPlasmid, 1962, 2485, 800, 1000, []string{})
	if err != nil {
		t.Fatal(err.Error())
	}

	chunks := append(append([]string{}, before...), after...)
	for _, chunk := range chunks {
		pieces, _, pieceErr := fragment.Fragment(chunk, 151, 176, []string{})
		if pieceErr != nil {
			t.Errorf("Failed to fragment chunk into pieces: %s", pieceErr)
		}
		if len(pieces) == 0 {
			t.Errorf("Expected at least one piece per chunk")
		}
	}
}
