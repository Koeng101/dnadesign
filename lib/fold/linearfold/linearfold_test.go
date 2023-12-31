package linearfold

import (
	"fmt"
	"testing"

	"github.com/koeng101/dnadesign/lib/fold/mfe"
	"github.com/koeng101/dnadesign/lib/fold/mfe/energy_params"
)

func ExampleCONTRAfoldV2() {
	result, score := CONTRAfoldV2("UGAGUUCUCGAUCUCUAAAAUCG", DefaultBeamSize)

	fmt.Printf("result: %v , score: %v", result, score)
	// Output: result: ....................... , score: -0.22376699999999988
}

func ExampleViennaRNAFold() {
	sequence := "UCUAGACUUUUCGAUAUCGCGAAAAAAAAU"
	result, score := ViennaRNAFold(sequence, DefaultTemperature, DefaultEnergyParamsSet, mfe.DefaultDanglingEndsModel, DefaultBeamSize)
	fmt.Printf("result: %v , score: %.2f", result, score)
	// Output: result: .......((((((......))))))..... , score: -3.90
}

func TestLinearFold(t *testing.T) {
	sequence := "UGAGUUCUCGAUCUCUAAAAUCG"
	expectedStructure := "......................."
	expectedScore := -0.22376699999999988
	CONTRAfoldV2Test(sequence, expectedStructure, expectedScore, t)

	sequence = "AAAACGGUCCUUAUCAGGACCAAACA"
	expectedStructure = ".....((((((....))))))....."
	expectedScore = 4.90842861201
	CONTRAfoldV2Test(sequence, expectedStructure, expectedScore, t)

	sequence = "AUUCUUGCUUCAACAGUGUUUGAACGGAAU"
	expectedStructure = ".............................."
	expectedScore = -0.2918699999999998
	CONTRAfoldV2Test(sequence, expectedStructure, expectedScore, t)

	sequence = "UCGGCCACAAACACACAAUCUACUGUUGGUCGA"
	expectedStructure = "(((((((...................)))))))"
	expectedScore = 0.9879274133299999
	CONTRAfoldV2Test(sequence, expectedStructure, expectedScore, t)

	sequence = "GUUUUUAUCUUACACACGCUUGUGUAAGAUAGUUA"
	expectedStructure = ".....(((((((((((....)))))))))))...."
	expectedScore = 6.660038360289999
	CONTRAfoldV2Test(sequence, expectedStructure, expectedScore, t)

	sequence = "GGGCUCGUAGAUCAGCGGUAGAUCGCUUCCUUCGCAAGGAAGCCCUGGGUUCAAAUCCCAGCGAGUCCACCA"
	expectedStructure = "(((((((..((((.......))))(((((((.....))))))).(((((.......))))))))))))...."
	expectedScore = 13.974768784808
	CONTRAfoldV2Test(sequence, expectedStructure, expectedScore, t)
}

