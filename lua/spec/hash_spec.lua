-- hash_spec.lua
local dnadesign = require("dnadesign")
local hash = dnadesign.hash

describe("Bit manipulation", function()
    local bit32 = hash.bit32

    it("should perform bitwise AND correctly", function()
        assert.equal(0x0F, bit32.band(0xFF, 0x0F))
        assert.equal(0x00, bit32.band(0xF0, 0x0F))
        assert.equal(0xFF, bit32.band(0xFF, 0xFF))
    end)

    it("should perform bitwise XOR correctly", function()
        assert.equal(0xF0, bit32.bxor(0xFF, 0x0F))
        assert.equal(0xFF, bit32.bxor(0xF0, 0x0F))
        assert.equal(0x00, bit32.bxor(0xFF, 0xFF))
    end)

    it("should perform right shift correctly", function()
        assert.equal(0x0F, bit32.rshift(0xF0, 4))
        assert.equal(0x01, bit32.rshift(0x10, 4))
        assert.equal(0x00, bit32.rshift(0x0F, 4))
    end)

    it("should perform left shift correctly", function()
        assert.equal(0xF0, bit32.lshift(0x0F, 4))
        assert.equal(0x10, bit32.lshift(0x01, 4))
        assert.equal(0x00, bit32.lshift(0x00, 4))
    end)

    it("should perform bitwise OR correctly", function()
        assert.equal(0xFF, bit32.bor(0xF0, 0x0F))
        assert.equal(0xFF, bit32.bor(0xFF, 0x0F))
        assert.equal(0xFF, bit32.bor(0xFF, 0xFF))
    end)

    it("should perform bitwise NOT correctly", function()
        assert.equal(0xFFFFFFFF - 0xFF, bit32.bnot(0xFF))
        assert.equal(0xFFFFFFFF, bit32.bnot(0))
        assert.equal(0, bit32.bnot(0xFFFFFFFF))
    end)
end)

describe("Hash functions", function()
    describe("CRC32", function()
		it("should match known online CRC32 calculator values", function()
            local test_strings = {
                -- Values checked against https://crccalc.com/ using CRC-32 (IEEE 802.3)
                { 
                    input = "The quick brown fox jumps over the lazy dog", 
                    expected = {0x41, 0x4F, 0xA3, 0x39}
                },
                { 
                    input = "hello world", 
                    expected = {0x0D, 0x4A, 0x11, 0x85}
                },
                { 
                    input = "", 
                    expected = {0x00, 0x00, 0x00, 0x00}
                }
            }

            for _, test in ipairs(test_strings) do
                local h = hash.new_crc32()
                h:write(test.input)
                assert.are.same(test.expected, h:sum())
            end
        end)

        it("should produce consistent hashes", function()
            local h1 = hash.new_crc32()
            local h2 = hash.new_crc32()
            
            h1:write("hello")
            h2:write("hello")
            
            assert.are.same(h1:sum(), h2:sum())
        end)
        
        it("should produce different hashes for different inputs", function()
            local h1 = hash.new_crc32()
            local h2 = hash.new_crc32()
            
            h1:write("hello")
            h2:write("world")
            
            assert.are_not.same(h1:sum(), h2:sum())
        end)
        
        it("should handle empty strings", function()
            local h = hash.new_crc32()
            h:write("")
            local sum = h:sum()[1]
            assert.are.equal(0, sum)
        end)
        
        it("should reset properly", function()
            local h = hash.new_crc32()
            h:write("hello")
            local first_sum = h:sum()[1]
            
            h:reset()
            h:write("hello")
            local second_sum = h:sum()[1]
            
            assert.are.equal(first_sum, second_sum)
        end)
        
        it("should match known CRC32 values", function()
            local h = hash.new_crc32()
            h:write("The quick brown fox jumps over the lazy dog")
            local sum = h:sum()
            -- This is the IEEE CRC32 hash of the test string
			assert.are.same({0x41, 0x4F, 0xA3, 0x39}, sum)
        end)
    end)
	describe("SHA256", function()
        it("should match known SHA256 values", function()
            local test_strings = {
                {
                    input = "The quick brown fox jumps over the lazy dog",
                    expected = {
                        0xD7, 0xA8, 0xFB, 0xB3, 0x07, 0xD7, 0x80, 0x94,
                        0x69, 0xCA, 0x9A, 0xBC, 0xB0, 0x08, 0x2E, 0x4F,
                        0x8D, 0x56, 0x51, 0xE4, 0x6D, 0x3C, 0xDB, 0x76,
                        0x2D, 0x02, 0xD0, 0xBF, 0x37, 0xC9, 0xE5, 0x92
                    }
                },
                {
                    input = "ATGC",
                    expected = {
                        0x98, 0x20, 0xF5, 0xA8, 0x4C, 0xC4, 0x04, 0x33,
                        0x0E, 0x6D, 0x97, 0xBD, 0xE1, 0x3B, 0x58, 0x0F,
                        0xE9, 0xBF, 0x68, 0xB9, 0xCA, 0x09, 0x74, 0x62,
                        0x45, 0x93, 0xF6, 0x32, 0x45, 0x53, 0x08, 0x8C
                    }
                },
                {
                    input = "abc",
                    expected = {
                        0xBA, 0x78, 0x16, 0xBF, 0x8F, 0x01, 0xCF, 0xEA,
                        0x41, 0x41, 0x40, 0xDE, 0x5D, 0xAE, 0x22, 0x23,
                        0xB0, 0x03, 0x61, 0xA3, 0x96, 0x17, 0x7A, 0x9C,
                        0xB4, 0x10, 0xFF, 0x61, 0xF2, 0x00, 0x15, 0xAD
                    }
                }
            }

            for _, test in ipairs(test_strings) do
                local h = hash.new_sha256()
                h:write(test.input)
                assert.are.same(test.expected, h:sum(), 
                    string.format("failed on input: %s", test.input))
            end
        end)

	    it("should produce consistent hashes", function()
	        local h1 = hash.new_sha256()
	        local h2 = hash.new_sha256()
	        
	        h1:write("hello")
	        h2:write("hello")
	        
	        assert.are.same(h1:sum(), h2:sum())
	    end)
	    
	    it("should produce different hashes for different inputs", function()
	        local h1 = hash.new_sha256()
	        local h2 = hash.new_sha256()
	        
	        h1:write("hello")
	        h2:write("world")
	        
	        assert.are_not.same(h1:sum(), h2:sum())
	    end)
	    
	    it("should handle incremental updates", function()
	        local h1 = hash.new_sha256()
	        local h2 = hash.new_sha256()
	        
	        h1:write("hello ")
	        h1:write("world")
	        
	        h2:write("hello world")
	        
	        assert.are.same(h1:sum(), h2:sum())
	    end)
	    
	    it("should reset properly", function()
	        local h = hash.new_sha256()
	        h:write("hello")
	        local first_sum = h:sum()
	        
	        h:reset()
	        h:write("hello")
	        local second_sum = h:sum()
	        
	        assert.are.same(first_sum, second_sum)
	    end)
	end)
end)
