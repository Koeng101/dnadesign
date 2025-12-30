package codon_test

import (
	"fmt"
	"os"
	"time"

	"github.com/koeng101/dnadesign/lib/bio"
	"github.com/koeng101/dnadesign/lib/synthesis/codon"
)

const puc19path = "../../bio/genbank/data/puc19.gbk"
const phix174path = "../../bio/genbank/data/phix174.gb"

func ExampleTranslationTable_Translate() {
	gfpTranslation := "MASKGEELFTGVVPILVELDGDVNGHKFSVSGEGEGDATYGKLTLKFICTTGKLPVPWPTLVTTFSYGVQCFSRYPDHMKRHDFFKSAMPEGYVQERTISFKDDGNYKTRAEVKFEGDTLVNRIELKGIDFKEDGNILGHKLEYNYNSHNVYITADKQKNGIKANFKIRHNIEDGSVQLADHYQQNTPIGDGPVLLPDNHYLSTQSALSKDPNEKRDHMVLLEFVTAAGITHGMDELYK*"
	gfpDnaSequence := "ATGGCTAGCAAAGGAGAAGAACTTTTCACTGGAGTTGTCCCAATTCTTGTTGAATTAGATGGTGATGTTAATGGGCACAAATTTTCTGTCAGTGGAGAGGGTGAAGGTGATGCTACATACGGAAAGCTTACCCTTAAATTTATTTGCACTACTGGAAAACTACCTGTTCCATGGCCAACACTTGTCACTACTTTCTCTTATGGTGTTCAATGCTTTTCCCGTTATCCGGATCATATGAAACGGCATGACTTTTTCAAGAGTGCCATGCCCGAAGGTTATGTACAGGAACGCACTATATCTTTCAAAGATGACGGGAACTACAAGACGCGTGCTGAAGTCAAGTTTGAAGGTGATACCCTTGTTAATCGTATCGAGTTAAAAGGTATTGATTTTAAAGAAGATGGAAACATTCTCGGACACAAACTCGAGTACAACTATAACTCACACAATGTATACATCACGGCAGACAAACAAAAGAATGGAATCAAAGCTAACTTCAAAATTCGCCACAACATTGAAGATGGATCCGTTCAACTAGCAGACCATTATCAACAAAATACTCCAATTGGCGATGGCCCTGTCCTTTTACCAGACAACCATTACCTGTCGACACAATCTGCCCTTTCGAAAGATCCCAACGAAAAGCGTGACCACATGGTCCTTCTTGAGTTTGTAACTGCTGCTGGGATTACACATGGCATGGATGAGCTCTACAAATAA"
	testTranslation, _ := codon.NewTranslationTable(11).Translate(gfpDnaSequence) // need to specify which codons map to which amino acids per NCBI table

	fmt.Println(gfpTranslation == testTranslation)
	// output: true
}

func ExampleTranslationTable_Optimize() {
	gfpTranslation := "MASKGEELFTGVVPILVELDGDVNGHKFSVSGEGEGDATYGKLTLKFICTTGKLPVPWPTLVTTFSYGVQCFSRYPDHMKRHDFFKSAMPEGYVQERTISFKDDGNYKTRAEVKFEGDTLVNRIELKGIDFKEDGNILGHKLEYNYNSHNVYITADKQKNGIKANFKIRHNIEDGSVQLADHYQQNTPIGDGPVLLPDNHYLSTQSALSKDPNEKRDHMVLLEFVTAAGITHGMDELYK*"

	file, _ := os.Open(puc19path)
	defer file.Close()
	parser := bio.NewGenbankParser(file)
	sequence, _ := parser.Next()
	codonTable := codon.NewTranslationTable(11)
	_ = codonTable.UpdateWeightsWithSequence(sequence)

	// Here, we double check if the number of genes is equal to the number of stop codons
	stopCodonCount := 0
	for _, aa := range codonTable.AminoAcids {
		if aa.Letter == "*" {
			for _, codon := range aa.Codons {
				stopCodonCount = stopCodonCount + codon.Weight
			}
		}
	}

	if stopCodonCount != codonTable.Stats.GeneCount {
		fmt.Println("Stop codons don't equal number of genes!")
	}

	seed := time.Now().UTC().UnixNano()
	optimizedSequence, _ := codonTable.Optimize(gfpTranslation, seed)
	optimizedSequenceTranslation, _ := codonTable.Translate(optimizedSequence)

	fmt.Println(optimizedSequenceTranslation == gfpTranslation)
	// output: true
}

