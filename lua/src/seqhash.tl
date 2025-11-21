--[[ Module seqhash
Module seqhash contains the seqhash algorithm.

This module contains the reference seqhash algorithm in Teal.

There is a big problem with current sequence databases - they all use different
identifiers and accession numbers. This means cross-referencing databases is
a complicated exercise, especially as the quantity of databases increases, or if
you need to compare "wild" DNA sequences.

Seqhash is a simple algorithm to produce consistent identifiers for any genetic sequence. The
basic premise of the Seqhash algorithm is to hash sequences with the hash being a robust
cross-database identifier. Sequences themselves shouldn't be used as a database index
(often, they're too big), so a hash based off of a sequence is the next best thing.

Usability wise, you should be able to Seqhash any rotation of a sequence in any direction and
get a consistent hash.

The Seqhash algorithm makes several opinionated design choices, primarily to make working
with Seqhashes more consistent and nice. The Seqhash algorithm only uses a single hash function,
SHA256 (different from Go implementation which uses Blake3), and only operates on DNA, RNA, and 
Protein sequences. These identifiers will be seen by human beings, so versioning and metadata 
is attached to the front of the hashes so that a human operator can quickly identify problems 
with hashing.

For DNA or RNA sequences, only ATUGCYRSWKMBDHVNZ characters are allowed. For Proteins,
only ACDEFGHIKLMNPQRSTVWYUO*BXZ characters are allowed in sequences. Selenocysteine (Sec; U) and pyrrolysine
(Pyl; O) are included in the protein character set - usually U and O don't occur within protein sequences,
but for certain organisms they do, and it is certainly a relevant amino acid for those particular proteins.
]]

local record seqhash
    version2_flag: function(version: integer, sequence_type: string, circularity: boolean, double_stranded: boolean): integer
    decode_flag: function(flag: integer): (integer, string, boolean, boolean)
    booth_least_rotation: function(sequence: string): integer
    rotate_sequence: function(sequence: string): string
	circular_equality: function(sequence1: string, sequence2: string): boolean
    hash2: function(sequence: string, sequence_type: string, circular: boolean, double_stranded: boolean): ({integer}, string)
    hash2_fragment: function(sequence: string, fwd_overhang_length: integer, rev_overhang_length: integer): ({integer}, string) 
    encode_hash2: function(hash: {integer}, err: string): (string, string)
    decode_hash2: function(encoded_string: string): ({integer}, string)
    encode_to_base58: function(input: {integer}): string
    decode_from_base58: function(input: string): ({integer}, string)

	hash2_metadata: {string:string}
end

-- Import required modules
local hash = require("hash") -- MODULE:hash
local transform = require("transform") -- MODULE:hash

-- Define the Sequence Type constants
local SEQ_TYPE_DNA: string = "DNA"
local SEQ_TYPE_RNA: string = "RNA" 
local SEQ_TYPE_PROTEIN: string = "PROTEIN"
local SEQ_TYPE_FRAGMENT: string = "FRAGMENT"

-- Define bit masks and shifts for flags
local HASH2_VERSION_MASK: integer = 0xF0 -- 11110000
local HASH2_CIRCULARITY_MASK: integer = 0x08 -- 00001000
local HASH2_DOUBLE_STRANDED_MASK: integer = 0x04 -- 00000100
local HASH2_TYPE_MASK: integer = 0x03 -- 00000011

local HASH2_VERSION_SHIFT: integer = 4
local HASH2_CIRCULARITY_SHIFT: integer = 3
local HASH2_DOUBLE_STRANDED_SHIFT: integer = 2

-- Base58 alphabet
local ALPHABET: string = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

--[[ SequenceType flags mapping ]]
local sequence_type_to_flag: {string:integer} = {
    [SEQ_TYPE_DNA] = 0x00,
    [SEQ_TYPE_RNA] = 0x01, 
    [SEQ_TYPE_PROTEIN] = 0x02,
    [SEQ_TYPE_FRAGMENT] = 0x03
}

local flag_to_sequence_type: {integer:string} = {
    [0x00] = SEQ_TYPE_DNA,
    [0x01] = SEQ_TYPE_RNA,
    [0x02] = SEQ_TYPE_PROTEIN, 
    [0x03] = SEQ_TYPE_FRAGMENT
}

-- Create a direct mapping of metadata combinations to letters
local function create_metadata_map(): {string:string}
    local m: {string:string} = {
        -- DNA
        ["DNA_circular_double"] = "A",
        ["DNA_circular_single"] = "B",
        ["DNA_linear_double"] = "C",
        ["DNA_linear_single"] = "D",
        -- RNA
        ["RNA_circular_double"] = "E",
        ["RNA_circular_single"] = "F",
        ["RNA_linear_double"] = "G",
        ["RNA_linear_single"] = "H",
        -- Protein
        ["PROTEIN_linear_single"] = "I",
        ["PROTEIN_circular_single"] = "J",
        -- Fragment
        ["FRAGMENT_linear_single"] = "K",
        ["FRAGMENT_circular_single"] = "L",
        ["FRAGMENT_linear_double"] = "M",
        ["FRAGMENT_circular_double"] = "N"
    }
	return m
end

seqhash.hash2_metadata = create_metadata_map()

--[[ Base58 encoding/decoding functions ]]
local function encode_to_base58(input: {integer}): string
    -- Input validation
    if #input == 0 then 
        return "1"
    end

    -- Count leading zeros
    local zeros = 0
    for i = 1, #input do
        if input[i] == 0 then
            zeros = zeros + 1
        else
            break
        end
    end
    
    -- Convert to regular bytes array
    local bytes: {integer} = {}
    for i = zeros + 1, #input do
        bytes[#bytes + 1] = input[i]
    end
    
    -- Convert to base58
    local result: {string} = {}
    while #bytes > 0 do
        local remainder = 0
        local new_bytes: {integer} = {}
        
        for i = 1, #bytes do
            local value = bytes[i]
            local temp = remainder * 256 + value
            remainder = temp % 58
            local quotient = math.floor(temp / 58)
            if #new_bytes > 0 or quotient > 0 then
                new_bytes[#new_bytes + 1] = quotient
            end
        end
        
        result[#result + 1] = ALPHABET:sub(remainder + 1, remainder + 1)
        bytes = new_bytes
    end
    
    -- Add leading '1's for zeros
    for _ = 1, zeros do
        result[#result + 1] = '1'
    end
    
    -- Reverse and concatenate
    local final: {string} = {}
    for i = #result, 1, -1 do
        final[#final + 1] = result[i]
    end
    
    return table.concat(final)
end

local function decode_from_base58(input: string): ({integer}, string)
    if #input == 0 then
        return {}, "empty input string"
    end
    
    -- Count leading '1's
    local zeros = 0
    for i = 1, #input do
        if input:sub(i,i) == '1' then
            zeros = zeros + 1
        else
            break
        end
    end
    
    -- Convert from base58 to array of base-256 values
    local bytes: {integer} = {0}  -- Start with single zero byte
    
    for i = zeros + 1, #input do
        local char_index = ALPHABET:find(input:sub(i,i))
        if not char_index then
            return {}, "invalid base58 character"
        end
        local digit = char_index - 1
        
        -- Multiply existing bytes by 58 and add digit
        local carry = digit
        for j = #bytes, 1, -1 do
            local x = bytes[j] * 58 + carry
            bytes[j] = x % 256
            carry = math.floor(x / 256)
        end
        
        while carry > 0 do
            table.insert(bytes, 1, carry % 256)
            carry = math.floor(carry / 256)
        end
    end
    
    -- Add leading zeros
    for _ = 1, zeros do
        table.insert(bytes, 1, 0)
    end
    
    return bytes, ""
end

--[[ 
EncodeFlag encodes the version, circularity, double-strandedness, and type into a single byte flag.
Used for seqhash v2
]]
local function version2_flag(version: integer, sequence_type: string, circularity: boolean, double_stranded: boolean): integer
    local flag: integer = 0
    flag = hash.bit32.bor(flag, hash.bit32.lshift(version, HASH2_VERSION_SHIFT))
    if circularity then
        flag = hash.bit32.bor(flag, hash.bit32.lshift(1, HASH2_CIRCULARITY_SHIFT))
    end
    if double_stranded then
        flag = hash.bit32.bor(flag, hash.bit32.lshift(1, HASH2_DOUBLE_STRANDED_SHIFT))
    end
    local type_flag = sequence_type_to_flag[sequence_type]
    flag = hash.bit32.bor(flag, hash.bit32.band(type_flag, HASH2_TYPE_MASK))
    return flag
end

--[[
DecodeFlag decodes the single byte flag into its constituent parts.
Outputs: version, sequence_type, circularity, doubleStranded.
Used for seqhash v2
]]
local function decode_flag(flag: integer): (integer, string, boolean, boolean)
    local version = hash.bit32.rshift(hash.bit32.band(flag, HASH2_VERSION_MASK), HASH2_VERSION_SHIFT)
    local circularity = hash.bit32.band(flag, HASH2_CIRCULARITY_MASK) ~= 0
    local double_stranded = hash.bit32.band(flag, HASH2_DOUBLE_STRANDED_MASK) ~= 0
    local seq_type = flag_to_sequence_type[hash.bit32.band(flag, HASH2_TYPE_MASK)]
    return version, seq_type, circularity, double_stranded
end

--[[ 
boothLeastRotation gets the least rotation of a circular string.
https://en.wikipedia.org/wiki/Lexicographically_minimal_string_rotation
]]
--local function booth_least_rotation(sequence: string): integer
--    local doubled = string.upper(sequence) .. string.upper(sequence)
--    local n = #doubled
--
--    local k = 0
--    
--    local f: {integer} = {}
--    for i = 0, n-1 do  -- Change to 0-based indexing
--        f[i] = -1
--    end
--
--    for j = 0, n-2 do  -- Change to 0-based indexing
--        local char = doubled:byte(j+2)  -- Adjust for 1-based string indexing
--        local failure = f[j-k]
--
--        while failure ~= -1 and char ~= doubled:byte(k+failure+2) do  -- Adjust for 1-based string indexing
--            if char < doubled:byte(k+failure+2) then
--                k = j - failure
--            end
--            failure = f[failure]
--        end
--
--        if char ~= doubled:byte(k+failure+2) then
--            if char < doubled:byte(k+2) then  -- Adjust for 1-based string indexing
--                k = j
--            end
--            f[j-k] = -1
--        else
--            f[j-k] = failure + 1
--        end
--    end
--
--    return k
--end
local function booth_least_rotation(seq:string): integer
    local s = string.upper(seq)
    local doubled = s .. s           -- work on s + s
    local n = #doubled               -- length of doubled string

    -- failure function f[0..n-1], 0-based
    local f = {}
    for i = 0, n - 1 do
        f[i] = -1
    end

    local k = 0   -- 0-based index of current best rotation in doubled

    -- j corresponds to 0-based j in the textbook algorithm, but we loop from 1
    -- because j = 0 case is trivial.
    for j = 1, n - 1 do
        -- s[j] in 0-based => doubled:byte(j+1) in Lua
        local sj = doubled:byte(j + 1)

        -- i = f[j - k - 1]  (this index is always >= 0 if invariant k < j holds)
        local i = f[j - k - 1]

        -- while i != -1 and s[j] != s[k + i + 1]
        while i ~= -1 and sj ~= doubled:byte(k + i + 2) do
            if sj < doubled:byte(k + i + 2) then
                k = j - i - 1
            end
            i = f[i]
        end

        if sj ~= doubled:byte(k + i + 2) then
            -- if s[j] < s[k]
            if sj < doubled:byte(k + 1) then
                k = j
            end
            f[j - k] = -1
        else
            f[j - k] = i + 1
        end
    end

    -- k is an index into doubled; reduce to index in original seq (length m)
    local m = #s
    if k >= m then
        k = k - m
    end
    return k  -- 0-based starting index of minimal rotation in original string
end



--[[ rotate_sequence rotates circular sequences to deterministic point ]]
local function rotate_sequence(sequence: string): string
    local k = booth_least_rotation(sequence)
    -- Since k is now 0-based but string:sub is 1-based, add 1 to k+1
    return sequence:sub(k+2) .. sequence:sub(1, k+1)
end

--[[ circular_equality checks if two circular sequences are the same. We assume
double stranded sequence.]]
local function circular_equality(sequence1: string, sequence2: string): boolean
	local sequence1U = string.upper(sequence1)
	local sequence2U = string.upper(sequence2)
	local deterministic_sequence1: string
	local seq1 = rotate_sequence(sequence1U)
    local rev_comp = transform.reverse_complement(sequence1U)
    local seq2 = rotate_sequence(rev_comp)
    deterministic_sequence1 = seq1 < seq2 and seq1 or seq2

	local deterministic_sequence2: string
    local seq2_1 = rotate_sequence(sequence2U)
    local rev_comp_2 = transform.reverse_complement(sequence2U)
    local seq2_2 = rotate_sequence(rev_comp_2)
    deterministic_sequence2 = seq2_1 < seq2_2 and seq2_1 or seq2_2

	return deterministic_sequence1 == deterministic_sequence2
end

--[[
Hash2 creates a version 2 seqhash.
]]
local function hash2(sequence: string, sequence_type: string, circular: boolean, double_stranded: boolean): ({integer}, string)
    -- Validate sequence type
    if sequence_type ~= SEQ_TYPE_DNA and sequence_type ~= SEQ_TYPE_RNA and sequence_type ~= SEQ_TYPE_PROTEIN then
        return {}, "Invalid sequence type"
    end
    
    -- Convert to uppercase and handle RNA
    sequence = string.upper(sequence)
    if sequence_type == SEQ_TYPE_RNA then
        sequence = string.gsub(sequence, "U", "T")
    end
    
    -- Validate sequence characters
    if sequence_type == SEQ_TYPE_DNA or sequence_type == SEQ_TYPE_RNA then
        for i = 1, #sequence do
            local char = sequence:sub(i,i)
            if not string.match("ATUGCYRSWKMBDHVNZ", char) then
                return {}, "Invalid DNA/RNA character: " .. char
            end
        end
    elseif sequence_type == SEQ_TYPE_PROTEIN then
        for i = 1, #sequence do
            local char = sequence:sub(i,i)
            if not string.match("ACDEFGHIKLMNPQRSTVWYUO*BXZ", char) then
                return {}, "Invalid protein character: " .. char
            end
        end
    end
    
	-- Get deterministic sequence
	local deterministic_sequence: string
	if circular and double_stranded then
	    local seq1 = rotate_sequence(sequence)
	    local rev_comp = transform.reverse_complement(sequence)
	    local seq2 = rotate_sequence(rev_comp)
	    deterministic_sequence = seq1 < seq2 and seq1 or seq2
	elseif circular then
	    deterministic_sequence = rotate_sequence(sequence)
	elseif double_stranded then
	    local seq2 = transform.reverse_complement(sequence)
	    deterministic_sequence = sequence < seq2 and sequence or seq2
	else
	    deterministic_sequence = sequence
	end
        
	-- Initialize result array
    local result = {}
    for i = 1, 16 do
        result[i] = 0
    end

	-- Set the flag byte
    result[1] = version2_flag(2, sequence_type, circular, double_stranded)
    
    -- Calculate SHA256 hash
    local h = hash.new_sha256()
    h:write(deterministic_sequence)
    local sha_bytes = h:sum()
    
    -- Copy exactly 15 bytes from the SHA256 hash into result (bytes 2-16)
	for i = 2, 16 do
		result[i] = sha_bytes[i-1]
	end
    
    return result, ""
end

--[[
Hash2Fragment creates a version 2 fragment seqhash. Fragment seqhashes are
a special kind of seqhash that are used to identify fragments, usually
released by restriction enzyme digestion, rather than complete DNA
sequences. This is very useful for tracking genetic parts in a database: as
abstractions away from their container vectors, so that many fragments in
different vectors can be identified consistently.

fwdOverhangLength and revOverhangLength are the lengths of both overhangs.
Hashed sequences are hashed with their overhangs attached. Most of the time,
both of these will equal 4, as they are released by TypeIIS restriction
enzymes.

In order to make sure fwdOverhangLength and revOverhangLength fit in the
hash, the hash is truncated at 13 bytes rather than 16, and both int8 are
inserted. So the bytes would be:

    flag + fwdOverhangLength + revOverhangLength + [13]byte(hash)

The fwdOverhang and revOverhang are the lengths of the overhangs of the
input sequence. The hash, however, contains the forward and reverse overhang
lengths of the deterministic sequence - ie, the alphabetically less-than
strand, when comparing the uppercase forward and reverse complement strand.
This means if the input sequence is not less than its reverse complement (for
example, GTT is greater than AAC), then the output hash will have the forward
and reverse overhang lengths of the reverse complement, not the input strand.
]]
local function hash2_fragment(sequence: string, fwd_overhang_length: integer, rev_overhang_length: integer): ({integer}, string)
    -- Validate sequence characters
    sequence = string.upper(sequence)
    for i = 1, #sequence do
        local char = sequence:sub(i,i)
        if not string.match("ATUGCYRSWKMBDHVNZ", char) then
            return {}, "Invalid DNA/RNA character: " .. char
        end
    end
    
    -- Get deterministic sequence and overhang lengths
    local forward, reverse = fwd_overhang_length, rev_overhang_length
    local deterministic_sequence = sequence
    local reverse_complement = transform.reverse_complement(sequence)
    
    if sequence > reverse_complement then
        forward = rev_overhang_length
        reverse = fwd_overhang_length
        deterministic_sequence = reverse_complement
    end
    
	-- Initialize result array
    local result = {}
    for i = 1, 16 do
        result[i] = 0
    end
    
    -- Set flag and overhang lengths
    result[1] = version2_flag(2, SEQ_TYPE_FRAGMENT, false, false)
    result[2] = forward
    result[3] = reverse
    
    -- Calculate SHA256 hash
    local h = hash.new_sha256()
    h:write(deterministic_sequence)
    local sha_bytes = h:sum()
    
	-- Copy exactly 13 bytes from the SHA256 hash into result (bytes 4-16)
	for i = 4, 16 do
	    result[i] = sha_bytes[i-3]
	end

    return result, ""
end

--[[
EncodeHash2 encodes Hash2 as a base58 string. It also adds a single
letter metadata tag that can be used as an easy heuristic for an LLM to
identify misbehaving code.
]]
local function encode_hash2(hash_to_encode: {integer}, err: string): (string, string)
    if err ~= "" then
        return "", err
    end

    local _, sequence_type, circularity, double_stranded = decode_flag(hash_to_encode[1])
    local metadata_key = sequence_type .. "_" ..
                        (circularity and "circular" or "linear") .. "_" ..
                        (double_stranded and "double" or "single")

    local metadata = seqhash.hash2_metadata[metadata_key]
    if not metadata then
        return "", "invalid metadata combination"
    end

    return metadata .. "_" .. encode_to_base58(hash_to_encode), ""
end

--[[ DecodeHash2 decodes a seqhash into a [16]byte, including the metadata tag ]]
local function decode_hash2(encoded_string: string): ({integer}, string)
    local parts = string.gmatch(encoded_string, "[^_]+")
    local metadata = parts()
    local hash_part = parts()
    
    if not metadata or not hash_part then
        return {}, "invalid encoded string format"
    end
    
    local decoded_bytes, err = decode_from_base58(hash_part)
    if err ~= "" then
        return {}, err
    end
    
    if #decoded_bytes ~= 16 then
        return {}, "decoded hash does not match expected length"
    end
    
    return decoded_bytes, ""
end

-- Create and return the seqhash interface
seqhash.version2_flag = version2_flag
seqhash.decode_flag = decode_flag
seqhash.booth_least_rotation = booth_least_rotation
seqhash.rotate_sequence = rotate_sequence
seqhash.circular_equality = circular_equality
seqhash.hash2 = hash2
seqhash.hash2_fragment = hash2_fragment
seqhash.encode_hash2 = encode_hash2
seqhash.decode_hash2 = decode_hash2
seqhash.encode_to_base58 = encode_to_base58
seqhash.decode_from_base58 = decode_from_base58

return seqhash 
