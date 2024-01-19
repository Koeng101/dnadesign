package minimap2_test

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/koeng101/dnadesign/external/minimap2"
)

func ExampleMinimap2() {
	// Get template io.Reader
	templateFile, _ := os.Open("./data/templates.fasta")
	defer templateFile.Close()

	// Get fastq reads io.Reader
	fastqFile, _ := os.Open("./data/reads.fastq")
	defer fastqFile.Close()

	// Create output buffer
	var buf bytes.Buffer

	// Execute the Minimap2Raw function
	_ = minimap2.Minimap2(templateFile, fastqFile, &buf)
	output := buf.String()
	line2 := strings.Split(output, "\n")[2]

	fmt.Println(line2)
	// Output: @SQ	SN:oligo2	LN:158
}
