local _tl_compat; if (tonumber((_VERSION or ''):match('[%d.]*$')) or 0) < 5.3 then local p, m = pcall(require, 'compat53.module'); if p then _tl_compat = m end end; local ipairs = _tl_compat and _tl_compat.ipairs or ipairs; local math = _tl_compat and _tl_compat.math or math; local pairs = _tl_compat and _tl_compat.pairs or pairs; local string = _tl_compat and _tl_compat.string or string; local table = _tl_compat and _tl_compat.table or table

























































local rng = {}
















local m1 = 2147483647
local m2 = 2147483399

local a1 = 48271
local a2 = 40692




local seed1 = (_G.SEED_1) or 12345
local seed2 = (_G.SEED_2) or 67890


function rng.new(seed_x, seed_y)
   return {
      x = seed_x % m1,
      y = seed_y % m2,
   }
end


function rng.next(state)
   local new_x = (state.x * a1) % m1
   local new_y = (state.y * a2) % m2


   local result = (new_x - new_y) % m1
   if result <= 0 then
      result = result + m1 - 1
   end

   return result, { x = new_x, y = new_y }
end


function rng.random(state, min, max)
   local result, new_state = rng.next(state)


   if not min and not max then

      return result / m1, new_state
   elseif not max then

      max = min
      min = 1
   end

   return min + (result % (max - min + 1)), new_state
end


rng.state = rng.new(seed1, seed2)


function rng.random_wrapper(m, n)
   if m == nil and n == nil then

      local result, new_state = rng.random(rng.state)
      rng.state = new_state
      return result
   elseif n == nil then

      local result, new_state = rng.random(rng.state, 1, m)
      rng.state = new_state
      return result
   else

      local result, new_state = rng.random(rng.state, m, n)
      rng.state = new_state
      return result
   end
end


_G.math.random = rng.random_wrapper


_G.math.randomseed = function(_)
   error("math.randomseed is disabled in hermetic environment. Use SEED_1 and SEED_2 environment variables instead.")
   return 0, 0
end












local hash = {}







local bit32 = {}








function bit32.band(a, b)
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

function bit32.bxor(a, b)
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

function bit32.rshift(a, b)
   if b < 0 then return bit32.lshift(a, -b) end
   return math.floor(a / math.floor(2 ^ b))
end

function bit32.lshift(a, b)
   if b < 0 then return bit32.rshift(a, -b) end
   return math.floor(a * math.floor(2 ^ b))
end

function bit32.bor(a, b)
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

function bit32.bnot(a)
   return math.floor((-1 - a) % math.floor(2 ^ 32))
end














local CRC32_TABLE = {}


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

local function new_crc32()
   return {
      value = 0xFFFFFFFF,
      reset = function(self)
         self.value = 0xFFFFFFFF
      end,
      write = function(self, str)
         local crc = self.value
         for i = 1, #str do
            local byte = str:byte(i)
            crc = bit32.bxor(crc, byte)
            crc = bit32.bxor(bit32.rshift(crc, 8), CRC32_TABLE[bit32.band(crc, 0xFF) + 1])
         end
         self.value = crc
      end,
      sum = function(self)
         local value = bit32.bxor(self.value, 0xFFFFFFFF)
         return {
            bit32.band(bit32.rshift(value, 24), 0xFF),
            bit32.band(bit32.rshift(value, 16), 0xFF),
            bit32.band(bit32.rshift(value, 8), 0xFF),
            bit32.band(value, 0xFF),
         }
      end,
      sum32 = function(self)
         return bit32.bxor(self.value, 0xFFFFFFFF)
      end,
   }
end




















local SHA_BLOCKSIZE = 64



local function ROR(x, y)
   y = bit32.band(y, 31)
   return bit32.band(bit32.bor(
   bit32.rshift(bit32.band(x, 0xffffffff), y),
   bit32.lshift(x, (32 - y))),
   0xffffffff)
end

local function Ch(x, y, z)
   return bit32.bxor(z, bit32.band(x, bit32.bxor(y, z)))
end

local function Maj(x, y, z)
   return bit32.bor(bit32.band(bit32.bor(x, y), z), bit32.band(x, y))
end

local function S(x, n)
   return ROR(x, n)
end

local function R(x, n)
   return bit32.rshift(bit32.band(x, 0xffffffff), n)
end

local function Sigma0(x)
   return bit32.bxor(bit32.bxor(S(x, 2), S(x, 13)), S(x, 22))
end

local function Sigma1(x)
   return bit32.bxor(bit32.bxor(S(x, 6), S(x, 11)), S(x, 25))
end

local function Gamma0(x)
   return bit32.bxor(bit32.bxor(S(x, 7), S(x, 18)), R(x, 3))
end

local function Gamma1(x)
   return bit32.bxor(bit32.bxor(S(x, 17), S(x, 19)), R(x, 10))
end

local function new_state()
   local state = {
      digest = { 0, 0, 0, 0, 0, 0, 0, 0 },
      count_lo = 0,
      count_hi = 0,
      data = {},
      local_ = 0,
      digestsize = 0,
   }
   for i = 1, SHA_BLOCKSIZE do
      state.data[i] = 0
   end
   return state
end

local K = {
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
   0x90befffa, 0xa4506ceb, 0xbef9a3f7, 0xc67178f2,
}

local function hash_transform(state)
   local W = {}


   for i = 0, 15 do
      W[i] = bit32.lshift(state.data[4 * i + 1], 24) +
      bit32.lshift(state.data[4 * i + 2], 16) +
      bit32.lshift(state.data[4 * i + 3], 8) +
      state.data[4 * i + 4]
   end

   for i = 16, 63 do
      W[i] = bit32.band(Gamma1(W[i - 2]) + W[i - 7] + Gamma0(W[i - 15]) + W[i - 16], 0xffffffff)
   end

   local ss = {}
   for i = 1, 8 do
      ss[i] = state.digest[i]
   end

   local function RND(a, b, c, d,
      e, f, g, h,
      i, ki)
      local t0 = h + Sigma1(e) + Ch(e, f, g) + ki + W[i]
      local t1 = Sigma0(a) + Maj(a, b, c)
      d = bit32.band(d + t0, 0xffffffff)
      h = bit32.band(t0 + t1, 0xffffffff)
      return d, h
   end

   for i = 0, 63 do
      local j = i % 8
      if j == 0 then
         ss[4], ss[8] = RND(ss[1], ss[2], ss[3], ss[4], ss[5], ss[6], ss[7], ss[8], i, K[i + 1])
      elseif j == 1 then
         ss[3], ss[7] = RND(ss[8], ss[1], ss[2], ss[3], ss[4], ss[5], ss[6], ss[7], i, K[i + 1])
      elseif j == 2 then
         ss[2], ss[6] = RND(ss[7], ss[8], ss[1], ss[2], ss[3], ss[4], ss[5], ss[6], i, K[i + 1])
      elseif j == 3 then
         ss[1], ss[5] = RND(ss[6], ss[7], ss[8], ss[1], ss[2], ss[3], ss[4], ss[5], i, K[i + 1])
      elseif j == 4 then
         ss[8], ss[4] = RND(ss[5], ss[6], ss[7], ss[8], ss[1], ss[2], ss[3], ss[4], i, K[i + 1])
      elseif j == 5 then
         ss[7], ss[3] = RND(ss[4], ss[5], ss[6], ss[7], ss[8], ss[1], ss[2], ss[3], i, K[i + 1])
      elseif j == 6 then
         ss[6], ss[2] = RND(ss[3], ss[4], ss[5], ss[6], ss[7], ss[8], ss[1], ss[2], i, K[i + 1])
      elseif j == 7 then
         ss[5], ss[1] = RND(ss[2], ss[3], ss[4], ss[5], ss[6], ss[7], ss[8], ss[1], i, K[i + 1])
      end
   end

   for i = 1, 8 do
      state.digest[i] = bit32.band(state.digest[i] + ss[i], 0xffffffff)
   end
end

local function init_state()
   local state = new_state()
   state.digest = {
      0x6A09E667, 0xBB67AE85, 0x3C6EF372, 0xA54FF53A,
      0x510E527F, 0x9B05688C, 0x1F83D9AB, 0x5BE0CD19,
   }
   state.count_lo = 0
   state.count_hi = 0
   state.local_ = 0
   state.digestsize = 32
   return state
end

local function update_state(state, buffer)
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

      for j = 0, i - 1 do
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


   for i = 1, count do
      state.data[state.local_ + i] = string.byte(buffer:sub(buffer_idx + i - 1, buffer_idx + i - 1))
   end
   state.local_ = state.local_ + count
end

local function final_state(state)
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

local function new_sha256()
   local internal_state = init_state()

   return {
      h = internal_state.digest,
      state = internal_state,
      length = 0,
      buffer = "",

      reset = function(self)
         internal_state = init_state()
         self.h = internal_state.digest
         self.state = internal_state
         self.length = 0
         self.buffer = ""
      end,

      write = function(self, msg)
         self.buffer = self.buffer .. msg
         self.length = self.length + #msg
         update_state(self.state, msg)
         self.h = self.state.digest
      end,

      sum = function(self)
         local digest = final_state(self.state)
         local bytes = {}
         for i = 1, 8 do
            local value = digest[i]
            bytes[4 * i - 3] = bit32.band(bit32.rshift(value, 24), 0xFF)
            bytes[4 * i - 2] = bit32.band(bit32.rshift(value, 16), 0xFF)
            bytes[4 * i - 1] = bit32.band(bit32.rshift(value, 8), 0xFF)
            bytes[4 * i] = bit32.band(value, 0xFF)
         end
         return bytes
      end,
   }
end

hash.bit32 = bit32
hash.new_crc32 = new_crc32
hash.new_sha256 = new_sha256
hash.Hash32 = Hash32







local transform = {}
















local complement_table = {
   ['A'] = 'T', ['B'] = 'V', ['C'] = 'G', ['D'] = 'H',
   ['G'] = 'C', ['H'] = 'D', ['K'] = 'M', ['M'] = 'K',
   ['N'] = 'N', ['R'] = 'Y', ['S'] = 'S', ['T'] = 'A',
   ['V'] = 'B', ['W'] = 'W', ['Y'] = 'R',
   ['a'] = 't', ['b'] = 'v', ['c'] = 'g', ['d'] = 'h',
   ['g'] = 'c', ['h'] = 'd', ['k'] = 'm', ['m'] = 'k',
   ['n'] = 'n', ['r'] = 'y', ['s'] = 's', ['t'] = 'a',
   ['v'] = 'b', ['w'] = 'w', ['y'] = 'r',
}





local complement_table_rna = {
   ['A'] = 'U', ['B'] = 'V', ['C'] = 'G', ['D'] = 'H',
   ['G'] = 'C', ['H'] = 'D', ['K'] = 'M', ['M'] = 'K',
   ['N'] = 'N', ['R'] = 'Y', ['S'] = 'S', ['U'] = 'A',
   ['V'] = 'B', ['W'] = 'W', ['Y'] = 'R', ['X'] = 'X',
   ['a'] = 'u', ['b'] = 'v', ['c'] = 'g', ['d'] = 'h',
   ['g'] = 'c', ['h'] = 'd', ['k'] = 'm', ['m'] = 'k',
   ['n'] = 'n', ['r'] = 'y', ['s'] = 's', ['u'] = 'a',
   ['v'] = 'b', ['w'] = 'w', ['y'] = 'r', ['x'] = 'x',
}










function transform.reverse_complement(sequence)
   local sequence_length = #sequence
   local new_sequence = {}

   for i = 1, sequence_length do
      local char = sequence:sub(sequence_length - i + 1, sequence_length - i + 1)
      new_sequence[i] = complement_table[char]
   end

   return table.concat(new_sequence)
end

















function transform.complement(sequence)
   local sequence_length = #sequence
   local new_sequence = {}

   for i = 1, sequence_length do
      local char = sequence:sub(i, i)
      new_sequence[i] = complement_table[char]
   end

   return table.concat(new_sequence)
end








function transform.reverse(sequence)
   local sequence_length = #sequence
   local new_sequence = {}

   for i = 1, sequence_length do
      new_sequence[i] = sequence:sub(sequence_length - i + 1, sequence_length - i + 1)
   end

   return table.concat(new_sequence)
end








function transform.complement_base(base_pair)
   local complement = complement_table[base_pair]
   if complement == nil then
      return " "
   end
   return complement
end










function transform.reverse_complement_rna(sequence)
   local sequence_length = #sequence
   local new_sequence = {}

   for i = 1, sequence_length do
      local char = sequence:sub(sequence_length - i + 1, sequence_length - i + 1)
      new_sequence[i] = complement_table_rna[char]
   end

   return table.concat(new_sequence)
end

















function transform.complement_rna(sequence)
   local sequence_length = #sequence
   local new_sequence = {}

   for i = 1, sequence_length do
      local char = sequence:sub(i, i)
      new_sequence[i] = complement_table_rna[char]
   end

   return table.concat(new_sequence)
end









function transform.complement_base_rna(base_pair)
   local complement = complement_table_rna[base_pair]
   if complement == nil then
      return " "
   end
   return complement
end






function transform.get_window(target_sequence, left_flank, right_flank, circular)
   local search_sequence = target_sequence
   if circular then
      search_sequence = target_sequence .. target_sequence
   end

   local target_upper = search_sequence:upper()
   local left_flank_upper = left_flank:upper()
   local right_flank_upper = right_flank:upper()
   local left_start, left_end = target_upper:find(left_flank_upper, 1, true)

   if left_start then
      local right_start, _ = target_upper:find(right_flank_upper, left_end + 1, true)
      if right_start then
         return search_sequence:sub(left_end + 1, right_start - 1)
      end
   end

   local rev_comp = transform.reverse_complement(search_sequence)
   local rev_comp_upper = rev_comp:upper()

   left_start, left_end = rev_comp_upper:find(left_flank_upper, 1, true)

   if left_start then
      local right_start, _ = rev_comp_upper:find(right_flank_upper, left_end + 1, true)

      if right_start then
         return rev_comp:sub(left_end + 1, right_start - 1)
      end
   end


   return nil
end






































































local align = {}























local default_matrix = {
   data = {
      A = { A = 1, C = -1, G = -1, T = -1, U = -1, ["-"] = -1 },
      C = { A = -1, C = 1, G = -1, T = -1, U = -1, ["-"] = -1 },
      G = { A = -1, C = -1, G = 1, T = -1, U = -1, ["-"] = -1 },
      T = { A = -1, C = -1, G = -1, T = 1, U = -1, ["-"] = -1 },
      U = { A = -1, C = -1, G = -1, T = -1, U = 1, ["-"] = -1 },
      ["-"] = { A = -1, C = -1, G = -1, T = -1, U = -1, ["-"] = -1 },
   },
}
align.default_matrix = default_matrix


function align.new_scoring(substitution_matrix, gap_penalty)
   if not substitution_matrix then
      substitution_matrix = default_matrix
   end
   return {
      substitution_matrix = substitution_matrix,
      gap_penalty = gap_penalty,
   }
end


local function score(scoring, a, b)
   if not scoring.substitution_matrix.data[a] or not scoring.substitution_matrix.data[a][b] then
      error("Invalid characters for scoring: " .. a .. ", " .. b)
   end
   return scoring.substitution_matrix.data[a][b]
end


local function max(a, b)
   if a > b then
      return a
   end
   return b
end














function align.needleman_wunsch(string_a, string_b, scoring)


   local column_length_m, row_length_n = #string_a, #string_b


   local matrix = {}
   for i = 0, column_length_m do
      matrix[i] = {}
      for j = 0, row_length_n do
         matrix[i][j] = 0
      end
   end


   for i = 1, column_length_m do
      matrix[i][0] = matrix[i - 1][0] + scoring.gap_penalty
   end


   for j = 1, row_length_n do
      matrix[0][j] = matrix[0][j - 1] + scoring.gap_penalty
   end


   for i = 1, column_length_m do
      for j = 1, row_length_n do
         local match_score = score(scoring, string_a:sub(i, i), string_b:sub(j, j))
         matrix[i][j] = max(
         matrix[i - 1][j - 1] + match_score,
         max(
         matrix[i - 1][j] + scoring.gap_penalty,
         matrix[i][j - 1] + scoring.gap_penalty))


      end
   end


   local align_a = ""
   local align_b = ""
   local i, j = column_length_m, row_length_n

   while i > 0 and j > 0 do
      local match_score = score(scoring, string_a:sub(i, i), string_b:sub(j, j))
      if matrix[i][j] == matrix[i - 1][j - 1] + match_score then
         align_a = string_a:sub(i, i) .. align_a
         align_b = string_b:sub(j, j) .. align_b
         i = i - 1
         j = j - 1
      elseif matrix[i][j] == matrix[i - 1][j] + scoring.gap_penalty then
         align_a = string_a:sub(i, i) .. align_a
         align_b = "-" .. align_b
         i = i - 1
      else
         align_a = "-" .. align_a
         align_b = string_b:sub(j, j) .. align_b
         j = j - 1
      end
   end

   return matrix[column_length_m][row_length_n], align_a, align_b
end




function align.smith_waterman(string_a, string_b, scoring)
   local column_length_m, row_length_n = #string_a, #string_b


   local matrix = {}
   for i = 0, column_length_m do
      matrix[i] = {}
      for j = 0, row_length_n do
         matrix[i][j] = 0
      end
   end


   local max_score = 0
   local max_score_row = 0
   local max_score_col = 0


   for i = 1, column_length_m do
      for j = 1, row_length_n do
         local match_score = score(scoring, string_a:sub(i, i), string_b:sub(j, j))
         local diag_score = matrix[i - 1][j - 1] + match_score
         local up_score = matrix[i - 1][j] + scoring.gap_penalty
         local left_score = matrix[i][j - 1] + scoring.gap_penalty
         matrix[i][j] = max(0, max(diag_score, max(up_score, left_score)))

         if matrix[i][j] > max_score then
            max_score = matrix[i][j]
            max_score_row = i
            max_score_col = j
         end
      end
   end


   local align_a = ""
   local align_b = ""
   local i, j = max_score_row, max_score_col

   while matrix[i][j] > 0 do
      local match_score = score(scoring, string_a:sub(i, i), string_b:sub(j, j))
      if matrix[i][j] == matrix[i - 1][j - 1] + match_score then
         align_a = string_a:sub(i, i) .. align_a
         align_b = string_b:sub(j, j) .. align_b
         i = i - 1
         j = j - 1
      elseif matrix[i][j] == matrix[i - 1][j] + scoring.gap_penalty then
         align_a = string_a:sub(i, i) .. align_a
         align_b = "-" .. align_b
         i = i - 1
      else
         align_a = "-" .. align_a
         align_b = string_b:sub(j, j) .. align_b
         j = j - 1
      end
   end

   return max_score, align_a, align_b
end













