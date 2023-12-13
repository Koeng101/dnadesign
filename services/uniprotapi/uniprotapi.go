package uniprotapi

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"io"

	_ "modernc.org/sqlite"

	"github.com/koeng101/dnadesign/lib/bio/uniprot"
	"github.com/koeng101/dnadesign/lib/seqhash"
	"github.com/koeng101/dnadesign/services/uniprotapi/uniprotapisql"
)

//go:embed schema.sql
var Schema string

// CreateDatabase creates a new database.
func CreateDatabase(db *sql.DB, schema string) error {
	_, err := db.Exec(schema)
	return err
}

// MakeTestDatabase creates a database for testing purposes.
func MakeTestDatabase(dbLocation string, schema string) *sql.DB {
	db, err := sql.Open("sqlite", dbLocation)
	if err != nil {
		panic(err)
	}
	err = CreateDatabase(db, schema)
	if err != nil {
		panic(err)
	}
	return db
}

// UniprotToDatabase reads a gzipped uniprot file from an io.Reader and inserts
// the contents into an SQLite database. This function is meant to run
// separately from the API, constructing the database before the API uses it.
func UniprotToDatabase(db *sql.DB, ctx context.Context, insertsPerTransaction int, r io.Reader) (int, error) {
	var entryCount int

	// Create parser
	parser, err := uniprot.NewParser(r)
	if err != nil {
		return entryCount, err
	}

	// Start transaction
	var tx *sql.Tx
	tx, err = db.Begin()
	if err != nil {
		return entryCount, err
	}
	insertConnection := uniprotapisql.New(tx)

	// Parse each entry and insert into the database
	for {
		// Check for EOF
		entry, err := parser.Next()
		if err != nil {
			if err != io.EOF {
				return entryCount, err
			} else {
				break
			}
		}

		// Create seqhash ID
		sequence := entry.Sequence.Value
		sequenceSeqhash, err := seqhash.EncodeHashV2(seqhash.HashV2(sequence, seqhash.PROTEIN, false, false))
		if err != nil {
			return entryCount, err
		}

		// Marshal Entry into JSON
		entryJSON, err := json.Marshal(entry)
		if err != nil {
			return entryCount, err
		}

		// Insert into database
		_, err = insertConnection.InsertEntry(ctx, uniprotapisql.InsertEntryParams{ID: entry.Accession[0], Seqhash: sequenceSeqhash, Entry: string(entryJSON)})
		if err != nil {
			return entryCount, err
		}
		entryCount++

		if entryCount%insertsPerTransaction == 0 {
			err = tx.Commit()
			if err != nil {
				_ = tx.Rollback()
				return entryCount, err
			}
			tx, err = db.Begin()
			if err != nil {
				return entryCount, err
			}
			insertConnection = uniprotapisql.New(tx)
		}
	}
	err = tx.Commit()
	return entryCount, err
}