func TestViennaRNAFoldDanglingModels(t *testing.T) {
	sequence := "GAGAUACCUACAGCGUGAGCUAUGAGAAAGCGCCACGCUUCCCGAAGGGAGAAAGGCGGACAGGUAUCCGGUAAGCGGCAGGGUCGGAACAGGAGAGCGCACGAGGGAGCUUCCAGGGGGAAACGCCUGGUAUCUUUAUAGUCCUGUCGGGUUUCGCCACCUCUGACUUGAGCGUCGAUUUUUGUGAUGCUCGUCAGGGGGGCGGAGCCUAUGGAAAAACGCCAGCAACGCGGCCUUUUUACGGUUCCUGGCCUUUUGCUGGCCUUUUGCUCACAUGUUCUUUCCUGCGUUAUCCCCUGAUUCUGUGGAUAACCGUAUUACCGCCUUUGAGUGAGCUGAUACCGCUCGCCGCAGCCGAACGACCGAGCGCAGCGAGUCAGUGAGCGAGGAAGCGGAAGAGCGCCCAAUACGCAAACCGCCUCUCCCCGCGCGUUGGCCGAUUCAUUAAUGCAGCUGGCACGACAGGUUUCCCGACUGGAAAGCGGGCAGUGAGCGCAACGCAAUUAAUGUGAGUUAGCUCACUCAUUAGGCACCCCAGGCUUUACACUUUAUGCUUCCGGCUCGUAUGUUGUGUGGAAUUGUGAGCGGAUAACAAUUUCACACAGGAAACAGCUAUGACCAUGAUUACGCCAAGCUUGCAUGCCUGCAGGUCGACUCUAGAGGAUCCCCGGGUACCGAGCUCGAAUUCACUGGCCGUCGUUUUACAACGUCGUGACUGGGAAAACCCUGGCGUUACCCAACUUAAUCGCCUUGCAGCACAUCCCCCUUUCGCCAGCUGGCGUAAUAGCGAAGAGGCCCGCACCGAUCGCCCUUCCCAACAGUUGCGCAGCCUGAAUGGCGAAUGGCGCCUGAUGCGGUAUUUUCUCCUUACGCAUCUGUGCGGUAUUUCACACCGCAUAUGGUGCACUCUCAGUACAAUCUGCUCUGAUGCCGCAUAGUUAAGCCAGCCCCGACACCCGCCAACACCCGCUGACGCGCCCUGACGGGCUUGUCUGCUCCCGGCAUCCGCUUACAGACAAGCUGUGACCGUCUCCGGGAGCUGCAUGUGUCAGAGGUUUUCACCGUCAUCACCGAAACGCGCGAGACGAAAGGGCCUCGUGAUACGCCUAUUUUUAUAGGUUAAUGUCAUGAUAAUAAUGGUUUCUUAGACGUCAGGUGGCACUUUUCGGGGAAAUGUGCGCGGAACCCCUAUUUGUUUAUUUUUCUAAAUACAUUCAAAUAUGUAUCCGCUCAUGAGACAAUAACCCUGAUAAAUGCUUCAAUAAUAUUGAAAAAGGAAGAGUAUGAGUAUUCAACAUUUCCGUGUCGCCCUUAUUCCCUUUUUUGCGGCAUUUUGCCUUCCUGUUUUUGCUCACCCAGAAACGCUGGUGAAAGUAAAAGAUGCUGAAGAUCAGUUGGGUGCACGAGUGGGUUACAUCGAACUGGAUCUCAACAGCGGUAAGAUCCUUGAGAGUUUUCGCCCCGAAGAACGUUUUCCAAUGAUGAGCACUUUUAAAGUUCUGCUAUGUGGCGCGGUAUUAUCCCGUAUUGACGCCGGGCAAGAGCAACUCGGUCGCCGCAUACACUAUUCUCAGAAUGACUUGGUUGAGUACUCACCAGUCACAGAAAAGCAUCUUACGGAUGGCAUGACAGUAAGAGAAUUAUGCAGUGCUGCCAUAACCAUGAGUGAUAACACUGCGGCCAACUUACUUCUGACAACGAUCGGAGGACCGAAGGAGCUAACCGCUUUUUUGCACAACAUGGGGGAUCAUGUAACUCGCCUUGAUCGUUGGGAACCGGAGCUGAAUGAAGCCAUACCAAACGACGAGCGUGACACCACGAUGCCUGUAGCAAUGGCAACAACGUUGCGCAAACUAUUAACUGGCGAACUACUUACUCUAGCUUCCCGGCAACAAUUAAUAGACUGGAUGGAGGCGGAUAAAGUUGCAGGACCACUUCUGCGCUCGGCCCUUCCGGCUGGCUGGUUUAUUGCUGAUAAAUCUGGAGCCGGUGAGCGUGGGUCUCGCGGUAUCAUUGCAGCACUGGGGCCAGAUGGUAAGCCCUCCCGUAUCGUAGUUAUCUACACGACGGGGAGUCAGGCAACUAUGGAUGAACGAAAUAGACAGAUCGCUGAGAUAGGUGCCUCACUGAUUAAGCAUUGGUAACUGUCAGACCAAGUUUACUCAUAUAUACUUUAGAUUGAUUUAAAACUUCAUUUUUAAUUUAAAAGGAUCUAGGUGAAGAUCCUUUUUGAUAAUCUCAUGACCAAAAUCCCUUAACGUGAGUUUUCGUUCCACUGAGCGUCAGACCCCGUAGAAAAGAUCAAAGGAUCUUCUUGAGAUCCUUUUUUUCUGCGCGUAAUCUGCUGCUUGCAAACAAAAAAACCACCGCUACCAGCGGUGGUUUGUUUGCCGGAUCAAGAGCUACCAACUCUUUUUCCGAAGGUAACUGGCUUCAGCAGAGCGCAGAUACCAAAUACUGUUCUUCUAGUGUAGCCGUAGUUAGGCCACCACUUCAAGAACUCUGUAGCACCGCCUACAUACCUCGCUCUGCUAAUCCUGUUACCAGUGGCUGCUGCCAGUGGCGAUAAGUCGUGUCUUACCGGGUUGGACUCAAGACGAUAGUUACCGGAUAAGGCGCAGCGGUCGGGCUGAACGGGGGGUUCGUGCACACAGCCCAGCUUGGAGCGAACGACCUACACCGAACU"
	expectedStructure := "(((((......((((((.(((.......)))..))))))((((...)))).((((((((...((((..((((.....((((((((((....(((.((((((.((((((((..(((((........)))))...........(((....((((((((((.((((((((..((((((((.......))))))))))))))))))))))))))..)))((((.((((((((...((((..............))))..)))))))).))))........)))))))).)))))).))).)))))))))).....)))))))).))))))))((((((((((((...((((((.((.((((......)).)))).))))))......(((.(((.((((....(((.......)))...))))...))).)))((((((..((.((((((......(((.((....(((.........)))....)).)))))))))))))))))...........))))))))))))..((((((...(((((((......(((((.((..((((.((...((((((((((((((.........))))))))))))))...))))))..)).))))).......)))))))..)))))).(((((((.(((((((......(((((.....))))).......(((((((.((.(((..((.(((((...(((....))).(((((((.......))).))))..(((((((.....((.(((((((((((.((((((...((((.(((.((...))...)))))))......)))))))))).....))))))).))(((.....)))((((.(((((.(((((((((((((((((.......)))))))).))))).....(((......)))....(((((((((.((............((((((..((........)).((((.((((....)))).)))).((((((((....((.((((((.....)))))).))...)))))))).((((.((((((.((((......(((.((.((((((.((((..((((.(((((((((((((((.((((((....))))))...)))))))).......)).)))))...))))..)).))...))))))......(((.(((((((...((((((..(((((....)))))....)))))).)).))))).))))))))...)))))))))).))))((((((...))))))...(((((.((.(((....))))).))))))))))).............)).))))))))).((((.(((.....(((((((((.((.....)))))))..))))((((((((..(((..(((((((((..((((((((((..((((..(((........)))))))..)))))((((((((.((((....(((((...)))))......)))).)))))))).((((((((...(((((((((((......)))))).)))))))).))))).)))))..)))).))...)))..))).....(((((.(((.((....))))))))))......))))))))..))).)))).....)))))))))..)))).)))))))......)))))))..))).)).))))))).(((((((((..(((((((((.(((..(((((((((.....)))))))))....((((((....))))))......))))))))))))((..((((.(((((((((.(((..............(((.(((.(((((.....(((.((((((.......)))))))))..(((((((((((.((((((.......))).)))))))......))))))).(((((.((.(((.((....((.((((((....)))))))))).))))))))))((((((((((((((...))))))))...))))))....)))))))).)))))).))))).)))).))))..))))).))))))...((((((...((((....))))...))))))(((.(((((.((((.......(((..........))).....)))).))))).)))((.(((((.(((((.........)))))))))).)).......))))))))))))))...............(((.((((((((((......)))))))))).))).)))))...................(((.((((((((((((..(((((.((((...(((((((((((....((((((((...))))))))))))))))))).....)))).)))))((((((((....(((((((((...)))))))))))))))))((((..(((((.......)))))..))))..((((((.((.((.(((((((((.((............((((((..((((..(((.......)))...))))..))))))..(((((......)))))...))))))))))))).)).))))))...(((((((((....(((.....)))..((((((((((((((.........)))).....)))).)))))))))))))))((((((.((((.....)))).....))))))....))))))))).))).)))........"
	expectedScore := -936.50
	ViennaRNAFoldTest(sequence, expectedStructure, expectedScore, DefaultTemperature, energy_params.Turner2004, mfe.DoubleDanglingEnds, t)
}

