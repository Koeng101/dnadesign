--[[ Module clone
Module clone provides functions for cloning DNA sequences.

Since 1973, the most common way to make recombinant DNA has been restriction
enzyme cloning (though lately, homologous recombination based methods like
Gibson assembly have attracted a lot of use). The cloning functions here allow
for simulation of restriction enzyme cloning.

For a historical review leading up to the discovery:
https://doi.org/10.1073/pnas.1313397110

The idea of restriction enzyme cloning is that you can cut DNA at specific
locations with restriction enzyme and then glue them back together in different
patterns using ligase. The final product is (99.9% of the time) a circular plasmid
that you can transform into a bacterial cell for propagation.

While simulation is simple for simple cases, there are a lot of edge cases to handle, for example:
  - Which input sequences are circular? How do we handle their rotations?
  - Is the enzyme that is cutting directional? How do we handle that directionality?
  - Are there multiple possible outputs of our ligation reaction? For example, ligations may be
    able to create a "library" of plasmids, in which there are millions of valid combinations.
  - How do we handle sequences that get ligated in multiple orientations?

These cloning functions handle all those problems so that they appear simple to the end user.

In particular, there is a focus here on GoldenGate Assembly:
https://en.wikipedia.org/wiki/Golden_Gate_Cloning
https://www.neb.com/applications/cloning-and-synthetic-biology/dna-assembly-and-cloning/golden-gate-assembly

GoldenGate is a particular kind of restriction enzyme cloning reaction that you can do
in a single tube and that is extraordinarily efficient (up to 50 parts) and is popular
for new modular DNA part toolkits. Users can easily simulate GoldenGate assembly reactions
with just their input fragments + the enzyme name.

Unlike many other GoldenGate simulators, we support simulating GoldenGate with
methylated DNA sequences, which are represented as lowercased sequences in user
inputted sequences. Normally, this can be turned off, but can be used in the
special case of recursive GoldenGate reactions.

Let's build some DNA!

PS: We do NOT (yet) handle restriction enzymes which recognize one site but cut
in multiple places (Type IIG enzymes) such as BcgI.
]]

-- Main interface
local record clone
   -- Core cloning functions
   cut_with_enzyme_by_name: function(part: Part, directional: boolean, name: string, methylated: boolean): ({Fragment}, string)
   cut_with_enzyme: function(part: Part, directional: boolean, enzyme: Enzyme, methylated: boolean): {Fragment}
   ligate: function(fragments: {Fragment}, circular: boolean): (string, {integer}, string)
   golden_gate: function(sequences: {Part}, cutting_enzyme: Enzyme, methylated: boolean): (string, {integer}, string)

   default_enzymes: {string:Enzyme}

   -- Kmer analysis functions for analyzing ligation junctions
   find_kmer_overlaps: function(fragments: {Fragment}, ligation_product: string, ligation_pattern: {integer}, kmer_size: integer): ({KmerOverlap}, string)
   find_kmers: function(kmer_overlaps: {KmerOverlap}, sequence: string): {KmerOverlap}

   Overhang: Overhang
end

-- Core type definitions
local record Part
   -- Part is a simple struct that can carry a circular or linear DNA sequence.
   -- In the field of synthetic biology, the term "DNA Part" was popularized by
   -- the iGEM competition http://parts.igem.org/Main_Page , so we use that term
   -- here.
   sequence: string
   circular: boolean
end

local record Overhang
   -- Overhang is a struct that represents the ends of a linearized sequence where Enzymes had cut.
   length: integer
   position: integer
   forward: boolean
   recognition_site_plus_skip_length: integer
end

local record Fragment
   -- Fragment is a struct that represents linear DNA sequences with sticky ends.
   sequence: string
   forward_overhang: string
   reverse_overhang: string
end

local record Enzyme
   -- Enzyme is a struct that represents restriction enzymes.
   name: string
   pattern_for: string  -- Lua pattern instead of regexp
   pattern_rev: string  -- Lua pattern instead of regexp
   skip: integer 
   overhead_length: integer
   recognition_site: string
end

local record KmerOverlap
   -- KmerOverlap represents the overlap between two fragments indicative of a
   -- ligation event.
   kmer: string
   fragment1: Fragment
   fragment2: Fragment
end

