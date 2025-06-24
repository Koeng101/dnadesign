local dnadesign = require("dnadesign")
local pcr = dnadesign.pcr

-- Test gene sequence
local gene = "aataattacaccgagataacacatcatggataaaccgatactcaaagattctatgaagctatttgaggcacttggtacgatcaagtcgcgctcaatgtttggtggcttcggacttttcgctgatgaaacgatgtttgcactggttgtgaatgatcaacttcacatacgagcagaccagcaaacttcatctaacttcgagaagcaagggctaaaaccgtacgtttataaaaagcgtggttttccagtcgttactaagtactacgcgatttccgacgacttgtgggaatccagtgaacgcttgatagaagtagcgaagaagtcgttagaacaagccaatttggaaaaaaagcaacaggcaagtagtaagcccgacaggttgaaagacctgcctaacttacgactagcgactgaacgaatgcttaagaaagctggtataaaatcagttgaacaacttgaagagaaaggtgcattgaatgcttacaaagcgatacgtgactctcactccgcaaaagtaagtattgagctactctgggctttagaaggagcgataaacggcacgcactggagcgtcgttcctcaatctcgcagagaagagctggaaaatgcgctttcttaa"

describe("PCR Module", function()
    it("should reject primers with too low tm", function()
        local primers = {
            "TATATGGTCTCTTCATTTAAGAAAGCGCATTTTCCAGC",
            "TTATAGGTCTCATACTAATAATTACACCGAGATAACACATCATGG",
            "CTGCAGGTCGACTCTAG"  -- too low tm
        }
        local fragments = pcr.simulate_simple({gene}, 55.0, false, primers)
        assert.equals(1, #fragments)
    end)

    it("should handle more than one forward primer", function()
        -- Internal primer that occurs inside gene
        local internal_primer = "gatactcaaagattctatgaagctatttgaggcacttggtacg"
        
        -- Different primer that will bind inside gene
        local reverse_primer = "tatcgctttgtaagcattcaatgcacctttctcttcaagttg"
        
        -- Primer that binds out of range of reverse_primer
        local outside_forward_primer = "gtcgttcctcaatctcgcagagaagagctggaaaatg"
        
        local primers = {internal_primer, reverse_primer, outside_forward_primer}
        local fragments = pcr.simulate_simple({gene}, 55.0, false, primers)
        assert.equals(1, #fragments)
    end)

    it("should handle circular DNA templates", function()
        -- Forward primer binds near end of gene
        local forward_primer = "actctgggctttagaaggagcgataaacggc"
        -- Reverse primer binds to beginning of gene
        local reverse_primer = "aagtgcctcaaatagcttcatagaatctttgagtatcgg"
        
        local target_fragment = "ACTCTGGGCTTTAGAAGGAGCGATAAACGGCACGCACTGGAGCGTCGTTCCTCAATCTCGCAGAGAAGAGCTGGAAAATGCGCTTTCTTAAAATAATTACACCGAGATAACACATCATGGATAAACCGATACTCAAAGATTCTATGAAGCTATTTGAGGCACTT"
        
        local fragments = pcr.simulate_simple({gene}, 55.0, true, {forward_primer, reverse_primer})
        assert.equals(target_fragment, fragments[1])
    end)

    it("should detect concatemerization", function()
        local forward_primer = "AATAATTACACCGAGATAACACATCATGG"
        -- This reverse primer adds a forward primer binding site allowing concatemerization
        local reverse_primer = "CCATGATGTGTTATCTCGGTGTAATTATTTTAAGAAAGCGCATTTTCCAGC"
        
        local fragments, err = pcr.simulate({gene}, 55.0, false, {forward_primer, reverse_primer})
        assert.is_not_nil(err)
        assert.matches("Concatemerization detected in PCR.", err)
    end)

    it("should correctly handle issue #279 PCR bug", function()
        local test_gene = "aataattacaccgagataacacatcatggataaaccgatactcaaagattctatgaagctatttgaggcacttggtacgatcaagtcgcgctcaatgtttggtggcttcggacttttcgctgatgaaacgatgtttgcactggttgtgaatgatcaacttcacatacgagcagaccagcaaacttcatctaacttcgagaagcaagggctaaaaccgtacgtttataaaaagcgtggttttccagtcgttactaagtactacgcgatttccgacgacttgtgggaatccagtgaacgcttgatagaagtagcgaagaagtcgttagaacaagccaatttggaaaaaaagcaacaggcaagtagtaagcccgacaggttgaaagacctgcctaacttacgactagcgactgaacgaatgcttaagaaagctggtataaaatcagttgaacaacttgaagagaaaggtgcattgaatgcttacaaagcgatacgtgactctcactccgcaaaagtaagtattgagctactctgggctttagaaggagcgataaacggcacgcactggagcgtcgttcctcaatctcgcagagaagagctggaaaatgcgctttcttaa"
        
        local primers = {
            "TATATGGTCTCTTCATTTAAGAAAGCGCATTTTCCAGC",
            "TTATAGGTCTCATACTAATAATTACACCGAGATAACACATCATGG",
            "CTGCAGGTCGACTCTAG"
        }
        
        local expected = "TTATAGGTCTCATACTAATAATTACACCGAGATAACACATCATGGATAAACCGATACTCAAAGATTCTATGAAGCTATTTGAGGCACTTGGTACGATCAAGTCGCGCTCAATGTTTGGTGGCTTCGGACTTTTCGCTGATGAAACGATGTTTGCACTGGTTGTGAATGATCAACTTCACATACGAGCAGACCAGCAAACTTCATCTAACTTCGAGAAGCAAGGGCTAAAACCGTACGTTTATAAAAAGCGTGGTTTTCCAGTCGTTACTAAGTACTACGCGATTTCCGACGACTTGTGGGAATCCAGTGAACGCTTGATAGAAGTAGCGAAGAAGTCGTTAGAACAAGCCAATTTGGAAAAAAAGCAACAGGCAAGTAGTAAGCCCGACAGGTTGAAAGACCTGCCTAACTTACGACTAGCGACTGAACGAATGCTTAAGAAAGCTGGTATAAAATCAGTTGAACAACTTGAAGAGAAAGGTGCATTGAATGCTTACAAAGCGATACGTGACTCTCACTCCGCAAAAGTAAGTATTGAGCTACTCTGGGCTTTAGAAGGAGCGATAAACGGCACGCACTGGAGCGTCGTTCCTCAATCTCGCAGAGAAGAGCTGGAAAATGCGCTTTCTTAAATGAAGAGACCATATA"
        
        local fragments, err = pcr.simulate({test_gene}, 55.0, false, primers)
        assert.equal(err, "")
        assert.equals(expected, fragments[1])
    end)

    it("should correctly handle issue #70 PCR bug", function()
        local test_gene = "CGAGACcAAGTCGTCATAGCTGTTTCCTGAGAGCTTGGCAGGTGATGACACACATTAACAAATTTCGTGAGGAGTCTCCAGAAGAATGCCATTAATTTCCATAGGCTCCGCCCCCCTGACGAGCATCACAAAAATCGACGCTCAAGTCAGAGGTGGCGAAACCCGACAGGACTATAAAGATACCAGGCGTTTCCCCCTGGAAGCTCCCTCGTGCGCTCTCCTGTTCCGACCCTGCCGCTTACCGGATACCTGTCCGCCTTTCTCCCTTCGGGAAGCGTGGCGCTTTCTCATAGCTCACGCTGTAGGTATCTCAGTTCGGTGTAGGTCGTTCGCTCCAAGCTGGGCTGTGTGCACGAACCCCCCGTTCAGCCCGACCGCTGCGCCTTATCCGGTAACTATCGTCTTGAGTCCAACCCGGTAAGACACGACTTATCGCCACTGGCAGCAGCCACTGGTAACAGGATTAGCAGAGCGAGGTATGTAGGCGGTGCTACAGAGTTCTTGAAGTGGTGGCCTAACTACGGCTACACTAGAAGAACAGTATTTGGTATCTGCGCTCTGCTGAAGCCAGTTACCTTCGGAAAAAGAGTTGGTAGCTCTTGATCCGGCAAACAAACCACCGCTGGTAGCGGTGGTTTTTTTGTTTGCAAGCAGCAGATTACGCGCAGAAAAAAAGGATCTCAAGAAGGCCTACTATTAGCAACAACGATCCTTTGATCTTTTCTACGGGGTCTGACGCTCAGTGGAACGAAAACTCACGTTAAGGGATTTTGGTCATGAGATTATCAAAAAGGATCTTCACCTAGATCCTTTTAAATTAAAAATGAAGTTTTAAATCAATCTAAAGTATATATGAGTAAACTTGGTCTGACAGTTACCAATGCTTAATCAGTGAGGCACCTATCTCAGCGATCTGTCTATTTCGTTCATCCATAGTTGCCTGACTCCCCGTCGTGTAGATAACTACGATACGGGAGGGCTTACCATCTGGCCCCAGTGCTGCAATGATACCGCGAGAACCACGCTCACCGGCTCCAGATTTATCAGCAATAAACCAGCCAGCCGGAAGGGCCGAGCGCAGAAGTGGTCCTGCAACTTTATCCGCCTCCATCCAGTCTATTAATTGTTGCCGGGAAGCTAGAGTAAGTAGTTCGCCAGTTAATAGTTTGCGCAACGTTGTTGCCATTGCTACAGGCATCGTGGTGTCACGCTCGTCGTTTGGTATGGCTTCATTCAGCTCCGGTTCCCAACGATCAAGGCGAGTTACATGATCCCCCATGTTGTGCAAAAAAGCGGTTAGCTCCTTCGGTCCTCCGATCGTTGTCAGAAGTAAGTTGGCCGCAGTGTTATCACTCATGGTTATGGCAGCACTGCATAATTCTCTTACTGTCATGCCATCCGTAAGATGCTTTTCTGTGACTGGTGAGTACTCAACCAAGTCATTCTGAGAATAGTGTATGCGGCGACCGAGTTGCTCTTGCCCGGCGTCAATACGGGATAATACCGCGCCACATAGCAGAACTTTAAAAGTGCTCATCATTGGAAAACGTTCTTCGGGGCGAAAACTCTCAAGGATCTTACCGCTGTTGAGATCCAGTTCGATGTAACCCACTCGTGCACCCAACTGATCTTCAGCATCTTTTACTTTCACCAGCGTTTCTGGGTGAGCAAAAACAGGAAGGCAAAATGCCGCAAAAAAGGGAATAAGGGCGACACGGAAATGTTGAATACTCATACTCTTCCTTTTTCAATATTATTGAAGCATTTATCAGGGTTATTGTCTCATGAGCGGATACATATTTGAATGTATTTAGAAAAATAAACAAATAGGGGTTCCGCGCACCTGCACCAGTCAGTAAAACGACGGCCAGTGACTTgGTCTCAGTCTCAGTCTCATCTTTCCCTTCGTCATGTGACCTGATATCGGGGGTTAGTTCGTCATCATTGATGAGGGTTGATTATCACAGTTTATTACTCTGAATTGGCTATCCGCGTGTGTACCTCTACCTGGAGTTTTTCCCACGGTGGATATTTCTTCTTGCGCTGAGCGTAAGAGCTATCTGACAGAACAGTTCTTCTTTGCTTCCTCGCCAGTTCGCTCGCTATGCTCGGTTACACGGCTGCGGCGAGCATCACGTGCTATAAAA"
        
        local primers = {"GTCATCACCTGCCAAGCTCT", "GGGTTATTGTCTCATGAGCGG"}
        local fragments, err = pcr.simulate({test_gene}, 50.0, true, primers)
        assert.equal(err, "")
        assert.equals(1, #fragments)
    end)
end)
