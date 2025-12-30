-- examples/pcr_examples.lua
local dnadesign = require("dnadesign")
local pcr = dnadesign.pcr

describe("PCR Examples", function()
    local gene = "aataattacaccgagataacacatcatggataaaccgatactcaaagattctatgaagctatttgaggcacttggtacgatcaagtcgcgctcaatgtttggtggcttcggacttttcgctgatgaaacgatgtttgcactggttgtgaatgatcaacttcacatacgagcagaccagcaaacttcatctaacttcgagaagcaagggctaaaaccgtacgtttataaaaagcgtggttttccagtcgttactaagtactacgcgatttccgacgacttgtgggaatccagtgaacgcttgatagaagtagcgaagaagtcgttagaacaagccaatttggaaaaaaagcaacaggcaagtagtaagcccgacaggttgaaagacctgcctaacttacgactagcgactgaacgaatgcttaagaaagctggtataaaatcagttgaacaacttgaagagaaaggtgcattgaatgcttacaaagcgatacgtgactctcactccgcaaaagtaagtattgagctactctgggctttagaaggagcgataaacggcacgcactggagcgtcgttcctcaatctcgcagagaagagctggaaaatgcgctttcttaa"
    
    it("demonstrates basic PCR usage", function()
        -- Our cloning scheme requires overhangs
        local forward_overhang = "TTATAGGTCTCATACT"
        local reverse_overhang = "ATGAAGAGACCATATA"
        
        -- Design primers with overhangs for a target Tm of 55
        local fwd, rev = pcr.design_primers_with_overhangs(gene, forward_overhang, reverse_overhang, 55.0)
        
        -- Simulate PCR with potential contaminating sequence
        local bad_fragment = "ATGACCATGATTACGCCAAGCTTGCATGCCTGCAGGTCGACTCTAGAGGATCCCCGGGTACCGAGCTCGAATTCACTGGCCGTCGTTTTACAACGTCGTGACTGGGAAAACCCTGGCGTTACCCAACTTAATCGCCTTGCAGCACATCCCCCTTTCGCCAGCTGGCGTAATAGCGAAGAGGCCCGCACCGATCGCCCTTCCCAACAGTTGCGCAGCCTGAATGGCGAATGGCGCCTGATGCGGTATTTTCTCCTTACGCATCTGTGCGGTATTTCACACCGCATATGGTGCACTCTCAGTACAATCTGCTCTGATGCCGCATAG"
        local fragments, err = pcr.simulate({gene, bad_fragment}, 55.0, false, {fwd, rev})
        
        -- Check we only got one fragment
        assert.are.equal(1, #fragments)
        
        -- Check expected primer sequences
        assert.are.equal("TTATAGGTCTCATACTAATAATTACACCGAGATAACACATCATGG", fwd)
        assert.are.equal("TATATGGTCTCTTCATTTAAGAAAGCGCATTTTCCAGC", rev)
    end)
    
    it("demonstrates designing primers with overhangs", function()
        local forward_overhang = "TTATAGGTCTCATACT"
        local reverse_overhang = "ATGAAGAGACCATATA"
        local fwd, rev = pcr.design_primers_with_overhangs(gene, forward_overhang, reverse_overhang, 55.0)
        
        assert.are.equal("TTATAGGTCTCATACTAATAATTACACCGAGATAACACATCATGG", fwd)
        assert.are.equal("TATATGGTCTCTTCATTTAAGAAAGCGCATTTTCCAGC", rev)
    end)
    
    it("demonstrates designing primers without overhangs", function()
        local fwd, rev = pcr.design_primers(gene, 55.0)
        
        assert.are.equal("AATAATTACACCGAGATAACACATCATGG", fwd)
        assert.are.equal("TTAAGAAAGCGCATTTTCCAGC", rev)
    end)
    
    it("demonstrates PCR simulation", function()
        local primers = {
            "TTATAGGTCTCATACTAATAATTACACCGAGATAACACATCATGG",
            "TATATGGTCTCTTCATTTAAGAAAGCGCATTTTCCAGC"
        }
        local fragments, err = pcr.simulate({gene}, 55.0, false, primers)
        
        assert.are.equal(1, #fragments)
        assert.are.equal("TTATAGGTCTCATACTAATAATTACACCGAGATAACACATCATGGATAAACCGATACTCAAAGATTCTATGAAGCTATTTGAGGCACTTGGTACGATCAAGTCGCGCTCAATGTTTGGTGGCTTCGGACTTTTCGCTGATGAAACGATGTTTGCACTGGTTGTGAATGATCAACTTCACATACGAGCAGACCAGCAAACTTCATCTAACTTCGAGAAGCAAGGGCTAAAACCGTACGTTTATAAAAAGCGTGGTTTTCCAGTCGTTACTAAGTACTACGCGATTTCCGACGACTTGTGGGAATCCAGTGAACGCTTGATAGAAGTAGCGAAGAAGTCGTTAGAACAAGCCAATTTGGAAAAAAAGCAACAGGCAAGTAGTAAGCCCGACAGGTTGAAAGACCTGCCTAACTTACGACTAGCGACTGAACGAATGCTTAAGAAAGCTGGTATAAAATCAGTTGAACAACTTGAAGAGAAAGGTGCATTGAATGCTTACAAAGCGATACGTGACTCTCACTCCGCAAAAGTAAGTATTGAGCTACTCTGGGCTTTAGAAGGAGCGATAAACGGCACGCACTGGAGCGTCGTTCCTCAATCTCGCAGAGAAGAGCTGGAAAATGCGCTTTCTTAAATGAAGAGACCATATA", fragments[1])
    end)
end)
