package transform_test

import (
	"fmt"

	"github.com/koeng101/dnadesign/lib/transform"
)

func ExampleReverseComplement() {
	sequence := "GATTACA"
	reverseComplement := transform.ReverseComplement(sequence)
	fmt.Println(reverseComplement)

	// Output: TGTAATC
}

func ExampleComplement() {
	sequence := "GATTACA"
	complement := transform.Complement(sequence)
	fmt.Println(complement)

	// Output: CTAATGT
}

func ExampleReverse() {
	sequence := "GATTACA"
	reverse := transform.Reverse(sequence)
	fmt.Println(reverse)

	// Output: ACATTAG
}
