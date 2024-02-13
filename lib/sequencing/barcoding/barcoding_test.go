package barcoding

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/koeng101/dnadesign/lib/bio"
	"github.com/koeng101/dnadesign/lib/sequencing/nanopore"
)

func ExampleSingleBarcodeSequence() {
	// Add in our barcode set. These barcodes are Nanopore native barcode
	// sequences.
	primerSetCsv := fmt.Sprintf("barcode01,%s\nbarcode02,%s", nanopore.NativeBarcodeMap["barcode01"].Forward, nanopore.NativeBarcodeMap["barcode02"].Forward)
	// Add in reads. The following are real reads from a Nanopore sequencing
	// run, barcoded with barcode01
	reads := `@ed433f97-a7e9-4b95-be59-2e18bda14fc2 runid=599e6407c739adc4ecb3169cea4c881b04635ea8 read=18 ch=90 start_time=2024-02-06T15:59:33.914899-08:00 flow_cell_id=ARF133 protocol_group_id=nseq30 sample_id=u12-u13-u14-u15-u16-u17-u18-u19-B4-B5 barcode=barcode01 barcode_alias=barcode01 parent_read_id=ed433f97-a7e9-4b95-be59-2e18bda14fc2 basecall_model_version_id=dna_r10.4.1_e8.2_400bps_sup@v4.2.0
TGTCCCTGTACTTCGTTCAGTTACGTATTGCTAAGGTTAACACAAAGACACCGACAACTTTCTTCAGCACCTGCGGAAGCGCAAGCCTAAGCGTAAAACGACGGCCAGTGAATCCGTAATCATGGTCATAGCTGTTTCCTGACTGGGTCTCGGTCTAGGTCAGGTGCTGAAGTTGTCGGTGTTGCTTGTGTTAACCTTTACGATGGT
+
'(*()(''*--.6+***+6---+,-./=CBDFIKSSGHFCHB@886797777ABCKICGDCAEFCBFDFSIHLSSGSSJSJISFMKSISJSOGSBA@??...--4?IFMRGSPGSRHLJGLEB(((18::=KMSGSJNSSGRGKSHGCCCECLSJRKIKHKJNHMJGH6.,)))((((())//---//0=8756:8892+&%%'((+
@8c5cf4e3-2595-4e1e-b014-9a6f9998b3d3 runid=599e6407c739adc4ecb3169cea4c881b04635ea8 read=31 ch=12 start_time=2024-02-06T15:59:35.914899-08:00 flow_cell_id=ARF133 protocol_group_id=nseq30 sample_id=u12-u13-u14-u15-u16-u17-u18-u19-B4-B5 barcode=barcode01 barcode_alias=barcode01 parent_read_id=8c5cf4e3-2595-4e1e-b014-9a6f9998b3d3 basecall_model_version_id=dna_r10.4.1_e8.2_400bps_sup@v4.2.0
TTGTCCTCTGCTTGGTTTCATTACGTATTGCTAACCAGCACAGAAGACACCGACAACTTTCTTCAGCACCTAGGCGAAGGCCAGCGTAAACGACGGCCAGTGGCTAAGGCGAAGGCCAGCGTAAAACAACACCCATGATCATCATCATCGTCATCACTCTTCCTCCGAGTCTGGC
+
%%%%$###$%&&'%''))))*-,,....0446-,*(('(()()()'((''(//123441,'',--/21112/..-.14457=>;;220-*()),-.234=?>>?=64443//001<34**+3/-,,*())'&(''))))*)*,++,,,*&&&&(((**%%%%(,)'&%%%%&%'&
@30119e9e-19a0-4174-be00-bc61117ce8d6 runid=599e6407c739adc4ecb3169cea4c881b04635ea8 read=57 ch=65 start_time=2024-02-06T16:00:40.914899-08:00 flow_cell_id=ARF133 protocol_group_id=nseq30 sample_id=u12-u13-u14-u15-u16-u17-u18-u19-B4-B5 barcode=barcode01 barcode_alias=barcode01 parent_read_id=30119e9e-19a0-4174-be00-bc61117ce8d6 basecall_model_version_id=dna_r10.4.1_e8.2_400bps_sup@v4.2.0
GTGTAACCTACTGGTTCAGTTACGTATTGCTAAGGTTAACACAAAGACACCGACAACTTTCTTCAGCACCTGTGAAGGTCAAGGGTAAGGGTAAAACGACGGCCAGTCACCAACCATAACCAGAACCATTAAAACGACGGCCAGTGAATCCGTAATCATGGTCATAGCTTTCCTGACTAACACTGTCACTGGAACATGGTCTAGCTGTTTCCTGGTGATAGGTG
+
$$$((45;=@GG7778454446889:AABCFSSOKPIDESGISSGJJSFSHBBB@>??EFBEJKGHGSSQNJJSKMSSOFFDFELFMSSNMHSOMNSIOHSNSSQSLLJSEHHGSISISKSHSHKDC/---,-47?@@ABSGFJICESNFKHSJIJSJSGHJKSKKRF?:7(<70111PKLSKSJSLHPNMFSOSJSLSLHIOGQSHFHA<@F<<<;:=)('''
@a931cb47-d99e-43e0-b6e1-7a52777ab1a2 runid=599e6407c739adc4ecb3169cea4c881b04635ea8 read=39 ch=70 start_time=2024-02-06T16:00:43.914899-08:00 flow_cell_id=ARF133 protocol_group_id=nseq30 sample_id=u12-u13-u14-u15-u16-u17-u18-u19-B4-B5 barcode=barcode01 barcode_alias=barcode01 parent_read_id=a931cb47-d99e-43e0-b6e1-7a52777ab1a2 basecall_model_version_id=dna_r10.4.1_e8.2_400bps_sup@v4.2.0
ATGTTGCCTACCTACTTGGTTCAGTTACGTATTGCTAAGGTTAACACAAAGACACCCGACAACTTTCTTCAGCACCTGCCAAGCGTAAAACGACGGCCAGTGGAAAAGCAAAACTAAAACGGTAAAACGACGGCCAGTGGAAAAGCAAAACTAAAACGGTAAACGACGGCCAGTTACGTTATTTGCTTTTCATGGTCATAGCTGTTTCCTGACTGGGTCTCGGTCTAGGTCAGGTGCTGAAGAAAGTTGTCGGTGTCTTTGTGTTAACCACGATGCGTTGT
+
%%%%%%####$%(*+,)&''));::80-''')0000BCFJOSNLQJKJKRSMSFKLSDHIMGJEHDABFCCHIJKSSISHGJMSKSSKSOLSJSJIOJMSMNRKJKSJQLISGHGIHSJHSLJKSMISJEEBBCBH>=:/---849000>@GEEEJOKJIJHGSISIOMHKKGFGFJMGHHROIPSIJSSJKSLSMMSKJSJIIIKNMMEISKMLSKSHJPHLMJNQLHGRJKKKSHKKOMISOSGFJIOSSLSOSSSSNMSIFH876'&'(....+++(%`

	primerSet, _ := ParseSinglePrimerSet(strings.NewReader(primerSetCsv))
	parser := bio.NewFastqParser(strings.NewReader(reads))
	records, _ := parser.Parse()

	var barcodes []string
	for _, record := range records {
		// Note: Nanopore has a score that requires a lower match (16) than the
		// default ScoreMagicNumber (18).
		barcode, _ := SingleBarcodeSequence(record.Sequence, primerSet)
		if barcode != "" {
			barcodes = append(barcodes, barcode)
		}
	}

	fmt.Println(barcodes)
	// Output: [barcode01 barcode01 barcode01]
}

