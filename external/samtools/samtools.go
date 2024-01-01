/*
Package samtools wraps the samtools cli to be used with Go.
*/
package samtools

import (
	"context"
	"io"
	"os"
	"os/exec"
	"syscall"

	"golang.org/x/sync/errgroup"
)

// Pileup runs a
func Pileup(templateFastas io.Reader, samAlignments io.Reader, w io.Writer) error {
	// Create a temporary file for the template fasta
	tmpFile, err := os.CreateTemp("", "template_*.fasta")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name()) // Clean up file afterwards

	// Write template fasta data to the temporary file
	if _, err := io.Copy(tmpFile, templateFastas); err != nil {
		return err
	}
	tmpFile.Close() // Close the file as it's no longer needed

	g, ctx := errgroup.WithContext(context.Background())

	// Setup pipe connections between commands
	viewSortReader, viewSortWriter := io.Pipe()
	sortMpileupReader, sortMpileupWriter := io.Pipe()

	// Define commands with context
	viewCmd := exec.CommandContext(ctx, "samtools", "view", "-bF", "4")
	sortCmd := exec.CommandContext(ctx, "samtools", "sort", "-")
	mpileupCmd := exec.CommandContext(ctx, "samtools", "mpileup", "-f", tmpFile.Name(), "-")

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
			viewCmd.Process.Signal(syscall.SIGTERM)
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
			sortCmd.Process.Signal(syscall.SIGTERM)
			return ctx.Err()
		default:
			return sortCmd.Wait()
		}
	})

	// Goroutine for the third command: samtools mpileup
	g.Go(func() error {
		mpileupCmd.Stdin = sortMpileupReader
		mpileupCmd.Stdout = w

		if err := mpileupCmd.Start(); err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			mpileupCmd.Process.Signal(syscall.SIGTERM)
			return ctx.Err()
		default:
			return mpileupCmd.Wait()
		}
	})

	// Wait for all goroutines to complete and return the first non-nil error
	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}