-- Default enzymes with Lua patterns instead of regexp
local default_enzymes: {string:Enzyme} = {
   BsaI = {
      name = "BsaI",
      pattern_for = "GGTCTC",  -- Will need to carefully handle pattern matching
      pattern_rev = "GAGACC",
      skip = 1,
      overhead_length = 4,
      recognition_site = "GGTCTC"
   },
   BbsI = {
      name = "BbsI",
      pattern_for = "GAAGAC",
      pattern_rev = "GTCTTC",
      skip = 2,
      overhead_length = 4,
      recognition_site = "GAAGAC"
   },
   BtgZI = {
      name = "BtgZI",
      pattern_for = "GCGATG",
      pattern_rev = "CATCGC",
      skip = 10,
      overhead_length = 4,
      recognition_site = "GCGATG"
   },
   PaqCI = {
      name = "PaqCI",
      pattern_for = "CACCTGC",
      pattern_rev = "GCAGGTG",
      skip = 4,
      overhead_length = 4,
      recognition_site = "CACCTGC"
   },
   BsmBI = {
      name = "BsmBI",
      pattern_for = "CGTCTC",
      pattern_rev = "GAGACG",
      skip = 1,
      overhead_length = 4,
      recognition_site = "CGTCTC"
   }
}

clone.default_enzymes = default_enzymes

--[[**************************************************************************

Base cloning functions begin here.

**************************************************************************]]

-- Regular utility functions
local function sort_overhangs(overhangs: {Overhang})
   table.sort(overhangs, function(a: Overhang, b: Overhang): boolean
      return a.position < b.position
   end)
end

-- cut_with_enzyme_by_name cuts a given sequence with an enzyme represented by the
-- enzyme's name. It is a convenience wrapper around cut_with_enzyme that
-- allows us to specify the enzyme by name. Set methylated flag to true if
-- there is lowercase methylated DNA as part of the sequence.
function clone.cut_with_enzyme_by_name(part: Part, directional: boolean, name: string, methylated: boolean): ({Fragment}, string)
   -- Get the enzyme from the enzyme map
   local enzyme = default_enzymes[name]
   if not enzyme then
      return {}, "enzyme not found"
   end
   -- Cut the sequence with the enzyme
   return clone.cut_with_enzyme(part, directional, enzyme, methylated), nil
end