func ExampleDualBarcodeSequence() {
	// Add in our barcode set. This is an abrdiged version of a full
	// dual barcode plate.
	primerSetCsv := `O1,GAAGCTCAAGCGTAAGCGGA,GAACCACAACCTTAACCTGA
B15,GCGGAAGCGCAAGCCTAAGC,GACCTAGACCGAGACCCAGT
J22,TAACGTGAACGTCAACGGTA,GTGGTAGTGGGAGTGGCAGT
E20,GGTGAAGGTCAAGGGTAAGG,GTTCCAGTGACAGTGTTAGT
C22,GAAGGACAAGGTTAAGGTGA,GTGGTAGTGGGAGTGGCAGT
A15,GGAAAAGCAAAACTAAAACG,GACCTAGACCGAGACCCAGT`
	// Add in reads. The following are real reads from a Nanopore sequencing
	// run, barcoded with the above barcodes.
	reads := `@ed433f97-a7e9-4b95-be59-2e18bda14fc2 runid=599e6407c739adc4ecb3169cea4c881b04635ea8 read=18 ch=90 start_time=2024-02-06T15:59:33.914899-08:00 flow_cell_id=ARF133 protocol_group_id=nseq30 sample_id=u12-u13-u14-u15-u16-u17-u18-u19-B4-B5 barcode=barcode01 barcode_alias=barcode01 parent_read_id=ed433f97-a7e9-4b95-be59-2e18bda14fc2 basecall_model_version_id=dna_r10.4.1_e8.2_400bps_sup@v4.2.0
TGTCCCTGTACTTCGTTCAGTTACGTATTGCTAAGGTTAACACAAAGACACCGACAACTTTCTTCAGCACCTGCGGAAGCGCAAGCCTAAGCGTAAAACGACGGCCAGTGAATCCGTAATCATGGTCATAGCTGTTTCCTGACTGGGTCTCGGTCTAGGTCAGGTGCTGAAGTTGTCGGTGTTGCTTGTGTTAACCTTTACGATGGT
+
'(*()(''*--.6+***+6---+,-./=CBDFIKSSGHFCHB@886797777ABCKICGDCAEFCBFDFSIHLSSGSSJSJISFMKSISJSOGSBA@??...--4?IFMRGSPGSRHLJGLEB(((18::=KMSGSJNSSGRGKSHGCCCECLSJRKIKHKJNHMJGH6.,)))((((())//---//0=8756:8892+&%%'((+
@8c5cf4e3-2595-4e1e-b014-9a6f9998b3d3 runid=599e6407c739adc4ecb3169cea4c881b04635ea8 read=31 ch=12 start_time=2024-02-06T15:59:35.914899-08:00 flow_cell_id=ARF133 protocol_group_id=nseq30 sample_id=u12-u13-u14-u15-u16-u17-u18-u19-B4-B5 barcode=barcode01 barcode_alias=barcode01 parent_read_id=8c5cf4e3-2595-4e1e-b014-9a6f9998b3d3 basecall_model_version_id=dna_r10.4.1_e8.2_400bps_sup@v4.2.0
TTGTCCTCTGCTTGGTTTCATTACGTATTGCTAACCAGCACAGAAGACACCGACAACTTTCTTCAGCACCTAGGCGAAGGCCAGCGTAAACGACGGCCAGTGGCTAAGGCGAAGGCCAGCGTAAAACAACACCCATGATCATCATCATCGTCATCACTCTTCCTCCGAGTCTGGC
+
%%%%$###$%&&'%''))))*-,,....0446-,*(('(()()()'((''(//123441,'',--/21112/..-.14457=>;;220-*()),-.234=?>>?=64443//001<34**+3/-,,*())'&(''))))*)*,++,,,*&&&&(((**%%%%(,)'&%%%%&%'&
@07c30f6c-5d40-4422-8caf-acca91369ae2 runid=599e6407c739adc4ecb3169cea4c881b04635ea8 read=8 ch=68 start_time=2024-02-06T15:59:30.914899-08:00 flow_cell_id=ARF133 protocol_group_id=nseq30 sample_id=u12-u13-u14-u15-u16-u17-u18-u19-B4-B5 barcode=barcode01 barcode_alias=barcode01 parent_read_id=07c30f6c-5d40-4422-8caf-acca91369ae2 basecall_model_version_id=dna_r10.4.1_e8.2_400bps_sup@v4.2.0
ATGTCCTGTACTTCGTTCAGTTACGTATTGCTAAGGTTAACACAAAGACACCGACAACTTTCTTCAGCACCTCCCAGGAAACAGCTATGACCATGAACCACAACCTTAACCTGACAGGAAACAGCTATGACCATGATTACGGATTCACTGGCCGTCGTTTGCTCACCTTTTGCCGCTTCCCCTCGTATGCGGATGCGCATGCCCAGGAAACAGCTATGACCATGATTACGGATTCACTGGCCGTCGTTTTACTCCGCTTACGCTTGAGCTTCACTGGCAGCGTTTTACTCCGCTTACGCTTGAGCTTCACTGGCCGTCGTTTATCCGCTCATG
+
$%&&$$$$&'))10235****+55677?@@ABKNEKFHEFDEDEEEFGKMGKECCCDHFFEAA500188:???BBBEHGELLDDDBBC@ABBCDBAABAHEEHDDAAA?BFJFEFEEDDCGEFEGGFGEEDDFJMFD><=<<>>>@<<;88:77765577,++++++++31(12/...../:2<>?@BHGKGKJGEDCEEBCC?0FFCDGKFEDCCFDDFFKSHHMHFEEEDGFFFDEFE>79877857>B==<<<ABBBKHEIA==<<=;9:;A/--+(&&&*-0*8=?CEIFGDEGCMOEGEHEGG?????DDDH<;:;;AEFB==<%$$$
@d990f6ac-ba6b-4f91-810d-0982b3d300bd runid=599e6407c739adc4ecb3169cea4c881b04635ea8 read=30 ch=3 start_time=2024-02-06T15:59:37.914899-08:00 flow_cell_id=ARF133 protocol_group_id=nseq30 sample_id=u12-u13-u14-u15-u16-u17-u18-u19-B4-B5 barcode=barcode01 barcode_alias=barcode01 parent_read_id=d990f6ac-ba6b-4f91-810d-0982b3d300bd basecall_model_version_id=dna_r10.4.1_e8.2_400bps_sup@v4.2.0
ATGTCCTTGTTCTTCGACATTACGTATTGCTAAGGTTAACACAAAGACACCGACAACTTTCTTCAGCACCTAAGCTCAAGCGTAAGCGGAGTAAAACGACGGCCAGTGTTATTGAATAAATCAGATTTGCCATGGTCTAGCTGTTTCCTGCGATATAGATATGCATATCCATGGTCATAGCTGTTTCCTGTCAGGTTAAGGTTGTGGTTCAGGTGCTGAAGAAAGTTGTCGTTTCTCTCTAATG
+
%%%&&)((''(&&&'34544444454489:CFFGHFNFFE;:442.)&&&+()12=>>??665,+,(((////00566<<<@BBCCLESGMPILIMG=====HECGFEJF))2000;34444)))))*(:BCDDDFFCAA?@CA>>9:CDDCAA>8.,,,,2)))))43844444@ADDGLJ=@DG5H@@B10000EEDFEFSKPGSLNFESHHKJSKHGCABB@BD=;))&&$$$$%%'%&'*
@66889973-c411-4d86-8d80-5040be1f91d1 runid=599e6407c739adc4ecb3169cea4c881b04635ea8 read=20 ch=70 start_time=2024-02-06T15:59:36.914899-08:00 flow_cell_id=ARF133 protocol_group_id=nseq30 sample_id=u12-u13-u14-u15-u16-u17-u18-u19-B4-B5 barcode=barcode01 barcode_alias=barcode01 parent_read_id=66889973-c411-4d86-8d80-5040be1f91d1 basecall_model_version_id=dna_r10.4.1_e8.2_400bps_sup@v4.2.0
TATATGTCCTCTACTCGTTCAGTTACGTATTGCTAAGGTTAACACAAAGACACCGACAACTTTCTTCAGCACCTTAACGTGAACGTCAACGGTAGTAAAACGACGGCCAGTCCCACTACCGGCGTATGGTCATAGCTGTTTCCTCGACTGCCACTCCCACTACCACAGCTGCTGAAGAAAGTTGTAATTTGTCTTTGTGTTAACCTTCACGATAAAC
+
&'(),,)%$###$%*,,,EFDEDFFJELHSHSJSIISSSOSSKLJKIKKKHNOIIJILKOKSMLSIEELGDMGJ=::<>=3334-----44CHDIHHJHSHHHIGSSGHHS@?;89A543**))(()'))*@BCDHJJSFF<D4,((()DSKKLIFBDCFL?>>22('&((())))))-,(*0,,(((('''',+.-/@@>557992&&&&((''&&
@6e9f5fbd-cd9c-4aaf-9831-3fc6c957429a runid=599e6407c739adc4ecb3169cea4c881b04635ea8 read=24 ch=70 start_time=2024-02-06T15:59:38.914899-08:00 flow_cell_id=ARF133 protocol_group_id=nseq30 sample_id=u12-u13-u14-u15-u16-u17-u18-u19-B4-B5 barcode=barcode01 barcode_alias=barcode01 parent_read_id=6e9f5fbd-cd9c-4aaf-9831-3fc6c957429a basecall_model_version_id=dna_r10.4.1_e8.2_400bps_sup@v4.2.0
TTTGCCTGTACTTCGTTCAGTTACGTATTGCTAAGGTTAACACAAAGACACCGACAACTTTCTTCAGCACCTGTGGTAGTGGGAGTGGCAGTCAGGAAACAGCTATGACCATGGTGGTAGTGGGACTGGCCGTGGTTTACGCTTAGCTTAAGCTTGTGCTTATGATTCACTGGCCGTCGTTTTACTCACCTTAACCTTGTCCTTAGGTGCTGAAGAAAGTTGTCGGTGTCTTTGTGTTAACCTTAGCAATACGT
+
$&&&&'%%,+++.**++'****+./00;<?76666GSIGLGIGGFHGGJGSQMNHIGHFI611=/./00HJGJOJJMIJSMMQSIMGEJECJ<:::<EJGLHLSSNNSSJOSGSQPSSJISFSIMJJLSGJIR++++4+1001-%%%&,,;?@@F;;<MFKKNKIKGNSRSHIJGGS??@@@D=FGEEDHFISSLGKSSOJPJIEGHFHIKSJSNSGSLSKNSIJFIIGFJDC<BBKC;77333++/,,,+*)(
@4f88db81-18a2-4353-bde1-2b539425b79a runid=599e6407c739adc4ecb3169cea4c881b04635ea8 read=125 ch=46 start_time=2024-02-06T16:00:54.914899-08:00 flow_cell_id=ARF133 protocol_group_id=nseq30 sample_id=u12-u13-u14-u15-u16-u17-u18-u19-B4-B5 barcode=barcode01 barcode_alias=barcode01 parent_read_id=4f88db81-18a2-4353-bde1-2b539425b79a basecall_model_version_id=dna_r10.4.1_e8.2_400bps_sup@v4.2.0
TATCTGTGTACTTGGTTCAGTTAGTATTGCTAAGGTTAACACAAAGACACCGACAACTTTCTTCAGCACCTGTACCAGTTACAGTTTTAGTCAGGAAACAGCTATGACCATGATTACGGATTCACTGGCCGTCGTTTTACTAGCCTTGCCCTTTCCCTTAACTGGCCGTGGTTTTG
+
'&'&&$$$%&))2/00::;5552((((358;?<<<<<?>?>>>>==998877777:<;<?>==;::<<=??@98888;776668777889932222567889:9:98676765566=:;::;>=><<<<<=>==>??:8;87779;>>=>?>CA;CA/>87889;?@>>8665-*'
@30119e9e-19a0-4174-be00-bc61117ce8d6 runid=599e6407c739adc4ecb3169cea4c881b04635ea8 read=57 ch=65 start_time=2024-02-06T16:00:40.914899-08:00 flow_cell_id=ARF133 protocol_group_id=nseq30 sample_id=u12-u13-u14-u15-u16-u17-u18-u19-B4-B5 barcode=barcode01 barcode_alias=barcode01 parent_read_id=30119e9e-19a0-4174-be00-bc61117ce8d6 basecall_model_version_id=dna_r10.4.1_e8.2_400bps_sup@v4.2.0
GTGTAACCTACTGGTTCAGTTACGTATTGCTAAGGTTAACACAAAGACACCGACAACTTTCTTCAGCACCTGTGAAGGTCAAGGGTAAGGGTAAAACGACGGCCAGTCACCAACCATAACCAGAACCATTAAAACGACGGCCAGTGAATCCGTAATCATGGTCATAGCTTTCCTGACTAACACTGTCACTGGAACATGGTCTAGCTGTTTCCTGGTGATAGGTG
+
$$$((45;=@GG7778454446889:AABCFSSOKPIDESGISSGJJSFSHBBB@>??EFBEJKGHGSSQNJJSKMSSOFFDFELFMSSNMHSOMNSIOHSNSSQSLLJSEHHGSISISKSHSHKDC/---,-47?@@ABSGFJICESNFKHSJIJSJSGHJKSKKRF?:7(<70111PKLSKSJSLHPNMFSOSJSLSLHIOGQSHFHA<@F<<<;:=)('''
@a931cb47-d99e-43e0-b6e1-7a52777ab1a2 runid=599e6407c739adc4ecb3169cea4c881b04635ea8 read=39 ch=70 start_time=2024-02-06T16:00:43.914899-08:00 flow_cell_id=ARF133 protocol_group_id=nseq30 sample_id=u12-u13-u14-u15-u16-u17-u18-u19-B4-B5 barcode=barcode01 barcode_alias=barcode01 parent_read_id=a931cb47-d99e-43e0-b6e1-7a52777ab1a2 basecall_model_version_id=dna_r10.4.1_e8.2_400bps_sup@v4.2.0
ATGTTGCCTACCTACTTGGTTCAGTTACGTATTGCTAAGGTTAACACAAAGACACCCGACAACTTTCTTCAGCACCTGCCAAGCGTAAAACGACGGCCAGTGGAAAAGCAAAACTAAAACGGTAAAACGACGGCCAGTGGAAAAGCAAAACTAAAACGGTAAACGACGGCCAGTTACGTTATTTGCTTTTCATGGTCATAGCTGTTTCCTGACTGGGTCTCGGTCTAGGTCAGGTGCTGAAGAAAGTTGTCGGTGTCTTTGTGTTAACCACGATGCGTTGT
+
%%%%%%####$%(*+,)&''));::80-''')0000BCFJOSNLQJKJKRSMSFKLSDHIMGJEHDABFCCHIJKSSISHGJMSKSSKSOLSJSJIOJMSMNRKJKSJQLISGHGIHSJHSLJKSMISJEEBBCBH>=:/---849000>@GEEEJOKJIJHGSISIOMHKKGFGFJMGHHROIPSIJSSJKSLSMMSKJSJIIIKNMMEISKMLSKSHJPHLMJNQLHGRJKKKSHKKOMISOSGFJIOSSLSOSSSSNMSIFH876'&'(....+++(%
@d8b0c139-fdc3-4c31-9cdc-0c6d1b5e14a2 runid=599e6407c739adc4ecb3169cea4c881b04635ea8 read=48 ch=26 start_time=2024-02-06T16:00:46.914899-08:00 flow_cell_id=ARF133 protocol_group_id=nseq30 sample_id=u12-u13-u14-u15-u16-u17-u18-u19-B4-B5 barcode=barcode01 barcode_alias=barcode01 parent_read_id=d8b0c139-fdc3-4c31-9cdc-0c6d1b5e14a2 basecall_model_version_id=dna_r10.4.1_e8.2_400bps_sup@v4.2.0
GTGTGTACTTCGTTCAGTTACGTATGCTTAGTTGTGAACACAAAGACACCGACAACTTCTTCAGCACCTTCTTGATCTTCATCTGTATCCAGGAAACAGCTATGACCATGCCTGAACCTCAACCGTAACCCAGGAAACAGCTATGACCATGCCTGAACCTCAACCGTAACCCAGGAAACAGCTATGACCATGATTCATCACTGGCCGTGGTTTCGGTTCTGGT
+
$$$%%%%&',*+,//;=:9771111,-+*&&%%%%%'.1389:987777844443553((50+,,88:;2///;6555597677764334578879999:977778888888;=;99888766534424551666688767644445776777:<;:998986652222378*9:::<;:99876322211//..,,,,-0667873333488('''122../
`
	primerSet, _ := ParseDualPrimerSet(strings.NewReader(primerSetCsv))
	parser := bio.NewFastqParser(strings.NewReader(reads))
	records, _ := parser.Parse()

	var wells []string
	for _, record := range records[0:10] {
		well, _ := DualBarcodeSequence(record.Sequence, primerSet)
		if well != "" {
			wells = append(wells, well)
		}
	}

	fmt.Println(wells)
	// Output: [B15 O1 O1 J22 C22 E20 A15]
}
