package tokenizer

import (
	"compress/gzip"
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/koeng101/dnadesign/lib/bio"
	"golang.org/x/sync/errgroup"
)

func TestTokenizeProtein(t *testing.T) {
	proteinSequence := "ACDEFGHIKLMNPQRSTVWYUO*BXZ"
	tokenizer := DefaultAminoAcidTokenizer()
	tokens, err := tokenizer.TokenizeProtein(proteinSequence)
	if err != nil {
		t.Errorf("Should have successfully tokenized. Got error: %s", err)
	}
	for i, token := range tokens[1 : len(tokens)-1] {
		// The first amino acid token is 3
		if token != uint16(i+2) {
			t.Errorf("Expected %d, got: %d", i+2, token)
		}
	}
	badProtein := "J" // should fail
	_, err = tokenizer.TokenizeProtein(badProtein)
	if err == nil {
		t.Errorf("Should have failed on J")
	}
}

func TestWriteTokensToShards(t *testing.T) {
	// temporary directory
	tempDir, err := os.MkdirTemp("", "example")
	if err != nil {
		fmt.Println("Error creating a temporary directory:", err)
		return
	}
	defer os.RemoveAll(tempDir) // Clean up

	// Get a default tokenizer
	tokenizer := DefaultAminoAcidTokenizer()
	inputChannel := make(chan []uint16)
	shardSize := 2000
	contextLength := 1024
	ctx := context.Background()
	errorGroup, ctx := errgroup.WithContext(ctx)
	errorGroup.Go(func() error {
		return tokenizer.WriteTokensToShards(ctx, inputChannel, shardSize, contextLength, tempDir)
	})
	uniprotFile, _ := os.Open("data/gfp_rfp_lacZ.xml.gz")
	file, _ := gzip.NewReader(uniprotFile)
	parser := bio.NewUniprotParser(file)
	for {
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
	}
	close(inputChannel)

	// Now, we read the files we created:
	// Open the directory
	dir, err := os.Open(tempDir)
	if err != nil {
		t.Error("Error opening directory: ", err)
	}

	// Read the directory contents
	files, err := dir.Readdirnames(0) // 0 to read all files and directories
	if err != nil {
		t.Error("Error reading directory contents: ", err)
	}

	// Iterate over the files and print them
	count := 0
	for range files {
		count++
		// fmt.Println(file) // uncomment this to read the two files generated
	}
	if count != 2 {
		for _, file := range files {
			fmt.Println(file)
		}
		t.Error("Expected 2 generated files. Got: ", count)
	}
	dir.Close()
}
