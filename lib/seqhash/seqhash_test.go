package seqhash

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/koeng101/dnadesign/lib/bio"
)

func TestHash2(t *testing.T) {
	// Test TNA as sequenceType
	_, err := Hash2("ATGGGCTAA", "TNA", true, true)
	if err == nil {
		t.Errorf("TestHash2() has failed. TNA is not a valid sequenceType.")
	}
	// Test X in DNA or RNA
	_, err = Hash2("XTGGCCTAA", "DNA", true, true)
	if err == nil {
		t.Errorf("TestSeqhashSequenceString() has failed. X is not a valid DNA or RNA sequence character.")
	}
	// Test X in PROTEIN
	_, err = Hash2("MGCJ*", "PROTEIN", false, false)
	if err == nil {
		t.Errorf("TestSeqhashSequenceProteinString() has failed. J is not a valid PROTEIN sequence character.")
		fmt.Println(err)
	}
	// Test double stranded Protein
	_, err = Hash2("MGCS*", "PROTEIN", false, true)
	if err == nil {
		t.Errorf("TestSeqhashProteinDoubleStranded() has failed. Proteins cannot be double stranded.")
	}

	// Test circular double stranded hashing
	seqhash, _ := EncodeHash2(Hash2("TTAGCCCAT", "DNA", true, true))
	if seqhash != "A_6VAbBfXD8BSZh2HJZqgGgR" {
		t.Errorf("Circular double stranded hashing failed. Expected A_6VAbBfXD8BSZh2HJZqgGgR, got: %s", seqhash)
	}
	// Test circular single stranded hashing
	seqhash, _ = EncodeHash2(Hash2("TTAGCCCAT", "DNA", true, false))
	if seqhash != "B_5xKbuHELJCCQWJwQi7W1ak" {
		t.Errorf("Circular single stranded hashing failed. Expected B_5xKbuHELJCCQWJwQi7W1ak, got: %s", seqhash)
	}
	// Test linear double stranded hashing
	seqhash, _ = EncodeHash2(Hash2("TTAGCCCAT", "DNA", false, true))
	if seqhash != "C_5Z2pHCXbxWUPYiZj6J1Nag" {
		t.Errorf("Linear double stranded hashing failed. Expected C_5Z2pHCXbxWUPYiZj6J1Nag, got: %s", seqhash)
	}
	// Test linear single stranded hashing
	seqhash, _ = EncodeHash2(Hash2("TTAGCCCAT", "DNA", false, false))
	if seqhash != "D_4yT7etihWZHHNXUpbM5tUf" {
		t.Errorf("Linear single stranded hashing failed. Expected D_4yT7etihWZHHNXUpbM5tUf, got: %s", seqhash)
	}

	// Test RNA Seqhash
	seqhash, _ = EncodeHash2(Hash2("TTAGCCCAT", "RNA", false, false))
	if seqhash != "H_56cWv4dacvRJxUUcXYsdP5" {
		t.Errorf("Linear single stranded hashing failed. Expected H_56cWv4dacvRJxUUcXYsdP5, got: %s", seqhash)
	}
	// Test Protein Seqhash
	seqhash, _ = EncodeHash2(Hash2("MGC*", "PROTEIN", false, false))
	if seqhash != "I_5DQsEyDHLh2r4njCcupAuF" {
		t.Errorf("Linear single stranded hashing failed. Expected I_5DQsEyDHLh2r4njCcupAuF, got: %s", seqhash)
	}
}

func TestEncodeAndDecode(t *testing.T) {
	rawBytes, err := Hash2("ATGC", "DNA", false, true)
	if err != nil {
		t.Errorf("Got bad hash: %s", err)
	}
	encoded, err := EncodeHash2(rawBytes, err)
	if err != nil {
		t.Errorf("Failed to encode: %s", err)
	}
	decoded, err := DecodeHash2(encoded)
	if err != nil {
		t.Errorf("Failed to decode: %s", err)
	}
	for i := range rawBytes {
		if rawBytes[i] != decoded[i] {
			t.Errorf("Failed to decode properly.")
		}
	}
	_, err = EncodeHash2([16]byte{}, errors.New("test"))
	if err == nil {
		t.Errorf("should fail on test error")
	}

	// Test no metadata
	_, err = DecodeHash2("")
	if err == nil {
		t.Errorf("should fail on no metadata")
	}
	// Test empty decode
	_, err = DecodeHash2("A_")
	if err == nil {
		t.Errorf("should fail on empty data")
	}
	// Test bad char
	_, err = DecodeHash2("A_/")
	if err == nil {
		t.Errorf("should fail on bad character")
	}
	// Test 1s
	_, err = DecodeHash2("A_11111")
	if err == nil {
		t.Errorf("should fail on 1s because length is wrong.")
	}

	// just to make sure gocov goes through
	_ = encodeToBase58([]byte{0, 0, 0, 0})
}

func TestLeastRotation(t *testing.T) {
	file, _ := os.Open("../data/puc19.gbk")
	defer file.Close()
	parser := bio.NewGenbankParser(file)
	sequence, _ := parser.Next()
	var sequenceBuffer bytes.Buffer

	sequenceBuffer.WriteString(sequence.Sequence)
	bufferLength := sequenceBuffer.Len()

	var rotatedSequence string
	for elementIndex := 0; elementIndex < bufferLength; elementIndex++ {
		bufferElement, _, _ := sequenceBuffer.ReadRune()
		sequenceBuffer.WriteRune(bufferElement)
		if elementIndex == 0 {
			rotatedSequence = RotateSequence(sequenceBuffer.String())
		} else {
			newRotatedSequence := RotateSequence(sequenceBuffer.String())
			if rotatedSequence != newRotatedSequence {
				t.Errorf("TestLeastRotation() has failed. rotationSequence mutated.")
			}
		}
	}
}

func TestFlagEncoding(t *testing.T) {
	version := 2
	sequenceType := DNA
	circularity := true
	doubleStranded := true
	flag := EncodeFlag(version, sequenceType, circularity, doubleStranded)
	decodedVersion, decodedSequenceType, decodedCircularity, decodedDoubleStranded := DecodeFlag(flag)
	if (decodedVersion != version) || (decodedSequenceType != sequenceType) || (decodedCircularity != circularity) || (decodedDoubleStranded != doubleStranded) {
		t.Errorf("Got different decoded flag.")
	}
}

func TestHash2Fragment(t *testing.T) {
	// Test X failure
	_, err := Hash2Fragment("ATGGGCTAX", 4, 4)
	if err == nil {
		t.Errorf("TestHash2Fragment() has failed. X is not a valid sequenceType.")
	}
	// Test actual hash
	sqHash, _ := EncodeHash2(Hash2Fragment("ATGGGCTAA", 4, 4))
	expectedHash := "K_5KnZQEnPRzJSYPkbPwLCJF"
	if sqHash != expectedHash {
		t.Errorf("Expected %s, Got: %s", expectedHash, sqHash)
	}

	// Test another hash
	sqHash, _ = EncodeHash2(Hash2Fragment("TTAGCCCAT", 4, 4))
	expectedHash = "K_5KnZQEnPRzJSYPkbPwLCJF"
	if sqHash != expectedHash {
		t.Errorf("Expected %s, Got: %s", expectedHash, sqHash)
	}
}
