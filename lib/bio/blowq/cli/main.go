package main

import (
	"bufio"
	"os"

	"github.com/klauspost/compress/zstd"
	"github.com/koeng101/dnadesign/lib/bio"
	"github.com/koeng101/dnadesign/lib/bio/blowq"
)

func main() {
	// Initialize a reader from stdin and a writer to stdout
	reader := bufio.NewReader(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)

	// Initialize zstd writer with max compression
	zstdEncoder, _ := zstd.NewWriter(writer, zstd.WithEncoderLevel(zstd.EncoderLevelFromZstd(22)))
	defer zstdEncoder.Close()

	// Initialize the FASTQ parser
	fastqParser := bio.NewFastqParser(reader)
	records, _ := fastqParser.Parse()

	// Write each record to the zstd compressed output
	for _, record := range records {
		_, _ = blowq.WriteTo(zstdEncoder, record)
	}

	// Ensure everything is written and cleanup
	zstdEncoder.Close()
	writer.Flush()
}
