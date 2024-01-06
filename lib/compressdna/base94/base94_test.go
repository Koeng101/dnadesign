package base94

import "testing"

func TestReEncode(t *testing.T) {
	encodedOriginal := "$$&%&%#$)*59;/767C378411,***,('11<;:,0039/0&()&'2(/*((4.1.09751).601+'#&&&,-**/0-+3558,/)+&)'&&%&$$'%'%'&*/5978<9;**'3*'&&A?99:;:97:278?=9B?CLJHGG=9<@AC@@=>?=>D>=3<>=>3362$%/((+/%&+//.-,%-4:+..000,&$#%$$%+*)&*0%.//*?<<;>DE>.8942&&//074&$033)*&&&%**)%)962133-%'&*99><<=1144??6.027639.011/-)($#$(/422*4;:=122>?@6964:.5'8:52)*675=:4@;323&&##'.-57*4597)+0&:7<7-550REGB21/0+*79/&/6538())+)+23665+(''$$$'-2(&&*-.-#$&%%$$,-)&$$#$'&,);;<C<@454)#"

	// Decoding
	decodedBytes, _ := Decode(encodedOriginal)

	// Re-Encoding
	reEncoded := Encode(decodedBytes)

	// Comparing the re-encoded message to the original encoded string
	if reEncoded != encodedOriginal {
		t.Errorf("Failed! Original Encoded: %s, Re-Encoded: %s", encodedOriginal, reEncoded)
	}
}

// TestLeadingZerosEncoding tests encoding of leading zeros.
func TestLeadingZerosEncoding(t *testing.T) {
	original := []byte{0x00, 0x00, 0x01} // leading zeros
	encoded := Encode(original)
	expectedPrefix := "!!" // as '!' represents a 0 byte in the custom alphabet

	if encoded[:2] != expectedPrefix {
		t.Errorf("Encoding leading zeros failed. Expected prefix: %s, Got: %s", expectedPrefix, encoded[:2])
	}
}

// TestInvalidCharacterDecoding tests decoding with an invalid character.
func TestInvalidCharacterDecoding(t *testing.T) {
	invalidEncoded := "InvalidCharacterNotInAlphabet\n" // Ensure this contains characters not in your alphabet
	decoded, err := Decode(invalidEncoded)

	if err == nil || decoded != nil {
		t.Errorf("Expected error for invalid characters, got nil error and decoded value: %v", decoded)
	}
}

// TestLeadingZerosDecoding tests decoding of leading special characters.
func TestLeadingZerosDecoding(t *testing.T) {
	// '!' represents the zero byte in our custom base94 encoding
	encoded := "!!abcdefgh" // This should decode into something that has leading zeros
	decoded, _ := Decode(encoded)

	if len(decoded) > 0 && decoded[0] != 0x00 {
		t.Errorf("Decoding leading special characters failed. Expected leading 0x00, found: %x", decoded[0])
	}
}
