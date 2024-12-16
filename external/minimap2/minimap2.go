/*
Package minimap2 contains functions for working with minimap2.

minimap2 is a DNA alignment package written by Heng Li for aligning nanopore
reads as the spirtual successor to bwa-mem, which is a widely used alignment
algorithm for illumina sequencing reads.

minimap2 takes in fasta reference genomes and aligns them with fastq reads,
outputting a sam alignment file. In this package, all io is handled with
standard library io.Reader and io.Writer, both of which can be used with
dnadesign `bio` parsers. Data should be piped in using data `WriteTo`
functions, and can be read using a sam parser.

We use `os.Exec` instead of cgo in order to make the package simpler, and
also because the overhead of launching is minimal in comparison to how much
data is expected to run through minimap2.

For more information on minimap2, please visit Heng Li's git: https://github.com/lh3/minimap2
*/
package minimap2

import (
	"context"
	"io"
	"os"
	"os/exec"

	"github.com/koeng101/dnadesign/lib/bio"
	"github.com/koeng101/dnadesign/lib/bio/fasta"
	"github.com/koeng101/dnadesign/lib/bio/fastq"
	"github.com/koeng101/dnadesign/lib/bio/sam"
	"golang.org/x/sync/errgroup"
)

// Minimap2 aligns sequences using minimap2 over the command line. Right
// now, only nanopore (map-ont) is supported. If you need others enabled,
// please put in an issue.
func Minimap2(templateFastas []fasta.Record, fastqInput io.Reader, w io.Writer) error {
	/*
		Generally, this is how the function works:
		1. Create a temporary file for templates. Templates are rather small,
		   and environments will probably have a filesystem, and minimap2
		   sometimes randomly fails if you don't it as a file on the system.
		2. Start minimap2, capturing both stdout and stdin.
		3. Write fastqInput to stdin of minimap2.
		4. Copy stdout of minimap2 to w.
		5. Complete.
	*/

	// Create temporary file for templates.
	/*
		This took me a while to figure out. For whatver reason, named pipes
		don't work: they occasionally stall out minimap2 (like 1/10 the time).
		Writing a temporary file always works, for whatever reason. Stdin
		for sequencing files seems to work rather well.
	*/

	// Create a temporary file for the template fasta
	tmpFile, err := os.CreateTemp("", "template_*.fasta")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name()) // Clean up file afterwards

	// Write template fasta data to the temporary file
	for _, templateFasta := range templateFastas {
		_, err = templateFasta.WriteTo(tmpFile)
		if err != nil {
			return err
		}
	}
	tmpFile.Close() // Close the file as it's no longer needed

	// Start minimap2 pointing to the temporary file and stdin for sequencing data
	cmd := exec.Command("minimap2", "--cs=short", "-K", "100", "-ax", "map-ont", tmpFile.Name(), "-")
	cmd.Stdout = w
	cmd.Stdin = fastqInput
	if err := cmd.Start(); err != nil {
		return err
	}

	return cmd.Wait()
}

// Minimap2Channeled uses channels rather than io.Reader and io.Writers.
func Minimap2Channeled(ctx context.Context, fastaTemplates []fasta.Record, fastqChan <-chan fastq.Read, samChan chan<- sam.Alignment) error {
	g, ctx := errgroup.WithContext(ctx)

	// Create a pipe for writing fastq reads and reading them as an io.Reader
	fastqPr, fastqPw := io.Pipe()

	// Goroutine to consume fastq reads and write them to the PipeWriter
	g.Go(func() error {
		defer fastqPw.Close()
		for read := range fastqChan {
			_, err := read.WriteTo(fastqPw)
			if err != nil {
				return err // return error to be handled by errgroup
			}
		}
		return nil
	})

	// Create a pipe for SAM alignments.
	samPr, samPw := io.Pipe()

	// Use Minimap2 function to process the reads and write SAM alignments.
	g.Go(func() error {
		defer samPw.Close()
		return Minimap2(fastaTemplates, fastqPr, samPw) // Minimap2 writes to samPw
	})

	// Create a SAM parser from samPr (the PipeReader connected to Minimap2 output).
	samParser, err := bio.NewSamParser(samPr)
	if err != nil {
		return err
	}

	// Parsing SAM and sending to channel.
	g.Go(func() error {
		return samParser.ParseToChannel(ctx, samChan, false)
	})

	// Wait for all goroutines in the group to finish.
	if err := g.Wait(); err != nil {
		return err // This will return the first non-nil error from the group of goroutines
	}

	// At this point, all goroutines have finished successfully
	return nil
}
