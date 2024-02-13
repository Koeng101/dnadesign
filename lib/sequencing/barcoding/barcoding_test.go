package barcoding

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/koeng101/dnadesign/lib/bio"
)

//go:embed data/dual_barcodes.csv
var dualBarcodes string

//go:embed data/dual_reads.fastq
var dualReads string

func ExampleDualBarcodeSequence() {
	primerSet, _ := ParseDualPrimerSet(strings.NewReader(dualBarcodes))
	parser := bio.NewFastqParser(strings.NewReader(dualReads))
	records, _ := parser.Parse()

	var wells []string
	for _, record := range records[0:10] {
		well, _ := DualBarcodeSequence(record.Sequence, primerSet)
		if well != "" {
			wells = append(wells, well)
		}
	}

	fmt.Println(wells)
	// Output: [B15 O1 O1 J22 C22 E20 A15]
}
