package tokenizer

import "testing"

func TestTokenizeProtein(t *testing.T) {
	proteinSequence := "ACDEFGHIKLMNPQRSTVWYUO*BXZ"
	tokens, err := TokenizeProtein(proteinSequence)
	if err != nil {
		t.Errorf("Should have successfully tokenized. Got error: %s", err)
	}
	for i, token := range tokens[1 : len(tokens)-1] {
		// The first amino acid token is 3
		if token != uint16(i+2) {
			t.Errorf("Expected %d, got: %d", i+2, token)
		}
	}
	badProtein := "J" // should fail
	_, err = TokenizeProtein(badProtein)
	if err == nil {
		t.Errorf("Should have failed on J")
	}
}
