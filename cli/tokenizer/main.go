package main

import (
	"bufio"
	"crypto/md5"
	"database/sql"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	_ "modernc.org/sqlite"

	"github.com/koeng101/dnadesign/lib/bio"
	"github.com/koeng101/dnadesign/lib/tokenizer"
)

// Function to convert []uint16 to a byte slice
func uint16SliceToBytes(slice []uint16) []byte {
	buf := make([]byte, len(slice)*2)
	for i, v := range slice {
		binary.LittleEndian.PutUint16(buf[i*2:], v)
	}
	return buf
}

// Function to convert byte slice back to []uint16
func bytesToUint16Slice(buf []byte) []uint16 {
	slice := make([]uint16, len(buf)/2)
	for i := range slice {
		slice[i] = binary.LittleEndian.Uint16(buf[i*2:])
	}
	return slice
}

func main() {
	// Parse the command line flags
	flag.Parse()

	// Connect to database
	db, err := sql.Open("sqlite", "./sequences.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create the table if it doesn't exist
	_, err = db.Exec(`
PRAGMA journal_mode = WAL;
PRAGMA synchronous = NORMAL; -- https://news.ycombinator.com/item?id=34247738
PRAGMA cache_size = 20000; -- https://news.ycombinator.com/item?id=34247738
PRAGMA foreign_keys = ON;
PRAGMA strict = ON;
PRAGMA busy_timeout = 5000;

        CREATE TABLE IF NOT EXISTS sequences (
            checksum TEXT PRIMARY KEY,
            sequence TEXT,
            tokens BLOB
        );
    `)
	if err != nil {
		log.Fatal(err)
	}

	// Get a default tokenizer
	tokenizer := tokenizer.DefaultAminoAcidTokenizer()
	fmt.Println("initializing parser")
	tokenizerJSON, err := tokenizer.ToJSON()
	if err != nil {
		fmt.Println("Err: ", err)
	}
	fmt.Println(tokenizerJSON)
	refParser := bio.NewFastaParser(bufio.NewReader(os.Stdin))
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
		tokensBytes := uint16SliceToBytes(tokens)
		checksum := fmt.Sprintf("%x", md5.Sum([]byte(sequence)))
		count++

		// Insert into the database
		_, err = db.Exec(`
            INSERT INTO sequences (checksum, sequence, tokens)
            VALUES (?, ?, ?);
        `, checksum, sequence, tokensBytes)
		if err != nil {
			log.Fatal(err)
		}
	}
}
