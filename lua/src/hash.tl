--[[ Module hash
Module hash contains implementations of common hash algorithms. While normally
these would be imported from the system (and would be much faster there), these
are to ensure cross-platform compatibility.

TODO: sha256 should be able to return more than 32 bits
]]

local record hash
    bit32: bit32
    new_crc32: function(): CRC32
    new_sha256: function(): SHA256
    Hash32: Hash32
end

-- Bit manipulation functions since bit32 isn't available
local record bit32
	band: function(a: integer, b: integer): integer
	bxor: function(a: integer, b: integer): integer
	rshift: function(a: integer, b: integer): integer
	lshift: function(a: integer, b: integer): integer
	bor: function(a: integer, b: integer): integer
	bnot: function(a: integer): integer
end

function bit32.band(a: integer, b: integer): integer
    local result = 0
    local bit = 1
    for _ = 0, 31 do
        if (math.floor(a % 2) == 1) and (math.floor(b % 2) == 1) then
            result = result + bit
        end
        a = math.floor(a / 2)
        b = math.floor(b / 2)
        bit = math.floor(bit * 2)
    end
    return math.floor(result)
end

function bit32.bxor(a: integer, b: integer): integer
    local result = 0
    local bit = 1
    for _ = 0, 31 do
        if math.floor(a % 2) ~= math.floor(b % 2) then
            result = result + bit
        end
        a = math.floor(a / 2)
        b = math.floor(b / 2)
        bit = math.floor(bit * 2)
    end
    return math.floor(result)
end

function bit32.rshift(a: integer, b: integer): integer
    if b < 0 then return bit32.lshift(a, -b) end
    return math.floor(a / math.floor(2 ^ b))
end

function bit32.lshift(a: integer, b: integer): integer
    if b < 0 then return bit32.rshift(a, -b) end
    return math.floor(a * math.floor(2 ^ b))
end

function bit32.bor(a: integer, b: integer): integer
    local result = 0
    local bit = 1
    for _ = 0, 31 do
        if (math.floor(a % 2) == 1) or (math.floor(b % 2) == 1) then
            result = result + bit
        end
        a = math.floor(a / 2)
        b = math.floor(b / 2)
        bit = math.floor(bit * 2)
    end
    return math.floor(result)
end

function bit32.bnot(a: integer): integer
    return math.floor((-1 - a) % math.floor(2^32))
end

-- Hash interface
local interface Hash32
    reset: function(Hash32)
    write: function(Hash32, string)
    sum: function(Hash32): {integer}
	sum32: function(Hash32): integer
end

-- CRC32 implementation
local record CRC32 is Hash32
    value: integer
end

local CRC32_TABLE: {integer} = {}

-- Initialize CRC32 lookup table
do
    for i = 0, 255 do
        local crc = i
        for _ = 1, 8 do
            if bit32.band(crc, 1) == 1 then
                crc = bit32.bxor(bit32.rshift(crc, 1), 0xEDB88320)
            else
                crc = bit32.rshift(crc, 1)
            end
        end
        CRC32_TABLE[i + 1] = crc
    end
end

local function new_crc32(): CRC32
    return {
        value = 0xFFFFFFFF,
        reset = function(self: CRC32)
            self.value = 0xFFFFFFFF
        end,
        write = function(self: CRC32, str: string)
            local crc = self.value
            for i = 1, #str do
                local byte = str:byte(i)
                crc = bit32.bxor(crc, byte)
                crc = bit32.bxor(bit32.rshift(crc, 8), CRC32_TABLE[bit32.band(crc, 0xFF) + 1])
            end
            self.value = crc
        end,
		sum = function(self: CRC32): {integer}
		    local value = bit32.bxor(self.value, 0xFFFFFFFF)
		    return {
		        bit32.band(bit32.rshift(value, 24), 0xFF),
		        bit32.band(bit32.rshift(value, 16), 0xFF),
		        bit32.band(bit32.rshift(value, 8), 0xFF),
		        bit32.band(value, 0xFF)
		    }
		end,
		sum32 = function(self: CRC32): integer
			return bit32.bxor(self.value, 0xFFFFFFFF)
		end
    }
end

-- Add these type declarations after the SHA256 record definition
local record SHA256State
    digest: {integer}
    count_lo: integer
    count_hi: integer
    data: {integer}
    local_: integer
    digestsize: integer
end

-- SHA256 implementation
local record SHA256 is Hash32
    h: {integer}
    length: integer
    buffer: string
	state: SHA256State
end

-- Constants
local SHA_BLOCKSIZE <const> = 64
--local SHA_DIGESTSIZE <const> = 32

-- Helper functions (keep the same implementations, just add type signatures)
local function ROR(x: integer, y: integer): integer
    y = bit32.band(y, 31)
    return bit32.band(bit32.bor(
        bit32.rshift(bit32.band(x, 0xffffffff), y),
        bit32.lshift(x, (32 - y))
    ), 0xffffffff)
end

local function Ch(x: integer, y: integer, z: integer): integer
    return bit32.bxor(z, bit32.band(x, bit32.bxor(y, z)))
end

