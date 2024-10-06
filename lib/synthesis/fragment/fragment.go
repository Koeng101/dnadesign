/*
Package fragment optimally fragments DNA for GoldenGate systems.

Optimal fragmentation is accomplished by using empirical fidelity data derived
by NEB in the paper "Enabling one-pot Golden Gate assemblies of unprecedented
complexity using data-optimized assembly design". We use the BsaI-T4 ligase
data provided in table S1.

Paper link: https://doi.org/10.1371/journal.pone.0238592
Data link: https://doi.org/10.1371/journal.pone.0238592.s001
*/
package fragment

import (
	"errors"
	"fmt"
	"strings"

	"github.com/koeng101/dnadesign/lib/checks"
	"github.com/koeng101/dnadesign/lib/transform"
)

// SetEfficiency gets the estimated fidelity rate of a given set of
// GoldenGate overhangs.
func SetEfficiency(overhangs []string) float64 {
	var efficiency = float64(1.0)
	for _, overhang := range overhangs {
		nCorrect := mismatches[key{overhang, overhang}]
		nTotal := 0
		for _, overhang2 := range overhangs {
			nTotal = nTotal + mismatches[key{overhang, overhang2}]
			nTotal = nTotal + mismatches[key{overhang, transform.ReverseComplement(overhang2)}]
		}
		if nTotal != nCorrect {
			efficiency = efficiency * (float64(nCorrect) / float64(nTotal))
		}
	}
	return efficiency
}

// NextOverhangs gets a list of possible next overhangs to use for an overhang
// list, along with their efficiencies. This can be used for more optimal
// fragmentation of sequences with potential degeneracy.
func NextOverhangs(currentOverhangs []string) ([]string, []float64) {
	currentOverhangMap := make(map[string]bool)
	for _, overhang := range currentOverhangs {
		currentOverhangMap[overhang] = true
	}

	// These 4 for loops generate all combinations of 4 base pairs
	// checking all 256 4mer combinations for palindromes. Palindromes
	// can cause problems in large combinatorial reactions, so are
	// removed here.
	bases := []rune{'A', 'T', 'G', 'C'}
	var overhangsToTest []string
	for _, base1 := range bases {
		for _, base2 := range bases {
			for _, base3 := range bases {
				for _, base4 := range bases {
					newOverhang := string([]rune{base1, base2, base3, base4})
					_, ok := currentOverhangMap[newOverhang]
					_, okReverse := currentOverhangMap[transform.ReverseComplement(newOverhang)]
					if !ok && !okReverse {
						if !checks.IsPalindromic(newOverhang) {
							overhangsToTest = append(overhangsToTest, newOverhang)
						}
					}
				}
			}
		}
	}

	var efficiencies []float64
	for _, overhang := range overhangsToTest {
		strandEffieciency := SetEfficiency(append(currentOverhangs, overhang))
		complementEfficiency := SetEfficiency(append(currentOverhangs, transform.ReverseComplement(overhang)))
		efficiencies = append(efficiencies, (strandEffieciency+complementEfficiency)/2)
	}
	return overhangsToTest, efficiencies
}

// NextOverhang gets next most efficient overhang to use for a given set of
// overhangs. This is useful for when developing a new set of standard
// overhangs. Note: NextOverhang is biased towards high AT overhangs, but this
// will not affect fidelity at all.
func NextOverhang(currentOverhangs []string) string {
	overhangsToTest, efficiencies := NextOverhangs(currentOverhangs)
	var efficiency float64
	var newOverhang string
	maxEfficiency := float64(0)
	for i, overhang := range overhangsToTest {
		efficiency = efficiencies[i]
		if efficiency > maxEfficiency {
			maxEfficiency = efficiency
			newOverhang = overhang
		}
	}
	return newOverhang
}

