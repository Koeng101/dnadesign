package fold

import (
	"math"
	"strings"
	"testing"
)

func TestFold(t *testing.T) {
	t.Run("FoldCache", func(t *testing.T) {
		seq := "ATGGATTTAGATAGAT"
		foldContext, err := newFoldingContext(seq, 37.0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		res, err := Zuker(seq, 37.0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		seqDg := res.MinimumFreeEnergy()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if math.Abs(seqDg-foldContext.unpairedMinimumFreeEnergyW[0][len(seq)-1].energy) > 1 {
			t.Errorf("expected values to be within delta of 1, got %v and %v", seqDg, foldContext.unpairedMinimumFreeEnergyW[0][len(seq)-1].energy)
		}
	})
	t.Run("FoldDNA", func(t *testing.T) {
		// unafold's estimates for free energy estimates of DNA oligos
		unafoldDgs := map[string]float64{
			"GGGAGGTCGTTACATCTGGGTAACACCGGTACTGATCCGGTGACCTCCC":                         -10.94, // three branched structure
			"GGGAGGTCGCTCCAGCTGGGAGGAGCGTTGGGGGTATATACCCCCAACACCGGTACTGATCCGGTGACCTCCC": -23.4,  // four branched structure
			"CGCAGGGAUACCCGCG":                         -3.8,
			"TAGCTCAGCTGGGAGAGCGCCTGCTTTGCACGCAGGAGGT": -6.85,
			"GGGGGCATAGCTCAGCTGGGAGAGCGCCTGCTTTGCACGCAGGAGGTCTGCGGTTCGATCCCGCGCGCTCCCACCA": -15.50,
			"TGAGACGGAAGGGGATGATTGTCCCCTTCCGTCTCA":                                         -18.10,
			"ACCCCCTCCTTCCTTGGATCAAGGGGCTCAA":                                              -3.65,
		}

		for seq, ufold := range unafoldDgs {
			res, err := Zuker(seq, 37.0)
			if err != nil {
				t.Fatalf("unexpected error for sequence %s: %v", seq, err)
			}
			d := res.MinimumFreeEnergy()

			// accepting a 60% difference
			delta := math.Abs(0.6 * math.Min(d, ufold))
			if math.Abs(d-ufold) > delta {
				t.Errorf("for sequence %s, expected value within delta of %v, got %v and %v", seq, delta, d, ufold)
			}
		}
	})
	t.Run("FoldRNA", func(t *testing.T) {
		// unafold's estimates for free energy estimates of RNA oligos
		// most tests available at https://github.com/jaswindersingh2/SPOT-RNA/blob/master/sample_inputs/batch_seq.fasta
		unafoldDgs := map[string]float64{
			"ACCCCCUCCUUCCUUGGAUCAAGGGGCUCAA":        -9.5,
			"AAGGGGUUGGUCGCCUCGACUAAGCGGCUUGGAAUUCC": -10.1,
			"UUGGAGUACACAACCUGUACACUCUUUC":           -4.3,
			"AGGGAAAAUCCC":                           -3.3,
			"GCUUACGAGCAAGUUAAGCAAC":                 -4.6,
			"UGGGAGGUCGUCUAACGGUAGGACGGCGGACUCUGGAUCCGCUGGUGGAGGUUCGAGUCCUCCCCUCCCAGCCA":   -32.8,
			"GGGCGAUGAGGCCCGCCCAAACUGCCCUGAAAAGGGCUGAUGGCCUCUACUG":                         -20.7,
			"GGGGGCAUAGCUCAGCUGGGAGAGCGCCUGCUUUGCACGCAGGAGGUCUGCGGUUCGAUCCCGCGCGCUCCCACCA": -31.4,
		}

		for seq, ufold := range unafoldDgs {
			res, err := Zuker(seq, 37.0)
			if err != nil {
				t.Fatalf("unexpected error for sequence %s: %v", seq, err)
			}
			d := res.MinimumFreeEnergy()

			// accepting a 50% difference
			delta := math.Abs(0.5 * math.Min(d, ufold))
			if math.Abs(d-ufold) > delta {
				t.Errorf("for sequence %s, expected value within delta of %v, got %v and %v", seq, delta, d, ufold)
			}
		}
	})
	t.Run("DotBracket", func(t *testing.T) {
		seq := "GGGAGGTCGTTACATCTGGGTAACACCGGTACTGATCCGGTGACCTCCC"
		res, err := Zuker(seq, 37.0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expected := "((((((((.((((......))))..((((.......)))).))))))))"
		if res.DotBracket() != expected {
			t.Errorf("expected %s, got %s", expected, res.DotBracket())
		}
	})
	t.Run("multibranch", func(t *testing.T) {
		seq := "GGGAGGTCGTTACATCTGGGTAACACCGGTACTGATCCGGTGACCTCCC" // three branch

		res, err := Zuker(seq, 37.0)
		if err != nil {
			t.Fatalf("Error: %v", err) // Fatal stops the test immediately
		}

		found := false
		foundIJ := subsequence{7, 41}
		for _, s := range res.structs {
			if strings.Contains(s.description, "BIFURCATION") {
				for _, ij := range s.inner {
					if ij == foundIJ {
						found = true
					}
				}
			}
		}
		if !found {
			t.Errorf("not found a BIFURCATION with (7, 41) in ij") // Error continues the test
		}
	})
	t.Run("pair", func(t *testing.T) {
		seq := "ATGGAATAGTG"
		result := pair(seq, 0, 1, 9, 10)
		expected := "AT/TG"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})
	t.Run("stack", func(t *testing.T) {
		seq := "GCUCAGCUGGGAGAGC"
		temp := 37.0
		foldContext, err := newFoldingContext(seq, temp)
		if err != nil {
			t.Fatalf("Error creating folding context: %v", err)
		}

		e := stack(1, 2, 14, 13, foldContext)
		expectedDelta := -2.1
		tolerance := 0.1
		if e < expectedDelta-tolerance || e > expectedDelta+tolerance {
			t.Errorf("Expected value to be within %.1f +/- %.1f, got %.2f", expectedDelta, tolerance, e)
		}
	})
	t.Run("Bulge", func(t *testing.T) {
		// mock bulge of CAT on one side and AG on other
		// from pg 429 of SantaLucia, 2004
		seq := "ACCCCCATCCTTCCTTGAGTCAAGGGGCTCAA"
		foldContext, err := newFoldingContext(seq, 37)
		if err != nil {
			t.Fatalf("Error creating folding context: %v", err)
		}

		pairDg, err := Bulge(5, 7, 18, 17, foldContext)
		if err != nil {
			t.Fatalf("Error in Bulge function: %v", err)
		}
		expectedDelta := 3.22
		tolerance := 0.4
		if pairDg < expectedDelta-tolerance || pairDg > expectedDelta+tolerance {
			t.Errorf("Expected value to be within %.2f +/- %.1f, got %.2f", expectedDelta, tolerance, pairDg)
		}
	})
	t.Run("hairpin", func(t *testing.T) {
		// Test case 1
		seq := "ACCCCCTCCTTCCTTGGATCAAGGGGCTCAA"
		i, j := 11, 16

		foldContext, err := newFoldingContext(seq, 37)
		if err != nil {
			t.Fatalf("Error creating folding context: %v", err)
		}

		hairpinDg, err := hairpin(i, j, foldContext)
		if err != nil {
			t.Fatalf("Error in hairpin function: %v", err)
		}
		if hairpinDg < 4.3-1.0 || hairpinDg > 4.3+1.0 {
			t.Errorf("hairpinDg = %v, want %v +/- %v", hairpinDg, 4.3, 1.0)
		}

		// Test case 2
		seq = "ACCCGCAAGCCCTCCTTCCTTGGATCAAGGGGCTCAA"
		i, j = 3, 8

		foldContext, err = newFoldingContext(seq, 37)
		if err != nil {
			t.Fatalf("Error creating folding context: %v", err)
		}

		hairpinDg, err = hairpin(i, j, foldContext)
		if err != nil {
			t.Fatalf("Error in hairpin function: %v", err)
		}
		if hairpinDg < 0.67-0.1 || hairpinDg > 0.67+0.1 {
			t.Errorf("hairpinDg = %v, want %v +/- %v", hairpinDg, 0.67, 0.1)
		}

		// Test case 3
		seq = "CUUUGCACG"
		i, j = 0, 8

		foldContext, err = newFoldingContext(seq, 37)
		if err != nil {
			t.Fatalf("Error creating folding context: %v", err)
		}

		hairpinDg, err = hairpin(i, j, foldContext)
		if err != nil {
			t.Fatalf("Error in hairpin function: %v", err)
		}
		if hairpinDg < 4.5-0.2 || hairpinDg > 4.5+0.2 {
			t.Errorf("hairpinDg = %v, want %v +/- %v", hairpinDg, 4.5, 0.2)
		}

	})
	t.Run("internalLoop", func(t *testing.T) {
		seq := "ACCCCCTCCTTCCTTGGATCAAGGGGCTCAA"
		i := 6
		j := 21

		foldContext, err := newFoldingContext(seq, 37)
		if err != nil {
			t.Fatalf("Error creating folding context: %v", err)
		}

		dg, err := internalLoop(i, i+4, j, j-4, foldContext)
		if err != nil {
			t.Fatalf("Error in internalLoop function: %v", err)
		}

		expectedDg := 3.5
		delta := 0.1
		if dg < expectedDg-delta || dg > expectedDg+delta {
			t.Errorf("internalLoop() = %v, want %v +/- %v", dg, expectedDg, delta)
		}
	})
	t.Run("W", func(t *testing.T) {
		// Test case 1
		seq := "GCUCAGCUGGGAGAGC"
		i, j := 0, 15

		foldContext, err := newFoldingContext(seq, 37)
		if err != nil {
			t.Fatalf("Error creating folding context: %v", err)
		}

		struc, err := unpairedMinimumFreeEnergyW(i, j, foldContext)
		if err != nil {
			t.Fatalf("Error in unpairedMinimumFreeEnergyW function: %v", err)
		}
		if struc.energy < -3.8-0.2 || struc.energy > -3.8+0.2 {
			t.Errorf("struc.energy = %v, want %v +/- %v", struc.energy, -3.8, 0.2)
		}

		// Test case 2
		seq = "CCUGCUUUGCACGCAGG"
		i, j = 0, 16

		foldContext, err = newFoldingContext(seq, 37)
		if err != nil {
			t.Fatalf("Error creating folding context: %v", err)
		}

		struc, err = unpairedMinimumFreeEnergyW(i, j, foldContext)
		if err != nil {
			t.Fatalf("Error in unpairedMinimumFreeEnergyW function: %v", err)
		}
		if struc.energy < -6.4-0.2 || struc.energy > -6.4+0.2 {
			t.Errorf("struc.energy = %v, want %v +/- %v", struc.energy, -6.4, 0.2)
		}

		// Test case 3
		seq = "GCGGUUCGAUCCCGC"
		i, j = 0, 14

		foldContext, err = newFoldingContext(seq, 37)
		if err != nil {
			t.Fatalf("Error creating folding context: %v", err)
		}

		struc, err = unpairedMinimumFreeEnergyW(i, j, foldContext)
		if err != nil {
			t.Fatalf("Error in unpairedMinimumFreeEnergyW function: %v", err)
		}
		if struc.energy < -4.2-0.2 || struc.energy > -4.2+0.2 {
			t.Errorf("struc.energy = %v, want %v +/- %v", struc.energy, -4.2, 0.2)
		}
	})

}
