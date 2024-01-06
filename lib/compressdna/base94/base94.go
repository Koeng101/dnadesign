package base94

import (
	"bytes"
	"errors"
	"math/big"
)

const ALPHABET = "!\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~"

// Encode encodes a byte slice as a base94 string.
func Encode(input []byte) string {
	x := big.NewInt(0).SetBytes(input)
	base := big.NewInt(94) // Updated base to 94
	zero := big.NewInt(0)
	var result []byte

	mod := new(big.Int)
	for x.Cmp(zero) > 0 {
		x.DivMod(x, base, mod)
		result = append(result, ALPHABET[mod.Int64()])
	}

	for i := 0; i < len(input) && input[i] == 0; i++ {
		result = append(result, '!')
	}

	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return string(result)
}

// Decode decodes a base94 string into a slice array and returns an error if encountered.
func Decode(input string) ([]byte, error) {
	result := big.NewInt(0)

	for i := 0; i < len(input); i++ {
		charIndex := bytes.IndexByte([]byte(ALPHABET), input[i])
		if charIndex == -1 {
			// If character is not in the alphabet, return nil and an error
			return nil, errors.New("invalid character in input string")
		}
		result.Mul(result, big.NewInt(94)) // Base is 94
		result.Add(result, big.NewInt(int64(charIndex)))
	}

	// Convert big.Int to a byte slice
	decoded := result.Bytes()

	// Handle leading zero bytes
	zeroByte := ALPHABET[0]
	for i := 0; i < len(input); i++ {
		if input[i] == zeroByte {
			decoded = append([]byte{0x00}, decoded...)
		} else {
			break
		}
	}

	return decoded, nil // Return the decoded slice and no error
}
