/*
Package tokenzier contains tokenizers for biological data.

Large Language Models (LLMs) are increasingly taking over the machine learning
field. There are two fundamental innovations: the idea of token vectors and
self-attention.

Rather than encoding words (or perhaps, amino acids) as themselves in a machine
learning model, they are encoded as token vectors. Tokens can be full words,
but are usually fragments of words. In the case of amino acids, each amino acid
would be a "token". For example:

	Token	->	Amino Acid
	1		->	A
	2		->	C
	3		->	D
	...
	20		-> Y
	21		-> *

These tokens are usually just integers, corresponding with a map to the actual
words they represent. These tokens are then mapped to a vector embedding:

	1 -> [0.0, 0.2, 0.1, ... ] (length:512)
	2 -> [0.9, 0.0, 0.2, ... ] (length:512)
	3 -> [0.2, 0.4, 0.6, ... ] (length:512)

In the original instantiation of vector embeddings, one could think of them as
representing an idea in high-dimensional space. For example, the concept of
gender could be the difference between the vector between "mom" and "dad"
(which correspondingly would also be the difference between the vector between
"aunt" and "uncle").

The idea is that these vector embeddings can be compared to each other to find
the most relevant portions of a sequence for a model, otherwise known as
"attention". When the model is comparing to itself, this is called
"self-attention". A good example of self attention is looking at the words in a
sentence to find out the meaning, or the way each amino acid in a protein
interacts with each other amino acid.

Transformers are a specific deep learning model architecture that depends on
self-attention plus feed-forward neural networks, layed on top of each other.
Because of the multiple layers of self-attention, transformers are very good
at figuring out the context of information, and how it relates to other
information in a sequence. These have found their way into biotechnology
research.

Alphafold is a great example of transformer-architecture applied to biological
data: by utilizing the self-attention mechanisms of transformers, it is able
to more effectively predict protein structure than any other piece of software.

This package's intention is to make a tokenizer for amino acid data, such that
sources like uniprot can be used to train LLMs. Essentially, we want to convert
amino acid sequence data to a list of int32 integers in an easy-to-use way.

We will be using Karpathy's datafile format from llm.c, written here:

	https://github.com/karpathy/llm.c/blob/master/dev/data/data_common.py

In brief, there is a header with 256 int32, followed by tokens as uint16. The
header begins with the magic number 20240520, then a version number, then the
number of tokens encoded after the header.
*/
package tokenizer

import (
	"bufio"
	"compress/gzip"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/koeng101/dnadesign/lib/bio"
)

// init initializes default tokenizers. This is run when importing the package
// to generate the desired lists.
func init() {
	// Init DefaultAminoAcidTokenizer
	chars := "ACDEFGHIKLMNPQRSTVWYUO*BXZ"
	tokenValue := uint16(1)
	for _, char := range chars {
		DefaultAminoAcidTokenizer.TokenMap[string(char)] = tokenValue
		tokenValue++
	}
}

// Tokenizer is a struct defining a tokenizer. Start and End tokens are
// specially encoded, while normal tokens reside in TokenMap.
type Tokenizer struct {
	TokenMap       map[string]uint16
	StartToken     uint16
	StartTokenText string
	EndToken       uint16
	EndTokenText   string
}

// DefaultAminoAcidTokenizer is a default Tokenizer that can encode amino acid
// data as tokens.
var DefaultAminoAcidTokenizer = Tokenizer{
	TokenMap:     map[string]uint16{}, // initialized with init()
	EndToken:     0,
	EndTokenText: "<|endoftext|>",
}

// TokenizeProteins converts a protein sequence into a list of tokens.
func TokenizeProtein(proteinSequence string) ([]uint16, error) {
	// We know how long the protein should be, so we can pre-allocate space
	tokens := make([]uint16, 0, 2+len(proteinSequence)) // add start+end to len
	for _, aminoAcid := range proteinSequence {
		tokenInteger, ok := DefaultAminoAcidTokenizer.TokenMap[string(aminoAcid)]
		if !ok {
			return tokens, errors.New("Only letters ACDEFGHIKLMNPQRSTVWYUO*BXZ are allowed for Proteins. Got letter: " + string(aminoAcid))
		}
		tokens = append(tokens, tokenInteger)
	}
	tokens = append(tokens, DefaultAminoAcidTokenizer.EndToken)
	return tokens, nil
}

// https://ftp.uniprot.org/pub/databases/uniprot/uniref/uniref90/uniref90.fasta.gz
func TokenizeFastaFile(r io.Reader, shardSize int, contextLength int, outputDir string) error {
	// Create a gzip reader
	gzReader, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzReader.Close()

	// Create a buffered reader
	reader := bufio.NewReader(gzReader)

	// Initialize shard variables
	currentShard := make([]uint16, 0, shardSize+contextLength+1) // shardSize + max protein length + end token
	tokenCount := 0
	shardCount := 0

	// Parse the fasta file
	parser := bio.NewFastaParser(reader)
	for {
		record, err := parser.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		tokens, err := TokenizeProtein(record.Sequence)
		if err != nil {
			return err
		}
		currentShard = append(currentShard, tokens...)
		tokenCount += len(tokens)

		// If the current shard is full, write it to a file
		if tokenCount >= shardSize {
			err = writeShardToFile(currentShard[:tokenCount], shardCount, outputDir)
			if err != nil {
				return err
			}
			currentShard = currentShard[:0] // slice is cleared, but the memory is still allocated.
			tokenCount = 0
			shardCount++
		}
	}
	// Write any remaining tokens to a final shard
	if len(currentShard) > 0 {
		err = writeShardToFile(currentShard, shardCount, outputDir)
		if err != nil {
			return err
		}
	}
	return nil
}

// writeShardToFile is a helper function that wries a shard to a file.
func writeShardToFile(shard []uint16, shardIndex int, outputDir string) error {
	var shardType string
	if shardIndex == 0 { // the first shard is reserved for val, the rest is train
		shardType = "val"
	} else {
		shardType = "train"
	}
	// Create the output file
	outputFileName := filepath.Join(outputDir, fmt.Sprintf("shard_%s_%d.bin", shardType, shardIndex))
	outputFile, err := os.Create(outputFileName)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	// Create a buffered writer. This will help the file get written because the
	// filesystem won't be called on every write.
	bufferedWriter := bufio.NewWriter(outputFile)
	defer bufferedWriter.Flush()

	// We write the header here, as defined in Karpathy's llm.c
	header := make([]int32, 256)  // Create a slice for 256 int32
	header[0] = 20240520          // Set magic number
	header[1] = 1                 // Set version info
	header[2] = int32(len(shard)) // Set the third int with the length of the shard

	// Convert the header to bytes and write it.
	for _, value := range header {
		err := binary.Write(bufferedWriter, binary.LittleEndian, value)
		if err != nil {
			return err
		}
	}

	// Finally, write data.
	for _, token := range shard {
		_, err := bufferedWriter.Write([]byte{byte(token), byte(token >> 8)})
		if err != nil {
			return err
		}
	}
	return nil
}
