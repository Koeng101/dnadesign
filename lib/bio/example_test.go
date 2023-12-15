package bio_test

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/koeng101/dnadesign/lib/bio"
	"github.com/koeng101/dnadesign/lib/bio/fasta"
	"github.com/koeng101/dnadesign/lib/bio/uniprot"
)

// Example_read shows an example of reading a file from disk.
func Example_read() {
	// Read lets you read files from disk into a parser.
	file, _ := os.Open("fasta/data/base.fasta")
	parser := bio.NewFastaParser(file)

	records, _ := parser.Parse()

	fmt.Println(records[1].Sequence)
	// Output: ADQLTEEQIAEFKEAFSLFDKDGDGTITTKELGTVMRSLGQNPTEAELQDMINEVDADGNGTIDFPEFLTMMARKMKDTDSEEEIREAFRVFDKDGNGYISAAELRHVMTNLGEKLTDEEVDEMIREADIDGDGQVNYEEFVQMMTAK*
}

// Example_readGz shows an example of reading and parsing a gzipped file.
func Example_readGz() {
	fileGz, _ := os.Open("fasta/data/base.fasta.gz")
	file, _ := gzip.NewReader(fileGz)
	parser := bio.NewFastaParser(file)
	records, _ := parser.Parse()

	fmt.Println(records[1].Sequence)
	// Output: ADQLTEEQIAEFKEAFSLFDKDGDGTITTKELGTVMRSLGQNPTEAELQDMINEVDADGNGTIDFPEFLTMMARKMKDTDSEEEIREAFRVFDKDGNGYISAAELRHVMTNLGEKLTDEEVDEMIREADIDGDGQVNYEEFVQMMTAK*
}

func Example_newParserGz() {
	// First, lets make a file that is gzip'd, represented by this
	// buffer.
	var file bytes.Buffer
	zipWriter := gzip.NewWriter(&file)
	_, _ = zipWriter.Write([]byte(`>gi|5524211|gb|AAD44166.1| cytochrome b [Elephas maximus maximus]
LCLYTHIGRNIYYGSYLYSETWNTGIMLLLITMATAFMGYVLPWGQMSFWGATVITNLFSAIPYIGTNLV
EWIWGGFSVDKATLNRFFAFHFILPFTMVALAGVHLTFLHETGSNNPLGLTSDSDKIPFHPYYTIKDFLG
LLILILLLLLLALLSPDMLGDPDNHMPADPLNTPLHIKPEWYFLFAYAILRSVPNKLGGVLALFLSIVIL
GLMPFLHTSKHRSMMLRPLSQALFWTLTMDLLTLTWIGSQPVEYPYTIIGQMASILYFSIILAFLPIAGX
IENY

>MCHU - Calmodulin - Human, rabbit, bovine, rat, and chicken
ADQLTEEQIAEFKEAFSLFDKDGDGTITTKELGTVMRSLGQNPTEAELQDMINEVDADGNGTID
FPEFLTMMARKMKDTDSEEEIREAFRVFDKDGNGYISAAELRHVMTNLGEKLTDEEVDEMIREA
DIDGDGQVNYEEFVQMMTAK*`))
	zipWriter.Close()

	fileDecompressed, _ := gzip.NewReader(&file) // Decompress the file
	parser := bio.NewFastaParser(fileDecompressed)
	records, _ := parser.Parse() // Parse all data records from file

	fmt.Println(records[1].Sequence)
	// Output: ADQLTEEQIAEFKEAFSLFDKDGDGTITTKELGTVMRSLGQNPTEAELQDMINEVDADGNGTIDFPEFLTMMARKMKDTDSEEEIREAFRVFDKDGNGYISAAELRHVMTNLGEKLTDEEVDEMIREADIDGDGQVNYEEFVQMMTAK*
}

