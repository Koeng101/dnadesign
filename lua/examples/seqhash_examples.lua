local dnadesign = require("dnadesign")
local seqhash = dnadesign.seqhash

describe("Seqhash Examples", function()
    it("demonstrates basic sequence hashing", function()
        -- Hash a DNA sequence
        local sequence = "TTAGCCCAT"
        local hash, err = seqhash.hash2(sequence, "DNA", false, true)
        local encoded, _ = seqhash.encode_hash2(hash, "")
        
        -- Result: "C_5Z2pHCXbxWUPYiZj6J1Nag"
        -- The prefix indicates: C = linear double stranded DNA
        assert.equal("C_5Z2pHCXbxWUPYiZj6J1Nag", encoded)
    end)
    
    it("demonstrates different sequence types", function()
        local sequence = "TTAGCCCAT"
        
        -- Circular double stranded DNA (like plasmids)
        -- Prefix A = circular double stranded
        local hash_plasmid, _ = seqhash.hash2(sequence, "DNA", true, true)
        local encoded_plasmid, _ = seqhash.encode_hash2(hash_plasmid, "")
        assert.equal("A_6X4AWQfBTKsbdShxSfv5Am", encoded_plasmid)
        
        -- Linear single stranded DNA (like primers)
        -- Prefix D = linear single stranded
        local hash_primer, _ = seqhash.hash2(sequence, "DNA", false, false)
        local encoded_primer, _ = seqhash.encode_hash2(hash_primer, "")
        assert.equal("D_4yT7etihWZHHNXUpbM5tUf", encoded_primer)
        
        -- RNA sequences
        -- Prefix H = RNA
        local hash_rna, _ = seqhash.hash2(sequence, "RNA", false, false)
        local encoded_rna, _ = seqhash.encode_hash2(hash_rna, "")
        assert.equal("H_56cWv4dacvRJxUUcXYsdP5", encoded_rna)
        
        -- Protein sequences
        -- Prefix I = protein
        local hash_protein, _ = seqhash.hash2("MGC*", "PROTEIN", false, false)
        local encoded_protein, _ = seqhash.encode_hash2(hash_protein, "")
        assert.equal("I_5DQsEyDHLh2r4njCcupAuF", encoded_protein)
    end)
    
    it("demonstrates circular sequences hash identically", function()
        -- Rotations of the same circular sequence produce the same hash
        local hash1, _ = seqhash.hash2("ATGC", "DNA", true, true)
        local hash2, _ = seqhash.hash2("GCAT", "DNA", true, true)  -- rotation
        
        local encoded1, _ = seqhash.encode_hash2(hash1, "")
        local encoded2, _ = seqhash.encode_hash2(hash2, "")
        
        -- These will be identical!
        assert.equal(encoded1, encoded2)
    end)
    
    it("demonstrates circular equality checking", function()
        -- Quick way to check if sequences are circularly equivalent
        assert.is_true(seqhash.circular_equality("ATGC", "GCAT"))
        
        -- Also handles reverse complements
        assert.is_true(seqhash.circular_equality("AAAA", "TTTT"))
    end)
    
    it("demonstrates fragment hashing", function()
        -- Hash fragments for Golden Gate assembly design
        -- Prefix K = fragment hash
        local sequence = "ATGGGCTAA"
        local hash, _ = seqhash.hash2_fragment(sequence, 4, 4)
        local encoded, _ = seqhash.encode_hash2(hash, "")
        
        assert.equal("K_5KnZQEnPRzJSYPkbPwLCJF", encoded)
        
        -- Reverse complement produces the same fragment hash
        local rc_hash, _ = seqhash.hash2_fragment("TTAGCCCAT", 4, 4)
        local rc_encoded, _ = seqhash.encode_hash2(rc_hash, "")
        assert.equal(encoded, rc_encoded)
    end)
    
    it("demonstrates decoding hashes", function()
        -- You can decode a hash string back to raw bytes
        local encoded = "C_5Z2pHCXbxWUPYiZj6J1Nag"
        local decoded, err = seqhash.decode_hash2(encoded)
        
        assert.equal("", err)
        assert.is_truthy(#decoded > 0)
    end)
end)
