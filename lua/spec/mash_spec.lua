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

	it("should build a containment sketch with unique kmers only", function()
        -- Sequence "AAAAAA" with k=3 has 4 k-mers (AAA, AAA, AAA, AAA)
        -- but only 1 unique k-mer.
        local seq = "AAAAAA"
        local fp = mash.new_containment_sketch(3, seq, hash.new_crc32())

        -- Sketch size should equal number of unique k-mers, not total windows
        assert.are.equal(1, fp.sketch_size)
    end)

    it("should report full containment for identical sequences", function()
        local seq = "ATGCGATCGATCGATC"
        local fp1 = mash.new_containment_sketch(5, seq, hash.new_crc32())
        local fp2 = mash.new_containment_sketch(5, seq, hash.new_crc32())

        -- containment(a, b) with identical sequences should be 1 in both directions
        assert.are.equal(1, mash.containment(fp1, fp2))
        assert.are.equal(1, mash.containment(fp2, fp1))
    end)

    it("should detect feature contained in a larger plasmid", function()
        local feature_seq = "ATGCGATCGATCGATC"
        local plasmid_seq = feature_seq .. "TTTTTT"  -- plasmid strictly contains the feature

        local feature_fp = mash.new_containment_sketch(5, feature_seq, hash.new_crc32())
        local plasmid_fp = mash.new_containment_sketch(5, plasmid_seq, hash.new_crc32())

        local feature_in_plasmid = mash.containment(feature_fp, plasmid_fp)
        local plasmid_in_feature = mash.containment(plasmid_fp, feature_fp)

        -- Feature should be fully contained in plasmid
        assert.are.equal(1, feature_in_plasmid)

        -- But plasmid is larger, so its containment in the feature must be < 1
        assert.is_true(plasmid_in_feature < 1)
    end)

    it("should report zero containment when there are no shared kmers", function()
        -- Choose sequences with disjoint k-mers for k=3
        local seq1 = "AAAAAA"
        local seq2 = "CCCCCC"

        local fp1 = mash.new_containment_sketch(3, seq1, hash.new_crc32())
        local fp2 = mash.new_containment_sketch(3, seq2, hash.new_crc32())

        assert.are.equal(0, mash.containment(fp1, fp2))
        assert.are.equal(0, mash.containment(fp2, fp1))
    end)
end)
