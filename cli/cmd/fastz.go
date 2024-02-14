package cmd

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/google/uuid"
	"github.com/klauspost/compress/zstd"
	"github.com/koeng101/dnadesign/lib/align/megamash"
	"github.com/koeng101/dnadesign/lib/bio"
	"github.com/koeng101/dnadesign/lib/bio/ddidx"
	"github.com/koeng101/dnadesign/lib/bio/fasta"
	"github.com/koeng101/dnadesign/lib/bio/fastq"
	"github.com/koeng101/dnadesign/lib/sequencing"
	"github.com/koeng101/dnadesign/lib/sequencing/barcoding"
	"github.com/spf13/cobra"
	"gitlab.com/rackn/seekable-zstd"
	"golang.org/x/sync/errgroup"
)

// fastzCmd represents the fastz command
var fastzCmd = &cobra.Command{
	Use:   "fastz",
	Short: "Compresses FASTQ files using zstd with additional indexing",
	Long: `fastz is a tool for compressing FASTQ files using zstd compression, while also generating a .ddidx index file.
The command requires a primer set file and a template map file to function properly. The output is a zstd compressed FASTQ file streamed to stdout, and a .ddidx index file is generated at the specified output location.

This command also supports optional parameters for adjusting the k-mer size and threshold used in megamash, as well as a score parameter for filtering.

Usage example:
cat input.fastq | ./dnadesign fastz --primerSet path/to/primerSet --templateMap path/to/templateMap --ddidxOutput path/to/output.ddidx --kmerSize 16 --threshold 10 --score 0.8 > output.fastq.zstd`,
	Run: func(cmd *cobra.Command, args []string) {
		// You can retrieve the flag values here and add your logic for processing the FASTQ file
		primerSetCsvLocation, _ := cmd.Flags().GetString("primerSet")
		templateMapLocation, _ := cmd.Flags().GetString("templateMap")
		ddidxOutputLocation, _ := cmd.Flags().GetString("ddidxOutput")
		kmerSize, _ := cmd.Flags().GetUint("kmerSize")
		threshold, _ := cmd.Flags().GetUint("threshold")
		score, _ := cmd.Flags().GetFloat64("score")
		cpus, _ := cmd.Flags().GetInt("cpus")

		// Open the primerSet CSV file
		primerSetCsv, err := os.Open(primerSetCsvLocation)
		if err != nil {
			// Handle error
			fmt.Println("Error opening primer set CSV:", err)
			return
		}
		defer primerSetCsv.Close() // Make sure to close the file when you're done

		// Open the templateMap file
		templateMap, err := os.Open(templateMapLocation)
		if err != nil {
			// Handle error
			fmt.Println("Error opening template map:", err)
			return
		}
		defer templateMap.Close() // Make sure to close the file when you're done

		// Create/Open the ddidxOutput file for writing
		// If you only need to write to it, use os.Create to create or truncate an existing file
		ddidxOutput, err := os.Create(ddidxOutputLocation)
		if err != nil {
			// Handle error
			fmt.Println("Error creating/opening ddidx output file:", err)
			return
		}
		defer ddidxOutput.Close() // Make sure to close the file when you're done

		/*
			Step 1: Parse initial data sets
		*/
		// Read primer set
		primerSet, err := barcoding.ParseDualPrimerSet(primerSetCsv)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing primerset: %v\n", err)
			os.Exit(1)
		}

		// Read template map
		var templates []fasta.Record
		reader := csv.NewReader(templateMap)

		for {
			// Read each record from csv
			record, err := reader.Read()
			// Break the loop at the end of the file
			if err == io.EOF {
				break
			}
			// Handle any other error
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing templateMap: %v\n", err)
				os.Exit(1)
			}

			if len(record) == 2 {
				templates = append(templates, fasta.Record{Identifier: record[0], Sequence: record[1]})
			}
		}

		/*
			Step 2: setup megamash
		*/
		m, err := megamash.NewMegamashMap(templates, kmerSize, threshold, score)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating megamash: %v\n", err)
			os.Exit(1)
		}

		/*
			Step 3: setup concurrent processing.
		*/
		parser := bio.NewFastqParser(os.Stdin)
		ctx := context.Background()
		errorGroup, ctx := errgroup.WithContext(ctx)

		fastqReads := make(chan fastq.Read)
		fastqBarcoded := make(chan fastq.Read)
		fastqBarcodedFiltered := make(chan fastq.Read)
		fastqBarcodedFilteredMegamashed := make(chan fastq.Read)

		// Read fastqs into channel
		errorGroup.Go(func() error {
			return parser.ParseToChannel(ctx, fastqReads, false)
		})
		// Barcoding can be an expensive operation
		errorGroup.Go(func() error {
			// We're going to start multiple workers within this errorGroup. This
			// helps when doing computationally intensive operations on channels.
			return bio.RunWorkers(ctx, cpus, fastqBarcoded, func(ctx context.Context) error {
				return sequencing.DualBarcodeFastq(ctx, primerSet, fastqReads, fastqBarcoded)
			})
		})
		// Filtering is a cheap operation, so we only have 1 worker doing it.
		errorGroup.Go(func() error {
			return bio.RunWorkers(ctx, 1, fastqBarcodedFiltered, func(ctx context.Context) error {
				return bio.FilterData(ctx, fastqBarcoded, fastqBarcodedFiltered, func(data fastq.Read) bool {
					_, ok := data.Optionals["dual_barcode"]
					return ok
				})
			})
		})
		// Megamash is very expensive, so we spawn many works to do it.
		errorGroup.Go(func() error {
			return bio.RunWorkers(ctx, cpus, fastqBarcodedFilteredMegamashed, func(ctx context.Context) error {
				return sequencing.MegamashFastq(ctx, m, fastqBarcodedFiltered, fastqBarcodedFilteredMegamashed)
			})
		})

		/*
			Step 4: Write to stdout
		*/
		// Setup seekable zstd
		// Initialize the zstd encoder with desired settings
		encoder, err := zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.SpeedBestCompression))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create zstd encoder: %v\n", err)
			os.Exit(1)
		}
		defer encoder.Close()

		// Create a seekable zstd writer on the temp file
		writer, err := seekable.NewWriter(os.Stdout, encoder)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create seekable zstd writer: %v\n", err)
			os.Exit(1)
		}
		// Now write to stdout
		var indexes []ddidx.Index
		var startPosition uint64
		for read := range fastqBarcodedFilteredMegamashed {
			writtenBytes, err := read.WriteTo(writer)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error writing to stdout: %v\n", err)
				os.Exit(1)
			}
			identifierBytes, err := uuid.Parse(read.Identifier)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Identifier cannot be written as 16byte uuid: %s . Got error: %v\n", read.Identifier, err)
				os.Exit(1)
			}
			indexes = append(indexes, ddidx.Index{Identifier: identifierBytes, StartPosition: startPosition, Length: uint64(writtenBytes)})
			startPosition = startPosition + uint64(writtenBytes)
		}
		// Close the writer to flush the seek table
		if err := writer.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to close seekable zstd writer: %v\n", err)
			os.Exit(1)
		}
		// Now write ddidx file
		for _, index := range indexes {
			_, err := index.WriteTo(ddidxOutput)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error writing to ddidx: %v\n", err)
				os.Exit(1)
			}
		}
		return
	},
}

func init() {
	rootCmd.AddCommand(fastzCmd)

	// Defining flags for primerSet and templateMap files, and the output location for the ddidx file
	fastzCmd.Flags().String("primerSet", "", "Path to the primer set file")
	fastzCmd.Flags().String("templateMap", "", "Path to the template map file")
	fastzCmd.Flags().String("ddidxOutput", "", "Output location for the .ddidx index file")
	fastzCmd.Flags().Uint("kmerSize", 16, "K-mer size for megamash")
	fastzCmd.Flags().Uint("threshold", 10, "Threshold for megamash")
	fastzCmd.Flags().Float64("score", 0.8, "Score for filtering")
	defaultCPUs := runtime.NumCPU()
	fastzCmd.Flags().Int("cpus", defaultCPUs, "Number of CPUs to use")

	// Marking the flags as required
	fastzCmd.MarkFlagRequired("primerSet")
	fastzCmd.MarkFlagRequired("templateMap")
	fastzCmd.MarkFlagRequired("ddidxOutput")
}