func ExampleParser_ParseWithHeader() {
	// The following can be replaced with a any io.Reader. For example,
	// `file, err := os.Open(path)` for file would also work.
	file := strings.NewReader(`#slow5_version	0.2.0
#num_read_groups	1
@asic_id	4175987214
#char*	uint32_t	double	double	double	double	uint64_t	int16_t*	uint64_t	int32_t	uint8_t	double	enum{unknown,partial,mux_change,unblock_mux_change,data_service_unblock_mux_change,signal_positive,signal_negative}	char*
#read_id	read_group	digitisation	offset	range	sampling_rate	len_raw_signal	raw_signal	start_time	read_number	start_mux	median_before	end_reason	channel_number
0026631e-33a3-49ab-aa22-3ab157d71f8b	0	8192	16	1489.52832	4000	5347	430,472,463	8318394	5383	1	219.133423	5	10
`)
	parser, _ := bio.NewSlow5Parser(file)
	reads, header, _ := parser.ParseWithHeader() // Parse all data records from file

	fmt.Printf("%s, %s\n", header.HeaderValues[0].Slow5Version, reads[0].ReadID)
	// Output: 0.2.0, 0026631e-33a3-49ab-aa22-3ab157d71f8b
}

func ExampleParser_ParseToChannel() {
	// The following can be replaced with a any io.Reader. For example,
	// `file, err := os.Open(path)` for file would also work.
	file := strings.NewReader(`>gi|5524211|gb|AAD44166.1| cytochrome b [Elephas maximus maximus]
LCLYTHIGRNIYYGSYLYSETWNTGIMLLLITMATAFMGYVLPWGQMSFWGATVITNLFSAIPYIGTNLV
EWIWGGFSVDKATLNRFFAFHFILPFTMVALAGVHLTFLHETGSNNPLGLTSDSDKIPFHPYYTIKDFLG
LLILILLLLLLALLSPDMLGDPDNHMPADPLNTPLHIKPEWYFLFAYAILRSVPNKLGGVLALFLSIVIL
GLMPFLHTSKHRSMMLRPLSQALFWTLTMDLLTLTWIGSQPVEYPYTIIGQMASILYFSIILAFLPIAGX
IENY

>MCHU - Calmodulin - Human, rabbit, bovine, rat, and chicken
ADQLTEEQIAEFKEAFSLFDKDGDGTITTKELGTVMRSLGQNPTEAELQDMINEVDADGNGTID
FPEFLTMMARKMKDTDSEEEIREAFRVFDKDGNGYISAAELRHVMTNLGEKLTDEEVDEMIREA
DIDGDGQVNYEEFVQMMTAK*`)
	parser := bio.NewFastaParser(file)

	channel := make(chan *fasta.Record)
	ctx := context.Background()
	go func() { _ = parser.ParseToChannel(ctx, channel, false) }()

	var records []*fasta.Record
	for record := range channel {
		records = append(records, record)
	}

	fmt.Println(records[1].Sequence)
	// Output: ADQLTEEQIAEFKEAFSLFDKDGDGTITTKELGTVMRSLGQNPTEAELQDMINEVDADGNGTIDFPEFLTMMARKMKDTDSEEEIREAFRVFDKDGNGYISAAELRHVMTNLGEKLTDEEVDEMIREADIDGDGQVNYEEFVQMMTAK*
}

func ExampleManyToChannel() {
	file1 := strings.NewReader(`>gi|5524211|gb|AAD44166.1| cytochrome b [Elephas maximus maximus]
LCLYTHIGRNIYYGSYLYSETWNTGIMLLLITMATAFMGYVLPWGQMSFWGATVITNLFSAIPYIGTNLV
EWIWGGFSVDKATLNRFFAFHFILPFTMVALAGVHLTFLHETGSNNPLGLTSDSDKIPFHPYYTIKDFLG
LLILILLLLLLALLSPDMLGDPDNHMPADPLNTPLHIKPEWYFLFAYAILRSVPNKLGGVLALFLSIVIL
GLMPFLHTSKHRSMMLRPLSQALFWTLTMDLLTLTWIGSQPVEYPYTIIGQMASILYFSIILAFLPIAGX
IENY
`)
	file2 := strings.NewReader(`>MCHU - Calmodulin - Human, rabbit, bovine, rat, and chicken
ADQLTEEQIAEFKEAFSLFDKDGDGTITTKELGTVMRSLGQNPTEAELQDMINEVDADGNGTID
FPEFLTMMARKMKDTDSEEEIREAFRVFDKDGNGYISAAELRHVMTNLGEKLTDEEVDEMIREA
DIDGDGQVNYEEFVQMMTAK*`)
	parser1 := bio.NewFastaParser(file1)
	parser2 := bio.NewFastaParser(file2)

	channel := make(chan *fasta.Record)
	ctx := context.Background()
	go func() { _ = bio.ManyToChannel(ctx, channel, parser1, parser2) }()

	var records []*fasta.Record
	for record := range channel {
		records = append(records, record)
	}

	fmt.Println(len(records)) // Records come out in a stochastic order, so we just make sure there are 2
	// Output: 2
}

