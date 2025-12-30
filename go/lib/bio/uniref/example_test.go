package uniref_test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/koeng101/dnadesign/lib/bio/uniref"
)

func Example() {
	// Open the gzipped UniRef file
	file, _ := os.Open(filepath.Join("data", "uniref90_mini.xml"))
	defer file.Close()

	// Create new parser
	parser, _ := uniref.NewParser(file)

	// Read and print the first entry
	entry, _ := parser.Next()

	fmt.Printf("Entry ID: %s\n", entry.ID)
	fmt.Printf("Name: %s\n", entry.Name)
	fmt.Printf("Sequence Length: %d\n", entry.RepMember.Sequence.Length)

	// Output:
	// Entry ID: UniRef50_UPI002E2621C6
	// Name: Cluster: uncharacterized protein LOC134193701
	// Sequence Length: 49499
}
