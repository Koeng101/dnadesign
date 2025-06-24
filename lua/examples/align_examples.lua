-- examples/align_examples.lua
local dnadesign = require("dnadesign")
local align = dnadesign.align

describe("Alignment Examples", function()
    it("demonstrates Needleman-Wunsch alignment", function()
        local a = "GATTACA"
        local b = "GCATGCU"

        local sub_matrix = align.default_matrix
        local scoring = align.new_scoring(sub_matrix, -1)
        
        local score, align_a, align_b = align.needleman_wunsch(a, b, scoring)
        
        assert.are.equal(0, score)
        assert.are.equal("G-ATTACA", align_a)
        assert.are.equal("GCA-TGCU", align_b)
    end)

    it("demonstrates Smith-Waterman alignment", function()
        local a = "GATTACA"
        local b = "GCATGCU"

        local alphabet = {"A", "C", "G", "T", "U"}
        local sub_matrix = align.default_matrix
        local scoring = align.new_scoring(sub_matrix, -1)
        
        local score, align_a, align_b = align.smith_waterman(a, b, scoring)
        
        assert.are.equal(2, score)
        assert.are.equal("AT", align_a)
        assert.are.equal("AT", align_b)
    end)
end)
