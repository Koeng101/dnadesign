package seqhash

import (
	"bytes"
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
	if seqhash != "A_LGxts7bxq55Uiq+E94pcYg==" {
		t.Errorf("Circular double stranded hashing failed. Expected A_LGxts7bxq55Uiq+E94pcYg==, got: " + seqhash)
	}
	// Test circular single stranded hashing
	seqhash, _ = EncodeHash2(Hash2("TTAGCCCAT", "DNA", true, false))
	if seqhash != "B_KB3s/EXx/C9wJvVE/gzw7Q==" {
		t.Errorf("Circular single stranded hashing failed. Expected B_KB3s/EXx/C9wJvVE/gzw7Q==, got: " + seqhash)
	}
	// Test linear double stranded hashing
	seqhash, _ = EncodeHash2(Hash2("TTAGCCCAT", "DNA", false, true))
	if seqhash != "C_JN15Uk5YpkXcKaJt0ozLRQ==" {
		t.Errorf("Linear double stranded hashing failed. Expected C_JN15Uk5YpkXcKaJt0ozLRQ==, got: " + seqhash)
	}
	// Test linear single stranded hashing
	seqhash, _ = EncodeHash2(Hash2("TTAGCCCAT", "DNA", false, false))
	if seqhash != "D_IC0pLlPHC/zPQpSqU6hy0A==" {
		t.Errorf("Linear single stranded hashing failed. Expected D_IC0pLlPHC/zPQpSqU6hy0A==, got: " + seqhash)
	}

	// Test RNA Seqhash
	seqhash, _ = EncodeHash2(Hash2("TTAGCCCAT", "RNA", false, false))
	if seqhash != "H_IS0pLlPHC/zPQpSqU6hy0A==" {
		t.Errorf("Linear single stranded hashing failed. Expected H_IS0pLlPHC/zPQpSqU6hy0A==, got: " + seqhash)
	}
	// Test Protein Seqhash
	seqhash, _ = EncodeHash2(Hash2("MGC*", "PROTEIN", false, false))
	if seqhash != "I_IiAwHj+EfYcQCf6Ty64wUg==" {
		t.Errorf("Linear single stranded hashing failed. Expected I_IiAwHj+EfYcQCf6Ty64wUg==, got: " + seqhash)
	}
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
	expectedHash := "K_IwQE3XlSTlimRdwpom3SjA=="
	if sqHash != expectedHash {
		t.Errorf("Expected %s, Got: %s", expectedHash, sqHash)
	}
}