func ExampleReadCodonJSON() {
	codontable := codon.ReadCodonJSON("../../data/bsub_codon_test.json")

	fmt.Println(codontable.GetWeightedAminoAcids()[0].Codons[0].Weight)
	//output: 28327
}

func ExampleParseCodonJSON() {
	file, _ := os.ReadFile("../../data/bsub_codon_test.json")
	codontable := codon.ParseCodonJSON(file)

	fmt.Println(codontable.GetWeightedAminoAcids()[0].Codons[0].Weight)
	//output: 28327
}

func ExampleWriteCodonJSON() {
	codontable := codon.ReadCodonJSON("../../data/bsub_codon_test.json")
	codon.WriteCodonJSON(codontable, "../../data/codon_test.json")
	testCodonTable := codon.ReadCodonJSON("../../data/codon_test.json")

	// cleaning up test data
	os.Remove("../../data/codon_test.json")

	fmt.Println(testCodonTable.GetWeightedAminoAcids()[0].Codons[0].Weight)
	//output: 28327
}

func ExampleCompromiseCodonTable() {
	file, _ := os.Open(puc19path)
	defer file.Close()
	parser := bio.NewGenbankParser(file)
	sequence, _ := parser.Next()

	// weight our codon optimization table using the regions we collected from the genbank file above
	optimizationTable := codon.NewTranslationTable(11)
	err := optimizationTable.UpdateWeightsWithSequence(sequence)
	if err != nil {
		panic(fmt.Errorf("got unexpected error in an example: %w", err))
	}

	file2, _ := os.Open(phix174path)
	defer file2.Close()
	parser2 := bio.NewGenbankParser(file2)
	sequence2, _ := parser2.Next()

	optimizationTable2 := codon.NewTranslationTable(11)
	err = optimizationTable2.UpdateWeightsWithSequence(sequence2)
	if err != nil {
		panic(fmt.Errorf("got unexpected error in an example: %w", err))
	}

	finalTable, _ := codon.CompromiseCodonTable(optimizationTable, optimizationTable2, 0.1)
	for _, aa := range finalTable.GetWeightedAminoAcids() {
		for _, codon := range aa.Codons {
			if codon.Triplet == "TAA" {
				fmt.Println(codon.Weight)
			}
		}
	}
	//output: 2727
}

func ExampleAddCodonTable() {
	file, _ := os.Open(puc19path)
	defer file.Close()
	parser := bio.NewGenbankParser(file)
	sequence, _ := parser.Next()

	// weight our codon optimization table using the regions we collected from the genbank file above
	optimizationTable := codon.NewTranslationTable(11)
	err := optimizationTable.UpdateWeightsWithSequence(sequence)
	if err != nil {
		panic(fmt.Errorf("got unexpected error in an example: %w", err))
	}

	file2, _ := os.Open(phix174path)
	defer file2.Close()
	parser2 := bio.NewGenbankParser(file2)
	sequence2, _ := parser2.Next()

	optimizationTable2 := codon.NewTranslationTable(11)
	err = optimizationTable2.UpdateWeightsWithSequence(sequence2)
	if err != nil {
		panic(fmt.Errorf("got unexpected error in an example: %w", err))
	}

	finalTable, err := codon.AddCodonTable(optimizationTable, optimizationTable2)
	if err != nil {
		panic(fmt.Errorf("got error in adding codon table example: %w", err))
	}

	for _, aa := range finalTable.AminoAcids {
		for _, codon := range aa.Codons {
			if codon.Triplet == "GGC" {
				fmt.Println(codon.Weight)
			}
		}
	}
	//output: 90
}
