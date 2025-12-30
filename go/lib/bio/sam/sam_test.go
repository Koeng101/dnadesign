package sam

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	file, err := os.Open("data/aln.sam")
	if err != nil {
		t.Errorf("Failed to open aln.sam: %s", err)
	}
	parser, header, err := NewParser(file, DefaultMaxLineSize)
	if err != nil {
		t.Errorf("Got error on new parser: %s", err)
	}
	if len(header.HD) != 3 {
		t.Errorf("HD should have 3 TAG:DATA pairs")
	}
	for {
		_, err := parser.Next()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				t.Errorf("Got unknown error: %s", err)
			}
			break
		}
	}
}

func ExampleNewParser() {
	file := strings.NewReader(`@HD	VN:1.6	SO:unsorted	GO:query
@SQ	SN:pOpen_V3_amplified	LN:2482
@PG	ID:minimap2	PN:minimap2	VN:2.24-r1155-dirty	CL:minimap2 -acLx map-ont - APX814_pass_barcode17_e229f2c8_109f9b91_0.fastq.gz
ae9a66f5-bf71-4572-8106-f6f8dbd3b799	16	pOpen_V3_amplified	1	60	8S54M1D3M1D108M1D1M1D62M226S	*	0	0	AGCATGCCGCTTTTCTGTGACTGGTGAGTACTCAACCAAGTCATTCTGAGAATAGTGTATGCGTGCTGAGTTGCTCTTGCCCGGCGTCAATACGGGATAATACCGCGCCACATAGCAGAACTTTAAAAGTGCTCATCATTGGAAAACGTTCTTCGGGGCGAAAACTCTCGACGTTTACCGCTGTTGAGATCCAGTTCGATGTAACCCACTCGTGCACCCAACTGATCTTCAGCATCAGGGCCGAGCGCAGAAGTGGTCCTGCAACTTTATCCGCCTCCATCCAGTCTATTAATTGTTGCCGGAAGCTAGAGTAAGTAGTTCGCCAGTTAATAGTTTGCGCAACGTTGTTGCCATTGCTACAGGCATCGTGGTTACTGTTGATGTTCATGTAGGTGCTGATCAGAGGTACTTTCCTGGAGGGTTTAACCTTAGCAATACGTAACGGAACGAAGTACAGGGCAT	%,<??@@{O{HS{{MOG{EHD@@=)))'&%%%%'(((6::::=?=;:7)'''/33387-)(*025557CBBDDFDECD;1+'(&&')(,-('))35@>AFDCBD{LNKKGIL{{JLKI{{IFG>==86668789=<><;056<;>=87:840/++1,++)-,-0{{&&%%&&),-13;<{HGVKCGFI{J{L{G{INJHEA@C540/3568;>EOI{{{I0000HHRJ{{{{{{{RH{N@@?AKLQEEC?==<433345588==FTA??A@G?@@@EC?==;10//2333?AB?<<<--(++*''&&-(((+@DBJQHJHGGPJH{.---@B?<''-++'--&%%&,,,FC:999IEGJ{HJHIGIFEGIFMDEF;8878{KJGFIJHIHDCAA=<<<<;DDB>:::EK{{@{E<==HM{{{KF{{{MDEQM{ECA?=>9--,.3))'')*++.-,**()%%	NM:i:8	ms:i:408	AS:i:408	nn:i:0	tp:A:P	cm:i:29	s1:i:195	s2:i:0	de:f:0.0345	SA:Z:pOpen_V3_amplified,2348,-,236S134M1D92S,60,1;	rl:i:0`)
	parser, _, _ := NewParser(file, DefaultMaxLineSize)
	samLine, _ := parser.Next()

	fmt.Println(samLine.CIGAR)
	// Output: 8S54M1D3M1D108M1D1M1D62M226S
}

