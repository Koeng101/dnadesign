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
	"io"
	"os"
	"os/exec"
)

// Minimap2 aligns sequences using minimap2 over the command line. Right
// now, only nanopore (map-ont) is supported. If you need others enabled,
// please put in an issue.
//
// Rarely Minimap2 will stall while reading in fastqInput. See examples of
// how to get around this problem.
func Minimap2(templateFastaInput io.Reader, fastqInput io.Reader, w io.Writer) error {
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
	if _, err := io.Copy(tmpFile, templateFastaInput); err != nil {
		return err
	}
	tmpFile.Close() // Close the file as it's no longer needed

	// Start minimap2 pointing to the temporary file and stdin for sequencing data
	cmd := exec.Command("minimap2", "-ax", "map-ont", tmpFile.Name(), "-")
	cmd.Stdout = w
	cmd.Stdin = fastqInput
	if err := cmd.Start(); err != nil {
		return err
	}

	return cmd.Wait()
}
