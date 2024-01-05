/*
Package megamash is an implementation of the megamash algorithm.

Megamash is an algorithm developed by Keoni Gandall to find templates from
sequencing reactions. For example, you may have a pool of amplicons, and need
to get a count of how many times each amplicon shows up in a given sequencing
reaction.
*/
package megamash

import (
	"encoding/base64"

	"github.com/koeng101/dnadesign/lib/transform"
)

func StandardizedCompressedDNA(sequence string) []byte {
	var deterministicSequence string
	reverseComplement := transform.ReverseComplement(sequence)
	if sequence > reverseComplement {
		deterministicSequence = reverseComplement
	} else {
		deterministicSequence = sequence
	}
	return CompressDNA(deterministicSequence)

}

type MegamashMap []map[string]bool

func MakeMegamashMap(sequences []string, kmerSize uint) MegamashMap {
	var megamashMap MegamashMap
	for _, sequence := range sequences {
		// First get all kmers with a given sequence
		kmerMap := make(map[string]bool)
		for i := 0; i <= len(sequence)-int(kmerSize); i++ {
			kmerBytes := StandardizedCompressedDNA(sequence[i : i+int(kmerSize)])
			kmerBase64 := base64.StdEncoding.EncodeToString(kmerBytes)
			kmerMap[kmerBase64] = true
		}

		// Then, get unique kmers for this sequence and only this sequence
		uniqueKmerMap := make(map[string]bool)
		for kmerBase64, _ := range kmerMap {
			unique := true
			for _, otherMegaMashMap := range megamashMap {
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

		// Now we have a unique Kmer map for the given sequence.
		// Add it to megamashMap
		megamashMap = append(megamashMap, uniqueKmerMap)
	}
	// Finally, go back through and make a final megamashMap without
	// all those falses.
	var finalMegamashMap MegamashMap
	for _, singleMegamashMap := range megamashMap {
		finalMap := make(map[string]bool)
		for kmerBase64, value := range singleMegamashMap {
			if value {
				finalMap[kmerBase64] = true
			}
		}
		finalMegamashMap = append(finalMegamashMap, finalMap)
	}
	return finalMegamashMap
}

func (m *MegamashMap) Score(sequence string) []float64 {
	var scores []float64
	// The algorithm is as follows:
	// - Go through each map.
	// - Get the number of matching kmers
	// - Divide that by the total kmers available for matching

	// First, get the kmer total
	var kmerSize int
out:
	for _, maps := range *m {
		for kmer, _ := range maps {
			decodedBytes, _ := base64.StdEncoding.DecodeString(kmer)
			sequence := DecompressDNA(decodedBytes)
			kmerSize = len(sequence)
			break out
		}
	}

	// Now, iterate through each map
	for _, sequenceMap := range *m {
		var score float64
		var totalKmers int = len(sequenceMap)
		var matchedKmers int
		for i := 0; i <= len(sequence)-int(kmerSize); i++ {
			kmerBytes := StandardizedCompressedDNA(sequence[i : i+int(kmerSize)])
			kmerBase64 := base64.StdEncoding.EncodeToString(kmerBytes)
			_, ok := sequenceMap[kmerBase64]
			if ok {
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
	return scores
}
