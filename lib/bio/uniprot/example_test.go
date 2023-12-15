package uniprot_test

import (
	"fmt"
	"os"

	"github.com/koeng101/dnadesign/lib/bio/uniprot"
)

// This example shows how to open a uniprot data dump file and read the results
// into a list. Directly using the channel without converting to an array
// should be used for the Trembl data dump
func Example_basic() {
	uniprotFile, _ := os.Open("data/uniprot_sprot_mini.xml.gz")
	defer uniprotFile.Close()
	parser, _ := uniprot.NewParser(uniprotFile)
	entry, _ := parser.Next()

	fmt.Println(entry.Accession[0])
	// Output: P0C9F0
}
