package main

import (
	"compress/gzip"
	"context"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"strings"

	"github.com/koeng101/dnadesign/lib/bio"
	"github.com/koeng101/dnadesign/lib/tokenizer"
	"golang.org/x/sync/errgroup"
)

func main() {
	// Define flags
	shardSize := flag.Int("shardSize", int(math.Pow(10, 7)), "Size of each shard") // uniprot sprot splits into 40 files, so 2.5% is retained for validation
	outputDir := flag.String("outputDir", "", "Output directory path")
	unirefInput := flag.String("unirefInput", "", "Uniref input directory")

	// Parse the command line flags
	flag.Parse()

	// Check if the directory path is provided
	if *outputDir == "" {
		fmt.Println("outputDir must be specified")
		os.Exit(1)
	}

	var uniref io.Reader
	// Open and decompress uniref file
	unirefFile, err := os.Open(*unirefInput)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer unirefFile.Close()

	uniref, err = gzip.NewReader(unirefFile)
	if err != nil {
		fmt.Println("Error creating gzip reader:", err)
		return
	}

	// Get a default tokenizer
	tokenizer := tokenizer.DefaultAminoAcidTokenizer()
	inputChannel := make(chan []uint16)
	ctx := context.Background()
	errorGroup, ctx := errgroup.WithContext(ctx)
	errorGroup.Go(func() error {
		return tokenizer.WriteTokensToShards(ctx, inputChannel, *shardSize, *outputDir)
	})
	fmt.Println("initializing parser")
	tokenizerJSON, err := tokenizer.ToJSON()
	if err != nil {
		fmt.Println("Err: ", err)
	}
	fmt.Println(tokenizerJSON)
	refParser := bio.NewFastaParser(uniref)
	count := 0
	for {
		if (count % 10000) == 0 {
			fmt.Printf("Processed sequence: %d\n", count)
		}
		protein, err := refParser.Next()
		if err != nil {
			break
		}
		sequence := strings.ToUpper(protein.Sequence)
		tokens, _ := tokenizer.TokenizeProtein(sequence)
		inputChannel <- tokens
		count++
	}
	close(inputChannel)
}
