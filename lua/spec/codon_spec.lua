-- codon_spec.lua
describe("codon", function()
	local dnadesign = require("dnadesign")
    local codon = dnadesign.codon
    
    describe("translation", function()
        local gfp_translation = "MASKGEELFTGVVPILVELDGDVNGHKFSVSGEGEGDATYGKLTLKFICTTGKLPVPWPTLVTTFSYGVQCFSRYPDHMKRHDFFKSAMPEGYVQERTISFKDDGNYKTRAEVKFEGDTLVNRIELKGIDFKEDGNILGHKLEYNYNSHNVYITADKQKNGIKANFKIRHNIEDGSVQLADHYQQNTPIGDGPVLLPDNHYLSTQSALSKDPNEKRDHMVLLEFVTAAGITHGMDELYK*"
        local gfp_dna_sequence = "ATGGCTAGCAAAGGAGAAGAACTTTTCACTGGAGTTGTCCCAATTCTTGTTGAATTAGATGGTGATGTTAATGGGCACAAATTTTCTGTCAGTGGAGAGGGTGAAGGTGATGCTACATACGGAAAGCTTACCCTTAAATTTATTTGCACTACTGGAAAACTACCTGTTCCATGGCCAACACTTGTCACTACTTTCTCTTATGGTGTTCAATGCTTTTCCCGTTATCCGGATCATATGAAACGGCATGACTTTTTCAAGAGTGCCATGCCCGAAGGTTATGTACAGGAACGCACTATATCTTTCAAAGATGACGGGAACTACAAGACGCGTGCTGAAGTCAAGTTTGAAGGTGATACCCTTGTTAATCGTATCGAGTTAAAAGGTATTGATTTTAAAGAAGATGGAAACATTCTCGGACACAAACTCGAGTACAACTATAACTCACACAATGTATACATCACGGCAGACAAACAAAAGAATGGAATCAAAGCTAACTTCAAAATTCGCCACAACATTGAAGATGGATCCGTTCAACTAGCAGACCATTATCAACAAAATACTCCAATTGGCGATGGCCCTGTCCTTTTACCAGACAACCATTACCTGTCGACACAATCTGCCCTTTCGAAAGATCCCAACGAAAAGCGTGACCACATGGTCCTTCTTGAGTTTGTAACTGCTGCTGGGATTACACATGGCATGGATGAGCTCTACAAATAA"
        
        it("translates DNA sequence correctly", function()
            local result, err = codon.new_translation_table(11):translate(gfp_dna_sequence)
            assert.is_nil(err)
            assert.are.equal(gfp_translation, result)
        end)
        
        it("handles mixed case", function()
            local mixed_case = gfp_dna_sequence:sub(1,100):upper() .. gfp_dna_sequence:sub(101):lower()
            local result, err = codon.new_translation_table(11):translate(mixed_case)
            assert.is_nil(err)
            assert.are.equal(gfp_translation, result)
        end)
        
        it("handles lower case", function()
            local result, err = codon.new_translation_table(11):translate(gfp_dna_sequence:lower())
            assert.is_nil(err)
            assert.are.equal(gfp_translation, result)
        end)
    end)
    
    describe("optimization", function()
        local gfp_translation = "MASKGEELFTGVVPILVELDGDVNGHKFSVSGEGEGDATYGKLTLKFICTTGKLPVPWPTLVTTFSYGVQCFSRYPDHMKRHDFFKSAMPEGYVQERTISFKDDGNYKTRAEVKFEGDTLVNRIELKGIDFKEDGNILGHKLEYNYNSHNVYITADKQKNGIKANFKIRHNIEDGSVQLADHYQQNTPIGDGPVLLPDNHYLSTQSALSKDPNEKRDHMVLLEFVTAAGITHGMDELYK*"
        
        it("errors on invalid amino acid", function()
            local result, err = codon.new_translation_table(1):optimize("TOP", 0)
            assert.are.equal('amino acid "O" is missing from codon table', err)
            assert.are.equal("", result)
        end)
    end)
    
    describe("codon frequency", function()
        it("counts codon frequencies correctly", function()
            local test_sequence = "ATGATGATG"  -- 3 ATG codons
            local frequencies = codon.get_codon_frequency(test_sequence)
            assert.are.equal(3, frequencies["ATG"])
        end)
        
        it("handles mixed case", function()
            local frequencies = codon.get_codon_frequency("ATGatgATG")
            assert.are.equal(3, frequencies["ATG"])
        end)
    end)
    
    describe("compromise tables", function()
        it("validates cut-off range", function()
            local t1 = codon.new_translation_table(11)
            local t2 = codon.new_translation_table(11)
            
            local _, err = codon.compromise_codon_table(t1, t2, -1.0)
            assert.are.equal("cut off too low, cannot be less than 0", err)
            
            local _, err2 = codon.compromise_codon_table(t1, t2, 10.0)
            assert.are.equal("cut off too high, cannot be greater than 1", err2)
        end)
        
        it("combines tables correctly", function()
            local t1 = codon.new_translation_table(11)
            local t2 = codon.new_translation_table(11)
            
            local result, err = codon.compromise_codon_table(t1, t2, 0.1)
            assert.is_nil(err)
            assert.is_not_nil(result)
        end)
    end)
    
    describe("adding tables", function()
        it("combines weights correctly", function()
            local t1 = codon.new_translation_table(11)
            local t2 = codon.new_translation_table(11)
            
            local result, err = codon.add_codon_table(t1, t2)
            assert.is_nil(err)
            assert.is_not_nil(result)
        end)
    end)
    
    describe("capitalization regression", function()
        it("handles mixed case amino acids", function()
            local mixed_case_gfp = "MaSKGEELFTGVVPILVELDGDVNGHKFSVSGEGEGDATYGKLTLKFICTTGKLPVPWPTLVTTFSYGVQCFSRYPDHMKRHDFFKSAMPEGYVQERTISFKDDGNYKTRAEVKFEGDTLVNRIELKGIDFKEDGNILGHKLEYNYNSHNVYITADKQKNGIKANFKIRHNIEDGSVQLADHYQQNTPIGDGPVLLPDNHYLSTQSALSKDPNEKRDHMVLLEFVTAAGITHGMDELYK*"
            local table = codon.default_tables["ecoli"]
            
            local sequence, err = table:optimize(mixed_case_gfp)
            assert.is_nil(err)
            
            local translation, err2 = table:translate(sequence)
            assert.is_nil(err2)
            assert.are.equal(mixed_case_gfp:upper(), translation)
        end)
    end)

    describe("standardize last codon", function()
        it("replaces last codon with standard codon", function()
            local table = codon.new_translation_table(11)
            local test_sequence = "ATGAAATTC"  -- ATG AAA TTC (M K F)
            
            local result, err = table:standardize_last_codon(test_sequence)
            assert.is_nil(err)
            assert.are.equal("ATGAAATTT", result)  -- Should end with TTT (standard F codon)
        end)
        
        it("handles empty sequence", function()
            local table = codon.new_translation_table(11)
            local result, err = table:standardize_last_codon("")
            assert.are.equal("", result)
            assert.are.equal("empty sequence string", err)
        end)
        
        it("handles sequence shorter than 3 bases", function()
            local table = codon.new_translation_table(11)
            local result, err = table:standardize_last_codon("AT")
            assert.are.equal("AT", result)
            assert.is_nil(err)
        end)
    end)
end)