func Example_writeAll() {
	// The following can be replaced with a any io.Reader. For example,
	// `file, err := os.Open(path)` for file would also work.
	file := strings.NewReader(`#slow5_version	0.2.0
#num_read_groups	1
@asic_id	4175987214
#char*	uint32_t	double	double	double	double	uint64_t	int16_t*	uint64_t	int32_t	uint8_t	double	enum{unknown,partial,mux_change,unblock_mux_change,data_service_unblock_mux_change,signal_positive,signal_negative}	char*
#read_id	read_group	digitisation	offset	range	sampling_rate	len_raw_signal	raw_signal	start_time	read_number	start_mux	median_before	end_reason	channel_number
0026631e-33a3-49ab-aa22-3ab157d71f8b	0	8192	16	1489.52832	4000	5347	430,472,463	8318394	5383	1	219.133423	5	10
`)
	parser, _ := bio.NewSlow5Parser(file)
	reads, header, _ := parser.ParseWithHeader() // Parse all data records from file

	// Write the files to an io.Writer.
	// All headers and all records implement io.WriterTo interfaces.
	var buffer bytes.Buffer
	_, _ = header.WriteTo(&buffer)
	for _, read := range reads {
		_, _ = read.WriteTo(&buffer)
	}

	fmt.Println(buffer.String())
	// Output: #slow5_version	0.2.0
	//#num_read_groups	1
	//@asic_id	4175987214
	//#char*	uint32_t	double	double	double	double	uint64_t	int16_t*	uint64_t	int32_t	uint8_t	double	enum{unknown,partial,mux_change,unblock_mux_change,data_service_unblock_mux_change,signal_positive,signal_negative}	char*
	//#read_id	read_group	digitisation	offset	range	sampling_rate	len_raw_signal	raw_signal	start_time	read_number	start_mux	median_before	end_reason	channel_number
	//0026631e-33a3-49ab-aa22-3ab157d71f8b	0	8192	16	1489.52832	4000	5347	430,472,463	8318394	5383	1	219.133423	5	10
	//
	//
}

func ExampleNewFastaParser() {
	// The following can be replaced with a any io.Reader. For example,
	// `file, err := os.Open(path)` for file would also work.
	file := strings.NewReader(`>gi|5524211|gb|AAD44166.1| cytochrome b [Elephas maximus maximus]
LCLYTHIGRNIYYGSYLYSETWNTGIMLLLITMATAFMGYVLPWGQMSFWGATVITNLFSAIPYIGTNLV
EWIWGGFSVDKATLNRFFAFHFILPFTMVALAGVHLTFLHETGSNNPLGLTSDSDKIPFHPYYTIKDFLG
LLILILLLLLLALLSPDMLGDPDNHMPADPLNTPLHIKPEWYFLFAYAILRSVPNKLGGVLALFLSIVIL
GLMPFLHTSKHRSMMLRPLSQALFWTLTMDLLTLTWIGSQPVEYPYTIIGQMASILYFSIILAFLPIAGX
IENY

>MCHU - Calmodulin - Human, rabbit, bovine, rat, and chicken
ADQLTEEQIAEFKEAFSLFDKDGDGTITTKELGTVMRSLGQNPTEAELQDMINEVDADGNGTID
FPEFLTMMARKMKDTDSEEEIREAFRVFDKDGNGYISAAELRHVMTNLGEKLTDEEVDEMIREA
DIDGDGQVNYEEFVQMMTAK*`)
	parser := bio.NewFastaParser(file)
	records, _ := parser.Parse() // Parse all data records from file

	fmt.Println(records[1].Sequence)
	// Output: ADQLTEEQIAEFKEAFSLFDKDGDGTITTKELGTVMRSLGQNPTEAELQDMINEVDADGNGTIDFPEFLTMMARKMKDTDSEEEIREAFRVFDKDGNGYISAAELRHVMTNLGEKLTDEEVDEMIREADIDGDGQVNYEEFVQMMTAK*
}

