/*
Package seqhash contains the seqhash algorithm.

This package contains the reference seqhash algorithm.

There is a big problem with current sequence databases - they all use different
identifiers and accession numbers. This means cross-referencing databases is
a complicated exercise, especially as the quantity of databases increases, or if
you need to compare "wild" DNA sequences.

Seqhash is a simple algorithm to produce consistent identifiers for any genetic sequence. The
basic premise of the Seqhash algorithm is to hash sequences with the hash being a robust
cross-database identifier. Sequences themselves shouldn't be used as a database index
(often, they're too big), so a hash based off of a sequence is the next best thing.

Usability wise, you should be able to Seqhash any rotation of a sequence in any direction and
get a consistent hash.

The Seqhash algorithm makes several opinionated design choices, primarily to make working
with Seqhashes more consistent and nice. The Seqhash algorithm only uses a single hash function,
Blake3, and only operates on DNA, RNA, and Protein sequences. These identifiers will be seen
by human beings, so versioning and metadata is attached to the front of the hashes so that
a human operator can quickly identify problems with hashing.

If the sequence is DNA or RNA, the Seqhash algorithm needs to know whether or not the nucleic
acid is circular and/or double stranded. If circular, the sequence is rotated to a deterministic
point. If double stranded, the sequence is compared to its reverse complement, and the lexicographically
minimal sequence is taken (whether or not the min or max is used doesn't matter, just needs to
be consistent).

If the sequence is RNA, the sequence will be converted to DNA before hashing. While the full Seqhash
will still be different between RNA and DNA (due to the metadata string), the hash afterwards will be the same.
This makes it easy to cross reference DNA and RNA sequences. This fact is important for parts of DnaDesign
store that relate to storing and searching large quantities of sequences - deduplication can easily
be used on those Seqhashes to save a lot of space.

For DNA or RNA sequences, only ATUGCYRSWKMBDHVNZ characters are allowed. For Proteins,
only ACDEFGHIKLMNPQRSTVWYUO*BXZ characters are allowed in sequences. Selenocysteine (Sec; U) and pyrrolysine
(Pyl; O) are included in the protein character set - usually U and O don't occur within protein sequences,
but for certain organisms they do, and it is certainly a relevant amino acid for those particular proteins.

# Seqhash version 2

Version 1 seqhashes are depreciated.

Version 1 seqhashes were rather long, and version 2 seqhashes are built to be
much shorter. The intended use case are for handling sequences with LLM systems
since these system's context window is a value resource, and smaller references
allows the system to be more focused. Seqhash version 2 are approximately 3x
smaller than version 1 seqhashes. Officially, they are [16]byte arrays, but can
be also encoded with base58 to get a hash that can be used as a string across
different systems. Here is a length comparison:

	version 1: v1_DLD_f4028f93e08c5c23cbb8daa189b0a9802b378f1a1c919dcbcf1508a615f46350
	version 2: C_5X6Hudy3K8ht7r4mvu9Gco

The metadata is now encoded in a 1 byte flag rather than a metadata string,
instead of 7 rune like in version 1. Rather than use 256 bits for encoding
the hash, we use 120 bits. Since seqhashes are not meant for security, this
is good enough (50% collision with 1.3x10^18 hashes), while making them
conveniently only 16 btyes long. Additionally, encoded prefixes are added
to the front of the base58 encoded hash as a heuristic device for LLMs while
processing batches of seqhashes.

In addition, seqhashes can now encode fragments. Fragments are double stranded
DNA that are the result of restriction digestion, with single stranded
overhangs flanking both sides. These fragments can encode genetic parts - and
an important part of any vector containing these parts would be the part
seqhash, rather than the vector seqhash. This enhancement allows you to
identify genetic parts irregardless of their context.

Base58 is used rather than base64 so that seqhashes can easily be added into
urls without a "/" in the identifier. Ironically, it also makes smaller hashes
than base64 due to base64 chunking 3 bytes at a time - at 16 bytes, 2 blank
bytes are added to make the seqhash divisible by 3. Base58 chunks differently,
and so doesn't encounter this problem.
*/
package seqhash

import (
	"crypto/sha256"
	"errors"
	"sort"
	"strings"

	"github.com/koeng101/dnadesign/lib/transform"
)

// Seqhash is a struct that contains the Seqhash algorithm sequence types.
type SequenceType string

const (
	DNA      SequenceType = "DNA"
	RNA      SequenceType = "RNA"
	PROTEIN  SequenceType = "PROTEIN"
	FRAGMENT SequenceType = "FRAGMENT"
)

