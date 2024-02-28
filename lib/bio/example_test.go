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
	"github.com/koeng101/dnadesign/lib/bio/fastq"
	"github.com/koeng101/dnadesign/lib/bio/sam"
	"golang.org/x/sync/errgroup"
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

	channel := make(chan fasta.Record)
	ctx := context.Background()
	go func() { _ = parser.ParseToChannel(ctx, channel, false) }()

	var records []fasta.Record
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

	channel := make(chan fasta.Record)
	ctx := context.Background()
	go func() { _ = bio.ManyToChannel(ctx, channel, parser1, parser2) }()

	var records []fasta.Record
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
	uniprotEntryText := strings.NewReader(`<entry dataset="Swiss-Prot" created="2009-05-05" modified="2020-08-12" version="9" xmlns="http://uniprot.org/uniprot">
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
</entry>`)
	// Now we load the parser, and get the first entry out.
	parser := bio.NewUniprotParser(uniprotEntryText)
	entry, _ := parser.Next()

	fmt.Println(entry.Accession[0])
	// Output: P0C9F0
}

func ExampleNewSamParser() {
	// The following can be replaced with a any io.Reader. For example,
	// `file, err := os.Open(path)` for file would also work.
	file := strings.NewReader(`@HD	VN:1.6	SO:unsorted	GO:query
@SQ	SN:pOpen_V3_amplified	LN:2482
@PG	ID:minimap2	PN:minimap2	VN:2.24-r1155-dirty	CL:minimap2 -acLx map-ont - APX814_pass_barcode17_e229f2c8_109f9b91_0.fastq.gz
ae9a66f5-bf71-4572-8106-f6f8dbd3b799	16	pOpen_V3_amplified	1	60	8S54M1D3M1D108M1D1M1D62M226S	*	0	0	AGCATGCCGCTTTTCTGTGACTGGTGAGTACTCAACCAAGTCATTCTGAGAATAGTGTATGCGTGCTGAGTTGCTCTTGCCCGGCGTCAATACGGGATAATACCGCGCCACATAGCAGAACTTTAAAAGTGCTCATCATTGGAAAACGTTCTTCGGGGCGAAAACTCTCGACGTTTACCGCTGTTGAGATCCAGTTCGATGTAACCCACTCGTGCACCCAACTGATCTTCAGCATCAGGGCCGAGCGCAGAAGTGGTCCTGCAACTTTATCCGCCTCCATCCAGTCTATTAATTGTTGCCGGAAGCTAGAGTAAGTAGTTCGCCAGTTAATAGTTTGCGCAACGTTGTTGCCATTGCTACAGGCATCGTGGTTACTGTTGATGTTCATGTAGGTGCTGATCAGAGGTACTTTCCTGGAGGGTTTAACCTTAGCAATACGTAACGGAACGAAGTACAGGGCAT	%,<??@@{O{HS{{MOG{EHD@@=)))'&%%%%'(((6::::=?=;:7)'''/33387-)(*025557CBBDDFDECD;1+'(&&')(,-('))35@>AFDCBD{LNKKGIL{{JLKI{{IFG>==86668789=<><;056<;>=87:840/++1,++)-,-0{{&&%%&&),-13;<{HGVKCGFI{J{L{G{INJHEA@C540/3568;>EOI{{{I0000HHRJ{{{{{{{RH{N@@?AKLQEEC?==<433345588==FTA??A@G?@@@EC?==;10//2333?AB?<<<--(++*''&&-(((+@DBJQHJHGGPJH{.---@B?<''-++'--&%%&,,,FC:999IEGJ{HJHIGIFEGIFMDEF;8878{KJGFIJHIHDCAA=<<<<;DDB>:::EK{{@{E<==HM{{{KF{{{MDEQM{ECA?=>9--,.3))'')*++.-,**()%%	NM:i:8	ms:i:408	AS:i:408	nn:i:0	tp:A:P	cm:i:29	s1:i:195	s2:i:0	de:f:0.0345	SA:Z:pOpen_V3_amplified,2348,-,236S134M1D92S,60,1;	rl:i:0`)
	parser, _ := bio.NewSamParser(file)
	records, _ := parser.Parse() // Parse all data records from file

	fmt.Println(records[0].CIGAR)
	// Output: 8S54M1D3M1D108M1D1M1D62M226S
}

func ExampleFilterData() {
	// Create channels for input and output
	inputChan := make(chan sam.Alignment, 2) // Buffered channel to prevent blocking
	outputChan := make(chan sam.Alignment)

	var results []sam.Alignment
	ctx := context.Background()
	errorGroup, ctx := errgroup.WithContext(ctx)
	errorGroup.Go(func() error {
		return bio.RunWorkers(ctx, 1, outputChan, func(ctx context.Context) error {
			return bio.FilterData(ctx, inputChan, outputChan, func(data sam.Alignment) bool { return (data.FLAG & 0x900) == 0 })
		})
	})

	// Send some example Alignments to the input channel
	inputChan <- sam.Alignment{FLAG: 0x900}              // Not primary, should not be outputted
	inputChan <- sam.Alignment{SEQ: "FAKE", FLAG: 0x000} // Primary, should be outputted
	close(inputChan)                                     // Close the input channel to signal no more data

	// Collect results from the output channel
	for alignment := range outputChan {
		results = append(results, alignment)
	}

	fmt.Println(results)
	// Output: [{ 0  0 0   0 0 FAKE  []}]
}

