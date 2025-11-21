package seqhash

import (
	"errors"
	"math/big"
	"strings"
)

const alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

// encodeToBase58 encodes a byte slice to a Base58 string
func encodeToBase58(input []byte) string {
	// Convert byte slice to a big.Int
	num := big.NewInt(0).SetBytes(input)
	base := big.NewInt(int64(len(alphabet)))
	mod := &big.Int{}
	var encoded strings.Builder

	// Convert to base58
	for num.Sign() > 0 {
		num.DivMod(num, base, mod)
		encoded.WriteByte(alphabet[mod.Int64()])
	}

	// Add '1' for each leading 0 byte
	for _, b := range input {
		if b != 0 {
			break
		}
		encoded.WriteByte('1')
	}

	// Reverse the encoded string
	result := []byte(encoded.String())
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return string(result)
}

// decodeFromBase58 decodes a Base58 string to a byte slice
func decodeFromBase58(input string) ([]byte, error) {
	if len(input) == 0 {
		return []byte{}, nil
	}

	num := big.NewInt(0)
	base := big.NewInt(int64(len(alphabet)))
	for _, c := range input {
		charIndex := strings.IndexRune(alphabet, c)
		if charIndex == -1 {
			return nil, errors.New("invalid character found")
		}
		num.Mul(num, base)
		num.Add(num, big.NewInt(int64(charIndex)))
	}

	decoded := num.Bytes()
	// Add leading zeros
	if input[0] == '1' {
		leadingZeros := len(input) - len(strings.TrimLeft(input, "1"))
		decoded = append(make([]byte, leadingZeros), decoded...)
	}

	return decoded, nil
}
