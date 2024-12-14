package uniref

import (
	"strings"
	"testing"
)

// Test data for each UniRef version
const (
	testData50 = `<?xml version="1.0" encoding="ISO-8859-1" ?>
<UniRef50 xmlns="http://uniprot.org/uniref" 
xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" 
xsi:schemaLocation="http://uniprot.org/uniref http://www.uniprot.org/support/docs/uniref.xsd" 
 releaseDate="2024-11-27" version="2024_06"> 
<entry id="UniRef50_UPI002E2621C6" updated="2024-05-29">
<name>Cluster: uncharacterized protein LOC134193701</name>
<property type="member count" value="1"/>
<property type="common taxon" value="Corticium candelabrum"/>
<property type="common taxon ID" value="121492"/>
<representativeMember>
<dbReference type="UniParc ID" id="UPI002E2621C6">
<property type="UniRef100 ID" value="UniRef100_UPI002E2621C6"/>
<property type="UniRef90 ID" value="UniRef90_UPI002E2621C6"/>
<property type="protein name" value="uncharacterized protein LOC134193701"/>
<property type="source organism" value="Corticium candelabrum"/>
<property type="NCBI taxonomy" value="121492"/>
<property type="length" value="49499"/>
<property type="isSeed" value="true"/>
</dbReference>
<sequence length="49499" checksum="428270C7C0D6A56C">MGR</sequence>
</representativeMember>
</entry>
</UniRef50>`

	testData90 = `<?xml version="1.0" encoding="ISO-8859-1" ?>
<UniRef90 xmlns="http://uniprot.org/uniref" 
xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" 
xsi:schemaLocation="http://uniprot.org/uniref http://www.uniprot.org/support/docs/uniref.xsd" 
 releaseDate="2024-11-27" version="2024_06"> 
<entry id="UniRef90_UPI002E2621C6" updated="2024-05-29">
<name>Cluster: uncharacterized protein LOC134193701</name>
<property type="member count" value="1"/>
<property type="common taxon" value="Corticium candelabrum"/>
<property type="common taxon ID" value="121492"/>
<representativeMember>
<dbReference type="UniParc ID" id="UPI002E2621C6">
<property type="UniRef100 ID" value="UniRef100_UPI002E2621C6"/>
<property type="protein name" value="uncharacterized protein LOC134193701"/>
<property type="source organism" value="Corticium candelabrum"/>
<property type="NCBI taxonomy" value="121492"/>
<property type="length" value="49499"/>
<property type="isSeed" value="true"/>
</dbReference>
<sequence length="49499" checksum="428270C7C0D6A56C">MGR</sequence>
</representativeMember>
</entry>
</UniRef90>`

	testData100 = `<?xml version="1.0" encoding="ISO-8859-1" ?>
<UniRef100 xmlns="http://uniprot.org/uniref" 
xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" 
xsi:schemaLocation="http://uniprot.org/uniref http://www.uniprot.org/support/docs/uniref.xsd" 
 releaseDate="2024-11-27" version="2024_06"> 
<entry id="UniRef100_UPI002E2621C6" updated="2024-05-29">
<name>Cluster: uncharacterized protein LOC134193701</name>
<property type="member count" value="1"/>
<property type="common taxon" value="Corticium candelabrum"/>
<property type="common taxon ID" value="121492"/>
<representativeMember>
<dbReference type="UniParc ID" id="UPI002E2621C6">
<property type="protein name" value="uncharacterized protein LOC134193701"/>
<property type="source organism" value="Corticium candelabrum"/>
<property type="NCBI taxonomy" value="121492"/>
<property type="length" value="49499"/>
<property type="isSeed" value="true"/>
</dbReference>
<sequence length="49499" checksum="428270C7C0D6A56C">MGR</sequence>
</representativeMember>
</entry>
</UniRef100>`
)

