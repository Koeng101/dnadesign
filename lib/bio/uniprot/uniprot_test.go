package uniprot_test

import (
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/koeng101/dnadesign/lib/bio/uniprot"
)

func TestRead(t *testing.T) {
	testFile, err := os.Open("data/test")
	if err != nil {
		t.Errorf("Should open file properly")
	}
	defer testFile.Close()
	_, err = uniprot.NewParser(testFile)
	if err == nil {
		t.Errorf("Failed to fail on non-gzipped file")
	}

	_, err = os.Open("data/FAKE")
	if err == nil {
		t.Errorf("Failed to fail on empty file")
	}

	uniprotFile, err := os.Open("data/uniprot_sprot_mini.xml.gz")
	if err != nil {
		t.Errorf("Should open file properly")
	}
	defer uniprotFile.Close()
	parser, err := uniprot.NewParser(uniprotFile)
	if err != nil {
		t.Errorf("Parser should succeed. Got err: %s", err)
	}
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
	var writer io.Writer = ioutil.Discard
	header := uniprot.Header{}
	_, err := header.WriteTo(writer)
	if err != nil {
		t.Errorf("should always be nil")
	}
}

func TestEntry(t *testing.T) {
	var writer io.Writer = ioutil.Discard
	entry := uniprot.Entry{}
	_, err := entry.WriteTo(writer)
	if err != nil {
		t.Errorf("should always be nil")
	}
}
