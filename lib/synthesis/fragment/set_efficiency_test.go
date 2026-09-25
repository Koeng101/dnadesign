package fragment_test

import (
	"math"
	"math/rand"
	"testing"

	"github.com/koeng101/dnadesign/lib/checks"
	"github.com/koeng101/dnadesign/lib/synthesis/fragment"
	"github.com/koeng101/dnadesign/lib/transform"
)

// allOverhangs returns every 4bp overhang that is legal in a GoldenGate
// reaction, i.e. every 4-mer that is not palindromic.
func allOverhangs() []string {
	bases := []rune{'A', 'T', 'G', 'C'}
	var out []string
	for _, b1 := range bases {
		for _, b2 := range bases {
			for _, b3 := range bases {
				for _, b4 := range bases {
					overhang := string([]rune{b1, b2, b3, b4})
					if !checks.IsPalindromic(overhang) {
						out = append(out, overhang)
					}
				}
			}
		}
	}
	return out
}

// TestSetEfficiencyScoresEveryOverhang is the regression test for the bug this
// file is named after. The NEB ligation data is reverse-complement
// canonicalised -- it stores one member of each RC pair -- and SetEfficiency
// used to look an overhang up exactly as written. For the 120 overhangs stored
// under their reverse complement every lookup missed, a missing key in a Go map
// reads as 0, and the guard against dividing 0/0 then skipped the overhang
// entirely: a set the table had no data for scored a PERFECT 1.0.
//
// A set of one overhang repeated is the sharpest case available: the same
// overhang on both sides of a junction can only mis-ligate, so no real overhang
// may score 1.0 here.
func TestSetEfficiencyScoresEveryOverhang(t *testing.T) {
	var unscored []string
	for _, overhang := range allOverhangs() {
		if fragment.SetEfficiency([]string{overhang, overhang}) == 1.0 {
			unscored = append(unscored, overhang)
		}
	}
	if len(unscored) != 0 {
		t.Errorf("%d of %d overhangs scored a perfect 1.0 when duplicated, so the "+
			"fidelity table has no data for them: %v", len(unscored), len(allOverhangs()), unscored)
	}
}

// TestSetEfficiencyStrandInvariance pins the property that makes the fix
// correct. An overhang and its reverse complement are the two strands of one
// junction, so they are the same measurement: rewriting any subset of a set on
// the other strand must not change the set's fidelity. Before the fix, flipping
// an overhang could move its score between a real value and 1.0.
func TestSetEfficiencyStrandInvariance(t *testing.T) {
	overhangs := allOverhangs()
	random := rand.New(rand.NewSource(1))
	for iteration := 0; iteration < 2000; iteration++ {
		// Build a random set of 2-6 overhangs, no two of which are the same
		// junction written either way.
		size := 2 + random.Intn(5)
		var set []string
		used := make(map[string]bool)
		for len(set) < size {
			candidate := overhangs[random.Intn(len(overhangs))]
			if used[candidate] || used[transform.ReverseComplement(candidate)] {
				continue
			}
			used[candidate] = true
			set = append(set, candidate)
		}
		// Flip a random subset onto the other strand.
		flipped := make([]string, len(set))
		for index, overhang := range set {
			flipped[index] = overhang
			if random.Intn(2) == 0 {
				flipped[index] = transform.ReverseComplement(overhang)
			}
		}
		straight, other := fragment.SetEfficiency(set), fragment.SetEfficiency(flipped)
		if math.Abs(straight-other) > 1e-9 {
			t.Fatalf("SetEfficiency depends on which strand the overhangs are written as:\n"+
				"  %v -> %f\n  %v -> %f", set, straight, flipped, other)
		}
	}
}

// TestSetEfficiencyKnownPair is the concrete pair that cost a real oligo pool.
// CCCG and CCTG differ at a single position and are both GC rich, the regime
// where T4 ligase mis-ligates; both are stored under their reverse complements,
// so the set used to score 1.0 and the fragmenter chose it.
func TestSetEfficiencyKnownPair(t *testing.T) {
	asOrdered := fragment.SetEfficiency([]string{"CCCG", "CCTG"})
	asComplement := fragment.SetEfficiency([]string{"CGGG", "CAGG"})
	if math.Abs(asOrdered-asComplement) > 1e-9 {
		t.Errorf("the same junction pair scored %f written one way and %f the other",
			asOrdered, asComplement)
	}
	if asOrdered >= 1.0 {
		t.Errorf("a one-mismatch GC-rich overhang pair scored %f, expected well under 1.0", asOrdered)
	}
}

