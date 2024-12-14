package uniref

import (
	"io"
	"strings"
	"testing"
)

// Test data
const testData = `<?xml version="1.0" encoding="ISO-8859-1" ?>
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
<entry id="UniRef50_UPI00358F51CD" updated="2024-11-27">
<name>Cluster: LOW QUALITY PROTEIN: titin</name>
<property type="member count" value="1"/>
<property type="common taxon" value="Myxine glutinosa"/>
<property type="common taxon ID" value="7769"/>
<representativeMember>
<dbReference type="UniParc ID" id="UPI00358F51CD">
<property type="UniRef100 ID" value="UniRef100_UPI00358F51CD"/>
<property type="UniRef90 ID" value="UniRef90_UPI00358F51CD"/>
<property type="protein name" value="LOW QUALITY PROTEIN: titin"/>
<property type="source organism" value="Myxine glutinosa"/>
<property type="NCBI taxonomy" value="7769"/>
<property type="length" value="47063"/>
<property type="isSeed" value="true"/>
</dbReference>
<sequence length="47063" checksum="48729625616C010E">MSEQ</sequence>
</representativeMember>
</entry>
</UniRef50>`

func TestUniRefParser(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{"TestBasicParsing", testBasicParsing},
		{"TestEmptyHeader", testEmptyHeader},
		{"TestSequentialReading", testSequentialReading},
		{"TestXMLExport", testXMLExport},
		{"TestPropertyAccess", testPropertyAccess},
		{"TestSequenceData", testSequenceData},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.testFunc)
	}
}

func testBasicParsing(t *testing.T) {
	parser, err := NewParser(strings.NewReader(testData))
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	entry, err := parser.Next()
	if err != nil {
		t.Fatalf("Failed to parse first entry: %v", err)
	}

	// Test first entry
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

func testEmptyHeader(t *testing.T) {
	parser, err := NewParser(strings.NewReader(testData))
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

func testSequentialReading(t *testing.T) {
	parser, err := NewParser(strings.NewReader(testData))
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	// First entry
	entry1, err := parser.Next()
	if err != nil {
		t.Fatalf("Failed to parse first entry: %v", err)
	}
	if entry1.ID != "UniRef50_UPI002E2621C6" {
		t.Errorf("First entry: expected ID UniRef50_UPI002E2621C6, got %s", entry1.ID)
	}

	// Second entry
	entry2, err := parser.Next()
	if err != nil {
		t.Fatalf("Failed to parse second entry: %v", err)
	}
	if entry2.ID != "UniRef50_UPI00358F51CD" {
		t.Errorf("Second entry: expected ID UniRef50_UPI00358F51CD, got %s", entry2.ID)
	}

	// Should be EOF now
	_, err = parser.Next()
	if err != io.EOF {
		t.Errorf("Expected EOF after second entry, got %v", err)
	}
}

func testXMLExport(t *testing.T) {
	parser, err := NewParser(strings.NewReader(testData))
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

	// Test that exported XML contains key elements
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

func testPropertyAccess(t *testing.T) {
	parser, err := NewParser(strings.NewReader(testData))
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	entry, err := parser.Next()
	if err != nil {
		t.Fatalf("Failed to parse entry: %v", err)
	}

	// Test property access
	if len(entry.Properties) == 0 {
		t.Fatal("Expected properties to be present")
	}

	// Check specific property values
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

func testSequenceData(t *testing.T) {
	parser, err := NewParser(strings.NewReader(testData))
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	entry, err := parser.Next()
	if err != nil {
		t.Fatalf("Failed to parse entry: %v", err)
	}

	// Test sequence data
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
