/*
Package uniref provides a parser for UniRef XML files.

UniRef clusters uniprot proteins by similarity. This is useful for doing
bioinformatics on protein space, as many proteins are sequenced a ton of times
in different organisms, and you don't want those proteins to dominate your
training data.

UniRef data dumps are available as FASTA or XML formatted data. The XML has
more rich data, so we use that. The parser was created using AI.

UniProt Reference Clusters (UniRef) provide clustered sets of sequences from
the UniProt Knowledgebase (including isoforms) and selected UniParc records in
order to obtain complete coverage of the sequence space at several resolutions
while hiding redundant sequences (but not their descriptions) from view.
(taken from uniref reference https://www.uniprot.org/help/uniref)

Download uniref data dumps here: https://www.uniprot.org/downloads

UniRef comes in three formats:
- UniRef100: Clusters of sequences that have 100% sequence identity and same length
- UniRef90: Clusters of sequences with at least 90% sequence identity and 80% overlap
- UniRef50: Clusters of sequences with at least 50% sequence identity and 80% overlap
*/
package uniref

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// Header is an empty struct since UniRef files don't have headers
type Header struct{}

// Entry represents a UniRef entry
type Entry struct {
	XMLName    xml.Name             `xml:"entry"`
	ID         string               `xml:"id,attr"`
	Updated    string               `xml:"updated,attr"`
	Name       string               `xml:"name"`
	Properties []Property           `xml:"property"`
	RepMember  RepresentativeMember `xml:"representativeMember"`
	Members    []Member             `xml:"member"`
}

// Property represents a property element
type Property struct {
	Type  string `xml:"type,attr"`
	Value string `xml:"value,attr"`
}

// DBReference represents a database reference
type DBReference struct {
	Type       string     `xml:"type,attr"`
	ID         string     `xml:"id,attr"`
	Properties []Property `xml:"property"`
}

// Sequence represents a sequence element
type Sequence struct {
	Length   int    `xml:"length,attr"`
	Checksum string `xml:"checksum,attr"`
	Value    string `xml:",chardata"`
}

// Member represents a member element
type Member struct {
	DBRef    DBReference `xml:"dbReference"`
	Sequence *Sequence   `xml:"sequence"`
}

// RepresentativeMember represents the representative member
type RepresentativeMember Member

// UniRef represents the root element which can be UniRef50, UniRef90, or UniRef100
type UniRef struct {
	XMLName     xml.Name // This will automatically match the root element name
	ReleaseDate string   `xml:"releaseDate,attr"`
	Version     string   `xml:"version,attr"`
	Entries     []Entry  `xml:"entry"`
}

// GetUniRefVersion returns "50", "90", or "100" based on the XML root element name
func (u *UniRef) GetUniRefVersion() string {
	name := u.XMLName.Local
	if strings.HasPrefix(name, "UniRef") {
		return strings.TrimPrefix(name, "UniRef")
	}
	return ""
}

type Parser struct {
	decoder *xml.Decoder
}

func NewParser(r io.Reader) (*Parser, error) {
	decoder := xml.NewDecoder(r)
	decoder.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
		if strings.ToLower(charset) == "iso-8859-1" {
			return input, nil
		}
		return nil, fmt.Errorf("unsupported charset: %s", charset)
	}

	return &Parser{
		decoder: decoder,
	}, nil
}

// Header returns an empty header since UniRef files don't have headers
func (p *Parser) Header() (Header, error) {
	return Header{}, nil
}

// Next returns the next Entry from the UniRef file
func (p *Parser) Next() (Entry, error) {
	for {
		token, err := p.decoder.Token()
		if err == io.EOF {
			return Entry{}, io.EOF
		}
		if err != nil {
			return Entry{}, err
		}

		// Look for start element of an entry
		if startElement, ok := token.(xml.StartElement); ok {
			if startElement.Name.Local == "entry" {
				var entry Entry
				if err := p.decoder.DecodeElement(&entry, &startElement); err != nil {
					return Entry{}, err
				}
				return entry, nil
			}
		}
	}
}

// ToXML converts an Entry back to its XML representation
func (e *Entry) ToXML() (string, error) {
	buf := new(bytes.Buffer)
	enc := xml.NewEncoder(buf)
	enc.Indent("", "  ")
	if err := enc.Encode(e); err != nil {
		return "", err
	}
	return buf.String(), nil
}
