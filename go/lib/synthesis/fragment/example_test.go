package fragment_test

import (
	"fmt"

	"github.com/koeng101/dnadesign/lib/synthesis/fragment"
)

// This example shows how to use the fragmenter to fragment a gene in
// preparation for synthesis. Inputs are the sequence, the minimal fragment
// length, the maximum fragment length, and a list of overhangs that will
// also be used in the assembly reaction.
func Example_basic() {
	lacZ := "ATGACCATGATTACGCCAAGCTTGCATGCCTGCAGGTCGACTCTAGAGGATCCCCGGGTACCGAGCTCGAATTCACTGGCCGTCGTTTTACAACGTCGTGACTGGGAAAACCCTGGCGTTACCCAACTTAATCGCCTTGCAGCACATCCCCCTTTCGCCAGCTGGCGTAATAGCGAAGAGGCCCGCACCGATCGCCCTTCCCAACAGTTGCGCAGCCTGAATGGCGAATGGCGCCTGATGCGGTATTTTCTCCTTACGCATCTGTGCGGTATTTCACACCGCATATGGTGCACTCTCAGTACAATCTGCTCTGATGCCGCATAG"
	fragments, _, _ := fragment.Fragment(lacZ, 95, 105, []string{"AAAA"})

	fmt.Println(fragments)
	// Output: [ATGACCATGATTACGCCAAGCTTGCATGCCTGCAGGTCGACTCTAGAGGATCCCCGGGTACCGAGCTCGAATTCACTGGCCGTCGTTTTACAACGTCGTGACTGG CTGGGAAAACCCTGGCGTTACCCAACTTAATCGCCTTGCAGCACATCCCCCTTTCGCCAGCTGGCGTAATAGCGAAGAGGCCCGCACCGATCGCCCTTCCCAAC CAACAGTTGCGCAGCCTGAATGGCGAATGGCGCCTGATGCGGTATTTTCTCCTTACGCATC CATCTGTGCGGTATTTCACACCGCATATGGTGCACTCTCAGTACAATCTGCTCTGATGCCGCATAG]
}

// This example shows how to generate a new overhang onto a list of overhangs.
func ExampleNextOverhang() {
	primerOverhangs := []string{"ATAA"}
	primerOverhangs = append(primerOverhangs, fragment.NextOverhang(primerOverhangs))
	primerOverhangs = append(primerOverhangs, fragment.NextOverhang(primerOverhangs))
	primerOverhangs = append(primerOverhangs, fragment.NextOverhang(primerOverhangs))

	fmt.Println(primerOverhangs)
	// Output: [ATAA AAAT AATA AAGA]
}

// This example shows how to fragment a sequence with a break forced into a
// given window. In full plasmid synthesis the first assembly is ampR-selected
// while a kanR marker is deliberately interrupted so it is non-functional; the
// "before" and "after" fragments only come together in a later GoldenGate,
// reconstituting full-length kanR.
func ExampleFragmentWithBreak() {
	before, after, _, _ := fragment.FragmentWithBreak(breakPlasmid, 1962, 2485, 800, 1000, []string{})

	// The break overhang is the junction shared by the last before-fragment and
	// the first after-fragment.
	lastBefore := before[len(before)-1]
	breakOverhang := lastBefore[len(lastBefore)-4:]

	fmt.Println(breakOverhang, breakOverhang == after[0][:4])
	// Output: TCAC true
}

func ExampleFragment() {
	lacZ := "ATGACCATGATTACGCCAAGCTTGCATGCCTGCAGGTCGACTCTAGAGGATCCCCGGGTACCGAGCTCGAATTCACTGGCCGTCGTTTTACAACGTCGTGACTGGGAAAACCCTGGCGTTACCCAACTTAATCGCCTTGCAGCACATCCCCCTTTCGCCAGCTGGCGTAATAGCGAAGAGGCCCGCACCGATCGCCCTTCCCAACAGTTGCGCAGCCTGAATGGCGAATGGCGCCTGATGCGGTATTTTCTCCTTACGCATCTGTGCGGTATTTCACACCGCATATGGTGCACTCTCAGTACAATCTGCTCTGATGCCGCATAG"
	fragments, efficiency, _ := fragment.Fragment(lacZ, 95, 105, []string{})

	fmt.Printf("%s : %f", fragments[1], efficiency)
	// Output: CTGGGAAAACCCTGGCGTTACCCAACTTAATCGCCTTGCAGCACATCCCCCTTTCGCCAGCTGGCGTAATAGCGAAGAGGCCCGCACCGATCGCCCTTCCCAACA : 1.000000
}
