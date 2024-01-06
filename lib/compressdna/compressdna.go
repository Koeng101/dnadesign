/*
Package compressdna contains the CompressDNA algorithm.

The CompressDNA algorithm efficiently codes a known DNA sequence into a byte
slice. It does this by encoding each base pair as 2 bits, which are packed into
a slice of bytes. The first byte in the final byte slice contains a flag, which
contains information on the uint length of the final sequence - since sequences
will often not perfectly fit into the unit-4 byte slice.

Flag 0x00 is a uint8 length end sequence (1 byte), with 0x01, 0x02, 0x03
standing for uint16, uint32, and uint64, respectively. The next bytes after the
flag depend on the flag definition: for example, the next byte after flag 0x00
(uint8) will be the length, which the next 2 bytes after flag 0x01 (uint16)
will be the length, 4 bytes for flag 0x02 (uint32), and 8 bytes for flag 0x03
(uint64). The chosen uint unit is dependent on the length of the sequence. A
simple example would look like this, with brackets between bytes:

	[flag 0x00] [uint8 (14)] [4bp] [4bp] [4bp] [2bp] = 6 bytes

CompressDNA is not able to handle ambigious base pairs. It is designed to only
compress specific DNA sequences, like you would find in a fastq file, or a
plasmid file.
*/
package compressdna

import (
	"errors"
	"math"

	"github.com/koeng101/dnadesign/lib/compressdna/base94"
)

const (
	FourCharFlag = 0x00
	FiveCharFlag = 0x10
)

// CompressDNA takes a DNA sequence and converts it into a compressed byte slice.
func CompressDNA(dna string, fiveCharAlphabet bool) ([]byte, error) {
	length := len(dna)
	if length == 0 {
		return nil, errors.New("DNA sequence is empty")
	}

	var lengthBytes []byte
	var flag byte

	// Set flag for alphabet type
	if fiveCharAlphabet {
		flag = FiveCharFlag
	} else {
		flag = FourCharFlag
	}

	// Determine the size of the length field based on the sequence length
	switch {
	case length <= math.MaxUint8:
		lengthBytes = []byte{byte(length)}
	case length <= math.MaxUint16:
		lengthBytes = []byte{byte(length >> 8), byte(length)}
		flag |= 0x01 // modify flag for uint16
	case length <= math.MaxUint32:
		lengthBytes = []byte{byte(length >> 24), byte(length >> 16), byte(length >> 8), byte(length)}
		flag |= 0x02 // modify flag for uint32
	default:
		return nil, errors.New("sequence too long")
	}

	// Encode DNA sequence
	shift := 2
	if fiveCharAlphabet {
		shift = 3 // Use 3 bits for 5char alphabet
	}
	var dnaBytes []byte
	var currentByte byte
	bitCount := 0
	for _, nucleotide := range dna {
		var bits byte
		switch nucleotide {
		case 'A': // 00 or 000
		case 'T': // 01 or 001
			bits = 1
		case 'G': // 10 or 010
			bits = 2
		case 'C': // 11 or 011
			bits = 3
		case 'N': // only valid for 5char alphabet, 100
			if !fiveCharAlphabet {
				return nil, errors.New("invalid character in DNA sequence: " + string(nucleotide))
			}
			bits = 4
		default:
			return nil, errors.New("invalid character in DNA sequence: " + string(nucleotide))
		}
		currentByte = (currentByte << shift) | bits
		bitCount += shift
		if bitCount == 8 {
			dnaBytes = append(dnaBytes, currentByte)
			currentByte = 0
			bitCount = 0
		}
	}
	// Handle last byte
	if bitCount > 0 {
		currentByte <<= (8 - bitCount)
		dnaBytes = append(dnaBytes, currentByte)
	}

	// Combine all parts into the result
	result := append([]byte{flag}, lengthBytes...)
	result = append(result, dnaBytes...)
	return result, nil
}

