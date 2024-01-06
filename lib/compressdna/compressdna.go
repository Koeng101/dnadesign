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

import "math"

// CompressDNA takes a DNA sequence and converts it into a compressed byte slice.
func CompressDNA(dna string) []byte {
	length := len(dna)
	var lengthBytes []byte
	var flag byte

	// Determine the size of the length field based on the sequence length
	switch {
	case length <= math.MaxUint8:
		lengthBytes = []byte{byte(length)}
		flag = 0x00 // flag for uint8
	case length <= math.MaxUint16:
		lengthBytes = []byte{byte(length >> 8), byte(length)}
		flag = 0x01 // flag for uint16
	case length <= math.MaxUint32:
		lengthBytes = []byte{byte(length >> 24), byte(length >> 16), byte(length >> 8), byte(length)}
		flag = 0x02 // flag for uint32
	default:
		lengthBytes = []byte{byte(length >> 56), byte(length >> 48), byte(length >> 40), byte(length >> 32), byte(length >> 24), byte(length >> 16), byte(length >> 8), byte(length)}
		flag = 0x03 // flag for uint64
	}

	// Encode DNA sequence
	var dnaBytes []byte
	var currentByte byte
	for i, nucleotide := range dna {
		switch nucleotide {
		case 'A': // 00
		case 'T': // 01
			currentByte |= 1 << (6 - (i%4)*2)
		case 'G': // 10
			currentByte |= 2 << (6 - (i%4)*2)
		case 'C': // 11
			currentByte |= 3 << (6 - (i%4)*2)
		}
		if i%4 == 3 || i == length-1 {
			dnaBytes = append(dnaBytes, currentByte)
			currentByte = 0
		}
	}

	// Combine all parts into the result
	result := append([]byte{flag}, lengthBytes...)
	result = append(result, dnaBytes...)
	return result
}

// DecompressDNA takes a compressed byte slice and converts it back to the original DNA string.
func DecompressDNA(compressed []byte) string {
	flag := compressed[0]
	var length int
	var lengthBytes int

	switch flag {
	case 0x00:
		length = int(compressed[1])
		lengthBytes = 1
	case 0x01:
		length = int(compressed[1])<<8 + int(compressed[2])
		lengthBytes = 2
	case 0x02:
		length = int(compressed[1])<<24 + int(compressed[2])<<16 + int(compressed[3])<<8 + int(compressed[4])
		lengthBytes = 4
	case 0x03:
		length = int(compressed[1])<<56 + int(compressed[2])<<48 + int(compressed[3])<<40 + int(compressed[4])<<32 + int(compressed[5])<<24 + int(compressed[6])<<16 + int(compressed[7])<<8 + int(compressed[8])
		lengthBytes = 8
	default:
		return "Invalid flag"
	}

	// Decode DNA sequence
	var dna string
	for i, b := range compressed[1+lengthBytes:] {
		for j := 0; j < 4 && 4*i+j < length; j++ {
			switch (b >> (6 - j*2)) & 0x03 {
			case 0:
				dna += "A"
			case 1:
				dna += "T"
			case 2:
				dna += "G"
			case 3:
				dna += "C"
			}
		}
	}

	return dna
}