// optimizeOverhangIteration takes in a sequence and optimally fragments it.
func optimizeOverhangIteration(sequence string, minFragmentSize int, maxFragmentSize int, existingFragments []string, excludeOverhangs []string, includeOverhangs []string) ([]string, float64, error) {
	recurseMaxFragmentSize := maxFragmentSize // this is for the final iteration, so we don't get continuously shorter fragments
	// If the sequence is smaller than maxFragment size, stop iteration.
	if len(sequence) < maxFragmentSize {
		existingFragments = append(existingFragments, sequence)
		return existingFragments, SetEfficiency(excludeOverhangs), nil
	}

	// Make sure minFragmentSize > maxFragmentSize
	if minFragmentSize > maxFragmentSize {
		return []string{}, float64(0), fmt.Errorf("minFragmentSize (%d) larger than maxFragmentSize (%d)", minFragmentSize, maxFragmentSize)
	}

	// Minimum lengths (given oligos) for assembly is 8 base pairs
	// https://doi.org/10.1186/1756-0500-3-291
	// For GoldenGate, 2 8bp oligos create 12 base pairs (4bp overhangs on two sides of 4bp),
	// so we check for minimal size of 12 base pairs.
	if minFragmentSize < 12 {
		return []string{}, float64(0), fmt.Errorf("minFragmentSize must be equal to or greater than 12 . Got size of %d", minFragmentSize)
	}

	// If our iteration is approaching the end of the sequence, that means we need to gracefully handle
	// the end so we aren't left with a tiny fragment that cannot be synthesized. For example, if our goal
	// is fragments of 100bp, and we have 110 base pairs left, we want each final fragment to be 55bp, not
	// 100 and 10bp
	if len(sequence) < 2*maxFragmentSize {
		maxAndMinDifference := maxFragmentSize - minFragmentSize
		maxFragmentSizeBuffer := len(sequence) / 2
		minFragmentSize = maxFragmentSizeBuffer - maxAndMinDifference
		if minFragmentSize < 12 {
			minFragmentSize = 12
		}
		maxFragmentSize = maxFragmentSizeBuffer // buffer is needed equations above pass.
	}

	// Get all sets of 4 between the min and max FragmentSize
	var bestOverhangEfficiency float64
	var bestOverhangPosition int
	var alreadyExists bool
	var buildAvailable bool
	for overhangOffset := 0; overhangOffset <= maxFragmentSize-minFragmentSize; overhangOffset++ {
		// We go from max -> min, so we can maximize the size of our fragments
		overhangPosition := maxFragmentSize - overhangOffset
		overhangToTest := sequence[overhangPosition-4 : overhangPosition]

		// Make sure overhang isn't already in set
		alreadyExists = false
		for _, excludeOverhang := range excludeOverhangs {
			if excludeOverhang == overhangToTest || transform.ReverseComplement(excludeOverhang) == overhangToTest {
				alreadyExists = true
			}
		}
		// Make sure overhang is in set of includeOverhangs. If includeOverhangs is
		// blank, skip this check.
		buildAvailable = false
		if len(includeOverhangs) == 0 {
			buildAvailable = true
		}
		for _, includeOverhang := range includeOverhangs {
			if includeOverhang == overhangToTest || transform.ReverseComplement(includeOverhang) == overhangToTest {
				buildAvailable = true
			}
		}
		if !alreadyExists && buildAvailable {
			// See if this overhang is a palindrome
			if !checks.IsPalindromic(overhangToTest) {
				// Get this overhang set's efficiency
				setEfficiency := SetEfficiency(append(excludeOverhangs, overhangToTest))

				// If this overhang is more efficient than any other found so far, set it as the best!
				if setEfficiency > bestOverhangEfficiency {
					bestOverhangEfficiency = setEfficiency
					bestOverhangPosition = overhangPosition
				}
			}
		}
	}
	// Set variables
	if bestOverhangPosition == 0 {
		return []string{}, float64(0), fmt.Errorf("bestOverhangPosition failed by equaling zero")
	}
	existingFragments = append(existingFragments, sequence[:bestOverhangPosition])
	excludeOverhangs = append(excludeOverhangs, sequence[bestOverhangPosition-4:bestOverhangPosition])
	sequence = sequence[bestOverhangPosition-4:]
	return optimizeOverhangIteration(sequence, minFragmentSize, recurseMaxFragmentSize, existingFragments, excludeOverhangs, includeOverhangs)
}

// Fragment fragments a sequence into fragments between the min and max size,
// choosing fragment ends for optimal assembly efficiency. Since fragments will
// be inserted into either a vector or primer binding sites, the first 4 and
// last 4 base pairs are the initial overhang set.
func Fragment(sequence string, minFragmentSize int, maxFragmentSize int, excludeOverhangs []string) ([]string, float64, error) {
	sequence = strings.ToUpper(sequence)
	return optimizeOverhangIteration(sequence, minFragmentSize, maxFragmentSize, []string{}, append([]string{sequence[:4], sequence[len(sequence)-4:]}, excludeOverhangs...), []string{})
}

// FragmentWithOverhangs fragments a sequence with only a certain overhang set.
// This is useful if you are constraining the set of possible overhangs when
// doing more advanced forms of cloning.
func FragmentWithOverhangs(sequence string, minFragmentSize int, maxFragmentSize int, excludeOverhangs []string, includeOverhangs []string) ([]string, float64, error) {
	sequence = strings.ToUpper(sequence)
	return optimizeOverhangIteration(sequence, minFragmentSize, maxFragmentSize, []string{}, append([]string{sequence[:4], sequence[len(sequence)-4:]}, excludeOverhangs...), includeOverhangs)
}