// DecompressDNA takes a compressed byte slice and converts it back to the original DNA string.
func DecompressDNA(compressed []byte) (string, error) {
	if len(compressed) < 2 { // Needs at least a flag and one length byte
		return "", errors.New("invalid compressed data")
	}

	flag := compressed[0]
	var length int
	var lengthBytes int

	// Determining length based on flag
	switch flag & 0x0F { // Masking to only look at the length bits
	case 0x00:
		length = int(compressed[1])
		lengthBytes = 1
	case 0x01:
		length = int(compressed[1])<<8 + int(compressed[2])
		lengthBytes = 2
	case 0x02:
		length = int(compressed[1])<<24 + int(compressed[2])<<16 + int(compressed[3])<<8 + int(compressed[4])
		lengthBytes = 4
	default:
		return "", errors.New("invalid flag value")
	}

	fiveCharAlphabet := (flag & 0xF0) == FiveCharFlag // Determine if it's 5char alphabet

	// Decode DNA sequence
	shift := 2
	bitMask := byte(math.Pow(2, float64(shift)) - 1) // Bitmask to extract nucleotide bits
	if fiveCharAlphabet {
		shift = 3
		bitMask = byte(math.Pow(2, float64(shift)) - 1) // Update bitmask for 5char
	}
	var dna string
	var currentByte byte
	bitCount := 0
	for i := 0; i < length; i++ {
		if bitCount == 0 {
			currentByte = compressed[1+lengthBytes]
			lengthBytes++
			bitCount = 8
		}
		bits := (currentByte >> (bitCount - shift)) & bitMask
		bitCount -= shift
		switch bits {
		case 0:
			dna += "A"
		case 1:
			dna += "T"
		case 2:
			dna += "G"
		case 3:
			dna += "C"
		case 4:
			if !fiveCharAlphabet {
				return "", errors.New("invalid encoding for 4char alphabet")
			}
			dna += "N"
		default:
			return "", errors.New("invalid bits for nucleotide")
		}
		if bitCount <= 0 {
			bitCount = 0 // Reset for next byte
		}
	}

	return dna, nil
}

// CompressDNAWithQuality takes a DNA sequence and a base94 encoded quality string and converts it into a compressed byte slice.
func CompressDNAWithQuality(dna, qualityBase94 string, fiveCharAlphabet bool) ([]byte, error) {
	compressedDNA, err := CompressDNA(dna, fiveCharAlphabet)
	if err != nil {
		return nil, err
	}

	// Decode base94 quality string back to bytes
	qualityBytes, err := base94.Decode(qualityBase94)
	if err != nil {
		return nil, err
	}

	// Preallocate the result slice with the exact size needed
	result := make([]byte, len(compressedDNA)+len(qualityBytes))

	// Copy the compressed DNA and quality bytes into the result slice
	copy(result, compressedDNA)
	copy(result[len(compressedDNA):], qualityBytes)

	return result, nil
}

func DecompressDNAWithQuality(compressed []byte) (string, string, error) {
	if len(compressed) < 2 {
		return "", "", errors.New("invalid compressed data")
	}

	flag := compressed[0]
	var lengthBytes int
	var dnaLength int

	// Determining length based on flag
	switch flag & 0x0F { // Masking to only look at the length bits
	case 0x00:
		dnaLength = int(compressed[1])
		lengthBytes = 1
	case 0x01:
		dnaLength = int(compressed[1])<<8 + int(compressed[2])
		lengthBytes = 2
	case 0x02:
		dnaLength = int(compressed[1])<<24 + int(compressed[2])<<16 + int(compressed[3])<<8 + int(compressed[4])
		lengthBytes = 4
	default:
		return "", "", errors.New("invalid flag value")
	}

	fiveCharAlphabet := (flag & 0xF0) == FiveCharFlag
	shift := 2
	if fiveCharAlphabet {
		shift = 3
	}

	// Calculate the length of the DNA bytes in the compressed format
	dnaBytesLength := (dnaLength*shift + 7) / 8 // Calculate total bits and convert to bytes, round up

	dna, err := DecompressDNA(compressed[:1+lengthBytes+dnaBytesLength])
	if err != nil {
		return "", "", err
	}

	// Extract the quality bytes and encode them to base94
	qualityBytes := compressed[1+lengthBytes+dnaBytesLength:]
	qualityBase94 := base94.Encode(qualityBytes) // Assuming a base94 package exists

	return dna, qualityBase94, nil
}