func CONTRAfoldV2Test(sequence, expectedStructure string, expectedScore float64, t *testing.T) {
	result, score := CONTRAfoldV2(sequence, DefaultBeamSize)
	if result != expectedStructure || score != expectedScore {
		t.Errorf("Failed to fold %v. \nExpected \nresult: %v \nscore: %v\nGot \nresult: %v \nscore: %v", sequence, expectedStructure, expectedScore, result, score)
	}
}

func ViennaRNAFoldTest(sequence, expectedStructure string, expectedScore float64, temperature float64, energyParamsSet energy_params.EnergyParamsSet, danglingEndsModel mfe.DanglingEndsModel, t *testing.T) {
	result, score := ViennaRNAFold(sequence, temperature, energyParamsSet, danglingEndsModel, DefaultBeamSize)
	// we expect the calculated score to be within 5% of the expected value
	threshold := 0.05
	var lowerThreshold, upperThreshold float64
	if expectedScore < 0 {
		upperThreshold = expectedScore * (1 - threshold)
		lowerThreshold = expectedScore * (1 + threshold)
	} else {
		lowerThreshold = expectedScore * (1 - threshold)
		upperThreshold = expectedScore * (1 + threshold)
	}
	if score < lowerThreshold || score > upperThreshold {
		// if result != expectedStructure || score != expectedScore {
		t.Errorf("Failed to fold %v. \nExpected \nresult: %v \nscore: %v\nGot \nresult: %v \nscore: %v", sequence, expectedStructure, expectedScore, result, score)
	}
}