func TestWriteTo(t *testing.T) {
	fileString := `@HD	VN:1.6	SO:unsorted	GO:query
@SQ	SN:pOpen_V3_amplified	LN:2482
@PG	ID:minimap2	PN:minimap2	VN:2.24-r1155-dirty	CL:minimap2 -acLx map-ont - APX814_pass_barcode17_e229f2c8_109f9b91_0.fastq.gz
ae9a66f5-bf71-4572-8106-f6f8dbd3b799	16	pOpen_V3_amplified	1	60	8S54M1D3M1D108M1D1M1D62M226S	*	0	0	AGCATGCCGCTTTTCTGTGACTGGTGAGTACTCAACCAAGTCATTCTGAGAATAGTGTATGCGTGCTGAGTTGCTCTTGCCCGGCGTCAATACGGGATAATACCGCGCCACATAGCAGAACTTTAAAAGTGCTCATCATTGGAAAACGTTCTTCGGGGCGAAAACTCTCGACGTTTACCGCTGTTGAGATCCAGTTCGATGTAACCCACTCGTGCACCCAACTGATCTTCAGCATCAGGGCCGAGCGCAGAAGTGGTCCTGCAACTTTATCCGCCTCCATCCAGTCTATTAATTGTTGCCGGAAGCTAGAGTAAGTAGTTCGCCAGTTAATAGTTTGCGCAACGTTGTTGCCATTGCTACAGGCATCGTGGTTACTGTTGATGTTCATGTAGGTGCTGATCAGAGGTACTTTCCTGGAGGGTTTAACCTTAGCAATACGTAACGGAACGAAGTACAGGGCAT	%,<??@@{O{HS{{MOG{EHD@@=)))'&%%%%'(((6::::=?=;:7)'''/33387-)(*025557CBBDDFDECD;1+'(&&')(,-('))35@>AFDCBD{LNKKGIL{{JLKI{{IFG>==86668789=<><;056<;>=87:840/++1,++)-,-0{{&&%%&&),-13;<{HGVKCGFI{J{L{G{INJHEA@C540/3568;>EOI{{{I0000HHRJ{{{{{{{RH{N@@?AKLQEEC?==<433345588==FTA??A@G?@@@EC?==;10//2333?AB?<<<--(++*''&&-(((+@DBJQHJHGGPJH{.---@B?<''-++'--&%%&,,,FC:999IEGJ{HJHIGIFEGIFMDEF;8878{KJGFIJHIHDCAA=<<<<;DDB>:::EK{{@{E<==HM{{{KF{{{MDEQM{ECA?=>9--,.3))'')*++.-,**()%%	NM:i:8	ms:i:408	AS:i:408	nn:i:0	tp:A:P	cm:i:29	s1:i:195	s2:i:0	de:f:0.0345	SA:Z:pOpen_V3_amplified,2348,-,236S134M1D92S,60,1;	rl:i:0
`
	file := strings.NewReader(fileString)
	parser, _, _ := NewParser(file, DefaultMaxLineSize)
	read, _ := parser.Next()
	header, _ := parser.Header()
	var buffer bytes.Buffer
	_, _ = header.WriteTo(&buffer)
	_, _ = read.WriteTo(&buffer)

	if fileString != buffer.String() {
		t.Errorf("Got diff! First:\n%s\nSecond:\n%s\n====", fileString, buffer.String())
	}
}