func Example_runWorkflow() {
	// Workflows are a way of running bioinformatics programs replacing stdin/stdout
	// with go channels. This allows for concurrent processing of data.
	//
	// Currently, we just use standard errorGroup handling for workflows, but
	// aim to support multiple workers for maximizing throughput.

	// First, setup parser
	file := strings.NewReader(`@289a197e-4c05-4143-80e6-488e23044378 runid=bb4427242f6da39e67293199a11c6c4b6ab2b141 read=34575 ch=111 start_time=2023-12-29T16:06:13.719061-08:00 flow_cell_id=AQY258 protocol_group_id=nseq28 sample_id=build3-build3gg-u11 barcode=barcode06 barcode_alias=barcode06 parent_read_id=289a197e-4c05-4143-80e6-488e23044378 basecall_model_version_id=dna_r10.4.1_e8.2_400bps_sup@v4.2.0
TTTTGTCTACTTCGTTCCGTTGCGTATTGCTAAGGTTAAGACTACTTTCTGCCTTTGCGAGACGGCGCCTCCGTGCGACGAGATTTCAAGGGTCTCTGTGCTATATTGCCGCTAGTTCCGCTCTAGCTGCTCCAGTTAATACTACTACTGAAGATGAATTGGAGGGTGACTTCGATGTTGCTGTTCTGCCTTTTTCCGCTTCTGAGACCCAGATCGACTTTTAGATTCCTCAGGTGCTGTTCTCGCAAAGGCAGAAAGTAGTCTTAACCTTAGCAATACGTGG
+
$%%&%%$$%&'+)**,-+)))+866788711112=>A?@@@BDB@>?746@?>A@D2@970,-+..*++++;662/.-.+,++,//+167>A@A@@B=<887-,'&&%%&''((5555644578::<==B?ABCIJA>>>>@DCAA99::<BAA@-----DECJEDDEGEFHE;;;:;;:88754998989998887,-<<;<>>=<<<=67777+***)//+,,+)&&&+--.02:>442000/1225:=D?=<<=7;866/..../AAA226545+&%%$$
@af86ed57-1cfe-486f-8205-b2c8d1186454 runid=bb4427242f6da39e67293199a11c6c4b6ab2b141 read=2233 ch=123 start_time=2023-12-29T10:04:32.719061-08:00 flow_cell_id=AQY258 protocol_group_id=nseq28 sample_id=build3-build3gg-u11 barcode=barcode07 barcode_alias=barcode07 parent_read_id=af86ed57-1cfe-486f-8205-b2c8d1186454 basecall_model_version_id=dna_r10.4.1_e8.2_400bps_sup@v4.2.0
TGTCCTTTACTTCGTTCAGTTACGTATTGCTAAGGTTAAGACTACTTTCTGCCTTTGCGAGAACAGCACCTCTGCTAGGGGCTACTTATCGGGTCTCTAGTTCCGCTCTAGCTGCTCCAGTTAATACTACTACTGAAGATGAATTGGAGGGTGACTTCGATGTTGCTGTTCTGCCTTTTTCCGCTTCTATCTGAGACCGAAGTGGTTTGCCTAAACGCAGGTGCTGTTGGCAAAGGCAGAAAGTAGTCTTAACCTTGACAATGAGTGGTA
+
$%&$$$$$#')+)+,<>@B?>==<>>;;<<<B??>?@DA@?=>==>??<>??7;<706=>=>CBCCB????@CCBDAGFFFGJ<<<<<=54455>@?>:::9..++?@BDCCDCGECFHD@>=<<==>@@B@?@@>>>==>>===>>>A?@ADFGDCA@?????CCCEFDDDDDGJODAA@A;;ABBD<=<:92222223:>>@?@@B?@=<62212=<<<=>AAB=<'&&&'-,-.,**)'&'(,,,-.114888&&&&&'+++++,,*`)
	parser := bio.NewFastqParser(file)

	// We setup the error group here. It's context is used if we need to cancel
	// code that is running.
	ctx := context.Background()
	errorGroup, ctx := errgroup.WithContext(ctx)

	// Now we set up a workflow. We need two things: channels for the internal
	// workflow steps to pass between, and the workflow steps themselves. We
	// Set them up here.
	fastqReads := make(chan fastq.Read)
	fastqBarcoded := make(chan fastq.Read)

	// Read fastqs into channel
	errorGroup.Go(func() error {
		return parser.ParseToChannel(ctx, fastqReads, false)
	})

	// Filter the right barcode fastqs from channel
	barcode := "barcode07"
	errorGroup.Go(func() error {
		// We're going to start multiple workers within this errorGroup. This
		// helps when doing computationally intensive operations on channels.
		return bio.RunWorkers(ctx, 2, fastqBarcoded, func(ctx context.Context) error {
			return bio.FilterData(ctx, fastqReads, fastqBarcoded, func(data fastq.Read) bool { return data.Optionals["barcode"] == barcode })
		})
	})

	// Now, check the outputs. We should have sorted only for barcode07
	var reads []fastq.Read
	for read := range fastqBarcoded {
		reads = append(reads, read)
	}

	fmt.Println(reads[0].Identifier)
	// Output: af86ed57-1cfe-486f-8205-b2c8d1186454
}
