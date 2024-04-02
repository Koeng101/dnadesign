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

// ScoreMagicNumber is the SmithWaterman score needed to match a barcode. 28 is
// chosen as a magic number that fits pretty well with normal 40bp barcodes.
// (20bp primer + 20bp barcode)
var ScoreMagicNumber = 28

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
func DualBarcodeSequence(sequence string, forwardPrimer string, reversePrimer string, primerSet DualBarcodePrimerSet) (string, error) {
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
	forwardPrimer = strings.ToUpper(forwardPrimer)
	reversePrimer = strings.ToUpper(reversePrimer)
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
			scoreFwd, _, _, err := align.SmithWaterman(sequence, forwardBarcode+forwardPrimer, scoring)
			if err != nil {
				return "", err
			}
			complementScoreFwd, _, _, err := align.SmithWaterman(sequence, transform.ReverseComplement(forwardBarcode+forwardPrimer), scoring)
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
			scoreRev, _, _, err := align.SmithWaterman(sequence, reverseBarcode+reversePrimer, scoring)
			if err != nil {
				return "", err
			}

			complementScoreRev, _, _, err := align.SmithWaterman(sequence, transform.ReverseComplement(reverseBarcode+reversePrimer), scoring)
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
	// This is to handle the case that we did not find a barcode
	// Finally, get the well from the barcode map.
	well := primerSet.ReverseBarcodeMap[DualBarcode{Forward: topRankedForward, Reverse: topRankedReverse}]
	return well, nil
}

