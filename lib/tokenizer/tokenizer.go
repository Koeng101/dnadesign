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
self-attention plus feed-forward neural networks, laid on top of each other.
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
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

func TokenizeProtein(sequence string) ([]uint8, error) {
	// Switch statements are faster than maps
	// https://adayinthelifeof.nl/2021/03/04/go-map-vs-switch.html
	// https://www.reddit.com/r/golang/comments/lxju7f/benchmarking_maps_vs_switches/
	tokens := make([]uint8, len(sequence)+1) // +1 for end token, which is the default 0
	var token uint8

	// Tokens: end_token, "ACDEFGHIKLMNPQRSTVWYUO*BXZ"
	// {"A":1,"C":2,"D":3,"E":4,"F":5,"G":6,"H":7,"I":8,"K":9,"L":10,"M":11,"N":12,"P":13,"Q":14,"R":15,"S":16,"T":17,"V":18,"W":19,"Y":20,"U":21,"O":22,"*":23,"B":24,"X":25,"Z":26}
	for i, aminoAcid := range sequence {
		switch aminoAcid {
		case 'A':
			token = 1
		case 'C':
			token = 2
		case 'D':
			token = 3
		case 'E':
			token = 4
		case 'F':
			token = 5
		case 'G':
			token = 6
		case 'H':
			token = 7
		case 'I':
			token = 8
		case 'K':
			token = 9
		case 'L':
			token = 10
		case 'M':
			token = 11
		case 'N':
			token = 12
		case 'P':
			token = 13
		case 'Q':
			token = 14
		case 'R':
			token = 15
		case 'S':
			token = 16
		case 'T':
			token = 17
		case 'V':
			token = 18
		case 'W':
			token = 19
		case 'Y':
			token = 20
		case 'U': // Selenocysteine
			token = 21
		case 'O': // Pyrrolysine
			token = 22
		case '*': // Stop codon
			token = 23
		case 'B': // Aspartic acid or Asparagine
			token = 24
		case 'X': // Any amino acid
			token = 25
		case 'Z': // Glutamic acid or Glutamine
			token = 26
		default:
			return tokens, fmt.Errorf("Got unknown amino acid. Must be in list of ACDEFGHIKLMNPQRSTVWYUO*BXZ. Got: %c", aminoAcid)
		}
		tokens[i] = token
	}
	return tokens, nil
}

// Tokenizer is a struct defining a tokenizer. Start and End tokens are
// specially encoded, while normal tokens reside in TokenMap.
type Tokenizer struct {
	TokenMap     sync.Map // concurrent safe
	EndToken     uint16
	EndTokenText string
}

// ToJSON converts the Tokenizer struct to JSON.
func (t *Tokenizer) ToJSON() (string, error) {
	// Convert sync.Map to a regular map
	tokenMap := make(map[string]uint16)
	t.TokenMap.Range(func(key, value interface{}) bool {
		tokenMap[key.(string)] = value.(uint16)
		return true
	})

	// Create a temporary struct for JSON marshalling
	temp := struct {
		TokenMap     map[string]uint16 `json:"token_map"`
		EndToken     uint16            `json:"end_token"`
		EndTokenText string            `json:"end_token_text"`
	}{
		TokenMap:     tokenMap,
		EndToken:     t.EndToken,
		EndTokenText: t.EndTokenText,
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(temp)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

// DefaultAminoAcidTokenizer returns a default Tokenizer that can encode amino
// acid data as tokens. It is a function rather than just directly encoded so
// modifications can be made to it as an application runs.
func DefaultAminoAcidTokenizer() *Tokenizer {
	var tokenizer = Tokenizer{
		TokenMap:     *new(sync.Map),
		EndToken:     0,
		EndTokenText: "<|endoftext|>",
	}
	chars := "ACDEFGHIKLMNPQRSTVWYUO*BXZ"
	tokenValue := uint16(1)
	for _, char := range chars {
		tokenizer.TokenMap.Store(string(char), tokenValue)
		tokenValue++
	}
	return &tokenizer
}

// TokenizeProteins converts a protein sequence into a list of tokens.
func (t *Tokenizer) TokenizeProtein(proteinSequence string) ([]uint16, error) {
	// We know how long the protein should be, so we can pre-allocate space
	tokens := make([]uint16, 0, 1+len(proteinSequence)) // add end to len
	for _, aminoAcid := range proteinSequence {
		tokenInteger, ok := t.TokenMap.Load(string(aminoAcid))
		if !ok {
			return tokens, errors.New("Only letters ACDEFGHIKLMNPQRSTVWYUO*BXZ are allowed for Proteins. Got letter: " + string(aminoAcid))
		}
		tokenIntegerTyped, ok := tokenInteger.(uint16)
		if ok {
			tokens = append(tokens, tokenIntegerTyped)
		} else {
			return tokens, errors.New("Failed to uint16 type. HINT: Are you adding custom tokens?")
		}
	}
	tokens = append(tokens, t.EndToken)
	return tokens, nil
}

// WriteTokensToShards is a function that takes in a tokenChannel and writes to
// shards. The idea is that, normally, you will be reading a very large
// quantity of data, so you want to have a concurrent process writing those
// shards to disk. Unlike many functions which use `io.Writer`, these shards
// are intended to be larger than a single file can hold, and thus they are
// written to a directory. The first shard is retained as a validation set,
// and the remaining shards are written as training sets.
//
// ShardSize is the number of tokens per file. ContextLength is the context
// length of the model. OutputDir is where the training / validation shards get
// written to.
func (t *Tokenizer) WriteTokensToShards(ctx context.Context, tokenChannel <-chan []uint16, shardSize int, outputDir string) error {
	var err error
	tokenCount := 0
	shardCount := 0
	currentShard := make([]uint16, 0, shardSize*2) // shardSize*2 is preallocated
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case tokens, ok := <-tokenChannel:
			if !ok {
				// Write any remaining tokens to a final shard
				if len(currentShard) > 0 {
					return writeShardToFile(currentShard, shardCount, outputDir)
				}
			}
			// Write data
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
	}
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
