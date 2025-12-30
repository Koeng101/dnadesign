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

describe("align_many Example", function()
    it("demonstrates finding best matches from multiple candidates", function()
        local target = "GATTACA"
        local candidates = {
            "GATTACA",    -- Perfect match
            "GCATGCT",    -- Some similarity
            "ACGT",       -- Short partial match
            "TTTTTTT"     -- Poor match
        }

        local substitution_matrix = {
            data = {
                ["-"] = {["-"] = 0, A = 0, C = 0, G = 0, T = 0},
                A = {["-"] = 0, A = 3, C = -3, G = -3, T = -3},
                C = {["-"] = 0, A = -3, C = 3, G = -3, T = -3},
                G = {["-"] = 0, A = -3, C = -3, G = 3, T = -3},
                T = {["-"] = 0, A = -3, C = -3, G = -3, T = 3}
            }
        }
        local scoring = align.new_scoring(substitution_matrix, -2)
        
        -- Get top 2 matches using Smith-Waterman local alignment
        local results = align.align_many(align.smith_waterman, target, candidates, scoring, 2)
        
        -- First result should be the perfect match
        assert.are.equal(21, results[1][1])  -- score: 7 chars * 3 points
        assert.are.equal("GATTACA", results[1][2])
        assert.are.equal("GATTACA", results[1][3])
		assert.are.equal(1, results[1][4]) -- The index in the candidates list are in [4]
        
        -- Second result should have lower score
        assert.is_true(results[2][1] < results[1][1])
    end)
end)
