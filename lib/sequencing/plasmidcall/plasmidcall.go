/*
Package plasmidcall contains functions from calling whether or not a plasmid
is mutated or not given sequencing data.

In particular, it assesses a plasmid given cs alignments from minimap2. The
cs alignments are compiled, and at each position, mutations are assigned.
There are a couple kinds of mutations:

  - DepthMutation. Not enough sequencing depth to make any call (at least 5x)
  - NanoporeMutation. Occurs when 80% of mutations are only on 1 strand,
    so long as we also have 5 examples of correct sequence on the other strand.
  - CleanMutation. Occurs when over 80% of a given position is mutated to a
    single mutation type.
  - MixedMutation. Occurs when 20% to 80% of a given position are not correct.
*/
package plasmidcall

import (
	"math"

	"github.com/koeng101/dnadesign/lib/align/cs"
	"github.com/koeng101/dnadesign/lib/bio/fastq"
)

// MutationType contains the potential mutation types, including
// depth, nanopore_error, point, indel, insertion, and noisy.
type MutationType string

const (
	DepthError    MutationType = "depth"
	NanoporeError MutationType = "nanopore_error"
	Point         MutationType = "point"
	Indel         MutationType = "indel"
	Insertion     MutationType = "insertion"
)

// Mutation contains mutation data at a certain location.
type Mutation struct {
	Position       int
	MutatioNString string
	MutationType   MutationType
	Change         string
	CleanMutation  bool // False if MixedMutation, True if CleanMutation
}

type CsAlignment struct {
	Read              fastq.Read
	ReverseComplement bool
	CS                []cs.CS
}

// CallMutations calls mutations from a range of cs alignments.
func CallMutations(referenceSequence string, alignments []CsAlignment) ([]Mutation, error) {
	totalDigestedCS := make(map[int][]cs.DigestedCS)
	totalDigestedInsertions := make(map[int][]cs.DigestedInsertion)
	for _, alignment := range alignments {
		// Produce digestedCS and digestedInsertion for each alignment
		digestedCS, digestedInsertions := cs.DigestCS(alignment.CS, alignment.Read.Quality, alignment.ReverseComplement)
		for i := range digestedCS {
			totalDigestedCS[i] = append(totalDigestedCS[i], digestedCS[i])
		}
		for i := range digestedInsertions {
			totalDigestedInsertions[int(digestedInsertions[i].Position)] = append(totalDigestedInsertions[int(digestedInsertions[i].Position)], digestedInsertions[i])
		}
	}
	var mutations []Mutation
	// Now we have a map of all the potential mutations at each point in the sequence. Now, we iterate over them.
	for i := range referenceSequence {
		// i is the particular position of the reference sequence.
		csAtPosition := totalDigestedCS[i]
		insertionsAtPosition := totalDigestedInsertions[i]

		// Check for depth mutation
		if len(csAtPosition) < 5 {
			mutations = append(mutations, Mutation{
				Position:      i,
				MutationType:  DepthError,
				CleanMutation: true,
			})
			continue
		}

		// First, we check for insertions
		if len(insertionsAtPosition) > 0 {
			insertionPercentage := float64(len(insertionsAtPosition)) / float64(len(csAtPosition))
			insertionCounts := make(map[string]int)
			var maxCount int
			var mostCommon string

			for _, insertion := range insertionsAtPosition {
				insertionCounts[insertion.Insertion]++
				if insertionCounts[insertion.Insertion] > maxCount {
					maxCount = insertionCounts[insertion.Insertion]
					mostCommon = insertion.Insertion
				}
			}
			if insertionPercentage > 0.8 {
				mutations = append(mutations, Mutation{
					Position:      i,
					MutationType:  Insertion,
					Change:        mostCommon,
					CleanMutation: true,
				})
			} else if insertionPercentage >= 0.25 {
				mutations = append(mutations, Mutation{
					Position:      i,
					MutationType:  Insertion,
					Change:        mostCommon,
					CleanMutation: false,
				})
			}
		}

		// Count occurrences of each mutation type
		mutationCounts := make(map[uint8]int) // used later
		totalMutations := 0
		forwardStrand := make(map[uint8]int)
		reverseStrand := make(map[uint8]int)
		for _, digestedCS := range csAtPosition {
			if !digestedCS.ReverseComplement {
				forwardStrand[digestedCS.Type]++
			} else {
				reverseStrand[digestedCS.Type]++
			}

			mutationCounts[digestedCS.Type]++
			if digestedCS.Type != '.' {
				totalMutations++
			}
		}
		// Now, get the mutational ratios of forward vs reverse.
		// These will be used to call nanopore errors. Essentially,
		// if we see a mutation to correct ratio that only occurs on
		// a single strand (80% difference), we will call it as a nanopore
		// error. Practically speaking, this means if there is a 5x
		// representation of mutations on only one strand, we believe this
		// is a nanopore sequencing error, and treat it as such.
		mutationalRatios := make(map[uint8]float64)
		for k := range forwardStrand {
			mutationalRatios[k] = 0
		}
		for k := range reverseStrand {
			mutationalRatios[k] = 0
		}
		var foundNanoporeMutation bool
		for k := range mutationalRatios {
			if k != '.' {
				forwardRatio := forwardStrand[k] / forwardStrand['.']
				reverseRatio := reverseStrand[k] / reverseStrand['.']
				// This just runs the test where if a strands correct/incorrect ratio is 5x on one strand and has 5
				// correct reads on the other strand, then we
				if ((forwardRatio > reverseRatio*5) && (reverseStrand['.'] > 5)) || ((reverseRatio > forwardRatio*5) && (forwardStrand['.'] > 5)) {
					mutations = append(mutations, Mutation{Position: i, MutationType: NanoporeError, CleanMutation: true})
					foundNanoporeMutation = true
					break
				}
			}
		}
		// Skip checking other mutations if we find a nanopore mutation
		if foundNanoporeMutation {
			continue
		}

		// Check for clean or mixed mutations
		if totalMutations > 0 {
			var mostCommonMutation uint8
			var maxCount int
			for mut, count := range mutationCounts {
				if count > maxCount {
					mostCommonMutation = mut
					maxCount = count
				}
			}

			mutationType := Point
			if mostCommonMutation == '*' {
				mutationType = Indel
			}

			// Check for a clean mutation
			mutationPercentage := float64(maxCount) / float64(len(csAtPosition))
			if mutationPercentage > 0.8 {
				if mostCommonMutation != '.' {
					mutations = append(mutations, Mutation{
						Position:      i,
						MutationType:  mutationType,
						Change:        string(mostCommonMutation),
						CleanMutation: true,
					})
					continue
				}
			}

			// Now, there might now a mixture of mutations
			var foundMixedMutation bool
			for mut, count := range mutationCounts {
				if mut != '.' {
					if (float64(count)/float64(mutationCounts['.']) + math.SmallestNonzeroFloat64) > .25 { // we add epsilon to prevent divide by zero
						mutations = append(mutations, Mutation{
							Position:      i,
							MutationType:  mutationType,
							Change:        string(mostCommonMutation),
							CleanMutation: false,
						})
						foundMixedMutation = true
						break
					}
				}
			}
			if foundMixedMutation {
				continue
			}
		}
	}
	return mutations, nil
}
