-- examples/primers_examples.lua
local dnadesign = require("dnadesign")
local primers = dnadesign.primers

describe("Primers Examples", function()
    it("demonstrates Marmur-Doty melting temperature calculation", function()
        local sequence = "ACGTCCGGACTT"
        local melting_temp = primers.marmur_doty(sequence)
        
        assert.are.equal(31, melting_temp)
    end)

    it("demonstrates SantaLucia melting temperature calculation", function()
        local sequence = "ACGATGGCAGTAGCATGC"
        local primer_conc = 0.1e-6  -- primer concentration
        local salt_conc = 350e-3    -- salt concentration
        local mg_conc = 0.0         -- magnesium concentration
        local expected_tm = 62.7    -- expected melting temperature
        
        local melting_temp = primers.santa_lucia(sequence, primer_conc, salt_conc, mg_conc)
        local within_margin = math.abs(expected_tm - melting_temp) / expected_tm < 0.02
        
        assert.is_true(within_margin)
    end)

    it("demonstrates standard melting temperature calculation", function()
        local sequence = "GTAAAACGACGGCCAGT" -- M13 forward primer
        local melting_temp = primers.melting_temp(sequence)
        local expected_tm = 52.8
        local within_margin = math.abs(expected_tm - melting_temp) / expected_tm < 0.02
        
        assert.is_true(within_margin)
    end)

    it("demonstrates De Bruijn sequence generation", function()
        local sequence = primers.nucleobase_debruijn_sequence(4)
        local expected_start = "AAAATAAAGAAACAATTAATGAATC"  -- checking just the start for brevity
        
        assert.are.equal(expected_start, sequence:sub(1, #expected_start))
    end)

    it("demonstrates barcode generation with banned sequences", function()
        local barcodes = primers.create_barcodes_with_banned_sequences(20, 4, {"CTCTCGGTCGCTCC"}, {})
        
        assert.are.equal("AAAATAAAGAAACAATTAAT", barcodes[1])
    end)

    it("demonstrates simple barcode generation", function()
        local barcodes = primers.create_barcodes(20, 4)
        
        assert.are.equal("AAAATAAAGAAACAATTAAT", barcodes[1])
    end)

    it("demonstrates GC-range barcode generation", function()
        local barcodes = primers.create_barcodes_gc_range(20, 4, 0.25, 0.75)
        
        assert.are.equal("GAAACAATTAATGAATCAAG", barcodes[1])
    end)
end)