func ExampleNewFastqParser() {
	// The following can be replaced with a any io.Reader. For example,
	// `file, err := os.Open(path)` for file would also work.
	file := strings.NewReader(`@e3cc70d5-90ef-49b6-bbe1-cfef99537d73 runid=99790f25859e24307203c25273f3a8be8283e7eb read=13956 ch=53 start_time=2020-11-11T01:49:01Z flow_cell_id=AEI083 protocol_group_id=NanoSav2 sample_id=nanosavseq2
GATGTGCGCCGTTCCAGTTGCGACGTACTATAATCCCCGGCAACACGGTGCTGATTCTCTTCCTGTTCCAGAAAGCATAAACAGATGCAAGTCTGGTGTGATTAACTTCACCAAAGGGCTGGTTGTAATATTAGGAAATCTAACAATAGATTCTGTTGGTTGGACTCTAAAATTAGAAATTTGATAGATTCCTTTTCCCAAATGAAAGTTTAACGTACACTTTGTTTCTAAAGGAAGGTCAAATTACAGTCTACAGCATCGTAATGGTTCATTTTCATTTATATTTTAATACTAGAAAAGTCCTAGGTTGAAGATAACCACATAATAAGCTGCAACTTCAGCTGTCCCAACCTGAAGAAGAATCGCAGGAGTCGAAATAACTTCTGTAAAGCAAGTAGTTTGAACCTATTGATGTTTCAACATGAGCAATACGTAACT
+
$$&%&%#$)*59;/767C378411,***,('11<;:,0039/0&()&'2(/*((4.1.09751).601+'#&&&,-**/0-+3558,/)+&)'&&%&$$'%'%'&*/5978<9;**'3*'&&A?99:;:97:278?=9B?CLJHGG=9<@AC@@=>?=>D>=3<>=>3362$%/((+/%&+//.-,%-4:+..000,&$#%$$%+*)&*0%.//*?<<;>DE>.8942&&//074&$033)*&&&%**)%)962133-%'&*99><<=1144??6.027639.011/-)($#$(/422*4;:=122>?@6964:.5'8:52)*675=:4@;323&&##'.-57*4597)+0&:7<7-550REGB21/0+*79/&/6538())+)+23665+(''$$$'-2(&&*-.-#$&%%$$,-)&$$#$'&,);;<C<@454)#'`) // This is a real sequencing output, btw
	parser := bio.NewFastqParser(file)
	records, _ := parser.Parse() // Parse all data records from file

	fmt.Println(records[0].Sequence)
	// Output:GATGTGCGCCGTTCCAGTTGCGACGTACTATAATCCCCGGCAACACGGTGCTGATTCTCTTCCTGTTCCAGAAAGCATAAACAGATGCAAGTCTGGTGTGATTAACTTCACCAAAGGGCTGGTTGTAATATTAGGAAATCTAACAATAGATTCTGTTGGTTGGACTCTAAAATTAGAAATTTGATAGATTCCTTTTCCCAAATGAAAGTTTAACGTACACTTTGTTTCTAAAGGAAGGTCAAATTACAGTCTACAGCATCGTAATGGTTCATTTTCATTTATATTTTAATACTAGAAAAGTCCTAGGTTGAAGATAACCACATAATAAGCTGCAACTTCAGCTGTCCCAACCTGAAGAAGAATCGCAGGAGTCGAAATAACTTCTGTAAAGCAAGTAGTTTGAACCTATTGATGTTTCAACATGAGCAATACGTAACT
}

