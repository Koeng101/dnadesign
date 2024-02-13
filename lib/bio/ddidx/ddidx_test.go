package ddidx

import (
	"bytes"
	"reflect"
	"testing"
)

func TestIndexWriteToAndReadIndexes(t *testing.T) {
	// Prepare a slice of Index instances for testing
	indexes := []Index{
		{
			Identifier:    [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
			StartPosition: 100,
			Length:        200,
		},
		{
			Identifier:    [16]byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1},
			StartPosition: 300,
			Length:        400,
		},
	}

	// Create a buffer to write the indexes to
	var buf bytes.Buffer

	// Write each index to the buffer
	for _, idx := range indexes {
		if _, err := idx.WriteTo(&buf); err != nil {
			t.Fatalf("WriteTo failed: %v", err)
		}
	}

	// Now read the indexes back from the buffer
	readIndexes, err := ReadIndexes(&buf)
	if err != nil {
		t.Fatalf("ReadIndexes failed: %v", err)
	}

	// Compare the original indexes with the ones read back
	if !reflect.DeepEqual(indexes, readIndexes) {
		t.Errorf("Original indexes %+v do not match read indexes %+v", indexes, readIndexes)
	}
}
