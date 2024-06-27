package main

import (
	"context"
	"flag"
	"fmt"
	"math"
	"os"

	"github.com/koeng101/dnadesign/lib/bio"
	"github.com/koeng101/dnadesign/lib/tokenizer"
	"golang.org/x/sync/errgroup"
)

func main() {
	// Define flags
	shardSize := flag.Int("shardSize", int(math.Pow(10, 7)), "Size of each shard") // uniprot sprot splits into 40 files, so 2.5% is retained for validation
	outputDir := flag.String("outputDir", "", "Output directory path")

	// Parse the command line flags
	flag.Parse()

	// Check if the directory path is provided
	if *outputDir == "" {
		fmt.Println("outputDir must be specified")
		os.Exit(1)
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
	parser := bio.NewUniprotParser(os.Stdin)
	count := 0
	for {
		if (count % 10000) == 0 {
			fmt.Println("Processed: ", count)
		}
		entry, err := parser.Next()
		if err != nil {
			break
		}
		// If the pfam is not in the tokenizer, add it
		var id string
		for _, reference := range entry.DbReference {
			if reference.Type == "Pfam" {
				id = reference.Id
				// First, check if the key already exists
				if _, ok := tokenizer.TokenMap.Load(id); !ok {
					// Key doesn't exist, count the entries.
					var count uint16
					tokenizer.TokenMap.Range(func(_, _ interface{}) bool {
						count++
						return true
					})
					// Add the new key with its value as the current count.
					tokenizer.TokenMap.Store(id, count)
				}
				// Now that the pfam is in the token map, get it.
				pfamTokenUntyped, _ := tokenizer.TokenMap.Load(id)
				pfamToken, _ := pfamTokenUntyped.(uint16)
				tokens, _ := tokenizer.TokenizeProtein(entry.Sequence.Value)

				// Append tokens together
				allTokens := make([]uint16, 0, 1+len(tokens))
				allTokens = append(allTokens, pfamToken)
				allTokens = append(allTokens, tokens...)
				inputChannel <- allTokens
			}
		}
		count++
	}
	tokenizerJSON, err := tokenizer.ToJSON()
	if err != nil {
		fmt.Println("Err: ", err)
	}
	fmt.Println(tokenizerJSON)
	close(inputChannel)
}
