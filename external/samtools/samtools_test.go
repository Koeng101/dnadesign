package samtools_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/koeng101/dnadesign/external/samtools"
	"github.com/koeng101/dnadesign/lib/bio"
)

func TestPileup(t *testing.T) {
	// Open the template FASTA file
	templateFile, err := os.Open("./data/template.fasta")
	if err != nil {
		t.Fatalf("Failed to open template FASTA file: %v", err)
	}
	defer templateFile.Close()

	// Open the sam file
	samFile, err := os.Open("./data/aln.sam")
	if err != nil {
		t.Fatalf("Failed to open sam alignment file: %v", err)
	}
	defer samFile.Close()

	// Prepare the writer to capture the output
	var buf bytes.Buffer

	// Execute the pileup function
	err = samtools.Pileup(templateFile, samFile, &buf)
	if err != nil {
		t.Errorf("Pileup returned error: %s", err)
	}

	// Read as pileup file
	parser := bio.NewPileupParser(&buf)
	lines, err := parser.Parse()
	if err != nil {
		t.Errorf("Failed while parsing: %s", err)
	}

	expectedQuality := "3555457667367556657568659884340:7"
	if lines[5].Quality != expectedQuality {
		t.Errorf("Bad quality return. Got: %s Expected: %s", lines[5].Quality, expectedQuality)
	}
}