func BenchmarkViennaRNAFold(b *testing.B) {
	// Run linearfold on a 2915bp sequence
	for n := 0; n < b.N; n++ {
		ViennaRNAFold("GGUCAAGAUGGUAAGGGCCCACGGUGGAUGCCUCGGCACCCGAGCCGAUGAAGGACGUGGCUACCUGCGAUAAGCCAGGGGGAGCCGGUAGCGGGCGUGGAUCCCUGGAUGUCCGAAUGGGGGAACCCGGCCGGCGGGAACGCCGGUCACCGCGCUUUUGCGCGGGGGGAACCUGGGGAACUGAAACAUCUCAGUACCCAGAGGAGAGGAAAGAGAAAUCGACUCCCUGAGUAGCGGCGAGCGAAAGGGGACCAGCCUAAACCGUCCGGCUUGUCCGGGCGGGGUCGUGGGGCCCUCGGACACCGAAUCCCCAGCCUAGCCGAAGCUGUUGGGAAGCAGCGCCAGAGAGGGUGAAAGCCCCGUAGGCGAAAGGUGGGGGGAUAGGUGAGGGUACCCGAGUACCCCGUGGUUCGUGGAGCCAUGGGGGAAUCUGGGCGGACCACCGGCCUAAGGCUAAGUACUCCGGGUGACCGAUAGCGCACCAGUACCGUGAGGGAAAGGUGAAAAGAACCCCGGGAGGGAGUGAAAUAGAGCCUGAAACCGUGGGCUUACAAGCAGUCACGGCCCCGCAAGGGGUUGUGGCGUGCCUAUUGAAGCAUGAGCCGGCGACUCACGGUCGUGGGCGAGCUUAAGCCGUUGAGGCGGAGGCGUAGGGAAACCGAGUCCGAACAGGGCGCAAGCGGGCCGCACGCGGCCCGCAAAGUCCGCGGCCGUGGACCCGAAACCGGGCGAGCUAGCCCUGGCCAGGGUGAAGCUGGGGUGAGACCCAGUGGAGGCCCGAACCGGUGGGGGAUGCAAACCCCUCGGAUGAGCUGGGGCUAGGAGUGAAAAGCUAACCGAGCCCGGAGAUAGCUGGUUCUCCCCGAAAUGACUUUAGGGUCAGCCUCAGGCGCUGACUGGGGCCUGUAGAGCACUGAUAGGGCUAGGGGGCCCACCAGCCUACCAAACCCUGUCAAACUCCGAAGGGUCCCAGGUGGAGCCUGGGAGUGAGGGCGCGAGCGAUAACGUCCGCGUCCGAGCGCGGGAACAACCGAGACCGCCAGCUAAGGCCCCCAAGUCUGGGCUAAGUGGUAAAGGAUGUGGCGCCGCGAAGACAGCCAGGAGGUUGGCUUAGAAGCAGCCAUCCUUUAAAGAGUGCGUAAUAGCUCACUGGUCGAGUGGCGCCGCGCCGAAAAUGAUGCGGGGCUUAAGCCCAGCGCCGAAGCUGCGGGUCUGGGGGAUGACCCCAGGCGGUAGGGGAGCGUUCCCGAUGCCGAUGAAGGCCGACCCGCGAGGCGGCUGGAGGUAAGGGAAGUGCGAAUGCCGGCAUGAGUAACGAUAAAGAGGGUGAGAAUCCCUCUCGCCGUAAGCCCAAGGGUUCCUACGCAAUGGUCGUCAGCGUAGGGUUAGGCGGGACCUAAGGUGAAGCCGAAAGGCGUAGCCGAAGGGCAGCCGGUUAAUAUUCCGGCCCUUCCCGCAGGUGCGAUGGGGGGACGCUCUAGGCUAGGGGGACCGGAGCCAUGGACGAGCCCGGCCAGAAGCGCAGGGUGGGAGGUAGGCAAAUCCGCCUCCCAACAAGCUCUGCGUGGUGGGGAAGCCCGUACGGGUGACAACCCCCCGAAGCCAGGGAGCCAAGAAAAGCCUCUAAGCACAACCUGCGGGAACCCGUACCGCAAACCGACACAGGUGGGCGGGUGCAAGAGCACUCAGGCGCGCGGGAGAACCCUCGCCAAGGAACUCUGCAAGUUGGCCCCGUAACUUCGGGAGAAGGGGUGCUCCCUGGGGUGAUGAGCCCCGGGGAGCCGCAGUGAACAGGCUCUGGCGACUGUUUACCAAAAACACAGCUCUCUGCGAACUCGUAAGAGGAGGUAUAGGGAGCGACGCUUGCCCGGUGCCGGAAGGUCAAGGGGAGGGGUGCAAGCCCCGAACCGAAGCCCCGGUGAACGGCGGCCGUAACUAUAACGGUCCUAAGGUAGCGAAAUUCCUUGUCGGGUAAGUUCCGACCUGCACGAAAAGCGUAACGACCGGAGCGCUGUCUCGGCGAGGGACCCGGUGAAAUUGAACUGGCCGUGAAGAUGCGGCCUACCCGUGGCAGGACGAAAAGACCCCGUGGAGCUUUACUGCAGCCUGGUGUUGGCUCUUGGUCGCGCCUGCGUAGGAUAGGUGGGAGCCUGUGAACCCCCGCCUCCGGGUGGGGGGGAGGCGCCGGUGAAAUACCACCCUGGCGCGGCUGGGGGCCUAACCCUCGGAUGGGGGGACAGCGCUUGGCGGGCAGUUUGACUGGGGCGGUCGCCUCCUAAAAGGUAACGGAGGCGCCCAAAGGUCCCCUCAGGCGGGACGGAAAUCCGCCGGAGAGCGCAAGGGUAGAAGGGGGCCUGACUGCGAGGCCUGCAAGCCGAGCAGGGGCGAAAGCCGGGCCUAGUGAACCGGUGGUCCCGUGUGGAAGGGCCAUCGAUCAACGGAUAAAAGUUACCCCGGGGAUAACAGGCUGAUCUCCCCCGAGCGUCCACAGCGGCGGGGAGGUUUGGCACCUCGAUGUCGGCUCGUCGCAUCCUGGGGCUGAAGAAGGUCCCAAGGGUUGGGCUGUUCGCCCAUUAAAGCGGCACGCGAGCUGGGUUCAGAACGUCGUGAGACAGUUCGGUCUCUAUCCGCCACGGGCGCAGGAGGCUUGAGGGGGGCUCUUCCUAGUACGAGAGGACCGGAAGGGACGCACCUCUGGUUUCCCAGCUGUCCCUCCAGGGGCAUAAGCUGGGUAGCCAUGUGCGGAAGGGAUAACCGCUGAAAGCAUCUAAGCGGGAAGCCCGCCCCAAGAUGAGGCCUCCCACGGCGUCAAGCCGGUAAGGACCCGGGAAGACCACCCGGUGGAUGGGCCGGGGGUGUAAGCGCCGCGAGGCGUUGAGCCGACCGGUCCCAAUCGUCCGAGGUCUUGACCCCUC", DefaultTemperature, DefaultEnergyParamsSet, mfe.DefaultDanglingEndsModel, DefaultBeamSize)
	}
}
