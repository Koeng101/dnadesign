/*
Package megamash is an implementation of the megamash algorithm.

Megamash is an algorithm developed by Keoni Gandall to find templates from
sequencing reactions. For example, you may have a pool of amplicons, and need
to get a count of how many times each amplicon shows up in a given sequencing
reaction.
*/
package megamash

import (
	"fmt"

	"github.com/koeng101/dnadesign/lib/bio/fasta"
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
	Kmers                 map[string]string
	IdentifierToKmerCount map[string]uint
	KmerSize              uint
	KmerMinimalCount      uint
	Threshold             float64
}

// NewMegamashMap creates a megamash map that can be searched against.
func NewMegamashMap(sequences []fasta.Record, kmerSize uint, kmerMinimalCount uint, threshold float64) (MegamashMap, error) {
	var megamashMap MegamashMap
	megamashMap.KmerSize = kmerSize
	megamashMap.KmerMinimalCount = kmerMinimalCount
	megamashMap.Threshold = threshold
	megamashMap.Kmers = make(map[string]string)

	kmerMap := make(map[string]string)
	bannedKmers := make(map[string]int)
	for _, fastaRecord := range sequences {
		sequence := fastaRecord.Sequence
		sequenceSpecificKmers := make(map[string]bool)
		for i := 0; i <= len(sequence)-int(kmerSize); i++ {
			kmerString := StandardizedDNA(sequence[i : i+int(kmerSize)])
			kmerMap[kmerString] = fastaRecord.Identifier
			sequenceSpecificKmers[kmerString] = true
		}
		for kmerString := range sequenceSpecificKmers {
			_, ok := bannedKmers[kmerString]
			if !ok {
				bannedKmers[kmerString] = 1
			} else {
				bannedKmers[kmerString]++
			}
		}
	}
	for kmerString, identifier := range kmerMap {
		kmerCount, ok := bannedKmers[kmerString]
		if ok {
			if kmerCount == 1 {
				megamashMap.Kmers[kmerString] = identifier
			}
		}
	}
	// Check for minimal kmerCount
	identifierToCount := make(map[string]uint)
	for _, fastaRecord := range sequences {
		identifierToCount[fastaRecord.Identifier] = 0
	}
	for _, identifier := range megamashMap.Kmers {
		identifierToCount[identifier]++
	}
	megamashMap.IdentifierToKmerCount = identifierToCount
	for identifier, count := range identifierToCount {
		if count < kmerMinimalCount {
			return megamashMap, fmt.Errorf("Got only %d unique kmers of required %d for sequence %s", count, kmerMinimalCount, identifier)
		}
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
	identifierToCounts := make(map[string]uint)
	for identifier := range m.IdentifierToKmerCount {
		identifierToCounts[identifier] = 0
	}
	for i := 0; i <= len(sequence)-int(m.KmerSize); i++ {
		kmerString := StandardizedDNA(sequence[i : i+int(m.KmerSize)])
		identifier, ok := m.Kmers[kmerString]
		if ok {
			identifierToCounts[identifier]++
		}
	}
	// Now we check which has above the threshold
	var matches []Match
	for identifier, totalKmers := range m.IdentifierToKmerCount {
		matchedKmers := identifierToCounts[identifier]
		score := float64(matchedKmers) / float64(totalKmers)
		if score > m.Threshold {
			matches = append(matches, Match{Identifier: identifier, Score: score})
		}
	}
	return matches
}