// TestFragmentDoesNotPreferUnscoredOverhangs is the end-to-end consequence.
// Because 1.0 is the top of the scale, an unscored overhang beat every scored
// candidate in optimizeOverhangIteration, so the fragmenter actively preferred
// the overhangs it had no data for. This is a real chunk from an oligo pool
// whose L4 subassembly gave zero sequence-perfect clones out of 37: the
// fragmenter had chosen CCCG and CCTG for it and reported 100% fidelity.
func TestFragmentDoesNotPreferUnscoredOverhangs(t *testing.T) {
	fragments, efficiency, err := fragment.Fragment(rubyNatMXL4, 136, 196, []string{})
	if err != nil {
		t.Fatalf("fragmenting: %s", err)
	}
	overhangs := []string{fragments[0][:4]}
	for _, f := range fragments {
		overhangs = append(overhangs, f[len(f)-4:])
	}
	// Verified against the other strand, which is the point: a set whose
	// score changes when it is rewritten on the complementary strand is a set
	// scored partly on overhangs the table cannot see.
	complemented := make([]string, len(overhangs))
	for index, overhang := range overhangs {
		complemented[index] = transform.ReverseComplement(overhang)
	}
	straight, other := fragment.SetEfficiency(overhangs), fragment.SetEfficiency(complemented)
	if math.Abs(straight-other) > 1e-9 {
		t.Errorf("the fragmenter chose overhangs %v, which score %f on one strand and %f on the other",
			overhangs, straight, other)
	}
	if math.Abs(straight-efficiency) > 1e-9 {
		t.Errorf("reported efficiency %f but the chosen overhang set %v scores %f", efficiency, overhangs, straight)
	}
	if straight < 0.99 {
		t.Errorf("chose overhangs %v scoring only %f, though a clean set exists for this sequence", overhangs, straight)
	}
}

// rubyNatMXL4 is the fourth chunk of a ruby_natMX construct, flanked for
// cloning. Its natMX half is 71-73% GC, which is what left the fragmenter
// choosing between GC-rich overhangs it could not score.
const rubyNatMXL4 = "GTCTCTCCATGTAAAATGACCACTCTTGACGACACGGCTTACCGGTACCGCACCAGTGTCCCGGGGGACGCCGAGGCCATCGAGGCACTGGATGGGTCCTTCACCACCGACACCGTCTTCCGCGTCACCGCCACCGGGGACGGCTTCACCCTGCGGGAGGTGCCGGTGGACCCGCCCCTGACCAAGGTGTTCCCCGACGACGAATCGGACGACGAATCGGACGACGGGGAGGACGGCGACCCGGACTCCCGGACGTTCGTCGCGTACGGGGACGACGGCGACCTGGCGGGCTTCGTGGTCGTCTCGTACTCCGGCTGGAACCGCCGGCTGACCGTCGAGGACATCGAGGTCGCCCCGGAGCACCGGGGGCACGGGGTCGGGCGCGCGCTGATGGGGCTCGCGACGGAGTTCGCCCGCGAGCGGGGCGCCGGGCACCTCTGGCTGGAGGTCACCAACGTCAACGCACCGGCGATCCACGCGTACCGGCGGATGGGGTTCACCCTCTGCGGCCTGGACACCGCCCTGTACGACGGCACCGCCTCGGACGGCGAGCAGGCGCTCTACATGAGCATGCCCTGCCCCTAGTACTGACAATAAAAAGATTCTTGTTTTCAAGAACTTGTCATTTGTATAGTTTTTTTATATTGTAGTTGTTCTATTTTAATCAAATGTTAGCGTGATTTATATTTTTTTTCGCCTCGACATCATCTGCCCAGATGCGAAGTTAAGTGCGCAGAAAGTAATATCATGCGTCAATCGTATGTGAATGCTGGTCGCTATACTGCGTGTTATCCAGTCCCAGAACGAGCGAG"
