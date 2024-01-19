/*
Package megamash is an implementation of the megamash algorithm.

Megamash is an algorithm developed by Keoni Gandall to find templates from
sequencing reactions. For example, you may have a pool of amplicons, and need
to get a count of how many times each amplicon shows up in a given sequencing
reaction.
*/
package megamash

import (
	"context"
	"fmt"

	"github.com/koeng101/dnadesign/lib/bio/fasta"
	"github.com/koeng101/dnadesign/lib/bio/fastq"
	"github.com/koeng101/dnadesign/lib/transform"
)

// StandardizedDNA returns the alphabetically lesser strand of a double
// stranded DNA molecule.
func StandardizedDNA(sequence string) string {
	var deterministicSequence string
	reverseComplement := transform.ReverseComplement(sequence)
	if sequence > reverseComplement {
		deterministicSequence = reverseComplement
	} else {
		deterministicSequence = sequence
	}
	return deterministicSequence
}

var (
	DefaultKmerSize         uint    = 16
	DefaultMinimalKmerCount uint    = 10
	DefaultScoreThreshold   float64 = 0.2
)

type MegamashMap struct {
	Identifiers      []string
	Kmers            []map[string]bool
	KmerSize         uint
	KmerMinimalCount uint
	Threshold        float64
}

// NewMegamashMap creates a megamash map that can be searched against.
func NewMegamashMap(sequences []fasta.Record, kmerSize uint, kmerMinimalCount uint, threshold float64) (MegamashMap, error) {
	var megamashMap MegamashMap
	megamashMap.KmerSize = kmerSize
	megamashMap.KmerMinimalCount = kmerMinimalCount
	megamashMap.Threshold = threshold

	for _, fastaRecord := range sequences {
		megamashMap.Identifiers = append(megamashMap.Identifiers, fastaRecord.Identifier)
		sequence := fastaRecord.Sequence

		// First get all kmers with a given sequence
		kmerMap := make(map[string]bool)
		for i := 0; i <= len(sequence)-int(kmerSize); i++ {
			kmerString := StandardizedDNA(sequence[i : i+int(kmerSize)])
			kmerMap[kmerString] = true
		}

		// Then, get unique kmers for this sequence and only this sequence
		uniqueKmerMap := make(map[string]bool)
		for kmerBase64 := range kmerMap {
			unique := true
			for _, otherMegaMashMap := range megamashMap.Kmers {
				_, ok := otherMegaMashMap[kmerBase64]
				// If this kmer is found, set both to false
				if ok {
					otherMegaMashMap[kmerBase64] = false
					unique = false
					break
				}
			}
			if unique {
				uniqueKmerMap[kmerBase64] = true
			}
		}
		// Check if we have the minimal kmerCount
		var kmerCount uint = 0
		for _, unique := range uniqueKmerMap {
			if unique {
				kmerCount++
			}
		}
		if kmerCount < kmerMinimalCount {
			return megamashMap, fmt.Errorf("Got only %d unique kmers of required %d for sequence %s", kmerCount, kmerMinimalCount, fastaRecord.Identifier)
		}

		// Now we have a unique Kmer map for the given sequence.
		// Add it to megamashMap
		megamashMap.Kmers = append(megamashMap.Kmers, uniqueKmerMap)
	}
	return megamashMap, nil
}

// Match contains the identifier and score of a potential match to the searched
// sequence.
type Match struct {
	Identifier string
	Score      float64
}

// Match matches a sequence to all the sequences in a megamash map.
func (m *MegamashMap) Match(sequence string) []Match {
	var scores []float64
	// The algorithm is as follows:
	// - Go through each map.
	// - Get the number of matching kmers
	// - Divide that by the total kmers available for matching

	// First, get the kmer total
	var kmerSize int
out:
	for _, maps := range m.Kmers {
		for kmer := range maps {
			kmerSize = len(kmer)
			break out
		}
	}

	// Now, iterate through each map
	for _, sequenceMap := range m.Kmers {
		var score float64
		var totalKmers = len(sequenceMap)
		var matchedKmers int
		for i := 0; i <= len(sequence)-kmerSize; i++ {
			kmerString := StandardizedDNA(sequence[i : i+kmerSize])
			unique, ok := sequenceMap[kmerString]
			if ok && unique {
				matchedKmers++
			}
		}
		if totalKmers == 0 {
			score = 0
		} else {
			score = float64(matchedKmers) / float64(totalKmers)
		}
		scores = append(scores, score)
	}

	var matches []Match
	for i, score := range scores {
		if score > m.Threshold {
			matches = append(matches, Match{Identifier: m.Identifiers[i], Score: score})
		}
	}
	return matches
}

// FastqMatchChannel processes a channel of fastq.Read and pushes to a channel of matches.
func (m *MegamashMap) FastqMatchChannel(ctx context.Context, sequences <-chan fastq.Read, matches chan<- []Match) error {
	for {
		select {
		case <-ctx.Done():
			// Clean up resources, handle cancellation.
			return ctx.Err()
		case sequence, ok := <-sequences:
			if !ok {
				close(matches)
				return nil
			}
			sequenceMatches := m.Match(sequence.Sequence)
			matches <- sequenceMatches
		}
	}
}
