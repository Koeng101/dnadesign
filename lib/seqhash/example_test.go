package seqhash_test

import (
	"fmt"
	"os"

	"github.com/koeng101/dnadesign/lib/bio"
	"github.com/koeng101/dnadesign/lib/seqhash"
)

// This example shows how to seqhash a sequence.
func Example_basic() {
	sequence := "ATGC"
	sequenceType := seqhash.DNA
	circular := false
	doubleStranded := true

	sequenceSeqhash, _ := seqhash.EncodeHash2(seqhash.Hash2(sequence, sequenceType, circular, doubleStranded))
	fmt.Println(sequenceSeqhash)
	// Output: C_JJgg9ahMxAQzDm2XveE7WA==
}

func ExampleRotateSequence() {
	file, _ := os.Open("../data/puc19.gbk")
	defer file.Close()
	parser := bio.NewGenbankParser(file)
	sequence, _ := parser.Next()

	sequenceLength := len(sequence.Sequence)
	testSequence := sequence.Sequence[sequenceLength/2:] + sequence.Sequence[0:sequenceLength/2]

	fmt.Println(seqhash.RotateSequence(sequence.Sequence) == seqhash.RotateSequence(testSequence))
	// output: true
}

func ExampleHash2() {
	sequence := "ATGC"
	sequenceType := seqhash.DNA
	circular := false
	doubleStranded := true

	sequenceSeqhash, _ := seqhash.Hash2(sequence, sequenceType, circular, doubleStranded)
	fmt.Println(sequenceSeqhash)
	// Output: [36 152 32 245 168 76 196 4 51 14 109 151 189 225 59 88]
}
