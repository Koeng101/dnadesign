package blowq_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/koeng101/dnadesign/lib/bio/blowq"
	"github.com/koeng101/dnadesign/lib/bio/fastq"
)

// Helper function to simulate encoding a fastq.Read for testing.
func encodeMockRead() []byte {
	// Mock a fastq.Read -- fill in actual data as per your schema
	mockRead := fastq.Read{
		Identifier: "e3cc70d5-90ef-49b6-bbe1-cfef99537d73",
		Optionals:  map[string]string{"key": "value"}, // Mock optional data
		Sequence:   "ACTG",                            // Mock sequence
		Quality:    "aaaa",                            // Mock quality
	}

	// Use the EncodeRead function to get binary representation
	encoded, err := blowq.EncodeRead(mockRead)
	if err != nil {
		panic("Failed to encode mock read: " + err.Error())
	}
	return encoded
}

const maxLineSize = 2 * 32 * 1024

func TestWithFastq(t *testing.T) {
	read := fastq.Read{Identifier: "e3cc70d5-90ef-49b6-bbe1-cfef99537d73", Optionals: map[string]string{"runid": "99790f25859e24307203c25273f3a8be8283e7eb", "read": "13956", "ch": "53", "start_time": "2020-11-11T01:49:01Z", "flow_cell_id": "AEI083", "protocol_group_id": "NanoSav2", "sample_id": "nanosavseq2"}, Sequence: "GATGTGCGCCGTTCCAGTTGCGACGTACTATAATCCCCGGCAACACGGTGCTGATTCTCTTCCTGTTCCAGAAAGCATAAACAGATGCAAGTCTGGTGTGATTAACTTCACCAAAGGGCTGGTTGTAATATTAGGAAATCTAACAATAGATTCTGTTGGTTGGACTCTAAAATTAGAAATTTGATAGATTCCTTTTCCCAAATGAAAGTTTAACGTACACTTTGTTTCTAAAGGAAGGTCAAATTACAGTCTACAGCATCGTAATGGTTCATTTTCATTTATATTTTAATACTAGAAAAGTCCTAGGTTGAAGATAACCACATAATAAGCTGCAACTTCAGCTGTCCCAACCTGAAGAAGAATCGCAGGAGTCGAAATAACTTCTGTAAAGCAAGTAGTTTGAACCTATTGATGTTTCAACATGAGCAATACGTAACT", Quality: "$$&%&%#$)*59;/767C378411,***,('11<;:,0039/0&()&'2(/*((4.1.09751).601+'#&&&,-**/0-+3558,/)+&)'&&%&$$'%'%'&*/5978<9;**'3*'&&A?99:;:97:278?=9B?CLJHGG=9<@AC@@=>?=>D>=3<>=>3362$%/((+/%&+//.-,%-4:+..000,&$#%$$%+*)&*0%.//*?<<;>DE>.8942&&//074&$033)*&&&%**)%)962133-%'&*99><<=1144??6.027639.011/-)($#$(/422*4;:=122>?@6964:.5'8:52)*675=:4@;323&&##'.-57*4597)+0&:7<7-550REGB21/0+*79/&/6538())+)+23665+(''$$$'-2(&&*-.-#$&%%$$,-)&$$#$'&,);;<C<@454)#'`"}
	encoded, err := blowq.EncodeRead(read)
	if err != nil {
		t.Errorf("Failed to encode: %s", err)
	}
	_, err = blowq.DecodeRead(encoded)
	if err != nil {
		t.Errorf("Failed to encode: %s", err)
	}
}

func TestNextWithValidData(t *testing.T) {
	// Use the helper function to get encoded data
	encodedData := encodeMockRead()

	// Create a new parser with the encoded data as input
	parser := blowq.NewParser(bytes.NewReader(encodedData), maxLineSize)

	// Attempt to read the next fastq.Read
	read, err := parser.Next()
	if err != nil {
		t.Fatalf("Expected no error, but got %s", err)
	}

	// Validate the read data -- add assertions as per your schema
	if read.Identifier != "e3cc70d5-90ef-49b6-bbe1-cfef99537d73" {
		t.Errorf("Expected identifier 'e3cc70d5-90ef-49b6-bbe1-cfef99537d73', but got '%s'", read.Identifier)
	}
	// Add more validations for sequence, quality, and optionals
}