func ExampleNewGenbankParser() {
	// The following can be replaced with a any io.Reader. For example,
	// `file, err := os.Open(path)` for file would also work.
	file := strings.NewReader(`LOCUS       pUC19_lacZ         336 bp DNA     linear   UNA 12-SEP-2023
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
`)
	parser := bio.NewGenbankParser(file)
	records, _ := parser.Parse()

	fmt.Println(records[0].Features[2].Attributes["translation"])
	// Output: [MTMITPSLHACRSTLEDPRVPSSNSLAVVLQRRDWENPGVTQLNRLAAHPPFASWRNSEEARTDRPSQQLRSLNGEWRLMRYFLLTHLCGISHRIWCTLSTICSDAA]
}

func ExampleNewSlow5Parser() {
	// The following can be replaced with a any io.Reader. For example,
	// `file, err := os.Open(path)` for file would also work.
	file := strings.NewReader(`#slow5_version	0.2.0
#num_read_groups	1
@asic_id	4175987214
#char*	uint32_t	double	double	double	double	uint64_t	int16_t*	uint64_t	int32_t	uint8_t	double	enum{unknown,partial,mux_change,unblock_mux_change,data_service_unblock_mux_change,signal_positive,signal_negative}	char*
#read_id	read_group	digitisation	offset	range	sampling_rate	len_raw_signal	raw_signal	start_time	read_number	start_mux	median_before	end_reason	channel_number
0026631e-33a3-49ab-aa22-3ab157d71f8b	0	8192	16	1489.52832	4000	5347	430,472,463	8318394	5383	1	219.133423	5	10
`)
	parser, _ := bio.NewSlow5Parser(file)
	reads, _ := parser.Parse() // Parse all data records from file

	fmt.Println(reads[0].RawSignal)
	// Output: [430 472 463]
}

func ExampleNewPileupParser() {
	file := strings.NewReader(`seq1 	272 	T 	24 	,.$.....,,.,.,...,,,.,..^+. 	<<<+;<<<<<<<<<<<=<;<;7<&
seq1 	273 	T 	23 	,.....,,.,.,...,,,.,..A 	<<<;<<<<<<<<<3<=<<<;<<+
seq1 	274 	T 	23 	,.$....,,.,.,...,,,.,... 	7<7;<;<<<<<<<<<=<;<;<<6
seq1 	275 	A 	23 	,$....,,.,.,...,,,.,...^l. 	<+;9*<<<<<<<<<=<<:;<<<<
seq1 	276 	G 	22 	...T,,.,.,...,,,.,.... 	33;+<<7=7<<7<&<<1;<<6<
seq1 	277 	T 	22 	....,,.,.,.C.,,,.,..G. 	+7<;<<<<<<<&<=<<:;<<&<
seq1 	278 	G 	23 	....,,.,.,...,,,.,....^k. 	%38*<<;<7<<7<=<<<;<<<<<
seq1 	279 	C 	23 	A..T,,.,.,...,,,.,..... 	75&<<<<<<<<<=<<<9<<:<<<`)
	parser := bio.NewPileupParser(file)
	lines, _ := parser.Parse() // Parse all lines from file

	fmt.Println(lines[1].Quality)
	// Output: <<<;<<<<<<<<<3<=<<<;<<+
}

