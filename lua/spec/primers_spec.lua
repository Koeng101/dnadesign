-- spec/primers_spec.lua
local dnadesign = require("dnadesign")
local primers = dnadesign.primers
local transform = dnadesign.transform

describe("Primers", function()
    describe("MarmurDoty", function()
        it("calculates melting temperature correctly", function()
            local sequence = "ACGTCCGGACTT"
            local expected_tm = 31.0
            local calc_tm = primers.marmur_doty(sequence)
            assert.are.equal(expected_tm, calc_tm)
        end)
    end)

    describe("SantaLucia", function()
        it("calculates melting temperature within expected margin", function()
            local sequence = "ACGATGGCAGTAGCATGC"
            local primer_conc = 0.1e-6
            local salt_conc = 350e-3
            local mg_conc = 0.0
            local expected_tm = 62.7

            local calc_tm = primers.santa_lucia(sequence, primer_conc, salt_conc, mg_conc)
            local margin = math.abs(expected_tm - calc_tm) / expected_tm
            assert.is_true(margin < 0.02)
        end)

        it("handles reverse complement sequences correctly", function()
            local sequence = "ACGTAGATCTACGT"
            local rev_comp = transform.reverse_complement(sequence)
            
            -- Test if sequence is its own reverse complement
            assert.are.equal(sequence, rev_comp)

            local primer_conc = 0.1e-6
            local salt_conc = 350e-3
            local mg_conc = 0.0
            local expected_tm = 47.428514

            local calc_tm = primers.santa_lucia(sequence, primer_conc, salt_conc, mg_conc)
            local margin = math.abs(expected_tm - calc_tm) / expected_tm
            assert.is_true(margin < 0.02)
        end)
    end)

    describe("MeltingTemp", function()
        it("calculates M13 forward primer melting temperature correctly", function()
            local sequence = "GTAAAACGACGGCCAGT"
            local expected_tm = 52.8
            local calc_tm = primers.melting_temp(sequence)
            local margin = math.abs(expected_tm - calc_tm) / expected_tm
            assert.is_true(margin < 0.02)
        end)
    end)

    describe("NucleobaseDeBruijnSequence", function()
        it("generates expected sequence", function()
            local sequence = primers.nucleobase_debruijn_sequence(4)
            local expected = "AAAATAAAGAAACAATTAATGAATCAAGTAAGGAAGCAACTAACGAACCATATAGATACATTTATTGATTCATGTATGGATGCATCTATCGATCCAGAGACAGTTAGTGAGTCAGGTAGGGAGGCAGCTAGCGAGCCACACTTACTGACTCACGTACGGACGCACCTACCGACCCTTTTGTTTCTTGGTTGCTTCGTTCCTGTGTCTGGGTGGCTGCGTGCCTCTCGGTCGCTCCGTCCCGGGGCGGCCGCGCCCCAAA"
            assert.are.equal(expected, sequence)
        end)
    end)

    describe("CreateBarcodes", function()
        it("creates barcodes with expected sequences", function()
            local barcodes = primers.create_barcodes(20, 4)
            assert.are.equal("AAAATAAAGAAACAATTAAT", barcodes[1])
        end)

        it("creates barcodes with banned sequences", function()
            local barcodes = primers.create_barcodes_with_banned_sequences(20, 4, {"CTCTCGGTCGCTCC"}, {})
            assert.are.equal("AAAATAAAGAAACAATTAAT", barcodes[1])
        end)

        it("creates barcodes within GC range", function()
            local barcodes = primers.create_barcodes_gc_range(20, 4, 0.25, 0.75)
            assert.are.equal("GAAACAATTAATGAATCAAG", barcodes[1])
        end)

        it("handles banned sequence functions", function()
            local test_func = function(s)
                return not string.find(s, "GGCCGCGCCCC", 1, true)
            end
            
            local barcodes = primers.create_barcodes_with_banned_sequences(20, 4, {}, {test_func})
            local output = barcodes[#barcodes]
            assert.are.equal("CTCTCGGTCGCTCCGTCCCG", output)

            -- Test with banned string
            barcodes = primers.create_barcodes_with_banned_sequences(20, 4, {"GGCCGCGCCCC"}, {})
            output = barcodes[#barcodes]
            assert.are.equal("CTCTCGGTCGCTCCGTCCCG", output)

            -- Test with reverse complement of banned string
            barcodes = primers.create_barcodes_with_banned_sequences(20, 4, {transform.reverse_complement("GGCCGCGCCCC")}, {})
            output = barcodes[#barcodes]
            assert.are.equal("CTCTCGGTCGCTCCGTCCCG", output)
        end)
    end)
end)