function align.align_many(func, target, candidates, scoring, ntop)
   local results = {}
   for i = 1, #candidates do
      local alignment_score, align_a, align_b = func(target, candidates[i], scoring)
      results[#results + 1] = { alignment_score, align_a, align_b, i }
   end


   table.sort(results, function(a, b)
      return a[1] > b[1]
   end)


   if ntop < 0 then
      ntop = 0
   end
   if ntop > #results then
      ntop = #results
   end


   local out = {}
   for i = 1, ntop do
      out[i] = results[i]
   end

   return out
end



















































local mash = {}




















local function new(kmer_size, sketch_size, hash_algorithm)
   local sketches = {}
   for i = 1, sketch_size do
      sketches[i] = 0xFFFFFFFF
   end

   return {
      kmer_size = kmer_size,
      sketch_size = sketch_size,
      sketches = sketches,
      hash_algorithm = hash_algorithm,
   }
end



local function sketch_sort(sketch_mash, sequence, sort_after)

   for i = 1, sketch_mash.sketch_size do
      sketch_mash.sketches[i] = 0xFFFFFFFF
   end


   local seen_hashes = {}


   for kmer_start = 1, #sequence - sketch_mash.kmer_size + 1 do
      local kmer = sequence:sub(kmer_start, kmer_start + sketch_mash.kmer_size - 1)
      sketch_mash.hash_algorithm:reset()
      sketch_mash.hash_algorithm:write(kmer)
      local hash_value = sketch_mash.hash_algorithm:sum32()


      if not seen_hashes[hash_value] then
         seen_hashes[hash_value] = true

         if not sort_after then


            if hash_value < sketch_mash.sketches[sketch_mash.sketch_size] then
               sketch_mash.sketches[sketch_mash.sketch_size] = hash_value
               table.sort(sketch_mash.sketches)
            end
         else
            table.insert(sketch_mash.sketches, hash_value)
         end
      end
   end
   if sort_after then
      table.sort(sketch_mash.sketches)
   end
end


local function sketch(sketch_mash, sequence)
   sketch_sort(sketch_mash, sequence, false)
end




local function new_containment_sketch(kmer_size, sequence, hash_algorithm)

   local seq_len = #sequence
   local max_kmers = seq_len - kmer_size + 1
   if max_kmers < 1 then
      return {
         kmer_size = kmer_size,
         sketch_size = 0,
         sketches = {},
         hash_algorithm = hash_algorithm,
      }
   end


   local mash_obj = new(kmer_size, max_kmers, hash_algorithm)
   sketch_sort(mash_obj, sequence, true)


   local sketches = mash_obj.sketches
   local effective = {}

   for i = 1, mash_obj.sketch_size do
      local h = sketches[i]
      if h ~= 0xFFFFFFFF then
         effective[#effective + 1] = h
      else


         break
      end
   end


   mash_obj.sketches = effective
   mash_obj.sketch_size = #effective
   return mash_obj
end

local function similarity(a, b)

   if not a.sketches[1] or not b.sketches[1] or
      a.sketches[1] == 0xFFFFFFFF or b.sketches[1] == 0xFFFFFFFF then
      return 0
   end

   local same_hashes = 0


   local larger_size
   local smaller_size
   local larger_sketches
   local smaller_sketches

   if a.sketch_size < b.sketch_size then
      smaller_size = a.sketch_size
      larger_size = b.sketch_size
      smaller_sketches = a.sketches
      larger_sketches = b.sketches
   else
      smaller_size = b.sketch_size
      larger_size = a.sketch_size
      smaller_sketches = b.sketches
      larger_sketches = a.sketches
   end

   local small_idx, large_idx = 1, 1
   while small_idx <= smaller_size and large_idx <= larger_size do
      if smaller_sketches[small_idx] == larger_sketches[large_idx] then
         same_hashes = same_hashes + 1
         small_idx = small_idx + 1
         large_idx = large_idx + 1
      elseif smaller_sketches[small_idx] < larger_sketches[large_idx] then
         small_idx = small_idx + 1
      else
         large_idx = large_idx + 1
      end
   end

   return same_hashes / smaller_size
end


local function distance(a, b)
   return 1 - similarity(a, b)
end




local function containment(a, b)

   if not a.sketches[1] or not b.sketches[1] or
      a.sketches[1] == 0xFFFFFFFF or b.sketches[1] == 0xFFFFFFFF then
      return 0
   end


   local i, j = 1, 1
   local same_hashes = 0




   while i <= a.sketch_size and j <= b.sketch_size do
      if a.sketches[i] == b.sketches[j] then
         same_hashes = same_hashes + 1
         i = i + 1
         j = j + 1
      elseif a.sketches[i] < b.sketches[j] then
         i = i + 1
      else
         j = j + 1
      end
   end
   return same_hashes / a.sketch_size
end

mash.new = new
mash.sketch = sketch
mash.similarity = similarity
mash.distance = distance
mash.new_containment_sketch = new_containment_sketch
mash.containment = containment



































local seqhash = {}




















local SEQ_TYPE_DNA = "DNA"
local SEQ_TYPE_RNA = "RNA"
local SEQ_TYPE_PROTEIN = "PROTEIN"
local SEQ_TYPE_FRAGMENT = "FRAGMENT"


local HASH2_VERSION_MASK = 0xF0
local HASH2_CIRCULARITY_MASK = 0x08
local HASH2_DOUBLE_STRANDED_MASK = 0x04
local HASH2_TYPE_MASK = 0x03

local HASH2_VERSION_SHIFT = 4
local HASH2_CIRCULARITY_SHIFT = 3
local HASH2_DOUBLE_STRANDED_SHIFT = 2


local ALPHABET = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"


local sequence_type_to_flag = {
   [SEQ_TYPE_DNA] = 0x00,
   [SEQ_TYPE_RNA] = 0x01,
   [SEQ_TYPE_PROTEIN] = 0x02,
   [SEQ_TYPE_FRAGMENT] = 0x03,
}

local flag_to_sequence_type = {
   [0x00] = SEQ_TYPE_DNA,
   [0x01] = SEQ_TYPE_RNA,
   [0x02] = SEQ_TYPE_PROTEIN,
   [0x03] = SEQ_TYPE_FRAGMENT,
}


local function create_metadata_map()
   local m = {

      ["DNA_circular_double"] = "A",
      ["DNA_circular_single"] = "B",
      ["DNA_linear_double"] = "C",
      ["DNA_linear_single"] = "D",

      ["RNA_circular_double"] = "E",
      ["RNA_circular_single"] = "F",
      ["RNA_linear_double"] = "G",
      ["RNA_linear_single"] = "H",

      ["PROTEIN_linear_single"] = "I",
      ["PROTEIN_circular_single"] = "J",

      ["FRAGMENT_linear_single"] = "K",
      ["FRAGMENT_circular_single"] = "L",
      ["FRAGMENT_linear_double"] = "M",
      ["FRAGMENT_circular_double"] = "N",
   }
   return m
end

seqhash.hash2_metadata = create_metadata_map()


local function encode_to_base58(input)

   if #input == 0 then
      return "1"
   end


   local zeros = 0
   for i = 1, #input do
      if input[i] == 0 then
         zeros = zeros + 1
      else
         break
      end
   end


   local bytes = {}
   for i = zeros + 1, #input do
      bytes[#bytes + 1] = input[i]
   end


   local result = {}
   while #bytes > 0 do
      local remainder = 0
      local new_bytes = {}

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


   for _ = 1, zeros do
      result[#result + 1] = '1'
   end


   local final = {}
   for i = #result, 1, -1 do
      final[#final + 1] = result[i]
   end

   return table.concat(final)
end

local function decode_from_base58(input)
   if #input == 0 then
      return {}, "empty input string"
   end


   local zeros = 0
   for i = 1, #input do
      if input:sub(i, i) == '1' then
         zeros = zeros + 1
      else
         break
      end
   end


   local bytes = { 0 }

   for i = zeros + 1, #input do
      local char_index = ALPHABET:find(input:sub(i, i))
      if not char_index then
         return {}, "invalid base58 character"
      end
      local digit = char_index - 1


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


   for _ = 1, zeros do
      table.insert(bytes, 1, 0)
   end

   return bytes, ""
end





local function version2_flag(version, sequence_type, circularity, double_stranded)
   local flag = 0
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






local function decode_flag(flag)
   local version = hash.bit32.rshift(hash.bit32.band(flag, HASH2_VERSION_MASK), HASH2_VERSION_SHIFT)
   local circularity = hash.bit32.band(flag, HASH2_CIRCULARITY_MASK) ~= 0
   local double_stranded = hash.bit32.band(flag, HASH2_DOUBLE_STRANDED_MASK) ~= 0
   local seq_type = flag_to_sequence_type[hash.bit32.band(flag, HASH2_TYPE_MASK)]
   return version, seq_type, circularity, double_stranded
end







































local function booth_least_rotation(seq)
   local s = string.upper(seq)
   local doubled = s .. s
   local n = #doubled


   local f = {}
   for i = 0, n - 1 do
      f[i] = -1
   end

   local k = 0



   for j = 1, n - 1 do

      local sj = doubled:byte(j + 1)


      local i = f[j - k - 1]


      while i ~= -1 and sj ~= doubled:byte(k + i + 2) do
         if sj < doubled:byte(k + i + 2) then
            k = j - i - 1
         end
         i = f[i]
      end

      if sj ~= doubled:byte(k + i + 2) then

         if sj < doubled:byte(k + 1) then
            k = j
         end
         f[j - k] = -1
      else
         f[j - k] = i + 1
      end
   end


   local m = #s
   if k >= m then
      k = k - m
   end
   return k
end




local function rotate_sequence(sequence)
   local k = booth_least_rotation(sequence)

   return sequence:sub(k + 2) .. sequence:sub(1, k + 1)
end



local function circular_equality(sequence1, sequence2)
   local sequence1U = string.upper(sequence1)
   local sequence2U = string.upper(sequence2)
   local deterministic_sequence1
   local seq1 = rotate_sequence(sequence1U)
   local rev_comp = transform.reverse_complement(sequence1U)
   local seq2 = rotate_sequence(rev_comp)
   deterministic_sequence1 = seq1 < seq2 and seq1 or seq2

   local deterministic_sequence2
   local seq2_1 = rotate_sequence(sequence2U)
   local rev_comp_2 = transform.reverse_complement(sequence2U)
   local seq2_2 = rotate_sequence(rev_comp_2)
   deterministic_sequence2 = seq2_1 < seq2_2 and seq2_1 or seq2_2

   return deterministic_sequence1 == deterministic_sequence2
end




local function hash2(sequence, sequence_type, circular, double_stranded)

   if sequence_type ~= SEQ_TYPE_DNA and sequence_type ~= SEQ_TYPE_RNA and sequence_type ~= SEQ_TYPE_PROTEIN then
      return {}, "Invalid sequence type"
   end


   sequence = string.upper(sequence)
   if sequence_type == SEQ_TYPE_RNA then
      sequence = string.gsub(sequence, "U", "T")
   end


   if sequence_type == SEQ_TYPE_DNA or sequence_type == SEQ_TYPE_RNA then
      for i = 1, #sequence do
         local char = sequence:sub(i, i)
         if not string.match("ATUGCYRSWKMBDHVNZ", char) then
            return {}, "Invalid DNA/RNA character: " .. char
         end
      end
   elseif sequence_type == SEQ_TYPE_PROTEIN then
      for i = 1, #sequence do
         local char = sequence:sub(i, i)
         if not string.match("ACDEFGHIKLMNPQRSTVWYUO*BXZ", char) then
            return {}, "Invalid protein character: " .. char
         end
      end
   end


   local deterministic_sequence
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


   local result = {}
   for i = 1, 16 do
      result[i] = 0
   end


   result[1] = version2_flag(2, sequence_type, circular, double_stranded)


   local h = hash.new_sha256()
   h:write(deterministic_sequence)
   local sha_bytes = h:sum()


   for i = 2, 16 do
      result[i] = sha_bytes[i - 1]
   end

   return result, ""
end




























local function hash2_fragment(sequence, fwd_overhang_length, rev_overhang_length)

   sequence = string.upper(sequence)
   for i = 1, #sequence do
      local char = sequence:sub(i, i)
      if not string.match("ATUGCYRSWKMBDHVNZ", char) then
         return {}, "Invalid DNA/RNA character: " .. char
      end
   end


   local forward, reverse = fwd_overhang_length, rev_overhang_length
   local deterministic_sequence = sequence
   local reverse_complement = transform.reverse_complement(sequence)

   if sequence > reverse_complement then
      forward = rev_overhang_length
      reverse = fwd_overhang_length
      deterministic_sequence = reverse_complement
   end


   local result = {}
   for i = 1, 16 do
      result[i] = 0
   end


   result[1] = version2_flag(2, SEQ_TYPE_FRAGMENT, false, false)
   result[2] = forward
   result[3] = reverse


   local h = hash.new_sha256()
   h:write(deterministic_sequence)
   local sha_bytes = h:sum()


   for i = 4, 16 do
      result[i] = sha_bytes[i - 3]
   end

   return result, ""
end






local function encode_hash2(hash_to_encode, err)
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


local function decode_hash2(encoded_string)
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
















local primers = {}





























local nearest_neighbors_thermodynamics = {
   ["AA"] = { h = -7.6, s = -21.3 },
   ["TT"] = { h = -7.6, s = -21.3 },
   ["AT"] = { h = -7.2, s = -20.4 },
   ["TA"] = { h = -7.2, s = -21.3 },
   ["CA"] = { h = -8.5, s = -22.7 },
   ["TG"] = { h = -8.5, s = -22.7 },
   ["GT"] = { h = -8.4, s = -22.4 },
   ["AC"] = { h = -8.4, s = -22.4 },
   ["CT"] = { h = -7.8, s = -21.0 },
   ["AG"] = { h = -7.8, s = -21.0 },
   ["GA"] = { h = -8.2, s = -22.2 },
   ["TC"] = { h = -8.2, s = -22.2 },
   ["CG"] = { h = -10.6, s = -27.2 },
   ["GC"] = { h = -9.8, s = -24.4 },
   ["GG"] = { h = -8.0, s = -19.9 },
   ["CC"] = { h = -8.0, s = -19.9 },
}

local initial_thermodynamic_penalty = { h = 0.2, s = -5.7 }
local symmetry_thermodynamic_penalty = { h = 0, s = -1.4 }
local terminal_at_thermodynamic_penalty = { h = 2.2, s = 6.9 }





function primers.santa_lucia(sequence, primer_concentration, salt_concentration, magnesium_concentration)
   sequence = string.upper(sequence)

   local gas_constant = 1.9872
   local symmetry_factor = 1
   local dh = 0
   local ds = 0


   dh = dh + initial_thermodynamic_penalty.h
   ds = ds + initial_thermodynamic_penalty.s


   if sequence == transform.reverse_complement(sequence) then
      dh = dh + symmetry_thermodynamic_penalty.h
      ds = ds + symmetry_thermodynamic_penalty.s
      symmetry_factor = 1
   else
      symmetry_factor = 4
   end


   local last_base = sequence:sub(-1)
   if last_base == "A" or last_base == "T" then
      dh = dh + terminal_at_thermodynamic_penalty.h
      ds = ds + terminal_at_thermodynamic_penalty.s
   end


   local salt_effect = salt_concentration + (magnesium_concentration * 140)
   ds = ds + (0.368 * (#sequence - 1) * math.log(salt_effect))


   for i = 1, #sequence - 1 do
      local pair = sequence:sub(i, i + 1)
      local dt = nearest_neighbors_thermodynamics[pair]
      if dt then
         dh = dh + dt.h
         ds = ds + dt.s
      end
   end

   local melting_temp = dh * 1000 / (ds + gas_constant * math.log(primer_concentration / symmetry_factor)) - 273.15
   return melting_temp, dh, ds
end






function primers.marmur_doty(sequence)
   sequence = string.upper(sequence)

   local a_count = 0
   local t_count = 0
   local c_count = 0
   local g_count = 0

   for i = 1, #sequence do
      local base = sequence:sub(i, i)
      if base == "A" then a_count = a_count + 1
      elseif base == "T" then t_count = t_count + 1
      elseif base == "C" then c_count = c_count + 1
      elseif base == "G" then g_count = g_count + 1
      end
   end

   local melting_temp = 2 * (a_count + t_count) + 4 * (c_count + g_count) - 7.0
   return melting_temp
end




function primers.melting_temp(sequence)
   local primer_concentration = 500e-9
   local salt_concentration = 50e-3
   local magnesium_concentration = 0.0

   local melting_temp = primers.santa_lucia(sequence, primer_concentration, salt_concentration, magnesium_concentration)
   return melting_temp
end





































































































function primers.nucleobase_debruijn_sequence(substring_length)
   local alphabet = "ATGC"
   local alphabet_length = #alphabet
   local a = {}


   for i = 1, alphabet_length * substring_length do
      a[i] = 0
   end

   local seq = {}

   local function construct_debruijn(t, p)
      if t > substring_length then
         if substring_length % p == 0 then
            for i = 1, p do
               table.insert(seq, a[i] or 0)
            end
         end
      else
         local tp_idx = (t - p)
         a[t] = a[tp_idx] or 0
         construct_debruijn(t + 1, p)
         local start_val = (a[tp_idx] or 0)
         for j = start_val + 1, alphabet_length - 1 do
            a[t] = j
            construct_debruijn(t + 1, t)
         end
      end
   end

   construct_debruijn(1, 1)

   local result = ""
   for _, i in ipairs(seq) do
      local idx = (i + 1)
      result = result .. alphabet:sub(idx, idx)
   end

   return result .. result:sub(1, (substring_length - 1))
end







function primers.create_barcodes_with_banned_sequences(length, max_subsequence, banned_sequences, banned_functions)
   local barcodes = {}
   local start = 1
   local finish = 1
   local debruijn = primers.nucleobase_debruijn_sequence(max_subsequence)

   local barcode_num = 0
   while (barcode_num * (length - (max_subsequence - 1))) + length <= #debruijn do
      start = barcode_num * (length - (max_subsequence - 1)) + 1
      finish = start + length - 1
      barcode_num = barcode_num + 1


      for _, banned_sequence in ipairs(banned_sequences) do

         while debruijn:sub(start, finish):find(banned_sequence, 1, true) do
            if finish + 1 > #debruijn then
               return barcodes
            end
            start = start + 1
            finish = finish + 1
            barcode_num = barcode_num + 1
         end


         while debruijn:sub(start, finish):find(transform.reverse_complement(banned_sequence), 1, true) do
            if finish + 1 > #debruijn then
               return barcodes
            end
            start = start + 1
            finish = finish + 1
            barcode_num = barcode_num + 1
         end
      end


      for _, banned_function in ipairs(banned_functions) do
         while not banned_function(debruijn:sub(start, finish)) do
            if finish + 1 > #debruijn then
               return barcodes
            end
            start = start + 1
            finish = finish + 1
            barcode_num = barcode_num + 1
         end
      end

      table.insert(barcodes, debruijn:sub(start, finish))
   end

   return barcodes
end




function primers.create_barcodes(length, max_subsequence)
   return primers.create_barcodes_with_banned_sequences(length, max_subsequence, {}, {})
end




function primers.create_barcodes_gc_range(length, max_subsequence, min_gc_content, max_gc_content)
   local function gc_content(sequence)
      local gc_count = 0
      sequence = string.upper(sequence)
      for i = 1, #sequence do
         local base = sequence:sub(i, i)
         if base == "G" or base == "C" then
            gc_count = gc_count + 1
         end
      end
      return gc_count / #sequence
   end

   local function gc_barcode_func(barcode_to_check)
      local gc = gc_content(barcode_to_check)
      if gc < min_gc_content or gc > max_gc_content then
         return false
      end
      return true
   end

   return primers.create_barcodes_with_banned_sequences(length, max_subsequence, {}, { gc_barcode_func })
end

















local orthoprimers = {}



































local ortho_primers = {
   "AAACACGTGGCAAACATTCC",
   "AAACCGGAGCCATACAGTAC",
   "AAAGCACTCTTAGGCCTCTG",
   "AAAGGGGCCGTCAATATCAG",
   "AAATAAGACGACGACCCTCG",
   "AACGATGATGCTCACTCTCG",
   "AAGAATTACTGACCCCTCGG",
   "AAGACGATCCGAGCCATTAC",
   "AAGGAACTATGGCATCGAGC",
   "AAGGACTGCATACCAGGTTG",
   "AAGGATATGTAGACACCGCC",
   "AAGGCCCAGAAGGATACAAC",
   "AAGGCGCTCGGATAATACTC",
   "AAGGTATGTATAGCGACCGC",
   "AATAGGAACCTCTTACGCGG",
   "AATATCACGCAAAAGCACCG",
   "AATCAGTTTCTTTGGCAGCC",
   "AATGCAAAGCTATTAGCGCG",
   "AATGCGTCATTTTACACGGC",
   "AATGTCCTTAGGCAGTCGTC",
   "ACAACGAGCAGACCGAATAG",
   "ACAAGGAGTCGGCATATCAC",
   "ACAGAACGAACAGGCACTAC",
   "ACAGGAAGCAAGGTATACGC",
   "ACAGGGTATATTGAGTGCCC",
   "ACATAAGCGATCCCAAGGTC",
   "ACATCGCATACCAGAACAGG",
   "ACATTAAATTTCGCCGTGGC",
   "ACCACAGGTCAAGATTCACG",
   "ACCCGTATCGCATAAGGATG",
   "ACGAGATGATGCACCGATAG",
   "ACGATGGGGACATAGAACAC",
   "ACGGAGCCCTTATTGTAACC",
   "ACGTATGGGGAACACTACAC",
   "ACGTGAAACTGTATCGAGCC",
   "ACGTTCAGTTTTCCAATGGC",
   "ACTAGATTAGCAAGGCACCC",
   "ACTGGACCCAATAAAAGGCC",
   "ACTTCGATTGGCAAGGACTG",
   "AGAACATAGCATTCACGGGG",
   "AGACAACAATCTGAGGCTGG",
   "AGACAAGCCTTAACCGTAGG",
   "AGACACAAGGCTGATTCCAG",
   "AGACATGGGATTGACCACAC",
   "AGAGAGGCATGATTGACCTC",
   "AGAGTTGCACCTAGAATCCG",
   "AGATAGATGCTCCGTCAAGC",
   "AGATAGTCACGCACAAGACC",
   "AGATTAGCCGACTTTCCTGG",
   "AGATTAGCTGCCGATACTGG",
   "AGATTGTTACTCCGACGGAC",
   "AGATTTCCGACGAGATTCCC",
   "AGCATCCGTCTAAATCTCGG",
   "AGCTATAAGAATTGCCGGGC",
   "AGCTATGATCCCGGTGTAAC",
   "AGCTCAATCTAACAGTGGGG",
   "AGGACACCAGACCAATGAAG",
   "AGGGCTAATTACCATCAGCG",
   "AGGTGATCTGACGAATGTCC",
   "AGTAAAGCATAGTGCCCAGC",
   "AGTAGTATCCGAATCGCTGC",
   "AGTATCTCAGCAAGGGCAAC",
   "AGTATTAGGCGTCAAGGTCC",
   "AGTATTCTTACAGCCAGCCG",
   "AGTATTGCCGGACTAAACCC",
   "AGTCCCAAGTTCAGACGTAC",
   "AGTCCGACACAATGTGACAC",
   "AGTGAACTGACCGAATCCTC",
   "AGTGGTCTGTAAACCGTACC",
   "AGTGTTTTCCATTTTCCGCG",
   "AGTTATAAGGGTCCGATGCC",
   "AGTTGCAGTATCTAACCCGC",
   "AGTTGTAATATCACCCGCGC",
   "ATACGTGGCTAGCATGAGAC",
   "ATACTGTAAGAACCACGCGG",
   "ATAGATCATGTCGGCAGTCG",
   "ATAGATGGTGCCTACATGCG",
   "ATCACAACAAAGGACGGGTC",
   "ATCAGACAACACAGAGGCTG",
   "ATCCAGGAGGTCTAGGAACC",
   "ATCCTAGAAAAGGCGAAGGC",
   "ATGCCATGACGACAACTAGC",
   "ATGCTAGCTGGAACTATCGG",
   "ATTAGGATTGCGAGCGACAC",
   "ATTAGTACACTCCGTGAGCG",
   "ATTCAAGGGTTGGACGACTC",
   "ATTCTCACGACGCAAGATGG",
   "ATTGACGGGAACTACACTCG",
   "CACTCGATAGGTACAACCGG",
   "CAGACCTACGGATCTTAGCG",
   "CCACGAGATAAGAGGATGGC",
   "CCAGAGCTTAGGGGACATAC",
   "CCCGAGGGGAGAAATATACC",
   "CCGAGGGAACCATGATACAG",
   "CCGGGAGGAAGATATAGCAC",
   "CCGGTTGTACCTATCGAGTG",
   "CCGTGCGACAAGATTTCAAG",
   "CCTTTAACAGGACATGCAGC",
   "CGAACGCAAAAGTCCTCAAG",
   "CGATAGAACGACCAGGTAGC",
   "CGGATCGAACTTAGGTAGCC",
   "CGGGAGGAAGTCTTTAGACC",
   "CTAATATCCCTGAGCGACGG",
   "CTAGGGAACCAGGCTTAACG",
   "CTAGGGGATGGTCCAATACG",
   "CTATAGAATCCGGGCTGGTC",
   "CTGCTAGGGGCTACTTATCG",
   "GAAAAGTCCCAATGAGTGCC",
   "GAAGTGGTTTGCCTAAACGC",
   "GACCATGCAAGGAGAGGTAC",
   "GATACATAGACTTGGCCCCG",
   "GCACGCAAAAGGACATAACC",
   "GCAGCGTTTTAGCCTACAAG",
   "GCATAAAGTTGACAGGCCAG",
   "GCTAAATAGAGGGAAGCCCC",
   "GGAAAACTAAGACAAGGCGC",
   "GGAAACAATAACCATCGGCG",
   "GGGCACCGATTAAGAAATGC",
   "GGGTTGTCTCCTCTGATAGC",
   "GTACTCAGAGATTGCCGGAG",
   "GTATAAGATCAGCCGGACCC",
   "GTATGTCGGCTCTCGTATCG",
   "GTTCAGAGGTACGAACCCTC",
   "GTTGCATCTAAGCCAAGTGC",
   "TAAAGAGAGGGCGTCCAATC",
   "TAACGACGTGCCGAACTTAG",
   "TAAGATAGCACCACGGATGG",
   "TAAGGATTCATCAGGTGCGC",
   "TAAGGGACGATGCTTAACCC",
   "TACCACGAAATGCACAGGAG",
   "TACTGATAATTCGGACGCCC",
   "TACTTGAATACCACGTGGCC",
   "TAGCCAGGCAAAAGAGATCC",
   "TAGCTCGATAATCAAGGGGC",
   "TAGTGACCTAATGCCATGGG",
   "TAGTTGAGAACACGAACCCG",
   "TATAACAGGCTGCTGAGACC",
   "TATACTGAAGAACGGCCCAG",
   "TATCAATCCGGAACCAGTGC",
   "TATCACGGAAGGACTCAACG",
   "TCAAAGGAGCACGAACCTAC",
   "TCAAGGTCCGTTATGGAACC",
   "TCACATAGAAGGACATGGCG",
   "TCACTTGGTATCGAGAACGG",
   "TCAGCCTTTCATTGATTGCG",
   "TCATCGACAAGATACAGGCG",
   "TCCAATTATACGGAGCAGGC",
   "TCGAATATGCTGTAACCCCG",
   "TCGACCAGGTTATCATGAGC",
   "TCGAGACAAGAACGATTCCC",
   "TCTAGGACTATCACCGGAGG",
   "TCTTCATAAGCCAGAGTGCC",
   "TCTTGCGATAGACACAAGCC",
   "TGAGCCATAAAAGCAAAGCG",
   "TGAGCGCAGAACTATCAGAC",
   "TGCATAGTATCCCAACAGGG",
   "TGCCAAAGGGTAGAGACATC",
   "TGCTGAATGAGAAACCTCGG",
   "TGGGGACGACTTATAATGCC",
   "TGTGGACCCTATCAAACGAG",
   "TTAGCTCAGGTCCAAAGTCC",
   "TTAGTAGGCAAGCATACCCG",
   "TTCGGGAGCGGATTATACAC",
   "TTCTGGGACTGGATAACACG",
   "TTGACAGACAATCCGTAGGC",
}

local function new_primer_set(self)

   local forward_primer = ""
   local reverse_primer = ""


   local least_used = 1000000
   for i = #self.primers, 1, -1 do
      if self.primer_use_quantity[self.primers[i]] <= least_used then
         forward_primer = self.primers[i]
         least_used = self.primer_use_quantity[self.primers[i]]
      end
   end



   local possible_reverse_primers = {}
   for _, primer in ipairs(self.primers) do
      local pair_key = orthoprimers.make_primer_pair_key(forward_primer, primer)
      if not self.primer_pairs[pair_key] and forward_primer ~= primer then
         table.insert(possible_reverse_primers, primer)
      end
   end


   least_used = 1000000
   for i = #possible_reverse_primers, 1, -1 do
      if self.primer_use_quantity[possible_reverse_primers[i]] <= least_used then
         reverse_primer = possible_reverse_primers[i]
         least_used = self.primer_use_quantity[possible_reverse_primers[i]]
      end
   end


   if forward_primer == "" or reverse_primer == "" then
      return "", "", "Not enough primers for genes in pool"
   end


   self.primer_use_quantity[forward_primer] = self.primer_use_quantity[forward_primer] + 1
   self.primer_use_quantity[reverse_primer] = self.primer_use_quantity[reverse_primer] + 1
   self.primer_pairs[orthoprimers.make_primer_pair_key(forward_primer, reverse_primer)] = true

   return forward_primer, reverse_primer, nil
end





function orthoprimers.make_primer_pair_key(forward, reverse)
   if forward > reverse then
      return forward .. "|" .. reverse
   end
   return reverse .. "|" .. forward
end



function orthoprimers.new_orthogonal_primer_set(primer_set)
   local quantity_map = {}
   for _, primer in ipairs(primer_set) do
      quantity_map[primer] = 0
   end
   local primer_pairs = {}
   return { primers = primer_set, primer_use_quantity = quantity_map, primer_pairs = primer_pairs, new_primer_set = new_primer_set }
end








function orthoprimers.new_default_orthogonal_primer_set()
   local first_96 = {}
   for i = 1, 96 do
      first_96[i] = ortho_primers[i]
   end
   return orthoprimers.new_orthogonal_primer_set(first_96)
end

orthoprimers.ortho_primers = ortho_primers






























local pcr = {}








local minimal_primer_length = 15


local function find_all(str, pattern)
   local positions = {}
   local pos = 1
   while true do
      local start, finish = str:find(pattern, pos, true)
      if not start then break end
      table.insert(positions, start)
      pos = finish + 1
   end
   return positions
end


local function generate_pcr_fragments(
   sequence,
   forward_location,
   reverse_location,
   forward_primer_indices,
   reverse_primer_indices,
   minimal_primers,
   primer_list)

   local pcr_fragments = {}

   for _, forward_primer_index in ipairs(forward_primer_indices) do
      local minimal_primer = minimal_primers[forward_primer_index]
      local full_primer_forward = primer_list[forward_primer_index]

      for _, reverse_primer_index in ipairs(reverse_primer_indices) do
         local full_primer_reverse = transform.reverse_complement(primer_list[reverse_primer_index])
         local pcr_fragment = full_primer_forward:sub(1, #full_primer_forward - #minimal_primer) ..
         sequence:sub(forward_location, reverse_location - 1) ..
         full_primer_reverse
         table.insert(pcr_fragments, pcr_fragment)
      end
   end

   return pcr_fragments
end







function pcr.design_primers_with_overhangs(sequence, forward_overhang, reverse_overhang, target_tm)
   sequence = string.upper(sequence)
   local forward_primer = sequence:sub(1, minimal_primer_length)


   local additional_nucleotides = 0
   while primers.melting_temp(forward_primer) < target_tm do
      additional_nucleotides = additional_nucleotides + 1
      forward_primer = sequence:sub(1, minimal_primer_length + additional_nucleotides)
   end


   local reverse_primer = transform.reverse_complement(sequence:sub(#sequence - minimal_primer_length + 1))
   additional_nucleotides = 0
   while primers.melting_temp(reverse_primer) < target_tm do
      additional_nucleotides = additional_nucleotides + 1
      reverse_primer = transform.reverse_complement(sequence:sub(#sequence - (minimal_primer_length + additional_nucleotides) + 1))
   end


   forward_primer = forward_overhang .. forward_primer
   reverse_primer = transform.reverse_complement(reverse_overhang) .. reverse_primer

   return forward_primer, reverse_primer
end




function pcr.design_primers(sequence, target_tm)
   return pcr.design_primers_with_overhangs(sequence, "", "", target_tm)
end








function pcr.simulate_simple(sequences, target_tm, circular, primer_list)

   local upper_primers = {}
   for _, primer in ipairs(primer_list) do
      table.insert(upper_primers, string.upper(primer))
   end

   local pcr_fragments = {}

   for _, sequence in ipairs(sequences) do
      sequence = string.upper(sequence)

      local forward_locations = {}
      local reverse_locations = {}
      local minimal_primers = {}


      for primer_index, primer in ipairs(upper_primers) do
         local minimal_length = minimal_primer_length
         while primers.melting_temp(primer:sub(-minimal_length)) < target_tm do
            minimal_length = minimal_length + 1
            if minimal_length > #primer then break end
         end

         local minimal_primer = primer:sub(-minimal_length)
         if minimal_primer ~= primer then
            minimal_primers[primer_index] = minimal_primer
         end


         for _, loc in ipairs(find_all(sequence, minimal_primer)) do
            if not forward_locations[loc] then forward_locations[loc] = {} end
            table.insert(forward_locations[loc], primer_index)
         end


         local rev_minimal_primer = transform.reverse_complement(minimal_primer)
         for _, loc in ipairs(find_all(sequence, rev_minimal_primer)) do
            if not reverse_locations[loc] then reverse_locations[loc] = {} end
            table.insert(reverse_locations[loc], primer_index)
         end
      end


      local forward_locs = {}
      local reverse_locs = {}
      for loc, _ in pairs(forward_locations) do table.insert(forward_locs, loc) end
      for loc, _ in pairs(reverse_locations) do table.insert(reverse_locs, loc) end
      table.sort(forward_locs)
      table.sort(reverse_locs)


      for i, forward_loc in ipairs(forward_locs) do
         if i < #forward_locs then
            for _, reverse_loc in ipairs(reverse_locs) do
               if forward_loc < reverse_loc and reverse_loc < forward_locs[i + 1] then
                  local new_fragments = generate_pcr_fragments(
                  sequence,
                  forward_loc,
                  reverse_loc,
                  forward_locations[forward_loc],
                  reverse_locations[reverse_loc],
                  minimal_primers,
                  upper_primers)

                  for _, fragment in ipairs(new_fragments) do
                     table.insert(pcr_fragments, fragment)
                  end
                  break
               end
            end
         else
            local found_fragment = false
            for _, reverse_loc in ipairs(reverse_locs) do
               if forward_loc < reverse_loc then
                  local new_fragments = generate_pcr_fragments(
                  sequence,
                  forward_loc,
                  reverse_loc,
                  forward_locations[forward_loc],
                  reverse_locations[reverse_loc],
                  minimal_primers,
                  upper_primers)

                  for _, fragment in ipairs(new_fragments) do
                     table.insert(pcr_fragments, fragment)
                  end
                  found_fragment = true
               end
            end


            if circular and not found_fragment then
               for _, reverse_loc in ipairs(reverse_locs) do
                  if forward_locs[1] > reverse_loc then
                     local rotated_sequence = sequence:sub(forward_loc) .. sequence:sub(1, forward_loc - 1)
                     local rotated_forward_loc = 1
                     local rotated_reverse_loc = #sequence - forward_loc + reverse_loc + 1
                     local new_fragments = generate_pcr_fragments(
                     rotated_sequence,
                     rotated_forward_loc,
                     rotated_reverse_loc,
                     forward_locations[forward_loc],
                     reverse_locations[reverse_loc],
                     minimal_primers,
                     upper_primers)

                     for _, fragment in ipairs(new_fragments) do
                        table.insert(pcr_fragments, fragment)
                     end
                  end
               end
            end
         end
      end
   end

   return pcr_fragments
end








function pcr.simulate(sequences, target_tm, circular, primer_list)
   local initial_amplification = pcr.simulate_simple(sequences, target_tm, circular, primer_list)


   local combined_primers = {}
   for _, primer in ipairs(primer_list) do
      table.insert(combined_primers, primer)
   end
   for _, fragment in ipairs(initial_amplification) do
      table.insert(combined_primers, fragment)
   end

   local subsequent_amplification = pcr.simulate_simple(sequences, target_tm, circular, combined_primers)

   if #initial_amplification ~= #subsequent_amplification then
      return initial_amplification, "Concatemerization detected in PCR."
   end

   return initial_amplification, ""
end
















local bio = {}
































































local DEFAULT_MAX_LINE_LENGTH = 65536






local DEFAULT_MAX_LENGTHS = {
   ["FASTA"] = DEFAULT_MAX_LINE_LENGTH,
   ["FASTQ"] = 8 * 1024 * 1024,
   ["GENBANK"] = DEFAULT_MAX_LINE_LENGTH,
   ["SLOW5"] = 128 * 1024 * 1024,
   ["SAM"] = DEFAULT_MAX_LINE_LENGTH,
   ["PILEUP"] = DEFAULT_MAX_LINE_LENGTH,
}
bio.DEFAULT_MAX_LENGTHS = DEFAULT_MAX_LENGTHS






























































function bio.new_string_reader(content)
   if type(content) ~= "string" then
      error("StringReader content must be a string")
   end

   local reader = {
      content = content,
      position = 1,
      line_number = 0,
      _eof_emitted = false,

      read = nil,
      close = nil,
      get_line_number = nil,
   }



   function reader:read(n)

      if self._eof_emitted then
         return nil, "EOF"
      end


      if self.position > #self.content then
         self._eof_emitted = true
         return nil, "EOF"
      end


      local bytes_available = #self.content - self.position + 1
      local bytes_to_read = math.min(n, bytes_available)


      local chunk = self.content:sub(self.position, self.position + bytes_to_read - 1)
      self.position = self.position + bytes_to_read


      self.line_number = self.line_number + (select(2, chunk:gsub("\n", "")))

      return chunk, nil
   end


   function reader:close()
      self._eof_emitted = true
      return true, nil
   end


   function reader:get_line_number()
      return self.line_number
   end

   return reader
end











function bio.new_string_writer()
   local writer = {
      content = "",
      write = nil,
      close = nil,
      get_content = nil,
   }


   function writer:write(data)
      self.content = self.content .. data
      return #data, nil
   end


   function writer:close()
      return true, nil
   end


   function writer:get_content()
      return self.content
   end

   return writer
end














function bio.new_buffered_reader(reader, max_line_size)
   local br = {
      wrapped = reader,
      buffer = "",
      buffer_pos = 1,
      buffer_size = 0,
      max_line_size = max_line_size,
      read = nil,
      close = nil,
      read_line = nil,
   }


   function br:read(n)
      return self.wrapped:read(n)
   end


   function br:close()
      return self.wrapped:close()
   end


   function br:read_line()
      local line_parts = {}
      local found_newline = false

      while not found_newline do

         if self.buffer_pos > self.buffer_size then
            local chunk, err = self.wrapped:read(self.max_line_size)
            if err == "EOF" then
               if #line_parts == 0 then
                  return nil, "EOF"
               end
               break
            elseif err then
               return nil, err
            end
            self.buffer = chunk
            self.buffer_pos = 1
            self.buffer_size = #chunk
         end


         local newline_pos = self.buffer:find("\n", self.buffer_pos)
         if newline_pos then

            local part = self.buffer:sub(self.buffer_pos, newline_pos - 1)
            if part ~= "" then
               table.insert(line_parts, part)
            end
            self.buffer_pos = newline_pos + 1
            found_newline = true
         else

            local part = self.buffer:sub(self.buffer_pos)
            if part ~= "" then
               table.insert(line_parts, part)
            end
            self.buffer_pos = self.buffer_size + 1
         end
      end

      local line = table.concat(line_parts)

      if line:sub(-1) == "\r" then
         line = line:sub(1, -2)
      end
      return line, nil
   end

   return br
end

















































local fasta = {}







local FastaRecord = {}










local FastaHeader = {}


























function FastaRecord:to_string()
   local result = ">" .. self.identifier .. "\n"

   local pos = 1
   while pos <= #self.sequence do
      local chunk = self.sequence:sub(pos, pos + 79)
      result = result .. chunk .. "\n"
      pos = pos + 80
   end
   return result .. "\n"
end

function FastaRecord:write(writer)
   local _, err = writer:write(">" .. self.identifier .. "\n")
   if err then return false, err end


   local i = 1
   while i <= #self.sequence do
      local chunk = self.sequence:sub(i, i + 79)
      _, err = writer:write(chunk .. "\n")
      if err then return false, err end
      i = i + 80
   end

   return true, nil
end


function FastaHeader:to_string()
   return ""
end

function FastaHeader:write(_)
   return true, nil
end


local function new_fasta_record(identifier, sequence)
   local fasta_rec = {
      identifier = identifier,
      sequence = sequence,
      format = "FASTA",
      to_string = FastaRecord.to_string,
      write = FastaRecord.write,
   }
   return fasta_rec
end

local function new_fasta_header()
   local fasta_head = {
      format = "FASTA",
      to_string = FastaHeader.to_string,
      write = FastaHeader.write,
   }
   return fasta_head
end


function fasta.new_parser(reader, max_line_size)
   local buffered = bio.new_buffered_reader(reader, max_line_size)
   local parser = {
      reader = buffered,
      identifier = "",
      sequence_buffer = {},
      start = true,
      line = 0,
      more = true,
      header = nil,
      next = nil,
      get_format = nil,
   }

   function parser:header()
      return new_fasta_header(), nil
   end

   function parser:get_format()
      return "FASTA"
   end


   local function new_record(self)
      local sequence = table.concat(self.sequence_buffer)
      if sequence == "" then
         return nil, string.format("%s has no sequence", self.identifier)
      end
      self.sequence_buffer = {}
      return new_fasta_record(self.identifier, sequence), nil
   end

   function parser:next()
      if not self.more then
         return nil, "EOF"
      end

      while true do
         local line, err = self.reader:read_line()
         if err == "EOF" then
            self.more = false
            if #self.sequence_buffer > 0 then
               return new_record(self)
            end
            return nil, "EOF"
         elseif err then
            return nil, err
         end

         self.line = self.line + 1


         if #line > 0 and line:sub(1, 1) ~= ";" then

            if line:sub(1, 1) ~= ">" and self.start then

               local err2 = string.format("invalid input: missing sequence identifier for sequence starting at line %d", self.line)
               local fasta_rec, _ = new_record(self)
               return fasta_rec, err2
            elseif line:sub(1, 1) ~= ">" then

               table.insert(self.sequence_buffer, line)
            elseif line:sub(1, 1) == ">" and not self.start then

               local fasta_rec, err3 = new_record(self)
               self.identifier = line:sub(2)
               return fasta_rec, err3
            elseif line:sub(1, 1) == ">" and self.start then

               self.identifier = line:sub(2)
               self.start = false
            end
         end
      end
   end

   return parser
end


















local fastq = {}







local FastqRead = {}













local FastqHeader = {}






















function FastqRead:to_string()
   local result = "@" .. self.identifier

   local keys = {}
   for key in pairs(self.optionals) do
      table.insert(keys, key)
   end
   table.sort(keys)
   for _, key in ipairs(keys) do
      result = result .. " " .. key .. "=" .. self.optionals[key]
   end
   result = result .. "\n" .. self.sequence .. "\n+\n" .. self.quality .. "\n"
   return result
end

function FastqRead:write(writer)
   local ok, err = writer:write(self:to_string())
   if not ok then
      return false, err
   end
   return true, nil
end

function FastqRead:deep_copy()
   local new_read = {
      identifier = self.identifier,
      sequence = self.sequence,
      quality = self.quality,
      format = "FASTQ",
      optionals = {},
      to_string = FastqRead.to_string,
      write = FastqRead.write,
      deep_copy = FastqRead.deep_copy,
   }
   for key, value in pairs(self.optionals) do
      new_read.optionals[key] = value
   end
   return new_read
end


function FastqHeader:to_string()
   return ""
end

function FastqHeader:write(_)
   return true, nil
end


local function new_fastq_read(identifier, sequence, quality, optionals)
   local fastq_read = {
      identifier = identifier,
      sequence = sequence,
      quality = quality,
      optionals = optionals or {},
      format = "FASTQ",
      to_string = FastqRead.to_string,
      write = FastqRead.write,
      deep_copy = FastqRead.deep_copy,
   }
   return fastq_read
end

local function new_fastq_header()
   local fastq_head = {
      format = "FASTQ",
      to_string = FastqHeader.to_string,
      write = FastqHeader.write,
   }
   return fastq_head
end


function fastq.new_parser(reader, max_line_size)
   local buffered = bio.new_buffered_reader(reader, max_line_size)
   local parser = {
      reader = buffered,
      line = 0,
      at_eof = false,
      header = nil,
      next = nil,
      get_format = nil,
   }

   function parser:header()
      return new_fastq_header(), nil
   end

   function parser:get_format()
      return "FASTQ"
   end


   local function parse_optionals(line)
      local optionals = {}
      local parts = line:gmatch("%S+")

      parts()
      for part in parts do
         local key, value = part:match("([^=]+)=(.+)")
         if key and value then
            optionals[key] = value
         end
      end
      return optionals
   end


   local function validate_sequence(seq)
      if #seq <= 0 or seq == "+" then
         return "empty fastq sequence"
      end
      local valid_chars = { ["A"] = true, ["T"] = true, ["G"] = true, ["C"] = true, ["N"] = true }
      for i = 1, #seq do
         local char = seq:sub(i, i)
         if not valid_chars[char] then
            return string.format("Only letters ATGCN are allowed for DNA/RNA in fastq file. Got letter: %s", char)
         end
      end
      return nil
   end

   function parser:next()
      if self.at_eof then
         return nil, "EOF"
      end


      local identifier_line, err = self.reader:read_line()
      self.line = self.line + 1
      if err then
         return nil, err
      end
      if #identifier_line == 0 or identifier_line:sub(1, 1) ~= "@" then
         return nil, string.format("did not find fastq start '@', got to line %d", self.line)
      end

      local identifier = identifier_line:match("^@([^%s]+)")
      local optionals = parse_optionals(identifier_line)


      local sequence, err2 = self.reader:read_line()
      self.line = self.line + 1
      if err2 then
         return nil, err2
      end
      if #sequence <= 0 then
         return nil, string.format("empty fastq sequence for %q, got to line %d", identifier, self.line)
      end

      local seq_err = validate_sequence(sequence)
      if seq_err then
         return nil, seq_err
      end


      local _, err3 = self.reader:read_line()
      self.line = self.line + 1
      if err3 then
         return nil, err3
      end


      local quality, err4 = self.reader:read_line()
      self.line = self.line + 1
      if err4 then
         if err4 == "EOF" then
            self.at_eof = true
         else
            return nil, err4
         end
      end

      if not quality or #quality <= 0 then
         return nil, string.format("empty quality sequence for %q, got to line %d", identifier, self.line)
      end

      if #sequence ~= #quality then
         return nil, string.format("Got different lengths for sequence(%d) and quality(%d)", #sequence, #quality)
      end

      return new_fastq_read(identifier, sequence, quality, optionals), nil
   end

   return parser
end










































local pileup = {}












local PileupLine = {}














local PileupHeader = {}











































function PileupLine:to_string()
   local result = string.format("%s\t%d\t%s\t%d\t%s\t%s\n",
   self.sequence,
   self.position,
   self.reference_base,
   self.read_count,
   table.concat(self.read_results),
   self.quality)
   return result
end

function PileupLine:write(writer)
   local ok, err = writer:write(self:to_string())
   if not ok then
      return false, err
   end
   return true, nil
end


function PileupHeader:to_string()
   return ""
end

function PileupHeader:write(_)
   return true, nil
end


local function new_pileup_line(sequence, position, reference_base,
   read_count, read_results, quality)
   local pileup_line = {
      sequence = sequence,
      position = position,
      reference_base = reference_base,
      read_count = read_count,
      read_results = read_results,
      quality = quality,
      format = "PILEUP",
      to_string = PileupLine.to_string,
      write = PileupLine.write,
   }
   return pileup_line
end

local function new_pileup_header()
   local pileup_head = {
      format = "PILEUP",
      to_string = PileupHeader.to_string,
      write = PileupHeader.write,
   }
   return pileup_head
end


function pileup.new_parser(reader, max_line_size)
   local buffered = bio.new_buffered_reader(reader, max_line_size)
   local parser = {
      reader = buffered,
      line = 0,
      at_eof = false,
      header = nil,
      next = nil,
      get_format = nil,
   }

   function parser:header()
      return new_pileup_header(), nil
   end

   function parser:get_format()
      return "PILEUP"
   end

   function parser:next()
      if self.at_eof then
         return nil, "EOF"
      end


      local line, err = self.reader:read_line()
      if err then
         if err == "EOF" then
            self.at_eof = true
            return nil, "EOF"
         end
         return nil, err
      end

      self.line = self.line + 1


      if #line == 0 then
         if self.at_eof then
            return nil, "EOF"
         end
         return self:next()
      end


      local values = {}
      for value in line:gmatch("[^\t]+") do
         table.insert(values, value)
      end
      if #values ~= 6 then
         return nil, string.format("Error on line %d: Got %d values, expected 6.", self.line, #values)
      end


      local position = tonumber(values[2])
      if not position then
         return nil, string.format("Error on line %d: Invalid position value", self.line)
      end

      local read_count = tonumber(values[4])
      if not read_count then
         return nil, string.format("Error on line %d: Invalid read count value", self.line)
      end


      local read_results = {}
      local results_string = values[5]
      local i = 1
      while i <= #results_string do
         local result_char = results_string:sub(i, i)
         if result_char == " " then
            i = i + 1
         elseif result_char == "^" then

            if i + 2 <= #results_string then
               table.insert(read_results, results_string:sub(i, i + 2))
               i = i + 3
            else
               return nil, string.format("Error on line %d: Invalid read start marker", self.line)
            end
         elseif result_char == "$" then

            if #read_results > 0 then
               read_results[#read_results] = read_results[#read_results] .. "$"
            end
            i = i + 1
         elseif result_char:match("[.,*ATGCNatgcn]") then

            table.insert(read_results, result_char)
            i = i + 1
         elseif result_char:match("[+-]") then

            local indel_size = ""
            local j = i + 1
            while j <= #results_string and results_string:sub(j, j):match("%d") do
               indel_size = indel_size .. results_string:sub(j, j)
               j = j + 1
            end

            if #indel_size == 0 then
               return nil, string.format("Error on line %d: Invalid indel format", self.line)
            end

            local size = tonumber(indel_size)
            if not size then
               return nil, string.format("Error on line %d: Invalid indel size", self.line)
            end

            local indel_end = j + size
            if indel_end > #results_string then
               return nil, string.format("Error on line %d: Indel extends beyond line", self.line)
            end

            local indel_seq = results_string:sub(i, indel_end - 1)
            if not indel_seq:match("^[+-][0-9]+[ATGCNatgcn]+$") then
               return nil, string.format("Error on line %d: Invalid indel sequence", self.line)
            end

            table.insert(read_results, indel_seq)
            i = indel_end
         else
            return nil, string.format("Error on line %d: Invalid character in read results: %s", self.line, result_char)
         end
      end

      return new_pileup_line(
      values[1],
      position,
      values[3],
      read_count,
      read_results,
      values[6]),
      nil
   end

   return parser
end







function pileup.call_mutations(read_results, reference_base, minimal_ratio)

   local reads = {}
   for _, read_result in ipairs(read_results) do
      if #read_result == 1 then

         reads[read_result] = (reads[read_result] or 0) + 1
      else

         reads[read_result] = (reads[read_result] or 0) + 1
      end
   end


   local no_mutation = (reads["."] or 0) + (reads[","] or 0)


   local function has_sufficient_ratio(mutation_count)
      if #read_results == 0 then
         return false
      end
      return (mutation_count / #read_results) > minimal_ratio
   end


   local a_mutation = (reads["A"] or 0) + (reads["a"] or 0)
   local t_mutation = (reads["T"] or 0) + (reads["t"] or 0)
   local g_mutation = (reads["G"] or 0) + (reads["g"] or 0)
   local c_mutation = (reads["C"] or 0) + (reads["c"] or 0)
   local point_indel = reads["*"] or 0

   if has_sufficient_ratio(a_mutation) then
      return {
         type = "point",
         from = reference_base,
         to = "A",
         length = 0,
         total_correct = no_mutation,
         total_mutated = a_mutation,
         total_aligned = #read_results,
      }
   elseif has_sufficient_ratio(t_mutation) then
      return {
         type = "point",
         from = reference_base,
         to = "T",
         length = 0,
         total_correct = no_mutation,
         total_mutated = t_mutation,
         total_aligned = #read_results,
      }
   elseif has_sufficient_ratio(g_mutation) then
      return {
         type = "point",
         from = reference_base,
         to = "G",
         length = 0,
         total_correct = no_mutation,
         total_mutated = g_mutation,
         total_aligned = #read_results,
      }
   elseif has_sufficient_ratio(c_mutation) then
      return {
         type = "point",
         from = reference_base,
         to = "C",
         length = 0,
         total_correct = no_mutation,
         total_mutated = c_mutation,
         total_aligned = #read_results,
      }
   elseif has_sufficient_ratio(point_indel) then
      return {
         type = "point_indel",
         from = reference_base,
         to = "*",
         length = 0,
         total_correct = no_mutation,
         total_mutated = point_indel,
         total_aligned = #read_results,
      }
   end


   for read_result, count in pairs(reads) do
      if #read_result > 1 then
         local first_char = read_result:sub(1, 1)
         if first_char == "-" or first_char == "+" then
            if has_sufficient_ratio(count) then

               local length_str = read_result:match("%d+")
               if length_str then
                  local length = tonumber(length_str)
                  return {
                     type = first_char == "-" and "indel" or "insertion",
                     from = reference_base,
                     to = read_result,
                     length = length,
                     total_correct = no_mutation,
                     total_mutated = count,
                     total_aligned = #read_results,
                  }
               end
            end
         end
      end
   end


   local noisy_count = 0
   for read_result, count in pairs(reads) do
      if read_result ~= "." and read_result ~= "," then
         noisy_count = noisy_count + count
      end
   end

   if has_sufficient_ratio(noisy_count) then
      return {
         type = "noisy",
         from = reference_base,
         to = "?",
         length = 0,
         total_correct = no_mutation,
         total_mutated = noisy_count,
         total_aligned = #read_results,
      }
   end


   return {
      type = "no_mutation",
      from = reference_base,
      to = ".",
      length = 0,
      total_correct = no_mutation,
      total_mutated = noisy_count,
      total_aligned = #read_results,
   }
end




























local sam = {}














sam.DEFAULT_MAX_LINE_SIZE = 1024 * 32 * 2






































local SamHeader = {}





















local SamAlignment = {}





































local function header_write_helper(sb, header_string, header_map, ordered_keys)
   table.insert(sb, header_string)

   for _, key in ipairs(ordered_keys) do
      if header_map[key] then
         table.insert(sb, string.format("\t%s:%s", key, header_map[key]))
      end
   end

   for key, value in pairs(header_map) do

      local skip = false
      for _, ordered_key in ipairs(ordered_keys) do
         if key == ordered_key then
            skip = true
            break
         end
      end
      if not skip then
         table.insert(sb, string.format("\t%s:%s", key, value))
      end
   end
   table.insert(sb, "\n")
end


function SamHeader:to_string()
   local sb = {}
   if self.HD and next(self.HD) ~= nil then
      header_write_helper(sb, "@HD", self.HD, { "VN", "SO", "GO", "SS" })
   end
   for _, sq in ipairs(self.SQ) do
      header_write_helper(sb, "@SQ", sq, { "SN", "LN", "AH", "AN", "AS", "DS", "M5", "SP", "TP", "UR" })
   end
   for _, rg in ipairs(self.RG) do
      header_write_helper(sb, "@RG", rg, { "ID", "BC", "CN", "DS", "DT", "FO", "KS", "LB", "PG", "PI", "PL", "PM", "PU", "SM" })
   end
   for _, pg in ipairs(self.PG) do
      header_write_helper(sb, "@PG", pg, { "ID", "PN", "VN", "CL", "PP", "DS" })
   end
   for _, co in ipairs(self.CO) do
      table.insert(sb, string.format("@CO %s\n", co))
   end
   return table.concat(sb)
end

function SamHeader:write(writer)
   local _, err = writer:write(self:to_string())
   if err then return false, err end
   return true, nil
end


function SamHeader:validate()

   if self.HD.VN then
      if not self.HD.VN:match("^%d+%.%d+$") then
         return string.format("Invalid format for @HD VN: %s", self.HD.VN)
      end
   end


   local sn_map = {}
   for _, sq in ipairs(self.SQ) do

      if sq.SN then
         if sn_map[sq.SN] then
            return string.format("Non-unique @SQ SN: %s", sq.SN)
         end
         sn_map[sq.SN] = true
      end


      if sq.LN then
         local ln = tonumber(sq.LN)
         if not ln or ln < 1 or ln > 2147483647 then
            return string.format("Invalid value for @SQ LN: %s", sq.LN)
         end
      end


      if sq.TP then
         local valid = false
         for _, tp in ipairs({ "linear", "circular" }) do
            if sq.TP == tp then
               valid = true
               break
            end
         end
         if not valid then
            return string.format("Invalid value for @SQ TP: %s", sq.TP)
         end
      end
   end


   if self.HD.SO then
      local valid = false
      for _, so in ipairs({ "unknown", "unsorted", "queryname", "coordinate" }) do
         if self.HD.SO == so then
            valid = true
            break
         end
      end
      if not valid then
         return string.format("Invalid value for @HD SO: %s", self.HD.SO)
      end
   end


   if self.HD.GO then
      local valid = false
      for _, go in ipairs({ "none", "query", "reference" }) do
         if self.HD.GO == go then
            valid = true
            break
         end
      end
      if not valid then
         return string.format("Invalid value for @HD GO: %s", self.HD.GO)
      end
   end


   local rg_id_map = {}
   for _, rg in ipairs(self.RG) do

      if rg.ID then
         if rg_id_map[rg.ID] then
            return string.format("Non-unique @RG ID: %s", rg.ID)
         end
         rg_id_map[rg.ID] = true
      end


      if rg.PL then
         local valid = false
         for _, pl in ipairs({ "CAPILLARY", "DNBSEQ", "ELEMENT", "HELICOS", "ILLUMINA", "IONTORRENT", "LS454", "ONT", "PACBIO", "SOLID", "ULTIMA" }) do
            if rg.PL == pl then
               valid = true
               break
            end
         end
         if not valid then
            return string.format("Invalid value for @RG PL: %s", rg.PL)
         end
      end
   end

   return nil
end


function SamAlignment:to_string()
   local parts = {
      self.QNAME,
      tostring(self.FLAG),
      self.RNAME,
      tostring(self.POS),
      tostring(self.MAPQ),
      self.CIGAR,
      self.RNEXT,
      tostring(self.PNEXT),
      tostring(self.TLEN),
      self.SEQ,
      self.QUAL,
   }

   for _, opt in ipairs(self.optionals) do
      table.insert(parts, string.format("%s:%s:%s", opt.tag, opt.tag_type, opt.data))
   end

   return table.concat(parts, "\t") .. "\n"
end

function SamAlignment:write(writer)
   local _, err = writer:write(self:to_string())
   if err then return false, err end
   return true, nil
end

function SamAlignment:validate()

   if self.FLAG < 0 or self.FLAG > 65535 then
      return "Invalid FLAG range"
   end


   if self.CIGAR ~= "*" then

      local valid_ops = ""
      for op in self.CIGAR:gmatch("%d+[MIDNSHPX=]") do
         valid_ops = valid_ops .. op
      end


      if valid_ops ~= self.CIGAR then
         return "Invalid CIGAR format"
      end
   end



   if not self.QNAME:match("^[!-?A-~]{1,254}$") then
      return "Invalid QNAME format"
   end


   if self.POS < 0 or self.POS > 2147483647 then
      return "Invalid POS range"
   end


   if self.MAPQ < 0 or self.MAPQ > 255 then
      return "Invalid MAPQ range"
   end


   if not self.SEQ:match("^%*|[A-Za-z=.]+$") then
      return "Invalid SEQ format"
   end


   if not self.QUAL:match("^[!-~]+$") then
      return "Invalid QUAL format"
   end

   return nil
end


function sam.new_parser(reader, max_line_size)
   local buffered = bio.new_buffered_reader(reader, max_line_size)
   local parser = {
      reader = buffered,
      line = 0,
      file_header = nil,
      first_line = "",
      read_first_line = false,
      header = nil,
      next = nil,
      get_format = nil,
   }


   local header = {
      HD = {},
      SQ = {},
      RG = {},
      PG = {},
      CO = {},
      format = "SAM",
      to_string = SamHeader.to_string,
      write = SamHeader.write,
      validate = SamHeader.validate,
   }

   local hd_parsed = false


   while true do
      local line
      local err
      line, err = buffered:read_line()
      if err then
         if err == "EOF" then
            if line and line:match("^[^@]") then
               parser.first_line = line
               break
            end

            if hd_parsed then
               parser.file_header = header
               return parser, header, nil
            end
            return nil, nil, "EOF"
         end
         return nil, nil, err
      end
      parser.line = parser.line + 1

      if #line == 0 then
         if not hd_parsed then
            return nil, nil, string.format("Line %d is empty. Empty lines not allowed in headers.", parser.line)
         else
            return nil, nil, string.format("Line %d is empty", parser.line)
         end
      end


      if line:sub(1, 1) ~= "@" then
         parser.first_line = line
         break
      end

      local values = {}
      for value in line:gmatch("[^\t]+") do
         table.insert(values, value)
      end


      if not hd_parsed then
         if values[1] ~= "@HD" then
            return nil, nil, string.format("First line must contain @HD. Got: %s", line)
         end
         for i = 2, #values do
            local k, v = values[i]:match("([^:]+):(.+)")
            if k and v then
               header.HD[k] = v
            end
         end
         hd_parsed = true
      else

         if values[1] == "@CO" then
            table.insert(header.CO, line)
         else

            local section = {}
            for i = 2, #values do
               local k, v = values[i]:match("([^:]+):(.+)")
               if k and v then
                  section[k] = v
               end
            end

            if values[1] == "@SQ" then
               table.insert(header.SQ, section)
            elseif values[1] == "@RG" then
               table.insert(header.RG, section)
            elseif values[1] == "@PG" then
               table.insert(header.PG, section)
            else
               return nil, nil, string.format("Line %d has invalid header tag: %s", parser.line, values[1])
            end
         end
      end
   end

   parser.file_header = header

   function parser:header()
      return self.file_header, nil
   end

   function parser:get_format()
      return "SAM"
   end

   function parser:next()
      local line
      local err


      repeat
         if not self.read_first_line then
            line = self.first_line
            self.read_first_line = true
         else
            line, err = self.reader:read_line()
            if err then
               if err == "EOF" then
                  return nil, "EOF"
               end
               return nil, err
            end
         end
      until #line > 0

      self.line = self.line + 1


      local values = {}
      for value in line:gmatch("[^\t]+") do
         table.insert(values, value)
      end

      if #values < 11 then
         return nil, string.format("Line %d must have at least 11 tab-delimited values. Had %d", self.line, #values)
      end


      local alignment = {
         QNAME = values[1],
         FLAG = tonumber(values[2]),
         RNAME = values[3],
         POS = tonumber(values[4]),
         MAPQ = tonumber(values[5]),
         CIGAR = values[6],
         RNEXT = values[7],
         PNEXT = tonumber(values[8]),
         TLEN = tonumber(values[9]),
         SEQ = values[10],
         QUAL = values[11],
         optionals = {},
         format = "SAM",
         to_string = SamAlignment.to_string,
         write = SamAlignment.write,
         validate = SamAlignment.validate,
      }


      if not alignment.FLAG or not alignment.POS or not alignment.MAPQ or
         not alignment.PNEXT or not alignment.TLEN then
         return nil, string.format("Line %d contains invalid number format", self.line)
      end


      if #values > 11 then
         for i = 12, #values do
            local tag, tag_type, data = values[i]:match("([^:]+):([^:]+):(.+)")
            if not (tag and tag_type and data) then
               return nil, string.format("Line %d has invalid optional field format: %s", self.line, values[i])
            end
            table.insert(alignment.optionals, {
               tag = tag,
               tag_type = tag_type,
               data = data,
            })
         end
      end

      return alignment, nil
   end

   return parser, header, nil
end


function sam.is_primary(alignment)


   local has_secondary = (math.floor(alignment.FLAG / 256) % 2) == 1
   local has_supplementary = (math.floor(alignment.FLAG / 2048) % 2) == 1


   return not (has_secondary or has_supplementary)
end

























local slow5 = {}
















local Slow5Header = {}









local Slow5Read = {}












































local knownend_reasons = {
   ["unknown"] = true,
   ["partial"] = true,
   ["mux_change"] = true,
   ["unblock_mux_change"] = true,
   ["data_service_unblock_mux_change"] = true,
   ["signal_positive"] = true,
   ["signal_negative"] = true,
}


function Slow5Read:to_string()
   local parts = {
      self.read_id,
      tostring(self.read_group_id),
      string.format("%g", self.digitisation),
      string.format("%g", self.offset),
      string.format("%g", self.read_range),
      string.format("%g", self.sampling_rate),
      tostring(self.len_raw_signal),
   }


   local signals = {}
   for i = 1, #self.raw_signal do
      table.insert(signals, tostring(self.raw_signal[i]))
   end
   table.insert(parts, table.concat(signals, ","))


   table.insert(parts, tostring(self.start_time))
   table.insert(parts, tostring(self.read_number))
   table.insert(parts, tostring(self.start_mux))
   table.insert(parts, string.format("%g", self.median_before))
   table.insert(parts, tostring(self.end_reason_map[self.end_reason]))
   table.insert(parts, self.channel_number)

   return table.concat(parts, "\t") .. "\n"
end

function Slow5Read:write(writer)
   local _, err = writer:write(self:to_string())
   if err then return false, err end
   return true, nil
end


function Slow5Header:to_string()
   local sb = {}
   table.insert(sb, "#slow5_version\t" .. self.header_values[1].slow5_version .. "\n")
   table.insert(sb, "#num_read_groups\t" .. tostring(#self.header_values) .. "\n")


   local attributes = {}
   for key in pairs(self.header_values[1].attributes) do
      table.insert(attributes, key)
   end
   table.sort(attributes)

   for _, key in ipairs(attributes) do
      local line = key
      for _, header in ipairs(self.header_values) do
         line = line .. "\t" .. (header.attributes[key] or ".")
      end
      table.insert(sb, line .. "\n")
   end

   return table.concat(sb)
end

function Slow5Header:write(writer)
   local _, err = writer:write(self:to_string())
   if err then return false, err end
   return true, nil
end


function slow5.new_parser(reader, max_line_size)
   local buffered = bio.new_buffered_reader(reader, max_line_size)

   local parser = {
      reader = buffered,
      line = 0,
      headerMap = {},
      endReasonMap = {},
      endReasonSlow5HeaderMap = {},
      header = nil,
      hitEOF = false,
   }


   local header = {
      header_values = {},
      format = "SLOW5",
      to_string = Slow5Header.to_string,
      write = Slow5Header.write,
   }

   local slow5Version = ""
   local numReadGroups = 0


   while true do
      local lineBytes, err = parser.reader:read_line()
      if err then
         if err == "EOF" then
            return nil, err
         end
         return nil, err
      end

      parser.line = parser.line + 1
      local line = lineBytes:gsub("^%s*(.-)%s*$", "%1")


      local values = {}
      for value in line:gmatch("[^\t]+") do
         table.insert(values, value)
      end

      if #values < 2 then
         return nil, string.format("Got following line without tabs: %s", line)
      end


      if numReadGroups == 0 then
         if values[1] == "#slow5_version" then
            slow5Version = values[2]
         elseif values[1] == "#num_read_groups" then
            numReadGroups = tonumber(values[2])
            if not numReadGroups then
               return nil, "Invalid num_read_groups value"
            end


            for id = 0, numReadGroups - 1 do
               table.insert(header.header_values, {
                  read_group_id = id,
                  slow5_version = slow5Version,
                  attributes = {},
                  end_reason_header_map = {},
               })
            end
         end
      else

         if values[1] == "#char*" then
            for _, typeInfo in ipairs(values) do
               if typeInfo:match("enum{") then
                  local enumStr = typeInfo:gsub("enum{(.-)}", "%1")
                  local endReasons = {}
                  for reason in enumStr:gmatch("[^,]+") do
                     table.insert(endReasons, reason)
                  end

                  for index, reason in ipairs(endReasons) do
                     if not knownend_reasons[reason] then
                        return nil, string.format("unknown end reason '%s' found in end_reason enum", reason)
                     end
                     parser.endReasonMap[index - 1] = reason
                     parser.endReasonSlow5HeaderMap[reason] = index - 1
                  end


                  for i = 1, #header.header_values do
                     header.header_values[i].end_reason_header_map = parser.endReasonSlow5HeaderMap
                  end
               end
            end
         elseif values[1] == "#read_id" then

            parser.headerMap[1] = "read_id"
            for i = 2, #values do
               parser.headerMap[i] = values[i]
            end
            break
         else

            if #values ~= numReadGroups + 1 then
               return nil, string.format("Improper amount of information for read groups. Needed %d, got %d", numReadGroups + 1, #values)
            end

            for id = 1, numReadGroups do
               header.header_values[id].attributes[values[1]] = values[id + 1]
            end
         end
      end
   end

   parser.headerValue = header

   function parser:header()
      return self.headerValue, nil
   end

   function parser:get_format()
      return "SLOW5"
   end

   function parser:next()
      if self.hitEOF then
         return nil, "EOF"
      end

      local lineBytes, err = self.reader:read_line()
      if err then
         if err == "EOF" then
            self.hitEOF = true
            if lineBytes == nil then
               return nil, "EOF"
            end
         else
            return nil, err
         end
      end

      self.line = self.line + 1
      local line = lineBytes:gsub("^%s*(.-)%s*$", "%1")


      local values = {}
      for value in line:gmatch("[^\t]+") do
         table.insert(values, value)
      end

      local read = {
         format = "SLOW5",
         to_string = Slow5Read.to_string,
         write = Slow5Read.write,
      }


      for valueIndex, value in ipairs(values) do
         local fieldValue = self.headerMap[valueIndex]
         if value ~= "." then
            if fieldValue == "read_id" then
               read.read_id = value
            elseif fieldValue == "read_group" then
               read.read_group_id = tonumber(value)
               if not read.read_group_id then
                  return nil, string.format("Failed convert read_group '%s' to number on line %d", value, self.line)
               end
            elseif fieldValue == "digitisation" then
               read.digitisation = tonumber(value)
               if not read.digitisation then
                  return nil, string.format("Failed to convert digitisation '%s' to number on line %d", value, self.line)
               end
            elseif fieldValue == "offset" then
               read.offset = tonumber(value)
               if not read.offset then
                  return nil, string.format("Failed to convert offset '%s' to number on line %d", value, self.line)
               end
            elseif fieldValue == "range" then
               read.read_range = tonumber(value)
               if not read.read_range then
                  return nil, string.format("Failed to convert range '%s' to number on line %d", value, self.line)
               end
            elseif fieldValue == "sampling_rate" then
               read.sampling_rate = tonumber(value)
               if not read.sampling_rate then
                  return nil, string.format("Failed to convert sampling_rate '%s' to number on line %d", value, self.line)
               end
            elseif fieldValue == "len_raw_signal" then
               read.len_raw_signal = tonumber(value)
               if not read.len_raw_signal then
                  return nil, string.format("Failed to convert len_raw_signal '%s' to number on line %d", value, self.line)
               end
            elseif fieldValue == "raw_signal" then
               read.raw_signal = {}
               for signal in value:gmatch("[^,]+") do
                  local num = tonumber(signal)
                  if not num then
                     return nil, string.format("Failed to convert raw signal '%s' to number on line %d", signal, self.line)
                  end
                  table.insert(read.raw_signal, num)
               end
            elseif fieldValue == "start_time" then
               read.start_time = tonumber(value)
               if not read.start_time then
                  return nil, string.format("Failed to convert start_time '%s' to number on line %d", value, self.line)
               end
            elseif fieldValue == "read_number" then
               read.read_number = tonumber(value)
               if not read.read_number then
                  return nil, string.format("Failed to convert read_number '%s' to number on line %d", value, self.line)
               end
            elseif fieldValue == "start_mux" then
               read.start_mux = tonumber(value)
               if not read.start_mux then
                  return nil, string.format("Failed to convert start_mux '%s' to number on line %d", value, self.line)
               end
            elseif fieldValue == "median_before" then
               read.median_before = tonumber(value)
               if not read.median_before then
                  return nil, string.format("Failed to convert median_before '%s' to number on line %d", value, self.line)
               end
            elseif fieldValue == "end_reason" then
               local endReasonIndex = tonumber(value)
               if not endReasonIndex then
                  return nil, string.format("Failed to convert end_reason '%s' to number on line %d", value, self.line)
               end
               if not self.endReasonMap[endReasonIndex] then
                  return nil, string.format("End reason out of range. Got '%d' on line %d", endReasonIndex, self.line)
               end
               read.end_reason = self.endReasonMap[endReasonIndex]
               read.end_reason_map = self.endReasonSlow5HeaderMap
            elseif fieldValue == "channel_number" then
               read.channel_number = value
            else
               return nil, string.format("Unknown field to parser '%s' found on line %d", fieldValue, self.line)
            end
         end
      end
      return read, nil
   end

   return parser, nil
end



















local genbank = {}








































































local Feature = {}
















local Genbank = {}












local GenbankParser = {}











































function GenbankParser:header()
   return { format = "GENBANK" }, nil
end

function GenbankParser:get_format()
   return "GENBANK"
end


local function deep_copy_location(loc)
   if not loc then return nil end

   local copy = {
      Start = loc.Start,
      End = loc.End,
      Complement = loc.Complement,
      Join = loc.Join,
      FivePrimePartial = loc.FivePrimePartial,
      ThreePrimePartial = loc.ThreePrimePartial,
      GbkLocationString = loc.GbkLocationString,
      SubLocations = {},
   }


   if loc.SubLocations then
      for _, subloc in ipairs(loc.SubLocations) do
         table.insert(copy.SubLocations, deep_copy_location(subloc))
      end
   end

   return copy
end

local function deep_copy_attributes(attrs)
   if not attrs then return {} end

   local copy = {}
   for k, v in pairs(attrs) do
      copy[k] = {}
      for _, val in ipairs(v) do
         table.insert(copy[k], val)
      end
   end
   return copy
end


local QUALIFIER_INDENT = 21
local FEATURE_INDENT = 5



local function clean_whitespace(s)
   local result, _ = s:gsub("^%s+", "")
   result, _ = result:gsub("%s+$", "")
   result, _ = result:gsub("%s+", " ")
   return result
end

local function count_leading_spaces(line)
   local trimmed, _ = line:gsub("^%s+", "")
   return #line - #trimmed
end


local function parse_simple_location(loc_str)

   local location = {
      Start = 0,
      End = 0,
      Complement = false,
      Join = false,
      FivePrimePartial = false,
      ThreePrimePartial = false,
      GbkLocationString = loc_str,
      SubLocations = {},
   }


   if loc_str:find("<") then
      location.FivePrimePartial = true
      loc_str = loc_str:gsub("<", "")
   end
   if loc_str:find(">") then
      location.ThreePrimePartial = true
      loc_str = loc_str:gsub(">", "")
   end


   if not loc_str:find("%.%.") then

      local pos = tonumber(loc_str)
      if not pos then
         return nil, "Invalid position in location string"
      end

      location.Start = pos - 1
      location.End = pos
   else

      local start_str, end_str = loc_str:match("(%d+)%.%.(%d+)")
      if not start_str or not end_str then
         return nil, "Invalid range format"
      end
      local start = tonumber(start_str)
      local end_pos = tonumber(end_str)
      if not start or not end_pos then
         return nil, "Invalid range numbers"
      end

      location.Start = start - 1
      location.End = end_pos
   end

   return location, nil
end


local function split_location_parts(inner)
   local parts = {}
   local current = ""
   local paren_count = 0

   for i = 1, #inner do
      local char = inner:sub(i, i)
      if char == "(" then
         paren_count = paren_count + 1
         current = current .. char
      elseif char == ")" then
         paren_count = paren_count - 1
         current = current .. char
      elseif char == "," and paren_count == 0 then

         local cleaned = clean_whitespace(current)
         if #cleaned > 0 then
            table.insert(parts, cleaned)
         end
         current = ""
      else
         current = current .. char
      end
   end

   if paren_count ~= 0 then
      return nil, "Unbalanced parentheses"
   end


   local cleaned = clean_whitespace(current)
   if #cleaned > 0 then
      table.insert(parts, cleaned)
   end

   return parts, nil
end


local function parse_location_recursive(expr, depth)
   if depth > 10 then
      return nil, "Maximum recursion depth exceeded"
   end


   local cleaned = clean_whitespace(expr)
   if #cleaned == 0 then
      return nil, "Empty location string"
   end


   if not cleaned:find("%(") then
      return parse_simple_location(cleaned)
   end


   local command, inner = cleaned:match("^(%w+)%((.+)%)$")
   if not command or not inner then
      return nil, "Invalid location format"
   end


   local location = {
      Start = 0,
      End = 0,
      Complement = false,
      Join = false,
      FivePrimePartial = false,
      ThreePrimePartial = false,
      GbkLocationString = expr,
      SubLocations = {},
   }

   if command == "complement" then
      local subloc, err = parse_location_recursive(inner, depth + 1)
      if err then
         return nil, err
      end


      local complement_loc = {
         Start = subloc.Start,
         End = subloc.End,
         Complement = true,
         Join = subloc.Join,
         FivePrimePartial = subloc.FivePrimePartial,
         ThreePrimePartial = subloc.ThreePrimePartial,
         GbkLocationString = expr,
         SubLocations = {},
      }


      for _, sub in ipairs(subloc.SubLocations) do
         table.insert(complement_loc.SubLocations, deep_copy_location(sub))
      end

      table.insert(location.SubLocations, complement_loc)
      return location, nil

   elseif command == "join" then
      location.Join = true


      local parts, err = split_location_parts(inner)
      if err then
         return nil, err
      end


      for _, part in ipairs(parts) do
         local subloc, err2 = parse_location_recursive(part, depth + 1)
         if err2 then
            return nil, err2
         end

         table.insert(location.SubLocations, deep_copy_location(subloc))
      end

      return location, nil
   else
      return nil, "Unknown location command: " .. command
   end
end

local function parse_location(location_string)
   if not location_string or #location_string == 0 then
      return nil, "Empty location string"
   end

   local location, err = parse_location_recursive(location_string, 0)
   if err then
      return nil, err
   end


   if location.Start == 0 and location.End == 0 and
      not location.Join and not location.Complement and
      #location.SubLocations == 1 then

      local subloc = location.SubLocations[1]
      location = deep_copy_location(subloc)
   end

   location.GbkLocationString = location_string
   return location, nil
end


local function finalize_current_feature(state)


   local loc_str = state.current_feature.Location.GbkLocationString
   if loc_str and #loc_str > 0 then
      local parsed_loc, perr = parse_location(loc_str)
      if not perr and parsed_loc then

         parsed_loc.GbkLocationString = loc_str
         state.current_feature.Location = parsed_loc
      end
   end


   local feature_copy = {
      Type = state.current_feature.Type,
      Description = state.current_feature.Description,
      Attributes = deep_copy_attributes(state.current_feature.Attributes),
      Location = deep_copy_location(state.current_feature.Location),
      Sequence = state.current_feature.Sequence,
      SequenceHash = state.current_feature.SequenceHash,
      SequenceHashFunction = state.current_feature.SequenceHashFunction,
      ParentSequence = nil,
   }

   table.insert(state.features_accumulated, feature_copy)


   state.current_feature = {
      Type = "",
      Description = "",
      Attributes = {},
      Location = {
         Start = 0,
         End = 0,
         Complement = false,
         Join = false,
         FivePrimePartial = false,
         ThreePrimePartial = false,
         SubLocations = {},
         GbkLocationString = "",
      },
      Sequence = "",
      SequenceHash = "",
      SequenceHashFunction = "",
   }

   state.current_qualifier_key = ""
   state.current_qualifier_value = ""
   state.in_qualifier = false

end


local function parse_qualifier(line)
   local trimmed = clean_whitespace(line)
   if #trimmed < 2 or trimmed:sub(1, 1) ~= "/" then
      return "", "", false
   end


   trimmed = trimmed:sub(2)


   local equals_pos = trimmed:find("=")
   if not equals_pos then
      return clean_whitespace(trimmed), "", true
   end

   local key = clean_whitespace(trimmed:sub(1, equals_pos - 1))
   local value = trimmed:sub(equals_pos + 1)


   if value:sub(1, 1) == '"' then
      value = value:sub(2, -2)
   end

   return key, clean_whitespace(value), false
end


local function is_location_syntax(text)
   if not text or #text == 0 then return false end


   if text:find("%.%.") then return true end
   if text:find("join%(") then return true end
   if text:find("complement%(") then return true end
   if text:find("[<>]") then return true end
   if text:match("^%d+$") then return true end

   return false
end


local METADATA_SECTIONS = {
   ["LOCUS"] = true,
   ["DEFINITION"] = true,
   ["ACCESSION"] = true,
   ["VERSION"] = true,
   ["KEYWORDS"] = true,
   ["SOURCE"] = true,
   ["REFERENCE"] = true,
   ["COMMENT"] = true,
   ["FEATURES"] = true,
   ["ORIGIN"] = true,
}


local function process_metadata_line(state, line)
   if #line == 0 then
      return "unexpected empty metadata line"
   end


   if line:sub(1, 1) ~= " " or state.current_tag == "FEATURES" then

      if state.current_tag ~= "" then
         local content = table.concat(state.metadata_buffer, " ")
         content = clean_whitespace(content)

         if state.current_tag == "DEFINITION" then
            state.result.Meta.Definition = content
         elseif state.current_tag == "ACCESSION" then
            state.result.Meta.Accession = content
         elseif state.current_tag == "VERSION" then
            state.result.Meta.Version = content
         elseif state.current_tag == "KEYWORDS" then
            state.result.Meta.Keywords = content
         elseif state.current_tag == "SOURCE" then

            local source_parts = {}
            local organism = ""
            local taxonomy = {}

            for _, line_data in ipairs(state.metadata_buffer) do
               if line_data:find("ORGANISM") then
                  local _, org_start = line_data:find("ORGANISM%s*")
                  if org_start then
                     organism = clean_whitespace(line_data:sub(org_start + 1))
                  end
               elseif line_data:find(";") then

                  for part in line_data:gmatch("[^;]+") do
                     local cleaned = clean_whitespace(part)
                     if cleaned:sub(-1) == "." then
                        cleaned = cleaned:sub(1, -2)
                     end
                     if #cleaned > 0 then
                        table.insert(taxonomy, cleaned)
                     end
                  end
               else
                  table.insert(source_parts, clean_whitespace(line_data))
               end
            end

            state.result.Meta.Source = table.concat(source_parts, " ")
            state.result.Meta.Organism = organism
            state.result.Meta.Taxonomy = taxonomy

         elseif state.current_tag == "REFERENCE" then

            local ref = {
               Authors = "",
               Title = "",
               Journal = "",
               PubMed = "",
               Remark = "",
               Range = "",
               Consortium = "",
            }

            local current_field = ""
            for _, line_data in ipairs(state.metadata_buffer) do
               local trimmed = clean_whitespace(line_data)
               if trimmed:find("^AUTHORS") then
                  current_field = "AUTHORS"
                  ref.Authors = clean_whitespace(trimmed:sub(8))
               elseif trimmed:find("^TITLE") then
                  current_field = "TITLE"
                  ref.Title = clean_whitespace(trimmed:sub(6))
               elseif trimmed:find("^JOURNAL") then
                  current_field = "JOURNAL"
                  ref.Journal = clean_whitespace(trimmed:sub(8))
               elseif trimmed:find("^PUBMED") then
                  current_field = "PUBMED"
                  ref.PubMed = clean_whitespace(trimmed:sub(7))
               elseif trimmed:find("^CONSRTM") then
                  current_field = "CONSRTM"
                  ref.Consortium = clean_whitespace(trimmed:sub(8))
               elseif #trimmed > 0 then

                  if current_field == "AUTHORS" then
                     ref.Authors = ref.Authors .. " " .. trimmed
                  elseif current_field == "TITLE" then
                     ref.Title = ref.Title .. " " .. trimmed
                  elseif current_field == "JOURNAL" then
                     ref.Journal = ref.Journal .. " " .. trimmed
                  elseif current_field == "CONSRTM" then
                     ref.Consortium = ref.Consortium .. " " .. trimmed
                  end
               end
            end


            local first_line = state.metadata_buffer[1]
            if first_line then
               local range_start = first_line:find("%(")
               if range_start then
                  ref.Range = first_line:sub(range_start)
               end
            end
            table.insert(state.result.Meta.References, ref)

         elseif state.current_tag == "FEATURES" then
            state.parsing_mode = "features"
         else

            state.result.Meta.Other[state.current_tag] = content
         end
      end


      if state.current_tag ~= "FEATURES" then
         local tag = line:match("^%s*(%S+)")
         if tag and METADATA_SECTIONS[tag] then
            state.current_tag = tag
            if state.current_tag == "FEATURES" then
               state.parsing_mode = "features"
               state.feature_start = true
            end
            state.metadata_buffer = {}
            local content = line:sub(#tag + 1)
            if #content > 0 then
               table.insert(state.metadata_buffer, clean_whitespace(content))
            end
         end
      end
   else

      table.insert(state.metadata_buffer, line)
   end

   return nil
end


local function process_feature_line(state, line)
   local trimmed = clean_whitespace(line)
   if #trimmed == 0 then return nil end

   local indent = count_leading_spaces(line)


   if line:find("BASE COUNT") then

      for count_str, base in line:gmatch("(%d+)%s+([a-zA-Z])") do
         local count = tonumber(count_str)
         if count then
            table.insert(state.result.Meta.BaseCount, {
               Base = base,
               Count = count,
            })
         end
      end
      return nil
   elseif line:find("ORIGIN") or line:find("CONTIG") then
      state.parsing_mode = "sequence"
      finalize_current_feature(state)


      for _, feature in ipairs(state.features_accumulated) do

         local new_feature = {
            Type = feature.Type,
            Description = feature.Description,
            Attributes = deep_copy_attributes(feature.Attributes),
            Location = deep_copy_location(feature.Location),
            Sequence = feature.Sequence,
            SequenceHash = feature.SequenceHash,
            SequenceHashFunction = feature.SequenceHashFunction,
            ParentSequence = state.result,
            get_sequence = Feature.get_sequence,
            store_sequence = Feature.store_sequence,
            copy = Feature.copy,
         }
         new_feature:store_sequence()
         table.insert(state.result.Features, new_feature)
      end

      if line:find("CONTIG") then
         local contig_parts = {}
         for part in line:gmatch("%S+") do
            table.insert(contig_parts, part)
         end
         if #contig_parts > 1 then
            table.remove(contig_parts, 1)
            state.result.Meta.Other["CONTIG"] = table.concat(contig_parts, " ")
         end
      end
      return nil
   end


   if (indent <= QUALIFIER_INDENT and count_leading_spaces(state.prev_line) > indent) or state.feature_start then

      if state.feature_start == nil then
         finalize_current_feature(state)
      else
         state.feature_start = nil
      end


      local parts = {}
      for part in trimmed:gmatch("%S+") do
         table.insert(parts, part)
      end

      if #parts < 2 then
         return "malformed feature line"
      end


      state.current_feature = {
         Type = parts[1],
         Description = "",
         Attributes = {},
         Location = {
            Start = 0,
            End = 0,
            Complement = false,
            Join = false,
            FivePrimePartial = false,
            ThreePrimePartial = false,
            SubLocations = {},
            GbkLocationString = parts[#parts],
         },
         Sequence = "",
         SequenceHash = "",
         SequenceHashFunction = "",
         ParentSequence = nil,
      }

   elseif line:sub(indent + 1, indent + 1) == "/" then

      local key, value, is_empty = parse_qualifier(trimmed)
      if #key > 0 then

         if not state.current_feature.Attributes[key] then
            state.current_feature.Attributes[key] = {}
         end

         if is_empty then

            table.insert(state.current_feature.Attributes[key], "")
         else

            if value and #value > 0 then
               table.insert(state.current_feature.Attributes[key], value)
            end
         end

         state.current_qualifier_key = key
         state.current_qualifier_value = value
         state.in_qualifier = true
      end

   else

      if indent >= QUALIFIER_INDENT then
         if is_location_syntax(trimmed) then

            if #state.current_feature.Location.GbkLocationString > 0 then
               state.current_feature.Location.GbkLocationString =
               state.current_feature.Location.GbkLocationString .. " " .. trimmed
            else
               state.current_feature.Location.GbkLocationString = trimmed
            end
         elseif state.in_qualifier and state.current_qualifier_key ~= "" then

            local current_values = state.current_feature.Attributes[state.current_qualifier_key]
            if current_values and #current_values > 0 then
               local last_idx = #current_values
               current_values[last_idx] = current_values[last_idx] .. " " .. trimmed
            end
         end
      end
   end

   return nil
end


local function build_location_string(location)

   local function build_position_string(pos, partial)
      local s = tostring(pos + 1)
      if partial then
         s = "<" .. s
      end
      return s
   end


   local function build_range_string(start, stop, five_prime, three_prime)
      local s = build_position_string(start, five_prime) ..
      ".." .. tostring(stop)
      if three_prime then
         s = s .. ">"
      end
      return s
   end


   if location.Join then
      local parts = {}
      for _, subloc in ipairs(location.SubLocations) do
         table.insert(parts, build_location_string(subloc))
      end
      return "join(" .. table.concat(parts, ",") .. ")"


   elseif location.Complement then

      local inner_loc = {
         Start = location.Start,
         End = location.End,
         Complement = false,
         Join = location.Join,
         FivePrimePartial = location.FivePrimePartial,
         ThreePrimePartial = location.ThreePrimePartial,
         GbkLocationString = "",
         SubLocations = location.SubLocations,
      }
      return "complement(" .. build_location_string(inner_loc) .. ")"


   else
      return build_range_string(
      location.Start,
      location.End,
      location.FivePrimePartial,
      location.ThreePrimePartial)

   end
end


local function get_feature_sequence(feature, location)
   if not feature.ParentSequence then
      return "", "Feature has no parent sequence"
   end

   if not feature.ParentSequence.Sequence then
      return "", "Parent sequence is missing"
   end

   local sequence_parts = {}


   if location.SubLocations and #location.SubLocations > 0 then
      for _, subloc in ipairs(location.SubLocations) do
         local seq, err = get_feature_sequence(feature, subloc)
         if err then
            return "", err
         end
         table.insert(sequence_parts, seq)
      end
   else

      if location.Start >= 0 and location.End > location.Start and
         location.End <= #feature.ParentSequence.Sequence then

         local seq = feature.ParentSequence.Sequence:sub(location.Start + 1, location.End)
         table.insert(sequence_parts, seq)
      else
         return "", "Invalid location bounds"
      end
   end


   local sequence = table.concat(sequence_parts)


   if location.Complement then
      sequence = transform.reverse_complement(sequence)
   end

   return sequence, nil
end


function Feature:get_sequence()
   return get_feature_sequence(self, self.Location)
end

function Feature:store_sequence()
   if self.Sequence ~= "" then
      return self.Sequence, nil
   end

   local seq, err = get_feature_sequence(self, self.Location)
   if err == nil then
      self.Sequence = seq
   end
   return seq, err
end

function Feature:copy()

   local copy = {
      Type = self.Type,
      Description = self.Description,
      Attributes = deep_copy_attributes(self.Attributes),
      Location = deep_copy_location(self.Location),
      Sequence = self.Sequence,
      SequenceHash = self.SequenceHash,
      SequenceHashFunction = self.SequenceHashFunction,
      ParentSequence = self.ParentSequence,


      get_sequence = self.get_sequence,
      store_sequence = self.store_sequence,
      copy = self.copy,
   }

   return copy
end


function Genbank:store_feature_sequences()
   for _, feature in ipairs(self.Features) do
      local _, err = feature:store_sequence()
      if err then
         return err
      end
   end
   return nil
end

function Genbank:add_feature(feature)

   local new_feature = feature:copy()
   new_feature.ParentSequence = self
   table.insert(self.Features, new_feature)
   return nil
end


local function build_feature_string(feature)
   if not feature then return "" end


   local type_space = 16 - #feature.Type
   if type_space < 0 then type_space = 0 end
   local white_space_trail = string.rep(" ", type_space)


   local location = feature.Location.GbkLocationString
   if #location == 0 then
      location = build_location_string(feature.Location)
   end


   local feature_string = string.rep(" ", FEATURE_INDENT) ..
   feature.Type ..
   white_space_trail ..
   location .. "\n"


   if feature.Attributes then

      local keys = {}
      for k in pairs(feature.Attributes) do
         table.insert(keys, k)
      end
      table.sort(keys)


      for _, key in ipairs(keys) do
         local values = feature.Attributes[key]
         if values then
            for _, value in ipairs(values) do
               feature_string = feature_string ..
               string.rep(" ", QUALIFIER_INDENT) ..
               "/" .. key
               if value and #value > 0 then
                  feature_string = feature_string .. "=\"" .. value .. "\""
               end
               feature_string = feature_string .. "\n"
            end
         end
      end
   end

   return feature_string
end


function genbank.parse_locus(locus_string)
   local locus = {
      Name = "",
      SequenceLength = "",
      MoleculeType = "",
      GenbankDivision = "",
      ModificationDate = "",
      SequenceCoding = "",
      Circular = false,
   }


   local parts = {}
   for part in locus_string:gmatch("%S+") do
      table.insert(parts, part)
   end

   if #parts < 2 then
      return locus
   end


   locus.Name = parts[2]


   local base_sequence = locus_string:match("%s+(%d+)%s+([a-zA-Z][a-zA-Z])%s+")
   if base_sequence then
      local length, coding = base_sequence:match("(%d+)%s+([a-zA-Z][a-zA-Z])")
      if length and coding then
         locus.SequenceLength = length
         locus.SequenceCoding = coding
      end
   end


   for _, mol_type in ipairs({
         "DNA", "genomic DNA", "genomic RNA", "mRNA", "tRNA", "rRNA",
         "other RNA", "other DNA", "transcribed RNA", "viral cRNA",
         "unassigned DNA", "unassigned RNA",
      }) do
      if locus_string:find("%s+" .. mol_type .. "%s+") then
         locus.MoleculeType = mol_type
         break
      end
   end


   if locus_string:find("%s+circular%s+") then
      locus.Circular = true
   end


   for _, division in ipairs({
         "PRI", "ROD", "MAM", "VRT", "INV", "PLN", "BCT", "VRL",
         "PHG", "SYN", "UNA", "EST", "PAT", "STS", "GSS", "HTG",
         "HTC", "ENV",
      }) do
      if locus_string:find("%s+" .. division .. "%s+") then
         locus.GenbankDivision = division
         break
      end
   end


   local date = locus_string:match("%d%d%-[A-Z][A-Z][A-Z]-%d%d%d%d")
   if date then
      locus.ModificationDate = date
   end

   return locus
end


local function wrap_text(text, width)
   local lines = {}
   local line = ""
   local space_left = width

   for word in text:gmatch("%S+") do
      if #word + 1 > space_left then
         if #line > 0 then
            table.insert(lines, line)
         end
         line = word
         space_left = width - #word
      else
         if #line > 0 then
            line = line .. " " .. word
            space_left = space_left - (#word + 1)
         else
            line = word
            space_left = width - #word
         end
      end
   end

   if #line > 0 then
      table.insert(lines, line)
   end

   return lines
end


function Genbank:write_to(writer)
   local written = 0


   local function write_line(line)
      local n, err = writer:write(line)
      if err then
         return err
      end
      written = written + n
      return nil
   end


   local locus = self.Meta.Locus
   local shape = locus.Circular and "circular" or "linear"
   local fivespace = string.rep(" ", 5)

   local locus_data = locus.Name .. fivespace ..
   locus.SequenceLength .. " bp" .. fivespace ..
   locus.MoleculeType .. fivespace ..
   shape .. fivespace ..
   locus.GenbankDivision .. fivespace ..
   locus.ModificationDate

   local err = write_line("LOCUS" .. string.rep(" ", 7) .. locus_data .. "\n")
   if err then return written, err end


   local meta_sections = {
      { "DEFINITION", self.Meta.Definition },
      { "ACCESSION", self.Meta.Accession },
      { "VERSION", self.Meta.Version },
      { "KEYWORDS", self.Meta.Keywords },
      { "SOURCE", self.Meta.Source },
      { "  ORGANISM", self.Meta.Organism },
   }

   for _, section in ipairs(meta_sections) do
      if section[2] and #section[2] > 0 then

         local indent = section[1] == "  ORGANISM" and 12 or #section[1]
         local content = section[2]
         local wrapped = wrap_text(content, 80 - indent - 1)


         err = write_line(section[1] .. string.rep(" ", 12 - indent) .. wrapped[1] .. "\n")
         if err then return written, err end


         for i = 2, #wrapped do
            err = write_line(string.rep(" ", 12) .. wrapped[i] .. "\n")
            if err then return written, err end
         end
      end
   end


   if #self.Meta.Taxonomy > 0 then
      local taxonomy = table.concat(self.Meta.Taxonomy, "; ") .. "."
      local wrapped = wrap_text(taxonomy, 68)
      for _, line in ipairs(wrapped) do
         err = write_line(string.rep(" ", 12) .. line .. "\n")
         if err then return written, err end
      end
   end


   for i, ref in ipairs(self.Meta.References) do

      local ref_header = string.format("REFERENCE   %d", i)
      if ref.Range and #ref.Range > 0 then
         ref_header = ref_header .. "  " .. ref.Range
      end
      err = write_line(ref_header .. "\n")
      if err then return written, err end


      local ref_fields = {
         { "  AUTHORS", ref.Authors },
         { "  TITLE", ref.Title },
         { "  JOURNAL", ref.Journal },
         { "  PUBMED", ref.PubMed },
         { "  CONSRTM", ref.Consortium },
      }

      for _, field in ipairs(ref_fields) do
         if field[2] and #field[2] > 0 then
            local wrapped = wrap_text(field[2], 68)
            err = write_line(field[1] .. string.rep(" ", 12 - #field[1]) .. wrapped[1] .. "\n")
            if err then return written, err end

            for j = 2, #wrapped do
               err = write_line(string.rep(" ", 12) .. wrapped[j] .. "\n")
               if err then return written, err end
            end
         end
      end
   end

   for tag, content in pairs(self.Meta.Other) do
      if content and #content > 0 then

         if tag == "COMMENT" then

            local wrapped = wrap_text(content, 68)
            err = write_line("COMMENT             " .. wrapped[1] .. "\n")
            if err then return written, err end


            for i = 2, #wrapped do
               err = write_line(string.rep(" ", 21) .. wrapped[i] .. "\n")
               if err then return written, err end
            end
         else

            local wrapped = wrap_text(content, 68)
            err = write_line(tag .. string.rep(" ", 12 - #tag) .. wrapped[1] .. "\n")
            if err then return written, err end

            for i = 2, #wrapped do
               err = write_line(string.rep(" ", 12) .. wrapped[i] .. "\n")
               if err then return written, err end
            end
         end
      end
   end


   err = write_line("FEATURES             Location/Qualifiers\n")
   if err then return written, err end

   for _, feature in ipairs(self.Features) do
      err = write_line(build_feature_string(feature))
      if err then return written, err end
   end


   if #self.Meta.BaseCount > 0 then
      err = write_line("BASE COUNT   ")
      if err then return written, err end

      for _, base_count in ipairs(self.Meta.BaseCount) do
         err = write_line(string.format(" %d %s", base_count.Count, base_count.Base))
         if err then return written, err end
      end
      err = write_line("\n")
      if err then return written, err end
   end


   err = write_line("ORIGIN\n")
   if err then return written, err end


   for i = 1, #self.Sequence, 60 do

      local line_num = tostring(i)
      local leading_spaces = string.rep(" ", 9 - #line_num)
      err = write_line(leading_spaces .. line_num .. " ")
      if err then return written, err end


      local block = self.Sequence:sub(i, math.min(i + 59, #self.Sequence))
      for j = 1, #block, 10 do
         local chunk = block:sub(j, math.min(j + 9, #block))
         err = write_line(chunk)
         if err then return written, err end

         if j + 9 < #block then
            err = write_line(" ")
            if err then return written, err end
         end
      end

      err = write_line("\n")
      if err then return written, err end
   end


   err = write_line("//\n")
   if err then return written, err end

   return written, nil
end

function GenbankParser:next()

   if not self.parameters then
      self.parameters = {
         parsing_mode = "metadata",
         in_feature = false,
         in_qualifier = false,
         line_number = 0,
         current_line = "",
         prev_line = "",
         feature_text = {},
         sequence_builder = {},
         metadata_buffer = {},
         current_tag = "",
         current_feature = {
            Type = "",
            Description = "",
            Attributes = {},
            Location = {
               Start = 0,
               End = 0,
               Complement = false,
               Join = false,
               FivePrimePartial = false,
               ThreePrimePartial = false,
               SubLocations = {},
               GbkLocationString = "",
            },
            Sequence = "",
            SequenceHash = "",
            SequenceHashFunction = "",
         },
         current_qualifier_key = "",
         current_qualifier_value = "",
         features_accumulated = {},
         result = {
            Meta = {
               Date = "",
               Definition = "",
               Accession = "",
               Version = "",
               Keywords = "",
               Organism = "",
               Source = "",
               Taxonomy = {},
               Origin = "",
               Locus = {
                  Name = "",
                  SequenceLength = "",
                  MoleculeType = "",
                  GenbankDivision = "",
                  ModificationDate = "",
                  SequenceCoding = "",
                  Circular = false,
               },
               References = {},
               BaseCount = {},
               Other = {},
               Name = "",
               SequenceHash = "",
               SequenceHashFunction = "",
            },
            Features = {},
            Sequence = "",
            format = "GENBANK",
         },
      }
   end

   local found_locus = false

   while true do
      local line, err = self.reader:read_line()
      if err then
         if err == "EOF" then
            if not found_locus then
               return nil, "EOF"
            elseif self.parameters.parsing_mode == "sequence" then

               self.parameters.result.Sequence = table.concat(self.parameters.sequence_builder)
               local result = self.parameters.result

               self.parameters = nil
               return result, nil
            else
               return nil, "Unexpected EOF in " .. self.parameters.parsing_mode .. " section"
            end
         end
         return nil, err
      end


      self.parameters.prev_line = self.parameters.current_line
      self.parameters.current_line = line
      self.parameters.line_number = self.parameters.line_number + 1


      if not found_locus then
         if line:find("LOCUS") then
            found_locus = true
            self.parameters.result.Meta.Locus = genbank.parse_locus(line)
         end
      else

         local err2
         if self.parameters.parsing_mode == "metadata" then
            err2 = process_metadata_line(self.parameters, line)
         elseif self.parameters.parsing_mode == "features" then
            err2 = process_feature_line(self.parameters, line)
         elseif self.parameters.parsing_mode == "sequence" then

            if line:sub(1, 2) == "//" then

               self.parameters.result.Sequence = table.concat(self.parameters.sequence_builder)
               local result = self.parameters.result
               result.write_to = Genbank.write_to

               self.parameters = nil
               return result, nil
            else

               local cleaned = line:gsub("[^a-zA-Z]", "")
               if #cleaned > 0 then
                  table.insert(self.parameters.sequence_builder, cleaned)
               end
            end
         else
            return nil, "Unknown parse state: " .. self.parameters.parsing_mode
         end

         if err2 then
            return nil, err2
         end
      end
   end
end


function genbank.new_parser(reader, max_line_size)
   if not reader then
      error("reader cannot be nil")
   end


   if not max_line_size or max_line_size < 1 then
      max_line_size = bio.DEFAULT_MAX_LENGTHS["GENBANK"]
   end


   local buffered_reader = bio.new_buffered_reader(reader, max_line_size)


   local parser = {
      reader = buffered_reader,
      parameters = nil,
      format = "GENBANK",


      header = GenbankParser.header,
      next = GenbankParser.next,
      get_format = GenbankParser.get_format,
   }

   return parser
end












local fragment_frequencies = {
   ["AAAA"] = {
      ["AAAA"] = 635,
      ["AAAC"] = 3,
      ["AAAG"] = 1,
      ["AAAT"] = 8,
      ["AATA"] = 2,
      ["ATAA"] = 7,
      ["TAAA"] = 4,
      ["TTTC"] = 3,
      ["TTTG"] = 1,
   },
   ["AAAC"] = {
      ["AAAA"] = 8,
      ["AAAC"] = 476,
      ["AAAG"] = 1,
      ["AAAT"] = 4,
      ["AACC"] = 8,
      ["AAGC"] = 5,
      ["AATC"] = 12,
      ["ACAC"] = 4,
      ["AGAC"] = 15,
      ["ATAC"] = 36,
      ["GAAC"] = 39,
      ["GTTG"] = 31,
      ["TAAC"] = 60,
   },
   ["AAAG"] = {
      ["AAAA"] = 40,
      ["AAAC"] = 4,
      ["AAAG"] = 596,
      ["AAAT"] = 1,
      ["AACG"] = 5,
      ["AAGG"] = 1,
      ["AATG"] = 6,
      ["AGAG"] = 6,
      ["ATAG"] = 40,
      ["CAAG"] = 14,
      ["GAAG"] = 28,
      ["TAAG"] = 60,
   },
   ["AAAT"] = {
      ["AAAA"] = 16,
      ["AAAC"] = 45,
      ["AAAG"] = 2,
      ["AAAT"] = 642,
      ["AACT"] = 5,
      ["AAGT"] = 2,
      ["ATCT"] = 6,
      ["ATGT"] = 1,
      ["ATTA"] = 44,
      ["ATTC"] = 15,
      ["ATTG"] = 13,
   },
   ["AACA"] = {
      ["AAAA"] = 2,
      ["AACA"] = 493,
      ["AACC"] = 7,
      ["AACG"] = 20,
      ["AACT"] = 26,
      ["AAGA"] = 1,
      ["AATA"] = 3,
      ["ACCA"] = 6,
      ["AGCA"] = 10,
      ["ATCA"] = 9,
      ["TACA"] = 29,
      ["TGTC"] = 3,
      ["TGTG"] = 3,
   },
   ["AACC"] = {
      ["AAAC"] = 20,
      ["AACA"] = 16,
      ["AACC"] = 411,
      ["AACG"] = 3,
      ["AACT"] = 16,
      ["AATC"] = 16,
      ["ACCC"] = 13,
      ["AGCC"] = 24,
      ["ATCC"] = 23,
      ["GACC"] = 33,
      ["GGTG"] = 16,
      ["TACC"] = 59,
   },
   ["AACG"] = {
      ["AAAG"] = 18,
      ["AACA"] = 64,
      ["AACC"] = 3,
      ["AACG"] = 548,
      ["AACT"] = 8,
      ["AAGG"] = 1,
      ["AATG"] = 18,
      ["ACCG"] = 36,
      ["AGCG"] = 49,
      ["ATAG"] = 1,
      ["ATCG"] = 41,
      ["CACG"] = 24,
      ["GACG"] = 46,
      ["TACG"] = 73,
   },
   ["AACT"] = {
      ["AAAC"] = 1,
      ["AAAT"] = 6,
      ["AACA"] = 56,
      ["AACC"] = 44,
      ["AACG"] = 14,
      ["AACT"] = 552,
      ["AGGA"] = 1,
      ["AGGT"] = 10,
      ["AGTA"] = 57,
      ["AGTC"] = 25,
      ["AGTG"] = 6,
      ["ATCG"] = 1,
      ["ATCT"] = 15,
   },
   ["AAGA"] = {
      ["AAAA"] = 7,
      ["AACA"] = 2,
      ["AAGA"] = 766,
      ["AAGC"] = 8,
      ["AAGG"] = 6,
      ["AAGT"] = 12,
      ["ACGA"] = 1,
      ["ATGA"] = 6,
      ["TAGA"] = 12,
      ["TCTC"] = 4,
      ["TCTG"] = 1,
   },
   ["AAGC"] = {
      ["AAAC"] = 15,
      ["AAAT"] = 1,
      ["AACC"] = 13,
      ["AAGA"] = 6,
      ["AAGC"] = 486,
      ["AAGG"] = 1,
      ["AAGT"] = 38,
      ["AATC"] = 9,
      ["ACGC"] = 9,
      ["AGGC"] = 17,
      ["ATGC"] = 26,
      ["GAGC"] = 44,
      ["GCTG"] = 35,
      ["TAGC"] = 71,
   },
   ["AAGG"] = {
      ["AAAG"] = 11,
      ["AACG"] = 20,
      ["AAGA"] = 61,
      ["AAGC"] = 4,
      ["AAGG"] = 637,
      ["AAGT"] = 12,
      ["AATG"] = 7,
      ["ACGG"] = 9,
      ["AGAG"] = 1,
      ["AGGG"] = 15,
      ["AGTG"] = 1,
      ["ATGG"] = 26,
      ["CAGG"] = 27,
      ["GAGG"] = 54,
      ["TAGG"] = 60,
   },
   ["AAGT"] = {
      ["AAAT"] = 9,
      ["AACT"] = 18,
      ["AAGA"] = 22,
      ["AAGC"] = 62,
      ["AAGG"] = 6,
      ["AAGT"] = 670,
      ["ACTA"] = 35,
      ["ACTC"] = 19,
      ["ACTG"] = 13,
      ["AGGT"] = 3,
      ["ATGT"] = 19,
      ["GAGC"] = 1,
      ["TCTC"] = 1,
   },
   ["AATA"] = {
      ["AAAA"] = 5,
      ["AACA"] = 9,
      ["AATA"] = 556,
      ["AATC"] = 18,
      ["AATG"] = 12,
      ["ACTA"] = 1,
      ["AGTA"] = 2,
      ["ATTA"] = 2,
      ["TATC"] = 4,
   },
   ["AATC"] = {
      ["AAAC"] = 16,
      ["AACC"] = 33,
      ["AAGC"] = 12,
      ["AATA"] = 9,
      ["AATC"] = 484,
      ["AATG"] = 2,
      ["ACTC"] = 5,
      ["AGTC"] = 7,
      ["ATTC"] = 10,
      ["GATG"] = 12,
      ["TATC"] = 46,
   },
   ["AATG"] = {
      ["AAAG"] = 50,
      ["AACG"] = 41,
      ["AAGG"] = 13,
      ["AATA"] = 49,
      ["AATC"] = 3,
      ["AATG"] = 625,
      ["ACTG"] = 9,
      ["AGTG"] = 17,
      ["ATTG"] = 14,
      ["GATG"] = 32,
      ["TATG"] = 49,
   },
   ["ACAA"] = {
      ["ACAA"] = 479,
      ["ACAC"] = 7,
      ["ACAG"] = 2,
      ["ACCA"] = 1,
      ["ACGA"] = 3,
      ["ACTA"] = 5,
      ["ATGT"] = 8,
      ["TTGA"] = 10,
      ["TTGC"] = 2,
      ["TTGG"] = 7,
   },
   ["ACAC"] = {
      ["AAAC"] = 8,
      ["ACAA"] = 2,
      ["ACAC"] = 360,
      ["ACCC"] = 15,
      ["ACGC"] = 21,
      ["ACTC"] = 16,
      ["AGAC"] = 4,
      ["ATAC"] = 14,
      ["ATGT"] = 12,
      ["GTGC"] = 70,
      ["GTGG"] = 58,
      ["GTTG"] = 1,
      ["TAAC"] = 6,
      ["TCAC"] = 61,
   },
   ["ACAG"] = {
      ["AAAG"] = 5,
      ["ACAA"] = 38,
      ["ACAC"] = 5,
      ["ACAG"] = 593,
      ["ACCG"] = 11,
      ["ACGG"] = 14,
      ["ACTA"] = 1,
      ["ACTG"] = 14,
      ["AGAG"] = 6,
      ["ATAG"] = 12,
      ["ATGT"] = 11,
      ["CAAG"] = 2,
      ["CTCG"] = 1,
      ["CTGG"] = 82,
      ["GCAG"] = 109,
      ["TCAG"] = 56,
      ["TGAG"] = 1,
   },
   ["ACCA"] = {
      ["AACA"] = 4,
      ["ACAA"] = 2,
      ["ACCA"] = 468,
      ["ACCC"] = 19,
      ["ACCG"] = 20,
      ["ACTA"] = 4,
      ["AGGT"] = 19,
      ["ATCA"] = 5,
      ["TGGA"] = 21,
      ["TGGC"] = 27,
      ["TGGG"] = 17,
   },
   ["ACCC"] = {
      ["AACC"] = 12,
      ["ACAC"] = 23,
      ["ACCA"] = 10,
      ["ACCC"] = 393,
      ["ACCG"] = 4,
      ["ACGC"] = 2,
      ["ACTC"] = 23,
      ["AGCC"] = 4,
      ["AGGT"] = 20,
      ["ATAC"] = 1,
      ["ATCC"] = 17,
      ["GGGC"] = 58,
      ["GGGG"] = 44,
      ["GTCC"] = 1,
      ["GTGG"] = 2,
      ["TCCC"] = 45,
   },
   ["ACCG"] = {
      ["AACG"] = 14,
      ["AACT"] = 1,
      ["AATG"] = 1,
      ["ACAG"] = 31,
      ["ACCA"] = 76,
      ["ACCC"] = 6,
      ["ACCG"] = 521,
      ["ACGG"] = 5,
      ["ACTG"] = 37,
      ["AGCG"] = 10,
      ["AGGT"] = 33,
      ["ATCG"] = 20,
      ["CGGG"] = 61,
      ["GCCG"] = 99,
      ["TCCG"] = 74,
   },
   ["ACGA"] = {
      ["AAGA"] = 5,
      ["ACAA"] = 27,
      ["ACCA"] = 6,
      ["ACGA"] = 714,
      ["ACGC"] = 38,
      ["ACGG"] = 26,
      ["ACTA"] = 3,
      ["AGGA"] = 3,
      ["ATGA"] = 4,
      ["TAGA"] = 1,
      ["TCGC"] = 50,
      ["TCGG"] = 19,
      ["TTGC"] = 1,
   },
   ["ACGC"] = {
      ["AAGC"] = 12,
      ["AATC"] = 1,
      ["ACAC"] = 30,
      ["ACCC"] = 16,
      ["ACGA"] = 8,
      ["ACGC"] = 335,
      ["ACTC"] = 6,
      ["AGGC"] = 11,
      ["ATGC"] = 32,
      ["GACC"] = 1,
      ["GCGG"] = 40,
      ["GCTG"] = 2,
      ["TAGC"] = 24,
      ["TCGC"] = 63,
      ["TGGC"] = 1,
   },
   ["ACGG"] = {
      ["AAGG"] = 19,
      ["AATG"] = 1,
      ["ACAG"] = 43,
      ["ACCG"] = 19,
      ["ACGA"] = 85,
      ["ACGC"] = 4,
      ["ACGG"] = 522,
      ["ACTG"] = 15,
      ["AGGG"] = 12,
      ["ATGA"] = 1,
      ["ATGG"] = 28,
      ["CAGG"] = 4,
      ["GAGG"] = 1,
      ["GCGG"] = 119,
      ["TAGG"] = 9,
      ["TCGC"] = 4,
      ["TCGG"] = 59,
   },
   ["ACTA"] = {
      ["AAGT"] = 35,
      ["ACAA"] = 14,
      ["ACCA"] = 8,
      ["ACGA"] = 1,
      ["ACTA"] = 631,
      ["ACTC"] = 22,
      ["ACTG"] = 25,
      ["ATGT"] = 2,
      ["ATTA"] = 1,
      ["TAGA"] = 3,
      ["TAGC"] = 5,
      ["TAGG"] = 4,
   },
   ["ACTC"] = {
      ["AAGT"] = 19,
      ["AATC"] = 4,
      ["ACAC"] = 35,
      ["ACCC"] = 30,
      ["ACGC"] = 37,
      ["ACTA"] = 14,
      ["ACTC"] = 509,
      ["ACTG"] = 3,
      ["AGTC"] = 2,
      ["ATTC"] = 7,
      ["GAGC"] = 50,
      ["GAGG"] = 37,
      ["TCTC"] = 24,
   },
   ["ACTG"] = {
      ["AAAG"] = 1,
      ["AAGT"] = 13,
      ["AATG"] = 3,
      ["ACAA"] = 1,
      ["ACAG"] = 71,
      ["ACCA"] = 1,
      ["ACCG"] = 56,
      ["ACGG"] = 31,
      ["ACTA"] = 44,
      ["ACTC"] = 8,
      ["ACTG"] = 595,
      ["AGTG"] = 2,
      ["ATTG"] = 7,
      ["CAGG"] = 68,
      ["GCTG"] = 59,
      ["TCTG"] = 27,
   },
   ["AGAA"] = {
      ["AAAA"] = 7,
      ["ACAA"] = 7,
      ["AGAA"] = 748,
      ["AGAC"] = 14,
      ["AGAG"] = 11,
      ["AGCA"] = 2,
      ["AGGA"] = 1,
      ["AGTA"] = 11,
      ["ATAA"] = 4,
      ["ATCT"] = 26,
      ["TTCA"] = 8,
      ["TTCC"] = 7,
      ["TTCG"] = 6,
   },
   ["AGAC"] = {
      ["AAAC"] = 37,
      ["ACAC"] = 35,
      ["AGAA"] = 2,
      ["AGAC"] = 495,
      ["AGAG"] = 2,
      ["AGCC"] = 26,
      ["AGGA"] = 1,
      ["AGGC"] = 28,
      ["AGTC"] = 32,
      ["ATAC"] = 29,
      ["ATCT"] = 21,
      ["GTCC"] = 42,
      ["GTCG"] = 51,
      ["TGAC"] = 93,
   },
   ["AGAG"] = {
      ["AAAG"] = 20,
      ["ACAA"] = 1,
      ["ACAG"] = 21,
      ["AGAA"] = 58,
      ["AGAC"] = 9,
      ["AGAG"] = 567,
      ["AGCG"] = 17,
      ["AGGA"] = 1,
      ["AGGG"] = 16,
      ["AGTG"] = 17,
      ["ATAG"] = 5,
      ["ATCT"] = 10,
      ["CTCG"] = 13,
      ["GGAG"] = 38,
      ["TGAG"] = 47,
   },
   ["AGCA"] = {
      ["AACA"] = 20,
      ["AACT"] = 1,
      ["ACCA"] = 16,
      ["AGAA"] = 8,
      ["AGCA"] = 487,
      ["AGCC"] = 34,
      ["AGCG"] = 74,
      ["AGTA"] = 21,
      ["ATCA"] = 19,
      ["ATGC"] = 1,
      ["TGCC"] = 25,
      ["TGCG"] = 14,
   },
   ["AGCC"] = {
      ["AACC"] = 28,
      ["ACCC"] = 31,
      ["AGAC"] = 33,
      ["AGCA"] = 17,
      ["AGCC"] = 403,
      ["AGCG"] = 30,
      ["AGGC"] = 3,
      ["AGTC"] = 67,
      ["ATCC"] = 28,
      ["GGCG"] = 30,
      ["GGTG"] = 1,
      ["GTCC"] = 1,
      ["TGCC"] = 79,
      ["TTCC"] = 1,
   },
   ["AGCG"] = {
      ["AACG"] = 29,
      ["ACAG"] = 1,
      ["ACCG"] = 24,
      ["AGAG"] = 25,
      ["AGCA"] = 64,
      ["AGCC"] = 7,
      ["AGCG"] = 395,
      ["AGGG"] = 1,
      ["AGTG"] = 48,
      ["ATCG"] = 25,
      ["GGCG"] = 52,
      ["TCCG"] = 1,
      ["TGCG"] = 60,
      ["TTCG"] = 1,
   },
   ["AGGA"] = {
      ["AACT"] = 1,
      ["AAGA"] = 6,
      ["ACGA"] = 8,
      ["AGAA"] = 23,
      ["AGAG"] = 1,
      ["AGCA"] = 3,
      ["AGGA"] = 789,
      ["AGGC"] = 52,
      ["AGGG"] = 26,
      ["AGGT"] = 59,
      ["AGTA"] = 10,
      ["ATGA"] = 4,
      ["TCCC"] = 10,
      ["TCCG"] = 12,
      ["TGGA"] = 9,
   },
   ["AGGC"] = {
      ["AAGC"] = 27,
      ["ACGC"] = 36,
      ["AGAC"] = 29,
      ["AGCC"] = 23,
      ["AGGA"] = 12,
      ["AGGC"] = 609,
      ["AGGT"] = 59,
      ["AGTC"] = 51,
      ["ATGC"] = 29,
      ["GCCG"] = 47,
      ["GCTG"] = 1,
      ["GGGC"] = 52,
      ["TGGC"] = 78,
   },
   ["AGGG"] = {
      ["AAGG"] = 20,
      ["ACGG"] = 28,
      ["AGAG"] = 20,
      ["AGCG"] = 20,
      ["AGGA"] = 69,
      ["AGGC"] = 13,
      ["AGGG"] = 438,
      ["AGGT"] = 42,
      ["AGTG"] = 23,
      ["ATGG"] = 17,
      ["CGGG"] = 27,
      ["GGGG"] = 27,
      ["GTGG"] = 2,
      ["TGGG"] = 43,
   },
   ["AGGT"] = {
      ["AACT"] = 10,
      ["AAGT"] = 22,
      ["ACCA"] = 19,
      ["ACCC"] = 20,
      ["ACCG"] = 33,
      ["ACGC"] = 1,
      ["AGCC"] = 1,
      ["AGGA"] = 20,
      ["AGGC"] = 79,
      ["AGGG"] = 8,
      ["AGGT"] = 539,
      ["ATCT"] = 20,
      ["ATGT"] = 10,
      ["TGGC"] = 2,
   },
   ["AGTA"] = {
      ["AACT"] = 57,
      ["AATA"] = 6,
      ["ACTA"] = 5,
      ["AGAA"] = 9,
      ["AGCA"] = 19,
      ["AGGA"] = 4,
      ["AGTA"] = 613,
      ["AGTC"] = 47,
      ["AGTG"] = 32,
      ["ATCT"] = 3,
      ["ATTA"] = 2,
      ["TACA"] = 2,
      ["TACC"] = 12,
      ["TACG"] = 4,
   },
   ["AGTC"] = {
      ["AACT"] = 25,
      ["AAGC"] = 1,
      ["AATC"] = 40,
      ["ACCC"] = 1,
      ["ACTC"] = 27,
      ["AGAC"] = 29,
      ["AGCC"] = 62,
      ["AGGC"] = 34,
      ["AGTA"] = 14,
      ["AGTC"] = 606,
      ["AGTG"] = 13,
      ["ATTC"] = 16,
      ["GACC"] = 47,
      ["GACG"] = 35,
      ["GATG"] = 1,
      ["TATC"] = 1,
      ["TGTC"] = 62,
   },
   ["AGTG"] = {
      ["AAAG"] = 2,
      ["AACT"] = 6,
      ["AAGG"] = 2,
      ["AATG"] = 27,
      ["ACTG"] = 20,
      ["AGAG"] = 45,
      ["AGCA"] = 2,
      ["AGCG"] = 52,
      ["AGGG"] = 26,
      ["AGTA"] = 53,
      ["AGTC"] = 10,
      ["AGTG"] = 389,
      ["ATTG"] = 4,
      ["CACG"] = 31,
      ["GGTG"] = 36,
      ["GTTG"] = 1,
      ["TGTG"] = 44,
   },
   ["ATAA"] = {
      ["AAAA"] = 9,
      ["ACAA"] = 5,
      ["ATAA"] = 489,
      ["ATAC"] = 15,
      ["ATAG"] = 2,
      ["ATGA"] = 2,
      ["ATTA"] = 8,
      ["ATTC"] = 1,
   },
   ["ATAC"] = {
      ["AAAC"] = 32,
      ["AATC"] = 1,
      ["ACAC"] = 34,
      ["AGAC"] = 3,
      ["ATAC"] = 505,
      ["ATCC"] = 21,
      ["ATCT"] = 1,
      ["ATGC"] = 10,
      ["ATTC"] = 22,
      ["GTAG"] = 12,
      ["GTTG"] = 1,
      ["TAAC"] = 12,
      ["TCAC"] = 1,
      ["TTAC"] = 34,
   },
   ["ATAG"] = {
      ["AAAG"] = 53,
      ["AAGG"] = 1,
      ["ACAG"] = 27,
      ["AGAG"] = 12,
      ["ATAA"] = 24,
      ["ATAC"] = 5,
      ["ATAG"] = 575,
      ["ATCG"] = 10,
      ["ATGG"] = 12,
      ["ATTG"] = 9,
      ["GCAG"] = 1,
      ["GTAG"] = 11,
      ["TAAG"] = 3,
      ["TTAG"] = 18,
   },
   ["ATCA"] = {
      ["AACA"] = 6,
      ["ACCA"] = 15,
      ["AGCA"] = 13,
      ["ATAA"] = 2,
      ["ATCA"] = 553,
      ["ATCC"] = 23,
      ["ATCG"] = 24,
      ["ATCT"] = 42,
      ["ATGA"] = 1,
      ["ATTA"] = 7,
      ["TACA"] = 1,
      ["TGAC"] = 7,
      ["TGAG"] = 10,
      ["TTCA"] = 26,
   },
   ["ATCC"] = {
      ["AACC"] = 37,
      ["ACCC"] = 37,
      ["AGCC"] = 11,
      ["ATAC"] = 20,
      ["ATCA"] = 11,
      ["ATCC"] = 575,
      ["ATCG"] = 16,
      ["ATCT"] = 19,
      ["ATGC"] = 4,
      ["ATTC"] = 19,
      ["GGAG"] = 14,
      ["GGTG"] = 1,
      ["GTCC"] = 45,
      ["TACC"] = 3,
      ["TTCC"] = 39,
   },
   ["ATCG"] = {
      ["AACG"] = 46,
      ["ACCG"] = 47,
      ["AGCG"] = 36,
      ["AGGG"] = 1,
      ["AGTG"] = 1,
      ["ATAG"] = 13,
      ["ATCA"] = 76,
      ["ATCC"] = 6,
      ["ATCG"] = 572,
      ["ATCT"] = 42,
      ["ATGG"] = 4,
      ["ATTG"] = 14,
      ["CACG"] = 1,
      ["CTCG"] = 45,
      ["GCCG"] = 1,
      ["GTCG"] = 43,
      ["TACG"] = 1,
      ["TCCG"] = 1,
      ["TTCG"] = 49,
   },
   ["ATCT"] = {
      ["AAAT"] = 6,
      ["AACT"] = 27,
      ["AGAA"] = 26,
      ["AGAC"] = 21,
      ["AGAG"] = 10,
      ["AGGT"] = 20,
      ["AGTA"] = 3,
      ["ATCA"] = 54,
      ["ATCC"] = 48,
      ["ATCG"] = 13,
      ["ATCT"] = 648,
      ["ATGT"] = 1,
      ["ATTC"] = 1,
      ["TTCC"] = 1,
   },
   ["ATGA"] = {
      ["AAGA"] = 25,
      ["ACGA"] = 14,
      ["ATAA"] = 9,
      ["ATCA"] = 2,
      ["ATGA"] = 687,
      ["ATGC"] = 50,
      ["ATGG"] = 26,
      ["ATGT"] = 38,
      ["ATTA"] = 2,
      ["TCAC"] = 5,
      ["TCAG"] = 4,
      ["TTGA"] = 7,
   },
   ["ATGC"] = {
      ["AAGC"] = 63,
      ["ACGC"] = 59,
      ["AGCA"] = 2,
      ["AGCG"] = 1,
      ["AGGC"] = 11,
      ["ATAC"] = 21,
      ["ATCC"] = 15,
      ["ATGA"] = 4,
      ["ATGC"] = 548,
      ["ATGT"] = 31,
      ["ATTC"] = 2,
      ["GCAG"] = 14,
      ["GCTG"] = 4,
      ["GTGC"] = 35,
      ["TAGC"] = 8,
      ["TTGC"] = 42,
   },
   ["ATGG"] = {
      ["AAGG"] = 51,
      ["ACGG"] = 56,
      ["AGGG"] = 6,
      ["AGTG"] = 1,
      ["ATAG"] = 23,
      ["ATCG"] = 10,
      ["ATGA"] = 40,
      ["ATGC"] = 7,
      ["ATGG"] = 514,
      ["ATGT"] = 45,
      ["ATTG"] = 3,
      ["CAGG"] = 4,
      ["CGGG"] = 1,
      ["CTGG"] = 19,
      ["GAGG"] = 2,
      ["GCGG"] = 1,
      ["GTGG"] = 28,
      ["TAGG"] = 17,
      ["TCGG"] = 1,
      ["TGGG"] = 1,
      ["TTGG"] = 29,
   },
   ["ATGT"] = {
      ["AAAT"] = 1,
      ["AAGT"] = 39,
      ["ACAA"] = 8,
      ["ACAC"] = 12,
      ["ACAG"] = 11,
      ["ACTA"] = 2,
      ["AGGT"] = 2,
      ["AGTA"] = 1,
      ["ATCT"] = 19,
      ["ATGA"] = 12,
      ["ATGC"] = 75,
      ["ATGG"] = 3,
      ["ATGT"] = 614,
   },
   ["ATTA"] = {
      ["AAAT"] = 44,
      ["AATA"] = 12,
      ["ACTA"] = 2,
      ["ATAA"] = 11,
      ["ATCA"] = 9,
      ["ATGA"] = 3,
      ["ATTA"] = 555,
      ["ATTC"] = 20,
      ["ATTG"] = 15,
      ["TAAA"] = 4,
      ["TAAC"] = 6,
   },
   ["ATTC"] = {
      ["AAAT"] = 15,
      ["AAGC"] = 1,
      ["AATC"] = 58,
      ["ACTC"] = 18,
      ["AGTC"] = 5,
      ["ATAC"] = 29,
      ["ATAG"] = 1,
      ["ATCA"] = 1,
      ["ATCC"] = 38,
      ["ATGC"] = 6,
      ["ATTA"] = 5,
      ["ATTC"] = 673,
      ["ATTG"] = 7,
      ["GAAC"] = 18,
      ["GAAG"] = 9,
      ["TTTC"] = 35,
   },
   ["ATTG"] = {
      ["AAAG"] = 4,
      ["AAAT"] = 13,
      ["AAGG"] = 1,
      ["AATG"] = 54,
      ["ACCG"] = 1,
      ["ACTG"] = 19,
      ["AGCG"] = 1,
      ["AGTG"] = 9,
      ["ATAG"] = 40,
      ["ATCA"] = 1,
      ["ATCG"] = 59,
      ["ATGG"] = 9,
      ["ATTA"] = 36,
      ["ATTC"] = 6,
      ["ATTG"] = 522,
      ["CAAG"] = 20,
      ["GTTG"] = 13,
      ["TATG"] = 1,
      ["TTTG"] = 26,
   },
   ["CAAG"] = {
      ["AAAG"] = 14,
      ["ATTG"] = 20,
      ["CAAG"] = 519,
      ["CACG"] = 8,
      ["CAGG"] = 7,
      ["CGGG"] = 1,
      ["CTCG"] = 31,
      ["CTGG"] = 8,
      ["GAAG"] = 11,
      ["GTTG"] = 15,
      ["TAAG"] = 37,
      ["TCTG"] = 1,
      ["TTTG"] = 54,
   },
   ["CACG"] = {
      ["AACG"] = 29,
      ["AGTG"] = 31,
      ["CAAG"] = 11,
      ["CACG"] = 351,
      ["CAGG"] = 8,
      ["CGGG"] = 50,
      ["CTCG"] = 76,
      ["CTGG"] = 2,
      ["GACG"] = 23,
      ["GGTG"] = 8,
      ["TAAG"] = 1,
      ["TACG"] = 51,
      ["TGAG"] = 1,
      ["TGCG"] = 1,
      ["TGGG"] = 1,
      ["TGTG"] = 67,
   },
   ["CAGG"] = {
      ["AAGG"] = 24,
      ["ACTG"] = 68,
      ["CAAG"] = 13,
      ["CACG"] = 20,
      ["CAGG"] = 552,
      ["CGGG"] = 42,
      ["CTCG"] = 1,
      ["CTGG"] = 66,
      ["GAGG"] = 17,
      ["GCTG"] = 10,
      ["TAGG"] = 37,
      ["TCTG"] = 99,
   },
   ["CGGG"] = {
      ["ACCG"] = 61,
      ["AGGG"] = 25,
      ["CACG"] = 50,
      ["CAGG"] = 47,
      ["CGGG"] = 386,
      ["CTCG"] = 20,
      ["CTGG"] = 30,
      ["GCCG"] = 14,
      ["GGGG"] = 25,
      ["GTGG"] = 1,
      ["TCAG"] = 1,
      ["TCCG"] = 82,
      ["TCTG"] = 1,
      ["TGGG"] = 43,
      ["TTCG"] = 1,
   },
   ["CTCG"] = {
      ["ACAG"] = 1,
      ["AGAG"] = 13,
      ["ATCG"] = 4,
      ["CAAG"] = 32,
      ["CACG"] = 45,
      ["CAGG"] = 3,
      ["CGGG"] = 20,
      ["CTCG"] = 460,
      ["CTGG"] = 4,
      ["GGAG"] = 4,
      ["TGAG"] = 29,
      ["TTCG"] = 16,
   },
   ["CTGG"] = {
      ["ACAG"] = 82,
      ["ATGG"] = 6,
      ["CAAG"] = 10,
      ["CACG"] = 2,
      ["CAGG"] = 68,
      ["CGGG"] = 21,
      ["CTCG"] = 8,
      ["CTGG"] = 534,
      ["GCAG"] = 10,
      ["GTGG"] = 1,
      ["TCAG"] = 62,
      ["TCGG"] = 1,
      ["TTGG"] = 9,
   },
   ["GAAC"] = {
      ["AAAC"] = 61,
      ["AGAC"] = 1,
      ["ATTC"] = 18,
      ["GAAC"] = 551,
      ["GAAG"] = 1,
      ["GACC"] = 14,
      ["GAGC"] = 35,
      ["GTCC"] = 35,
      ["GTGC"] = 14,
      ["GTTG"] = 6,
      ["TAAC"] = 50,
      ["TTTC"] = 3,
   },
   ["GAAG"] = {
      ["AAAG"] = 56,
      ["AGAG"] = 1,
      ["ATTC"] = 9,
      ["CAAG"] = 6,
      ["GAAC"] = 10,
      ["GAAG"] = 562,
      ["GACG"] = 12,
      ["GAGG"] = 8,
      ["GATG"] = 7,
      ["GCAG"] = 10,
      ["GGAG"] = 35,
      ["GGCG"] = 1,
      ["GTAG"] = 36,
      ["GTCG"] = 4,
      ["TAAG"] = 35,
      ["TCTC"] = 1,
      ["TTTC"] = 50,
   },
   ["GACC"] = {
      ["AACC"] = 66,
      ["AGCC"] = 1,
      ["AGTC"] = 47,
      ["GAAC"] = 26,
      ["GACC"] = 499,
      ["GACG"] = 16,
      ["GAGC"] = 4,
      ["GGGC"] = 33,
      ["GGTG"] = 11,
      ["GTCC"] = 44,
      ["GTGC"] = 1,
      ["TACC"] = 39,
      ["TGTC"] = 32,
   },
   ["GACG"] = {
      ["AACA"] = 1,
      ["AACG"] = 61,
      ["ACCG"] = 3,
      ["AGCG"] = 2,
      ["AGTC"] = 35,
      ["CACG"] = 11,
      ["GAAG"] = 20,
      ["GACC"] = 10,
      ["GACG"] = 452,
      ["GAGG"] = 4,
      ["GATG"] = 32,
      ["GCCG"] = 72,
      ["GGCG"] = 76,
      ["GTCG"] = 37,
      ["TACG"] = 51,
      ["TATG"] = 1,
      ["TCCG"] = 1,
      ["TGCC"] = 1,
      ["TGGC"] = 3,
      ["TGTC"] = 72,
   },
   ["GAGC"] = {
      ["AAGC"] = 75,
      ["ACTC"] = 50,
      ["AGGC"] = 2,
      ["ATGC"] = 1,
      ["GAAC"] = 33,
      ["GACC"] = 28,
      ["GAGC"] = 477,
      ["GAGG"] = 2,
      ["GCTG"] = 12,
      ["GGGC"] = 40,
      ["GTGC"] = 42,
      ["TAGC"] = 43,
      ["TCTC"] = 6,
      ["TTGC"] = 1,
   },
   ["GAGG"] = {
      ["AAGG"] = 59,
      ["ACTC"] = 37,
      ["CAGG"] = 7,
      ["GAAG"] = 12,
      ["GACG"] = 19,
      ["GAGC"] = 8,
      ["GAGG"] = 420,
      ["GATG"] = 4,
      ["GCGG"] = 9,
      ["GGCG"] = 1,
      ["GGGG"] = 17,
      ["GGTG"] = 1,
      ["GTGG"] = 29,
      ["TAGG"] = 31,
      ["TCAC"] = 1,
      ["TCTC"] = 60,
   },
   ["GATG"] = {
      ["AAAG"] = 1,
      ["AATC"] = 12,
      ["AATG"] = 58,
      ["AGTC"] = 1,
      ["GAAG"] = 56,
      ["GACG"] = 52,
      ["GAGG"] = 16,
      ["GATG"] = 395,
      ["GCTG"] = 27,
      ["GGCG"] = 1,
      ["GGTG"] = 36,
      ["GTAG"] = 1,
      ["GTTG"] = 20,
      ["TATC"] = 41,
      ["TATG"] = 29,
      ["TGTC"] = 3,
      ["TTTC"] = 1,
   },
   ["GCAG"] = {
      ["ACAG"] = 85,
      ["ATGC"] = 14,
      ["CTGG"] = 10,
      ["GAAG"] = 16,
      ["GATG"] = 3,
      ["GCAG"] = 455,
      ["GCCG"] = 38,
      ["GCGG"] = 35,
      ["GCTG"] = 42,
      ["GGAG"] = 15,
      ["GTAG"] = 21,
      ["GTGC"] = 11,
      ["TCAG"] = 71,
      ["TCGC"] = 2,
      ["TTGC"] = 53,
   },
   ["GCCG"] = {
      ["AACG"] = 1,
      ["ACAG"] = 1,
      ["ACCA"] = 2,
      ["ACCG"] = 91,
      ["AGGC"] = 47,
      ["CGGG"] = 14,
      ["GACG"] = 59,
      ["GCAG"] = 31,
      ["GCCG"] = 447,
      ["GCGG"] = 18,
      ["GCTG"] = 49,
      ["GGCG"] = 7,
      ["GGGC"] = 10,
      ["GTAG"] = 1,
      ["GTCG"] = 63,
      ["TCCG"] = 75,
      ["TCTC"] = 1,
      ["TGGC"] = 80,
   },
   ["GCGG"] = {
      ["AAGG"] = 1,
      ["ACAG"] = 1,
      ["ACGA"] = 2,
      ["ACGC"] = 40,
      ["ACGG"] = 76,
      ["CAGG"] = 1,
      ["GAGG"] = 37,
      ["GCAG"] = 51,
      ["GCCG"] = 18,
      ["GCGG"] = 265,
      ["GCTG"] = 16,
      ["GGGC"] = 1,
      ["GGGG"] = 12,
      ["GTGG"] = 45,
      ["TCAC"] = 1,
      ["TCGC"] = 72,
      ["TCGG"] = 39,
   },
   ["GCTG"] = {
      ["AAGC"] = 35,
      ["ACAG"] = 2,
      ["ACGC"] = 2,
      ["ACGG"] = 1,
      ["ACTA"] = 2,
      ["ACTG"] = 60,
      ["AGGC"] = 1,
      ["ATGC"] = 4,
      ["CAGG"] = 10,
      ["GACG"] = 2,
      ["GAGC"] = 12,
      ["GAGG"] = 1,
      ["GATG"] = 26,
      ["GCAG"] = 73,
      ["GCCG"] = 61,
      ["GCGG"] = 42,
      ["GCTG"] = 352,
      ["GGAG"] = 1,
      ["GGGC"] = 1,
      ["GGTG"] = 6,
      ["GTTG"] = 19,
      ["TAGC"] = 47,
      ["TCTG"] = 33,
      ["TGGC"] = 1,
      ["TTGC"] = 7,
   },
   ["GGAG"] = {
      ["AAAG"] = 2,
      ["AGAA"] = 3,
      ["AGAG"] = 63,
      ["ATCC"] = 14,
      ["CTCG"] = 4,
      ["GAAG"] = 49,
      ["GACG"] = 1,
      ["GCAG"] = 20,
      ["GGAG"] = 396,
      ["GGCG"] = 33,
      ["GGGG"] = 13,
      ["GGTG"] = 20,
      ["GTAG"] = 19,
      ["GTCC"] = 12,
      ["TCTC"] = 1,
      ["TGAG"] = 43,
      ["TTCC"] = 39,
   },
   ["GGCG"] = {
      ["ACCG"] = 1,
      ["AGCA"] = 2,
      ["AGCC"] = 30,
      ["AGCG"] = 61,
      ["ATCG"] = 1,
      ["GACG"] = 49,
      ["GAGG"] = 1,
      ["GATG"] = 1,
      ["GCCG"] = 30,
      ["GGAG"] = 34,
      ["GGCG"] = 280,
      ["GGGG"] = 6,
      ["GGTG"] = 43,
      ["GTCG"] = 50,
      ["TGCC"] = 89,
      ["TGCG"] = 51,
   },
   ["GGGC"] = {
      ["ACCC"] = 58,
      ["AGGC"] = 82,
      ["GACC"] = 33,
      ["GAGC"] = 45,
      ["GCCG"] = 10,
      ["GCTG"] = 1,
      ["GGCG"] = 1,
      ["GGGC"] = 423,
      ["GGGG"] = 5,
      ["GTCC"] = 37,
      ["GTGC"] = 45,
      ["TCCC"] = 16,
      ["TGGC"] = 52,
   },
   ["GGGG"] = {
      ["ACCC"] = 44,
      ["AGCG"] = 1,
      ["AGGA"] = 1,
      ["AGGG"] = 63,
      ["ATGG"] = 1,
      ["CGGG"] = 6,
      ["GAGG"] = 31,
      ["GCGG"] = 6,
      ["GGAG"] = 23,
      ["GGCG"] = 11,
      ["GGGC"] = 10,
      ["GGGG"] = 273,
      ["GGTG"] = 17,
      ["GTGG"] = 13,
      ["TCCC"] = 51,
      ["TGGG"] = 21,
      ["TTCC"] = 1,
   },
   ["GGTG"] = {
      ["AACC"] = 16,
      ["AATG"] = 2,
      ["AGAG"] = 2,
      ["AGCC"] = 1,
      ["AGCG"] = 1,
      ["AGTG"] = 47,
      ["ATCC"] = 1,
      ["CACG"] = 8,
      ["GAAG"] = 2,
      ["GACC"] = 11,
      ["GACG"] = 1,
      ["GATG"] = 43,
      ["GCTG"] = 16,
      ["GGAG"] = 67,
      ["GGCG"] = 48,
      ["GGGG"] = 20,
      ["GGTG"] = 252,
      ["GTTG"] = 20,
      ["TACC"] = 43,
      ["TATG"] = 1,
      ["TCTG"] = 1,
      ["TGAG"] = 1,
      ["TGCC"] = 3,
      ["TGTG"] = 19,
      ["TTCC"] = 2,
   },
   ["GTAG"] = {
      ["ATAC"] = 12,
      ["ATAG"] = 42,
      ["GAAG"] = 36,
      ["GACG"] = 2,
      ["GATG"] = 1,
      ["GCAG"] = 51,
      ["GGAG"] = 23,
      ["GGCG"] = 3,
      ["GTAG"] = 393,
      ["GTCG"] = 32,
      ["GTGG"] = 17,
      ["GTTG"] = 30,
      ["TAAC"] = 1,
      ["TCAC"] = 1,
      ["TGAC"] = 1,
      ["TTAC"] = 26,
      ["TTAG"] = 8,
   },
   ["GTCC"] = {
      ["AACC"] = 1,
      ["AGAC"] = 42,
      ["AGCC"] = 3,
      ["ATCC"] = 64,
      ["ATCG"] = 1,
      ["GAAC"] = 35,
      ["GACC"] = 60,
      ["GGAG"] = 12,
      ["GGGC"] = 37,
      ["GTCC"] = 536,
      ["GTCG"] = 39,
      ["GTGC"] = 11,
      ["TGAC"] = 37,
      ["TTCC"] = 15,
   },
   ["GTCG"] = {
      ["AGAC"] = 51,
      ["AGCG"] = 9,
      ["ATCA"] = 1,
      ["ATCG"] = 75,
      ["CTCG"] = 11,
      ["GAAG"] = 1,
      ["GACG"] = 52,
      ["GATG"] = 1,
      ["GCCG"] = 53,
      ["GGCG"] = 53,
      ["GGTG"] = 1,
      ["GTAG"] = 26,
      ["GTCC"] = 21,
      ["GTCG"] = 382,
      ["GTGG"] = 7,
      ["GTTG"] = 38,
      ["TGAC"] = 65,
      ["TGGC"] = 1,
      ["TTCG"] = 36,
   },
   ["GTGC"] = {
      ["AAGC"] = 1,
      ["ACAC"] = 70,
      ["AGGC"] = 2,
      ["ATAC"] = 1,
      ["ATGC"] = 78,
      ["ATGT"] = 2,
      ["GAAC"] = 14,
      ["GACC"] = 1,
      ["GAGC"] = 49,
      ["GCAG"] = 11,
      ["GGGC"] = 33,
      ["GTCC"] = 25,
      ["GTGC"] = 388,
      ["GTGG"] = 5,
      ["TCAC"] = 14,
      ["TTGC"] = 20,
   },
   ["GTGG"] = {
      ["AAGG"] = 1,
      ["ACAC"] = 58,
      ["ACCC"] = 2,
      ["ACGG"] = 1,
      ["ATAG"] = 1,
      ["ATGA"] = 1,
      ["ATGG"] = 51,
      ["CTGG"] = 12,
      ["GAAG"] = 1,
      ["GAGG"] = 56,
      ["GCAG"] = 1,
      ["GCGG"] = 41,
      ["GGGG"] = 29,
      ["GGTG"] = 2,
      ["GTAG"] = 29,
      ["GTCG"] = 14,
      ["GTGC"] = 13,
      ["GTGG"] = 283,
      ["GTTG"] = 8,
      ["TCAC"] = 31,
      ["TCTC"] = 1,
      ["TTGG"] = 9,
   },
   ["GTTG"] = {
      ["AAAC"] = 31,
      ["ACAC"] = 1,
      ["ATAC"] = 1,
      ["ATGG"] = 1,
      ["ATTG"] = 44,
      ["CAAG"] = 15,
      ["GAAC"] = 6,
      ["GAAG"] = 1,
      ["GATG"] = 44,
      ["GCAG"] = 1,
      ["GCTG"] = 28,
      ["GGCG"] = 2,
      ["GGTG"] = 27,
      ["GTAG"] = 47,
      ["GTCG"] = 61,
      ["GTGG"] = 7,
      ["GTTG"] = 272,
      ["TAAC"] = 31,
      ["TGAC"] = 4,
      ["TTAC"] = 1,
      ["TTTG"] = 7,
   },
   ["TAAA"] = {
      ["AAAA"] = 3,
      ["ATTA"] = 4,
      ["TAAA"] = 361,
      ["TAAC"] = 3,
      ["TTCA"] = 1,
      ["TTTG"] = 10,
   },
   ["TAAC"] = {
      ["AAAC"] = 43,
      ["ATAC"] = 1,
      ["ATTA"] = 6,
      ["GAAC"] = 29,
      ["GTAG"] = 1,
      ["GTTG"] = 31,
      ["TAAC"] = 475,
      ["TACC"] = 4,
      ["TAGC"] = 4,
      ["TATC"] = 2,
      ["TCAC"] = 3,
      ["TGAC"] = 9,
      ["TTAC"] = 18,
   },
   ["TAAG"] = {
      ["AAAG"] = 35,
      ["ACAG"] = 1,
      ["CAAG"] = 20,
      ["GAAG"] = 23,
      ["TAAA"] = 11,
      ["TAAC"] = 2,
      ["TAAG"] = 542,
      ["TACG"] = 2,
      ["TAGG"] = 1,
      ["TATG"] = 3,
      ["TCAG"] = 2,
      ["TGAG"] = 6,
      ["TTAG"] = 8,
   },
   ["TACA"] = {
      ["AACA"] = 18,
      ["AGTA"] = 2,
      ["TACA"] = 527,
      ["TACC"] = 8,
      ["TACG"] = 11,
      ["TGTC"] = 2,
      ["TGTG"] = 31,
      ["TTCA"] = 3,
   },
   ["TACC"] = {
      ["AACC"] = 37,
      ["ACCG"] = 1,
      ["AGTA"] = 12,
      ["ATCC"] = 1,
      ["GACC"] = 23,
      ["GGTG"] = 43,
      ["TAAC"] = 10,
      ["TACA"] = 4,
      ["TACC"] = 474,
      ["TATC"] = 4,
      ["TCCC"] = 11,
      ["TGCC"] = 18,
      ["TTCC"] = 12,
   },
   ["TACG"] = {
      ["AACA"] = 1,
      ["AACG"] = 52,
      ["AGTA"] = 4,
      ["CACG"] = 52,
      ["GACG"] = 42,
      ["TAAG"] = 7,
      ["TACA"] = 32,
      ["TACG"] = 507,
      ["TATG"] = 8,
      ["TCCG"] = 23,
      ["TGCG"] = 33,
      ["TTCG"] = 22,
   },
   ["TAGA"] = {
      ["AAGA"] = 12,
      ["ACTA"] = 3,
      ["TAAA"] = 2,
      ["TAGA"] = 608,
      ["TAGG"] = 1,
      ["TCTC"] = 8,
      ["TCTG"] = 25,
      ["TGGA"] = 1,
      ["TTGA"] = 1,
   },
   ["TAGC"] = {
      ["AAGC"] = 49,
      ["ACTA"] = 5,
      ["GAGC"] = 46,
      ["GCTG"] = 47,
      ["TAAC"] = 6,
      ["TACC"] = 4,
      ["TAGA"] = 1,
      ["TAGC"] = 524,
      ["TATC"] = 8,
      ["TCGC"] = 10,
      ["TGGC"] = 6,
      ["TTGC"] = 12,
   },
   ["TAGG"] = {
      ["AAGG"] = 58,
      ["ACTA"] = 4,
      ["CAGG"] = 53,
      ["CTGG"] = 3,
      ["GAGG"] = 47,
      ["TAAG"] = 4,
      ["TACG"] = 11,
      ["TAGA"] = 28,
      ["TAGC"] = 1,
      ["TAGG"] = 528,
      ["TATG"] = 3,
      ["TCGG"] = 6,
      ["TGGG"] = 8,
      ["TTGG"] = 13,
   },
   ["TATC"] = {
      ["AATA"] = 4,
      ["AATC"] = 44,
      ["GATG"] = 41,
      ["TAAC"] = 7,
      ["TACC"] = 16,
      ["TAGC"] = 1,
      ["TATC"] = 535,
      ["TATG"] = 1,
      ["TGCC"] = 1,
      ["TGTC"] = 6,
      ["TTTC"] = 6,
   },
   ["TATG"] = {
      ["AATG"] = 44,
      ["GATG"] = 22,
      ["GTAG"] = 2,
      ["TAAG"] = 34,
      ["TACG"] = 21,
      ["TAGG"] = 3,
      ["TATC"] = 4,
      ["TATG"] = 484,
      ["TCTG"] = 3,
      ["TGTG"] = 5,
      ["TTTG"] = 6,
   },
   ["TCAC"] = {
      ["ACAC"] = 28,
      ["ATAC"] = 1,
      ["ATGA"] = 5,
      ["GAGG"] = 1,
      ["GCGG"] = 1,
      ["GTAG"] = 1,
      ["GTGC"] = 14,
      ["GTGG"] = 31,
      ["TAAC"] = 2,
      ["TCAC"] = 442,
      ["TCCC"] = 6,
      ["TCGC"] = 16,
      ["TCTC"] = 6,
      ["TGAC"] = 4,
      ["TGGA"] = 1,
      ["TTAC"] = 14,
      ["TTGA"] = 1,
   },
   ["TCAG"] = {
      ["ACAG"] = 22,
      ["AGAG"] = 1,
      ["ATAG"] = 2,
      ["ATGA"] = 4,
      ["CAAG"] = 1,
      ["CACG"] = 1,
      ["CGGG"] = 1,
      ["CTGG"] = 62,
      ["GCAG"] = 9,
      ["TAAG"] = 5,
      ["TCAC"] = 2,
      ["TCAG"] = 712,
      ["TCCG"] = 4,
      ["TCGG"] = 6,
      ["TCTG"] = 8,
      ["TGAG"] = 7,
      ["TTAG"] = 8,
      ["TTGA"] = 37,
   },
   ["TCCC"] = {
      ["ACCC"] = 38,
      ["AGGA"] = 10,
      ["GGGC"] = 16,
      ["GGGG"] = 51,
      ["TACC"] = 6,
      ["TCAC"] = 18,
      ["TCCC"] = 550,
      ["TCCG"] = 2,
      ["TCGC"] = 3,
      ["TCTC"] = 5,
      ["TGCC"] = 2,
      ["TGGA"] = 12,
      ["TTCC"] = 12,
      ["TTGC"] = 1,
   },
   ["TCCG"] = {
      ["ACCG"] = 37,
      ["AGGA"] = 12,
      ["CACG"] = 1,
      ["CGGG"] = 82,
      ["GCCG"] = 17,
      ["TACG"] = 8,
      ["TCAG"] = 11,
      ["TCCC"] = 2,
      ["TCCG"] = 747,
      ["TCGG"] = 3,
      ["TCTG"] = 16,
      ["TGCG"] = 3,
      ["TGGA"] = 78,
      ["TGGG"] = 1,
      ["TTCG"] = 12,
   },
   ["TCGC"] = {
      ["ACGA"] = 50,
      ["ACGC"] = 42,
      ["ACGG"] = 4,
      ["GCAG"] = 2,
      ["GCGG"] = 72,
      ["TAGC"] = 10,
      ["TCAC"] = 18,
      ["TCCC"] = 10,
      ["TCGC"] = 563,
      ["TCGG"] = 1,
      ["TCTC"] = 2,
      ["TGGC"] = 9,
      ["TTGC"] = 22,
   },
   ["TCGG"] = {
      ["ACGA"] = 20,
      ["ACGG"] = 48,
      ["CGGG"] = 2,
      ["CTGG"] = 3,
      ["GCGG"] = 16,
      ["TAGG"] = 11,
      ["TCAG"] = 24,
      ["TCCG"] = 9,
      ["TCGC"] = 4,
      ["TCGG"] = 637,
      ["TCTG"] = 5,
      ["TGGG"] = 7,
      ["TTGG"] = 9,
   },
   ["TCTC"] = {
      ["AAGA"] = 4,
      ["AAGT"] = 1,
      ["ACTC"] = 27,
      ["ATTC"] = 1,
      ["GAAG"] = 1,
      ["GAGC"] = 6,
      ["GAGG"] = 60,
      ["GCCG"] = 1,
      ["GGAG"] = 1,
      ["GTGG"] = 1,
      ["TAGA"] = 8,
      ["TATC"] = 2,
      ["TCAC"] = 22,
      ["TCCC"] = 15,
      ["TCGC"] = 12,
      ["TCTC"] = 651,
      ["TCTG"] = 2,
      ["TTTC"] = 3,
   },
   ["TCTG"] = {
      ["AAGA"] = 1,
      ["ACAG"] = 1,
      ["ACTG"] = 35,
      ["CAAG"] = 1,
      ["CAGG"] = 99,
      ["CGGG"] = 1,
      ["GCTG"] = 10,
      ["TAGA"] = 25,
      ["TATG"] = 2,
      ["TCAG"] = 49,
      ["TCCG"] = 42,
      ["TCGG"] = 13,
      ["TCTC"] = 1,
      ["TCTG"] = 646,
      ["TGGA"] = 1,
      ["TGTG"] = 1,
      ["TTTG"] = 2,
   },
   ["TGAC"] = {
      ["AGAC"] = 89,
      ["AGCC"] = 1,
      ["ATCA"] = 7,
      ["GTAG"] = 1,
      ["GTCC"] = 37,
      ["GTCG"] = 65,
      ["GTTG"] = 4,
      ["TAAC"] = 33,
      ["TCAC"] = 6,
      ["TGAC"] = 564,
      ["TGCC"] = 47,
      ["TGGC"] = 25,
      ["TGTC"] = 29,
      ["TTAC"] = 3,
      ["TTCA"] = 1,
   },
   ["TGAG"] = {
      ["AGAA"] = 2,
      ["AGAG"] = 63,
      ["ATCA"] = 10,
      ["CAAG"] = 1,
      ["CACG"] = 1,
      ["CTCG"] = 29,
      ["GGAG"] = 20,
      ["GTAG"] = 1,
      ["TAAG"] = 17,
      ["TCAG"] = 4,
      ["TGAC"] = 5,
      ["TGAG"] = 475,
      ["TGCG"] = 16,
      ["TGGG"] = 4,
      ["TGTG"] = 14,
      ["TTAG"] = 5,
      ["TTCA"] = 22,
   },
   ["TGCC"] = {
      ["AACC"] = 1,
      ["AGCA"] = 25,
      ["AGCC"] = 77,
      ["AGTC"] = 1,
      ["ATCC"] = 1,
      ["GACC"] = 3,
      ["GACG"] = 1,
      ["GGCG"] = 89,
      ["GGTG"] = 3,
      ["TAAC"] = 1,
      ["TACC"] = 31,
      ["TATC"] = 1,
      ["TCCC"] = 10,
      ["TGAC"] = 23,
      ["TGCC"] = 509,
      ["TGCG"] = 13,
      ["TGGC"] = 4,
      ["TGTC"] = 49,
      ["TTAC"] = 1,
      ["TTCC"] = 9,
   },
   ["TGCG"] = {
      ["AGAG"] = 1,
      ["AGCA"] = 16,
      ["AGCG"] = 78,
      ["CACG"] = 4,
      ["GGCG"] = 24,
      ["TACG"] = 23,
      ["TAGG"] = 1,
      ["TATG"] = 2,
      ["TCCG"] = 12,
      ["TGAG"] = 14,
      ["TGCC"] = 7,
      ["TGCG"] = 356,
      ["TGGG"] = 4,
      ["TGTG"] = 29,
      ["TTCG"] = 6,
   },
   ["TGGA"] = {
      ["ACCA"] = 21,
      ["AGGA"] = 47,
      ["AGGT"] = 1,
      ["TAGA"] = 3,
      ["TCAC"] = 1,
      ["TCCC"] = 12,
      ["TCCG"] = 78,
      ["TCTG"] = 1,
      ["TGGA"] = 645,
      ["TGGC"] = 27,
      ["TGGG"] = 12,
      ["TTCA"] = 6,
   },
   ["TGGC"] = {
      ["AAGC"] = 2,
      ["ACCA"] = 27,
      ["AGAC"] = 3,
      ["AGGC"] = 95,
      ["AGTC"] = 2,
      ["GACG"] = 3,
      ["GCCG"] = 80,
      ["GCTG"] = 1,
      ["GGGC"] = 29,
      ["GTCG"] = 1,
      ["TAGC"] = 35,
      ["TCGC"] = 12,
      ["TGAC"] = 28,
      ["TGCC"] = 28,
      ["TGGA"] = 2,
      ["TGGC"] = 494,
      ["TGTC"] = 51,
      ["TTGC"] = 11,
   },
   ["TGGG"] = {
      ["ACCA"] = 17,
      ["AGGG"] = 66,
      ["AGTG"] = 1,
      ["CACG"] = 1,
      ["CAGG"] = 2,
      ["CGGG"] = 62,
      ["GAGG"] = 2,
      ["GGGG"] = 24,
      ["TAGG"] = 23,
      ["TCCG"] = 1,
      ["TCGG"] = 3,
      ["TGAG"] = 13,
      ["TGCG"] = 12,
      ["TGGA"] = 34,
      ["TGGC"] = 8,
      ["TGGG"] = 383,
      ["TGTG"] = 18,
      ["TTGG"] = 5,
   },
   ["TGTC"] = {
      ["AACA"] = 3,
      ["AGTC"] = 67,
      ["GACC"] = 32,
      ["GACG"] = 72,
      ["GATG"] = 3,
      ["TACA"] = 2,
      ["TAGC"] = 2,
      ["TATC"] = 21,
      ["TCTC"] = 7,
      ["TGAC"] = 26,
      ["TGCC"] = 40,
      ["TGGC"] = 28,
      ["TGTC"] = 587,
      ["TGTG"] = 3,
   },
   ["TGTG"] = {
      ["AACA"] = 3,
      ["AGTG"] = 57,
      ["CACG"] = 67,
      ["GATG"] = 1,
      ["GGTG"] = 19,
      ["TACA"] = 31,
      ["TATG"] = 22,
      ["TCTG"] = 2,
      ["TGAG"] = 27,
      ["TGCG"] = 32,
      ["TGGG"] = 18,
      ["TGTC"] = 9,
      ["TGTG"] = 328,
      ["TTTG"] = 1,
   },
   ["TTAC"] = {
      ["AAAC"] = 3,
      ["ATAC"] = 18,
      ["GTAG"] = 26,
      ["GTTG"] = 1,
      ["TAAC"] = 13,
      ["TCAC"] = 25,
      ["TGAC"] = 1,
      ["TTAC"] = 607,
      ["TTCC"] = 5,
      ["TTGC"] = 4,
      ["TTTC"] = 10,
   },
   ["TTAG"] = {
      ["ATAG"] = 23,
      ["GTAG"] = 2,
      ["TAAG"] = 24,
      ["TCAG"] = 19,
      ["TGAG"] = 3,
      ["TTAC"] = 1,
      ["TTAG"] = 507,
      ["TTCG"] = 3,
      ["TTGA"] = 1,
      ["TTGG"] = 2,
      ["TTTG"] = 2,
   },
   ["TTCA"] = {
      ["AGAA"] = 8,
      ["ATCA"] = 6,
      ["TAAA"] = 1,
      ["TACA"] = 5,
      ["TGAC"] = 1,
      ["TGAG"] = 22,
      ["TGGA"] = 6,
      ["TTCA"] = 540,
      ["TTCC"] = 4,
      ["TTCG"] = 5,
      ["TTGA"] = 1,
   },
   ["TTCC"] = {
      ["AGAA"] = 7,
      ["ATCC"] = 36,
      ["GGAG"] = 39,
      ["GGGG"] = 1,
      ["GGTG"] = 2,
      ["GTCC"] = 10,
      ["TACC"] = 20,
      ["TCAC"] = 1,
      ["TCCC"] = 24,
      ["TGCC"] = 9,
      ["TTAC"] = 5,
      ["TTCA"] = 2,
      ["TTCC"] = 718,
      ["TTCG"] = 1,
      ["TTTC"] = 4,
   },
   ["TTCG"] = {
      ["AGAA"] = 6,
      ["ATCG"] = 43,
      ["CACG"] = 1,
      ["CGGG"] = 1,
      ["CTCG"] = 87,
      ["GTCG"] = 10,
      ["TACG"] = 30,
      ["TATG"] = 1,
      ["TCCG"] = 50,
      ["TGCG"] = 8,
      ["TTAG"] = 4,
      ["TTCA"] = 28,
      ["TTCC"] = 2,
      ["TTCG"] = 655,
      ["TTTG"] = 4,
   },
   ["TTGA"] = {
      ["ACAA"] = 10,
      ["ATGA"] = 13,
      ["TAGA"] = 12,
      ["TCAC"] = 1,
      ["TCAG"] = 37,
      ["TTAG"] = 1,
      ["TTGA"] = 490,
      ["TTGC"] = 16,
      ["TTGG"] = 4,
   },
   ["TTGC"] = {
      ["AAGC"] = 1,
      ["ACAA"] = 2,
      ["ACGA"] = 1,
      ["ATGC"] = 24,
      ["GCAG"] = 53,
      ["GCTG"] = 7,
      ["GTGC"] = 4,
      ["TAGC"] = 23,
      ["TCAC"] = 1,
      ["TCGC"] = 42,
      ["TGGC"] = 5,
      ["TTAC"] = 3,
      ["TTCC"] = 13,
      ["TTGA"] = 2,
      ["TTGC"] = 530,
      ["TTGG"] = 1,
      ["TTTC"] = 2,
   },
   ["TTGG"] = {
      ["AAGG"] = 1,
      ["ACAA"] = 7,
      ["ATGA"] = 1,
      ["ATGG"] = 36,
      ["CAGG"] = 5,
      ["CTGG"] = 69,
      ["GAGG"] = 1,
      ["GTGG"] = 4,
      ["TAGG"] = 30,
      ["TCGG"] = 52,
      ["TGGG"] = 1,
      ["TTAG"] = 7,
      ["TTCG"] = 5,
      ["TTGA"] = 25,
      ["TTGC"] = 2,
      ["TTGG"] = 396,
   },
   ["TTTC"] = {
      ["AAAA"] = 3,
      ["AATC"] = 1,
      ["ATTC"] = 23,
      ["GAAC"] = 3,
      ["GAAG"] = 50,
      ["GATG"] = 1,
      ["TATC"] = 17,
      ["TCTC"] = 12,
      ["TGTC"] = 2,
      ["TTAC"] = 14,
      ["TTCC"] = 20,
      ["TTGC"] = 2,
      ["TTTC"] = 715,
   },
   ["TTTG"] = {
      ["AAAA"] = 1,
      ["AATG"] = 3,
      ["AGTG"] = 1,
      ["ATTG"] = 14,
      ["CAAG"] = 54,
      ["GATG"] = 1,
      ["GTTG"] = 4,
      ["TAAA"] = 10,
      ["TAAG"] = 1,
      ["TATG"] = 17,
      ["TCTG"] = 16,
      ["TGTG"] = 6,
      ["TTAG"] = 24,
      ["TTCG"] = 23,
      ["TTGG"] = 3,
      ["TTTC"] = 1,
      ["TTTG"] = 485,
   },
}
















local fragment = { Assembly = {} }






















local function is_palindromic(sequence)
   return sequence == transform.reverse_complement(sequence)
end



function fragment.set_efficiency(overhangs)
   local efficiency = 1.0
   for _, overhang in ipairs(overhangs) do

      if not fragment_frequencies[overhang] then
         fragment_frequencies[overhang] = {}
      end

      local n_correct = (fragment_frequencies[overhang][overhang] or 0)
      local n_total = 0
      for _, overhang2 in ipairs(overhangs) do
         n_total = n_total + (fragment_frequencies[overhang][overhang2] or 0)
         n_total = n_total + (fragment_frequencies[overhang][transform.reverse_complement(overhang2)] or 0)
      end
      if n_total ~= n_correct and n_total > 0 then
         efficiency = efficiency * (n_correct / n_total)
      end
   end
   return efficiency
end




function fragment.next_overhangs(current_overhangs)

   local current_overhang_map = {}
   for _, overhang in ipairs(current_overhangs) do
      current_overhang_map[overhang] = true
   end





   local bases = { "A", "T", "G", "C" }
   local overhangs_to_test = {}
   for _, base1 in ipairs(bases) do
      for _, base2 in ipairs(bases) do
         for _, base3 in ipairs(bases) do
            for _, base4 in ipairs(bases) do
               local new_overhang = base1 .. base2 .. base3 .. base4
               local in_current = current_overhang_map[new_overhang]
               local in_current_reverse = current_overhang_map[transform.reverse_complement(new_overhang)]
               if not in_current and not in_current_reverse then
                  if not is_palindromic(new_overhang) then
                     table.insert(overhangs_to_test, new_overhang)
                  end
               end
            end
         end
      end
   end

   local efficiencies = {}
   for _, overhang in ipairs(overhangs_to_test) do
      local current_with_overhang = {}
      for _, o in ipairs(current_overhangs) do table.insert(current_with_overhang, o) end
      table.insert(current_with_overhang, overhang)

      local strand_efficiency = fragment.set_efficiency(current_with_overhang)

      local current_with_complement = {}
      for _, o in ipairs(current_overhangs) do table.insert(current_with_complement, o) end
      table.insert(current_with_complement, transform.reverse_complement(overhang))

      local complement_efficiency = fragment.set_efficiency(current_with_complement)
      table.insert(efficiencies, (strand_efficiency + complement_efficiency) / 2)
   end

   return overhangs_to_test, efficiencies
end





function fragment.next_overhang(current_overhangs)
   local overhangs_to_test, efficiencies = fragment.next_overhangs(current_overhangs)
   local max_efficiency = 0.0
   local new_overhang = ""

   for i, overhang in ipairs(overhangs_to_test) do
      local efficiency = efficiencies[i]
      if efficiency > max_efficiency then
         max_efficiency = efficiency
         new_overhang = overhang
      end
   end

   return new_overhang
end


local function optimize_overhang_iteration(sequence, min_fragment_size, max_fragment_size, existing_fragments, exclude_overhangs, include_overhangs)
   local recurse_max_fragment_size = max_fragment_size


   if #sequence < max_fragment_size then
      table.insert(existing_fragments, sequence)
      return existing_fragments, fragment.set_efficiency(exclude_overhangs), nil
   end


   if min_fragment_size > max_fragment_size then
      return {}, 0, string.format("min_fragment_size (%d) larger than max_fragment_size (%d)", min_fragment_size, max_fragment_size)
   end





   if min_fragment_size < 12 then
      return {}, 0, string.format("min_fragment_size must be equal to or greater than 12. Got size of %d", min_fragment_size)
   end





   if #sequence < 2 * max_fragment_size then
      local max_and_min_difference = max_fragment_size - min_fragment_size
      local max_fragment_size_buffer = math.floor(#sequence / 2)
      min_fragment_size = max_fragment_size_buffer - max_and_min_difference
      if min_fragment_size < 12 then
         min_fragment_size = 12
      end
      max_fragment_size = max_fragment_size_buffer
   end


   local best_overhang_efficiency = 0.0
   local best_overhang_position = 0

   for overhang_offset = 0, max_fragment_size - min_fragment_size do
      local overhang_position = max_fragment_size - overhang_offset
      if overhang_position <= #sequence - 4 then
         local overhang_to_test = sequence:sub(overhang_position - 3, overhang_position)


         local already_exists = false
         for _, exclude_overhang in ipairs(exclude_overhangs) do
            if exclude_overhang == overhang_to_test or transform.reverse_complement(exclude_overhang) == overhang_to_test then
               already_exists = true
               break
            end
         end



         local build_available = #include_overhangs == 0
         for _, include_overhang in ipairs(include_overhangs) do
            if include_overhang == overhang_to_test or transform.reverse_complement(include_overhang) == overhang_to_test then
               build_available = true
               break
            end
         end

         if not already_exists and build_available and not is_palindromic(overhang_to_test) then
            local temp_exclude = {}
            for _, o in ipairs(exclude_overhangs) do table.insert(temp_exclude, o) end
            table.insert(temp_exclude, overhang_to_test)
            local set_efficiency = fragment.set_efficiency(temp_exclude)

            if set_efficiency > best_overhang_efficiency then
               best_overhang_efficiency = set_efficiency
               best_overhang_position = overhang_position
            end
         end
      end
   end

   if best_overhang_position == 0 then
      return {}, 0, "best_overhang_position failed by equaling zero"
   end

   table.insert(existing_fragments, sequence:sub(1, best_overhang_position))
   table.insert(exclude_overhangs, sequence:sub(best_overhang_position - 3, best_overhang_position))

   return optimize_overhang_iteration(
   sequence:sub(best_overhang_position - 3),
   min_fragment_size,
   recurse_max_fragment_size,
   existing_fragments,
   exclude_overhangs,
   include_overhangs)

end





function fragment.fragment(sequence, min_fragment_size, max_fragment_size, exclude_overhangs)
   sequence = string.upper(sequence)
   local initial_overhangs = { sequence:sub(1, 4), sequence:sub(#sequence - 3) }
   for _, o in ipairs(exclude_overhangs) do
      table.insert(initial_overhangs, o)
   end
   return optimize_overhang_iteration(sequence, min_fragment_size, max_fragment_size, {}, initial_overhangs, {})
end




function fragment.fragment_with_overhangs(sequence, min_fragment_size, max_fragment_size, exclude_overhangs, include_overhangs)
   sequence = string.upper(sequence)
   local initial_overhangs = { sequence:sub(1, 4), sequence:sub(#sequence - 3) }
   for _, o in ipairs(exclude_overhangs) do
      table.insert(initial_overhangs, o)
   end
   return optimize_overhang_iteration(sequence, min_fragment_size, max_fragment_size, {}, initial_overhangs, include_overhangs)
end














local function recursive_fragment_iteration(sequence, max_coding_size_oligo, assembly_pattern, exclude_overhangs, include_overhangs, forward_flank, reverse_flank, iteration)

















   local SMALLEST_MIN_FRAGMENT_SIZE_SUBTRACTION = 60
   local MIN_FRAGMENT_SIZE_SUBTRACTION = 100


   sequence = string.gsub(sequence, "\n", "")
   local sequence_len = #sequence


   local append_length = #forward_flank + #reverse_flank
   local sizes = {}
   local max_size = (max_coding_size_oligo - append_length) * assembly_pattern[1]

   for i = 1, #assembly_pattern do
      if i == 1 then
         sizes[i] = max_size
      else
         sizes[i] = sizes[i - 1] * assembly_pattern[i] - SMALLEST_MIN_FRAGMENT_SIZE_SUBTRACTION
      end
   end

   local target_sequence = sequence
   if iteration ~= 0 then
      target_sequence = forward_flank .. sequence .. reverse_flank
   end


   if sequence_len <= sizes[1] then
      local fragments, efficiency, err = fragment.fragment_with_overhangs(
      target_sequence,
      max_coding_size_oligo - 60,
      max_coding_size_oligo,
      exclude_overhangs,
      include_overhangs)

      if err ~= nil then
         return fragment.Assembly, err
      end
      return {
         sequence = sequence,
         fragments = fragments,
         efficiency = efficiency,
         sub_assemblies = {},
      }, nil
   end


   for i = 1, #sizes - 1 do
      if sequence_len <= sizes[i + 1] then
         local fragments, efficiency, err = fragment.fragment_with_overhangs(
         target_sequence,
         sizes[i] - MIN_FRAGMENT_SIZE_SUBTRACTION,
         sizes[i],
         exclude_overhangs,
         include_overhangs)

         if err ~= nil then
            return fragment.Assembly, err
         end

         local sub_assemblies = {}
         for _, frag in ipairs(fragments) do
            local sub_assembly, sub_err = recursive_fragment_iteration(
            frag,
            max_coding_size_oligo,
            assembly_pattern,
            exclude_overhangs,
            include_overhangs,
            forward_flank,
            reverse_flank,
            iteration + 1)

            if sub_err ~= "" then
               return fragment.Assembly, sub_err
            end
            table.insert(sub_assemblies, sub_assembly)
         end

         return {
            sequence = sequence,
            fragments = {},
            efficiency = efficiency,
            sub_assemblies = sub_assemblies,
         }, nil
      end
   end

   return fragment.Assembly, "Fragment too long!"
end
























function fragment.recursive_fragment(sequence, max_coding_size_oligo, assembly_pattern, exclude_overhangs, include_overhangs, forward_flank, reverse_flank)
   return recursive_fragment_iteration(sequence, max_coding_size_oligo, assembly_pattern, exclude_overhangs, include_overhangs, forward_flank, reverse_flank, 0)
end


















































local clone = {}

































































local default_enzymes = {
   BsaI = {
      name = "BsaI",
      pattern_for = "GGTCTC",
      pattern_rev = "GAGACC",
      skip = 1,
      overhead_length = 4,
      recognition_site = "GGTCTC",
   },
   BbsI = {
      name = "BbsI",
      pattern_for = "GAAGAC",
      pattern_rev = "GTCTTC",
      skip = 2,
      overhead_length = 4,
      recognition_site = "GAAGAC",
   },
   BtgZI = {
      name = "BtgZI",
      pattern_for = "GCGATG",
      pattern_rev = "CATCGC",
      skip = 10,
      overhead_length = 4,
      recognition_site = "GCGATG",
   },
   PaqCI = {
      name = "PaqCI",
      pattern_for = "CACCTGC",
      pattern_rev = "GCAGGTG",
      skip = 4,
      overhead_length = 4,
      recognition_site = "CACCTGC",
   },
   BsmBI = {
      name = "BsmBI",
      pattern_for = "CGTCTC",
      pattern_rev = "GAGACG",
      skip = 1,
      overhead_length = 4,
      recognition_site = "CGTCTC",
   },
}

clone.default_enzymes = default_enzymes








local function sort_overhangs(overhangs)
   table.sort(overhangs, function(a, b)
      return a.position < b.position
   end)
end





function clone.cut_with_enzyme_by_name(part, directional, name, methylated)

   local enzyme = default_enzymes[name]
   if not enzyme then
      return {}, "enzyme not found"
   end

   return clone.cut_with_enzyme(part, directional, enzyme, methylated), nil
end




function clone.cut_with_enzyme(part, directional, enzyme, methylated)
   local fragments = {}
   local fragment_sequences = {}


   local sequence = part.sequence
   if part.circular then
      sequence = sequence .. sequence
   end



   if not methylated then
      sequence = string.upper(sequence)
   end


   local palindromic = enzyme.recognition_site == string.reverse(enzyme.recognition_site)


   local overhangs = {}
   local forward_overhangs = {}
   local reverse_overhangs = {}


   local pos = 1
   while true do
      local start = string.find(sequence, enzyme.pattern_for, pos, true)
      if not start then break end
      table.insert(forward_overhangs, {
         length = enzyme.overhead_length,
         position = start + #enzyme.pattern_for + enzyme.skip,
         forward = true,
         recognition_site_plus_skip_length = #enzyme.recognition_site + enzyme.skip,
      })
      pos = start + 1
   end


   if not palindromic then
      pos = 1
      while true do
         local start = string.find(sequence, enzyme.pattern_rev, pos, true)
         if not start then break end
         table.insert(reverse_overhangs, {
            length = enzyme.overhead_length,
            position = start - enzyme.skip,
            forward = false,
            recognition_site_plus_skip_length = #enzyme.recognition_site + enzyme.skip,
         })
         pos = start + 1
      end
   end


   for _, overhang_set in ipairs({ forward_overhangs, reverse_overhangs }) do
      if #overhang_set > 0 then
         if not part.circular and
            (overhang_set[#overhang_set].position + enzyme.skip + enzyme.overhead_length > #sequence) then
            overhang_set[#overhang_set] = nil
         end
      end
      for _, overhang in ipairs(overhang_set) do
         table.insert(overhangs, overhang)
      end
   end


   sort_overhangs(overhangs)



   if #overhangs == 1 and not directional and not part.circular then
      local fragment_sequence1
      local fragment_sequence2
      local overhang_sequence

      if #forward_overhangs > 0 then

         fragment_sequence1 = sequence:sub(overhangs[1].position + overhangs[1].length)
         fragment_sequence2 = sequence:sub(1, overhangs[1].position - 1)
         overhang_sequence = sequence:sub(overhangs[1].position,
         overhangs[1].position + overhangs[1].length - 1)
         table.insert(fragments, {
            sequence = fragment_sequence1,
            forward_overhang = overhang_sequence,
            reverse_overhang = "",
         })
         table.insert(fragments, {
            sequence = fragment_sequence2,
            forward_overhang = "",
            reverse_overhang = overhang_sequence,
         })
      else
         fragment_sequence1 = sequence:sub(overhangs[1].position)
         fragment_sequence2 = sequence:sub(1, overhangs[1].position - overhangs[1].length - 1)
         overhang_sequence = sequence:sub(overhangs[1].position - overhangs[1].length,
         overhangs[1].position - 1)
         table.insert(fragments, {
            sequence = fragment_sequence2,
            forward_overhang = "",
            reverse_overhang = overhang_sequence,
         })
         table.insert(fragments, {
            sequence = fragment_sequence1,
            forward_overhang = overhang_sequence,
            reverse_overhang = "",
         })
      end
      return fragments
   end


   if #overhangs == 2 and not directional and part.circular then
      local fragment_sequence1 = sequence:sub(overhangs[1].position + overhangs[1].length,
      #part.sequence)
      local fragment_sequence2 = sequence:sub(1, overhangs[1].position - 1)
      local fragment_sequence = fragment_sequence1 .. fragment_sequence2
      local overhang_sequence = sequence:sub(overhangs[1].position,
      overhangs[1].position + overhangs[1].length - 1)
      table.insert(fragments, {
         sequence = fragment_sequence,
         forward_overhang = overhang_sequence,
         reverse_overhang = overhang_sequence,
      })
      return fragments
   end


   if #overhangs > 1 then
      for i = 1, #overhangs - 1 do
         local current_overhang = overhangs[i]
         local next_overhang = overhangs[i + 1]
         if directional and not palindromic then
            if current_overhang.forward and not next_overhang.forward then
               table.insert(fragment_sequences,
               sequence:sub(current_overhang.position, next_overhang.position - 1))
            end
            if next_overhang.position - next_overhang.recognition_site_plus_skip_length > #part.sequence then
               break
            end
         else
            table.insert(fragment_sequences,
            sequence:sub(current_overhang.position, next_overhang.position - 1))
            if next_overhang.position - next_overhang.recognition_site_plus_skip_length > #part.sequence then
               break
            end
         end
      end


      for _, fragment_sequence in ipairs(fragment_sequences) do
         if #fragment_sequence > 8 then
            local fragment_seq = fragment_sequence:sub(enzyme.overhead_length + 1,
            #fragment_sequence - enzyme.overhead_length)
            local forward_overhang = fragment_sequence:sub(1, enzyme.overhead_length)
            local reverse_overhang = fragment_sequence:sub(#fragment_sequence - enzyme.overhead_length + 1)
            table.insert(fragments, {
               sequence = fragment_seq,
               forward_overhang = forward_overhang,
               reverse_overhang = reverse_overhang,
            })
         end
      end
   end

   return fragments
end











function clone.ligate(fragments, circular)
   if #fragments == 0 then
      return "", {}, "no fragments to ligate"
   end



   local ligation_pattern = { 1 }

   local final_fragment = {
      sequence = fragments[1].sequence,
      forward_overhang = fragments[1].forward_overhang,
      reverse_overhang = fragments[1].reverse_overhang,
   }

   local used = { [1] = true }
   local match_found = true


   while match_found do
      match_found = false
      for i = 1, #fragments do
         if not used[i] then
            if final_fragment.reverse_overhang == fragments[i].forward_overhang then
               final_fragment.sequence = final_fragment.sequence ..
               final_fragment.reverse_overhang ..
               fragments[i].sequence
               final_fragment.reverse_overhang = fragments[i].reverse_overhang
               used[i] = true
               match_found = true
               table.insert(ligation_pattern, i)
               break
            end
            if final_fragment.reverse_overhang == transform.reverse_complement(fragments[i].reverse_overhang) then
               final_fragment.sequence = final_fragment.sequence ..
               final_fragment.reverse_overhang ..
               transform.reverse_complement(fragments[i].sequence)
               final_fragment.reverse_overhang = transform.reverse_complement(fragments[i].forward_overhang)
               used[i] = true
               match_found = true
               table.insert(ligation_pattern, i)
               break
            end
         end
      end
   end


   if circular then
      if final_fragment.forward_overhang ~= final_fragment.reverse_overhang then
         return "", ligation_pattern, "does not circularize"
      end
      return final_fragment.forward_overhang .. final_fragment.sequence,
      ligation_pattern,
      nil
   end

   return final_fragment.forward_overhang ..
   final_fragment.sequence ..
   final_fragment.reverse_overhang,
   ligation_pattern,
   nil
end






function clone.golden_gate(sequences, cutting_enzyme, methylated)
   local fragments = {}

   for _, sequence in ipairs(sequences) do
      local new_fragments = clone.cut_with_enzyme(sequence, true, cutting_enzyme, methylated)
      for _, frag in ipairs(new_fragments) do
         table.insert(fragments, frag)
      end
   end

   return clone.ligate(fragments, true)
end
















local function standardize_kmer(kmer)

   kmer = string.upper(kmer)

   local rev_comp = transform.reverse_complement(kmer)

   if kmer < rev_comp then
      return kmer
   end
   return rev_comp
end

function clone.find_kmer_overlaps(
   fragments,
   ligation_product,
   ligation_pattern,
   kmer_size)

   if #fragments < 2 then
      return {}, "need at least two fragments to find overlaps"
   end

   local bp_from_each = math.floor((kmer_size - 4) / 2)
   if bp_from_each < 4 then
      return {}, "need at least a kmer of 12"
   end

   local ligation = ligation_product .. ligation_product
   local position = 1
   local kmer_overlaps = {}

   for i = 1, #ligation_pattern do
      local frag1 = fragments[ligation_pattern[i]]
      local frag2
      if i == #ligation_pattern then
         frag2 = fragments[ligation_pattern[1]]
      else
         frag2 = fragments[ligation_pattern[i + 1]]
      end

      position = position + #frag1.forward_overhang + #frag1.sequence
      local kmer = ligation:sub(
      position - bp_from_each,
      position + 3 + bp_from_each)


      kmer = standardize_kmer(kmer)

      table.insert(kmer_overlaps, {
         kmer = kmer,
         fragment1 = frag1,
         fragment2 = frag2,
      })
   end

   return kmer_overlaps, nil
end






function clone.find_kmers(kmer_overlaps, sequence)
   local output_kmer_overlaps = {}
   sequence = string.upper(sequence)

   for _, kmer_overlap in ipairs(kmer_overlaps) do
      if string.find(sequence, string.upper(kmer_overlap.kmer), 1, true) or
         string.find(sequence, string.upper(transform.reverse_complement(kmer_overlap.kmer)), 1, true) then
         table.insert(output_kmer_overlaps, {
            kmer = kmer_overlap.kmer,
            fragment1 = kmer_overlap.fragment1,
            fragment2 = kmer_overlap.fragment2,
         })
      end
   end

   return output_kmer_overlaps
end









function clone.get_window_from_fragment(fragment, left_flank_length, right_flank_length)

   local full_sequence = fragment.sequence


   local left_flank = full_sequence:sub(1, left_flank_length)


   local right_flank = full_sequence:sub(#full_sequence - right_flank_length + 1)


   return right_flank, left_flank
end























local codon = {}























local errEmptyAminoAcidString = "empty amino acid string"
local errEmptySequenceString = "empty sequence string"
























function codon.copy_translation_table(t)
   local new_t = {
      start_codons = {},
      stop_codons = {},
      amino_acids = {},
      translation_map = {},
      start_codon_table = {},
      translate = t.translate,
      optimize = t.optimize,
      standardize_last_codon = t.standardize_last_codon,
   }


   for _, cdn in ipairs(t.start_codons) do
      table.insert(new_t.start_codons, cdn)
   end
   for _, cdn in ipairs(t.stop_codons) do
      table.insert(new_t.stop_codons, cdn)
   end


   for _, aa in ipairs(t.amino_acids) do
      local new_codons = {}
      for _, cdn in ipairs(aa.codons) do
         table.insert(new_codons, {
            triplet = cdn.triplet,
            weight = cdn.weight,
         })
      end
      table.insert(new_t.amino_acids, {
         letter = aa.letter,
         codons = new_codons,
      })
   end


   for k, v in pairs(t.translation_map) do
      new_t.translation_map[k] = v
   end
   for k, v in pairs(t.start_codon_table) do
      new_t.start_codon_table[k] = v
   end

   return new_t
end


function codon.get_stochastic_codon(aa)
   local total_weight = 0
   local cumulative_weights = {}


   for i, cdn in ipairs(aa.codons) do
      total_weight = total_weight + cdn.weight
      cumulative_weights[i] = total_weight
   end

   if total_weight == 0 then
      return nil
   end


   local r = math.random(0, total_weight - 1)


   for i, cdn in ipairs(aa.codons) do
      if r < cumulative_weights[i] then
         return cdn.triplet
      end
   end

   return nil
end


function codon.get_codon_frequency(sequence)
   local cdn_freq_map = {}
   sequence = string.upper(sequence)

   for i = 1, #sequence - 2, 3 do
      local cdn = sequence:sub(i, i + 2)
      cdn_freq_map[cdn] = (cdn_freq_map[cdn] or 0) + 1
   end

   return cdn_freq_map
end


function codon.weight_amino_acids(sequence, amino_acids)
   sequence = string.upper(sequence)
   local cdn_freq_map = codon.get_codon_frequency(sequence)

   local weighted_amino_acids = {}

   for _, amino_acid in ipairs(amino_acids) do
      local new_codons = {}
      for _, cdn in ipairs(amino_acid.codons) do
         table.insert(new_codons, {
            triplet = cdn.triplet,
            weight = cdn_freq_map[cdn.triplet] or 0,
         })
      end
      table.insert(weighted_amino_acids, {
         letter = amino_acid.letter,
         codons = new_codons,
      })
   end

   return weighted_amino_acids
end


local function translate(self, dna_seq)
   if dna_seq == "" then
      return "", errEmptySequenceString
   end

   local current_amino_acids = {}
   dna_seq = string.upper(dna_seq)

   for i = 1, #dna_seq - 2, 3 do
      local cdn = dna_seq:sub(i, i + 2)
      local aa = self.translation_map[cdn]
      if aa then
         table.insert(current_amino_acids, aa)
      end
   end

   return table.concat(current_amino_acids), nil
end


local function optimize(self, aas)

   local amino_acids = string.upper(aas)

   if amino_acids == "" then
      return "", errEmptyAminoAcidString
   end

   local codons = {}
   for i = 1, #amino_acids do
      local amino_acid = amino_acids:sub(i, i)
      local found = false

      for _, aa in ipairs(self.amino_acids) do
         if amino_acid == aa.letter then
            table.insert(codons, codon.get_stochastic_codon(aa))
            found = true
            break
         end
      end

      if not found then
         return "", string.format('amino acid "%s" is missing from codon table', amino_acid)
      end
   end

   return table.concat(codons), nil
end


local standard_codons = {
   K = "AAA", L = "TTA", R = "CGA", F = "TTT", Q = "CAA", W = "TGG",
   H = "CAC", P = "CCG", V = "GTG", C = "TGC", T = "ACA", S = "TCC",
   Y = "TAC", I = "ATA", G = "GGT", N = "AAC", A = "GCC", M = "ATG",
   D = "GAC", E = "GAA",
}


local function standardize_last_codon(self, dna_seq)
   if dna_seq == "" then
      return "", errEmptySequenceString
   end

   dna_seq = string.upper(dna_seq)
   local seq_len = #dna_seq

   if seq_len < 3 then
      return dna_seq, nil
   end

   local last_codon = dna_seq:sub(seq_len - 2, seq_len)
   local aa = self.translation_map[last_codon]

   if not aa then
      return dna_seq, string.format('unknown codon "%s"', last_codon)
   end

   local standard_codon = standard_codons[aa]
   if not standard_codon then
      return dna_seq, string.format('no standard codon for amino acid "%s"', aa)
   end

   return dna_seq:sub(1, seq_len - 3) .. standard_codon, nil
end


function codon.generate_codon_table(amino_acids, starts)

   local base1 = "TTTTTTTTTTTTTTTTCCCCCCCCCCCCCCCCAAAAAAAAAAAAAAAAGGGGGGGGGGGGGGGG"
   local base2 = "TTTTCCCCAAAAGGGGTTTTCCCCAAAAGGGGTTTTCCCCAAAAGGGGTTTTCCCCAAAAGGGG"
   local base3 = "TCAGTCAGTCAGTCAGTCAGTCAGTCAGTCAGTCAGTCAGTCAGTCAGTCAGTCAGTCAGTCAG"


   local amino_acid_map = {}
   local start_codons = {}
   local stop_codons = {}
   local translation_map = {}


   for i = 1, #amino_acids do
      local aa = amino_acids:sub(i, i)
      local start = starts:sub(i, i)
      local triplet = base1:sub(i, i) .. base2:sub(i, i) .. base3:sub(i, i)


      if not amino_acid_map[aa] then
         amino_acid_map[aa] = {}
      end


      table.insert(amino_acid_map[aa], { triplet = triplet, weight = 1 })
      translation_map[triplet] = aa


      if start == "M" then
         table.insert(start_codons, triplet)
      end


      if aa == "*" then
         table.insert(stop_codons, triplet)
      end
   end


   local amino_acids_array = {}
   for aa, codons in pairs(amino_acid_map) do
      table.insert(amino_acids_array, {
         letter = aa,
         codons = codons,
      })
   end


   local start_codon_table = {}
   for _, cdn in ipairs(start_codons) do
      start_codon_table[cdn] = "M"
   end


   return {
      start_codons = start_codons,
      stop_codons = stop_codons,
      amino_acids = amino_acids_array,
      translation_map = translation_map,
      start_codon_table = start_codon_table,
      translate = translate,
      optimize = optimize,
      standardize_last_codon = standardize_last_codon,
   }
end



function codon.compromise_codon_table(first_t, second_t, cut_off)

   if cut_off < 0 then
      return first_t, "cut off too low, cannot be less than 0"
   end
   if cut_off > 1 then
      return first_t, "cut off too high, cannot be greater than 1"
   end


   local merged_t = codon.copy_translation_table(first_t)
   local final_amino_acids = {}


   for _, first_aa in ipairs(first_t.amino_acids) do
      local first_triplets = {}
      local first_weights = {}
      local first_total = 0

      local second_weights = {}
      local second_total = 0


      for _, first_cdn in ipairs(first_aa.codons) do
         table.insert(first_triplets, first_cdn.triplet)
         table.insert(first_weights, first_cdn.weight)
         first_total = first_total + first_cdn.weight


         for _, second_aa in ipairs(second_t.amino_acids) do
            if second_aa.letter == first_aa.letter then
               for _, second_cdn in ipairs(second_aa.codons) do
                  if second_cdn.triplet == first_cdn.triplet then
                     table.insert(second_weights, second_cdn.weight)
                     second_total = second_total + second_cdn.weight
                     break
                  end
               end
            end
         end
      end

      local final_codons = {}
      local cut_off_weight = math.floor(10000 * cut_off)


      for i, triplet in ipairs(first_triplets) do
         local first_weight = math.floor((first_weights[i] / first_total) * 10000)
         local second_weight = math.floor((second_weights[i] / second_total) * 10000)

         if first_weight < cut_off_weight or second_weight < cut_off_weight then
            table.insert(final_codons, { triplet = triplet, weight = 0 })
         else
            table.insert(final_codons, {
               triplet = triplet,
               weight = math.floor((first_weight + second_weight) / 2),
            })
         end
      end

      table.insert(final_amino_acids, {
         letter = first_aa.letter,
         codons = final_codons,
      })
   end

   merged_t.amino_acids = final_amino_acids
   return merged_t, nil
end


function codon.add_codon_table(first_t, second_t)
   local final_amino_acids = {}

   for _, first_aa in ipairs(first_t.amino_acids) do
      local final_codons = {}

      for _, first_cdn in ipairs(first_aa.codons) do
         for _, second_aa in ipairs(second_t.amino_acids) do
            for _, second_cdn in ipairs(second_aa.codons) do
               if first_cdn.triplet == second_cdn.triplet then
                  table.insert(final_codons, {
                     triplet = first_cdn.triplet,
                     weight = first_cdn.weight + second_cdn.weight,
                  })
               end
            end
         end
      end

      table.insert(final_amino_acids, {
         letter = first_aa.letter,
         codons = final_codons,
      })
   end

   local merged_t = codon.copy_translation_table(first_t)
   merged_t.amino_acids = final_amino_acids
   return merged_t, nil
end


local translation_tables_by_number = {
   [1] = codon.generate_codon_table("FFLLSSSSYY**CC*WLLLLPPPPHHQQRRRRIIIMTTTTNNKKSSRRVVVVAAAADDEEGGGG", "---M------**--*----M---------------M----------------------------"),
   [2] = codon.generate_codon_table("FFLLSSSSYY**CCWWLLLLPPPPHHQQRRRRIIMMTTTTNNKKSS**VVVVAAAADDEEGGGG", "----------**--------------------MMMM----------**---M------------"),
   [3] = codon.generate_codon_table("FFLLSSSSYY**CCWWTTTTPPPPHHQQRRRRIIMMTTTTNNKKSSRRVVVVAAAADDEEGGGG", "----------**----------------------MM---------------M------------"),
   [4] = codon.generate_codon_table("FFLLSSSSYY**CCWWLLLLPPPPHHQQRRRRIIIMTTTTNNKKSSRRVVVVAAAADDEEGGGG", "--MM------**-------M------------MMMM---------------M------------"),
   [5] = codon.generate_codon_table("FFLLSSSSYY**CCWWLLLLPPPPHHQQRRRRIIMMTTTTNNKKSSSSVVVVAAAADDEEGGGG", "---M------**--------------------MMMM---------------M------------"),
   [6] = codon.generate_codon_table("FFLLSSSSYYQQCC*WLLLLPPPPHHQQRRRRIIIMTTTTNNKKSSRRVVVVAAAADDEEGGGG", "--------------*--------------------M----------------------------"),
   [9] = codon.generate_codon_table("FFLLSSSSYY**CCWWLLLLPPPPHHQQRRRRIIIMTTTTNNNKSSSSVVVVAAAADDEEGGGG", "----------**-----------------------M---------------M------------"),
   [10] = codon.generate_codon_table("FFLLSSSSYY**CCCWLLLLPPPPHHQQRRRRIIIMTTTTNNKKSSRRVVVVAAAADDEEGGGG", "----------**-----------------------M----------------------------"),
   [11] = codon.generate_codon_table("FFLLSSSSYY**CC*WLLLLPPPPHHQQRRRRIIIMTTTTNNKKSSRRVVVVAAAADDEEGGGG", "---M------**--*----M------------MMMM---------------M------------"),
   [12] = codon.generate_codon_table("FFLLSSSSYY**CC*WLLLSPPPPHHQQRRRRIIIMTTTTNNKKSSRRVVVVAAAADDEEGGGG", "----------**--*----M---------------M----------------------------"),
   [13] = codon.generate_codon_table("FFLLSSSSYY**CCWWLLLLPPPPHHQQRRRRIIMMTTTTNNKKSSGGVVVVAAAADDEEGGGG", "---M------**----------------------MM---------------M------------"),
   [14] = codon.generate_codon_table("FFLLSSSSYYY*CCWWLLLLPPPPHHQQRRRRIIIMTTTTNNNKSSSSVVVVAAAADDEEGGGG", "-----------*-----------------------M----------------------------"),
   [16] = codon.generate_codon_table("FFLLSSSSYY*LCC*WLLLLPPPPHHQQRRRRIIIMTTTTNNKKSSRRVVVVAAAADDEEGGGG", "----------*---*--------------------M----------------------------"),
   [21] = codon.generate_codon_table("FFLLSSSSYY**CCWWLLLLPPPPHHQQRRRRIIMMTTTTNNNKSSSSVVVVAAAADDEEGGGG", "----------**-----------------------M---------------M------------"),
   [22] = codon.generate_codon_table("FFLLSS*SYY*LCC*WLLLLPPPPHHQQRRRRIIIMTTTTNNKKSSRRVVVVAAAADDEEGGGG", "------*---*---*--------------------M----------------------------"),
   [23] = codon.generate_codon_table("FF*LSSSSYY**CC*WLLLLPPPPHHQQRRRRIIIMTTTTNNKKSSRRVVVVAAAADDEEGGGG", "--*-------**--*-----------------M--M---------------M------------"),
   [24] = codon.generate_codon_table("FFLLSSSSYY**CCWWLLLLPPPPHHQQRRRRIIIMTTTTNNKKSSSKVVVVAAAADDEEGGGG", "---M------**-------M---------------M---------------M------------"),
   [25] = codon.generate_codon_table("FFLLSSSSYY**CCGWLLLLPPPPHHQQRRRRIIIMTTTTNNKKSSRRVVVVAAAADDEEGGGG", "---M------**-----------------------M---------------M------------"),
   [26] = codon.generate_codon_table("FFLLSSSSYY**CC*WLLLAPPPPHHQQRRRRIIIMTTTTNNKKSSRRVVVVAAAADDEEGGGG", "----------**--*----M---------------M----------------------------"),
   [27] = codon.generate_codon_table("FFLLSSSSYYQQCCWWLLLLPPPPHHQQRRRRIIIMTTTTNNKKSSRRVVVVAAAADDEEGGGG", "--------------*--------------------M----------------------------"),
   [28] = codon.generate_codon_table("FFLLSSSSYYQQCCWWLLLLPPPPHHQQRRRRIIIMTTTTNNKKSSRRVVVVAAAADDEEGGGG", "----------**--*--------------------M----------------------------"),
   [29] = codon.generate_codon_table("FFLLSSSSYYYYCC*WLLLLPPPPHHQQRRRRIIIMTTTTNNKKSSRRVVVVAAAADDEEGGGG", "--------------*--------------------M----------------------------"),
   [30] = codon.generate_codon_table("FFLLSSSSYYEECC*WLLLLPPPPHHQQRRRRIIIMTTTTNNKKSSRRVVVVAAAADDEEGGGG", "--------------*--------------------M----------------------------"),
   [31] = codon.generate_codon_table("FFLLSSSSYYEECCWWLLLLPPPPHHQQRRRRIIIMTTTTNNKKSSRRVVVVAAAADDEEGGGG", "----------**-----------------------M----------------------------"),
   [33] = codon.generate_codon_table("FFLLSSSSYYY*CCWWLLLLPPPPHHQQRRRRIIIMTTTTNNKKSSSKVVVVAAAADDEEGGGG", "---M-------*-------M---------------M---------------M------------"),
}
codon.translation_tables_by_number = translation_tables_by_number


function codon.new_translation_table(index)
   return codon.copy_translation_table(translation_tables_by_number[index])
end


codon.default_tables = {
   ["ecoli"] = {
      start_codons = { "TTG", "CTG", "ATT", "ATC", "ATA", "ATG", "GTG" },
      stop_codons = { "TAA", "TAG", "TGA" },
      amino_acids = {
         { letter = "Y", codons = { { triplet = "TAT", weight = 42 }, { triplet = "TAC", weight = 58 } } },
         { letter = "C", codons = { { triplet = "TGT", weight = 42 }, { triplet = "TGC", weight = 58 } } },
         { letter = "I", codons = { { triplet = "ATT", weight = 49 }, { triplet = "ATC", weight = 51 }, { triplet = "ATA", weight = 0 } } },
         { letter = "V", codons = { { triplet = "GTT", weight = 35 }, { triplet = "GTC", weight = 28 }, { triplet = "GTA", weight = 3 }, { triplet = "GTG", weight = 34 } } },
         { letter = "G", codons = { { triplet = "GGT", weight = 60 }, { triplet = "GGC", weight = 39 }, { triplet = "GGA", weight = 0 }, { triplet = "GGG", weight = 0 } } },
         { letter = "L", codons = { { triplet = "TTA", weight = 3 }, { triplet = "TTG", weight = 14 }, { triplet = "CTT", weight = 2 }, { triplet = "CTC", weight = 3 }, { triplet = "CTA", weight = 0 }, { triplet = "CTG", weight = 78 } } },
         { letter = "W", codons = { { triplet = "TGG", weight = 1 } } },
         { letter = "K", codons = { { triplet = "AAA", weight = 51 }, { triplet = "AAG", weight = 49 } } },
         { letter = "S", codons = { { triplet = "TCT", weight = 10 }, { triplet = "TCC", weight = 13 }, { triplet = "TCA", weight = 2 }, { triplet = "TCG", weight = 5 }, { triplet = "AGT", weight = 0 }, { triplet = "AGC", weight = 68 } } },
         { letter = "*", codons = { { triplet = "TAA", weight = 2015 }, { triplet = "TAG", weight = 1667 }, { triplet = "TGA", weight = 1300 } } },
         { letter = "Q", codons = { { triplet = "CAA", weight = 45 }, { triplet = "CAG", weight = 55 } } },
         { letter = "R", codons = { { triplet = "CGT", weight = 62 }, { triplet = "CGC", weight = 35 }, { triplet = "CGA", weight = 0 }, { triplet = "CGG", weight = 0 }, { triplet = "AGA", weight = 3 }, { triplet = "AGG", weight = 0 } } },
         { letter = "A", codons = { { triplet = "GCT", weight = 12 }, { triplet = "GCC", weight = 19 }, { triplet = "GCA", weight = 24 }, { triplet = "GCG", weight = 44 } } },
         { letter = "E", codons = { { triplet = "GAA", weight = 43 }, { triplet = "GAG", weight = 57 } } },
         { letter = "F", codons = { { triplet = "TTT", weight = 45 }, { triplet = "TTC", weight = 55 } } },
         { letter = "P", codons = { { triplet = "CCT", weight = 9 }, { triplet = "CCC", weight = 0 }, { triplet = "CCA", weight = 10 }, { triplet = "CCG", weight = 81 } } },
         { letter = "H", codons = { { triplet = "CAT", weight = 38 }, { triplet = "CAC", weight = 62 } } },
         { letter = "M", codons = { { triplet = "ATG", weight = 1 } } },
         { letter = "T", codons = { { triplet = "ACT", weight = 10 }, { triplet = "ACC", weight = 57 }, { triplet = "ACA", weight = 0 }, { triplet = "ACG", weight = 33 } } },
         { letter = "N", codons = { { triplet = "AAT", weight = 47 }, { triplet = "AAC", weight = 53 } } },
         { letter = "D", codons = { { triplet = "GAT", weight = 46 }, { triplet = "GAC", weight = 54 } } },
      },
      translation_map = {},
      start_codon_table = {},
   },
   ["pichia"] = {
      start_codons = { "TTG", "CTG", "ATT", "ATC", "ATA", "ATG", "GTG" },
      stop_codons = { "TAA", "TAG", "TGA" },
      amino_acids = {
         { letter = "Y", codons = { { triplet = "TAT", weight = 40017 }, { triplet = "TAC", weight = 37740 } } },
         { letter = "C", codons = { { triplet = "TGT", weight = 17099 }, { triplet = "TGC", weight = 10242 } } },
         { letter = "I", codons = { { triplet = "ATT", weight = 68516 }, { triplet = "ATC", weight = 43651 }, { triplet = "ATA", weight = 35059 } } },
         { letter = "V", codons = { { triplet = "GTT", weight = 54750 }, { triplet = "GTC", weight = 30526 }, { triplet = "GTA", weight = 25054 }, { triplet = "GTG", weight = 30581 } } },
         { letter = "G", codons = { { triplet = "GGT", weight = 42959 }, { triplet = "GGC", weight = 18853 }, { triplet = "GGA", weight = 43541 }, { triplet = "GGG", weight = 14618 } } },
         { letter = "L", codons = { { triplet = "TTA", weight = 41481 }, { triplet = "TTG", weight = 68335 }, { triplet = "CTT", weight = 40288 }, { triplet = "CTC", weight = 20003 }, { triplet = "CTA", weight = 29034 }, { triplet = "CTG", weight = 35916 } } },
         { letter = "W", codons = { { triplet = "TGG", weight = 23941 } } },
         { letter = "K", codons = { { triplet = "AAA", weight = 83571 }, { triplet = "AAG", weight = 77197 } } },
         { letter = "S", codons = { { triplet = "TCT", weight = 53665 }, { triplet = "TCC", weight = 35643 }, { triplet = "TCA", weight = 43185 }, { triplet = "TCG", weight = 19746 }, { triplet = "AGT", weight = 32769 }, { triplet = "AGC", weight = 21832 } } },
         { letter = "*", codons = { { triplet = "TAA", weight = 2015 }, { triplet = "TAG", weight = 1667 }, { triplet = "TGA", weight = 1300 } } },
         { letter = "Q", codons = { { triplet = "CAA", weight = 58688 }, { triplet = "CAG", weight = 38500 } } },
         { letter = "R", codons = { { triplet = "CGT", weight = 14716 }, { triplet = "CGC", weight = 5515 }, { triplet = "CGA", weight = 12855 }, { triplet = "CGG", weight = 5643 }, { triplet = "AGA", weight = 47972 }, { triplet = "AGG", weight = 19381 } } },
         { letter = "A", codons = { { triplet = "GCT", weight = 51452 }, { triplet = "GCC", weight = 30978 }, { triplet = "GCA", weight = 35840 }, { triplet = "GCG", weight = 10148 } } },
         { letter = "E", codons = { { triplet = "GAA", weight = 93407 }, { triplet = "GAG", weight = 64293 } } },
         { letter = "F", codons = { { triplet = "TTT", weight = 60424 }, { triplet = "TTC", weight = 43704 } } },
         { letter = "P", codons = { { triplet = "CCT", weight = 35821 }, { triplet = "CCC", weight = 18924 }, { triplet = "CCA", weight = 39324 }, { triplet = "CCG", weight = 10585 } } },
         { letter = "H", codons = { { triplet = "CAT", weight = 30739 }, { triplet = "CAC", weight = 19034 } } },
         { letter = "M", codons = { { triplet = "ATG", weight = 42837 } } },
         { letter = "T", codons = { { triplet = "ACT", weight = 47886 }, { triplet = "ACC", weight = 31320 }, { triplet = "ACA", weight = 36947 }, { triplet = "ACG", weight = 16313 } } },
         { letter = "N", codons = { { triplet = "AAT", weight = 66744 }, { triplet = "AAC", weight = 57670 } } },
         { letter = "D", codons = { { triplet = "GAT", weight = 84985 }, { triplet = "GAC", weight = 52486 } } },
      },
      translation_map = {},
      start_codon_table = {},
   },
   ["scerevisiae"] = {
      start_codons = { "TTG", "CTG", "ATG" },
      stop_codons = { "TAA", "TAG", "TGA" },
      amino_acids = {
         { letter = "F", codons = { { triplet = "TTT", weight = 69255 }, { triplet = "TTC", weight = 47162 } } },
         { letter = "L", codons = { { triplet = "TTA", weight = 69592 }, { triplet = "TTG", weight = 69917 }, { triplet = "CTT", weight = 32571 }, { triplet = "CTC", weight = 14569 }, { triplet = "CTA", weight = 35726 }, { triplet = "CTG", weight = 27963 } } },
         { letter = "S", codons = { { triplet = "TCT", weight = 61396 }, { triplet = "TCC", weight = 36989 }, { triplet = "TCA", weight = 50416 }, { triplet = "TCG", weight = 23012 }, { triplet = "AGT", weight = 38486 }, { triplet = "AGC", weight = 26233 } } },
         { letter = "Y", codons = { { triplet = "TAT", weight = 50491 }, { triplet = "TAC", weight = 38471 } } },
         { letter = "K", codons = { { triplet = "AAA", weight = 113360 }, { triplet = "AAG", weight = 80638 } } },
         { letter = "W", codons = { { triplet = "TGG", weight = 27342 } } },
         { letter = "V", codons = { { triplet = "GTT", weight = 56557 }, { triplet = "GTC", weight = 29536 }, { triplet = "GTA", weight = 31958 }, { triplet = "GTG", weight = 28127 } } },
         { letter = "*", codons = { { triplet = "TAA", weight = 2561 }, { triplet = "TAG", weight = 1243 }, { triplet = "TGA", weight = 1614 } } },
         { letter = "C", codons = { { triplet = "TGT", weight = 20749 }, { triplet = "TGC", weight = 12623 } } },
         { letter = "P", codons = { { triplet = "CCT", weight = 35886 }, { triplet = "CCC", weight = 18238 }, { triplet = "CCA", weight = 47055 }, { triplet = "CCG", weight = 14424 } } },
         { letter = "Q", codons = { { triplet = "CAA", weight = 71668 }, { triplet = "CAG", weight = 32701 } } },
         { letter = "I", codons = { { triplet = "ATT", weight = 79418 }, { triplet = "ATC", weight = 44763 }, { triplet = "ATA", weight = 48474 } } },
         { letter = "M", codons = { { triplet = "ATG", weight = 54773 } } },
         { letter = "T", codons = { { triplet = "ACT", weight = 52985 }, { triplet = "ACC", weight = 32808 }, { triplet = "ACA", weight = 47831 }, { triplet = "ACG", weight = 21401 } } },
         { letter = "A", codons = { { triplet = "GCT", weight = 53513 }, { triplet = "GCC", weight = 31917 }, { triplet = "GCA", weight = 42870 }, { triplet = "GCG", weight = 16338 } } },
         { letter = "D", codons = { { triplet = "GAT", weight = 100401 }, { triplet = "GAC", weight = 53688 } } },
         { letter = "E", codons = { { triplet = "GAA", weight = 120741 }, { triplet = "GAG", weight = 51544 } } },
         { letter = "G", codons = { { triplet = "GGT", weight = 59816 }, { triplet = "GGC", weight = 25856 }, { triplet = "GGA", weight = 29566 }, { triplet = "GGG", weight = 16019 } } },
         { letter = "H", codons = { { triplet = "CAT", weight = 36801 }, { triplet = "CAC", weight = 20528 } } },
         { letter = "R", codons = { { triplet = "CGT", weight = 16654 }, { triplet = "CGC", weight = 7037 }, { triplet = "CGA", weight = 8246 }, { triplet = "CGG", weight = 4843 }, { triplet = "AGA", weight = 55776 }, { triplet = "AGG", weight = 25020 } } },
         { letter = "N", codons = { { triplet = "AAT", weight = 96750 }, { triplet = "AAC", weight = 65451 } } },
      },
      translation_map = {},
      start_codon_table = {},
   },
   ["homo_sapiens"] = {
      start_codons = { "ATG" },
      stop_codons = { "TAA", "TAG", "TGA" },
      amino_acids = {
         { letter = "F", codons = { { triplet = "TTT", weight = 714298 }, { triplet = "TTC", weight = 824692 } } },
         { letter = "L", codons = { { triplet = "TTA", weight = 311881 }, { triplet = "TTG", weight = 525688 }, { triplet = "CTT", weight = 536515 }, { triplet = "CTC", weight = 796638 }, { triplet = "CTA", weight = 290751 }, { triplet = "CTG", weight = 1611801 } } },
         { letter = "I", codons = { { triplet = "ATT", weight = 650473 }, { triplet = "ATC", weight = 846466 }, { triplet = "ATA", weight = 304565 } } },
         { letter = "M", codons = { { triplet = "ATG", weight = 896005 } } },
         { letter = "V", codons = { { triplet = "GTT", weight = 448607 }, { triplet = "GTC", weight = 588138 }, { triplet = "GTA", weight = 287712 }, { triplet = "GTG", weight = 1143534 } } },
         { letter = "S", codons = { { triplet = "TCT", weight = 618711 }, { triplet = "TCC", weight = 718892 }, { triplet = "TCA", weight = 496448 }, { triplet = "TCG", weight = 179419 }, { triplet = "AGT", weight = 493429 }, { triplet = "AGC", weight = 791383 } } },
         { letter = "P", codons = { { triplet = "CCT", weight = 713233 }, { triplet = "CCC", weight = 804620 }, { triplet = "CCA", weight = 688038 }, { triplet = "CCG", weight = 281570 } } },
         { letter = "T", codons = { { triplet = "ACT", weight = 533609 }, { triplet = "ACC", weight = 768147 }, { triplet = "ACA", weight = 614523 }, { triplet = "ACG", weight = 246105 } } },
         { letter = "A", codons = { { triplet = "GCT", weight = 750096 }, { triplet = "GCC", weight = 1127679 }, { triplet = "GCA", weight = 643471 }, { triplet = "GCG", weight = 299495 } } },
         { letter = "Y", codons = { { triplet = "TAT", weight = 495699 }, { triplet = "TAC", weight = 622407 } } },
         { letter = "*", codons = { { triplet = "TAA", weight = 40285 }, { triplet = "TAG", weight = 32109 }, { triplet = "TGA", weight = 63237 } } },
         { letter = "H", codons = { { triplet = "CAT", weight = 441711 }, { triplet = "CAC", weight = 613713 } } },
         { letter = "Q", codons = { { triplet = "CAA", weight = 501911 }, { triplet = "CAG", weight = 1391973 } } },
         { letter = "N", codons = { { triplet = "AAT", weight = 689701 }, { triplet = "AAC", weight = 776603 } } },
         { letter = "K", codons = { { triplet = "AAA", weight = 993621 }, { triplet = "AAG", weight = 1295568 } } },
         { letter = "D", codons = { { triplet = "GAT", weight = 885429 }, { triplet = "GAC", weight = 1020595 } } },
         { letter = "E", codons = { { triplet = "GAA", weight = 1177632 }, { triplet = "GAG", weight = 1609975 } } },
         { letter = "C", codons = { { triplet = "TGT", weight = 430311 }, { triplet = "TGC", weight = 513028 } } },
         { letter = "W", codons = { { triplet = "TGG", weight = 535595 } } },
         { letter = "R", codons = { { triplet = "CGT", weight = 184609 }, { triplet = "CGC", weight = 423516 }, { triplet = "CGA", weight = 250760 }, { triplet = "CGG", weight = 464485 }, { triplet = "AGA", weight = 494682 }, { triplet = "AGG", weight = 486463 } } },
         { letter = "G", codons = { { triplet = "GGT", weight = 437126 }, { triplet = "GGC", weight = 903565 }, { triplet = "GGA", weight = 669873 }, { triplet = "GGG", weight = 669768 } } },
      },
      translation_map = {},
      start_codon_table = {},
   },
}


local function init_translation_maps(t)

   for _, aa in ipairs(t.amino_acids) do
      for _, cdn in ipairs(aa.codons) do
         t.translation_map[cdn.triplet] = aa.letter
      end
   end


   for _, cdn in ipairs(t.start_codons) do
      t.start_codon_table[cdn] = "M"
   end
   t.translate = translate
   t.optimize = optimize
   t.standardize_last_codon = standardize_last_codon

   return
end


for _, t in pairs(codon.default_tables) do
   init_translation_maps(t)
end













































local fix = {}













local function gc_content(sequence)
   sequence = string.upper(sequence)
   local guanine_count = 0
   local cytosine_count = 0


   for i = 1, #sequence do
      local base = string.sub(sequence, i, i)
      if base == "G" then
         guanine_count = guanine_count + 1
      elseif base == "C" then
         cytosine_count = cytosine_count + 1
      end
   end


   local guanine_and_cytosine_percentage = (guanine_count + cytosine_count) / #sequence
   return guanine_and_cytosine_percentage
end


function fix.remove_sequence(sequences_to_remove, reason)
   return function(sequence)
      local suggestions = {}
      local sequences_to_remove_for_reverse = {}

      for _, seq in ipairs(sequences_to_remove) do
         local reverse_complement_to_remove = transform.reverse_complement(seq)

         if reverse_complement_to_remove == seq then
            sequences_to_remove_for_reverse = { seq }
         else
            sequences_to_remove_for_reverse = { seq, reverse_complement_to_remove }
         end

         for _, site in ipairs(sequences_to_remove_for_reverse) do

            local start_pos = 1
            while true do
               local start_idx, end_idx = string.find(sequence, site, start_pos, true)
               if not start_idx then break end

               local codon_length = 3
               local position = math.floor((start_idx - 1) / codon_length) + 1
               table.insert(suggestions, {
                  start = position,
                  stop = math.floor((end_idx - 1) / codon_length) + 1,
                  bias = "NA",
                  quantity_fixes = 1,
                  suggestion_type = reason,
               })
               start_pos = start_idx + 1
            end
         end
      end
      return suggestions
   end
end


function fix.remove_repeat(repeat_len)
   return function(sequence)
      local suggestions = {}
      local kmers = {}


      for sequence_position = 1, #sequence - repeat_len do
         local kmer = string.sub(sequence, sequence_position, sequence_position + repeat_len - 1)
         local already_found_forward = kmers[kmer]
         local already_found_reverse = kmers[transform.reverse_complement(kmer)]
         kmers[kmer] = true

         if already_found_forward or already_found_reverse then
            local codon_length = 3
            local position = math.floor((sequence_position - 1) / codon_length) + 1
            local leftover = sequence_position % codon_length
            local end_position = math.floor((sequence_position + repeat_len - 1) / codon_length) + 1

            if leftover == 0 then
               table.insert(suggestions, {
                  start = position,
                  stop = end_position,
                  bias = "NA",
                  quantity_fixes = 1,
                  suggestion_type = "Repeat sequence",
               })
            else
               table.insert(suggestions, {
                  start = position,
                  stop = end_position - 1,
                  bias = "NA",
                  quantity_fixes = 1,
                  suggestion_type = "Repeat sequence",
               })
            end
            sequence_position = sequence_position + leftover
         end
      end
      return suggestions
   end
end





function fix.gc_content_fixer(upper_bound, lower_bound)
   return function(sequence)
      local suggestions = {}
      local gc_content_percentage = gc_content(sequence)
      local codon_length = 3

      if gc_content_percentage > upper_bound then
         local number_of_changes = math.floor((gc_content_percentage - upper_bound) * #sequence) + 1
         table.insert(suggestions, {
            start = 1,
            stop = math.floor((#sequence - 1) / codon_length) + 1,
            bias = "AT",
            quantity_fixes = number_of_changes,
            suggestion_type = "GcContent too high",
         })
      end

      if gc_content_percentage < lower_bound then
         local number_of_changes = math.floor((lower_bound - gc_content_percentage) * #sequence) + 1
         table.insert(suggestions, {
            start = 1,
            stop = math.floor((#sequence - 1) / codon_length) + 1,
            bias = "GC",
            quantity_fixes = number_of_changes,
            suggestion_type = "GcContent too low",
         })
      end

      return suggestions
   end
end



local function find_problems(sequence, problematic_sequence_funcs)
   local all_suggestions = {}


   for _, func in ipairs(problematic_sequence_funcs) do
      local suggestions = func(sequence)
      for _, suggestion in ipairs(suggestions) do
         table.insert(all_suggestions, suggestion)
      end
   end

   return all_suggestions
end

































function fix.cds(sequence, codon_table, problematic_sequence_funcs)
   local codon_length = 3
   if #sequence % codon_length ~= 0 then
      return "", {}, "this sequence isn't a complete CDS, please try to use a CDS without interrupted codons"
   end


   local historical_map = {}
   local weight_map = {}
   local na_bias_map = {}
   local gc_bias_map = {}
   local at_bias_map = {}


   local amino_acid_weight_table = {}
   for _, amino_acid in ipairs(codon_table.amino_acids) do
      local amino_acid_total = 0
      for _, cdn in ipairs(amino_acid.codons) do

         amino_acid_total = amino_acid_total + cdn.weight


         local codon_gc_count = 0
         for i = 1, #cdn.triplet do
            local base = string.sub(cdn.triplet, i, i)
            if base == "G" or base == "C" then
               codon_gc_count = codon_gc_count + 1
            end
         end

         for _, to_cdn in ipairs(amino_acid.codons) do
            if cdn.triplet ~= to_cdn.triplet then
               local to_codon_gc_count = 0
               for i = 1, #to_cdn.triplet do
                  local base = string.sub(to_cdn.triplet, i, i)
                  if base == "G" or base == "C" then
                     to_codon_gc_count = to_codon_gc_count + 1
                  end
               end

               if codon_gc_count > to_codon_gc_count then
                  if not at_bias_map[cdn.triplet] then at_bias_map[cdn.triplet] = {} end
                  table.insert(at_bias_map[cdn.triplet], to_cdn.triplet)
               elseif codon_gc_count < to_codon_gc_count then
                  if not gc_bias_map[cdn.triplet] then gc_bias_map[cdn.triplet] = {} end
                  table.insert(gc_bias_map[cdn.triplet], to_cdn.triplet)
               end
               if not na_bias_map[cdn.triplet] then na_bias_map[cdn.triplet] = {} end
               table.insert(na_bias_map[cdn.triplet], to_cdn.triplet)
            end
         end
      end


      if amino_acid_total == 0 then
         return "", {}, "incomplete codon table"
      end
      amino_acid_weight_table[amino_acid.letter] = amino_acid_total
   end


   for _, amino_acid in ipairs(codon_table.amino_acids) do
      for _, cdn in ipairs(amino_acid.codons) do
         local codon_weight_ratio = cdn.weight / amino_acid_weight_table[amino_acid.letter]
         local normalized_codon_weight = 100 * codon_weight_ratio
         weight_map[cdn.triplet] = normalized_codon_weight
      end
   end


   local sequence_length = 0


   for pos = 1, #sequence, codon_length do
      local cdn = string.sub(sequence, pos, pos + codon_length - 1)
      local position = math.floor((pos - 1) / codon_length) + 1
      if not historical_map[position] then historical_map[position] = {} end
      table.insert(historical_map[position], cdn)
      sequence_length = position
   end


   local function get_sequence(history)
      local result = {}
      for pos = 1, sequence_length do
         local codon_history = history[pos]
         table.insert(result, codon_history[#codon_history])
      end
      return table.concat(result)
   end

   local changes = {}
   local fix_iteration = 1

   while true do
      local suggestions = find_problems(sequence, problematic_sequence_funcs)


      if #suggestions == 0 then

         table.sort(changes, function(a, b)
            if a.step == b.step then
               return a.position < b.position
            end
            return a.step < b.step
         end)
         return sequence, changes, nil
      end

      for _, suggestion in ipairs(suggestions) do

         if suggestion.bias ~= "NA" and suggestion.bias ~= "GC" and suggestion.bias ~= "AT" then
            return sequence, {}, string.format("Invalid bias. Expected NA, GC, or AT, got %s", suggestion.bias)
         end


         local potential_changes = {}

         for position_selector = suggestion.start, suggestion.stop do
            if position_selector > sequence_length then break end

            local codon_list = historical_map[position_selector]
            local last_codon = codon_list[#codon_list]
            local unavailable_codons = {}

            for _, codon_site in ipairs(historical_map[position_selector]) do
               unavailable_codons[codon_site] = true
            end


            local bias_map
            if suggestion.bias == "NA" then
               bias_map = na_bias_map
            elseif suggestion.bias == "GC" then
               bias_map = gc_bias_map
            else
               bias_map = at_bias_map
            end


            if bias_map[last_codon] then
               for _, potential_codon in ipairs(bias_map[last_codon]) do
                  if not unavailable_codons[potential_codon] then
                     table.insert(potential_changes, {
                        position = position_selector,
                        step = fix_iteration,
                        from = last_codon,
                        to = potential_codon,
                        reason = suggestion.suggestion_type,
                     })
                  end
               end
            end
         end


         table.sort(potential_changes, function(a, b)
            return weight_map[a.to] > weight_map[b.to]
         end)


         local sorted_changes = {}
         local used_positions = {}

         for _, potential_change in ipairs(potential_changes) do
            if not used_positions[potential_change.position] then
               used_positions[potential_change.position] = true
               table.insert(sorted_changes, potential_change)
            end
         end

         if sorted_changes[1] ~= nil then

            if #sorted_changes < suggestion.quantity_fixes then
               return sequence, {}, string.format(
               "Too many fixes required. Number of potential fixes: %d, number of required fixes: %d",
               #potential_changes,
               suggestion.quantity_fixes)

            end


            for i = 1, suggestion.quantity_fixes do
               local target_change = sorted_changes[i]
               table.insert(historical_map[target_change.position], target_change.to)
               table.insert(changes, target_change)
               sequence = get_sequence(historical_map)
            end
         end
      end
      fix_iteration = fix_iteration + 1


      if fix_iteration > 100 then
         return sequence, changes, "maximum iterations reached"
      end
   end
end




function fix.cds_simple(sequence, codon_table, sequences_to_remove)
   local functions = {}


   table.insert(functions, fix.remove_sequence({ "AAAAAAAA", "GGGGGGGG" }, "Homopolymers"))


   table.insert(functions, fix.remove_sequence(sequences_to_remove, "Removal requested by user"))


   table.insert(functions, fix.remove_repeat(18))


   table.insert(functions, fix.gc_content_fixer(0.80, 0.20))

   return fix.cds(sequence, codon_table, functions)
end





return {
   rng = rng,
   hash = hash,
   transform = transform,
   align = align,
   mash = mash,
   seqhash = seqhash,
   primers = primers,
   orthoprimers = orthoprimers,
   pcr = pcr,
   bio = bio,
   fasta = fasta,
   fastq = fastq,
   pileup = pileup,
   sam = sam,
   slow5 = slow5,
   genbank = genbank,
   fragment_frequencies = fragment_frequencies,
   fragment = fragment,
   clone = clone,
   codon = codon,
   fix = fix,
}
