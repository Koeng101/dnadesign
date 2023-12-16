package sam

import (
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
