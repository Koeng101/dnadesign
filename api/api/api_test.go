package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/koeng101/dnadesign/api/gen"
)

var app App

func TestMain(m *testing.M) {
	app = InitializeApp()
	code := m.Run()
	os.Exit(code)
}

func TestIoFastaParse(t *testing.T) {
	baseFasta := `>gi|5524211|gb|AAD44166.1| cytochrome b [Elephas maximus maximus]
LCLYTHIGRNIYYGSYLYSETWNTGIMLLLITMATAFMGYVLPWGQMSFWGATVITNLFSAIPYIGTNLV
EWIWGGFSVDKATLNRFFAFHFILPFTMVALAGVHLTFLHETGSNNPLGLTSDSDKIPFHPYYTIKDFLG
LLILILLLLLLALLSPDMLGDPDNHMPADPLNTPLHIKPEWYFLFAYAILRSVPNKLGGVLALFLSIVIL
GLMPFLHTSKHRSMMLRPLSQALFWTLTMDLLTLTWIGSQPVEYPYTIIGQMASILYFSIILAFLPIAGX
IENY
`
	req := httptest.NewRequest("POST", "/api/io/fasta/parse", strings.NewReader(baseFasta))
	resp := httptest.NewRecorder()
	app.Router.ServeHTTP(resp, req)

	r := `[{"identifier":"gi|5524211|gb|AAD44166.1| cytochrome b [Elephas maximus maximus]","sequence":"LCLYTHIGRNIYYGSYLYSETWNTGIMLLLITMATAFMGYVLPWGQMSFWGATVITNLFSAIPYIGTNLVEWIWGGFSVDKATLNRFFAFHFILPFTMVALAGVHLTFLHETGSNNPLGLTSDSDKIPFHPYYTIKDFLGLLILILLLLLLALLSPDMLGDPDNHMPADPLNTPLHIKPEWYFLFAYAILRSVPNKLGGVLALFLSIVILGLMPFLHTSKHRSMMLRPLSQALFWTLTMDLLTLTWIGSQPVEYPYTIIGQMASILYFSIILAFLPIAGXIENY"}]`
	if strings.TrimSpace(resp.Body.String()) != r {
		t.Errorf("Unexpected response. Expected: " + r + "\nGot: " + resp.Body.String())
	}
}

