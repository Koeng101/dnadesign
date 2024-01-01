package minimap2_test

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/koeng101/dnadesign/external/minimap2"
)

// TestMinimap2 tests the Minimap2Raw function
func TestMinimap2(t *testing.T) {
	// Open the template FASTA file
	templateFile, err := os.Open("./data/templates.fasta")
	if err != nil {
		t.Fatalf("Failed to open template FASTA file: %v", err)
	}
	defer templateFile.Close()

	// Open the FASTQ file
	fastqFile, err := os.Open("./data/reads.fastq")
	if err != nil {
		t.Fatalf("Failed to open FASTQ file: %v", err)
	}
	defer fastqFile.Close()

	// Prepare the writer to capture the output
	var buf bytes.Buffer

	// Execute the Minimap2 function
	err = minimap2.Minimap2(templateFile, fastqFile, &buf)
	if err != nil {
		t.Errorf("Minimap2Raw returned an error: %v", err)
	}

	expectedHeader := `@HD	VN:1.6	SO:unsorted	GO:query
@SQ	SN:oligo1	LN:169
@SQ	SN:oligo2	LN:158
@SQ	SN:oligo3	LN:102`

	// Extract the relevant part of the output for comparison
	output := buf.String()
	headerLines := strings.SplitN(output, "\n", 5) // Assuming header is first 4 lines
	if len(headerLines) < 4 {
		t.Fatalf("Output header is too short, got: %v", headerLines)
	}
	outputHeader := strings.Join(headerLines[:4], "\n")

	// Perform comparison
	if outputHeader != expectedHeader {
		t.Errorf("Output header does not match expected header. Got %s, want %s", outputHeader, expectedHeader)
	}

}
