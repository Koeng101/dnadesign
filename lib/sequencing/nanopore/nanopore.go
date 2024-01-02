/*
Package nanopore contains data associated with nanopore sequencing.
*/
package nanopore

import (
	"fmt"
	"strings"

	"github.com/koeng101/dnadesign/lib/align"
	"github.com/koeng101/dnadesign/lib/align/matrix"
	"github.com/koeng101/dnadesign/lib/alphabet"
	"github.com/koeng101/dnadesign/lib/bio/fastq"
)

// ScoreMagicNumber is the score of a Smith Waterman alignment between a
// barcode and sequence used a threshold for whether the barcode exists.
// It was found from me playing around with sequences.
var ScoreMagicNumber = 12

// NativeBarcode contains the data structure defining a nanopore barcode.
// In between Forward and a target DNA sequence is 8bp: CAGCACC followed by a
// T, which is used for the TA ligation to the target end-prepped DNA.
type NativeBarcode struct {
	Forward string `json:"forward"`
	Reverse string `json:"reverse"`
}

// NativeBarcodeMap contains a map of native barcodes to their respective
// forward and reverse sequences.
var NativeBarcodeMap = map[string]NativeBarcode{
	"barcode01": {"CACAAAGACACCGACAACTTTCTT", "AAGAAAGTTGTCGGTGTCTTTGTG"},
	"barcode02": {"ACAGACGACTACAAACGGAATCGA", "TCGATTCCGTTTGTAGTCGTCTGT"},
	"barcode03": {"CCTGGTAACTGGGACACAAGACTC", "GAGTCTTGTGTCCCAGTTACCAGG"},
	"barcode04": {"TAGGGAAACACGATAGAATCCGAA", "TTCGGATTCTATCGTGTTTCCCTA"},
	"barcode05": {"AAGGTTACACAAACCCTGGACAAG", "CTTGTCCAGGGTTTGTGTAACCTT"},
	"barcode06": {"GACTACTTTCTGCCTTTGCGAGAA", "TTCTCGCAAAGGCAGAAAGTAGTC"},
	"barcode07": {"AAGGATTCATTCCCACGGTAACAC", "GTGTTACCGTGGGAATGAATCCTT"},
	"barcode08": {"ACGTAACTTGGTTTGTTCCCTGAA", "TTCAGGGAACAAACCAAGTTACGT"},
	"barcode09": {"AACCAAGACTCGCTGTGCCTAGTT", "AACTAGGCACAGCGAGTCTTGGTT"},
	"barcode10": {"GAGAGGACAAAGGTTTCAACGCTT", "AAGCGTTGAAACCTTTGTCCTCTC"},
	"barcode11": {"TCCATTCCCTCCGATAGATGAAAC", "GTTTCATCTATCGGAGGGAATGGA"},
	"barcode12": {"TCCGATTCTGCTTCTTTCTACCTG", "CAGGTAGAAAGAAGCAGAATCGGA"},
	"barcode13": {"AGAACGACTTCCATACTCGTGTGA", "TCACACGAGTATGGAAGTCGTTCT"},
	"barcode14": {"AACGAGTCTCTTGGGACCCATAGA", "TCTATGGGTCCCAAGAGACTCGTT"},
	"barcode15": {"AGGTCTACCTCGCTAACACCACTG", "CAGTGGTGTTAGCGAGGTAGACCT"},
	"barcode16": {"CGTCAACTGACAGTGGTTCGTACT", "AGTACGAACCACTGTCAGTTGACG"},
	"barcode17": {"ACCCTCCAGGAAAGTACCTCTGAT", "ATCAGAGGTACTTTCCTGGAGGGT"},
	"barcode18": {"CCAAACCCAACAACCTAGATAGGC", "GCCTATCTAGGTTGTTGGGTTTGG"},
	"barcode19": {"GTTCCTCGTGCAGTGTCAAGAGAT", "ATCTCTTGACACTGCACGAGGAAC"},
	"barcode20": {"TTGCGTCCTGTTACGAGAACTCAT", "ATGAGTTCTCGTAACAGGACGCAA"},
	"barcode21": {"GAGCCTCTCATTGTCCGTTCTCTA", "TAGAGAACGGACAATGAGAGGCTC"},
	"barcode22": {"ACCACTGCCATGTATCAAAGTACG", "CGTACTTTGATACATGGCAGTGGT"},
	"barcode23": {"CTTACTACCCAGTGAACCTCCTCG", "CGAGGAGGTTCACTGGGTAGTAAG"},
	"barcode24": {"GCATAGTTCTGCATGATGGGTTAG", "CTAACCCATCATGCAGAACTATGC"},
	"barcode25": {"GTAAGTTGGGTATGCAACGCAATG", "CATTGCGTTGCATACCCAACTTAC"},
	"barcode26": {"CATACAGCGACTACGCATTCTCAT", "ATGAGAATGCGTAGTCGCTGTATG"},
	"barcode27": {"CGACGGTTAGATTCACCTCTTACA", "TGTAAGAGGTGAATCTAACCGTCG"},
	"barcode28": {"TGAAACCTAAGAAGGCACCGTATC", "GATACGGTGCCTTCTTAGGTTTCA"},
	"barcode29": {"CTAGACACCTTGGGTTGACAGACC", "GGTCTGTCAACCCAAGGTGTCTAG"},
	"barcode30": {"TCAGTGAGGATCTACTTCGACCCA", "TGGGTCGAAGTAGATCCTCACTGA"},
	"barcode31": {"TGCGTACAGCAATCAGTTACATTG", "CAATGTAACTGATTGCTGTACGCA"},
	"barcode32": {"CCAGTAGAAGTCCGACAACGTCAT", "ATGACGTTGTCGGACTTCTACTGG"},
	"barcode33": {"CAGACTTGGTACGGTTGGGTAACT", "AGTTACCCAACCGTACCAAGTCTG"},
	"barcode34": {"GGACGAAGAACTCAAGTCAAAGGC", "GCCTTTGACTTGAGTTCTTCGTCC"},
	"barcode35": {"CTACTTACGAAGCTGAGGGACTGC", "GCAGTCCCTCAGCTTCGTAAGTAG"},
	"barcode36": {"ATGTCCCAGTTAGAGGAGGAAACA", "TGTTTCCTCCTCTAACTGGGACAT"},
	"barcode37": {"GCTTGCGATTGATGCTTAGTATCA", "TGATACTAAGCATCAATCGCAAGC"},
	"barcode38": {"ACCACAGGAGGACGATACAGAGAA", "TTCTCTGTATCGTCCTCCTGTGGT"},
	"barcode39": {"CCACAGTGTCAACTAGAGCCTCTC", "GAGAGGCTCTAGTTGACACTGTGG"},
	"barcode40": {"TAGTTTGGATGACCAAGGATAGCC", "GGCTATCCTTGGTCATCCAAACTA"},
	"barcode41": {"GGAGTTCGTCCAGAGAAGTACACG", "CGTGTACTTCTCTGGACGAACTCC"},
	"barcode42": {"CTACGTGTAAGGCATACCTGCCAG", "CTGGCAGGTATGCCTTACACGTAG"},
	"barcode43": {"CTTTCGTTGTTGACTCGACGGTAG", "CTACCGTCGAGTCAACAACGAAAG"},
	"barcode44": {"AGTAGAAAGGGTTCCTTCCCACTC", "GAGTGGGAAGGAACCCTTTCTACT"},
	"barcode45": {"GATCCAACAGAGATGCCTTCAGTG", "CACTGAAGGCATCTCTGTTGGATC"},
	"barcode46": {"GCTGTGTTCCACTTCATTCTCCTG", "CAGGAGAATGAAGTGGAACACAGC"},
	"barcode47": {"GTGCAACTTTCCCACAGGTAGTTC", "GAACTACCTGTGGGAAAGTTGCAC"},
	"barcode48": {"CATCTGGAACGTGGTACACCTGTA", "TACAGGTGTACCACGTTCCAGATG"},
	"barcode49": {"ACTGGTGCAGCTTTGAACATCTAG", "CTAGATGTTCAAAGCTGCACCAGT"},
	"barcode50": {"ATGGACTTTGGTAACTTCCTGCGT", "ACGCAGGAAGTTACCAAAGTCCAT"},
	"barcode51": {"GTTGAATGAGCCTACTGGGTCCTC", "GAGGACCCAGTAGGCTCATTCAAC"},
	"barcode52": {"TGAGAGACAAGATTGTTCGTGGAC", "GTCCACGAACAATCTTGTCTCTCA"},
	"barcode53": {"AGATTCAGACCGTCTCATGCAAAG", "CTTTGCATGAGACGGTCTGAATCT"},
	"barcode54": {"CAAGAGCTTTGACTAAGGAGCATG", "CATGCTCCTTAGTCAAAGCTCTTG"},
	"barcode55": {"TGGAAGATGAGACCCTGATCTACG", "CGTAGATCAGGGTCTCATCTTCCA"},
	"barcode56": {"TCACTACTCAACAGGTGGCATGAA", "TTCATGCCACCTGTTGAGTAGTGA"},
	"barcode57": {"GCTAGGTCAATCTCCTTCGGAAGT", "ACTTCCGAAGGAGATTGACCTAGC"},
	"barcode58": {"CAGGTTACTCCTCCGTGAGTCTGA", "TCAGACTCACGGAGGAGTAACCTG"},
	"barcode59": {"TCAATCAAGAAGGGAAAGCAAGGT", "ACCTTGCTTTCCCTTCTTGATTGA"},
	"barcode60": {"CATGTTCAACCAAGGCTTCTATGG", "CCATAGAAGCCTTGGTTGAACATG"},
	"barcode61": {"AGAGGGTACTATGTGCCTCAGCAC", "GTGCTGAGGCACATAGTACCCTCT"},
	"barcode62": {"CACCCACACTTACTTCAGGACGTA", "TACGTCCTGAAGTAAGTGTGGGTG"},
	"barcode63": {"TTCTGAAGTTCCTGGGTCTTGAAC", "GTTCAAGACCCAGGAACTTCAGAA"},
	"barcode64": {"GACAGACACCGTTCATCGACTTTC", "GAAAGTCGATGAACGGTGTCTGTC"},
	"barcode65": {"TTCTCAGTCTTCCTCCAGACAAGG", "CCTTGTCTGGAGGAAGACTGAGAA"},
	"barcode66": {"CCGATCCTTGTGGCTTCTAACTTC", "GAAGTTAGAAGCCACAAGGATCGG"},
	"barcode67": {"GTTTGTCATACTCGTGTGCTCACC", "GGTGAGCACACGAGTATGACAAAC"},
	"barcode68": {"GAATCTAAGCAAACACGAAGGTGG", "CCACCTTCGTGTTTGCTTAGATTC"},
	"barcode69": {"TACAGTCCGAGCCTCATGTGATCT", "AGATCACATGAGGCTCGGACTGTA"},
	"barcode70": {"ACCGAGATCCTACGAATGGAGTGT", "ACACTCCATTCGTAGGATCTCGGT"},
	"barcode71": {"CCTGGGAGCATCAGGTAGTAACAG", "CTGTTACTACCTGATGCTCCCAGG"},
	"barcode72": {"TAGCTGACTGTCTTCCATACCGAC", "GTCGGTATGGAAGACAGTCAGCTA"},
	"barcode73": {"AAGAAACAGGATGACAGAACCCTC", "GAGGGTTCTGTCATCCTGTTTCTT"},
	"barcode74": {"TACAAGCATCCCAACACTTCCACT", "AGTGGAAGTGTTGGGATGCTTGTA"},
	"barcode75": {"GACCATTGTGATGAACCCTGTTGT", "ACAACAGGGTTCATCACAATGGTC"},
	"barcode76": {"ATGCTTGTTACATCAACCCTGGAC", "GTCCAGGGTTGATGTAACAAGCAT"},
	"barcode77": {"CGACCTGTTTCTCAGGGATACAAC", "GTTGTATCCCTGAGAAACAGGTCG"},
	"barcode78": {"AACAACCGAACCTTTGAATCAGAA", "TTCTGATTCAAAGGTTCGGTTGTT"},
	"barcode79": {"TCTCGGAGATAGTTCTCACTGCTG", "CAGCAGTGAGAACTATCTCCGAGA"},
	"barcode80": {"CGGATGAACATAGGATAGCGATTC", "GAATCGCTATCCTATGTTCATCCG"},
	"barcode81": {"CCTCATCTTGTGAAGTTGTTTCGG", "CCGAAACAACTTCACAAGATGAGG"},
	"barcode82": {"ACGGTATGTCGAGTTCCAGGACTA", "TAGTCCTGGAACTCGACATACCGT"},
	"barcode83": {"TGGCTTGATCTAGGTAAGGTCGAA", "TTCGACCTTACCTAGATCAAGCCA"},
	"barcode84": {"GTAGTGGACCTAGAACCTGTGCCA", "TGGCACAGGTTCTAGGTCCACTAC"},
	"barcode85": {"AACGGAGGAGTTAGTTGGATGATC", "GATCATCCAACTAACTCCTCCGTT"},
	"barcode86": {"AGGTGATCCCAACAAGCGTAAGTA", "TACTTACGCTTGTTGGGATCACCT"},
	"barcode87": {"TACATGCTCCTGTTGTTAGGGAGG", "CCTCCCTAACAACAGGAGCATGTA"},
	"barcode88": {"TCTTCTACTACCGATCCGAAGCAG", "CTGCTTCGGATCGGTAGTAGAAGA"},
	"barcode89": {"ACAGCATCAATGTTTGGCTAGTTG", "CAACTAGCCAAACATTGATGCTGT"},
	"barcode90": {"GATGTAGAGGGTACGGTTTGAGGC", "GCCTCAAACCGTACCCTCTACATC"},
	"barcode91": {"GGCTCCATAGGAACTCACGCTACT", "AGTAGCGTGAGTTCCTATGGAGCC"},
	"barcode92": {"TTGTGAGTGGAAAGATACAGGACC", "GGTCCTGTATCTTTCCACTCACAA"},
	"barcode93": {"AGTTTCCATCACTTCAGACTTGGG", "CCCAAGTCTGAAGTGATGGAAACT"},
	"barcode94": {"GATTGTCCTCAAACTGCCACCTAC", "GTAGGTGGCAGTTTGAGGACAATC"},
	"barcode95": {"CCTGTCTGGAAGAAGAATGGACTT", "AAGTCCATTCTTCTTCCAGACAGG"},
	"barcode96": {"CTGAACGGTCATAGAGTCCACCAT", "ATGGTGGACTCTATGACCGTTCAG"},
}

