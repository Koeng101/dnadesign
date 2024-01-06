package blowq

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strings"

	"github.com/google/uuid"
	"github.com/koeng101/dnadesign/lib/bio/fastq"
	"github.com/koeng101/dnadesign/lib/compressdna"
)

// WriteTo writes to a writer.
func WriteTo(w io.Writer, read *fastq.Read) (n int64, err error) {
	encodedRead, err := EncodeRead(*read)
	if err != nil {
		return 0, err
	}
	bytesWritten, err := w.Write(encodedRead)
	return int64(bytesWritten), err
}

var usesN bool

// EncodeRead encodes a single Read into a binary format.
func EncodeRead(read fastq.Read) ([]byte, error) {
	// Convert optionals to the simplified string format
	var optionalsBuilder strings.Builder
	for key, value := range read.Optionals {
		optionalsBuilder.WriteString(key)
		optionalsBuilder.WriteString("=")
		optionalsBuilder.WriteString(value)
		optionalsBuilder.WriteString(" ")
	}
	optionalsStr := optionalsBuilder.String()
	if len(optionalsStr) > 0 {
		// Remove the last space if not empty
		optionalsStr = optionalsStr[:len(optionalsStr)-1]
	}

	// Compress DNA and Quality
	compressedData, err := compressdna.CompressDNAWithQuality(read.Sequence, read.Quality, usesN)
	if err != nil {
		return nil, err
	}

	// Prepare buffer to write binary data
	buf := new(bytes.Buffer)

	// Convert UUID string to UUID bytes
	id, err := uuid.Parse(read.Identifier)
	if err != nil {
		return nil, err
	}
	buf.Write(id[:]) // Write 16 bytes of UUID

	// Write lengths as uint32 followed by actual data
	_ = binary.Write(buf, binary.LittleEndian, uint32(len(optionalsStr)))
	_, _ = buf.WriteString(optionalsStr)
	_ = binary.Write(buf, binary.LittleEndian, uint32(len(compressedData)))
	_, _ = buf.Write(compressedData)

	return buf.Bytes(), nil
}

// DecodeRead decodes binary data into a Read struct.
func DecodeRead(data []byte) (*fastq.Read, error) {
	buf := bytes.NewReader(data)

	// Read 16-byte UUID
	uuidBytes := make([]byte, 16)
	_, _ = buf.Read(uuidBytes)
	identifier, err := uuid.FromBytes(uuidBytes)
	if err != nil {
		return nil, err
	}

	// Read length of optionals string
	var optionalsLen uint32
	_ = binary.Read(buf, binary.LittleEndian, &optionalsLen)
	optionalsBytes := make([]byte, optionalsLen)
	_, _ = buf.Read(optionalsBytes)
	optionalsStr := string(optionalsBytes)

	// Decode optionals string into a map
	optionals := make(map[string]string)
	pairs := strings.Split(optionalsStr, " ")
	for _, pair := range pairs {
		kv := strings.Split(pair, "=")
		if len(kv) == 2 {
			optionals[kv[0]] = kv[1]
		} else {
			// Handle error or unexpected format
			return nil, fmt.Errorf("invalid format for key-value pair in optionals: %s", pair)
		}
	}

	// Read lengths
	var compressedLen uint32

	_ = binary.Read(buf, binary.LittleEndian, &compressedLen)
	compressedData := make([]byte, compressedLen)
	_, _ = buf.Read(compressedData)

	// Decompress DNA and Quality
	sequence, quality, err := compressdna.DecompressDNAWithQuality(compressedData)
	if err != nil {
		return nil, err
	}

	return &fastq.Read{
		Identifier: identifier.String(), // Convert UUID back to string
		Optionals:  optionals,
		Sequence:   sequence,
		Quality:    quality,
	}, nil
}

// Parser struct holds the state of the parser including the underlying reader.
type Parser struct {
	reader *bufio.Reader
}

// NewParser returns a Parser that uses r as the source
// from which to parse fastq formatted sequences.
func NewParser(r io.Reader, maxLineSize int) *Parser {
	return &Parser{
		reader: bufio.NewReaderSize(r, maxLineSize),
	}
}

// Header returns nil,nil.
func (parser *Parser) Header() (*fastq.Header, error) {
	return &fastq.Header{}, nil
}

// Next uses DecodeRead function to return the next fastq.Read from the reader, or an error.
func (parser *Parser) Next() (*fastq.Read, error) {
	// Assuming the structure of the fastq.Read is known and follows your defined format
	// Read 16-byte UUID
	uuidBytes := make([]byte, 16)
	if _, err := io.ReadFull(parser.reader, uuidBytes); err != nil {
		return nil, err // Handle error or EOF
	}

	// Read length of optionals string
	optionalsLenBytes := make([]byte, 4)
	if _, err := io.ReadFull(parser.reader, optionalsLenBytes); err != nil {
		return nil, err
	}
	optionalsLen := binary.LittleEndian.Uint32(optionalsLenBytes)

	// Read optionals string
	optionalsBytes := make([]byte, optionalsLen)
	if _, err := io.ReadFull(parser.reader, optionalsBytes); err != nil {
		return nil, err
	}

	// Read the length of the compressed data (if your structure has this)
	compressedLenBytes := make([]byte, 4)
	if _, err := io.ReadFull(parser.reader, compressedLenBytes); err != nil {
		return nil, err
	}
	compressedLen := binary.LittleEndian.Uint32(compressedLenBytes)

	// Read compressed data
	compressedData := make([]byte, compressedLen)
	if _, err := io.ReadFull(parser.reader, compressedData); err != nil {
		return nil, err
	}

	// Combine all parts into one byte slice as expected by DecodeRead
	var buf bytes.Buffer
	buf.Write(uuidBytes)
	buf.Write(optionalsLenBytes)
	buf.Write(optionalsBytes)
	buf.Write(compressedLenBytes)
	buf.Write(compressedData)

	// Call DecodeRead to decode the byte slice into fastq.Read
	read, err := DecodeRead(buf.Bytes())
	if err != nil {
		return nil, err
	}

	return read, nil
}
