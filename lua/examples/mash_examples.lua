local dnadesign = require("dnadesign")
local mash = dnadesign.mash
local hash = dnadesign.hash

describe("Mash Examples", function()
    it("demonstrates basic Mash sketching and comparison", function()
        -- Create two sequences that are similar but not identical
        local seq1 = "ATGCGATCGATCGATCGATCG"
        local seq2 = "ATGCGATCGATCGATCGTTCG"
        
        -- Create Mash sketches with kmer size 5 and sketch size 10, using CRC32
        local hasher = hash.new_crc32()
        local sketch1 = mash.new(5, 10, hasher)
        local sketch2 = mash.new(5, 10, hasher)
        
        -- Generate sketches
        mash.sketch(sketch1, seq1)
        mash.sketch(sketch2, seq2)
        
        -- Calculate distance
        local distance = mash.distance(sketch1, sketch2)
        
        -- The sequences are very similar, so distance should be small
        assert.is_true(distance < 0.5)
    end)
    
    it("demonstrates effect of sketch size on precision", function()
        local seq = "ATGCGATCGATCGATCGATCG"
        local hasher = hash.new_crc32()
        
        -- Create sketches with different sizes
        local small_sketch = mash.new(5, 5, hasher)
        local large_sketch = mash.new(5, 15, hasher)
        
        mash.sketch(small_sketch, seq)
        mash.sketch(large_sketch, seq)
        
        -- Even though they're from the same sequence, the different sketch
        -- sizes will affect the distance calculation
        local distance = mash.distance(small_sketch, large_sketch)
        
        -- But they should still be recognized as very similar
        assert.is_true(distance < 0.2)
    end)

	it("demonstrates screening multiple features against one plasmid", function()
        local featureA = "ATGCGATCGATCGATC"   -- present
        local featureB = "GGGAAACCCGGGAAAC"   -- absent
        local plasmid  = featureA .. "TTTTTT" -- only featureA is present

        local hasher = hash.new_crc32()

        local sketchA = mash.new_containment_sketch(6, featureA, hasher)
        local sketchB = mash.new_containment_sketch(6, featureB, hasher)
        local plasmid_sketch = mash.new_containment_sketch(6, plasmid, hasher)

        local cA = mash.containment(sketchA, plasmid_sketch)
        local cB = mash.containment(sketchB, plasmid_sketch)

        -- Feature A should be strongly contained
        assert.is_true(cA > 0.9)

        -- Feature B should not be contained
        assert.are.equal(0, cB)
    end)
end)
