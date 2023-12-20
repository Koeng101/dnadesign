/*
Package uniprot provides an XML parser for Uniprot data dumps.

Uniprot is comprehensive, high-quality and freely accessible resource of protein
sequence and functional information. It is the best(1) protein database out there.

Uniprot database dumps are available as gzipped FASTA files or gzipped XML files.
The XML files have significantly more information than the FASTA files, and this
parser specifically works on the XML files from Uniprot.

Uniprot provides an XML schema of their data dumps(3), which is useful for
autogeneration of Golang structs. xsdgen was used to automatically generate
xml.go from uniprot.xsd.

Each protein in Uniprot is known as an "Entry" (as defined in xml.go).

(1) Opinion of Keoni Gandall as of May 18, 2021
(2) https://www.uniprot.org/downloads
(3) https://www.uniprot.org/docs/uniprot.xsd
*/
package uniprot

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
// from which to parse fasta formatted sequences.
func NewParser(r io.Reader) *Parser {
	decoder := xml.NewDecoder(r)
	return &Parser{decoder: decoder}
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

// BaseURL encodes the base URL for the Uniprot REST API.
var BaseURL string = "https://rest.uniprot.org/uniprotkb/"

// Get gets a uniprot from its accessionID
func Get(ctx context.Context, accessionID string) (*Entry, error) {
	var entry Entry

	// Parse the base URL
	baseURL, err := url.Parse(BaseURL)
	if err != nil {
		return &entry, err
	}

	// Resolve the full URL
	fullURL := baseURL.ResolveReference(&url.URL{Path: accessionID + ".xml"})

	// Create NewRequestWithContext. Note: since url.Parse catches errors in
	// the URL, no err is checked here.
	req, _ := http.NewRequestWithContext(ctx, "GET", fullURL.String(), nil)

	// Create a new HTTP client and send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return &entry, err
	}
	defer resp.Body.Close()

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		return &entry, fmt.Errorf("Got http status code: %d", resp.StatusCode)
	}

	// Return the first parsed XML
	return NewParser(resp.Body).Next()
}
