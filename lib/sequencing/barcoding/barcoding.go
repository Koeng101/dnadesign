/*
Package barcoding contains functions to help barcode sequencing data.

Normally, external software handles barcoding for you (for example, Oxford
Nanopore's MinKNOW app). However, specialized indexing strategies can be
employed for custom barcoding methods.
*/

package barcoding

import (
	"encoding/csv"
	"io"
	"sort"
	"strings"

	"github.com/koeng101/dnadesign/lib/align"
	"github.com/koeng101/dnadesign/lib/align/matrix"
	"github.com/koeng101/dnadesign/lib/alphabet"
	"github.com/koeng101/dnadesign/lib/transform"
)

var ScoreMagicNumber = 18
var MinimalReadSize = 200
var EdgeCheckSize = 120

/******************************************************************************
Feb 12, 2024
                            Single barcodes

This is code for detecting single barcodes. Hasn't been given as much love as
dual barcoding code, and is mainly just copy pasted from there.

Keoni
******************************************************************************/

// SingleBarcodePrimerSet is a list of single barcode-to-sequence pairs in a
// convenient format for barcoding.
type SingleBarcodePrimerSet struct {
	BarcodeMap        map[string]string // barcode to sequence
	ReverseBarcodeMap map[string]string // sequence to barcode
}

// ParseSinglePrimerSet parses a csv file with barcode primer pairs into a
// SingleBarcodePrimerSet.
func ParseSinglePrimerSet(csvFile io.Reader) (SingleBarcodePrimerSet, error) {
	var result SingleBarcodePrimerSet
	result.BarcodeMap = make(map[string]string)
	result.ReverseBarcodeMap = make(map[string]string)

	// Create a new CSV reader reading from the input io.Reader
	reader := csv.NewReader(csvFile)

	for {
		// Read each record from csv
		record, err := reader.Read()
		// Break the loop at the end of the file
		if err == io.EOF {
			break
		}
		// Handle any other error
		if err != nil {
			return result, err
		}

		if len(record) == 2 {
			result.BarcodeMap[record[0]] = record[1]
			result.ReverseBarcodeMap[record[1]] = record[0]
		}
	}
	return result, nil
}

// SingleBarcode barcodes a sequence with a single barcode.
func SingleBarcode(sequence string, primerSet DualBarcodePrimerSet) (string, error) {
	m := [][]int{
		/*       A C G T U */
		/* A */ {1, -1, -1, -1, -1},
		/* C */ {-1, 1, -1, -1, -1},
		/* G */ {-1, -1, 1, -1, -1},
		/* T */ {-1, -1, -1, 1, -1},
		/* U */ {-1, -1, -1, -1, 1},
	}

	alphabet := alphabet.NewAlphabet([]string{"A", "C", "G", "T", "U"})
	subMatrix, _ := matrix.NewSubstitutionMatrix(alphabet, alphabet, m)
	scoring, _ := align.NewScoring(subMatrix, -1)

	if len(sequence) < MinimalReadSize {
		return "", nil
	}
	sequence = strings.ToUpper(sequence) // make sure sequence is upper case for the purposes of barcoding
	sequenceForward := sequence[:EdgeCheckSize]
	sequenceReverse := sequence[len(sequence)-EdgeCheckSize:]
	var topRanked string
	var topRankedScore int
	for _, sequence := range []string{sequenceForward, sequenceReverse} {
		for _, barcodeSequence := range primerSet.ForwardBarcodes {
			score, _, _, err := align.SmithWaterman(sequence, barcodeSequence, scoring)
			if err != nil {
				return "", err
			}
			complementScore, _, _, err := align.SmithWaterman(sequence, transform.ReverseComplement(barcodeSequence), scoring)
			if err != nil {
				return "", err
			}

			switch {
			case score > topRankedScore && score > ScoreMagicNumber:
				topRanked = barcodeSequence
				topRankedScore = score
			case complementScore > topRankedScore && complementScore > ScoreMagicNumber:
				topRanked = barcodeSequence
				topRankedScore = complementScore
			}
		}
	}
	_ = topRanked

	return topRanked, nil
}

/******************************************************************************
Feb 12, 2024
							Dual barcodes

When using Nanopore sequencing, I can barcode both sides of a given sequence.
Dual barcodes can encode a combinatorial quantity of potential sequences, so
are nice for barcoding lots of different DNA wells at once.

Keoni
******************************************************************************/

// DualBarcodePrimerSet represents a list of dual-barcoded wells in a
// convenient format.
type DualBarcodePrimerSet struct {
	BarcodeMap        map[string]DualBarcode
	ReverseBarcodeMap map[DualBarcode]string
	ForwardBarcodes   []string
	ReverseBarcodes   []string
}

// DualBarcode contains a forward and reverse barcode.
type DualBarcode struct {
	Forward string
	Reverse string
}