func ExampleNewUniprotParser() {
	// The following is a real entry in Swiss-Prot. We're going to gzip it and
	// put the gzipped text as an io.Reader to mock a file. You can edit the
	// text here to see how the parser works.
	uniprotEntryText := `<entry dataset="Swiss-Prot" created="2009-05-05" modified="2020-08-12" version="9" xmlns="http://uniprot.org/uniprot">
  <accession>P0C9F0</accession>
  <name>1001R_ASFK5</name>
  <protein>
    <recommendedName>
      <fullName>Protein MGF 100-1R</fullName>
    </recommendedName>
  </protein>
  <gene>
    <name type="ordered locus">Ken-018</name>
  </gene>
  <organism>
    <name type="scientific">African swine fever virus (isolate Pig/Kenya/KEN-50/1950)</name>
    <name type="common">ASFV</name>
    <dbReference type="NCBI Taxonomy" id="561445"/>
    <lineage>
      <taxon>Viruses</taxon>
      <taxon>Varidnaviria</taxon>
      <taxon>Bamfordvirae</taxon>
      <taxon>Nucleocytoviricota</taxon>
      <taxon>Pokkesviricetes</taxon>
      <taxon>Asfuvirales</taxon>
      <taxon>Asfarviridae</taxon>
      <taxon>Asfivirus</taxon>
    </lineage>
  </organism>
  <organismHost>
    <name type="scientific">Ornithodoros</name>
    <name type="common">relapsing fever ticks</name>
    <dbReference type="NCBI Taxonomy" id="6937"/>
  </organismHost>
  <organismHost>
    <name type="scientific">Phacochoerus aethiopicus</name>
    <name type="common">Warthog</name>
    <dbReference type="NCBI Taxonomy" id="85517"/>
  </organismHost>
  <organismHost>
    <name type="scientific">Phacochoerus africanus</name>
    <name type="common">Warthog</name>
    <dbReference type="NCBI Taxonomy" id="41426"/>
  </organismHost>
  <organismHost>
    <name type="scientific">Potamochoerus larvatus</name>
    <name type="common">Bushpig</name>
    <dbReference type="NCBI Taxonomy" id="273792"/>
  </organismHost>
  <organismHost>
    <name type="scientific">Sus scrofa</name>
    <name type="common">Pig</name>
    <dbReference type="NCBI Taxonomy" id="9823"/>
  </organismHost>
  <reference key="1">
    <citation type="submission" date="2003-03" db="EMBL/GenBank/DDBJ databases">
      <title>African swine fever virus genomes.</title>
      <authorList>
        <person name="Kutish G.F."/>
        <person name="Rock D.L."/>
      </authorList>
    </citation>
    <scope>NUCLEOTIDE SEQUENCE [LARGE SCALE GENOMIC DNA]</scope>
  </reference>
  <comment type="function">
    <text evidence="1">Plays a role in virus cell tropism, and may be required for efficient virus replication in macrophages.</text>
  </comment>
  <comment type="similarity">
    <text evidence="2">Belongs to the asfivirus MGF 100 family.</text>
  </comment>
  <dbReference type="EMBL" id="AY261360">
    <property type="status" value="NOT_ANNOTATED_CDS"/>
    <property type="molecule type" value="Genomic_DNA"/>
  </dbReference>
  <dbReference type="Proteomes" id="UP000000861">
    <property type="component" value="Genome"/>
  </dbReference>
  <proteinExistence type="inferred from homology"/>
  <feature type="chain" id="PRO_0000373170" description="Protein MGF 100-1R">
    <location>
      <begin position="1"/>
      <end position="122"/>
    </location>
  </feature>
  <evidence type="ECO:0000250" key="1"/>
  <evidence type="ECO:0000305" key="2"/>
  <sequence length="122" mass="14969" checksum="C5E63C34B941711C" modified="2009-05-05" version="1">MVRLFYNPIKYLFYRRSCKKRLRKALKKLNFYHPPKECCQIYRLLENAPGGTYFITENMTNELIMIAKDPVDKKIKSVKLYLTGNYIKINQHYYINIYMYLMRYNQIYKYPLICFSKYSKIL</sequence>
</entry>`
	// Encode the string into an gzip io.Reader
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, _ = gz.Write([]byte(uniprotEntryText))
	_ = gz.Close()

	r := bytes.NewReader(buf.Bytes())

	// Now we load the parser, and get the first entry out.
	parser, _ := uniprot.NewParser(r)
	entry, _ := parser.Next()

	fmt.Println(entry.Accession[0])
	// Output: P0C9F0
}
