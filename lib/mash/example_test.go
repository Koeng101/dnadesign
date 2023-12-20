package mash_test

import (
	"fmt"
	"hash/crc32"

	"github.com/koeng101/dnadesign/lib/mash"
)

func ExampleMash() {
	fingerprint1 := mash.New(17, 10, crc32.NewIEEE())
	fingerprint1.Sketch("ATGCGATCGATCGATCGATCGATCGATCGATCGATCGATCGATCGATCGATCGATCGATCGA")

	fingerprint2 := mash.New(17, 9, crc32.NewIEEE())
	fingerprint2.Sketch("ATGCGATCGATCGATCGATCGATCGATCGATCGATCGATCGATCGATCGATCGATCGATCGA")

	distance := fingerprint1.Distance(fingerprint2)

	fmt.Println(distance)

	// Output:
	// 0
}