// TestValidate ensures that every aspect of validation is covered
func TestValidate(t *testing.T) {
	// Construct an alignment that is correct in all aspects
	validAlignment := Alignment{
		QNAME: "ValidName",
		FLAG:  255,
		RNAME: "*",
		POS:   123456,
		MAPQ:  50,
		CIGAR: "10M1I4M",
		RNEXT: "*",
		PNEXT: 234567,
		TLEN:  1000,
		SEQ:   "ACTGACTGAC",
		QUAL:  "~~~~~~~~~~",
	}

	// Should pass (no error)
	if err := validAlignment.Validate(); err != nil {
		t.Errorf("Valid alignment did not pass validation: %s", err)
	}

	// Test cases for each field
	testCases := []struct {
		modify   func(a *Alignment)
		expected string
	}{
		{ // Invalid QNAME
			func(a *Alignment) { a.QNAME = "Invalid QNAME due to length and spaces" },
			"Invalid QNAME",
		},
		{ // Invalid RNAME
			func(a *Alignment) { a.RNAME = "Invalid RNAME" },
			"Invalid RNAME",
		},
		{ // Invalid POS, out of range
			func(a *Alignment) { a.POS = -1 },
			"Invalid POS",
		},
		{ // Invalid CIGAR
			func(a *Alignment) { a.CIGAR = "X" },
			"Invalid CIGAR",
		},
		{ // Invalid RNEXT
			func(a *Alignment) { a.RNEXT = "Invalid RNEXT" },
			"Invalid RNEXT",
		},
		{ // Invalid PNEXT, out of range
			func(a *Alignment) { a.PNEXT = -1 },
			"Invalid PNEXT",
		},
		{ // Invalid SEQ
			func(a *Alignment) { a.SEQ = "ACTG123" },
			"Invalid SEQ",
		},
		{ // Invalid QUAL
			func(a *Alignment) { a.QUAL = "qual string with lower case or invalid characters" },
			"Invalid QUAL",
		},
	}

	for _, tc := range testCases {
		// Copy the valid alignment and modify it for the test
		invalidAlignment := validAlignment
		tc.modify(&invalidAlignment)

		// Now validate it
		err := invalidAlignment.Validate()
		if err == nil || !contains(err.Error(), tc.expected) {
			t.Errorf("Expected error for %s but got none or wrong error: %s", tc.expected, err)
		}
	}
}

// contains is a helper function to check if errStr contains the expected substring
func contains(errStr, expected string) bool {
	return errStr != "" && strings.Contains(errStr, expected)
}

