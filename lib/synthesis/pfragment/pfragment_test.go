package pfragment

import (
	"bytes"
	_ "embed"
	"sync"
	"testing"

	"github.com/koeng101/dnadesign/lib/bio"
	"github.com/koeng101/dnadesign/lib/synthesis/codon"
	"github.com/koeng101/dnadesign/lib/synthesis/fix"
)

//go:embed data/order.fasta
var orderTestFasta []byte

func TestNaiveFragmentProtein(t *testing.T) {
	parser, err := bio.NewFastaParser(bytes.NewReader(orderTestFasta))
	if err != nil {
		t.Errorf("Failed to make fasta parser: %s", err)
	}
	fastas, err := parser.Parse()
	if err != nil {
		t.Errorf("Failed to parse fasta file: %s", err)
	}
	var proteins []string
	for _, fsta := range fastas {
		proteins = append(proteins, fsta.Sequence)
	}
	_, _ = NaiveFragmentProtein(proteins, 4, 280, 310)
}

func TestNaiveOptimization(t *testing.T) {
	parser, err := bio.NewFastaParser(bytes.NewReader(orderTestFasta))
	if err != nil {
		t.Errorf("Failed to make fasta parser: %s", err)
	}
	fastas, err := parser.Parse()
	if err != nil {
		t.Errorf("Failed to parse fasta file: %s", err)
	}
	var proteins []string
	for _, fsta := range fastas {
		proteins = append(proteins, fsta.Sequence)
	}
	var functions []func(string, chan fix.DnaSuggestion, *sync.WaitGroup)
	functions = append(functions, fix.RemoveSequence([]string{"GGTCTC"}, "Removal requested by user"))
	functions = append(functions, fix.RemoveSequence([]string{"AAAAAAAA", "GGGGGGGG"}, "Homopolymers"))
	functions = append(functions, fix.RemoveRepeat(18))

	ct := codon.ParseCodonJSON(codon.EcoliCodonTable)

	_, err = NaiveProteinFragmentationAndOptimization(proteins, 4, 280, 310, *ct, functions, "AATG", "TAAATCC")
	if err != nil {
		t.Error(err)
	}
}
