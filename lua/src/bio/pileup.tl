--[[ Module pileup
Module pileup contains pileup parsers and writers.

The pileup format is a text-based bioinformatics format to summarize aligned
reads against a reference sequence. In comparison to simply getting a consensus
sequence from sequencing data, pileup files can contain more context about the
mutations in a sequencing run, which is especially useful when analyzing
plasmid sequencing data from Nanopore sequencing runs.

Pileup files are basically tsv files with 6 columns: Sequence Identifier, Position,
Reference Base, Read Count, Read Results, and Quality. An example from
wikipedia (https://en.wikipedia.org/wiki/Pileup_format) is shown below:

    ```
    seq1    272    T    24    ,.$.....,,.,.,...,,,.,..^+.    <<<+;<<<<<<<<<<<=<;<;7<&
    seq1    273    T    23    ,.....,,.,.,...,,,.,..A    <<<;<<<<<<<<<3<=<<<;<<+
    seq1    274    T    23    ,.$....,,.,.,...,,,.,... 7<7;<;<<<<<<<<<=<;<;<<6
    seq1    275    A    23    ,$....,,.,.,...,,,.,...^l.    <+;9*<<<<<<<<<=<<:;<<<<
    seq1    276    G    22    ...T,,.,.,...,,,.,....    33;+<<7=7<<7<&<<1;<<6<
    seq1    277    T    22    ....,,.,.,.C.,,,.,..G.    +7<;<<<<<<<&<=<<:;<<&<
    seq1    278    G    23    ....,,.,.,...,,,.,....^k.    %38*<<;<7<<7<=<<<;<<<<<
    seq1    279    C    23    A..T,,.,.,...,,,.,.....    75&<<<<<<<<<=<<<9<<:<<<
    ```

    1. Sequence Identifier: The sequence identifier of the reference sequence
    2. Position: Position of row in the reference sequence (indexed at 1)
    3. Reference Base: Base pair in reference sequence
    4. Read Count: Number of aligned reads to this particular base pair
    5. Read Results: The resultant alignments
    6. Quality: Phred quality scores associated with each base

This module provides a parser and writer for working with pileup files.

Wikipedia reference: https://en.wikipedia.org/wiki/Pileup_format
]]

local bio = require("bio")

local record pileup
   PileupParser: PileupParser
   PileupLine: PileupLine
   PileupHeader: PileupHeader
   new_parser: function(reader: bio.Reader, max_line_size: integer): PileupParser
   
   -- Mutation types
   MutationType: MutationType
   Mutation: Mutation
   call_mutations: function(read_results: {string}, reference_base: string, minimal_ratio: number): Mutation
end

-- PileupLine represents a single position in a pileup file
local record PileupLine is bio.Data
   where self.format == "PILEUP"

   sequence: string
   position: integer
   reference_base: string  
   read_count: integer
   read_results: {string}
   quality: string
   format: bio.Format
   to_string: function(self: PileupLine): string
   write: function(self: PileupLine, writer: bio.Writer): (boolean, string)
end

-- PileupHeader is a blank record, needed for compatibility with bio parsers
local record PileupHeader is bio.Header
   where self.format == "PILEUP"
   
   format: bio.Format
   to_string: function(self: PileupHeader): string
   write: function(self: PileupHeader, writer: bio.Writer): (boolean, string)
end

-- PileupParser implements an efficient buffered parser for PILEUP format
local record PileupParser is bio.Parser
   where self.format == "PILEUP"

   reader: bio.BufferedReader
   line: integer
   at_eof: boolean

   -- Implement bio.Parser interface
   header: function(self: PileupParser): (bio.Header, string)
   next: function(self: PileupParser): (PileupLine, string)
   get_format: function(self: PileupParser): bio.Format
end

-- MutationType represents different kinds of mutations that can be detected
local enum MutationType
   "no_mutation"
   "point"
   "point_indel" 
   "indel"
   "insertion"
   "noisy"
end

-- Mutation represents details about a detected mutation
local record Mutation
   type: MutationType
   from: string
   to: string
   length: integer -- Only for Indel and Insertions
   total_correct: integer
   total_mutated: integer
   total_aligned: integer
end

-- PileupLine methods
function PileupLine:to_string(): string
   local result = string.format("%s\t%d\t%s\t%d\t%s\t%s\n",
      self.sequence,
      self.position,
      self.reference_base,
      self.read_count,
      table.concat(self.read_results),
      self.quality)
   return result
end

function PileupLine:write(writer: bio.Writer): (boolean, string)
   local ok, err = writer:write(self:to_string())
   if not ok then
      return false, err
   end
   return true, nil
end

-- PileupHeader methods
function PileupHeader:to_string(): string
   return ""  -- Pileup has no header
end

function PileupHeader:write(_: bio.Writer): (boolean, string)
   return true, nil  -- Pileup has no header
end

-- Helper functions
local function new_pileup_line(sequence: string, position: integer, reference_base: string, 
                              read_count: integer, read_results: {string}, quality: string): PileupLine
   local pileup_line: PileupLine = {
      sequence = sequence,
      position = position,
      reference_base = reference_base,
      read_count = read_count,
      read_results = read_results,
      quality = quality,
      format = "PILEUP",
      to_string = PileupLine.to_string,
      write = PileupLine.write
   }
   return pileup_line
end

local function new_pileup_header(): PileupHeader
   local pileup_head: PileupHeader = {
      format = "PILEUP",
      to_string = PileupHeader.to_string,
      write = PileupHeader.write
   }
   return pileup_head
end

-- Create a new PileupParser
function pileup.new_parser(reader: bio.Reader, max_line_size: integer): PileupParser
   local buffered = bio.new_buffered_reader(reader, max_line_size)
   local parser: PileupParser = {
      reader = buffered,
      line = 0,
      at_eof = false,
      header = nil,
      next = nil,
      get_format = nil
   }

   function parser:header(): (bio.Header, string)
      return new_pileup_header(), nil
   end

   function parser:get_format(): bio.Format
      return "PILEUP"
   end

   function parser:next(): (PileupLine, string)
      if self.at_eof then
         return nil, "EOF"
      end

      -- Parse out a single line
      local line, err = self.reader:read_line()
      if err then
         if err == "EOF" then
            self.at_eof = true
            return nil, "EOF"
         end
         return nil, err
      end
      
      self.line = self.line + 1

      -- Skip empty lines
      if #line == 0 then
         if self.at_eof then
            return nil, "EOF"
         end
         return self:next()
      end

      -- Check that there are 6 values, as defined by the pileup format
      local values = {}
      for value in line:gmatch("[^\t]+") do
         table.insert(values, value)
      end
      if #values ~= 6 then
         return nil, string.format("Error on line %d: Got %d values, expected 6.", self.line, #values)
      end

      -- Convert Position and ReadCount to integers
      local position = tonumber(values[2]) as integer
      if not position then
         return nil, string.format("Error on line %d: Invalid position value", self.line)
      end
      
      local read_count = tonumber(values[4]) as integer
      if not read_count then
         return nil, string.format("Error on line %d: Invalid read count value", self.line)
      end

      -- Parse ReadResults
      local read_results = {}
      local results_string = values[5]
      local i = 1
      while i <= #results_string do
         local result_char = results_string:sub(i,i)
         if result_char == " " then
            i = i + 1
         elseif result_char == "^" then
            -- Start of read marker
            if i + 2 <= #results_string then
               table.insert(read_results, results_string:sub(i,i+2))
               i = i + 3
            else
               return nil, string.format("Error on line %d: Invalid read start marker", self.line)
            end
         elseif result_char == "$" then
            -- End of read marker
            if #read_results > 0 then
               read_results[#read_results] = read_results[#read_results] .. "$"
            end
            i = i + 1
         elseif result_char:match("[.,*ATGCNatgcn]") then
            -- Simple base call
            table.insert(read_results, result_char)
            i = i + 1
         elseif result_char:match("[+-]") then
            -- Indel marker
            local indel_size = ""
            local j = i + 1
            while j <= #results_string and results_string:sub(j,j):match("%d") do
               indel_size = indel_size .. results_string:sub(j,j)
               j = j + 1
            end
            
            if #indel_size == 0 then
               return nil, string.format("Error on line %d: Invalid indel format", self.line)
            end
            
            local size = tonumber(indel_size) as integer
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
         values[1],           -- sequence
         position,            -- position
         values[3],           -- reference_base
         read_count,          -- read_count
         read_results,        -- read_results
         values[6]            -- quality
      ), nil
   end

   return parser
end

--[[
Call mutations analyzes read results to detect mutations There are a few
problems with this function, (such as dual mutations in a single line may get
called as point mutation for one of the mutations rather than noisy), but it
seems to work well enough in practice.
--]]
function pileup.call_mutations(read_results: {string}, reference_base: string, minimal_ratio: number): Mutation
   -- Count occurrences of each read result
   local reads: {string:integer} = {}
   for _, read_result in ipairs(read_results) do
      if #read_result == 1 then
         -- Simple point mutations and correct sequences
         reads[read_result] = (reads[read_result] or 0) + 1
      else
         -- For indels/insertions or complex patterns, store entire string
         reads[read_result] = (reads[read_result] or 0) + 1
      end
   end

   -- Count correct alignments
   local no_mutation = (reads["."] or 0) + (reads[","] or 0)

   -- Helper function to check mutation ratio
   local function has_sufficient_ratio(mutation_count: integer): boolean
      if #read_results == 0 then
         return false
      end
      return (mutation_count / #read_results) > minimal_ratio
   end

   -- Check point mutations first
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
         total_aligned = #read_results
      }
   elseif has_sufficient_ratio(t_mutation) then
      return {
         type = "point",
         from = reference_base,
         to = "T",
         length = 0,
         total_correct = no_mutation,
         total_mutated = t_mutation,
         total_aligned = #read_results
      }
   elseif has_sufficient_ratio(g_mutation) then
      return {
         type = "point",
         from = reference_base,
         to = "G",
         length = 0,
         total_correct = no_mutation,
         total_mutated = g_mutation,
         total_aligned = #read_results
      }
   elseif has_sufficient_ratio(c_mutation) then
      return {
         type = "point",
         from = reference_base,
         to = "C",
         length = 0,
         total_correct = no_mutation,
         total_mutated = c_mutation,
         total_aligned = #read_results
      }
   elseif has_sufficient_ratio(point_indel) then
      return {
         type = "point_indel",
         from = reference_base,
         to = "*",
         length = 0,
         total_correct = no_mutation,
         total_mutated = point_indel,
         total_aligned = #read_results
      }
   end

   -- Check indels (both insertions and deletions)
   for read_result, count in pairs(reads) do
      if #read_result > 1 then
         local first_char = read_result:sub(1,1)
         if first_char == "-" or first_char == "+" then
            if has_sufficient_ratio(count) then
               -- Extract the length from the indel notation
               local length_str = read_result:match("%d+")
               if length_str then
                  local length = tonumber(length_str) as integer
                  return {
                     type = first_char == "-" and "indel" or "insertion",
                     from = reference_base,
                     to = read_result,
                     length = length,
                     total_correct = no_mutation,
                     total_mutated = count,
                     total_aligned = #read_results
                  }
               end
            end
         end
      end
   end

   -- Check for noisy reads
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
         total_aligned = #read_results
      }
   end

   -- No significant mutations found
   return {
      type = "no_mutation",
      from = reference_base,
      to = ".",
      length = 0,
      total_correct = no_mutation,
      total_mutated = noisy_count,
      total_aligned = #read_results
   }
end

return pileup
