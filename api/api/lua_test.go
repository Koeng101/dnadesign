package api

import (
	"testing"

	"github.com/koeng101/dnadesign/api/gen"
)

func TestApp_LuaIoFastaParse(t *testing.T) {
	luaScript := `
parsed_fasta = fasta_parse(attachments["input.fasta"])

output = parsed_fasta[1].identifier
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

func TestApp_LuaIoGenbankParse(t *testing.T) {
	luaScript := `
parsed_genbank = genbank_parse(attachments["input.gb"])

output = parsed_genbank[1].features[3].attributes["translation"][1]
`
	inputGenbank := `LOCUS       pUC19_lacZ         336 bp DNA     linear   UNA 12-SEP-2023
DEFINITION  natural linear DNA
ACCESSION   .
VERSION     .
KEYWORDS    .
SOURCE      natural DNA sequence
  ORGANISM  unspecified
REFERENCE   1  (bases 1 to 336)
  AUTHORS   Keoni Gandall
  TITLE     Direct Submission
  JOURNAL   Exported Sep 12, 2023 from SnapGene 6.2.2
            https://www.snapgene.com
FEATURES             Location/Qualifiers
     source          1..336
                     /mol_type="genomic DNA"
                     /organism="unspecified"
     primer_bind     1..17
                     /label=M13 rev
                     /note="common sequencing primer, one of multiple similar
                     variants"
     CDS             13..336
                     /codon_start=1
                     /gene="lacZ"
                     /product="LacZ-alpha fragment of beta-galactosidase"
                     /label=lacZ-alpha
                     /translation="MTMITPSLHACRSTLEDPRVPSSNSLAVVLQRRDWENPGVTQLNR
                     LAAHPPFASWRNSEEARTDRPSQQLRSLNGEWRLMRYFLLTHLCGISHRIWCTLSTICS
                     DAA"
     misc_feature    30..86
                     /label=MCS
                     /note="pUC19 multiple cloning site"
     primer_bind     complement(87..103)
                     /label=M13 fwd
                     /note="common sequencing primer, one of multiple similar
                     variants"
ORIGIN
        1 caggaaacag ctatgaccat gattacgcca agcttgcatg cctgcaggtc gactctagag
       61 gatccccggg taccgagctc gaattcactg gccgtcgttt tacaacgtcg tgactgggaa
      121 aaccctggcg ttacccaact taatcgcctt gcagcacatc cccctttcgc cagctggcgt
      181 aatagcgaag aggcccgcac cgatcgccct tcccaacagt tgcgcagcct gaatggcgaa
      241 tggcgcctga tgcggtattt tctccttacg catctgtgcg gtatttcaca ccgcatatgg
      301 tgcactctca gtacaatctg ctctgatgcc gcatag
//
`
	genbankAttachment := gen.Attachment{
		Name:    "input.gb",
		Content: inputGenbank,
	}

	_, output, err := app.ExecuteLua(luaScript, []gen.Attachment{genbankAttachment})
	if err != nil {
		t.Errorf("No error should be found. Got err: %s", err)
	}
	expectedOutput := "MTMITPSLHACRSTLEDPRVPSSNSLAVVLQRRDWENPGVTQLNRLAAHPPFASWRNSEEARTDRPSQQLRSLNGEWRLMRYFLLTHLCGISHRIWCTLSTICSDAA"
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
