package pfragment

import (
	"math"
	"sort"
	"strings"
	"sync"

	"github.com/koeng101/dnadesign/lib/synthesis/codon"
	"github.com/koeng101/dnadesign/lib/synthesis/fix"
)

// splitString splits a string into nearly equal parts, each of which is as close as possible to maxChunkSize.
func splitString(str string, maxChunkSize int) []string {
	var chunks []string
	strLength := len(str)

	// Determine the number of chunks
	numChunks := int(math.Ceil(float64(strLength) / float64(maxChunkSize)))

	// Calculate the size of each chunk
	chunkSize := int(math.Ceil(float64(strLength) / float64(numChunks)))

	for i := 0; i < strLength; i += chunkSize {
		end := i + chunkSize
		if end > strLength {
			end = strLength
		}
		chunks = append(chunks, str[i:end])
	}

	return chunks
}

// NaiveFragmentProtein naively fragments proteins into chunks.
func NaiveFragmentProtein(proteins []string, overhangLength int, minAaLength int, maxAaLength int) ([][]string, []string) {
	// First, parse out the best overhangs
	var overhangs []string

	// Get mean to iterate over
	meanAaLength := (minAaLength + maxAaLength) / 2
	meanToMin := meanAaLength - minAaLength
	meanToMax := maxAaLength - meanAaLength
	for start := meanAaLength; start+meanToMax < len(proteins[0]); start = start + meanAaLength {
		sequence := proteins[0][start-meanToMin : start+meanToMax]
		subsequences := make(map[string]int)
		for i := 0; i <= len(sequence)-overhangLength; i++ {
			subsequence := sequence[i : i+overhangLength]
			subsequences[subsequence] = 0
		}
		// Now, we check all other proteins to get a "ranking" for this particular site.
		for _, protein := range proteins {
			for subsequence := range subsequences {
				if strings.Count(protein, subsequence) == 1 {
					subsequences[subsequence]++
				}
			}
		}
		// Return the best overhang
		var maxRank int
		var maxSubSequence string
		for subsequence, ranking := range subsequences {
			if ranking >= maxRank {
				maxRank = ranking
				maxSubSequence = subsequence
			}
		}
		overhangs = append(overhangs, maxSubSequence)
	}

	// Now fragment proteins
	var fragmentedProteins [][]string

	for _, protein := range proteins {
		var splitIndices []int

		// First, check if it should be fragmented at certain
		// overhangs.
		var targetOverhangs []string
		for _, overhang := range overhangs {
			if strings.Count(protein, overhang) == 1 {
				targetOverhangs = append(targetOverhangs, overhang)
			}
		}

		// Collect indices for all overhang occurrences
		for _, overhang := range targetOverhangs {
			index := strings.Index(protein, overhang)
			for index != -1 {
				splitIndices = append(splitIndices, index)
				index = strings.Index(protein[index+len(overhang):], overhang)
				if index != -1 {
					index += len(overhang)
				}
			}
		}

		sort.Ints(splitIndices)

		// Use indices to split the protein
		var start int
		var fragments []string
		for _, index := range splitIndices {
			if start != index {
				fragments = append(fragments, protein[start:index])
				start = index
			}
		}
		fragments = append(fragments, protein[start:])

		// Split into equal sized fragments
		var splitFragments []string
		for _, fragment := range fragments {
			if len(fragment) > maxAaLength {
				splitFragments = append(splitFragments, splitString(fragment, maxAaLength)...)
			} else {
				splitFragments = append(splitFragments, fragment)
			}
		}

		fragmentedProteins = append(fragmentedProteins, splitFragments)
	}

	return fragmentedProteins, overhangs
}

// NaiveProteinFragmentationAndOptimization fragments proteins into DNA chunks,
// trying to maintain as much similarity between overhangs as possible.
func NaiveProteinFragmentationAndOptimization(proteins []string, overhangLength int, minAaLength int, maxAaLength int, codonTable codon.TranslationTable, problematicSequenceFuncs []func(string, chan fix.DnaSuggestion, *sync.WaitGroup), prefix string, suffix string) ([][]string, error) {
	// First, fragment the proteins
	proteinFragmentsList, _ := NaiveFragmentProtein(proteins, overhangLength, minAaLength, maxAaLength)

	// Next, build the checks we'll have for the overhangs. We want them to not
	// contain the suffix / prefix overhangs.
	var overhangFunctions []func(string, chan fix.DnaSuggestion, *sync.WaitGroup)
	overhangFunctions = append(overhangFunctions, problematicSequenceFuncs...)
	overhangFunctions = append(overhangFunctions, fix.RemoveSequence([]string{prefix[:4], suffix[len(suffix)-4:]}, "final protein overhangs"))

	// Even though we have overhangLength for the length of the protein
	// overhangs, we don't want to use them. We will use the first 2 amino
	// acids from each overhang, and create a standard library to pull from.
	overhangs := make(map[string]string)
	var allOptimizedSequences [][]string
	for _, fragmentedProtein := range proteinFragmentsList {
		var optimizedSequences []string
		for i, fragment := range fragmentedProtein {
			// We do not need to find overhangs for the first fragment.
			if i == 0 {
				optimizedSequence, err := codonTable.Optimize(fragment)
				if err != nil {
					return allOptimizedSequences, err
				}
				fixedSequence, _, err := fix.Cds(optimizedSequence, &codonTable, problematicSequenceFuncs)
				if err != nil {
					return allOptimizedSequences, err
				}
				optimizedSequences = append(optimizedSequences, fixedSequence)
				continue
			}
			// We need to standardize overhangs for all other fragments.
			// If not in our overhang map, append it!
			targetOverhang := fragment[0:2]
			_, ok := overhangs[targetOverhang]
			if !ok {
				optimizedOverhang, err := codonTable.Optimize(targetOverhang)
				if err != nil {
					return allOptimizedSequences, err
				}
				fixedOverhang, _, err := fix.Cds(optimizedOverhang, &codonTable, overhangFunctions)
				if err != nil {
					return allOptimizedSequences, err
				}
				overhangs[targetOverhang] = fixedOverhang
			}
			// Now that we have a standardized overhang, apply it while optimizing
			optimizedSequence, err := codonTable.Optimize(fragment[2:])
			if err != nil {
				return allOptimizedSequences, err
			}
			optimizedSequence = overhangs[targetOverhang] + optimizedSequence
			fixedSequence, _, err := fix.Cds(optimizedSequence, &codonTable, problematicSequenceFuncs)
			if err != nil {
				return allOptimizedSequences, err
			}
			optimizedSequences = append(optimizedSequences, fixedSequence)
		}
		allOptimizedSequences = append(allOptimizedSequences, optimizedSequences)
	}
	// However, we can't natively just synthesize these sequences quite yet.
	// We need to apply the prefix / suffix to the first / last fragment,
	// respectively, and then apply each overhang to the fragment preceeding.
	var sequences [][]string
	for _, fragmentedProtein := range allOptimizedSequences {
		var overhangsApplied []string
		for i, sequence := range fragmentedProtein {
			switch {
			// Check if first
			case i == 0:
				overhangsApplied = append(overhangsApplied, prefix+sequence+fragmentedProtein[i+1][:4])
			// Check if last
			case i == len(fragmentedProtein)-1:
				overhangsApplied = append(overhangsApplied, sequence+suffix)
			default:
				overhangsApplied = append(overhangsApplied, sequence+fragmentedProtein[i+1][:4])
			}
		}
		sequences = append(sequences, overhangsApplied)
	}
	return sequences, nil
}
