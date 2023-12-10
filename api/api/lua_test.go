package api

import (
	"testing"

	"github.com/koeng101/dnadesign/api/gen"
)

func TestApp_LuaIoFastaParse(t *testing.T) {
	luaScript := `
parsed_fasta = fasta_parse(attachments["input.fasta"])

output = parsed_fasta[1].name
`
	inputFasta := `>AAD44166.1
LCLYTHIGRNIYYGSYLYSETWNTGIMLLLITMATAFMGYVLPWGQMSFWGATVITNLFSAIPYIGTNLV
EWIWGGFSVDKATLNRFFAFHFILPFTMVALAGVHLTFLHETGSNNPLGLTSDSDKIPFHPYYTIKDFLG
LLILILLLLLLALLSPDMLGDPDNHMPADPLNTPLHIKPEWYFLFAYAILRSVPNKLGGVLALFLSIVIL
GLMPFLHTSKHRSMMLRPLSQALFWTLTMDLLTLTWIGSQPVEYPYTIIGQMASILYFSIILAFLPIAGX
IENY
`

	fastaAttachment := gen.Attachment{
		Name:    "input.fasta",
		Content: inputFasta,
	}

	_, output, err := app.ExecuteLua(luaScript, []gen.Attachment{fastaAttachment})
	if err != nil {
		t.Errorf("No error should be found. Got err: %s", err)
	}
	expectedOutput := "AAD44166.1"
	if output != expectedOutput {
		t.Errorf("Unexpected response. Expected: " + expectedOutput + "\nGot: " + output)
	}
}

func TestApp_LuaDesignCdsFix(t *testing.T) {
	luaScript := `
fixed = fix("ATGGGTCTCTAA", "Escherichia coli", {"GGTCTC"})

output = fixed.sequence
`
	_, output, err := app.ExecuteLua(luaScript, []gen.Attachment{})
	if err != nil {
		t.Errorf("No error should be found. Got err: %s", err)
	}
	expectedOutput := "ATGGGTCTGTAA"
	if output != expectedOutput {
		t.Errorf("Unexpected response. Expected: " + expectedOutput + "\nGot: " + output)
	}
}

func TestApp_LuaDesignCdsOptimizeTranslate(t *testing.T) {
	luaScript := `
optimized = optimize("MHELLQWQRLD", "Escherichia coli", 0)
translated = translate(optimized, 11)

output = optimized
log = translated
`
	log, output, err := app.ExecuteLua(luaScript, []gen.Attachment{})
	if err != nil {
		t.Errorf("No error should be found. Got err: %s", err)
	}
	expectedOutput := "ATGCATGAGCTGTTGCAGTGGCAGCGCTTGGAC"
	if output != expectedOutput {
		t.Errorf("Unexpected response. Expected: " + expectedOutput + "\nGot: " + output)
	}
	expectedLog := "MHELLQWQRLD"
	if log != expectedLog {
		t.Errorf("Unexpected response. Expected: " + expectedLog + "\nGot: " + log)
	}
}

func TestApp_LuaSimulateFragment(t *testing.T) {
	luaScript := `
lacZ = "ATGACCATGATTACGCCAAGCTTGCATGCCTGCAGGTCGACTCTAGAGGATCCCCGGGTACCGAGCTCGAATTCACTGGCCGTCGTTTTACAACGTCGTGACTGGGAAAACCCTGGCGTTACCCAACTTAATCGCCTTGCAGCACATCCCCCTTTCGCCAGCTGGCGTAATAGCGAAGAGGCCCGCACCGATCGCCCTTCCCAACAGTTGCGCAGCCTGAATGGCGAATGGCGCCTGATGCGGTATTTTCTCCTTACGCATCTGTGCGGTATTTCACACCGCATATGGTGCACTCTCAGTACAATCTGCTCTGATGCCGCATAG"
fragmentation = fragment(lacZ, 95, 105)

output = fragmentation.fragments[1]
`
	_, output, err := app.ExecuteLua(luaScript, []gen.Attachment{})
	if err != nil {
		t.Errorf("No error should be found. Got err: %s", err)
	}
	expectedOutput := "ATGACCATGATTACGCCAAGCTTGCATGCCTGCAGGTCGACTCTAGAGGATCCCCGGGTACCGAGCTCGAATTCACTGGCCGTCGTTTTACAACGTCGTGACTGG"
	if output != expectedOutput {
		t.Errorf("Unexpected response. Expected: " + expectedOutput + "\nGot: " + output)
	}
}