-- cut_with_enzyme cuts a given sequence with an enzyme represented by an Enzyme struct.
-- If there is methylated parts of the target DNA, set the "methylated" flag to
-- true and lowercase ONLY methylated DNA.
function clone.cut_with_enzyme(part: Part, directional: boolean, enzyme: Enzyme, methylated: boolean): {Fragment}
   local fragments: {Fragment} = {}
   local fragment_sequences: {string} = {}

   -- Setup circular sequences
   local sequence = part.sequence
   if part.circular then
      sequence = sequence .. sequence
   end

   -- If unmethylated, set everything to uppercase so that the enzyme pattern
   -- works on all the sequence
   if not methylated then
      sequence = string.upper(sequence)
   end

   -- Check for palindromes by comparing forward and reverse sequences
   local palindromic = enzyme.recognition_site == string.reverse(enzyme.recognition_site)

   -- Find and define overhangs
   local overhangs: {Overhang} = {}
   local forward_overhangs: {Overhang} = {}
   local reverse_overhangs: {Overhang} = {}

   -- Find forward cuts
   local pos = 1
   while true do
      local start = string.find(sequence, enzyme.pattern_for, pos, true)
      if not start then break end
      table.insert(forward_overhangs, {
         length = enzyme.overhead_length,
         position = start + #enzyme.pattern_for + enzyme.skip,
         forward = true,
         recognition_site_plus_skip_length = #enzyme.recognition_site + enzyme.skip
      })
      pos = start + 1
   end

   -- Find reverse cuts if not palindromic
   if not palindromic then
      pos = 1
      while true do
         local start = string.find(sequence, enzyme.pattern_rev, pos, true)
         if not start then break end
         table.insert(reverse_overhangs, {
            length = enzyme.overhead_length,
            position = start - enzyme.skip,
            forward = false,
            recognition_site_plus_skip_length = #enzyme.recognition_site + enzyme.skip
         })
         pos = start + 1
      end
   end

   -- Remove overhangs that would extend beyond sequence length for linear sequences
   for _, overhang_set in ipairs({forward_overhangs, reverse_overhangs}) do
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

   -- Sort overhangs by position
   sort_overhangs(overhangs)

   -- Convert Overhangs into Fragments
   -- Single cut in linear sequence case:
	if #overhangs == 1 and not directional and not part.circular then
	   local fragment_sequence1: string
	   local fragment_sequence2: string
	   local overhang_sequence: string
	
	   if #forward_overhangs > 0 then
	      -- Change from original implementation:
	      fragment_sequence1 = sequence:sub(overhangs[1].position + overhangs[1].length) -- This is correct
	      fragment_sequence2 = sequence:sub(1, overhangs[1].position - 1) -- Need -1 for inclusive end
	      overhang_sequence = sequence:sub(overhangs[1].position,
	                                     overhangs[1].position + overhangs[1].length - 1) -- Need -1 for inclusive end
	      table.insert(fragments, {
	         sequence = fragment_sequence1,
	         forward_overhang = overhang_sequence,
	         reverse_overhang = ""
	      })
	      table.insert(fragments, {
	         sequence = fragment_sequence2,
	         forward_overhang = "",
	         reverse_overhang = overhang_sequence
	      })
	   else
	      fragment_sequence1 = sequence:sub(overhangs[1].position)
	      fragment_sequence2 = sequence:sub(1, overhangs[1].position - overhangs[1].length - 1) -- Need -1 for inclusive end
	      overhang_sequence = sequence:sub(overhangs[1].position - overhangs[1].length,
	                                     overhangs[1].position - 1) -- Need -1 for inclusive end
	      table.insert(fragments, {
	         sequence = fragment_sequence2,
	         forward_overhang = "",
	         reverse_overhang = overhang_sequence
	      })
	      table.insert(fragments, {
	         sequence = fragment_sequence1,
	         forward_overhang = overhang_sequence,
	         reverse_overhang = ""
	      })
	   end
	   return fragments
	end
	
	-- Single cut in circular sequence case:
	if #overhangs == 2 and not directional and part.circular then
	   local fragment_sequence1 = sequence:sub(overhangs[1].position + overhangs[1].length,
	                                         #part.sequence)
	   local fragment_sequence2 = sequence:sub(1, overhangs[1].position - 1) -- Need -1 for inclusive end
	   local fragment_sequence = fragment_sequence1 .. fragment_sequence2
	   local overhang_sequence = sequence:sub(overhangs[1].position,
	                                        overhangs[1].position + overhangs[1].length - 1) -- Need -1 for inclusive end
	   table.insert(fragments, {
	      sequence = fragment_sequence,
	      forward_overhang = overhang_sequence,
	      reverse_overhang = overhang_sequence
	   })
	   return fragments
	end
	
	-- Multiple cuts case:
	if #overhangs > 1 then
	   for i = 1, #overhangs - 1 do
	      local current_overhang = overhangs[i]
	      local next_overhang = overhangs[i + 1]
	      if directional and not palindromic then
	         if current_overhang.forward and not next_overhang.forward then
	            table.insert(fragment_sequences,
	                        sequence:sub(current_overhang.position, next_overhang.position - 1)) -- Need -1 for inclusive end
	         end
	         if next_overhang.position - next_overhang.recognition_site_plus_skip_length > #part.sequence then
	            break
	         end
	      else
	         table.insert(fragment_sequences,
	                     sequence:sub(current_overhang.position, next_overhang.position - 1)) -- Need -1 for inclusive end
	         if next_overhang.position - next_overhang.recognition_site_plus_skip_length > #part.sequence then
	            break
	         end
	      end
	   end
	
	   -- Convert fragment sequences to fragments
	   for _, fragment_sequence in ipairs(fragment_sequences) do
	      if #fragment_sequence > 8 then -- Minimum length check (8bp)
	         local fragment_seq = fragment_sequence:sub(enzyme.overhead_length + 1,
	                                                  #fragment_sequence - enzyme.overhead_length)
	         local forward_overhang = fragment_sequence:sub(1, enzyme.overhead_length)
	         local reverse_overhang = fragment_sequence:sub(#fragment_sequence - enzyme.overhead_length + 1)
	         table.insert(fragments, {
	            sequence = fragment_seq,
	            forward_overhang = forward_overhang,
	            reverse_overhang = reverse_overhang
	         })
	      end
	   end
	end

   return fragments
end

-- Import required functions from other modules
local transform = require("transform")

--[[
ligate simulates ligations. It assumes that fragments can only be ligated
in a single way (no 2 fragments with the same overhangs), and also assumes
the first fragment WILL be used in the ligation reaction. This function
is a massive simplification of the original ligation code which can do more.
If this does not fulfill your needs, please leave an issue in git.
]]
function clone.ligate(fragments: {Fragment}, circular: boolean): (string, {integer}, string)
   if #fragments == 0 then
      return "", {}, "no fragments to ligate"
   end

   -- Ligation pattern is used in downstream functions for analyzing
   -- ligation patterns.
   local ligation_pattern: {integer} = {1} -- first fragment is the first ligation site

   local final_fragment = {
      sequence = fragments[1].sequence,
      forward_overhang = fragments[1].forward_overhang,
      reverse_overhang = fragments[1].reverse_overhang
   }
   
   local used: {integer:boolean} = {[1] = true}
   local match_found = true

   -- iterate until no fragments are found
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

   -- attempt circularization
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

--[[
golden_gate simulates a GoldenGate cloning reaction. As of right now, we only
support BsaI, BbsI, BtgZI, and BsmBI. Set methylated flag to true if there
is lowercase methylated DNA as part of the sequence.
]]
function clone.golden_gate(sequences: {Part}, cutting_enzyme: Enzyme, methylated: boolean): (string, {integer}, string)
   local fragments: {Fragment} = {}
   
   for _, sequence in ipairs(sequences) do
      local new_fragments = clone.cut_with_enzyme(sequence, true, cutting_enzyme, methylated)
      for _, frag in ipairs(new_fragments) do
         table.insert(fragments, frag)
      end
   end
   
   return clone.ligate(fragments, true)
end

--[[
find_kmer_overlaps finds kmerOverlaps from ligation reactions. It can be used
with find_kmers to find ligation events in sequence data.

Here is the problem: You have sequenced a GoldenGate or ligation you've run
to assemble some DNA. How do you quantify the efficiency of the ligation
from this raw data? You may want to do this to make sure your ligations are
working properly in a way that simple controls wouldn't: you can *directly*
observe single molecules of interest that are ligated and ones which aren't.

How would you do this? The simplest version, which we implement here, is to
check whether or not kmers indicative of ligation exist within the sequenced
fragments. This is simple and computationally inexpensive.
]]
-- Helper function to get standardized kmer
local function standardize_kmer(kmer: string): string
   -- Convert to uppercase
   kmer = string.upper(kmer)
   -- Get reverse complement
   local rev_comp = transform.reverse_complement(kmer)
   -- Return the lexicographically smaller one
   if kmer < rev_comp then
      return kmer
   end
   return rev_comp
end

function clone.find_kmer_overlaps(
   fragments: {Fragment}, 
   ligation_product: string, 
   ligation_pattern: {integer}, 
   kmer_size: integer
): ({KmerOverlap}, string)
   if #fragments < 2 then
      return {}, "need at least two fragments to find overlaps"
   end

   local bp_from_each = math.floor((kmer_size - 4) / 2)
   if bp_from_each < 4 then
      return {}, "need at least a kmer of 12"
   end

   local ligation = ligation_product .. ligation_product
   local position = 1
   local kmer_overlaps: {KmerOverlap} = {}

   for i = 1, #ligation_pattern do
      local frag1 = fragments[ligation_pattern[i]]
      local frag2: Fragment
      if i == #ligation_pattern then
         frag2 = fragments[ligation_pattern[1]]
      else
         frag2 = fragments[ligation_pattern[i + 1]]
      end

      position = position + #frag1.forward_overhang + #frag1.sequence
      local kmer = ligation:sub(
         position - bp_from_each,
         position + 3 + bp_from_each
      )
      -- Standardize kmer orientation
      kmer = standardize_kmer(kmer)
      
      table.insert(kmer_overlaps, {
         kmer = kmer,
         fragment1 = frag1,
         fragment2 = frag2
      })
   end

   return kmer_overlaps, nil
end

--[[
find_kmers finds kmers indicative of ligation events within sequence data.
If you need to just search raw sequence and not fastq reads, you can
simply pass in the sequence directly.
]]
function clone.find_kmers(kmer_overlaps: {KmerOverlap}, sequence: string): {KmerOverlap}
   local output_kmer_overlaps: {KmerOverlap} = {}
   sequence = string.upper(sequence)

   for _, kmer_overlap in ipairs(kmer_overlaps) do
      if string.find(sequence, string.upper(kmer_overlap.kmer), 1, true) or
         string.find(sequence, string.upper(transform.reverse_complement(kmer_overlap.kmer)), 1, true) then
         table.insert(output_kmer_overlaps, {
            kmer = kmer_overlap.kmer,
            fragment1 = kmer_overlap.fragment1,
            fragment2 = kmer_overlap.fragment2
         })
      end
   end

   return output_kmer_overlaps
end

return clone
