--[[ Module fasta
Module fasta contains fasta parsers and writers.

Fasta is a flat text file format developed in 1985 to store nucleotide and
amino acid sequences. It is extremely simple and well-supported across many
languages. However, this simplicity means that annotation of genetic objects
is not supported.

This module provides a parser and writer for working with Fasta formatted
genetic sequences.
]]

--[[
Apr 25, 2021

Fasta Parser begins here

Many thanks to Jordan Campbell (https://github.com/0x106) for building the first
parser for Poly and thanks to Tim Stiles (https://github.com/TimothyStiles)
for helping complete that PR. This work expands on the previous work by allowing
for concurrent  parsing and giving Poly a specific parser subpackage,
as well as few bug fixes.

Fasta is a very simple file format for working with DNA, RNA, or protein sequences.
It was first released in 1985 and is still widely used in bioinformatics.

https://en.wikipedia.org/wiki/FASTA_format

One interesting use of the concurrent parser is working with the Uniprot
fasta dump files, which are far too large to fit into RAM. This parser is able
to easily handle those files by doing computation actively while the data dump
is getting parsed.

https://www.uniprot.org/downloads

I have removed the  Parsers from the io.go file and moved them into this
subpackage.

Hack the Planet,

Keoni
]]

local bio = require("bio")

local record fasta
   FastaParser: FastaParser
   FastaRecord: FastaRecord
   FastaHeader: FastaHeader
   new_parser: function(reader: bio.Reader, max_line_size: integer): FastaParser
end

-- FastaRecord represents a single fasta file element with an Identifier and its corresponding Sequence
local record FastaRecord is bio.Data
   where self.format == "FASTA"

   identifier: string
   sequence: string
   format: bio.Format
   to_string: function(self: FastaRecord): string
   write: function(self: FastaRecord, writer: bio.Writer): (boolean, string)
end

-- FastaHeader is a blank record, needed for compatibility with bio parsers
local record FastaHeader is bio.Header
   where self.format == "FASTA"

   format: bio.Format
   to_string: function(self: FastaHeader): string
   write: function(self: FastaHeader, writer: bio.Writer): (boolean, string)
end


-- FastaParser implements an efficient buffered parser for FASTA format
local record FastaParser is bio.Parser
   where self.format == "FASTA"

   reader: bio.BufferedReader
   identifier: string
   sequence_buffer: {string}  -- Accumulates sequence data
   start: boolean
   line: integer
   more: boolean

   -- Implement bio.Parser interface
   header: function(self: FastaParser): (bio.Header, string)
   next: function(self: FastaParser): (FastaRecord, string)
   get_format: function(self: FastaParser): bio.Format
end

-- FastaRecord methods
function FastaRecord:to_string(): string
   local result = ">" .. self.identifier .. "\n"
   -- Write sequence in 80 character chunks
   local pos = 1
   while pos <= #self.sequence do
      local chunk = self.sequence:sub(pos, pos + 79)
      result = result .. chunk .. "\n"
      pos = pos + 80
   end
   return result .. "\n"
end

function FastaRecord:write(writer: bio.Writer): (boolean, string)
   local _, err = writer:write(">" .. self.identifier .. "\n")
   if err then return false, err end

   -- Write sequence in 80 character chunks
   local i = 1
   while i <= #self.sequence do
      local chunk = self.sequence:sub(i, i + 79)
      _, err = writer:write(chunk .. "\n")
      if err then return false, err end
      i = i + 80
   end

   return true, nil
end

-- FastaHeader methods
function FastaHeader:to_string(): string
   return ""  -- Fasta has no header
end

function FastaHeader:write(_: bio.Writer): (boolean, string)
   return true, nil  -- Fasta has no header
end

-- Helper functions
local function new_fasta_record(identifier: string, sequence: string): FastaRecord
   local fasta_rec: FastaRecord = {
      identifier = identifier,
      sequence = sequence,
      format = "FASTA",
      to_string = FastaRecord.to_string,
      write = FastaRecord.write
   }
   return fasta_rec
end

local function new_fasta_header(): FastaHeader
   local fasta_head: FastaHeader = {
      format = "FASTA",
      to_string = FastaHeader.to_string,
      write = FastaHeader.write
   }
   return fasta_head
end

-- Create a new FastaParser
function fasta.new_parser(reader: bio.Reader, max_line_size: integer): FastaParser
   local buffered = bio.new_buffered_reader(reader, max_line_size)
   local parser: FastaParser = {
      reader = buffered,
      identifier = "",
      sequence_buffer = {},
      start = true,
      line = 0,
      more = true,
      header = nil,
      next = nil,
      get_format = nil
   }

   function parser:header(): (bio.Header, string)
      return new_fasta_header(), nil
   end

   function parser:get_format(): bio.Format
      return "FASTA"
   end

   -- Create a new record from accumulated sequence data
   local function new_record(self: FastaParser): (FastaRecord, string)
      local sequence = table.concat(self.sequence_buffer)
      if sequence == "" then
         return nil, string.format("%s has no sequence", self.identifier)
      end
      self.sequence_buffer = {}  -- Reset buffer
      return new_fasta_record(self.identifier, sequence), nil
   end

   function parser:next(): (FastaRecord, string)
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

         -- Skip empty lines and comments
         if #line > 0 and line:sub(1,1) ~= ";" then
            -- Handle different line cases
            if line:sub(1,1) ~= ">" and self.start then
               -- Start of file with no identifier
               local err2 = string.format("invalid input: missing sequence identifier for sequence starting at line %d", self.line)
               local fasta_rec, _ = new_record(self)
               return fasta_rec, err2
            elseif line:sub(1,1) ~= ">" then
               -- Sequence line
               table.insert(self.sequence_buffer, line)
            elseif line:sub(1,1) == ">" and not self.start then
               -- New sequence - return current and prepare for next
               local fasta_rec, err3 = new_record(self)
               self.identifier = line:sub(2)  -- Remove '>' and store new identifier
               return fasta_rec, err3
            elseif line:sub(1,1) == ">" and self.start then
               -- First sequence in file
               self.identifier = line:sub(2)
               self.start = false
            end
         end
      end
   end

   return parser
end

return fasta
