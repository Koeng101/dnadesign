--[[ Module fastq
Module fastq contains fastq parsers and writers.

Fastq is a flat text file format developed in ~2000 to store nucleotide
sequencing data. While similar to fastq, fastq has a few differences. First,
the sequence identifier begins with @ instead of >, and includes quality
values for a sequence.

This module provides a parser and writer for working with Fastq formatted
sequencing data.
]]

local bio = require("bio")

local record fastq
   FastqParser: FastqParser
   FastqRead: FastqRead
   FastqHeader: FastqHeader
   new_parser: function(reader: bio.Reader, max_line_size: integer): FastqParser
end

-- FastqRead represents a single Fastq read element with an Identifier, its corresponding sequence, its quality score, and any optional pieces of data.
local record FastqRead is bio.Data
   where self.format == "FASTQ"

   identifier: string
   optionals: {string:string}  -- Nanopore, for example, carries along data like: `read=13956 ch=53 start_time=2020-11-11T01:49:01Z`
   sequence: string
   quality: string
   format: bio.Format
   to_string: function(self: FastqRead): string
   write: function(self: FastqRead, writer: bio.Writer): (boolean, string)
   deep_copy: function(self: FastqRead): FastqRead
end

-- FastqHeader is a blank record, needed for compatibility with bio parsers
local record FastqHeader is bio.Header
   where self.format == "FASTQ"

   format: bio.Format
   to_string: function(self: FastqHeader): string
   write: function(self: FastqHeader, writer: bio.Writer): (boolean, string)
end

-- FastqParser implements an efficient buffered parser for FASTQ format
local record FastqParser is bio.Parser
   where self.format == "FASTQ"

   reader: bio.BufferedReader
   line: integer
   at_eof: boolean

   -- Implement bio.Parser interface
   header: function(self: FastqParser): (bio.Header, string)
   next: function(self: FastqParser): (FastqRead, string)
   get_format: function(self: FastqParser): bio.Format
end

-- FastqRead methods
function FastqRead:to_string(): string
   local result = "@" .. self.identifier
   -- Sort keys for consistent output
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

function FastqRead:write(writer: bio.Writer): (boolean, string)
   local ok, err = writer:write(self:to_string())
   if not ok then
      return false, err
   end
   return true, nil
end

function FastqRead:deep_copy(): FastqRead
   local new_read: FastqRead = {
      identifier = self.identifier,
      sequence = self.sequence,
      quality = self.quality,
      format = "FASTQ",
      optionals = {},
      to_string = FastqRead.to_string,
      write = FastqRead.write,
      deep_copy = FastqRead.deep_copy
   }
   for key, value in pairs(self.optionals) do
      new_read.optionals[key] = value
   end
   return new_read
end

-- FastqHeader methods
function FastqHeader:to_string(): string
   return ""  -- Fastq has no header
end

function FastqHeader:write(_: bio.Writer): (boolean, string)
   return true, nil  -- Fastq has no header
end

-- Helper functions
local function new_fastq_read(identifier: string, sequence: string, quality: string, optionals: {string:string}): FastqRead
   local fastq_read: FastqRead = {
      identifier = identifier,
      sequence = sequence,
      quality = quality,
      optionals = optionals or {},
      format = "FASTQ",
      to_string = FastqRead.to_string,
      write = FastqRead.write,
      deep_copy = FastqRead.deep_copy
   }
   return fastq_read
end

local function new_fastq_header(): FastqHeader
   local fastq_head: FastqHeader = {
      format = "FASTQ",
      to_string = FastqHeader.to_string,
      write = FastqHeader.write
   }
   return fastq_head
end

-- Create a new FastqParser
function fastq.new_parser(reader: bio.Reader, max_line_size: integer): FastqParser
   local buffered = bio.new_buffered_reader(reader, max_line_size)
   local parser: FastqParser = {
      reader = buffered,
      line = 0,
      at_eof = false,
      header = nil,
      next = nil,
      get_format = nil
   }

   function parser:header(): (bio.Header, string)
      return new_fastq_header(), nil
   end

   function parser:get_format(): bio.Format
      return "FASTQ"
   end

   -- Parse optionals from a space-separated string
   local function parse_optionals(line: string): {string:string}
      local optionals: {string:string} = {}
      local parts = line:gmatch("%S+")
      -- Skip the identifier part
      parts() 
      for part in parts do
         local key, value = part:match("([^=]+)=(.+)")
         if key and value then
            optionals[key] = value
         end
      end
      return optionals
   end

   -- Check if sequence contains only valid characters
   local function validate_sequence(seq: string): string
      if #seq <= 0 or seq == "+" then
         return "empty fastq sequence"
	  end
      local valid_chars: {string:boolean} = {["A"]=true, ["T"]=true, ["G"]=true, ["C"]=true, ["N"]=true}
      for i = 1, #seq do
         local char = seq:sub(i,i)
         if not valid_chars[char] then
            return string.format("Only letters ATGCN are allowed for DNA/RNA in fastq file. Got letter: %s", char)
         end
      end
      return nil
   end

   function parser:next(): (FastqRead, string)
      if self.at_eof then
         return nil, "EOF"
      end

      -- Read identifier line
      local identifier_line, err = self.reader:read_line()
      self.line = self.line + 1
      if err then
         return nil, err
      end
      if #identifier_line == 0 or identifier_line:sub(1,1) ~= "@" then
         return nil, string.format("did not find fastq start '@', got to line %d", self.line)
      end
      
      local identifier = identifier_line:match("^@([^%s]+)")
      local optionals = parse_optionals(identifier_line)

      -- Read sequence line
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

      -- Skip + line
      local _, err3 = self.reader:read_line()
      self.line = self.line + 1
      if err3 then
         return nil, err3
      end

      -- Read quality line
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

return fastq
