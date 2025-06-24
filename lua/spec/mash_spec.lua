local dnadesign = require("dnadesign")
local mash = dnadesign.mash
local hash = dnadesign.hash

describe("Mash", function()
    it("should create identical sketches for identical sequences", function()
        local seq = "ATGCGATCGATCGATCGATCGATCGATCGATCGATCGATCGATCGATCGATCGATCGATCGA"
        
        local fp1 = mash.new(17, 10, hash.new_crc32())
        mash.sketch(fp1, seq)
        
        local fp2 = mash.new(17, 9, hash.new_crc32())
        mash.sketch(fp2, seq)
        
        assert.are.equal(0, mash.distance(fp1, fp2))
        assert.are.equal(0, mash.distance(fp2, fp1))
    end)
    
    it("should handle different sketch sizes", function()
        local seq1 = "ATGCGATCGATCGATCGATCGATCGATCGATCGATCGATCGATCGATCGATCGATCGATCGA"
        local seq2 = "ATCGATCGATCGATCGATCGATCGATCGATCGATCGAATGCGATCGATCGATCGATCGATCG"
        
        local fp1 = mash.new(17, 10, hash.new_crc32())
        mash.sketch(fp1, seq1)
        
        local fp2 = mash.new(17, 5, hash.new_crc32())
        mash.sketch(fp2, seq2)
        
        local distance = mash.distance(fp1, fp2)
        assert.is_true(distance > 0.19 and distance < 0.21)
    end)
    
    it("should detect completely different sequences", function()
        local h1 = hash.new_crc32()
        local h2 = hash.new_crc32()
        
        local fp1 = mash.new(17, 10, h1)
        mash.sketch(fp1, "ATGCGATCGATCGATCGATCG")
        
        local fp2 = mash.new(17, 10, h2)
		mash.sketch(fp2, "atgagtata")
        
        assert.are.equal(1, mash.distance(fp1, fp2))
    end)
    
    it("should handle empty sketches", function()
        local h1 = hash.new_crc32()
        local h2 = hash.new_crc32()
        
        local fp1 = mash.new(17, 10, h1)
        local fp2 = mash.new(17, 9, h2)
        
        assert.are.equal(1, mash.distance(fp1, fp2))
    end)
end)