/******************************************************************************

                            Higher level assembly

Practically speaking, if we are synthesizing DNA, there is an upper limit to
the quantity of DNA we can assemble in a single reaction because of the
mutation rate of the synthesis reaction and because assemble efficiency drops
as we have more fragments.

So, the functions here help with that. They allow DNA to be broken up into
sub-fragments, then re-assembled from there.

******************************************************************************/

// Assembly is a recursive fragmentation, where higher level assemblies are
// created from lower level assemblies, until you get to the foundation, which
// is a basic fragmentation.
type Assembly struct {
	Sequence      string
	Fragments     []string
	Efficiency    float64
	SubAssemblies []Assembly
}

// RecursiveFragment fragments a sequence recursively into an assembly, which
// can be created from sub-assemblies. This function is for designing large
// pieces of DNA.
//
// maxCodingSizeOligo should be for the max oligo size. If you are synthesizing
// from an oligo pool, I'd recommend about 56 shorter so you can add primers,
// BsaI, and a CA for recursive assembly.
//
// The assemblyPattern should be for how the oligo should be fragmented up.
// For example: []int{5,4,4,5} is a very reasonable standard if you have
// oligos with a 1/2000 mutation rate that are approximately 174bp - you
// would assemble ~870bp fragments, which should have a 64.72% success rate,
// or a ~95% success rate over 3 colonies. Assembly pattern is also just a
// rough... recommendation. Often times the lowest level of oligo has +1 in
// order to fit the right overhangs in. This doesn't matter that much because
// the limiting factor in assemblies is typically mutation rate at that size.
//
// The forwardFlank and reverseFlank are for preparing the sequences for
// recursive assembly. Generally, this involves appending a certain sequence
// to each oligo, and also to the edges of each subassembly. Do not add these
// to the maxCodingSizeOligo: that is done within the function.
func RecursiveFragment(sequence string, maxCodingSizeOligo int, assemblyPattern []int, excludeOverhangs []string, includeOverhangs []string, forwardFlank string, reverseFlank string) (Assembly, error) {
	/*
		Ok, so this is a note for you hackers out there: this algorithm can be
		greatly improved. The optimal way to do this would be to do a continuous
		fragmentation, so that each of the smallest possible assemblies are as
		large as possible. This reduces the number of total assembly reactions
		you'd have to do. However, this requires some special programming that is,
		frankly, difficult (some dynamic programming), so I just implemented the
		simplest thing that would work. I'd love to get contributions to improve
		this function, though.

		What we want to optimize for: as FEW assemblies as possible, with the
		smallest assembly always using the maximal number of bases per oligo.

		- Keoni
	*/

	// There are two magic numbers here, for defining small fragments.
	// While probably not optimal, they really work quite well in real
	// data, so they're used here.
	smallestMinFragmentSizeSubtraction := 60
	minFragmentSizeSubtraction := 100

	var assembly Assembly
	sequence = strings.ReplaceAll(sequence, "\n", "") // replace newlines, in case they crept in
	sequenceLen := len(sequence)

	// get size pattern. This size pattern maps how we need to fragment the sequences
	appendLength := len(forwardFlank) + len(reverseFlank)
	sizes := make([]int, len(assemblyPattern))
	maxSize := (maxCodingSizeOligo - appendLength) * assemblyPattern[0]
	for i := range assemblyPattern {
		if i == 0 {
			sizes[i] = maxSize
			continue
		}
		sizes[i] = sizes[i-1]*assemblyPattern[i] - smallestMinFragmentSizeSubtraction // subtract approx 60bp to give room for finding overhangs
	}
	if sequenceLen <= sizes[0] {
		fragments, efficiency, err := FragmentWithOverhangs(forwardFlank+sequence+reverseFlank, maxCodingSizeOligo-60, maxCodingSizeOligo, excludeOverhangs, includeOverhangs)
		if err != nil {
			return assembly, err
		}
		return Assembly{Sequence: sequence, Fragments: fragments, Efficiency: efficiency}, nil
	}
	// After the smallest possible block, begin iterating for each size.
	for i, size := range sizes[1:] {
		if sequenceLen <= size {
			fragments, efficiency, err := FragmentWithOverhangs(forwardFlank+sequence+reverseFlank, sizes[i]-minFragmentSizeSubtraction, sizes[i], excludeOverhangs, includeOverhangs)
			if err != nil {
				return assembly, err
			}
			// Now we need to get the derived fragments from this overall construction
			var subAssemblies []Assembly
			for _, fragment := range fragments {
				subAssembly, err := RecursiveFragment(fragment, maxCodingSizeOligo, assemblyPattern, excludeOverhangs, includeOverhangs, forwardFlank, reverseFlank)
				if err != nil {
					return subAssembly, err
				}
				subAssemblies = append(subAssemblies, subAssembly)
			}
			return Assembly{Sequence: sequence, Efficiency: efficiency, SubAssemblies: subAssemblies}, nil
		}
	}
	return assembly, errors.New("Fragment too long!")
}
