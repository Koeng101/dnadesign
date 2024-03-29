/*
Package barcoding contains functions to help barcode sequencing data.

DNA barcoding is a strategy during DNA sequencing to correlate reads to certain
samples. Modern DNA sequencers sequence a lot of DNA, and often you'll want to
split up one DNA sequencer flow cell to sequence a bunch of different samples.
However, you need a way to tell the samples apart. This is usually done with
DNA barcodes - you attach a barcode to all the sequences from a certain sample,
then use software after sequencing to sort out barcoded samples.

Normally, vendor software handles barcoding for you (for example, Oxford
Nanopore's MinKNOW app). However, specialized barcoding strategies can be used
for more advanced sample processing.
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

// ScoreMagicNumber is the SmithWaterman score needed to match a barcode. 18 is
// chosen as a magic number that fits pretty well with normal 20bp barcodes.
var ScoreMagicNumber = 18

// MinimalReadSize is the minimal size of reads needed to barcode. This is
// mostly important for Nanopore reads, since amplicon Nanopore sequencing
// tends to generate quite a few small reads from PCR products that we do not
// want to barcode.
var MinimalReadSize = 200

// EdgeCheckSize is how long to check from the edge of both sides of a read.
// This is important because SmithWaterman alignment, used for aligning
// barcodes, is a quadratic algorithm. Nanopore sequencing prep, especially
// ligation preparation, can add DNA to either end of an amplicon. The native
// barcode (NB) and native adapter (NA) sequence may or may not be on both
// sides, adding ~80bp to the read.
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
	Barcodes          []string          // sorted for determinism
}

// ParseSinglePrimerSet parses a csv file with barcode primer pairs into a
// SingleBarcodePrimerSet.
func ParseSinglePrimerSet(csvFile io.Reader) (SingleBarcodePrimerSet, error) {
	var result SingleBarcodePrimerSet
	result.BarcodeMap = make(map[string]string)
	result.ReverseBarcodeMap = make(map[string]string)
	var barcodes []string

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
			barcodes = append(barcodes, record[1])
		}
	}

	// Sort the slices
	sort.Strings(barcodes)
	result.Barcodes = barcodes

	return result, nil
}

// SingleBarcodeSequence barcodes a sequence with a single barcode.
func SingleBarcodeSequence(sequence string, primerSet SingleBarcodePrimerSet) (string, error) {
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
		for _, barcodeSequence := range primerSet.Barcodes {
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
				topRanked = primerSet.ReverseBarcodeMap[barcodeSequence]
				topRankedScore = score
			case complementScore > topRankedScore && complementScore > ScoreMagicNumber:
				topRanked = primerSet.ReverseBarcodeMap[barcodeSequence]
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

I personally use this for plasmid sequencing.

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
	Name    string
	Forward string
	Reverse string
}

// DualBarcodesToPrimerSet parsers a list of dual barcodes into a dual barcode
// primer set.
func DualBarcodesToPrimerSet(dualBarcodes []DualBarcode) DualBarcodePrimerSet {
	var result DualBarcodePrimerSet
	result.BarcodeMap = make(map[string]DualBarcode)
	result.ReverseBarcodeMap = make(map[DualBarcode]string)
	forwardBarcodesMap := make(map[string]bool)
	reverseBarcodesMap := make(map[string]bool)

	for _, barcode := range dualBarcodes {
		forwardBarcodesMap[barcode.Forward] = true
		reverseBarcodesMap[barcode.Reverse] = true
		newDualBarcode := DualBarcode{Forward: barcode.Forward, Reverse: barcode.Reverse}
		result.BarcodeMap[barcode.Name] = newDualBarcode
		result.ReverseBarcodeMap[newDualBarcode] = barcode.Name
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

	return result
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