// TestValidateAllInOne - testing all validation rules in one function
func TestValidateAllInOne(t *testing.T) {
	// Define a series of headers to test different validation scenarios
	tests := []struct {
		name          string
		header        *Header
		expectedError error
	}{
		// Valid Complete Header
		{
			name: "Valid Complete Header",
			header: &Header{
				HD: map[string]string{"VN": "1.0", "SO": "unsorted", "GO": "none", "SS": "coordinate:example"},
				SQ: []map[string]string{{"SN": "chr1", "LN": "1000", "TP": "linear"}},
				RG: []map[string]string{{"ID": "rg1", "PL": "ILLUMINA", "FO": "*", "DT": "2023-01-01"}},
				PG: []map[string]string{{"ID": "pg1"}},
				CO: []string{"This is a comment."},
			},
			expectedError: nil,
		},
		// Invalid @HD VN format
		{
			name: "Invalid @HD VN format",
			header: &Header{
				HD: map[string]string{"VN": "abc"}, // Invalid VN format
			},
			expectedError: fmt.Errorf("Invalid format for @HD VN. Accepted format: /^[0-9]+\\.[0-9]+$/.\nGot: %s", "abc"),
		},
		// Invalid @HD SO value
		{
			name: "Invalid @HD SO value",
			header: &Header{
				HD: map[string]string{"VN": "1.0", "SO": "invalid_so"}, // Invalid SO value
			},
			expectedError: fmt.Errorf("Invalid value for @HD SO. Valid values: unknown (default), unsorted, queryname and coordinate. Got: %s", "invalid_so"),
		},
		// Invalid @HD GO value
		{
			name: "Invalid @HD GO value",
			header: &Header{
				HD: map[string]string{"VN": "1.0", "GO": "invalid_go"}, // Invalid GO value
			},
			expectedError: fmt.Errorf("Invalid value for @HD GO. Valid values: none (default), query (alignments are grouped by QNAME), and reference (alignments are grouped by RNAME/POS). Got: %s", "invalid_go"),
		},
		// Invalid @HD SS format
		{
			name: "Invalid @HD SS format",
			header: &Header{
				HD: map[string]string{"VN": "1.0", "SS": "invalid_ss"}, // Invalid SS format
			},
			expectedError: fmt.Errorf("Invalid format for @HD SS. Needs to match: Regular expression: (coordinate|queryname|unsorted)(:[A-Za-z0-9_-]+)+\nGot: %s", "invalid_ss"),
		},
		// Invalid @SQ LN range
		{
			name: "Invalid @SQ LN range",
			header: &Header{
				SQ: []map[string]string{{"SN": "chr1", "LN": "2147483648"}}, // Invalid LN range
			},
			expectedError: fmt.Errorf("Invalid value for @SQ LN. Range: [1, 231 − 1], Got: %d", 2147483648),
		},
		// Invalid @SQ TP value
		{
			name: "Invalid @SQ TP value",
			header: &Header{
				SQ: []map[string]string{{"SN": "chr1", "LN": "1000", "TP": "invalid_tp"}}, // Invalid TP value
			},
			expectedError: fmt.Errorf("Invalid value for @SQ TP. Valid values: linear (default) and circular, Got: %s", "invalid_tp"),
		},
		// Non-unique @RG ID
		{
			name: "Non-unique @RG ID",
			header: &Header{
				RG: []map[string]string{{"ID": "rg1", "PL": "ILLUMINA"}, {"ID": "rg1", "PL": "SOLID"}},
			},
			expectedError: fmt.Errorf("Non-unique @RG ID. Got: %s", "rg1"),
		},
		// Invalid @RG FO format
		{
			name: "Invalid @RG FO format",
			header: &Header{
				RG: []map[string]string{{"ID": "rg1", "FO": "invalid_fo"}},
			},
			expectedError: fmt.Errorf("Invalid format for @RG FO. Required regexp format: /\\*|[ACMGRSVTWYHKDBN]+/\nGot: %s", "invalid_fo"),
		},
		// Invalid @RG PL value
		{
			name: "Invalid @RG PL value",
			header: &Header{
				RG: []map[string]string{{"ID": "rg1", "PL": "invalid_pl"}},
			},
			expectedError: fmt.Errorf("Invalid value for @RG PL. Valid values: CAPILLARY, DNBSEQ (MGI/BGI), ELEMENT, HELICOS, ILLUMINA, IONTORRENT, LS454, ONT (Oxford Nanopore), PACBIO (Pacific Bio-sciences), SOLID, and ULTIMA. Got: %s", "invalid_pl"),
		},
		// Non-unique @PG ID
		{
			name: "Non-unique @PG ID",
			header: &Header{
				PG: []map[string]string{{"ID": "pg1"}, {"ID": "pg1"}},
			},
			expectedError: fmt.Errorf("Non-unique @PG ID. Got: %s", "pg1"),
		},
		// Non-unique @SN SQ
		{
			name: "Invalid @SQ SN format",
			header: &Header{
				SQ: []map[string]string{{"SN": "invalid_sn", "LN": "1000"}, {"SN": "invalid_sn"}}, // Invalid SN format
			},
			expectedError: fmt.Errorf("Non-unique @SQ SN: %s", "invalid_sn"),
		},
	}

	// Iterate through each test case
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Run the validate function on the header
			err := tc.header.Validate()

			// Check if the error matches the expected error
			if (err != nil && tc.expectedError == nil) || (err == nil && tc.expectedError != nil) || (err != nil && tc.expectedError != nil && err.Error() != tc.expectedError.Error()) {
				t.Errorf("Test %v - Got error %v, want %v", tc.name, err, tc.expectedError)
			}
		})
	}
}

func TestPrimary(t *testing.T) {
	// Define test cases
	tests := []struct {
		name string
		flag uint16
		want bool
	}{
		{"No Flags", 0x0, true},
		{"Secondary Alignment", 0x100, false},
		{"Supplementary Alignment", 0x800, false},
		{"Secondary and Supplementary", 0x900, false},
		{"PCR or Optical Duplicate", 0x400, true},
		{"Reverse Complemented", 0x10, true},
		// ... other test cases
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create an Alignment with the given FLAG
			a := Alignment{FLAG: tc.flag}
			// Call the Primary function
			got := Primary(a)
			// Assert that the result is as expected
			if got != tc.want {
				t.Errorf("Primary() with FLAG 0x%x = %v, want %v", tc.flag, got, tc.want)
			}
		})
	}
}