// boothLeastRotation gets the least rotation of a circular string.
func boothLeastRotation(sequence string) int {
	// https://en.wikipedia.org/wiki/Lexicographically_minimal_string_rotation
	// this is generally over commented but I'm keeping it this way for now. - Tim

	// first concatenate the sequence to itself to avoid modular arithmetic
	sequence += sequence // maybe do this as a buffer just for speed? May get annoying with larger sequences.
	leastRotationIndex := 0

	//initializing failure slice.
	failureSlice := make([]int, len(sequence))
	for i := range failureSlice {
		failureSlice[i] = -1
	}
	// iterate through each character in the doubled over sequence
	for characterIndex := 1; characterIndex < len(sequence); characterIndex++ {
		// get character
		character := sequence[characterIndex]
		// get failure
		failure := failureSlice[characterIndex-leastRotationIndex-1]
		// while failure does not equal -1 and character does not equal the character found at the least rotation + failure + 1 <- why this?
		for failure != -1 && character != sequence[leastRotationIndex+failure+1] {
			// if character is lexically less than whatever is at the leastRotationIndex index update leastRotation index
			if character < sequence[leastRotationIndex+failure+1] {
				leastRotationIndex = characterIndex - failure - 1
			}
			// update failure using previous failure as index?
			failure = failureSlice[failure]
		}

		// if character does not equal whatever character is at leastRotationIndex plus failure.
		if character != sequence[leastRotationIndex+failure+1] {
			// if character is lexically less then what is rotated least leastRotationIndex gets value of character index.
			if character < sequence[leastRotationIndex] {
				leastRotationIndex = characterIndex
			}
			// assign -1 to whatever is at the index of difference between character and rotation indices.
			failureSlice[characterIndex-leastRotationIndex] = -1

			// if character does equal whatever character is at leastRotationIndex plus failure.
		} else {
			// assign failure + 1 at the index of difference between character and rotation indices.
			failureSlice[characterIndex-leastRotationIndex] = failure + 1
		}
	} // end loop

	return leastRotationIndex
}

// RotateSequence rotates circular sequences to deterministic point.
func RotateSequence(sequence string) string {
	rotationIndex := boothLeastRotation(sequence)
	var sequenceBuilder strings.Builder

	// writing the same sequence twice. using build incase of very long circular genome.
	sequenceBuilder.WriteString(sequence)
	sequenceBuilder.WriteString(sequence)

	concatenatedSequence := sequenceBuilder.String()
	sequence = concatenatedSequence[rotationIndex : rotationIndex+len(sequence)]
	return sequence
}

// The following consts are for seqhash version 2
const (
	// Define bit masks for each part of the flag
	hash2versionMask        byte = 0b11110000 // Version occupies the first 4 bits
	hash2circularityMask    byte = 0b00001000 // Circularity occupies the 5th bit
	hash2doubleStrandedMask byte = 0b00000100 // Double-strandedness occupies the 6th bit
	hash2typeMask           byte = 0b00000011 // DNA/RNA/PROTEIN occupies the last 2 bits

	// Define shift counts for each part
	hash2versionShift        = 4
	hash2circularityShift    = 3
	hash2doubleStrandedShift = 2
)

var (
	// sequenceTypeStringToByteFlagMap converts a sequenceType to a byte
	sequenceTypeStringToByteFlagMap = map[SequenceType]byte{
		DNA:      0b00,
		RNA:      0b01,
		PROTEIN:  0b10,
		FRAGMENT: 0b11,
	}
	// sequenceTypeByteToStringFlagMap converts a byte to a sequenceType
	sequenceTypeByteToStringFlagMap = map[byte]SequenceType{
		0b00: DNA,
		0b01: RNA,
		0b10: PROTEIN,
		0b11: FRAGMENT,
	}
)

// EncodeFlag encodes the version, circularity, double-strandedness, and type into a single byte flag.
// Used for seqhash v2
func EncodeFlag(version int, sequenceType SequenceType, circularity bool, doubleStranded bool) byte {
	var flag byte

	// Encode the version (assuming version is in the range 0-16)
	flag |= (byte(version) << hash2versionShift)

	// Encode the circularity
	if circularity {
		flag |= (1 << hash2circularityShift)
	}

	// Encode the double-strandedness
	if doubleStranded {
		flag |= (1 << hash2doubleStrandedShift)
	}

	// Encode the DNA/RNA/PROTEIN
	dnaRnaProtein := sequenceTypeStringToByteFlagMap[sequenceType]
	flag |= (dnaRnaProtein & hash2typeMask)

	return flag
}

