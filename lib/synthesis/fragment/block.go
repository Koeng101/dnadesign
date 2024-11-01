package fragment

import (
	"fmt"

	"github.com/koeng101/dnadesign/lib/checks"
	"github.com/koeng101/dnadesign/lib/transform"
)

func BlockFragment(sequence string, minFragmentSize int, maxFragmentSize int, maxBlocksPerAssembly int, minEfficiency float64, availableOverhangs []string) ([]string, error) {
	return blockFragment(sequence, minFragmentSize, maxFragmentSize, maxBlocksPerAssembly, minEfficiency, []string{}, availableOverhangs)
}

func blockFragment(sequence string, minFragmentSize int, maxFragmentSize int, maxBlocksPerAssembly int, minEfficiency float64, existingFragments []string, availableOverhangs []string) ([]string, error) {
	recurseMaxFragmentSize := maxFragmentSize // this is for the final iteration, so we don't get continuously shorter fragments
	// If the sequence is smaller than maxFragment size, stop iteration.
	if len(sequence) < maxFragmentSize {
		existingFragments = append(existingFragments, sequence)
		return existingFragments, nil
	}

	// Make sure minFragmentSize > maxFragmentSize
	if minFragmentSize > maxFragmentSize {
		return []string{}, fmt.Errorf("minFragmentSize (%d) larger than maxFragmentSize (%d)", minFragmentSize, maxFragmentSize)
	}

	// Minimum lengths (given oligos) for assembly is 8 base pairs
	// https://doi.org/10.1186/1756-0500-3-291
	// For GoldenGate, 2 8bp oligos create 12 base pairs (4bp overhangs on two sides of 4bp),
	// so we check for minimal size of 12 base pairs.
	if minFragmentSize < 12 {
		return []string{}, fmt.Errorf("minFragmentSize must be equal to or greater than 12 . Got size of %d", minFragmentSize)
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

	// Now that we've run all of our checks, find all the overhangs that are
	// currently a part of our existingFragments. This includes the last 4 base
	// pairs of every existingFragment up to maxBlocksPerAssembly. We also include
	// the first 4 base pairs of the very last fragment, or the first 4 base pairs
	// of the sequence input if existingFragment is shorter than maxBlocksPerAssembly.
	existingOverhangs := make(map[string]bool)
	numExistingFragments := len(existingFragments)
	fragmentsToProcess := min(numExistingFragments, maxBlocksPerAssembly)

	// Process fragments in reverse order
	for i := fragmentsToProcess - 1; i >= 0; i-- {
		fragment := existingFragments[i]
		if len(fragment) >= 4 {
			overhang := fragment[len(fragment)-4:]
			existingOverhangs[overhang] = true
		}
	}
	if numExistingFragments < maxBlocksPerAssembly {
		existingOverhangs[sequence[:4]] = true
	} else {
		existingOverhangs[existingFragments[maxBlocksPerAssembly][:4]] = true
	}
	// Figure out if we are going to have to reserve the last part of the sequence.
	if float64(len(sequence))/float64(recurseMaxFragmentSize)/float64(maxBlocksPerAssembly) < 0.8 {
		existingOverhangs[sequence[len(sequence)-4:]] = true
	}

	// Make existingOverhangs into a list for using with SetEfficiency
	var existingOverhangsList []string
	for overhang := range existingOverhangs {
		existingOverhangsList = append(existingOverhangsList, overhang)
	}
	// Now we have the existingOverhangs. Convert this into a map of available.
	overhangsToTest := make(map[string]float64)
	for _, availableOverhang := range availableOverhangs {
		if checks.IsPalindromic(availableOverhang) {
			return []string{}, fmt.Errorf("%s is palindromic", availableOverhang)
		}
		if !existingOverhangs[availableOverhang] && !existingOverhangs[transform.ReverseComplement(availableOverhang)] {
			setEfficiency := SetEfficiency(append(existingOverhangsList, availableOverhang))
			overhangsToTest[availableOverhang] = setEfficiency
		}
	}

	// Now we have the overhangs. Now actually find a target overhang:
	var bestOverhangEfficiency float64
	var bestOverhangPosition int
	for overhangOffset := 0; overhangOffset <= maxFragmentSize-minFragmentSize; overhangOffset++ {
		// We go from max -> min, so we can maximize the size of our fragments
		overhangPosition := maxFragmentSize - overhangOffset
		overhangToTest := sequence[overhangPosition-4 : overhangPosition]
		for _, seq := range []string{overhangToTest, transform.ReverseComplement(overhangToTest)} {
			if efficiency, ok := overhangsToTest[seq]; ok && efficiency > bestOverhangEfficiency && efficiency > minEfficiency {
				bestOverhangEfficiency = efficiency
				bestOverhangPosition = overhangPosition
			}
		}
	}
	// Set variables
	if bestOverhangPosition == 0 {
		return []string{}, fmt.Errorf("bestOverhangPosition failed by equaling zero")
	}
	existingFragments = append(existingFragments, sequence[:bestOverhangPosition])
	sequence = sequence[bestOverhangPosition-4:]
	return blockFragment(sequence, minFragmentSize, recurseMaxFragmentSize, maxBlocksPerAssembly, minEfficiency, existingFragments, availableOverhangs)
}

// Helper function for min value
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
