-- examples/hash_examples.lua
local dnadesign = require("dnadesign")
local hash = dnadesign.hash

describe("Hash Examples", function()
    it("demonstrates basic CRC32 usage", function()
        -- Create a new CRC32 hash
        local h = hash.new_crc32()
        
        -- Hash a DNA sequence
        h:write("ATGCGATCGATCGATCG")
        
        -- Get the hash value
        local hash_value = h:sum()[1]
        
        -- You can reset and reuse the hash
        h:reset()
        h:write("Another sequence")
        local new_hash = h:sum()[1]
        
        -- Hashes of different inputs should be different
        assert.are_not.equal(hash_value, new_hash)
    end)
    
    it("demonstrates incremental hashing", function()
        -- Create two CRC32 hashes
        local h1 = hash.new_crc32()
        local h2 = hash.new_crc32()
        
        -- Hash the same content in different chunks
        h1:write("ATGC")
        h1:write("GATC")
        h1:write("GATCG")
        
        -- Hash the content all at once
        h2:write("ATGCGATCGATCG")
        
        -- The hashes should be identical
        assert.are.same(h1:sum(), h2:sum())
    end)

	it("demonstrates basic SHA256 usage", function()
        -- Create a new SHA256 hash
        local h = hash.new_sha256()

        -- Hash a DNA sequence
        h:write("ATGCGATCGATCGATCG")

        -- Get the hash value
        local hash_value = h:sum()

		-- Sums will be 32 int numbers that correspond to bytes
		assert.are.equal(#hash_value, 32)
    end)
    
end)
