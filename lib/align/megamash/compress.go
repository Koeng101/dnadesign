package megamash

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
	lengthBytes := 1

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
