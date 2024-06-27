package main

import (
	"compress/gzip"
	"context"
	"crypto/md5"
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
	shardSize := flag.Int("shardSize", int(math.Pow(10, 8)), "Size of each shard") // uniprot sprot splits into 40 files, so 2.5% is retained for validation
	outputDir := flag.String("outputDir", "", "Output directory path")
	tremblInput := flag.String("tremblInput", "", "Trembl input directory")
	unirefInput := flag.String("unirefInput", "", "Uniref input directory")
	refFileFlag := flag.Bool("refFile", true, "use uniref file")
	refFile := *refFileFlag

	// Parse the command line flags
	flag.Parse()

	// Check if the directory path is provided
	if *outputDir == "" {
		fmt.Println("outputDir must be specified")
		os.Exit(1)
	}

	// Open and decompress trembl file
	tremblFile, err := os.Open(*tremblInput)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer tremblFile.Close()

	trembl, err := gzip.NewReader(tremblFile)
	if err != nil {
		fmt.Println("Error creating gzip reader:", err)
		return
	}

	var uniref io.Reader
	if refFile {
		// Open and decompress uniref file
		unirefFile, err := os.Open(*unirefInput)
		if err != nil {
			fmt.Println("Error opening file:", err)
			return
		}
		defer unirefFile.Close()

		uniref, err := gzip.NewReader(unirefFile)
		if err != nil {
			fmt.Println("Error creating gzip reader:", err)
			return
		}
		defer uniref.Close()
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
	parser := bio.NewUniprotParser(trembl)
	count := 0
	pfamMap := make(map[string][]string) // hash -> pfam
	for {
		if (count % 100000) == 0 {
			fmt.Printf("Processed pfam: %d\n", count)
		}
		entry, err := parser.Next()
		if err != nil {
			break
		}
		// Read uniprot trembl.
		var id string
		for _, reference := range entry.DbReference {
			if reference.Type == "Pfam" {
				id = reference.Id
				sequence := strings.ToUpper(entry.Sequence.Value)
				if sequence[len(sequence)-1] == '*' {
					sequence = sequence[:len(sequence)-1]
				}
				checkSum := fmt.Sprintf("%x", md5.Sum([]byte(sequence)))
				_, ok := pfamMap[checkSum]
				if !ok {
					pfamMap[checkSum] = []string{id}
				} else {
					found := false
					for _, pfam := range pfamMap[checkSum] {
						if pfam == id {
							found = true
						}
					}
					if !found {
						pfamMap[checkSum] = append(pfamMap[checkSum], id)
					}
				}
			}
		}
		count++
	}
	trembl.Close()

	// Write pfams to tokenizer
	var pfamCount uint16
	tokenizer.TokenMap.Range(func(_, _ interface{}) bool {
		pfamCount++
		return true
	})
	for _, values := range pfamMap {
		for _, pfam := range values {
			pfamCount++
			tokenizer.TokenMap.Store(pfam, pfamCount)
		}
	}
	tokenizerJSON, err := tokenizer.ToJSON()
	if err != nil {
		fmt.Println("Err: ", err)
	}
	fmt.Println(tokenizerJSON)

	if refFile {
		refParser := bio.NewFastaParser(uniref)
		count = 0
		for {
			if (count % 10000) == 0 {
				fmt.Printf("Processed sequence: %d\n", count)
			}
			protein, err := refParser.Next()
			if err != nil {
				break
			}
			sequence := strings.ToUpper(protein.Sequence)
			if sequence[len(sequence)-1] == '*' {
				sequence = sequence[:len(sequence)-1]
			}
			checkSum := fmt.Sprintf("%x", md5.Sum([]byte(sequence)))
			// Now that the pfam is in the token map, get it.
			pfams, ok := pfamMap[checkSum]
			if !ok {
				fmt.Println("Skipping: ", protein)
				continue
			}
			for _, pfam := range pfams {
				pfamTokenUntyped, _ := tokenizer.TokenMap.Load(pfam)
				pfamToken, _ := pfamTokenUntyped.(uint16)
				tokens, _ := tokenizer.TokenizeProtein(sequence)

				// Append tokens together
				allTokens := make([]uint16, 0, 1+len(tokens))
				allTokens = append(allTokens, pfamToken)
				allTokens = append(allTokens, tokens...)
				inputChannel <- allTokens
			}
			count++
		}
	} else {
		// Open and decompress trembl file
		tremblFile, err := os.Open(*tremblInput)
		if err != nil {
			fmt.Println("Error opening file:", err)
			return
		}
		defer tremblFile.Close()

		trembl, err := gzip.NewReader(tremblFile)
		if err != nil {
			fmt.Println("Error creating gzip reader:", err)
			return
		}
		count = 0
		parser := bio.NewUniprotParser(trembl)
		for {
			if (count % 100000) == 0 {
				fmt.Printf("Processed pfam: %d\n", count)
			}
			entry, err := parser.Next()
			if err != nil {
				break
			}
			// Read uniprot trembl.
			var pfam string
			for _, reference := range entry.DbReference {
				if reference.Type == "Pfam" {
					pfam = reference.Id
					sequence := strings.ToUpper(entry.Sequence.Value)
					if sequence[len(sequence)-1] == '*' {
						sequence = sequence[:len(sequence)-1]
					}
					pfamTokenUntyped, _ := tokenizer.TokenMap.Load(pfam)
					pfamToken, _ := pfamTokenUntyped.(uint16)
					tokens, _ := tokenizer.TokenizeProtein(sequence)

					// Append tokens together
					allTokens := make([]uint16, 0, 1+len(tokens))
					allTokens = append(allTokens, pfamToken)
					allTokens = append(allTokens, tokens...)
					inputChannel <- allTokens
				}
			}
			count++
		}
	}
	close(inputChannel)
}