var DefaultBarcodes = DualBarcodesToPrimerSet([]DualBarcode{
	{"A1", "GGAAAAGCAAAACTAAAACG", "GAACCACAACCTTAACCTGA"},
	{"C1", "GAAGGACAAGGTTAAGGTGA", "GAACCACAACCTTAACCTGA"},
	{"E1", "GGTGAAGGTCAAGGGTAAGG", "GAACCACAACCTTAACCTGA"},
	{"G1", "TAAGGGGAAGGGCAAGGCTA", "GAACCACAACCTTAACCTGA"},
	{"I1", "GGCTAAGGCGAAGGCCAAGC", "GAACCACAACCTTAACCTGA"},
	{"K1", "CAAGCATAAGCAGAAGCACA", "GAACCACAACCTTAACCTGA"},
	{"M1", "GCACAAGCTTAAGCTGAAGC", "GAACCACAACCTTAACCTGA"},
	{"O1", "GAAGCTCAAGCGTAAGCGGA", "GAACCACAACCTTAACCTGA"},
	{"A2", "GGAAAAGCAAAACTAAAACG", "CCTGAACCTCAACCGTAACC"},
	{"C2", "GAAGGACAAGGTTAAGGTGA", "CCTGAACCTCAACCGTAACC"},
	{"E2", "GGTGAAGGTCAAGGGTAAGG", "CCTGAACCTCAACCGTAACC"},
	{"G2", "TAAGGGGAAGGGCAAGGCTA", "CCTGAACCTCAACCGTAACC"},
	{"I2", "GGCTAAGGCGAAGGCCAAGC", "CCTGAACCTCAACCGTAACC"},
	{"K2", "CAAGCATAAGCAGAAGCACA", "CCTGAACCTCAACCGTAACC"},
	{"M2", "GCACAAGCTTAAGCTGAAGC", "CCTGAACCTCAACCGTAACC"},
	{"O2", "GAAGCTCAAGCGTAAGCGGA", "CCTGAACCTCAACCGTAACC"},
	{"A3", "GGAAAAGCAAAACTAAAACG", "TAACCGGAACCGCAACCCTA"},
	{"C3", "GAAGGACAAGGTTAAGGTGA", "TAACCGGAACCGCAACCCTA"},
	{"E3", "GGTGAAGGTCAAGGGTAAGG", "TAACCGGAACCGCAACCCTA"},
	{"G3", "TAAGGGGAAGGGCAAGGCTA", "TAACCGGAACCGCAACCCTA"},
	{"I3", "GGCTAAGGCGAAGGCCAAGC", "TAACCGGAACCGCAACCCTA"},
	{"K3", "CAAGCATAAGCAGAAGCACA", "TAACCGGAACCGCAACCCTA"},
	{"M3", "GCACAAGCTTAAGCTGAAGC", "TAACCGGAACCGCAACCCTA"},
	{"O3", "GAAGCTCAAGCGTAAGCGGA", "TAACCGGAACCGCAACCCTA"},
	{"A4", "GGAAAAGCAAAACTAAAACG", "CCCTAACCCGAACCCCATAT"},
	{"C4", "GAAGGACAAGGTTAAGGTGA", "CCCTAACCCGAACCCCATAT"},
	{"E4", "GGTGAAGGTCAAGGGTAAGG", "CCCTAACCCGAACCCCATAT"},
	{"G4", "TAAGGGGAAGGGCAAGGCTA", "CCCTAACCCGAACCCCATAT"},
	{"I4", "GGCTAAGGCGAAGGCCAAGC", "CCCTAACCCGAACCCCATAT"},
	{"K4", "CAAGCATAAGCAGAAGCACA", "CCCTAACCCGAACCCCATAT"},
	{"M4", "GCACAAGCTTAAGCTGAAGC", "CCCTAACCCGAACCCCATAT"},
	{"O4", "GAAGCTCAAGCGTAAGCGGA", "CCCTAACCCGAACCCCATAT"},
	{"A5", "GGAAAAGCAAAACTAAAACG", "GGATATGCATATCTATATCG"},
	{"C5", "GAAGGACAAGGTTAAGGTGA", "GGATATGCATATCTATATCG"},
	{"E5", "GGTGAAGGTCAAGGGTAAGG", "GGATATGCATATCTATATCG"},
	{"G5", "TAAGGGGAAGGGCAAGGCTA", "GGATATGCATATCTATATCG"},
	{"I5", "GGCTAAGGCGAAGGCCAAGC", "GGATATGCATATCTATATCG"},
	{"K5", "CAAGCATAAGCAGAAGCACA", "GGATATGCATATCTATATCG"},
	{"M5", "GCACAAGCTTAAGCTGAAGC", "GGATATGCATATCTATATCG"},
	{"O5", "GAAGCTCAAGCGTAAGCGGA", "GGATATGCATATCTATATCG"},
	{"A6", "GGAAAAGCAAAACTAAAACG", "CGTATGCGGATGCGCATGCC"},
	{"C6", "GAAGGACAAGGTTAAGGTGA", "CGTATGCGGATGCGCATGCC"},
	{"E6", "GGTGAAGGTCAAGGGTAAGG", "CGTATGCGGATGCGCATGCC"},
	{"G6", "TAAGGGGAAGGGCAAGGCTA", "CGTATGCGGATGCGCATGCC"},
	{"I6", "GGCTAAGGCGAAGGCCAAGC", "CGTATGCGGATGCGCATGCC"},
	{"K6", "CAAGCATAAGCAGAAGCACA", "CGTATGCGGATGCGCATGCC"},
	{"M6", "GCACAAGCTTAAGCTGAAGC", "CGTATGCGGATGCGCATGCC"},
	{"O6", "GAAGCTCAAGCGTAAGCGGA", "CGTATGCGGATGCGCATGCC"},
	{"A7", "GGAAAAGCAAAACTAAAACG", "ATGCCTATGCCGATGCCCAT"},
	{"C7", "GAAGGACAAGGTTAAGGTGA", "ATGCCTATGCCGATGCCCAT"},
	{"E7", "GGTGAAGGTCAAGGGTAAGG", "ATGCCTATGCCGATGCCCAT"},
	{"G7", "TAAGGGGAAGGGCAAGGCTA", "ATGCCTATGCCGATGCCCAT"},
	{"I7", "GGCTAAGGCGAAGGCCAAGC", "ATGCCTATGCCGATGCCCAT"},
	{"K7", "CAAGCATAAGCAGAAGCACA", "ATGCCTATGCCGATGCCCAT"},
	{"M7", "GCACAAGCTTAAGCTGAAGC", "ATGCCTATGCCGATGCCCAT"},
	{"O7", "GAAGCTCAAGCGTAAGCGGA", "ATGCCTATGCCGATGCCCAT"},
	{"A8", "GGAAAAGCAAAACTAAAACG", "CCCATCATCAGTATCAGGAT"},
	{"C8", "GAAGGACAAGGTTAAGGTGA", "CCCATCATCAGTATCAGGAT"},
	{"E8", "GGTGAAGGTCAAGGGTAAGG", "CCCATCATCAGTATCAGGAT"},
	{"G8", "TAAGGGGAAGGGCAAGGCTA", "CCCATCATCAGTATCAGGAT"},
	{"I8", "GGCTAAGGCGAAGGCCAAGC", "CCCATCATCAGTATCAGGAT"},
	{"K8", "CAAGCATAAGCAGAAGCACA", "CCCATCATCAGTATCAGGAT"},
	{"M8", "GCACAAGCTTAAGCTGAAGC", "CCCATCATCAGTATCAGGAT"},
	{"O8", "GAAGCTCAAGCGTAAGCGGA", "CCCATCATCAGTATCAGGAT"},
	{"A9", "GGAAAAGCAAAACTAAAACG", "AGGATCAGCATCACTATCAC"},
	{"C9", "GAAGGACAAGGTTAAGGTGA", "AGGATCAGCATCACTATCAC"},
	{"E9", "GGTGAAGGTCAAGGGTAAGG", "AGGATCAGCATCACTATCAC"},
	{"G9", "TAAGGGGAAGGGCAAGGCTA", "AGGATCAGCATCACTATCAC"},
	{"I9", "GGCTAAGGCGAAGGCCAAGC", "AGGATCAGCATCACTATCAC"},
	{"K9", "CAAGCATAAGCAGAAGCACA", "AGGATCAGCATCACTATCAC"},
	{"M9", "GCACAAGCTTAAGCTGAAGC", "AGGATCAGCATCACTATCAC"},
	{"O9", "GAAGCTCAAGCGTAAGCGGA", "AGGATCAGCATCACTATCAC"},
	{"A10", "GGAAAAGCAAAACTAAAACG", "ATCACGATCACCATCTAGAT"},
	{"C10", "GAAGGACAAGGTTAAGGTGA", "ATCACGATCACCATCTAGAT"},
	{"E10", "GGTGAAGGTCAAGGGTAAGG", "ATCACGATCACCATCTAGAT"},
	{"G10", "TAAGGGGAAGGGCAAGGCTA", "ATCACGATCACCATCTAGAT"},
	{"I10", "GGCTAAGGCGAAGGCCAAGC", "ATCACGATCACCATCTAGAT"},
	{"K10", "CAAGCATAAGCAGAAGCACA", "ATCACGATCACCATCTAGAT"},
	{"M10", "GCACAAGCTTAAGCTGAAGC", "ATCACGATCACCATCTAGAT"},
	{"O10", "GAAGCTCAAGCGTAAGCGGA", "ATCACGATCACCATCTAGAT"},
	{"A11", "GGAAAAGCAAAACTAAAACG", "TCTTGATCTTCATCTGTATC"},
	{"C11", "GAAGGACAAGGTTAAGGTGA", "TCTTGATCTTCATCTGTATC"},
	{"E11", "GGTGAAGGTCAAGGGTAAGG", "TCTTGATCTTCATCTGTATC"},
	{"G11", "TAAGGGGAAGGGCAAGGCTA", "TCTTGATCTTCATCTGTATC"},
	{"I11", "GGCTAAGGCGAAGGCCAAGC", "TCTTGATCTTCATCTGTATC"},
	{"K11", "CAAGCATAAGCAGAAGCACA", "TCTTGATCTTCATCTGTATC"},
	{"M11", "GCACAAGCTTAAGCTGAAGC", "TCTTGATCTTCATCTGTATC"},
	{"O11", "GAAGCTCAAGCGTAAGCGGA", "TCTTGATCTTCATCTGTATC"},
	{"A12", "GGAAAAGCAAAACTAAAACG", "CCAGACACAGACTTAGACTG"},
	{"C12", "GAAGGACAAGGTTAAGGTGA", "CCAGACACAGACTTAGACTG"},
	{"E12", "GGTGAAGGTCAAGGGTAAGG", "CCAGACACAGACTTAGACTG"},
	{"G12", "TAAGGGGAAGGGCAAGGCTA", "CCAGACACAGACTTAGACTG"},
	{"I12", "GGCTAAGGCGAAGGCCAAGC", "CCAGACACAGACTTAGACTG"},
	{"K12", "CAAGCATAAGCAGAAGCACA", "CCAGACACAGACTTAGACTG"},
	{"M12", "GCACAAGCTTAAGCTGAAGC", "CCAGACACAGACTTAGACTG"},
	{"O12", "GAAGCTCAAGCGTAAGCGGA", "CCAGACACAGACTTAGACTG"},
	{"A13", "GGAAAAGCAAAACTAAAACG", "GACTGAGACTCAGACGTAGA"},
	{"C13", "GAAGGACAAGGTTAAGGTGA", "GACTGAGACTCAGACGTAGA"},
	{"E13", "GGTGAAGGTCAAGGGTAAGG", "GACTGAGACTCAGACGTAGA"},
	{"G13", "TAAGGGGAAGGGCAAGGCTA", "GACTGAGACTCAGACGTAGA"},
	{"I13", "GGCTAAGGCGAAGGCCAAGC", "GACTGAGACTCAGACGTAGA"},
	{"K13", "CAAGCATAAGCAGAAGCACA", "GACTGAGACTCAGACGTAGA"},
	{"M13", "GCACAAGCTTAAGCTGAAGC", "GACTGAGACTCAGACGTAGA"},
	{"O13", "GAAGCTCAAGCGTAAGCGGA", "GACTGAGACTCAGACGTAGA"},
	{"A14", "GGAAAAGCAAAACTAAAACG", "GTAGACGGAGACGCAGACCT"},
	{"C14", "GAAGGACAAGGTTAAGGTGA", "GTAGACGGAGACGCAGACCT"},
	{"E14", "GGTGAAGGTCAAGGGTAAGG", "GTAGACGGAGACGCAGACCT"},
	{"G14", "TAAGGGGAAGGGCAAGGCTA", "GTAGACGGAGACGCAGACCT"},
	{"I14", "GGCTAAGGCGAAGGCCAAGC", "GTAGACGGAGACGCAGACCT"},
	{"K14", "CAAGCATAAGCAGAAGCACA", "GTAGACGGAGACGCAGACCT"},
	{"M14", "GCACAAGCTTAAGCTGAAGC", "GTAGACGGAGACGCAGACCT"},
	{"O14", "GAAGCTCAAGCGTAAGCGGA", "GTAGACGGAGACGCAGACCT"},
	{"A15", "GGAAAAGCAAAACTAAAACG", "GACCTAGACCGAGACCCAGT"},
	{"C15", "GAAGGACAAGGTTAAGGTGA", "GACCTAGACCGAGACCCAGT"},
	{"E15", "GGTGAAGGTCAAGGGTAAGG", "GACCTAGACCGAGACCCAGT"},
	{"G15", "TAAGGGGAAGGGCAAGGCTA", "GACCTAGACCGAGACCCAGT"},
	{"I15", "GGCTAAGGCGAAGGCCAAGC", "GACCTAGACCGAGACCCAGT"},
	{"K15", "CAAGCATAAGCAGAAGCACA", "GACCTAGACCGAGACCCAGT"},
	{"M15", "GCACAAGCTTAAGCTGAAGC", "GACCTAGACCGAGACCCAGT"},
	{"O15", "GAAGCTCAAGCGTAAGCGGA", "GACCTAGACCGAGACCCAGT"},
	{"A16", "GGAAAAGCAAAACTAAAACG", "CCAGTAGTAGGAGTAGCAGT"},
	{"C16", "GAAGGACAAGGTTAAGGTGA", "CCAGTAGTAGGAGTAGCAGT"},
	{"E16", "GGTGAAGGTCAAGGGTAAGG", "CCAGTAGTAGGAGTAGCAGT"},
	{"G16", "TAAGGGGAAGGGCAAGGCTA", "CCAGTAGTAGGAGTAGCAGT"},
	{"I16", "GGCTAAGGCGAAGGCCAAGC", "CCAGTAGTAGGAGTAGCAGT"},
	{"K16", "CAAGCATAAGCAGAAGCACA", "CCAGTAGTAGGAGTAGCAGT"},
	{"M16", "GCACAAGCTTAAGCTGAAGC", "CCAGTAGTAGGAGTAGCAGT"},
	{"O16", "GAAGCTCAAGCGTAAGCGGA", "CCAGTAGTAGGAGTAGCAGT"},
	{"A17", "GGAAAAGCAAAACTAAAACG", "GCAGTACTAGTACGAGTACC"},
	{"C17", "GAAGGACAAGGTTAAGGTGA", "GCAGTACTAGTACGAGTACC"},
	{"E17", "GGTGAAGGTCAAGGGTAAGG", "GCAGTACTAGTACGAGTACC"},
	{"G17", "TAAGGGGAAGGGCAAGGCTA", "GCAGTACTAGTACGAGTACC"},
	{"I17", "GGCTAAGGCGAAGGCCAAGC", "GCAGTACTAGTACGAGTACC"},
	{"K17", "CAAGCATAAGCAGAAGCACA", "GCAGTACTAGTACGAGTACC"},
	{"M17", "GCACAAGCTTAAGCTGAAGC", "GCAGTACTAGTACGAGTACC"},
	{"O17", "GAAGCTCAAGCGTAAGCGGA", "GCAGTACTAGTACGAGTACC"},
	{"A18", "GGAAAAGCAAAACTAAAACG", "GTACCAGTTACAGTTTTAGT"},
	{"C18", "GAAGGACAAGGTTAAGGTGA", "GTACCAGTTACAGTTTTAGT"},
	{"E18", "GGTGAAGGTCAAGGGTAAGG", "GTACCAGTTACAGTTTTAGT"},
	{"G18", "TAAGGGGAAGGGCAAGGCTA", "GTACCAGTTACAGTTTTAGT"},
	{"I18", "GGCTAAGGCGAAGGCCAAGC", "GTACCAGTTACAGTTTTAGT"},
	{"K18", "CAAGCATAAGCAGAAGCACA", "GTACCAGTTACAGTTTTAGT"},
	{"M18", "GCACAAGCTTAAGCTGAAGC", "GTACCAGTTACAGTTTTAGT"},
	{"O18", "GAAGCTCAAGCGTAAGCGGA", "GTACCAGTTACAGTTTTAGT"},
	{"A19", "GGAAAAGCAAAACTAAAACG", "AGTTTGAGTTTCAGTTGTAG"},
	{"C19", "GAAGGACAAGGTTAAGGTGA", "AGTTTGAGTTTCAGTTGTAG"},
	{"E19", "GGTGAAGGTCAAGGGTAAGG", "AGTTTGAGTTTCAGTTGTAG"},
	{"G19", "TAAGGGGAAGGGCAAGGCTA", "AGTTTGAGTTTCAGTTGTAG"},
	{"I19", "GGCTAAGGCGAAGGCCAAGC", "AGTTTGAGTTTCAGTTGTAG"},
	{"K19", "CAAGCATAAGCAGAAGCACA", "AGTTTGAGTTTCAGTTGTAG"},
	{"M19", "GCACAAGCTTAAGCTGAAGC", "AGTTTGAGTTTCAGTTGTAG"},
	{"O19", "GAAGCTCAAGCGTAAGCGGA", "AGTTTGAGTTTCAGTTGTAG"},
	{"A20", "GGAAAAGCAAAACTAAAACG", "GTTCCAGTGACAGTGTTAGT"},
	{"C20", "GAAGGACAAGGTTAAGGTGA", "GTTCCAGTGACAGTGTTAGT"},
	{"E20", "GGTGAAGGTCAAGGGTAAGG", "GTTCCAGTGACAGTGTTAGT"},
	{"G20", "TAAGGGGAAGGGCAAGGCTA", "GTTCCAGTGACAGTGTTAGT"},
	{"I20", "GGCTAAGGCGAAGGCCAAGC", "GTTCCAGTGACAGTGTTAGT"},
	{"K20", "CAAGCATAAGCAGAAGCACA", "GTTCCAGTGACAGTGTTAGT"},
	{"M20", "GCACAAGCTTAAGCTGAAGC", "GTTCCAGTGACAGTGTTAGT"},
	{"O20", "GAAGCTCAAGCGTAAGCGGA", "GTTCCAGTGACAGTGTTAGT"},
	{"A21", "GGAAAAGCAAAACTAAAACG", "TTAGTGTGAGTGTCAGTGGT"},
	{"C21", "GAAGGACAAGGTTAAGGTGA", "TTAGTGTGAGTGTCAGTGGT"},
	{"E21", "GGTGAAGGTCAAGGGTAAGG", "TTAGTGTGAGTGTCAGTGGT"},
	{"G21", "TAAGGGGAAGGGCAAGGCTA", "TTAGTGTGAGTGTCAGTGGT"},
	{"I21", "GGCTAAGGCGAAGGCCAAGC", "TTAGTGTGAGTGTCAGTGGT"},
	{"K21", "CAAGCATAAGCAGAAGCACA", "TTAGTGTGAGTGTCAGTGGT"},
	{"M21", "GCACAAGCTTAAGCTGAAGC", "TTAGTGTGAGTGTCAGTGGT"},
	{"O21", "GAAGCTCAAGCGTAAGCGGA", "TTAGTGTGAGTGTCAGTGGT"},
	{"A22", "GGAAAAGCAAAACTAAAACG", "GTGGTAGTGGGAGTGGCAGT"},
	{"C22", "GAAGGACAAGGTTAAGGTGA", "GTGGTAGTGGGAGTGGCAGT"},
	{"E22", "GGTGAAGGTCAAGGGTAAGG", "GTGGTAGTGGGAGTGGCAGT"},
	{"G22", "TAAGGGGAAGGGCAAGGCTA", "GTGGTAGTGGGAGTGGCAGT"},
	{"I22", "GGCTAAGGCGAAGGCCAAGC", "GTGGTAGTGGGAGTGGCAGT"},
	{"K22", "CAAGCATAAGCAGAAGCACA", "GTGGTAGTGGGAGTGGCAGT"},
	{"M22", "GCACAAGCTTAAGCTGAAGC", "GTGGTAGTGGGAGTGGCAGT"},
	{"O22", "GAAGCTCAAGCGTAAGCGGA", "GTGGTAGTGGGAGTGGCAGT"},
	{"A23", "GGAAAAGCAAAACTAAAACG", "GCAGTGCTAGTGCGAGTGCC"},
	{"C23", "GAAGGACAAGGTTAAGGTGA", "GCAGTGCTAGTGCGAGTGCC"},
	{"E23", "GGTGAAGGTCAAGGGTAAGG", "GCAGTGCTAGTGCGAGTGCC"},
	{"G23", "TAAGGGGAAGGGCAAGGCTA", "GCAGTGCTAGTGCGAGTGCC"},
	{"I23", "GGCTAAGGCGAAGGCCAAGC", "GCAGTGCTAGTGCGAGTGCC"},
	{"K23", "CAAGCATAAGCAGAAGCACA", "GCAGTGCTAGTGCGAGTGCC"},
	{"M23", "GCACAAGCTTAAGCTGAAGC", "GCAGTGCTAGTGCGAGTGCC"},
	{"O23", "GAAGCTCAAGCGTAAGCGGA", "GCAGTGCTAGTGCGAGTGCC"},
	{"A24", "GGAAAAGCAAAACTAAAACG", "GTGCCAGTCACAGTCTTAGT"},
	{"C24", "GAAGGACAAGGTTAAGGTGA", "GTGCCAGTCACAGTCTTAGT"},
	{"E24", "GGTGAAGGTCAAGGGTAAGG", "GTGCCAGTCACAGTCTTAGT"},
	{"G24", "TAAGGGGAAGGGCAAGGCTA", "GTGCCAGTCACAGTCTTAGT"},
	{"I24", "GGCTAAGGCGAAGGCCAAGC", "GTGCCAGTCACAGTCTTAGT"},
	{"K24", "CAAGCATAAGCAGAAGCACA", "GTGCCAGTCACAGTCTTAGT"},
	{"M24", "GCACAAGCTTAAGCTGAAGC", "GTGCCAGTCACAGTCTTAGT"},
	{"O24", "GAAGCTCAAGCGTAAGCGGA", "GTGCCAGTCACAGTCTTAGT"},
	{"B1", "GCGGAAGCGCAAGCCTAAGC", "GAACCACAACCTTAACCTGA"},
	{"D1", "TAAGCCGAAGCCCAACAACA", "GAACCACAACCTTAACCTGA"},
	{"F1", "ACATGAACATCAACAGTAAC", "GAACCACAACCTTAACCTGA"},
	{"H1", "CGAGAACGACAACGTTAACG", "GAACCACAACCTTAACCTGA"},
	{"J1", "TAACGTGAACGTCAACGGTA", "GAACCACAACCTTAACCTGA"},
	{"L1", "CGGTAACGGGAACGGCAACG", "GAACCACAACCTTAACCTGA"},
	{"N1", "CAACGCTAACGCGAACGCCA", "GAACCACAACCTTAACCTGA"},
	{"P1", "CGCCAACCATAACCAGAACC", "GAACCACAACCTTAACCTGA"},
	{"B2", "GCGGAAGCGCAAGCCTAAGC", "CCTGAACCTCAACCGTAACC"},
	{"D2", "TAAGCCGAAGCCCAACAACA", "CCTGAACCTCAACCGTAACC"},
	{"F2", "ACATGAACATCAACAGTAAC", "CCTGAACCTCAACCGTAACC"},
	{"H2", "CGAGAACGACAACGTTAACG", "CCTGAACCTCAACCGTAACC"},
	{"J2", "TAACGTGAACGTCAACGGTA", "CCTGAACCTCAACCGTAACC"},
	{"L2", "CGGTAACGGGAACGGCAACG", "CCTGAACCTCAACCGTAACC"},
	{"N2", "CAACGCTAACGCGAACGCCA", "CCTGAACCTCAACCGTAACC"},
	{"P2", "CGCCAACCATAACCAGAACC", "CCTGAACCTCAACCGTAACC"},
	{"B3", "GCGGAAGCGCAAGCCTAAGC", "TAACCGGAACCGCAACCCTA"},
	{"D3", "TAAGCCGAAGCCCAACAACA", "TAACCGGAACCGCAACCCTA"},
	{"F3", "ACATGAACATCAACAGTAAC", "TAACCGGAACCGCAACCCTA"},
	{"H3", "CGAGAACGACAACGTTAACG", "TAACCGGAACCGCAACCCTA"},
	{"J3", "TAACGTGAACGTCAACGGTA", "TAACCGGAACCGCAACCCTA"},
	{"L3", "CGGTAACGGGAACGGCAACG", "TAACCGGAACCGCAACCCTA"},
	{"N3", "CAACGCTAACGCGAACGCCA", "TAACCGGAACCGCAACCCTA"},
	{"P3", "CGCCAACCATAACCAGAACC", "TAACCGGAACCGCAACCCTA"},
	{"B4", "GCGGAAGCGCAAGCCTAAGC", "CCCTAACCCGAACCCCATAT"},
	{"D4", "TAAGCCGAAGCCCAACAACA", "CCCTAACCCGAACCCCATAT"},
	{"F4", "ACATGAACATCAACAGTAAC", "CCCTAACCCGAACCCCATAT"},
	{"H4", "CGAGAACGACAACGTTAACG", "CCCTAACCCGAACCCCATAT"},
	{"J4", "TAACGTGAACGTCAACGGTA", "CCCTAACCCGAACCCCATAT"},
	{"L4", "CGGTAACGGGAACGGCAACG", "CCCTAACCCGAACCCCATAT"},
	{"N4", "CAACGCTAACGCGAACGCCA", "CCCTAACCCGAACCCCATAT"},
	{"P4", "CGCCAACCATAACCAGAACC", "CCCTAACCCGAACCCCATAT"},
	{"B5", "GCGGAAGCGCAAGCCTAAGC", "GGATATGCATATCTATATCG"},
	{"D5", "TAAGCCGAAGCCCAACAACA", "GGATATGCATATCTATATCG"},
	{"F5", "ACATGAACATCAACAGTAAC", "GGATATGCATATCTATATCG"},
	{"H5", "CGAGAACGACAACGTTAACG", "GGATATGCATATCTATATCG"},
	{"J5", "TAACGTGAACGTCAACGGTA", "GGATATGCATATCTATATCG"},
	{"L5", "CGGTAACGGGAACGGCAACG", "GGATATGCATATCTATATCG"},
	{"N5", "CAACGCTAACGCGAACGCCA", "GGATATGCATATCTATATCG"},
	{"P5", "CGCCAACCATAACCAGAACC", "GGATATGCATATCTATATCG"},
	{"B6", "GCGGAAGCGCAAGCCTAAGC", "CGTATGCGGATGCGCATGCC"},
	{"D6", "TAAGCCGAAGCCCAACAACA", "CGTATGCGGATGCGCATGCC"},
	{"F6", "ACATGAACATCAACAGTAAC", "CGTATGCGGATGCGCATGCC"},
	{"H6", "CGAGAACGACAACGTTAACG", "CGTATGCGGATGCGCATGCC"},
	{"J6", "TAACGTGAACGTCAACGGTA", "CGTATGCGGATGCGCATGCC"},
	{"L6", "CGGTAACGGGAACGGCAACG", "CGTATGCGGATGCGCATGCC"},
	{"N6", "CAACGCTAACGCGAACGCCA", "CGTATGCGGATGCGCATGCC"},
	{"P6", "CGCCAACCATAACCAGAACC", "CGTATGCGGATGCGCATGCC"},
	{"B7", "GCGGAAGCGCAAGCCTAAGC", "ATGCCTATGCCGATGCCCAT"},
	{"D7", "TAAGCCGAAGCCCAACAACA", "ATGCCTATGCCGATGCCCAT"},
	{"F7", "ACATGAACATCAACAGTAAC", "ATGCCTATGCCGATGCCCAT"},
	{"H7", "CGAGAACGACAACGTTAACG", "ATGCCTATGCCGATGCCCAT"},
	{"J7", "TAACGTGAACGTCAACGGTA", "ATGCCTATGCCGATGCCCAT"},
	{"L7", "CGGTAACGGGAACGGCAACG", "ATGCCTATGCCGATGCCCAT"},
	{"N7", "CAACGCTAACGCGAACGCCA", "ATGCCTATGCCGATGCCCAT"},
	{"P7", "CGCCAACCATAACCAGAACC", "ATGCCTATGCCGATGCCCAT"},
	{"B8", "GCGGAAGCGCAAGCCTAAGC", "CCCATCATCAGTATCAGGAT"},
	{"D8", "TAAGCCGAAGCCCAACAACA", "CCCATCATCAGTATCAGGAT"},
	{"F8", "ACATGAACATCAACAGTAAC", "CCCATCATCAGTATCAGGAT"},
	{"H8", "CGAGAACGACAACGTTAACG", "CCCATCATCAGTATCAGGAT"},
	{"J8", "TAACGTGAACGTCAACGGTA", "CCCATCATCAGTATCAGGAT"},
	{"L8", "CGGTAACGGGAACGGCAACG", "CCCATCATCAGTATCAGGAT"},
	{"N8", "CAACGCTAACGCGAACGCCA", "CCCATCATCAGTATCAGGAT"},
	{"P8", "CGCCAACCATAACCAGAACC", "CCCATCATCAGTATCAGGAT"},
	{"B9", "GCGGAAGCGCAAGCCTAAGC", "AGGATCAGCATCACTATCAC"},
	{"D9", "TAAGCCGAAGCCCAACAACA", "AGGATCAGCATCACTATCAC"},
	{"F9", "ACATGAACATCAACAGTAAC", "AGGATCAGCATCACTATCAC"},
	{"H9", "CGAGAACGACAACGTTAACG", "AGGATCAGCATCACTATCAC"},
	{"J9", "TAACGTGAACGTCAACGGTA", "AGGATCAGCATCACTATCAC"},
	{"L9", "CGGTAACGGGAACGGCAACG", "AGGATCAGCATCACTATCAC"},
	{"N9", "CAACGCTAACGCGAACGCCA", "AGGATCAGCATCACTATCAC"},
	{"P9", "CGCCAACCATAACCAGAACC", "AGGATCAGCATCACTATCAC"},
	{"B10", "GCGGAAGCGCAAGCCTAAGC", "ATCACGATCACCATCTAGAT"},
	{"D10", "TAAGCCGAAGCCCAACAACA", "ATCACGATCACCATCTAGAT"},
	{"F10", "ACATGAACATCAACAGTAAC", "ATCACGATCACCATCTAGAT"},
	{"H10", "CGAGAACGACAACGTTAACG", "ATCACGATCACCATCTAGAT"},
	{"J10", "TAACGTGAACGTCAACGGTA", "ATCACGATCACCATCTAGAT"},
	{"L10", "CGGTAACGGGAACGGCAACG", "ATCACGATCACCATCTAGAT"},
	{"N10", "CAACGCTAACGCGAACGCCA", "ATCACGATCACCATCTAGAT"},
	{"P10", "CGCCAACCATAACCAGAACC", "ATCACGATCACCATCTAGAT"},
	{"B11", "GCGGAAGCGCAAGCCTAAGC", "TCTTGATCTTCATCTGTATC"},
	{"D11", "TAAGCCGAAGCCCAACAACA", "TCTTGATCTTCATCTGTATC"},
	{"F11", "ACATGAACATCAACAGTAAC", "TCTTGATCTTCATCTGTATC"},
	{"H11", "CGAGAACGACAACGTTAACG", "TCTTGATCTTCATCTGTATC"},
	{"J11", "TAACGTGAACGTCAACGGTA", "TCTTGATCTTCATCTGTATC"},
	{"L11", "CGGTAACGGGAACGGCAACG", "TCTTGATCTTCATCTGTATC"},
	{"N11", "CAACGCTAACGCGAACGCCA", "TCTTGATCTTCATCTGTATC"},
	{"P11", "CGCCAACCATAACCAGAACC", "TCTTGATCTTCATCTGTATC"},
	{"B12", "GCGGAAGCGCAAGCCTAAGC", "CCAGACACAGACTTAGACTG"},
	{"D12", "TAAGCCGAAGCCCAACAACA", "CCAGACACAGACTTAGACTG"},
	{"F12", "ACATGAACATCAACAGTAAC", "CCAGACACAGACTTAGACTG"},
	{"H12", "CGAGAACGACAACGTTAACG", "CCAGACACAGACTTAGACTG"},
	{"J12", "TAACGTGAACGTCAACGGTA", "CCAGACACAGACTTAGACTG"},
	{"L12", "CGGTAACGGGAACGGCAACG", "CCAGACACAGACTTAGACTG"},
	{"N12", "CAACGCTAACGCGAACGCCA", "CCAGACACAGACTTAGACTG"},
	{"P12", "CGCCAACCATAACCAGAACC", "CCAGACACAGACTTAGACTG"},
	{"B13", "GCGGAAGCGCAAGCCTAAGC", "GACTGAGACTCAGACGTAGA"},
	{"D13", "TAAGCCGAAGCCCAACAACA", "GACTGAGACTCAGACGTAGA"},
	{"F13", "ACATGAACATCAACAGTAAC", "GACTGAGACTCAGACGTAGA"},
	{"H13", "CGAGAACGACAACGTTAACG", "GACTGAGACTCAGACGTAGA"},
	{"J13", "TAACGTGAACGTCAACGGTA", "GACTGAGACTCAGACGTAGA"},
	{"L13", "CGGTAACGGGAACGGCAACG", "GACTGAGACTCAGACGTAGA"},
	{"N13", "CAACGCTAACGCGAACGCCA", "GACTGAGACTCAGACGTAGA"},
	{"P13", "CGCCAACCATAACCAGAACC", "GACTGAGACTCAGACGTAGA"},
	{"B14", "GCGGAAGCGCAAGCCTAAGC", "GTAGACGGAGACGCAGACCT"},
	{"D14", "TAAGCCGAAGCCCAACAACA", "GTAGACGGAGACGCAGACCT"},
	{"F14", "ACATGAACATCAACAGTAAC", "GTAGACGGAGACGCAGACCT"},
	{"H14", "CGAGAACGACAACGTTAACG", "GTAGACGGAGACGCAGACCT"},
	{"J14", "TAACGTGAACGTCAACGGTA", "GTAGACGGAGACGCAGACCT"},
	{"L14", "CGGTAACGGGAACGGCAACG", "GTAGACGGAGACGCAGACCT"},
	{"N14", "CAACGCTAACGCGAACGCCA", "GTAGACGGAGACGCAGACCT"},
	{"P14", "CGCCAACCATAACCAGAACC", "GTAGACGGAGACGCAGACCT"},
	{"B15", "GCGGAAGCGCAAGCCTAAGC", "GACCTAGACCGAGACCCAGT"},
	{"D15", "TAAGCCGAAGCCCAACAACA", "GACCTAGACCGAGACCCAGT"},
	{"F15", "ACATGAACATCAACAGTAAC", "GACCTAGACCGAGACCCAGT"},
	{"H15", "CGAGAACGACAACGTTAACG", "GACCTAGACCGAGACCCAGT"},
	{"J15", "TAACGTGAACGTCAACGGTA", "GACCTAGACCGAGACCCAGT"},
	{"L15", "CGGTAACGGGAACGGCAACG", "GACCTAGACCGAGACCCAGT"},
	{"N15", "CAACGCTAACGCGAACGCCA", "GACCTAGACCGAGACCCAGT"},
	{"P15", "CGCCAACCATAACCAGAACC", "GACCTAGACCGAGACCCAGT"},
	{"B16", "GCGGAAGCGCAAGCCTAAGC", "CCAGTAGTAGGAGTAGCAGT"},
	{"D16", "TAAGCCGAAGCCCAACAACA", "CCAGTAGTAGGAGTAGCAGT"},
	{"F16", "ACATGAACATCAACAGTAAC", "CCAGTAGTAGGAGTAGCAGT"},
	{"H16", "CGAGAACGACAACGTTAACG", "CCAGTAGTAGGAGTAGCAGT"},
	{"J16", "TAACGTGAACGTCAACGGTA", "CCAGTAGTAGGAGTAGCAGT"},
	{"L16", "CGGTAACGGGAACGGCAACG", "CCAGTAGTAGGAGTAGCAGT"},
	{"N16", "CAACGCTAACGCGAACGCCA", "CCAGTAGTAGGAGTAGCAGT"},
	{"P16", "CGCCAACCATAACCAGAACC", "CCAGTAGTAGGAGTAGCAGT"},
	{"B17", "GCGGAAGCGCAAGCCTAAGC", "GCAGTACTAGTACGAGTACC"},
	{"D17", "TAAGCCGAAGCCCAACAACA", "GCAGTACTAGTACGAGTACC"},
	{"F17", "ACATGAACATCAACAGTAAC", "GCAGTACTAGTACGAGTACC"},
	{"H17", "CGAGAACGACAACGTTAACG", "GCAGTACTAGTACGAGTACC"},
	{"J17", "TAACGTGAACGTCAACGGTA", "GCAGTACTAGTACGAGTACC"},
	{"L17", "CGGTAACGGGAACGGCAACG", "GCAGTACTAGTACGAGTACC"},
	{"N17", "CAACGCTAACGCGAACGCCA", "GCAGTACTAGTACGAGTACC"},
	{"P17", "CGCCAACCATAACCAGAACC", "GCAGTACTAGTACGAGTACC"},
	{"B18", "GCGGAAGCGCAAGCCTAAGC", "GTACCAGTTACAGTTTTAGT"},
	{"D18", "TAAGCCGAAGCCCAACAACA", "GTACCAGTTACAGTTTTAGT"},
	{"F18", "ACATGAACATCAACAGTAAC", "GTACCAGTTACAGTTTTAGT"},
	{"H18", "CGAGAACGACAACGTTAACG", "GTACCAGTTACAGTTTTAGT"},
	{"J18", "TAACGTGAACGTCAACGGTA", "GTACCAGTTACAGTTTTAGT"},
	{"L18", "CGGTAACGGGAACGGCAACG", "GTACCAGTTACAGTTTTAGT"},
	{"N18", "CAACGCTAACGCGAACGCCA", "GTACCAGTTACAGTTTTAGT"},
	{"P18", "CGCCAACCATAACCAGAACC", "GTACCAGTTACAGTTTTAGT"},
	{"B19", "GCGGAAGCGCAAGCCTAAGC", "AGTTTGAGTTTCAGTTGTAG"},
	{"D19", "TAAGCCGAAGCCCAACAACA", "AGTTTGAGTTTCAGTTGTAG"},
	{"F19", "ACATGAACATCAACAGTAAC", "AGTTTGAGTTTCAGTTGTAG"},
	{"H19", "CGAGAACGACAACGTTAACG", "AGTTTGAGTTTCAGTTGTAG"},
	{"J19", "TAACGTGAACGTCAACGGTA", "AGTTTGAGTTTCAGTTGTAG"},
	{"L19", "CGGTAACGGGAACGGCAACG", "AGTTTGAGTTTCAGTTGTAG"},
	{"N19", "CAACGCTAACGCGAACGCCA", "AGTTTGAGTTTCAGTTGTAG"},
	{"P19", "CGCCAACCATAACCAGAACC", "AGTTTGAGTTTCAGTTGTAG"},
	{"B20", "GCGGAAGCGCAAGCCTAAGC", "GTTCCAGTGACAGTGTTAGT"},
	{"D20", "TAAGCCGAAGCCCAACAACA", "GTTCCAGTGACAGTGTTAGT"},
	{"F20", "ACATGAACATCAACAGTAAC", "GTTCCAGTGACAGTGTTAGT"},
	{"H20", "CGAGAACGACAACGTTAACG", "GTTCCAGTGACAGTGTTAGT"},
	{"J20", "TAACGTGAACGTCAACGGTA", "GTTCCAGTGACAGTGTTAGT"},
	{"L20", "CGGTAACGGGAACGGCAACG", "GTTCCAGTGACAGTGTTAGT"},
	{"N20", "CAACGCTAACGCGAACGCCA", "GTTCCAGTGACAGTGTTAGT"},
	{"P20", "CGCCAACCATAACCAGAACC", "GTTCCAGTGACAGTGTTAGT"},
	{"B21", "GCGGAAGCGCAAGCCTAAGC", "TTAGTGTGAGTGTCAGTGGT"},
	{"D21", "TAAGCCGAAGCCCAACAACA", "TTAGTGTGAGTGTCAGTGGT"},
	{"F21", "ACATGAACATCAACAGTAAC", "TTAGTGTGAGTGTCAGTGGT"},
	{"H21", "CGAGAACGACAACGTTAACG", "TTAGTGTGAGTGTCAGTGGT"},
	{"J21", "TAACGTGAACGTCAACGGTA", "TTAGTGTGAGTGTCAGTGGT"},
	{"L21", "CGGTAACGGGAACGGCAACG", "TTAGTGTGAGTGTCAGTGGT"},
	{"N21", "CAACGCTAACGCGAACGCCA", "TTAGTGTGAGTGTCAGTGGT"},
	{"P21", "CGCCAACCATAACCAGAACC", "TTAGTGTGAGTGTCAGTGGT"},
	{"B22", "GCGGAAGCGCAAGCCTAAGC", "GTGGTAGTGGGAGTGGCAGT"},
	{"D22", "TAAGCCGAAGCCCAACAACA", "GTGGTAGTGGGAGTGGCAGT"},
	{"F22", "ACATGAACATCAACAGTAAC", "GTGGTAGTGGGAGTGGCAGT"},
	{"H22", "CGAGAACGACAACGTTAACG", "GTGGTAGTGGGAGTGGCAGT"},
	{"J22", "TAACGTGAACGTCAACGGTA", "GTGGTAGTGGGAGTGGCAGT"},
	{"L22", "CGGTAACGGGAACGGCAACG", "GTGGTAGTGGGAGTGGCAGT"},
	{"N22", "CAACGCTAACGCGAACGCCA", "GTGGTAGTGGGAGTGGCAGT"},
	{"P22", "CGCCAACCATAACCAGAACC", "GTGGTAGTGGGAGTGGCAGT"},
	{"B23", "GCGGAAGCGCAAGCCTAAGC", "GCAGTGCTAGTGCGAGTGCC"},
	{"D23", "TAAGCCGAAGCCCAACAACA", "GCAGTGCTAGTGCGAGTGCC"},
	{"F23", "ACATGAACATCAACAGTAAC", "GCAGTGCTAGTGCGAGTGCC"},
	{"H23", "CGAGAACGACAACGTTAACG", "GCAGTGCTAGTGCGAGTGCC"},
	{"J23", "TAACGTGAACGTCAACGGTA", "GCAGTGCTAGTGCGAGTGCC"},
	{"L23", "CGGTAACGGGAACGGCAACG", "GCAGTGCTAGTGCGAGTGCC"},
	{"N23", "CAACGCTAACGCGAACGCCA", "GCAGTGCTAGTGCGAGTGCC"},
	{"P23", "CGCCAACCATAACCAGAACC", "GCAGTGCTAGTGCGAGTGCC"},
	{"B24", "GCGGAAGCGCAAGCCTAAGC", "GTGCCAGTCACAGTCTTAGT"},
	{"D24", "TAAGCCGAAGCCCAACAACA", "GTGCCAGTCACAGTCTTAGT"},
	{"F24", "ACATGAACATCAACAGTAAC", "GTGCCAGTCACAGTCTTAGT"},
	{"H24", "CGAGAACGACAACGTTAACG", "GTGCCAGTCACAGTCTTAGT"},
	{"J24", "TAACGTGAACGTCAACGGTA", "GTGCCAGTCACAGTCTTAGT"},
	{"L24", "CGGTAACGGGAACGGCAACG", "GTGCCAGTCACAGTCTTAGT"},
	{"N24", "CAACGCTAACGCGAACGCCA", "GTGCCAGTCACAGTCTTAGT"},
	{"P24", "CGCCAACCATAACCAGAACC", "GTGCCAGTCACAGTCTTAGT"},
})
