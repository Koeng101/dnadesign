/*
Package ddidx contains information about the dnadesign index format.
*/
package ddidx

import (
	"encoding/binary"
	"errors"
	"io"
)

// Index is a 32 byte index for individual objects.
type Index struct {
	Identifier    [16]byte
	StartPosition uint64
	Length        uint64
}

// WriteTo writes the binary representation of the Index to the given writer.
// It returns the number of bytes written and any error encountered.
func (i *Index) WriteTo(w io.Writer) (int64, error) {
	// The total bytes written
	var totalBytes int64

	// Write Identifier
	n, err := w.Write(i.Identifier[:])
	totalBytes += int64(n)
	if err != nil {
		return totalBytes, err
	}

	// Create a buffer to write the uint64 values
	buf := make([]byte, 8)

	// Write StartPosition
	binary.BigEndian.PutUint64(buf, i.StartPosition)
	n, err = w.Write(buf)
	totalBytes += int64(n)
	if err != nil {
		return totalBytes, err
	}

	// Write Length
	binary.BigEndian.PutUint64(buf, i.Length)
	n, err = w.Write(buf)
	totalBytes += int64(n)
	if err != nil {
		return totalBytes, err
	}

	return totalBytes, nil
}

// ReadIndexes reads and returns a list of Index structs from the given reader.
func ReadIndexes(r io.Reader) ([]Index, error) {
	var indexes []Index

	for {
		var idx Index

		// Read Identifier
		if _, err := io.ReadFull(r, idx.Identifier[:]); err != nil {
			if errors.Is(err, io.EOF) {
				break // End of file, stop reading
			}
			return indexes, err
		}

		// Read StartPosition
		if err := binary.Read(r, binary.BigEndian, &idx.StartPosition); err != nil {
			return indexes, err
		}

		// Read Length
		if err := binary.Read(r, binary.BigEndian, &idx.Length); err != nil {
			return indexes, err
		}

		indexes = append(indexes, idx)
	}

	return indexes, nil
}