local function Maj(x: integer, y: integer, z: integer): integer
    return bit32.bor(bit32.band(bit32.bor(x, y), z), bit32.band(x, y))
end

local function S(x: integer, n: integer): integer
    return ROR(x, n)
end

local function R(x: integer, n: integer): integer
    return bit32.rshift(bit32.band(x, 0xffffffff), n)
end

local function Sigma0(x: integer): integer
    return bit32.bxor(bit32.bxor(S(x, 2), S(x, 13)), S(x, 22))
end

local function Sigma1(x: integer): integer
    return bit32.bxor(bit32.bxor(S(x, 6), S(x, 11)), S(x, 25))
end

local function Gamma0(x: integer): integer
    return bit32.bxor(bit32.bxor(S(x, 7), S(x, 18)), R(x, 3))
end

local function Gamma1(x: integer): integer
    return bit32.bxor(bit32.bxor(S(x, 17), S(x, 19)), R(x, 10))
end

local function new_state(): SHA256State
    local state: SHA256State = {
        digest = {0,0,0,0,0,0,0,0},
        count_lo = 0,
        count_hi = 0,
        data = {},
        local_ = 0,
        digestsize = 0
    }
    for i = 1, SHA_BLOCKSIZE do
        state.data[i] = 0
    end
    return state
end

local K: {integer} = {
    0x428a2f98, 0x71374491, 0xb5c0fbcf, 0xe9b5dba5,
    0x3956c25b, 0x59f111f1, 0x923f82a4, 0xab1c5ed5,
    0xd807aa98, 0x12835b01, 0x243185be, 0x550c7dc3,
    0x72be5d74, 0x80deb1fe, 0x9bdc06a7, 0xc19bf174,
    0xe49b69c1, 0xefbe4786, 0x0fc19dc6, 0x240ca1cc,
    0x2de92c6f, 0x4a7484aa, 0x5cb0a9dc, 0x76f988da,
    0x983e5152, 0xa831c66d, 0xb00327c8, 0xbf597fc7,
    0xc6e00bf3, 0xd5a79147, 0x06ca6351, 0x14292967,
    0x27b70a85, 0x2e1b2138, 0x4d2c6dfc, 0x53380d13,
    0x650a7354, 0x766a0abb, 0x81c2c92e, 0x92722c85,
    0xa2bfe8a1, 0xa81a664b, 0xc24b8b70, 0xc76c51a3,
    0xd192e819, 0xd6990624, 0xf40e3585, 0x106aa070,
    0x19a4c116, 0x1e376c08, 0x2748774c, 0x34b0bcb5,
    0x391c0cb3, 0x4ed8aa4a, 0x5b9cca4f, 0x682e6ff3,
    0x748f82ee, 0x78a5636f, 0x84c87814, 0x8cc70208,
    0x90befffa, 0xa4506ceb, 0xbef9a3f7, 0xc67178f2
}

local function hash_transform(state: SHA256State): nil 
    local W: {integer} = {}
    
    -- Fill W
    for i = 0, 15 do
        W[i] = bit32.lshift(state.data[4*i + 1], 24) + 
               bit32.lshift(state.data[4*i + 2], 16) + 
               bit32.lshift(state.data[4*i + 3], 8) + 
               state.data[4*i + 4]
    end
    
    for i = 16, 63 do
        W[i] = bit32.band(Gamma1(W[i-2]) + W[i-7] + Gamma0(W[i-15]) + W[i-16], 0xffffffff)
    end
    
    local ss: {integer} = {}
    for i = 1, 8 do
        ss[i] = state.digest[i]
    end
    
    local function RND(a: integer, b: integer, c: integer, d: integer, 
                      e: integer, f: integer, g: integer, h: integer, 
                      i: integer, ki: integer): integer, integer
        local t0 = h + Sigma1(e) + Ch(e, f, g) + ki + W[i]
        local t1 = Sigma0(a) + Maj(a, b, c)
        d = bit32.band(d + t0, 0xffffffff)
        h = bit32.band(t0 + t1, 0xffffffff)
        return d, h
    end

    for i = 0, 63 do
        local j = i % 8
        if j == 0 then
            ss[4], ss[8] = RND(ss[1],ss[2],ss[3],ss[4],ss[5],ss[6],ss[7],ss[8], i, K[i+1])
        elseif j == 1 then
            ss[3], ss[7] = RND(ss[8],ss[1],ss[2],ss[3],ss[4],ss[5],ss[6],ss[7], i, K[i+1])
        elseif j == 2 then
            ss[2], ss[6] = RND(ss[7],ss[8],ss[1],ss[2],ss[3],ss[4],ss[5],ss[6], i, K[i+1])
        elseif j == 3 then
            ss[1], ss[5] = RND(ss[6],ss[7],ss[8],ss[1],ss[2],ss[3],ss[4],ss[5], i, K[i+1])
        elseif j == 4 then
            ss[8], ss[4] = RND(ss[5],ss[6],ss[7],ss[8],ss[1],ss[2],ss[3],ss[4], i, K[i+1])
        elseif j == 5 then
            ss[7], ss[3] = RND(ss[4],ss[5],ss[6],ss[7],ss[8],ss[1],ss[2],ss[3], i, K[i+1])
        elseif j == 6 then
            ss[6], ss[2] = RND(ss[3],ss[4],ss[5],ss[6],ss[7],ss[8],ss[1],ss[2], i, K[i+1])
        elseif j == 7 then
            ss[5], ss[1] = RND(ss[2],ss[3],ss[4],ss[5],ss[6],ss[7],ss[8],ss[1], i, K[i+1])
        end
    end

    for i = 1, 8 do
        state.digest[i] = bit32.band(state.digest[i] + ss[i], 0xffffffff)
    end
