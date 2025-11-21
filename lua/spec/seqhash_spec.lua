local dnadesign = require("dnadesign")
local seqhash = dnadesign.seqhash

describe("seqhash", function()
    describe("hash2", function()
        it("should fail with invalid sequence type", function()
            local _, err = seqhash.hash2("ATGGGCTAA", "TNA", true, true)
            assert.is_not_nil(err)
        end)

        it("should fail with invalid DNA/RNA character", function()
            local _, err = seqhash.hash2("XTGGCCTAA", "DNA", true, true)
            assert.is_not_nil(err)
        end)

        it("should fail with invalid protein character", function()
            local _, err = seqhash.hash2("MGCJ*", "PROTEIN", false, false)
            assert.is_not_nil(err)
        end)

        it("should handle circular double stranded DNA correctly", function()
            local hash, err = seqhash.hash2("TTAGCCCAT", "DNA", true, true)
            local encoded, err = seqhash.encode_hash2(hash, err)
            assert.equal("A_6X4AWQfBTKsbdShxSfv5Am", encoded)
        end)

        it("should handle circular single stranded DNA correctly", function()
            local hash, err = seqhash.hash2("TTAGCCCAT", "DNA", true, false)
            local encoded, err = seqhash.encode_hash2(hash, err)
            assert.equal("B_63GUxYRGsH7kzSEbLh4KyG", encoded)
        end)

        it("should handle linear double stranded DNA correctly", function()
            local hash, err = seqhash.hash2("TTAGCCCAT", "DNA", false, true)
            local encoded, err = seqhash.encode_hash2(hash, err)
            assert.equal("C_5Z2pHCXbxWUPYiZj6J1Nag", encoded)
        end)

        it("should handle linear single stranded DNA correctly", function()
            local hash, err = seqhash.hash2("TTAGCCCAT", "DNA", false, false)
            local encoded, err = seqhash.encode_hash2(hash, err)
            assert.equal("D_4yT7etihWZHHNXUpbM5tUf", encoded)
        end)

        it("should handle RNA correctly", function()
            local hash, err = seqhash.hash2("TTAGCCCAT", "RNA", false, false)
            local encoded, err = seqhash.encode_hash2(hash, err)
            assert.equal("H_56cWv4dacvRJxUUcXYsdP5", encoded)
        end)

        it("should handle protein correctly", function()
            local hash, err = seqhash.hash2("MGC*", "PROTEIN", false, false)
            local encoded, err = seqhash.encode_hash2(hash, err)
            assert.equal("I_5DQsEyDHLh2r4njCcupAuF", encoded)
        end)
    end)

    describe("encode and decode", function()
        it("should encode and decode consistently", function()
            local raw_bytes, err = seqhash.hash2("ATGC", "DNA", false, true)
            assert.equal(err, "")
            
            local encoded, err = seqhash.encode_hash2(raw_bytes, "")
            assert.equal(err, "")
            
            local decoded, err = seqhash.decode_hash2(encoded)
            assert.equal(err, "")
            
            for i = 1, #raw_bytes do
                assert.equal(raw_bytes[i], decoded[i])
            end
        end)

        it("should fail on empty input", function()
            local _, err = seqhash.decode_hash2("")
            assert.is_not_nil(err)
        end)

        it("should fail on empty data", function()
            local _, err = seqhash.decode_hash2("A_")
            assert.is_not_nil(err)
        end)

        it("should fail on bad character", function()
            local _, err = seqhash.decode_hash2("A_/")
            assert.is_not_nil(err)
        end)

        it("should fail on wrong length", function()
            local _, err = seqhash.decode_hash2("A_11111")
            assert.is_not_nil(err)
        end)
    end)

    describe("sequence rotation", function()
        it("should find consistent least rotation", function()
            local test_sequences = {
                "AGCT",
                "GCTA",
                "CTAG",
                "TAGC"
            }
            
            local first_rotation = seqhash.rotate_sequence(test_sequences[1])
            for _, seq in ipairs(test_sequences) do
                local rotation = seqhash.rotate_sequence(seq)
                assert.equal(first_rotation, rotation)
            end
        end)
    end)

    describe("flag encoding", function()
        it("should encode and decode flags consistently", function()
            local version = 2
            local sequence_type = "DNA"
            local circularity = true
            local double_stranded = true
            
            local flag = seqhash.version2_flag(version, sequence_type, circularity, double_stranded)
            local decoded_version, decoded_sequence_type, decoded_circularity, decoded_double_stranded = 
                seqhash.decode_flag(flag)
            
            assert.equal(version, decoded_version)
            assert.equal(sequence_type, decoded_sequence_type)
            assert.equal(circularity, decoded_circularity)
            assert.equal(double_stranded, decoded_double_stranded)
        end)
    end)

    describe("hash2_fragment", function()
        it("should fail with invalid character", function()
            local _, err = seqhash.hash2_fragment("ATGGGCTAX", 4, 4)
            assert.is_not_nil(err)
        end)

        it("should hash fragments correctly", function()
            local hash, err = seqhash.hash2_fragment("ATGGGCTAA", 4, 4)
            local encoded, err = seqhash.encode_hash2(hash, err)
            assert.equal("K_5KnZQEnPRzJSYPkbPwLCJF", encoded)
        end)

        it("should hash reverse complement fragments consistently", function()
            local hash, err = seqhash.hash2_fragment("TTAGCCCAT", 4, 4)
            local encoded, err = seqhash.encode_hash2(hash, err)
            assert.equal("K_5KnZQEnPRzJSYPkbPwLCJF", encoded)
        end)
    end)

    describe("circular_equality", function()
        it("should return true for identical sequences", function()
            local result = seqhash.circular_equality("ATGC", "ATGC")
            assert.is_true(result)
        end)

        it("should return true for rotations of the same sequence", function()
            local result1 = seqhash.circular_equality("ATGC", "GCAT")
            local result2 = seqhash.circular_equality("ATGC", "CATG")
            local result3 = seqhash.circular_equality("ATGC", "TGCA")
            assert.is_true(result1)
            assert.is_true(result2)
            assert.is_true(result3)
        end)

        it("should return true for reverse complement", function()
            local result = seqhash.circular_equality("ATGC", "GCAT")
            assert.is_true(result)
        end)

        it("should return true for rotation of reverse complement", function()
            local result1 = seqhash.circular_equality("AAAA", "TTTT")
            assert.is_true(result1)
        end)

        it("should return false for completely different sequences", function()
            local result = seqhash.circular_equality("ATGC", "AAAA")
            assert.is_false(result)
        end)
    end)
end)
