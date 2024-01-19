/*
Package nanopore contains data associated with nanopore sequencing.
*/
package nanopore

// NativeBarcode contains the data structure defining a nanopore barcode.
// In between Forward and a target DNA sequence is 8bp: CAGCACC followed by a
// T, which is used for the TA ligation to the target end-prepped DNA.
type NativeBarcode struct {
	Forward string `json:"forward"`
	Reverse string `json:"reverse"`
}

// NativeBarcodeMap contains a map of native barcodes to their respective
// forward and reverse sequences.
var NativeBarcodeMap = map[string]NativeBarcode{
	"barcode01": {"CACAAAGACACCGACAACTTTCTT", "AAGAAAGTTGTCGGTGTCTTTGTG"},
	"barcode02": {"ACAGACGACTACAAACGGAATCGA", "TCGATTCCGTTTGTAGTCGTCTGT"},
	"barcode03": {"CCTGGTAACTGGGACACAAGACTC", "GAGTCTTGTGTCCCAGTTACCAGG"},
	"barcode04": {"TAGGGAAACACGATAGAATCCGAA", "TTCGGATTCTATCGTGTTTCCCTA"},
	"barcode05": {"AAGGTTACACAAACCCTGGACAAG", "CTTGTCCAGGGTTTGTGTAACCTT"},
	"barcode06": {"GACTACTTTCTGCCTTTGCGAGAA", "TTCTCGCAAAGGCAGAAAGTAGTC"},
	"barcode07": {"AAGGATTCATTCCCACGGTAACAC", "GTGTTACCGTGGGAATGAATCCTT"},
	"barcode08": {"ACGTAACTTGGTTTGTTCCCTGAA", "TTCAGGGAACAAACCAAGTTACGT"},
	"barcode09": {"AACCAAGACTCGCTGTGCCTAGTT", "AACTAGGCACAGCGAGTCTTGGTT"},
	"barcode10": {"GAGAGGACAAAGGTTTCAACGCTT", "AAGCGTTGAAACCTTTGTCCTCTC"},
	"barcode11": {"TCCATTCCCTCCGATAGATGAAAC", "GTTTCATCTATCGGAGGGAATGGA"},
	"barcode12": {"TCCGATTCTGCTTCTTTCTACCTG", "CAGGTAGAAAGAAGCAGAATCGGA"},
	"barcode13": {"AGAACGACTTCCATACTCGTGTGA", "TCACACGAGTATGGAAGTCGTTCT"},
	"barcode14": {"AACGAGTCTCTTGGGACCCATAGA", "TCTATGGGTCCCAAGAGACTCGTT"},
	"barcode15": {"AGGTCTACCTCGCTAACACCACTG", "CAGTGGTGTTAGCGAGGTAGACCT"},
	"barcode16": {"CGTCAACTGACAGTGGTTCGTACT", "AGTACGAACCACTGTCAGTTGACG"},
	"barcode17": {"ACCCTCCAGGAAAGTACCTCTGAT", "ATCAGAGGTACTTTCCTGGAGGGT"},
	"barcode18": {"CCAAACCCAACAACCTAGATAGGC", "GCCTATCTAGGTTGTTGGGTTTGG"},
	"barcode19": {"GTTCCTCGTGCAGTGTCAAGAGAT", "ATCTCTTGACACTGCACGAGGAAC"},
	"barcode20": {"TTGCGTCCTGTTACGAGAACTCAT", "ATGAGTTCTCGTAACAGGACGCAA"},
	"barcode21": {"GAGCCTCTCATTGTCCGTTCTCTA", "TAGAGAACGGACAATGAGAGGCTC"},
	"barcode22": {"ACCACTGCCATGTATCAAAGTACG", "CGTACTTTGATACATGGCAGTGGT"},
	"barcode23": {"CTTACTACCCAGTGAACCTCCTCG", "CGAGGAGGTTCACTGGGTAGTAAG"},
	"barcode24": {"GCATAGTTCTGCATGATGGGTTAG", "CTAACCCATCATGCAGAACTATGC"},
	"barcode25": {"GTAAGTTGGGTATGCAACGCAATG", "CATTGCGTTGCATACCCAACTTAC"},
	"barcode26": {"CATACAGCGACTACGCATTCTCAT", "ATGAGAATGCGTAGTCGCTGTATG"},
	"barcode27": {"CGACGGTTAGATTCACCTCTTACA", "TGTAAGAGGTGAATCTAACCGTCG"},
	"barcode28": {"TGAAACCTAAGAAGGCACCGTATC", "GATACGGTGCCTTCTTAGGTTTCA"},
	"barcode29": {"CTAGACACCTTGGGTTGACAGACC", "GGTCTGTCAACCCAAGGTGTCTAG"},
	"barcode30": {"TCAGTGAGGATCTACTTCGACCCA", "TGGGTCGAAGTAGATCCTCACTGA"},
	"barcode31": {"TGCGTACAGCAATCAGTTACATTG", "CAATGTAACTGATTGCTGTACGCA"},
	"barcode32": {"CCAGTAGAAGTCCGACAACGTCAT", "ATGACGTTGTCGGACTTCTACTGG"},
	"barcode33": {"CAGACTTGGTACGGTTGGGTAACT", "AGTTACCCAACCGTACCAAGTCTG"},
	"barcode34": {"GGACGAAGAACTCAAGTCAAAGGC", "GCCTTTGACTTGAGTTCTTCGTCC"},
	"barcode35": {"CTACTTACGAAGCTGAGGGACTGC", "GCAGTCCCTCAGCTTCGTAAGTAG"},
	"barcode36": {"ATGTCCCAGTTAGAGGAGGAAACA", "TGTTTCCTCCTCTAACTGGGACAT"},
	"barcode37": {"GCTTGCGATTGATGCTTAGTATCA", "TGATACTAAGCATCAATCGCAAGC"},
	"barcode38": {"ACCACAGGAGGACGATACAGAGAA", "TTCTCTGTATCGTCCTCCTGTGGT"},
	"barcode39": {"CCACAGTGTCAACTAGAGCCTCTC", "GAGAGGCTCTAGTTGACACTGTGG"},
	"barcode40": {"TAGTTTGGATGACCAAGGATAGCC", "GGCTATCCTTGGTCATCCAAACTA"},
	"barcode41": {"GGAGTTCGTCCAGAGAAGTACACG", "CGTGTACTTCTCTGGACGAACTCC"},
	"barcode42": {"CTACGTGTAAGGCATACCTGCCAG", "CTGGCAGGTATGCCTTACACGTAG"},
	"barcode43": {"CTTTCGTTGTTGACTCGACGGTAG", "CTACCGTCGAGTCAACAACGAAAG"},
	"barcode44": {"AGTAGAAAGGGTTCCTTCCCACTC", "GAGTGGGAAGGAACCCTTTCTACT"},
	"barcode45": {"GATCCAACAGAGATGCCTTCAGTG", "CACTGAAGGCATCTCTGTTGGATC"},
	"barcode46": {"GCTGTGTTCCACTTCATTCTCCTG", "CAGGAGAATGAAGTGGAACACAGC"},
	"barcode47": {"GTGCAACTTTCCCACAGGTAGTTC", "GAACTACCTGTGGGAAAGTTGCAC"},
	"barcode48": {"CATCTGGAACGTGGTACACCTGTA", "TACAGGTGTACCACGTTCCAGATG"},
	"barcode49": {"ACTGGTGCAGCTTTGAACATCTAG", "CTAGATGTTCAAAGCTGCACCAGT"},
	"barcode50": {"ATGGACTTTGGTAACTTCCTGCGT", "ACGCAGGAAGTTACCAAAGTCCAT"},
	"barcode51": {"GTTGAATGAGCCTACTGGGTCCTC", "GAGGACCCAGTAGGCTCATTCAAC"},
	"barcode52": {"TGAGAGACAAGATTGTTCGTGGAC", "GTCCACGAACAATCTTGTCTCTCA"},
	"barcode53": {"AGATTCAGACCGTCTCATGCAAAG", "CTTTGCATGAGACGGTCTGAATCT"},
	"barcode54": {"CAAGAGCTTTGACTAAGGAGCATG", "CATGCTCCTTAGTCAAAGCTCTTG"},
	"barcode55": {"TGGAAGATGAGACCCTGATCTACG", "CGTAGATCAGGGTCTCATCTTCCA"},
	"barcode56": {"TCACTACTCAACAGGTGGCATGAA", "TTCATGCCACCTGTTGAGTAGTGA"},
	"barcode57": {"GCTAGGTCAATCTCCTTCGGAAGT", "ACTTCCGAAGGAGATTGACCTAGC"},
	"barcode58": {"CAGGTTACTCCTCCGTGAGTCTGA", "TCAGACTCACGGAGGAGTAACCTG"},
	"barcode59": {"TCAATCAAGAAGGGAAAGCAAGGT", "ACCTTGCTTTCCCTTCTTGATTGA"},
	"barcode60": {"CATGTTCAACCAAGGCTTCTATGG", "CCATAGAAGCCTTGGTTGAACATG"},
	"barcode61": {"AGAGGGTACTATGTGCCTCAGCAC", "GTGCTGAGGCACATAGTACCCTCT"},
	"barcode62": {"CACCCACACTTACTTCAGGACGTA", "TACGTCCTGAAGTAAGTGTGGGTG"},
	"barcode63": {"TTCTGAAGTTCCTGGGTCTTGAAC", "GTTCAAGACCCAGGAACTTCAGAA"},
	"barcode64": {"GACAGACACCGTTCATCGACTTTC", "GAAAGTCGATGAACGGTGTCTGTC"},
	"barcode65": {"TTCTCAGTCTTCCTCCAGACAAGG", "CCTTGTCTGGAGGAAGACTGAGAA"},
	"barcode66": {"CCGATCCTTGTGGCTTCTAACTTC", "GAAGTTAGAAGCCACAAGGATCGG"},
	"barcode67": {"GTTTGTCATACTCGTGTGCTCACC", "GGTGAGCACACGAGTATGACAAAC"},
	"barcode68": {"GAATCTAAGCAAACACGAAGGTGG", "CCACCTTCGTGTTTGCTTAGATTC"},
	"barcode69": {"TACAGTCCGAGCCTCATGTGATCT", "AGATCACATGAGGCTCGGACTGTA"},
	"barcode70": {"ACCGAGATCCTACGAATGGAGTGT", "ACACTCCATTCGTAGGATCTCGGT"},
	"barcode71": {"CCTGGGAGCATCAGGTAGTAACAG", "CTGTTACTACCTGATGCTCCCAGG"},
	"barcode72": {"TAGCTGACTGTCTTCCATACCGAC", "GTCGGTATGGAAGACAGTCAGCTA"},
	"barcode73": {"AAGAAACAGGATGACAGAACCCTC", "GAGGGTTCTGTCATCCTGTTTCTT"},
	"barcode74": {"TACAAGCATCCCAACACTTCCACT", "AGTGGAAGTGTTGGGATGCTTGTA"},
	"barcode75": {"GACCATTGTGATGAACCCTGTTGT", "ACAACAGGGTTCATCACAATGGTC"},
	"barcode76": {"ATGCTTGTTACATCAACCCTGGAC", "GTCCAGGGTTGATGTAACAAGCAT"},
	"barcode77": {"CGACCTGTTTCTCAGGGATACAAC", "GTTGTATCCCTGAGAAACAGGTCG"},
	"barcode78": {"AACAACCGAACCTTTGAATCAGAA", "TTCTGATTCAAAGGTTCGGTTGTT"},
	"barcode79": {"TCTCGGAGATAGTTCTCACTGCTG", "CAGCAGTGAGAACTATCTCCGAGA"},
	"barcode80": {"CGGATGAACATAGGATAGCGATTC", "GAATCGCTATCCTATGTTCATCCG"},
	"barcode81": {"CCTCATCTTGTGAAGTTGTTTCGG", "CCGAAACAACTTCACAAGATGAGG"},
	"barcode82": {"ACGGTATGTCGAGTTCCAGGACTA", "TAGTCCTGGAACTCGACATACCGT"},
	"barcode83": {"TGGCTTGATCTAGGTAAGGTCGAA", "TTCGACCTTACCTAGATCAAGCCA"},
	"barcode84": {"GTAGTGGACCTAGAACCTGTGCCA", "TGGCACAGGTTCTAGGTCCACTAC"},
	"barcode85": {"AACGGAGGAGTTAGTTGGATGATC", "GATCATCCAACTAACTCCTCCGTT"},
	"barcode86": {"AGGTGATCCCAACAAGCGTAAGTA", "TACTTACGCTTGTTGGGATCACCT"},
	"barcode87": {"TACATGCTCCTGTTGTTAGGGAGG", "CCTCCCTAACAACAGGAGCATGTA"},
	"barcode88": {"TCTTCTACTACCGATCCGAAGCAG", "CTGCTTCGGATCGGTAGTAGAAGA"},
	"barcode89": {"ACAGCATCAATGTTTGGCTAGTTG", "CAACTAGCCAAACATTGATGCTGT"},
	"barcode90": {"GATGTAGAGGGTACGGTTTGAGGC", "GCCTCAAACCGTACCCTCTACATC"},
	"barcode91": {"GGCTCCATAGGAACTCACGCTACT", "AGTAGCGTGAGTTCCTATGGAGCC"},
	"barcode92": {"TTGTGAGTGGAAAGATACAGGACC", "GGTCCTGTATCTTTCCACTCACAA"},
	"barcode93": {"AGTTTCCATCACTTCAGACTTGGG", "CCCAAGTCTGAAGTGATGGAAACT"},
	"barcode94": {"GATTGTCCTCAAACTGCCACCTAC", "GTAGGTGGCAGTTTGAGGACAATC"},
	"barcode95": {"CCTGTCTGGAAGAAGAATGGACTT", "AAGTCCATTCTTCTTCCAGACAGG"},
	"barcode96": {"CTGAACGGTCATAGAGTCCACCAT", "ATGGTGGACTCTATGACCGTTCAG"},
}