func TestApp_LuaSimulatePCR(t *testing.T) {
	luaScript := `
gene = "aataattacaccgagataacacatcatggataaaccgatactcaaagattctatgaagctatttgaggcacttggtacgatcaagtcgcgctcaatgtttggtggcttcggacttttcgctgatgaaacgatgtttgcactggttgtgaatgatcaacttcacatacgagcagaccagcaaacttcatctaacttcgagaagcaagggctaaaaccgtacgtttataaaaagcgtggttttccagtcgttactaagtactacgcgatttccgacgacttgtgggaatccagtgaacgcttgatagaagtagcgaagaagtcgttagaacaagccaatttggaaaaaaagcaacaggcaagtagtaagcccgacaggttgaaagacctgcctaacttacgactagcgactgaacgaatgcttaagaaagctggtataaaatcagttgaacaacttgaagagaaaggtgcattgaatgcttacaaagcgatacgtgactctcactccgcaaaagtaagtattgagctactctgggctttagaaggagcgataaacggcacgcactggagcgtcgttcctcaatctcgcagagaagagctggaaaatgcgctttcttaa"
fwd_primer = "TTATAGGTCTCATACTAATAATTACACCGAGATAACACATCATGG"
rev_primer = "TATATGGTCTCTTCATTTAAGAAAGCGCATTTTCCAGC"
amplicons = complex_pcr({gene},{fwd_primer, rev_primer}, 55.0)

output = amplicons[1]
`
	_, output, err := app.ExecuteLua(luaScript, []gen.Attachment{})
	if err != nil {
		t.Errorf("No error should be found. Got err: %s", err)
	}
	expectedOutput := "TTATAGGTCTCATACTAATAATTACACCGAGATAACACATCATGGATAAACCGATACTCAAAGATTCTATGAAGCTATTTGAGGCACTTGGTACGATCAAGTCGCGCTCAATGTTTGGTGGCTTCGGACTTTTCGCTGATGAAACGATGTTTGCACTGGTTGTGAATGATCAACTTCACATACGAGCAGACCAGCAAACTTCATCTAACTTCGAGAAGCAAGGGCTAAAACCGTACGTTTATAAAAAGCGTGGTTTTCCAGTCGTTACTAAGTACTACGCGATTTCCGACGACTTGTGGGAATCCAGTGAACGCTTGATAGAAGTAGCGAAGAAGTCGTTAGAACAAGCCAATTTGGAAAAAAAGCAACAGGCAAGTAGTAAGCCCGACAGGTTGAAAGACCTGCCTAACTTACGACTAGCGACTGAACGAATGCTTAAGAAAGCTGGTATAAAATCAGTTGAACAACTTGAAGAGAAAGGTGCATTGAATGCTTACAAAGCGATACGTGACTCTCACTCCGCAAAAGTAAGTATTGAGCTACTCTGGGCTTTAGAAGGAGCGATAAACGGCACGCACTGGAGCGTCGTTCCTCAATCTCGCAGAGAAGAGCTGGAAAATGCGCTTTCTTAAATGAAGAGACCATATA"
	if output != expectedOutput {
		t.Errorf("Unexpected response. Expected: " + expectedOutput + "\nGot: " + output)
	}

	luaScript = `
gene = "aataattacaccgagataacacatcatggataaaccgatactcaaagattctatgaagctatttgaggcacttggtacgatcaagtcgcgctcaatgtttggtggcttcggacttttcgctgatgaaacgatgtttgcactggttgtgaatgatcaacttcacatacgagcagaccagcaaacttcatctaacttcgagaagcaagggctaaaaccgtacgtttataaaaagcgtggttttccagtcgttactaagtactacgcgatttccgacgacttgtgggaatccagtgaacgcttgatagaagtagcgaagaagtcgttagaacaagccaatttggaaaaaaagcaacaggcaagtagtaagcccgacaggttgaaagacctgcctaacttacgactagcgactgaacgaatgcttaagaaagctggtataaaatcagttgaacaacttgaagagaaaggtgcattgaatgcttacaaagcgatacgtgactctcactccgcaaaagtaagtattgagctactctgggctttagaaggagcgataaacggcacgcactggagcgtcgttcctcaatctcgcagagaagagctggaaaatgcgctttcttaa"
fwd_primer = "TTATAGGTCTCATACTAATAATTACACCGAGATAACACATCATGG"
rev_primer = "TATATGGTCTCTTCATTTAAGAAAGCGCATTTTCCAGC"
amplicon = pcr(gene, fwd_primer, rev_primer, 55.0)

output = amplicon
`
	_, output, err = app.ExecuteLua(luaScript, []gen.Attachment{})
	if err != nil {
		t.Errorf("No error should be found. Got err: %s", err)
	}
	if output != expectedOutput {
		t.Errorf("Unexpected response. Expected: " + expectedOutput + "\nGot: " + output)
	}
}
