package bcftools_test

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/brentp/vcfgo"
	"github.com/koeng101/dnadesign/external/bcftools"
	"github.com/koeng101/dnadesign/lib/bio"
)

func TestGenerateVCF(t *testing.T) {
	// Open the template FASTA file
	templateFile, err := os.Open("./data/template.fasta")
	if err != nil {
		t.Fatalf("Failed to open template FASTA file: %v", err)
	}
	fastaRecord, _ := bio.NewFastaParser(templateFile).Next()
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
	ctx := context.Background()
	err = bcftools.GenerateVCF(ctx, 6, 150, fastaRecord, samFile, &buf)
	if err != nil {
		t.Errorf("GenerateVCF returned error: %s", err)
	}

	// Read as vcf file
	// This turns out to be one of the only good file format parsers in Go for
	// bioinformatics. So we use it rather than writing our own.
	rdr, err := vcfgo.NewReader(&buf, false)
	if err != nil {
		t.Errorf("Failed to open new reader")
	}
	var variantPosition uint64
	for {
		variant := rdr.Read()
		if variant == nil {
			break
		}
		variantPosition = variant.Pos
	}
	if variantPosition != 136 {
		t.Errorf("Should have been error at position 136")
	}
}