// DecodeFlag decodes the single byte flag into its constituent parts.
// Outputs: version, circularity, doubleStranded, dnaRnaProtein.
// Used for seqhash v2
func DecodeFlag(flag byte) (int, SequenceType, bool, bool) {
	version := int((flag & hash2versionMask) >> hash2versionShift)
	circularity := (flag & hash2circularityMask) != 0
	doubleStranded := (flag & hash2doubleStrandedMask) != 0
	dnaRnaProtein := flag & hash2typeMask
	sequenceType := sequenceTypeByteToStringFlagMap[dnaRnaProtein]

	return version, sequenceType, circularity, doubleStranded
}

// Hash2 creates a version 2 seqhash.
func Hash2(sequence string, sequenceType SequenceType, circular bool, doubleStranded bool) ([16]byte, error) {
	var result [16]byte

	// First, get the determistic sequence of the hash

	// By definition, Seqhashes are of uppercase sequences
	sequence = strings.ToUpper(sequence)
	// If RNA, convert to a DNA sequence. The hash itself between a DNA and RNA sequence will not
	// be different, but their Seqhash will have a different metadata string (R vs D)
	if sequenceType == SequenceType("RNA") {
		sequence = strings.ReplaceAll(sequence, "U", "T")
	}

	// Run checks on the input
	if sequenceType != DNA && sequenceType != RNA && sequenceType != PROTEIN {
		return result, errors.New("Only sequenceTypes of DNA, RNA, or PROTEIN allowed. Got sequenceType: " + string(sequenceType))
	}
	if sequenceType == DNA || sequenceType == RNA {
		for _, char := range sequence {
			if !strings.Contains("ATUGCYRSWKMBDHVNZ", string(char)) {
				return result, errors.New("Only letters ATUGCYRSWKMBDHVNZ are allowed for DNA/RNA. Got letter: " + string(char))
			}
		}
	}
	if sequenceType == PROTEIN {
		for _, char := range sequence {
			// Selenocysteine (Sec; U) and pyrrolysine (Pyl; O) are added
			// in accordance with https://www.uniprot.org/help/sequences
			// The release notes https://web.expasy.org/docs/relnotes/relstat.html
			// also state there are Asx (B), Glx (Z), and Xaa (X) amino acids, so
			// these are added in as well.
			if !strings.Contains("ACDEFGHIKLMNPQRSTVWYUO*BXZ", string(char)) {
				return result, errors.New("Only letters ACDEFGHIKLMNPQRSTVWYUO*BXZ are allowed for Proteins. Got letter: " + string(char))
			}
		}
	}
	// There is no check for circular proteins since proteins can be circular
	if sequenceType == PROTEIN && doubleStranded {
		return result, errors.New("Proteins cannot be double stranded")
	}
	// Gets Deterministic sequence based off of metadata + sequence
	var deterministicSequence string
	switch {
	case circular && doubleStranded:
		potentialSequences := []string{RotateSequence(sequence), RotateSequence(transform.ReverseComplement(sequence))}
		sort.Strings(potentialSequences)
		deterministicSequence = potentialSequences[0]
	case circular && !doubleStranded:
		deterministicSequence = RotateSequence(sequence)
	case !circular && doubleStranded:
		potentialSequences := []string{sequence, transform.ReverseComplement(sequence)}
		sort.Strings(potentialSequences)
		deterministicSequence = potentialSequences[0]
	case !circular && !doubleStranded:
		deterministicSequence = sequence
	}

	// Build our byte flag
	flag := EncodeFlag(2, sequenceType, circular, doubleStranded)
	result[0] = flag

	// Compute BLAKE3, then copy those to the remaining 16 bytes
	newhash := sha256.Sum256([]byte(deterministicSequence))
	copy(result[1:], newhash[:16])

	return result, nil
}

