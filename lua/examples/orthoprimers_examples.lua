-- examples/orthoprimers_examples.lua
local dnadesign = require("dnadesign")
local orthoprimers = dnadesign.orthoprimers

describe("OrthoPrimers Examples", function()
    it("demonstrates basic primer set usage", function()
        -- Create a default set of 96 orthogonal primers
        local ops = orthoprimers.new_default_orthogonal_primer_set()
        
        -- Generate primer pairs for different samples
        local fwd1, rev1, err1 = ops:new_primer_set()
        local fwd2, rev2, err2 = ops:new_primer_set()
        local fwd3, rev3, err3 = ops:new_primer_set()
        
        -- Check we got valid primers
        assert.is_nil(err1)
        assert.is_nil(err2) 
        assert.is_nil(err3)
        
        -- Primers should be different
        assert.is_not.equal(fwd1, fwd2)
        assert.is_not.equal(rev1, rev2)
        
        -- Example outputs
        assert.are.equal("AAACACGTGGCAAACATTCC", fwd1)
        assert.are.equal("AAACCGGAGCCATACAGTAC", rev1)
    end)
    
    it("demonstrates custom primer set", function()
        -- Create a smaller custom set for testing
        local custom_primers = {
            "AAACACGTGGCAAACATTCC",
            "AAACCGGAGCCATACAGTAC", 
            "AAAGCACTCTTAGGCCTCTG",
            "AAAGGGGCCGTCAATATCAG"
        }
        local ops = orthoprimers.new_orthogonal_primer_set(custom_primers)
        
        -- Generate a primer pair
        local fwd, rev, err = ops:new_primer_set()
        
        assert.is_nil(err)
        assert.are.equal("AAACACGTGGCAAACATTCC", fwd)
        assert.are.equal("AAACCGGAGCCATACAGTAC", rev)
    end)
    
    it("demonstrates primer exhaustion", function()
        -- Small set will run out of pairs quickly
        local small_set = {
            "AAACACGTGGCAAACATTCC",
            "AAACCGGAGCCATACAGTAC"
        }
        local ops = orthoprimers.new_orthogonal_primer_set(small_set)
        
        -- Use the only possible pair
        local fwd, rev, err = ops:new_primer_set()
        assert.is_nil(err)
        
        -- Second attempt should fail
        local _, _, err2 = ops:new_primer_set()
        assert.are.equal("Not enough primers for genes in pool", err2)
    end)
end)
