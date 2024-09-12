package plasmidcall

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/koeng101/dnadesign/lib/align/cs"
	"github.com/koeng101/dnadesign/lib/bio/fastq"
)

func TestCallMutations(t *testing.T) {
	// Read the TSV file
	file, err := os.Open(filepath.Join(".", "data", "59QgiHyryWfVpvsAqFeboj.tsv"))
	if err != nil {
		t.Fatalf("Failed to open TSV file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// Read and discard the header
	if !scanner.Scan() {
		t.Fatalf("Failed to read TSV header: %v", scanner.Err())
	}

	var alignments []CsAlignment

	// Read and parse each line
	for scanner.Scan() {
		line := scanner.Text()
		fields := parseFields(line)

		if len(fields) < 7 {
			t.Fatalf("Invalid number of fields in line: %s", line)
		}

		flag, err := strconv.Atoi(fields[0])
		if err != nil {
			t.Fatalf("Failed to parse flag: %v", err)
		}

		if flag != 0 && flag != 16 {
			continue
		}

		alignment := CsAlignment{
			Read: fastq.Read{
				Identifier: fields[4],
				Sequence:   fields[5],
				Quality:    fields[6],
			},
			ReverseComplement: flag == 16,
			CS:                cs.ParseCS(fields[3]),
		}

		alignments = append(alignments, alignment)
	}
	fmt.Println(len(alignments))

	referenceSequence := "GGGTTATTGTCTCATGAGCGGATACATATTTGAATGTATTTAGAAAAATAAACAAATAGGGGTTCCGCGCACCTGCACCAGTCAGTAAAACGACGGCCAGTGACTTGGTCTCATACGCACTGATGCGCTACGAGGACGTGTACATGCCGATCGTGTACTGCGCAATCAACATCGCACAGAACTACGCATGGAAGATCAACAAGCAGGTGCTGGCAGTGGCACGCGTGATCACGAAGTGGAAGCACTGCAAGGTGGAGGACATCCCGGCAATCGAGAACGGTGAGCTGTACATGAAGCCGGAGGACATCGACATGAACCCGGAGGCACTGACGGCATGGAAGATGGCAGCAGCAGCAGTGTACCGCAAGGACAAGGCACGCAAGTCGCGCCGCATCATGCTGGAGTTCATGCTGGAGGGTGCACGCAAGTTCGCAAACCACAAGGCATGCTGGTTCCCGTCGAACATGGACTGGCGCGGTCGCCAGTACGCAGTGGGTATGTTCAACCCGCAGGGTAACGACATGACGAAGCTGCTGCTGACGCTGGCAAAGGGTAAGACGATCGGTAAGGAGGGTTACTACTGGCTGAAGATCCACGGTGCAAACTGCGCAGGTGTGCACAAGGTGCCGTTCCCGGAGCGCATCAAGTTCATCGAGGAGAACCACGAGAACATCATGGCATGCGCAAAGCTGCCGCTGGAGAACGTGTGGCACGCAGAGCAGGACTCGCCGTTCTGCTTCCTGGCATTCTGCTTCGAGTACGCAGGTGTGCAGCACCACGGTCTGTCGTACAACTGCTCGCTGGACCTGGCAGTGGACGGTTCGTGCTCGGGTATCCAGCACTTCTCGGCAATGCTGCGCGACGAGGTGGGTGGTCGCCTGGTGAACCTGCTGCCGTCGGAGACGGTGCAGGACATCTACGGTATCGTGGCAAAGAAGCCGAACGAGATCCTGCAGGCATCGGCAATCAACGGTACGGACAACGAGGTGGTGACGGTGACGGACGAGCTGACGGGTGAGACGAGACCAAGTCGTCATAGCTGTTTCCTGAGAGCTTGGCAGGTGATGAC"

	// Call the function under test
	mutations, err := CallMutations(referenceSequence, alignments)
	if err != nil {
		t.Fatalf("CallMutations failed: %v", err)
	}
	fmt.Println(mutations)

	// TODO: Add assertions to check the correctness of the mutations
	// For example:
	// if len(mutations) == 0 {
	//     t.Errorf("Expected mutations, but got none")
	// }

	// Print out the mutations for inspection
	t.Logf("Found %d mutations:", len(mutations))
	for i, mutation := range mutations {
		t.Logf("Mutation %d: %+v", i+1, mutation)
	}
}

// parseFields splits a line into fields, handling quoted fields
func parseFields(line string) []string {
	var fields []string
	var field strings.Builder
	inQuote := false

	for _, r := range line {
		switch r {
		case '"':
			inQuote = !inQuote
		case '\t':
			if !inQuote {
				fields = append(fields, field.String())
				field.Reset()
			} else {
				field.WriteRune(r)
			}
		default:
			field.WriteRune(r)
		}
	}

	// Add the last field
	fields = append(fields, field.String())

	return fields
}
