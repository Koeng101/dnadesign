-- codon_examples_spec.lua
describe("codon examples", function()
	local dnadesign = require("dnadesign")
    local codon = dnadesign.codon

    describe("translation table examples", function()
        it("translates GFP sequence correctly", function()
            local gfp_translation = "MASKGEELFTGVVPILVELDGDVNGHKFSVSGEGEGDATYGKLTLKFICTTGKLPVPWPTLVTTFSYGVQCFSRYPDHMKRHDFFKSAMPEGYVQERTISFKDDGNYKTRAEVKFEGDTLVNRIELKGIDFKEDGNILGHKLEYNYNSHNVYITADKQKNGIKANFKIRHNIEDGSVQLADHYQQNTPIGDGPVLLPDNHYLSTQSALSKDPNEKRDHMVLLEFVTAAGITHGMDELYK*"
            local gfp_dna_sequence = "ATGGCTAGCAAAGGAGAAGAACTTTTCACTGGAGTTGTCCCAATTCTTGTTGAATTAGATGGTGATGTTAATGGGCACAAATTTTCTGTCAGTGGAGAGGGTGAAGGTGATGCTACATACGGAAAGCTTACCCTTAAATTTATTTGCACTACTGGAAAACTACCTGTTCCATGGCCAACACTTGTCACTACTTTCTCTTATGGTGTTCAATGCTTTTCCCGTTATCCGGATCATATGAAACGGCATGACTTTTTCAAGAGTGCCATGCCCGAAGGTTATGTACAGGAACGCACTATATCTTTCAAAGATGACGGGAACTACAAGACGCGTGCTGAAGTCAAGTTTGAAGGTGATACCCTTGTTAATCGTATCGAGTTAAAAGGTATTGATTTTAAAGAAGATGGAAACATTCTCGGACACAAACTCGAGTACAACTATAACTCACACAATGTATACATCACGGCAGACAAACAAAAGAATGGAATCAAAGCTAACTTCAAAATTCGCCACAACATTGAAGATGGATCCGTTCAACTAGCAGACCATTATCAACAAAATACTCCAATTGGCGATGGCCCTGTCCTTTTACCAGACAACCATTACCTGTCGACACAATCTGCCCTTTCGAAAGATCCCAACGAAAAGCGTGACCACATGGTCCTTCTTGAGTTTGTAACTGCTGCTGGGATTACACATGGCATGGATGAGCTCTACAAATAA"
            
            local table = codon.new_translation_table(11)
            local test_translation, err = table:translate(gfp_dna_sequence)
            
            assert.is_nil(err)
            assert.are.equal(gfp_translation, test_translation)
        end)

        it("optimizes sequence and translates back correctly", function()
            local gfp_translation = "MASKGEELFTGVVPILVELDGDVNGHKFSVSGEGEGDATYGKLTLKFICTTGKLPVPWPTLVTTFSYGVQCFSRYPDHMKRHDFFKSAMPEGYVQERTISFKDDGNYKTRAEVKFEGDTLVNRIELKGIDFKEDGNILGHKLEYNYNSHNVYITADKQKNGIKANFKIRHNIEDGSVQLADHYQQNTPIGDGPVLLPDNHYLSTQSALSKDPNEKRDHMVLLEFVTAAGITHGMDELYK*"
            
            -- Use E. coli default table
            local table = codon.default_tables["ecoli"]
            
            -- Set deterministic seed for reproducible test
            _G.SEED_1 = 12345
            _G.SEED_2 = 67890
            
            local optimized_sequence, err1 = table:optimize(gfp_translation)
            assert.is_nil(err1)
            
            local optimized_translation, err2 = table:translate(optimized_sequence)
            assert.is_nil(err2)
            assert.are.equal(gfp_translation, optimized_translation)
        end)

        it("demonstrates compromise table creation", function()
            local ecoli_table = codon.default_tables["ecoli"]
            local pichia_table = codon.default_tables["pichia"]
            
            local final_table, err = codon.compromise_codon_table(ecoli_table, pichia_table, 0.1)
            assert.is_nil(err)
            
            -- Check a specific codon weight (TAA stop codon)
            local taa_weight = 0
            for _, aa in ipairs(final_table.amino_acids) do
                if aa.letter == "*" then
                    for _, cdn in ipairs(aa.codons) do
                        if cdn.triplet == "TAA" then
                            taa_weight = cdn.weight
                            break
                        end
                    end
                end
            end
            
            assert.is_true(taa_weight > 0)  -- Weight should be non-zero
        end)

        it("demonstrates adding tables together", function()
            local pichia_table = codon.default_tables["pichia"]
            local scerevisiae_table = codon.default_tables["scerevisiae"]
            
            local final_table, err = codon.add_codon_table(pichia_table, scerevisiae_table)
            assert.is_nil(err)
            
            -- Check a specific codon weight (GGC)
            local ggc_weight = 0
            for _, aa in ipairs(final_table.amino_acids) do
                for _, cdn in ipairs(aa.codons) do
                    if cdn.triplet == "GGC" then
                        ggc_weight = cdn.weight
                        break
                    end
                end
            end
            
            assert.is_true(ggc_weight > 0)  -- Combined weight should be non-zero
        end)
    end)
end)
