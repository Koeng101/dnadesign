/*
Package cs contains functions for parsing minimap2 cs optional tags.

The following is taken directly from the minimap2 README documentation:

# The cs optional tag

The `cs` SAM/PAF tag encodes bases at mismatches and INDELs. It matches regular
expression `/(:[0-9]+|\*[a-z][a-z]|[=\+\-][A-Za-z]+)+/`. Like CIGAR, `cs`
consists of series of operations.  Each leading character specifies the
operation; the following sequence is the one involved in the operation.

The `cs` tag is enabled by command line option `--cs`. The following alignment,
for example:

	CGATCGATAAATAGAGTAG---GAATAGCA
	||||||   ||||||||||   |||| |||
	CGATCG---AATAGAGTAGGTCGAATtGCA

is represented as `:6-ata:10+gtc:4*at:3`, where `:[0-9]+` represents an
identical block, `-ata` represents a deletion, `+gtc` an insertion and `*at`
indicates reference base `a` is substituted with a query base `t`. It is
similar to the `MD` SAM tag but is standalone and easier to parse.

If `--cs=long` is used, the `cs` string also contains identical sequences in
the alignment. The above example will become
`=CGATCG-ata=AATAGAGTAG+gtc=GAAT*at=GCA`. The long form of `cs` encodes both
reference and query sequences in one string. The `cs` tag also encodes intron
positions and splicing signals (see the [minimap2 manpage][manpage-cs] for
details).
*/
package cs

import (
	"strconv"
	"unicode"

	"github.com/koeng101/dnadesign/lib/transform"
)

// CS is a struct format of each element of a cs string. For example, the cs
//
//	:6-ata:10+gtc:4*at:3
//
// would be divided into 7 CS structs. These can be directly processed, or
// further processed into DigestedCS and DigestedInsertions.
type CS struct {
	Type   rune   // Acceptable types: [* + :]
	Size   int    // Size of cs. For example, :6 would be 6 or -ata would be 3
	Change string // if insertion or deletion, write out the change. -ata would be ata
}

/*
Context for developers:

DigestedCS is my way of encoding more simply changes for the needs of DNA
synthesis validation. The core idea is that if we remove the need for
specifically encoding insertions, we can flatten the data structure, and make
it a lot easier to use. The full insertion information can be stored separately
in a DigestedInsertion struct.
*/

// DigestedCS is a simplified way of viewing CS, specifically for finding
// mutations in sequences. It only contains the position of a given change
// and the type of change. For point mutations and indels, this contains full
// information of the change - a longer deletion will simply be a few * in a
// row. However, it doesn't contain the full information for insertions. For
// DNA synthesis applications, this is mostly OK because insertions are very
// rare.
type DigestedCS struct {
	Position          uint64 // Position in the sequence with the change
	Type              uint8  // The change type. Can only be [. A T G C * +]
	Qual              byte   // The byte of the quality
	ReverseComplement bool
}

// DigestedInsertion separately stores inserts from the DigestedCS, in case
// this data is needed for analysis later.
type DigestedInsertion struct {
	Position          uint64 // Position in the sequence with an insertion
	Insertion         string // the insertion string
	Qual              string // The string of the quality
	ReverseComplement bool
}

// toCS is a function that properly adds size to CS. We use a function here
// instead copying this work twice to get 100% testing.
func toCS(current CS, numBuffer string) CS {
	if current.Type == ':' {
		current.Size, _ = strconv.Atoi(numBuffer)
	} else if current.Type == '*' {
		current.Size = 1 // point mutations are always 1 base pair
	} else {
		current.Size = len(current.Change)
	}
	return current
}

// ParseCS takes a short CS string and returns a slice of CS structs
func ParseCS(csString string) []CS {
	var result []CS
	var current CS
	var numBuffer string

	for _, char := range csString {
		switch {
		case char == ':' || char == '*' || char == '+' || char == '-':
			if current.Type != 0 {
				result = append(result, toCS(current, numBuffer))
			}
			current = CS{Type: char}
			numBuffer = ""
		case unicode.IsDigit(char):
			numBuffer += string(char)
		default:
			current.Change += string(char)
		}
	}

	// Add the last CS struct
	if current.Type != 0 {
		result = append(result, toCS(current, numBuffer))
	}

	return result
}

// DigestCS takes a slice of CS structs and returns slices of DigestedCS and DigestedInsertion
func DigestCS(csSlice []CS, quality string, reverseComplement bool) ([]DigestedCS, []DigestedInsertion) {
	if reverseComplement {
		quality = transform.ReverseString(quality)
	}
	var digestedCS []DigestedCS
	var digestedInsertions []DigestedInsertion
	position := uint64(0)
	// We want the quality of the fragment to follow the query, not the
	// reference sequence.
	var qualityPosition int

	for _, cs := range csSlice {
		switch cs.Type {
		case ':':
			for i := 0; i < cs.Size; i++ {
				digestedCS = append(digestedCS, DigestedCS{Position: position, Type: '.', ReverseComplement: reverseComplement, Qual: quality[qualityPosition]})
				position++
				qualityPosition++
			}
		case '*':
			digestedCS = append(digestedCS, DigestedCS{Position: position, Type: cs.Change[1], ReverseComplement: reverseComplement, Qual: quality[qualityPosition]}) // *at we need t
			position++
			qualityPosition++
		case '-':
			for i := 0; i < cs.Size; i++ {
				digestedCS = append(digestedCS, DigestedCS{
					Position:          position,
					Type:              '*',
					ReverseComplement: reverseComplement,
					// No quality for deletions mutations.
				})
				position++
			}
		case '+':
			// Insertions are positioned at where they are inserted. For example,
			// if we have it between 18 and 19, the insertion "Position" would be 19
			digestedInsertions = append(digestedInsertions, DigestedInsertion{
				Position:          position,
				Insertion:         cs.Change,
				ReverseComplement: reverseComplement,
				Qual:              quality[qualityPosition : qualityPosition+len(cs.Change)],
			})
			// Don't increment position for insertions, but increment qualityPosition
			qualityPosition += len(cs.Change)
		}
	}

	return digestedCS, digestedInsertions
}
