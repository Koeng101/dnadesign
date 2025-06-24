-- examples/transform_examples.lua
local dnadesign = require("dnadesign")
local transform = dnadesign.transform

describe("Transform Examples", function()
    it("demonstrates reverse complement", function()
        local sequence = "GATTACA"
        local reverse_complement = transform.reverse_complement(sequence)
        
        assert.are.equal("TGTAATC", reverse_complement)
    end)

    it("demonstrates complement", function()
        local sequence = "GATTACA"
        local complement = transform.complement(sequence)
        
        assert.are.equal("CTAATGT", complement)
    end)

    it("demonstrates reverse", function()
        local sequence = "GATTACA"
        local reverse = transform.reverse(sequence)
        
        assert.are.equal("ACATTAG", reverse)
    end)

    it("demonstrates RNA complement", function()
        local sequence = "GAUUACA"
        local complement = transform.complement_rna(sequence)
        
        assert.are.equal("CUAAUGU", complement)
    end)

    it("demonstrates RNA reverse complement", function()
        local sequence = "GAUUACA"
        local reverse_complement = transform.reverse_complement_rna(sequence)
        
        assert.are.equal("UGUAAUC", reverse_complement)
    end)
end)
