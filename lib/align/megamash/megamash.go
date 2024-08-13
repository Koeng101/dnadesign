/*
Package megamash is an implementation of the megamash algorithm.

Megamash is an algorithm developed by Keoni Gandall to find templates from
sequencing reactions. For example, you may have a pool of amplicons, and need
to get a count of how many times each amplicon shows up in a given sequencing
reaction.
*/
package megamash

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/koeng101/dnadesign/lib/bio/fasta"
	"github.com/koeng101/dnadesign/lib/transform"
)

// StandardizeDNA returns the alphabetically lesser strand of a double
// stranded DNA molecule.
func StandardizeDNA(sequence string) string {
	sequence = strings.ToUpper(sequence)
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
	DefaultMinimalKmerCount int     = 10
	DefaultScoreThreshold   float64 = 0.5
)

type MegamashMap struct {
	Kmers                 map[string]string
	IdentifierToKmerCount map[string]int
	KmerSize              uint
	KmerMinimalCount      int
	Threshold             float64
}

// NewMegamashMap creates a megamash map that can be searched against.
func NewMegamashMap(sequences []fasta.Record, kmerSize uint, kmerMinimalCount int, threshold float64) (MegamashMap, error) {
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
			kmerString := StandardizeDNA(sequence[i : i+int(kmerSize)])
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
	identifierToCount := make(map[string]int)
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
	Identifier string  `json:"identifier"`
	Score      float64 `json:"score"`
}

// Match matches a sequence to all the sequences in a megamash map.
func (m *MegamashMap) Match(sequence string) []Match {
	identifierToCounts := make(map[string]uint)
	for identifier := range m.IdentifierToKmerCount {
		identifierToCounts[identifier] = 0
	}
	for i := 0; i <= len(sequence)-int(m.KmerSize); i++ {
		kmerString := StandardizeDNA(sequence[i : i+int(m.KmerSize)])
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

// MatchesToJSON converts a slice of Match structs to a JSON string.
func MatchesToJSON(matches []Match) (string, error) {
	jsonData, err := json.Marshal(matches)
	if err != nil {
		return "", err // Return an empty string and the error
	}
	return string(jsonData), nil // Convert byte slice to string and return
}

// JSONToMatches converts a JSON string to a slice of Match structs.
func JSONToMatches(jsonStr string) ([]Match, error) {
	var matches []Match
	err := json.Unmarshal([]byte(jsonStr), &matches)
	if err != nil {
		return nil, err // Return nil and the error
	}
	return matches, nil // Return the slice of matches
}