func TestUniRefVersions(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		version string
	}{
		{"UniRef50", testData50, "50"},
		{"UniRef90", testData90, "90"},
		{"UniRef100", testData100, "100"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, err := NewParser(strings.NewReader(tt.data))
			if err != nil {
				t.Fatalf("Failed to create parser for %s: %v", tt.name, err)
			}

			entry, err := parser.Next()
			if err != nil {
				t.Fatalf("Failed to parse first entry for %s: %v", tt.name, err)
			}

			expectedID := "UniRef" + tt.version + "_UPI002E2621C6"
			if entry.ID != expectedID {
				t.Errorf("Expected ID %s, got %s", expectedID, entry.ID)
			}

			if parser.uniref.GetUniRefVersion() != tt.version {
				t.Errorf("Expected version %s, got %s", tt.version, parser.uniref.GetUniRefVersion())
			}
		})
	}
}

func TestBasicParsing(t *testing.T) {
	parser, err := NewParser(strings.NewReader(testData50))
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	entry, err := parser.Next()
	if err != nil {
		t.Fatalf("Failed to parse first entry: %v", err)
	}

	if entry.ID != "UniRef50_UPI002E2621C6" {
		t.Errorf("Expected ID UniRef50_UPI002E2621C6, got %s", entry.ID)
	}
	if entry.Name != "Cluster: uncharacterized protein LOC134193701" {
		t.Errorf("Expected name 'Cluster: uncharacterized protein LOC134193701', got %s", entry.Name)
	}
	if len(entry.Properties) != 3 {
		t.Errorf("Expected 3 properties, got %d", len(entry.Properties))
	}
}

func TestEmptyHeader(t *testing.T) {
	parser, err := NewParser(strings.NewReader(testData50))
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	header, err := parser.Header()
	if err != nil {
		t.Errorf("Expected no error for empty header, got %v", err)
	}
	if header != (Header{}) {
		t.Error("Expected empty header struct")
	}
}

func TestSequenceData(t *testing.T) {
	parser, err := NewParser(strings.NewReader(testData50))
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	entry, err := parser.Next()
	if err != nil {
		t.Fatalf("Failed to parse entry: %v", err)
	}

	sequence := entry.RepMember.Sequence
	if sequence == nil {
		t.Fatal("Expected sequence to be present")
	}

	expectedTests := []struct {
		name     string
		got      interface{}
		expected interface{}
	}{
		{"Length", sequence.Length, 49499},
		{"Checksum", sequence.Checksum, "428270C7C0D6A56C"},
		{"Value", sequence.Value, "MGR"},
	}

	for _, tt := range expectedTests {
		if tt.got != tt.expected {
			t.Errorf("%s: expected %v, got %v", tt.name, tt.expected, tt.got)
		}
	}
}

func TestPropertyAccess(t *testing.T) {
	parser, err := NewParser(strings.NewReader(testData50))
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	entry, err := parser.Next()
	if err != nil {
		t.Fatalf("Failed to parse entry: %v", err)
	}

	if len(entry.Properties) == 0 {
		t.Fatal("Expected properties to be present")
	}

	memberCountFound := false
	for _, prop := range entry.Properties {
		if prop.Type == "member count" && prop.Value == "1" {
			memberCountFound = true
			break
		}
	}
	if !memberCountFound {
		t.Error("Expected to find member count property with value '1'")
	}
}

func TestXMLExport(t *testing.T) {
	parser, err := NewParser(strings.NewReader(testData50))
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	entry, err := parser.Next()
	if err != nil {
		t.Fatalf("Failed to parse entry: %v", err)
	}

	xml, err := entry.ToXML()
	if err != nil {
		t.Fatalf("Failed to export XML: %v", err)
	}

	expectedElements := []string{
		`id="UniRef50_UPI002E2621C6"`,
		`updated="2024-05-29"`,
		`<name>Cluster: uncharacterized protein LOC134193701</name>`,
		`checksum="428270C7C0D6A56C"`,
		`>MGR</sequence>`,
	}

	for _, expected := range expectedElements {
		if !strings.Contains(xml, expected) {
			t.Errorf("Expected XML to contain '%s', but it didn't", expected)
		}
	}
}
