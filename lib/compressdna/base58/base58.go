/*
Package base58 provides functions to encode and decode base58 strings.

Base58 encodes byte arrays into human-readable and human-typable strings. It
originally showed up in bitcoin in order to encode account numbers. The
original bitcoin source code explains its use in comparison to base64:

	// Why base-58 instead of standard base-64 encoding?
	// - Don't want 0OIl characters that look the same in some fonts and
	//      could be used to create visually identical looking account numbers.
	// - A string with non-alphanumeric characters is not as easily accepted as an account number.
	// - E-mail usually won't line-break if there's no punctuation to break at.
	// - Doubleclicking selects the whole number as one word if it's all alphanumeric.
*/
package base58

import (
	"bytes"
	"math/big"
)

const ALPHABET = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

// Encode encodes a byte slice as a base58 string.
func Encode(input []byte) string {
	x := big.NewInt(0).SetBytes(input)
	base := big.NewInt(58)
	zero := big.NewInt(0)
	var result []byte

	// Convert to base58
	mod := new(big.Int)
	for x.Cmp(zero) > 0 {
		x.DivMod(x, base, mod)
		result = append(result, ALPHABET[mod.Int64()])
	}

	// Append "1" for each leading 0 byte in the input
	for i := 0; i < len(input) && input[i] == 0; i++ {
		result = append(result, '1')
	}

	// Reverse the encoded bytes
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return string(result)
}

// Decode decodes a base58 string into a slice array.
func Decode(input string) []byte {
	result := big.NewInt(0)
	zeroByte := byte(ALPHABET[0]) // Convert the first alphabet character to a byte for comparison

	for i := 0; i < len(input); i++ {
		charIndex := bytes.IndexByte([]byte(ALPHABET), input[i]) // input[i] is byte now
		if charIndex == -1 {
			return nil // Character not in Base58 alphabet
		}
		result.Mul(result, big.NewInt(58))
		result.Add(result, big.NewInt(int64(charIndex)))
	}

	decoded := result.Bytes()

	// Add leading zeros
	for i := 0; i < len(input); i++ {
		if input[i] == zeroByte {
			decoded = append([]byte{0x00}, decoded...)
		} else {
			break
		}
	}

	return decoded
}