end

local function init_state(): SHA256State
    local state = new_state()
    state.digest = {
        0x6A09E667, 0xBB67AE85, 0x3C6EF372, 0xA54FF53A,
        0x510E527F, 0x9B05688C, 0x1F83D9AB, 0x5BE0CD19
    }
    state.count_lo = 0
    state.count_hi = 0
    state.local_ = 0
    state.digestsize = 32
    return state
end

local function update_state(state: SHA256State, buffer: string): nil
    local count = #buffer
    local buffer_idx = 1

    local clo = bit32.band(state.count_lo + bit32.lshift(count, 3), 0xffffffff)
    if clo < state.count_lo then
        state.count_hi = state.count_hi + 1
    end
    state.count_lo = clo
    state.count_hi = state.count_hi + bit32.rshift(count, 29)

    if state.local_ > 0 then
        local i = SHA_BLOCKSIZE - state.local_
        if i > count then i = count end

        for j = 0, i-1 do
            state.data[state.local_ + j + 1] = string.byte(buffer:sub(buffer_idx + j, buffer_idx + j))
        end

        count = count - i
        buffer_idx = buffer_idx + i
        state.local_ = state.local_ + i

        if state.local_ == SHA_BLOCKSIZE then
            hash_transform(state)
            state.local_ = 0
        else
            return
        end
    end

    while count >= SHA_BLOCKSIZE do
        for i = 1, SHA_BLOCKSIZE do
            state.data[i] = string.byte(buffer:sub(buffer_idx + i - 1, buffer_idx + i - 1))
        end
        count = count - SHA_BLOCKSIZE
        buffer_idx = buffer_idx + SHA_BLOCKSIZE
        hash_transform(state)
    end

    -- Handle remaining bytes
    for i = 1, count do
        state.data[state.local_ + i] = string.byte(buffer:sub(buffer_idx + i - 1, buffer_idx + i - 1))
    end
    state.local_ = state.local_ + count
end

local function final_state(state: SHA256State): {integer}
    local lo_bit_count = state.count_lo
    local hi_bit_count = state.count_hi
    local count = bit32.rshift(lo_bit_count, 3) % 64

    state.data[count + 1] = 0x80
    count = count + 1

    if count > SHA_BLOCKSIZE - 8 then
        while count < SHA_BLOCKSIZE do
            state.data[count + 1] = 0
            count = count + 1
        end
        hash_transform(state)
        count = 0
    end

    while count < SHA_BLOCKSIZE - 8 do
        state.data[count + 1] = 0
        count = count + 1
    end

    state.data[57] = bit32.rshift(hi_bit_count, 24) % 256
    state.data[58] = bit32.rshift(hi_bit_count, 16) % 256
    state.data[59] = bit32.rshift(hi_bit_count, 8) % 256
    state.data[60] = hi_bit_count % 256
    state.data[61] = bit32.rshift(lo_bit_count, 24) % 256
    state.data[62] = bit32.rshift(lo_bit_count, 16) % 256
    state.data[63] = bit32.rshift(lo_bit_count, 8) % 256
    state.data[64] = lo_bit_count % 256

    hash_transform(state)

    local digest = {}
    for i = 1, 8 do
		digest[i] = state.digest[i]
    end
    return digest
end

local function new_sha256(): SHA256
    local internal_state = init_state()
    
    return {
        h = internal_state.digest,
		state = internal_state,
        length = 0,
        buffer = "",
        
        reset = function(self: SHA256): nil
            internal_state = init_state()
            self.h = internal_state.digest
			self.state = internal_state
            self.length = 0
            self.buffer = ""
        end,

        write = function(self: SHA256, msg: string): nil 
            self.buffer = self.buffer .. msg
            self.length = self.length + #msg
            update_state(self.state, msg)
            self.h = self.state.digest
        end,

		sum = function(self: SHA256): {integer}
		    local digest = final_state(self.state)
		    local bytes = {}
		    for i = 1, 8 do
		        local value = digest[i]
		        bytes[4*i-3] = bit32.band(bit32.rshift(value, 24), 0xFF)
		        bytes[4*i-2] = bit32.band(bit32.rshift(value, 16), 0xFF)
		        bytes[4*i-1] = bit32.band(bit32.rshift(value, 8), 0xFF)
		        bytes[4*i] = bit32.band(value, 0xFF)
		    end
		    return bytes
		end
    }
end

hash.bit32 = bit32
hash.new_crc32 = new_crc32
hash.new_sha256 = new_sha256
hash.Hash32 = Hash32

return hash
