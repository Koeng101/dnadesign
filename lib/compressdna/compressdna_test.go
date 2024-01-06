package compressdna

import (
	"errors"
	"fmt"
	"testing"

	"github.com/koeng101/dnadesign/lib/random"
)

func TestCompressDNA(t *testing.T) {
	// Define test cases
	longDna, _ := random.DNASequence(300, 0)
	longerDna, _ := random.DNASequence(66000, 0)
	tests := []struct {
		name          string
		dna           string
		expectedLen   int  // Expected length of the compressed data
		expectedFlag  byte // Expected flag byte
		expectedError error
	}{
		{"Empty", "", 2, 0x00, errors.New("DNA sequence is empty")},
		{"Bad", "u", 1, 0x00, errors.New("invalid character in DNA sequence: u")},
		{"Short", "ATGC", 3, 0x00, errors.New("")},
		{"Medium", "ATGCGTATGCCGTAGC", 6, 0x00, errors.New("")},
		{"Long", longDna, 78, 0x01, errors.New("")},
		{"Longest", longerDna, 16505, 0x02, errors.New("")},
		// Add more test cases for longer sequences and edge cases
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			compressed, err := CompressDNA(tc.dna, false)
			if err != nil {
				if err.Error() != tc.expectedError.Error() {
					t.Errorf("Got unexpected error: %s", err)
				}
			} else {
				if len(compressed) != tc.expectedLen {
					t.Errorf("CompressDNA() with input %s, expected length %d, got %d", "", tc.expectedLen, len(compressed))
				}
				if compressed[0] != tc.expectedFlag {
					t.Errorf("CompressDNA() with input %s, expected flag %b, got %b", tc.dna, tc.expectedFlag, compressed[0])
				}
			}
		})
	}
}

func TestDecompressDNA(t *testing.T) {
	longDna, _ := random.DNASequence(300, 0)
	longerDna, _ := random.DNASequence(66000, 0)
	// Define test cases
	tests := []struct {
		name     string
		dna      string
		expected string
	}{
		{"Empty", "", ""},
		{"Short", "ATGC", "ATGC"},
		{"Medium", "ATGCGTATGCCGTAGC", "ATGCGTATGCCGTAGC"},
		{"Long", longDna, longDna},
		{"Longest", longerDna, longerDna},
		// Add more test cases as needed
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			compressed, _ := CompressDNA(tc.dna, false)
			decompressed, _ := DecompressDNA(compressed)
			if decompressed != tc.expected {
				t.Errorf("DecompressDNA() with input %v, expected %s, got %s", compressed, tc.expected, decompressed)
			}
		})
	}
	_, err := DecompressDNA([]byte{230, 0x00, 0x00, 0x00})
	if err == nil {
		t.Errorf("Expected bad flag")
	}
}

func TestCompressDNAWithQuality(t *testing.T) {
	quality := `$$&%&%#$)*59;/767C378411,***,('11<;:,0039/0&()&'2(/*((4.1.09751).601+'#&&&,-**/0-+3558,/)+&)'&&%&$$'%'%'&*/5978<9;**'3*'&&A?99:;:97:278?=9B?CLJHGG=9<@AC@@=>?=>D>=3<>=>3362$%/((+/%&+//.-,%-4:+..000,&$#%$$%+*)&*0%.//*?<<;>DE>.8942&&//074&$033)*&&&%**)%)962133-%'&*99><<=1144??6.027639.011/-)($#$(/422*4;:=122>?@6964:.5'8:52)*675=:4@;323&&##'.-57*4597)+0&:7<7-550REGB21/0+*79/&/6538())+)+23665+(''$$$'-2(&&*-.-#$&%%$$,-)&$$#$'&,);;<C<@454)#'`
	sequence := "GATGTGCGCCGTTCCAGTTGCGACGTACTATAATCCCCGGCAACACGGTGCTGATTCTCTTCCTGTTCCAGAAAGCATAAACAGATGCAAGTCTGGTGTGATTAACTTCACCAAAGGGCTGGTTGTAATATTAGGAAATCTAACAATAGATTCTGTTGGTTGGACTCTAAAATTAGAAATTTGATAGATTCCTTTTCCCAAATGAAAGTTTAACGTACACTTTGTTTCTAAAGGAAGGTCAAATTACAGTCTACAGCATCGTAATGGTTCATTTTCATTTATATTTTAATACTAGAAAAGTCCTAGGTTGAAGATAACCACATAATAAGCTGCAACTTCAGCTGTCCCAACCTGAAGAAGAATCGCAGGAGTCGAAATAACTTCTGTAAAGCAAGTAGTTTGAACCTATTGATGTTTCAACATGAGCAATACGTAACT"
	compressed, err := CompressDNAWithQuality(sequence, quality, false)
	if err != nil {
		t.Errorf("Got unexpected error while CompressDNAWithQuality: %s", err)
	}
	decompressedSequence, decompressedQuality, err := DecompressDNAWithQuality(compressed)
	if err != nil {
		t.Errorf("Got unexpected error while DecompressDNAWithQuality: %s", err)
	}
	if decompressedSequence != sequence {
		fmt.Println(decompressedSequence)
		t.Errorf("Got different sequences")
	}
	if decompressedQuality != quality {
		t.Errorf("Got different qualities")
	}
}