func TrimBarcodeWithChannels(barcodeName string, fastqInput <-chan *fastq.Read, fastqOutput chan<- *fastq.Read) error {
	for {
		select {
		case data, ok := <-fastqInput:
			if !ok {
				// If the input channel is closed, we close the output channel and return
				close(fastqOutput)
				return nil
			}
			trimmedRead, err := TrimBarcode(barcodeName, *data)
			if err != nil {
				close(fastqOutput)
				return err
			}
			fastqOutput <- &trimmedRead
		}
	}

}

// TrimBarcode takes a barcodeName and a fastqSequence and returns a trimmed
// barcode.
func TrimBarcode(barcodeName string, fastqRead fastq.Read) (fastq.Read, error) {
	// Copy into new fastq read
	var newFastqRead fastq.Read
	newFastqRead.Identifier = fastqRead.Identifier
	newFastqRead.Optionals = make(map[string]string)
	for key, value := range fastqRead.Optionals {
		newFastqRead.Optionals[key] = value
	}

	// Get the barcode
	barcode, ok := NativeBarcodeMap[barcodeName]
	if !ok {
		return newFastqRead, fmt.Errorf("barcode %s not found in NativeBarcodeMap.", barcodeName)
	}

	// Trim in the forward direction
	var sequence string
	var quality string
	sequence = fastqRead.Sequence
	quality = fastqRead.Quality
	score, alignment, err := Align(sequence[:80], barcode.Forward)
	if err != nil {
		return newFastqRead, err
	}
	// Empirically from looking around, it seems to me that approximately a
	// score of 21 looks like a barcode match. This is almost by definition a
	// magic number, so it is defined above as so.
	if score >= ScoreMagicNumber {
		modifiedAlignment := strings.ReplaceAll(alignment, "-", "")
		index := strings.Index(sequence, modifiedAlignment)
		// The index needs to be within the first 80 base pairs. Usually the
		// native adapter + native barcode is within the first ~70 bp, but we
		// put a small buffer.
		if index != -1 {
			// 7 here is a magic number between the native barcode and your
			// target sequence. It's just a number that exists irl, so it is
			// not defined above.
			newStart := index + 7
			if newStart < len(sequence) {
				sequence = sequence[newStart:]
				quality = quality[newStart:]
			}
		}
	}
	// Now do the same thing with the reverse strand.
	score, alignment, err = Align(sequence[len(sequence)-80:], barcode.Reverse)
	if err != nil {
		return newFastqRead, err
	}
	if score >= ScoreMagicNumber {
		modifiedAlignment := strings.ReplaceAll(alignment, "-", "")
		index := strings.Index(sequence, modifiedAlignment)
		// This time we need to check within the last 80 base pairs.
		if index != -1 {
			newEnd := index - 7
			if newEnd < len(sequence) {
				sequence = sequence[:newEnd]
				quality = quality[:newEnd]
			}
		}
	}
	newFastqRead.Sequence = sequence
	newFastqRead.Quality = quality
	return newFastqRead, err
}

// Align uses the SmithWaterman algorithm to align a barcode to a sequence.
// It is used because it is a rather simple algorithm, and since barcodes are
// sufficiently small, fast enough.
func Align(sequence string, barcodeSequence string) (int, string, error) {
	m := [][]int{
		/*       A C G T U */
		/* A */ {1, -1, -1, -1, -1},
		/* C */ {-1, 1, -1, -1, -1},
		/* G */ {-1, -1, 1, -1, -1},
		/* T */ {-1, -1, -1, 1, -1},
		/* U */ {-1, -1, -1, -1, 1},
	}

	alphabet := alphabet.NewAlphabet([]string{"A", "C", "G", "T", "U"})
	subMatrix, err := matrix.NewSubstitutionMatrix(alphabet, alphabet, m)
	if err != nil {
		return 0, "", err
	}
	scoring, err := align.NewScoring(subMatrix, -1)
	if err != nil {
		return 0, "", err
	}
	score, alignSequence, _, err := align.SmithWaterman(sequence, barcodeSequence, scoring)
	if err != nil {
		return 0, "", err
	}
	return score, alignSequence, nil
}
