package uniprot_test

import (
	"compress/gzip"
	_ "embed"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/koeng101/dnadesign/lib/bio/uniprot"
)

func TestRead(t *testing.T) {
	uniprotFile, err := os.Open("data/uniprot_sprot_mini.xml.gz")
	if err != nil {
		t.Errorf("Should open file properly")
	}
	defer uniprotFile.Close()
	unzippedFile, _ := gzip.NewReader(uniprotFile)
	parser := uniprot.NewParser(unzippedFile)
	for {
		_, err := parser.Next()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				t.Errorf("Failed to parse uniprot test file with err: %s", err)
				break
			}
		}
	}
}

func TestHeader(t *testing.T) {
	var writer = io.Discard
	header, err := uniprot.NewParser(nil).Header()
	if err != nil {
		t.Errorf("should always be nil")
	}
	_, err = header.WriteTo(writer)
	if err != nil {
		t.Errorf("should always be nil")
	}
}

func TestEntry(t *testing.T) {
	var writer = io.Discard
	entry := uniprot.Entry{}
	_, err := entry.WriteTo(writer)
	if err != nil {
		t.Errorf("should always be nil")
	}
}

//go:embed data/P42212.xml
var gfpXml []byte

func TestGet(t *testing.T) {
	// First, a successful get
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/P42212.xml":
			w.WriteHeader(http.StatusOK)
			w.Write(gfpXml)
		case "/asdf.xml":
			w.WriteHeader(404)
		}
	}))
	defer server.Close()

	uniprot.BaseURL = server.URL
	entry, err := uniprot.Get("P42212")
	if err != nil {
		t.Error(err)
	}
	if entry.Accession[0] != "P42212" {
		t.Errorf("Expected 'P42212', got %s", entry.Accession[0])
	}

	// Next, a bad base url
	uniprot.BaseURL = "//example.com#fragment"
	_, err = uniprot.Get("P42212")
	if err == nil {
		t.Errorf("Should fail with bad BaseURL")
	}

	// Next, with a bad accession number
	uniprot.BaseURL = server.URL
	_, err = uniprot.Get("asdf")
	if err == nil {
		t.Errorf("Should fail with bad accession number")
	}
}