func TestNextWithInvalidData(t *testing.T) {
	// Simulate invalid data
	invalidData := []byte("clearly not a valid encoded fastq read")

	// Create a new parser with the invalid data
	parser := blowq.NewParser(bytes.NewReader(invalidData), maxLineSize)

	// Attempt to read the next fastq.Read
	_, err := parser.Next()
	if err == nil {
		t.Error("Expected an error for invalid data, but got nil")
	}
}

func TestNextAtEOF(t *testing.T) {
	// Create a new parser with no data to simulate immediate EOF
	parser := blowq.NewParser(bytes.NewReader([]byte{}), maxLineSize)

	// Attempt to read the next fastq.Read
	_, err := parser.Next()
	if err != io.EOF {
		t.Errorf("Expected io.EOF, but got %s", err)
	}
}

// TestDecodeReadWithValidData tests the DecodeRead function with valid input data.
func TestDecodeReadWithValidData(t *testing.T) {
	// Generate some encoded data using the helper function
	encodedData := encodeMockRead()

	// Attempt to decode the data using DecodeRead
	read, err := blowq.DecodeRead(encodedData)
	if err != nil {
		t.Fatalf("Expected no error, but got %s", err)
	}

	// Verify the contents of 'read' are what you expect
	// This will depend on what was in your mock fastq.Read
	// For example:
	if read.Identifier != "e3cc70d5-90ef-49b6-bbe1-cfef99537d73" {
		t.Errorf("Expected identifier 'expected identifier', but got '%s'", read.Identifier)
	}

	// Continue with other verifications as necessary...
}

// TestDecodeReadWithInvalidData tests the DecodeRead function with invalid input data.
func TestDecodeReadWithInvalidData(t *testing.T) {
	// Simulate invalid data
	invalidData := []byte{}

	// Attempt to decode the invalid data
	_, err := blowq.DecodeRead(invalidData)
	if err == nil {
		t.Error("Expected an error for invalid data, but got nil")
	}

	// Optionally, check for the specific type of error if you expect one
}

func TestSRA(t *testing.T) {
	_ = `@SRR2962693.1 1 length=252
TAGGTAACTGGCTATATGAACTTGTAGAAGGTGCTCATTCCAGTCCTCTTGTTCCCAGAGGCTGTGGCTCAAGGCAGCTCTCATGGGTATATTCAAATTGATTGGAGATGCCACTGGAGAGGGTATNAAANNNCCTGGGCTCCTACAGGAACAATGACACTGNCNNNNNNNNNNNNNNNNGCAGCAGCTACAAGATACCCTCTNNNNNNNNANNNNCANNCAATNTGAATATACCCATGAGAGCTGCCTNNN
+SRR2962693.1 1 length=252
BBBCCGGGGGGGGGGGGFGGGGGGGCEGGGGGGGGGGGGGGGGGGGGGGGGGGGGEGG@FEGGGGGGGGGGGGFGBDGGGGGGGGGEEDGGGGGGGGCEB>@GG>0=DFGGGGGCFBG>EDGFCC0#<<?###;@@>FGGGGGGGGGGGGGGGGGGGGGGGG#=################===CGEGGGGGGGGFGGGGGGGG########0####::##:;;F#=:F=GGGGGGGGGG=EEG<E=GGG###
@SRR2962693.2 2 length=252
GTTTTTCCAACAGAACCAGACAGGTTTCTCCTGAAACTCTTTCATTATACGCCATGTACTGTTCATATCCTCATACATCTGCTTTGATCTTCCCCCTCCCCGCTCTCTCTCTCTAACACACACATAAGAAAATNAANTGTGTGTGTGTGTGTGTGAGAGAGAGAGAGAGAGAGAGACAGAGACAGAGAGAGAGAGACAGGGAGAGGGTATATCAAGTATGAGAAGGAACAATGTGTGTATGTGTGTGTGAGA
+SRR2962693.2 2 length=252
BBC?@;DC1=101EE>CDGG0>FGFFDFGE1EDGGFBFFGGGGGGGGGGGEFGC/F@FEGCGG11=1:1>:C1<1=EFCGFEGGE11?1=11010/:C//=/8:/0000<080000000C...880<AB00;1#00#>>0=FDDGGG@>EFGGGGFCBG0<:>0<DG0>FGFGGDECFDFCG1EFFGGGFDEG=CEB0>0=:=/=C/EFG0C00;<000<00088;;C0@00C0<00<088<FFGGG.6CG/`
}
