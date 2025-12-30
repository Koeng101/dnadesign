/*
Package bcftools wraps the bcftools cli to be used with Go.

Requires both bcftools and samtools.
*/
package bcftools

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"

	"github.com/koeng101/dnadesign/lib/bio/fasta"
	"golang.org/x/sync/errgroup"
)

// GenerateVCF generates a VCF file from sam alignments.
// Specifically, it runs the following commands, with sam alignments in stdin
// and the templateFasta written to a temporary file:
//
// `samtools view -bS - | samtools sort - | bcftools mpileup -Ou -f tmpFile.fasta -t {template:6-150} - | bcftools call -mv -Ov -`
//
// Here's a breakdown of the process and flags used:
//
// samtools view -bS -:
//   - Converts SAM to BAM format.
//
// samtools sort -:
//   - Sorts the BAM in preparation for bcftools.
//
// bcftools mpileup:
//   - -Ou: Outputs an uncompressed BCF, which is faster for piping.
//   - -f: Reads from the temporary template file.
//   - -t {template:6-150}: Targets certain sections of the template. Larger
//     regions around the region of interest are often used for better alignment.
//
// bcftools call:
//   - -m: Activates the multiallelic caller, better for calling mutations in
//     prone regions.
//   - -v: Outputs variants only, omitting non-variant sites.
//   - -Ov: Sets the output type to VCF text, which is easier to read and parse.
//
// The sequence starts with converting SAM to BAM, sorting it, then using
// bcftools mpileup and call to generate a VCF file from the alignments.
func GenerateVCF(ctx context.Context, startRegion int, endRegion int, templateFasta fasta.Record, samAlignments io.Reader, w io.Writer) error {
	/*
	   Due to how os.exec works in Golang, we can't directly have pipes as if
	   the whole thing was a script. However, we can attach pipes to each
	   command, and move data between all 4. This is how this function works.

	   First, we create a temporary template fasta (named pipes tend to be
	   unreliable). Then, we create each command, set up pipes between them,
	   and then run each in a errGroup as a goroutine.

	   Then, we wait for all the goroutines to finish. They will be sending
	   vcf text to the output io.Writer.
	*/
	// Create a temporary file for the template fasta
	tmpFile, err := os.CreateTemp("", "template_*.fasta")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name()) // Clean up file afterwards

	// Write template fasta data to the temporary file
	_, err = templateFasta.WriteTo(tmpFile)
	if err != nil {
		return err
	}
	tmpFile.Close() // Close the file as it's no longer needed

	// Index the temporary fasta file to generate the .fai file
	indexCmd := exec.Command("samtools", "faidx", tmpFile.Name())
	if err := indexCmd.Run(); err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name() + ".fai") // Ensure to clean up the .fai file afterwards

	g, ctx := errgroup.WithContext(ctx)

	// Setup pipe connections between commands
	viewSortReader, viewSortWriter := io.Pipe()
	sortMpileupReader, sortMpileupWriter := io.Pipe()
	callReader, callWriter := io.Pipe()

	// Define commands with context
	// bcftools mpileup -Ou -f tmpFile.fasta -t {template:6-150} - | bcftools call -mv -Ov -`
	viewCmd := exec.CommandContext(ctx, "samtools", "view", "-bS", "-")
	sortCmd := exec.CommandContext(ctx, "samtools", "sort", "-")
	mpileupCmd := exec.CommandContext(ctx, "bcftools", "mpileup", "-Ou", "-f", tmpFile.Name(), "-t", fmt.Sprintf("%s:%d-%d", templateFasta.Identifier, startRegion, endRegion), "-")
	callCmd := exec.CommandContext(ctx, "bcftools", "call", "-mv", "-Ov", "-")

	// Goroutine for the first command: samtools view
	g.Go(func() error {
		defer viewSortWriter.Close() // ensure the pipe is closed after this function exits

		viewCmd.Stdin = samAlignments
		viewCmd.Stdout = viewSortWriter

		if err := viewCmd.Start(); err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			_ = viewCmd.Process.Signal(syscall.SIGTERM)
			return ctx.Err()
		default:
			return viewCmd.Wait()
		}
	})

	// Goroutine for the second command: samtools sort
	g.Go(func() error {
		defer sortMpileupWriter.Close() // ensure the pipe is closed after this function exits

		sortCmd.Stdin = viewSortReader
		sortCmd.Stdout = sortMpileupWriter

		if err := sortCmd.Start(); err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			_ = sortCmd.Process.Signal(syscall.SIGTERM)
			return ctx.Err()
		default:
			return sortCmd.Wait()
		}
	})

	// Goroutine for the third command: bcftools mpileup
	g.Go(func() error {
		defer callWriter.Close() // ensure the pipe is closed after this function exits

		mpileupCmd.Stdin = sortMpileupReader
		mpileupCmd.Stdout = callWriter

		if err := mpileupCmd.Start(); err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			_ = mpileupCmd.Process.Signal(syscall.SIGTERM)
			return ctx.Err()
		default:
			return mpileupCmd.Wait()
		}
	})

	// Goroutine for the fourth command: bcftools call
	g.Go(func() error {
		callCmd.Stdin = callReader
		callCmd.Stdout = w

		if err := callCmd.Start(); err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			_ = callCmd.Process.Signal(syscall.SIGTERM)
			return ctx.Err()
		default:
			return callCmd.Wait()
		}
	})

	// Wait for all goroutines to complete and return the first non-nil error
	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}