// Hash2Fragment creates a version 2 fragment seqhash. Fragment seqhashes are
// a special kind of seqhash that are used to identify fragments, usually
// released by restriction enzyme digestion, rather than complete DNA
// sequences. This is very useful for tracking genetic parts in a database: as
// abstractions away from their container vectors, so that many fragments in
// different vectors can be identified consistently.
//
// fwdOverhangLength and revOverhangLength are the lengths of both overhangs.
// Hashed sequences are hashed with their overhangs attached. Most of the time,
// both of these will equal 4, as they are released by TypeIIS restriction
// enzymes.
//
// In order to make sure fwdOverhangLength and revOverhangLength fit in the
// hash, the hash is truncated at 13 bytes rather than 16, and both int8 are
// inserted. So the bytes would be:
//
//	flag + fwdOverhangLength + revOverhangLength + [13]byte(hash)
//
// fwdOverhangLength and revOverhangLength are both int8, and their negatives
// are considered if the the overhang is on the 3prime strand, rather than the
// 5prime strand.
//
// 13 bytes is considered enough, because the number of fragments is limited
// by our ability to physically produce them, while other other sequence types
// can be found in nature.
//
// The fwdOverhang and revOverhang are the lengths of the overhangs of the
// input sequence. The hash, however, contains the forward and reverse overhang
// lengths of the deterministic sequence - ie, the alphabetically less-than
// strand, when comparing the uppercase forward and reverse complement strand.
// This means if the input sequence is not less than its reverse complement (for
// example, GTT is greater than AAC), then the output hash will have the forward
// and reverse overhang lengths of the reverse complement, not the input strand.
func Hash2Fragment(sequence string, fwdOverhangLength int8, revOverhangLength int8) ([16]byte, error) {
	var result [16]byte

	// First, run checks and get the determistic sequence of the hash
	for _, char := range sequence {
		if !strings.Contains("ATUGCYRSWKMBDHVNZ", string(char)) {
			return result, errors.New("Only letters ATUGCYRSWKMBDHVNZ are allowed for DNA/RNA. Got letter: " + string(char))
		}
	}
	sequence = strings.ToUpper(sequence)
	var forward, reverse int8
	var deterministicSequence string
	reverseComplement := transform.ReverseComplement(sequence)
	if sequence > reverseComplement {
		// If the reverse complement is smaller, reverse the overhangs forward and reverse
		forward = revOverhangLength
		reverse = fwdOverhangLength
		deterministicSequence = reverseComplement
	} else {
		forward = fwdOverhangLength
		reverse = revOverhangLength
		deterministicSequence = sequence
	}

	// Build our byte flag and copy length flags
	flag := EncodeFlag(2, FRAGMENT, false, false)
	result[0] = flag
	result[1] = byte(forward)
	result[2] = byte(reverse)

	// Compute BLAKE3, then copy those to the remaining 13 bytes
	newhash := sha256.Sum256([]byte(deterministicSequence))
	copy(result[3:], newhash[:13])

	return result, nil
}

// Hash2MetadataKey is a key for a seqhash v2 single letter metadata tag.
type Hash2MetadataKey struct {
	SequenceType   SequenceType
	Circular       bool
	DoubleStranded bool
}

// Hash2Metadata contains the seqhash v2 single letter metadata tags.
var Hash2Metadata = map[Hash2MetadataKey]rune{
	{DNA, true, true}:        'A',
	{DNA, true, false}:       'B',
	{DNA, false, true}:       'C',
	{DNA, false, false}:      'D',
	{RNA, true, true}:        'E',
	{RNA, true, false}:       'F',
	{RNA, false, true}:       'G',
	{RNA, false, false}:      'H',
	{PROTEIN, false, false}:  'I',
	{PROTEIN, true, false}:   'J',
	{FRAGMENT, false, false}: 'K',
	{FRAGMENT, true, false}:  'L',
	{FRAGMENT, false, true}:  'M',
	{FRAGMENT, true, true}:   'N',
}

// EncodeHash2 encodes Hash2 as a base58 string. It also adds a single
// letter metadata tag that can be used as an easy heuristic for an LLM to
// identify misbehaving code.
func EncodeHash2(hash [16]byte, err error) (string, error) {
	if err != nil {
		return "", err
	}
	_, sequenceType, circularity, doubleStranded := DecodeFlag(hash[0])
	encoded := encodeToBase58(hash[:])

	return string(Hash2Metadata[Hash2MetadataKey{sequenceType, circularity, doubleStranded}]) + "_" + encoded, nil
}

// DecodeHash2 decodes a seqhash into a [16]byte, including the metadata tag.
func DecodeHash2(encodedString string) ([16]byte, error) {
	// First, we need to decompose the the string into the metadata and the
	// seqhash.
	parts := strings.SplitN(encodedString, "_", 2)
	if len(parts) != 2 {
		return [16]byte{}, errors.New("invalid encoded string format")
	}

	// Decode the Base58 encoded part
	decodedBytes, err := decodeFromBase58(parts[1])
	if err != nil {
		return [16]byte{}, err
	}

	// Ensure decoded bytes fit into a [16]byte array
	if len(decodedBytes) != 16 {
		return [16]byte{}, errors.New("decoded hash does not match expected length")
	}

	var hash [16]byte
	copy(hash[:], decodedBytes)
	return hash, nil
}
