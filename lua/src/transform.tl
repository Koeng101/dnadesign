--[[ Module transform
Module transform provides functions for transforming sequences.
]]
local record transform
    -- DNA/RNA sequence transformations
    reverse_complement: function(sequence: string): string
    complement: function(sequence: string): string
    reverse: function(sequence: string): string
    complement_base: function(base_pair: string): string
    
    -- RNA-specific transformations
    reverse_complement_rna: function(sequence: string): string
    complement_rna: function(sequence: string): string 
    complement_base_rna: function(base_pair: string): string

	-- Windowing functions
	get_window: function(target_sequence: string, left_flank: string, right_flank: string, circular: boolean): string
end

--[[ complementTable provides 1:1 mapping between bases and their complements ]]
local complement_table: {string:string} = {
    ['A'] = 'T', ['B'] = 'V', ['C'] = 'G', ['D'] = 'H',
    ['G'] = 'C', ['H'] = 'D', ['K'] = 'M', ['M'] = 'K',
    ['N'] = 'N', ['R'] = 'Y', ['S'] = 'S', ['T'] = 'A',
    ['V'] = 'B', ['W'] = 'W', ['Y'] = 'R',
    ['a'] = 't', ['b'] = 'v', ['c'] = 'g', ['d'] = 'h',
    ['g'] = 'c', ['h'] = 'd', ['k'] = 'm', ['m'] = 'k',
    ['n'] = 'n', ['r'] = 'y', ['s'] = 's', ['t'] = 'a',
    ['v'] = 'b', ['w'] = 'w', ['y'] = 'r'
}

--[[
complementTable provides 1:1 mapping between bases and their complements
see https://www.dnabaser.com/articles/IUPAC%20ambiguity%20codes.html
]]
local complement_table_rna: {string:string} = {
    ['A'] = 'U', ['B'] = 'V', ['C'] = 'G', ['D'] = 'H',
    ['G'] = 'C', ['H'] = 'D', ['K'] = 'M', ['M'] = 'K',
    ['N'] = 'N', ['R'] = 'Y', ['S'] = 'S', ['U'] = 'A',
    ['V'] = 'B', ['W'] = 'W', ['Y'] = 'R', ['X'] = 'X',
    ['a'] = 'u', ['b'] = 'v', ['c'] = 'g', ['d'] = 'h',
    ['g'] = 'c', ['h'] = 'd', ['k'] = 'm', ['m'] = 'k',
   ['n'] = 'n', ['r'] = 'y', ['s'] = 's', ['u'] = 'a',
    ['v'] = 'b', ['w'] = 'w', ['y'] = 'r', ['x'] = 'x'
}

--[[
ReverseComplement returns the reversed complement of sequence.
It is the equivalent of calling

    revComplement := Reverse(Complement(sequence))

This function expects byte characters in the range a-z and A-Z and
will not check for non-byte characters, i.e. utf-8 encoding.
]]
function transform.reverse_complement(sequence: string): string
  local sequence_length = #sequence
  local new_sequence: {string} = {}
  
  for i = 1, sequence_length do
    local char = sequence:sub(sequence_length - i + 1, sequence_length - i + 1)
    new_sequence[i] = complement_table[char]
  end

  return table.concat(new_sequence)
end

--[[
Complement returns the complement of sequence. In [DNA] each nucleotide
(A, T, C or G) is has a deterministic pair. A is paired with T and
C is paired with G. The complement of a sequence is formed by exchanging
these letters for their pair:

    'A' becomes 'T'
    'T' becomes 'A'
    'C' becomes 'G'
    'G' becomes 'C'

This function expects byte characters in the range a-z and A-Z and
will not check for non-byte characters, i.e. utf-8 encoding.

[DNA]: https://en.wikipedia.org/wiki/DNA
]]
function transform.complement(sequence: string): string
  local sequence_length = #sequence
  local new_sequence: {string} = {}

  for i = 1, sequence_length do
    local char = sequence:sub(i, i)
    new_sequence[i] = complement_table[char]
  end

  return table.concat(new_sequence)
end

--[[
Reverse returns the reverse of sequence. It performs a basic
string reversal by working on the bytes.

This function expects byte characters in the range a-z and A-Z and
will not check for non-byte character, i.e. utf-8 encoding.
]]
function transform.reverse(sequence: string): string
  local sequence_length = #sequence
  local new_sequence: {string} = {}

  for i = 1, sequence_length do
    new_sequence[i] = sequence:sub(sequence_length - i + 1, sequence_length - i + 1)
  end

  return table.concat(new_sequence)
end

--[[
ComplementBase accepts a base pair and returns its complement base pair. See Complement.

This function expects byte characters in the range a-z and A-Z and
will return a space ' ' (U+0020) for characters that are not matched
to any known base. This is subject to change.
]]
function transform.complement_base(base_pair: string): string
  local complement = complement_table[base_pair]
  if complement == nil then
    return " "
  end
  return complement
end

--[[
ReverseComplementRNA returns the reversed complement of sequence.
It is the equivalent of calling

    revComplement := Reverse(ComplementRNA(sequence))

This function expects byte characters in the range a-z and A-Z and
will not check for non-byte characters, i.e. utf-8 encoding.
]]
function transform.reverse_complement_rna(sequence: string): string
  local sequence_length = #sequence
  local new_sequence: {string} = {}
  
  for i = 1, sequence_length do
    local char = sequence:sub(sequence_length - i + 1, sequence_length - i + 1)
    new_sequence[i] = complement_table_rna[char]
  end

  return table.concat(new_sequence)
end

--[[
ComplementRNA returns the complement of sequence. In [RNA] each nucleotide
(A, U, C or G) is has a deterministic pair. A is paired with U and
C is paired with G. The complement of a sequence is formed by exchanging
these letters for their pair:

    'A' becomes 'U'
    'U' becomes 'A'
    'C' becomes 'G'
    'G' becomes 'C'

This function expects byte characters in the range a-z and A-Z and
will not check for non-byte characters, i.e. utf-8 encoding.

[RNA]: https://en.wikipedia.org/wiki/RNA
]]
function transform.complement_rna(sequence: string): string
  local sequence_length = #sequence
  local new_sequence: {string} = {}

  for i = 1, sequence_length do
    local char = sequence:sub(i, i)
    new_sequence[i] = complement_table_rna[char]
  end

  return table.concat(new_sequence)
end

--[[
ComplementBaseRNA accepts a RNA base pair and returns its complement base
pair. See Complement.

This function expects byte characters in the range a-z and A-Z and
will return a space ' ' (U+0020) for characters that are not matched
to any known base. This is subject to change.
]]
function transform.complement_base_rna(base_pair: string): string
  local complement = complement_table_rna[base_pair]
  if complement == nil then
    return " "
  end
  return complement
end

--[[
get_window takes in a target sequence and two flanks, and if it exists, returns
the window between that sequence. It checks both the given sequence and its
reverse complement.
--]]
function transform.get_window(target_sequence: string, left_flank: string, right_flank: string, circular: boolean): string
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
    
    -- Not found
    return nil
end

return transform
