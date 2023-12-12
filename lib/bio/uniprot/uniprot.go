/*
Package uniprot provides an XML parser for Uniprot data dumps.

Uniprot is comprehensive, high-quality and freely accessible resource of protein
sequence and functional information. It is the best(1) protein database out there.

Uniprot database dumps are available as gzipped FASTA files or gzipped XML files.
The XML files have significantly more information than the FASTA files, and this
parser specifically works on the gzipped XML files from Uniprot.

Uniprot provides an XML schema of their data dumps(3), which is useful for
autogeneration of Golang structs. xsdgen was used to automatically generate
xml.go from uniprot.xsd.

Each protein in Uniprot is known as an "Entry" (as defined in xml.go).

The function Parse stream-reads Uniprot into an Entry channel, from which you
can use the entries however you want. Read simplifies reading gzipped files
from a disk into an Entry channel.

(1) Opinion of Keoni Gandall as of May 18, 2021
(2) https://www.uniprot.org/downloads
(3) https://www.uniprot.org/docs/uniprot.xsd
*/
package uniprot

import (
	"compress/gzip"
	"encoding/json"
	"encoding/xml"
	"io"
)

// Decoder decodes XML elements2
type Decoder interface {
	DecodeElement(v interface{}, start *xml.StartElement) error
	Token() (xml.Token, error)
}

// Header is a blank struct, needed for compatibility with bio parsers. It contains nothing.
type Header struct{}

// Header_WriteTo is a blank function, needed for compatibility with bio parsers. It doesn't do anything.
func (header *Header) WriteTo(w io.Writer) (int64, error) {
	return 0, nil
}

// Header returns nil,nil.
func (p *Parser) Header() (*Header, error) {
	return &Header{}, nil
}

// Entry_WriteTo writes an entry to an io.Writer. It specifically writes a JSON
// representation, NOT an XML representation, of the uniprot data.
func (entry *Entry) WriteTo(w io.Writer) (int64, error) {
	b, err := json.Marshal(entry)
	if err != nil {
		return 0, err
	}
	n, err := w.Write(b)
	return int64(n), err
}

// Parser implements a bio parser with Next().
type Parser struct {
	decoder Decoder
}

// NewParser returns a Parser that uses r as the source
// from which to parse fasta formatted sequences. It expects a gzipped file,
// as the default uniprot dump is xml.gz
func NewParser(r io.Reader) (*Parser, error) {
	unzippedBytes, err := gzip.NewReader(r)
	if err != nil {
		return &Parser{}, err
	}
	decoder := xml.NewDecoder(unzippedBytes)
	return &Parser{decoder: decoder}, nil
}

func (p *Parser) Next() (*Entry, error) {
	decoderToken, err := p.decoder.Token()

	// Check decoding
	if err != nil {
		// If we are the end of the file, return io.EOF
		if err.Error() == "EOF" {
			return &Entry{}, io.EOF
		}
	}

	// Actual parsing
	startElement, ok := decoderToken.(xml.StartElement)
	if ok && startElement.Name.Local == "entry" {
		var e Entry
		err = p.decoder.DecodeElement(&e, &startElement)
		if err != nil {
			return &Entry{}, err
		}
		return &e, nil
	}
	return p.Next()
}

//// Read reads a gzipped Uniprot XML dump. Failing to open the XML dump
//// gives a single error, while errors encountered while decoding the XML dump
//// are added to the errors channel.
//func Read(path string) (chan Entry, chan error, error) {
//	entries := make(chan Entry, 100) // if you don't have a buffered channel, nothing will be read in loops on the channel.
//	decoderErrors := make(chan error, 100)
//	xmlFile, err := os.Open(path)
//	if err != nil {
//		return entries, decoderErrors, err
//	}
//	unzippedBytes, err := gzip.NewReader(xmlFile)
//	if err != nil {
//		return entries, decoderErrors, err
//	}
//	decoder := xml.NewDecoder(unzippedBytes)
//	go Parse(decoder, entries, decoderErrors)
//	return entries, decoderErrors, nil
//}
//
//// Parse parses Uniprot entries into a channel.
//func Parse(decoder Decoder, entries chan<- Entry, errors chan<- error) {
//	for {
//		decoderToken, err := decoder.Token()
//
//		if err != nil {
//			if err.Error() == "EOF" {
//				break
//			}
//			errors <- err
//		}
//		startElement, ok := decoderToken.(xml.StartElement)
//		if ok && startElement.Name.Local == "entry" {
//			var e Entry
//			err = decoder.DecodeElement(&e, &startElement)
//			if err != nil {
//				errors <- err
//			}
//			entries <- e
//		}
//	}
//	close(entries)
//	close(errors)
//}
