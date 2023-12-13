package uniprotapi

import (
	"context"
	"os"
	"testing"
)

func TestUniprotToDatabase(t *testing.T) {
	db := MakeTestDatabase(":memory:", Schema)
	uniprotFile, err := os.Open("data/uniprot_sprot_mini.xml.gz")
	if err != nil {
		t.Error(err)
	}
	defer uniprotFile.Close()

	ctx := context.Background()
	i, err := UniprotToDatabase(db, ctx, 10, uniprotFile)
	if err != nil {
		t.Errorf("Failed to UniprotToDatabase. Got err: %s", err)
	}
	if i != 20 {
		t.Errorf("Expected 20 entries. Got: %d", i)
	}
}