// ParseDualPrimerSet parses a csv file into a DualBarcodePrimerSet.
func ParseDualPrimerSet(csvFile io.Reader) (DualBarcodePrimerSet, error) {
	var result DualBarcodePrimerSet
	result.BarcodeMap = make(map[string]DualBarcode)
	result.ReverseBarcodeMap = make(map[DualBarcode]string)
	forwardBarcodesMap := make(map[string]bool)
	reverseBarcodesMap := make(map[string]bool)

	// Create a new CSV reader reading from the input io.Reader
	reader := csv.NewReader(csvFile)

	for {
		// Read each record from csv
		record, err := reader.Read()
		// Break the loop at the end of the file
		if err == io.EOF {
			break
		}
		// Handle any other error
		if err != nil {
			return result, err
		}
		if len(record) == 3 {
			well := record[0]
			forwardBarcode := record[1]
			reverseBarcode := record[2]

			forwardBarcodesMap[forwardBarcode] = true
			reverseBarcodesMap[reverseBarcode] = true
			newDualBarcode := DualBarcode{Forward: forwardBarcode, Reverse: reverseBarcode}
			result.BarcodeMap[well] = newDualBarcode
			result.ReverseBarcodeMap[newDualBarcode] = well
		}
	}

	// Convert maps to slices
	forwardBarcodes := make([]string, 0, len(forwardBarcodesMap))
	for barcode := range forwardBarcodesMap {
		forwardBarcodes = append(forwardBarcodes, barcode)
	}
	reverseBarcodes := make([]string, 0, len(reverseBarcodesMap))
	for barcode := range reverseBarcodesMap {
		reverseBarcodes = append(reverseBarcodes, barcode)
	}

	// Sort the slices
	sort.Strings(forwardBarcodes)
	sort.Strings(reverseBarcodes)

	// Append sorted barcodes to result
	result.ForwardBarcodes = forwardBarcodes
	result.ReverseBarcodes = reverseBarcodes

	return result, nil
}

// DualBarcodeSequence analyzes a sequence for both a forward and reverse
// barcode pair and returns their well.
func DualBarcodeSequence(sequence string, primerSet DualBarcodePrimerSet) (string, error) {
	m := [][]int{
		/*       A C G T U */
		/* A */ {1, -1, -1, -1, -1},
		/* C */ {-1, 1, -1, -1, -1},
		/* G */ {-1, -1, 1, -1, -1},
		/* T */ {-1, -1, -1, 1, -1},
		/* U */ {-1, -1, -1, -1, 1},
	}

	alphabet := alphabet.NewAlphabet([]string{"A", "C", "G", "T", "U"})
	subMatrix, _ := matrix.NewSubstitutionMatrix(alphabet, alphabet, m)
	scoring, _ := align.NewScoring(subMatrix, -1)

	if len(sequence) < MinimalReadSize {
		return "", nil
	}
	sequence = strings.ToUpper(sequence) // make sure sequence is upper case for the purposes of barcoding
	sequenceForward := sequence[:EdgeCheckSize]
	sequenceReverse := sequence[len(sequence)-EdgeCheckSize:]
	var topRankedForward string
	var topRankedForwardScore int
	var topRankedReverse string
	var topRankedReverseScore int
	// We want to check both ends for barcodes
	for _, sequence := range []string{sequenceForward, sequenceReverse} {
		for _, forwardBarcode := range primerSet.ForwardBarcodes {
			// We check both the barcode's sequence and its reverse complement.
			scoreFwd, _, _, err := align.SmithWaterman(sequence, forwardBarcode, scoring)
			if err != nil {
				return "", err
			}
			complementScoreFwd, _, _, err := align.SmithWaterman(sequence, transform.ReverseComplement(forwardBarcode), scoring)
			if err != nil {
				return "", err
			}

			// If the score is greater than any previous score and above the magic
			// number, replace the topRankedScore and continue on!
			switch {
			case scoreFwd > topRankedForwardScore && scoreFwd > ScoreMagicNumber:
				topRankedForward = forwardBarcode
				topRankedForwardScore = scoreFwd
			case complementScoreFwd > topRankedForwardScore && complementScoreFwd > ScoreMagicNumber:
				topRankedForward = forwardBarcode
				topRankedForwardScore = complementScoreFwd
			}
		}
		for _, reverseBarcode := range primerSet.ReverseBarcodes {
			scoreRev, _, _, err := align.SmithWaterman(sequence, reverseBarcode, scoring)
			if err != nil {
				return "", err
			}

			complementScoreRev, _, _, err := align.SmithWaterman(sequence, transform.ReverseComplement(reverseBarcode), scoring)
			if err != nil {
				return "", err
			}

			switch {
			case scoreRev > topRankedReverseScore && scoreRev > ScoreMagicNumber:
				topRankedReverse = reverseBarcode
				topRankedReverseScore = scoreRev
			case complementScoreRev > topRankedReverseScore && complementScoreRev > ScoreMagicNumber:
				topRankedReverse = reverseBarcode
				topRankedReverseScore = complementScoreRev
			}
		}
	}
	// Finally, get the well from the barcode map.
	well := primerSet.ReverseBarcodeMap[DualBarcode{Forward: topRankedForward, Reverse: topRankedReverse}]
	return well, nil
}