func TestIoGenbankParse(t *testing.T) {
	baseGenbank := `LOCUS       pUC19_lacZ         336 bp DNA     linear   UNA 12-SEP-2023
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
	req := httptest.NewRequest("POST", "/api/io/genbank/parse", strings.NewReader(baseGenbank))
	resp := httptest.NewRecorder()
	app.Router.ServeHTTP(resp, req)

	r := `[{"features":[{"attributes":{"mol_type":["genomic DNA"],"organism":["unspecified"]},"description":"","location":{"complement":false,"end":336,"fivePrimePartial":false,"gbkLocationString":"","join":false,"start":0,"subLocations":null,"threePrimePartial":false},"sequence":"caggaaacagctatgaccatgattacgccaagcttgcatgcctgcaggtcgactctagaggatccccgggtaccgagctcgaattcactggccgtcgttttacaacgtcgtgactgggaaaaccctggcgttacccaacttaatcgccttgcagcacatccccctttcgccagctggcgtaatagcgaagaggcccgcaccgatcgcccttcccaacagttgcgcagcctgaatggcgaatggcgcctgatgcggtattttctccttacgcatctgtgcggtatttcacaccgcatatggtgcactctcagtacaatctgctctgatgccgcatag","sequenceHash":"","sequenceHashFunction":"","type":"source"},{"attributes":{"label":["M13 rev"],"note":["common sequencing primer, one of multiple similarvariants"]},"description":"","location":{"complement":false,"end":17,"fivePrimePartial":false,"gbkLocationString":"","join":false,"start":0,"subLocations":null,"threePrimePartial":false},"sequence":"caggaaacagctatgac","sequenceHash":"","sequenceHashFunction":"","type":"primer_bind"},{"attributes":{"codon_start":["1"],"gene":["lacZ"],"label":["lacZ-alpha"],"product":["LacZ-alpha fragment of beta-galactosidase"],"translation":["MTMITPSLHACRSTLEDPRVPSSNSLAVVLQRRDWENPGVTQLNRLAAHPPFASWRNSEEARTDRPSQQLRSLNGEWRLMRYFLLTHLCGISHRIWCTLSTICSDAA"]},"description":"","location":{"complement":false,"end":336,"fivePrimePartial":false,"gbkLocationString":"","join":false,"start":12,"subLocations":null,"threePrimePartial":false},"sequence":"atgaccatgattacgccaagcttgcatgcctgcaggtcgactctagaggatccccgggtaccgagctcgaattcactggccgtcgttttacaacgtcgtgactgggaaaaccctggcgttacccaacttaatcgccttgcagcacatccccctttcgccagctggcgtaatagcgaagaggcccgcaccgatcgcccttcccaacagttgcgcagcctgaatggcgaatggcgcctgatgcggtattttctccttacgcatctgtgcggtatttcacaccgcatatggtgcactctcagtacaatctgctctgatgccgcatag","sequenceHash":"","sequenceHashFunction":"","type":"CDS"},{"attributes":{"label":["MCS"],"note":["pUC19 multiple cloning site"]},"description":"","location":{"complement":false,"end":86,"fivePrimePartial":false,"gbkLocationString":"","join":false,"start":29,"subLocations":null,"threePrimePartial":false},"sequence":"aagcttgcatgcctgcaggtcgactctagaggatccccgggtaccgagctcgaattc","sequenceHash":"","sequenceHashFunction":"","type":"misc_feature"},{"attributes":{"label":["M13 fwd"],"note":["common sequencing primer, one of multiple similarvariants"]},"description":"","location":{"complement":true,"end":103,"fivePrimePartial":false,"gbkLocationString":"complement(87..103)","join":false,"start":86,"subLocations":null,"threePrimePartial":false},"sequence":"gtaaaacgacggccagt","sequenceHash":"","sequenceHashFunction":"","type":"primer_bind"}],"meta":{"accession":".","baseCount":null,"date":"","definition":"natural linear DNA","keywords":".","locus":{"circular":false,"genbankDivision":"UNA","modificationDate":"12-SEP-2023","moleculeType":"DNA","name":"pUC19_lacZ","sequenceCoding":"bp","sequenceLength":"336"},"name":"","organism":"unspecified","origin":"","other":{},"references":[{"authors":"Keoni Gandall","consortium":"","journal":"Exported Sep 12, 2023 from SnapGene 6.2.2 https://www.snapgene.com","pubMed":"","range":"(bases 1 to 336)","remark":"","title":"Direct Submission"}],"sequenceHash":"","sequenceHashFunction":"","source":"natural DNA sequence","taxonomy":null,"version":"."},"sequence":"caggaaacagctatgaccatgattacgccaagcttgcatgcctgcaggtcgactctagaggatccccgggtaccgagctcgaattcactggccgtcgttttacaacgtcgtgactgggaaaaccctggcgttacccaacttaatcgccttgcagcacatccccctttcgccagctggcgtaatagcgaagaggcccgcaccgatcgcccttcccaacagttgcgcagcctgaatggcgaatggcgcctgatgcggtattttctccttacgcatctgtgcggtatttcacaccgcatatggtgcactctcagtacaatctgctctgatgccgcatag"}]`
	if strings.TrimSpace(resp.Body.String()) != r {
		t.Errorf("Unexpected response. Expected: " + r + "\nGot: " + resp.Body.String())
	}
}

func TestFix(t *testing.T) {
	req := httptest.NewRequest("POST", "/api/design/cds/fix", strings.NewReader(`{"organism":"Escherichia coli","sequence":"ATGGGTCTCTAA","removeSequences":["GGTCTC"]}`))
	resp := httptest.NewRecorder()
	app.Router.ServeHTTP(resp, req)

	r := `{"changes":[{"From":"CTC","Position":2,"Reason":"Common TypeIIS restriction enzymes - BsaI, BbsI, PaqCI","Step":0,"To":"CTG"}],"sequence":"ATGGGTCTGTAA"}`
	if strings.TrimSpace(resp.Body.String()) != r {
		t.Errorf("Unexpected response. Expected: " + r + "\nGot: " + resp.Body.String())
	}
}

func TestOptimize(t *testing.T) {
	gfp := "MASKGEELFTGVVPILVELDGDVNGHKFSVSGEGEGDATYGKLTLKFICTTGKLPVPWPTLVTTFSYGVQCFSRYPDHMKRHDFFKSAMPEGYVQERTISFKDDGNYKTRAEVKFEGDTLVNRIELKGIDFKEDGNILGHKLEYNYNSHNVYITADKQKNGIKANFKIRHNIEDGSVQLADHYQQNTPIGDGPVLLPDNHYLSTQSALSKDPNEKRDHMVLLEFVTAAGITHGMDELYK"
	req := httptest.NewRequest("POST", "/api/design/cds/optimize", strings.NewReader(fmt.Sprintf(`{"organism":"Escherichia coli","sequence":"%s"}`, gfp)))
	resp := httptest.NewRecorder()
	app.Router.ServeHTTP(resp, req)

	r := `"ATGGCATCCAAGGGCGAGGAGTTGTTCACCGGTGTTGTGCCGATCCTGGTGGAGCTGGACGGTGACGTGAACGGTCACAAATTTAGCGTGTCCGGTGAGGGTGAGGGTGATGCTACCTATGGCAAGCTGACCCTGAAATTCATTTGTACCACGGGTAAACTGCCGGTCCCGTGGCCGACGCTGGTGACCACCTTCAGCTATGGTGTGCAGTGTTTCAGCCGCTACCCGGACCACATGAAGCGCCACGACTTTTTCAAGAGCGCGATGCCGGAGGGTTATGTGCAAGAACGTACCATCAGCTTTAAAGATGATGGTAACTATAAGACCCGCGCGGAAGTCAAGTTTGAGGGTGACACGCTGGTGAATCGTATTGAGTTGAAGGGTATTGACTTTAAGGAGGATGGTAATATTTTGGGCCACAAACTGGAGTACAATTACAATAGCCACAATGTTTACATCACGGCAGATAAACAGAAGAACGGTATCAAGGCGAACTTCAAAATTCGTCACAACATTGAGGACGGTTCTGTTCAACTGGCGGACCATTACCAACAGAATACCCCGATCGGTGACGGCCCGGTTCTGCTGCCGGACAACCATTATTTGAGCACCCAGTCCGCCCTGAGCAAGGACCCGAATGAGAAGCGTGATCATATGGTTCTGCTGGAGTTTGTGACCGCGGCGGGCATCACCCACGGCATGGACGAGCTGTACAAG"`
	if strings.TrimSpace(resp.Body.String()) != r {
		t.Errorf("Unexpected response. Expected: " + r + "\nGot: " + resp.Body.String())
	}
}

func TestTranslate(t *testing.T) {
	gfp := "ATGGCATCCAAGGGCGAGGAGTTGTTCACCGGTGTTGTGCCGATCCTGGTGGAGCTGGACGGTGACGTGAACGGTCACAAATTTAGCGTGTCCGGTGAGGGTGAGGGTGATGCTACCTATGGCAAGCTGACCCTGAAATTCATTTGTACCACGGGTAAACTGCCGGTCCCGTGGCCGACGCTGGTGACCACCTTCAGCTATGGTGTGCAGTGTTTCAGCCGCTACCCGGACCACATGAAGCGCCACGACTTTTTCAAGAGCGCGATGCCGGAGGGTTATGTGCAAGAACGTACCATCAGCTTTAAAGATGATGGTAACTATAAGACCCGCGCGGAAGTCAAGTTTGAGGGTGACACGCTGGTGAATCGTATTGAGTTGAAGGGTATTGACTTTAAGGAGGATGGTAATATTTTGGGCCACAAACTGGAGTACAATTACAATAGCCACAATGTTTACATCACGGCAGATAAACAGAAGAACGGTATCAAGGCGAACTTCAAAATTCGTCACAACATTGAGGACGGTTCTGTTCAACTGGCGGACCATTACCAACAGAATACCCCGATCGGTGACGGCCCGGTTCTGCTGCCGGACAACCATTATTTGAGCACCCAGTCCGCCCTGAGCAAGGACCCGAATGAGAAGCGTGATCATATGGTTCTGCTGGAGTTTGTGACCGCGGCGGGCATCACCCACGGCATGGACGAGCTGTACAAG"
	req := httptest.NewRequest("POST", "/api/design/cds/translate", strings.NewReader(fmt.Sprintf(`{"translation_table":11,"sequence":"%s"}`, gfp)))
	resp := httptest.NewRecorder()
	app.Router.ServeHTTP(resp, req)

	r := `"MASKGEELFTGVVPILVELDGDVNGHKFSVSGEGEGDATYGKLTLKFICTTGKLPVPWPTLVTTFSYGVQCFSRYPDHMKRHDFFKSAMPEGYVQERTISFKDDGNYKTRAEVKFEGDTLVNRIELKGIDFKEDGNILGHKLEYNYNSHNVYITADKQKNGIKANFKIRHNIEDGSVQLADHYQQNTPIGDGPVLLPDNHYLSTQSALSKDPNEKRDHMVLLEFVTAAGITHGMDELYK"`
	if strings.TrimSpace(resp.Body.String()) != r {
		t.Errorf("Unexpected response. Expected: " + r + "\nGot: " + resp.Body.String())
	}
}

func TestFragment(t *testing.T) {
	lacZ := "ATGACCATGATTACGCCAAGCTTGCATGCCTGCAGGTCGACTCTAGAGGATCCCCGGGTACCGAGCTCGAATTCACTGGCCGTCGTTTTACAACGTCGTGACTGGGAAAACCCTGGCGTTACCCAACTTAATCGCCTTGCAGCACATCCCCCTTTCGCCAGCTGGCGTAATAGCGAAGAGGCCCGCACCGATCGCCCTTCCCAACAGTTGCGCAGCCTGAATGGCGAATGGCGCCTGATGCGGTATTTTCTCCTTACGCATCTGTGCGGTATTTCACACCGCATATGGTGCACTCTCAGTACAATCTGCTCTGATGCCGCATAG"
	type fragmentStruct struct {
		Sequence         string   `json:"sequence"`
		MinFragmentSize  int      `json:"min_fragment_size"`
		MaxFragmentSize  int      `json:"max_fragment_size"`
		ExcludeOverhangs []string `json:"exclude_overhangs"`
	}
	fragReq := &fragmentStruct{Sequence: lacZ, MinFragmentSize: 95, MaxFragmentSize: 105, ExcludeOverhangs: []string{"AAAA"}}
	b, err := json.Marshal(fragReq)
	if err != nil {
		t.Errorf("Failed to marshal: %s", err)
	}
	req := httptest.NewRequest("POST", "/api/simulate/fragment", bytes.NewBuffer(b))
	resp := httptest.NewRecorder()
	app.Router.ServeHTTP(resp, req)

	r := `{"efficiency":1,"fragments":["ATGACCATGATTACGCCAAGCTTGCATGCCTGCAGGTCGACTCTAGAGGATCCCCGGGTACCGAGCTCGAATTCACTGGCCGTCGTTTTACAACGTCGTGACTGG","CTGGGAAAACCCTGGCGTTACCCAACTTAATCGCCTTGCAGCACATCCCCCTTTCGCCAGCTGGCGTAATAGCGAAGAGGCCCGCACCGATCGCCCTTCCCAAC","CAACAGTTGCGCAGCCTGAATGGCGAATGGCGCCTGATGCGGTATTTTCTCCTTACGCATCTGTGC","GTGCGGTATTTCACACCGCATATGGTGCACTCTCAGTACAATCTGCTCTGATGCCGCATAG"]}`
	if strings.TrimSpace(resp.Body.String()) != r {
		t.Errorf("Unexpected response. Expected: " + r + "\nGot: " + resp.Body.String())
	}

}

func TestPCR(t *testing.T) {
	gene := "aataattacaccgagataacacatcatggataaaccgatactcaaagattctatgaagctatttgaggcacttggtacgatcaagtcgcgctcaatgtttggtggcttcggacttttcgctgatgaaacgatgtttgcactggttgtgaatgatcaacttcacatacgagcagaccagcaaacttcatctaacttcgagaagcaagggctaaaaccgtacgtttataaaaagcgtggttttccagtcgttactaagtactacgcgatttccgacgacttgtgggaatccagtgaacgcttgatagaagtagcgaagaagtcgttagaacaagccaatttggaaaaaaagcaacaggcaagtagtaagcccgacaggttgaaagacctgcctaacttacgactagcgactgaacgaatgcttaagaaagctggtataaaatcagttgaacaacttgaagagaaaggtgcattgaatgcttacaaagcgatacgtgactctcactccgcaaaagtaagtattgagctactctgggctttagaaggagcgataaacggcacgcactggagcgtcgttcctcaatctcgcagagaagagctggaaaatgcgctttcttaa"
	fwdPrimer := "TTATAGGTCTCATACTAATAATTACACCGAGATAACACATCATGG"
	revPrimer := "TATATGGTCTCTTCATTTAAGAAAGCGCATTTTCCAGC"
	primers := []string{fwdPrimer, revPrimer}
	circular := false
	complexReq := &gen.PostSimulateComplexPcrJSONBody{Circular: &circular, Primers: primers, TargetTm: 55.0, Templates: []string{gene}}
	b, err := json.Marshal(complexReq)
	if err != nil {
		t.Errorf("Failed to marshal: %s", err)
	}
	req := httptest.NewRequest("POST", "/api/simulate/complex_pcr", bytes.NewBuffer(b))
	resp := httptest.NewRecorder()
	app.Router.ServeHTTP(resp, req)
	r := `["TTATAGGTCTCATACTAATAATTACACCGAGATAACACATCATGGATAAACCGATACTCAAAGATTCTATGAAGCTATTTGAGGCACTTGGTACGATCAAGTCGCGCTCAATGTTTGGTGGCTTCGGACTTTTCGCTGATGAAACGATGTTTGCACTGGTTGTGAATGATCAACTTCACATACGAGCAGACCAGCAAACTTCATCTAACTTCGAGAAGCAAGGGCTAAAACCGTACGTTTATAAAAAGCGTGGTTTTCCAGTCGTTACTAAGTACTACGCGATTTCCGACGACTTGTGGGAATCCAGTGAACGCTTGATAGAAGTAGCGAAGAAGTCGTTAGAACAAGCCAATTTGGAAAAAAAGCAACAGGCAAGTAGTAAGCCCGACAGGTTGAAAGACCTGCCTAACTTACGACTAGCGACTGAACGAATGCTTAAGAAAGCTGGTATAAAATCAGTTGAACAACTTGAAGAGAAAGGTGCATTGAATGCTTACAAAGCGATACGTGACTCTCACTCCGCAAAAGTAAGTATTGAGCTACTCTGGGCTTTAGAAGGAGCGATAAACGGCACGCACTGGAGCGTCGTTCCTCAATCTCGCAGAGAAGAGCTGGAAAATGCGCTTTCTTAAATGAAGAGACCATATA"]`
	if strings.TrimSpace(resp.Body.String()) != r {
		t.Errorf("Unexpected response. Expected: " + r + "\nGot: " + resp.Body.String())
	}

	simpleReq := &gen.PostSimulatePcrJSONBody{Circular: &circular, TargetTm: 55.0, Template: gene, ForwardPrimer: fwdPrimer, ReversePrimer: revPrimer}
	b, err = json.Marshal(simpleReq)
	if err != nil {
		t.Errorf("Failed to marshal: %s", err)
	}
	req = httptest.NewRequest("POST", "/api/simulate/pcr", bytes.NewBuffer(b))
	resp = httptest.NewRecorder()
	app.Router.ServeHTTP(resp, req)
	r = `"TTATAGGTCTCATACTAATAATTACACCGAGATAACACATCATGGATAAACCGATACTCAAAGATTCTATGAAGCTATTTGAGGCACTTGGTACGATCAAGTCGCGCTCAATGTTTGGTGGCTTCGGACTTTTCGCTGATGAAACGATGTTTGCACTGGTTGTGAATGATCAACTTCACATACGAGCAGACCAGCAAACTTCATCTAACTTCGAGAAGCAAGGGCTAAAACCGTACGTTTATAAAAAGCGTGGTTTTCCAGTCGTTACTAAGTACTACGCGATTTCCGACGACTTGTGGGAATCCAGTGAACGCTTGATAGAAGTAGCGAAGAAGTCGTTAGAACAAGCCAATTTGGAAAAAAAGCAACAGGCAAGTAGTAAGCCCGACAGGTTGAAAGACCTGCCTAACTTACGACTAGCGACTGAACGAATGCTTAAGAAAGCTGGTATAAAATCAGTTGAACAACTTGAAGAGAAAGGTGCATTGAATGCTTACAAAGCGATACGTGACTCTCACTCCGCAAAAGTAAGTATTGAGCTACTCTGGGCTTTAGAAGGAGCGATAAACGGCACGCACTGGAGCGTCGTTCCTCAATCTCGCAGAGAAGAGCTGGAAAATGCGCTTTCTTAAATGAAGAGACCATATA"`
	if strings.TrimSpace(resp.Body.String()) != r {
		t.Errorf("Unexpected response. Expected: " + r + "\nGot: " + resp.Body.String())
	}
}
