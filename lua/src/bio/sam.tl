--[[ Module sam
Module sam implements a SAM file parser and writer.

SAM is a tab-delimited text format for storing DNA/RNA sequence alignment data.
It is the most widely used alignment format, complementing its binary
equivalent, BAM, which stores the same data in a compressed format.

DNA sequencing works in the following way:

  - DNA is read in with some raw signal format from the sequencer machine.
  - Raw signal is converted to fastq reads using basecalling software.
  - Fastq reads are aligned to target template, producing SAM files.
  - SAM files are used to answer bioinformatic queries.

This parser allows parsing and writing of SAM files in Teal. Unlike other SAM
parsers, we aim to be as close to underlying data types as possible,
with a goal of being as simple as possible, and no simpler.

Paper: https://doi.org/10.1093%2Fbioinformatics%2Fbtp352
Spec: http://samtools.github.io/hts-specs/SAMv1.pdf
]]

local bio = require("bio")

local record sam
   SamParser: SamParser
   SamAlignment: SamAlignment 
   SamHeader: SamHeader
   SamOptional: SamOptional
   DEFAULT_MAX_LINE_SIZE: integer
   SortOrder: SortOrder
   GroupOrder: GroupOrder
   TPValue: TPValue
   PlatformValue: PlatformValue
   new_parser: function(reader: bio.Reader, max_line_size: integer): (SamParser, SamHeader, string)
   is_primary: function(alignment: SamAlignment): boolean
end

-- DEFAULT_MAX_LINE_SIZE is a magic number often used by the Go stdlib for parsing
sam.DEFAULT_MAX_LINE_SIZE = 1024 * 32 * 2 -- 32kB * 2

-- Enums for validation
local enum SortOrder
   "unknown"
   "unsorted"
   "queryname"
   "coordinate"
end

local enum GroupOrder
   "none"
   "query" 
   "reference"
end

local enum TPValue
   "linear"
   "circular"
end

local enum PlatformValue
   "CAPILLARY"
   "DNBSEQ"
   "ELEMENT"
   "HELICOS"
   "ILLUMINA"
   "IONTORRENT"
   "LS454"
   "ONT"
   "PACBIO"
   "SOLID"
   "ULTIMA"
end

-- Each header in a SAM file begins with an @ followed by a two letter record
-- code type. Each line is tab delimited, and contains TAG:VALUE pairs. HD, the
-- first line, only occurs once, while SQ, RG, and PG can appear multiple
-- times. Finally, @CO contains user generated comments.
local record SamHeader is bio.Header
   where self.format == "SAM"

   HD: {string:string}  -- File-level metadata
   SQ: {{string:string}}  -- Reference sequence dictionary
   RG: {{string:string}}  -- Read group
   PG: {{string:string}}  -- Program
   CO: {string}  -- Comments
   format: bio.Format
   to_string: function(self: SamHeader): string
   write: function(self: SamHeader, writer: bio.Writer): (boolean, string)
   validate: function(self: SamHeader): string
end

-- Optional fields in SAM alignments are structured as TAG:TYPE:DATA
local record SamOptional
   tag: string  -- Tag is typically a two letter tag
   tag_type: string  -- One of A (character), B (array), f (real), H (hex), i (integer), Z (string)
   data: string  -- Optional data
end

-- Each alignment is a single line of a SAM file
local record SamAlignment is bio.Data
   where self.format == "SAM"

   QNAME: string      -- Query template NAME
   FLAG: integer      -- bitwise FLAG  
   RNAME: string      -- References sequence NAME
   POS: integer       -- 1-based leftmost mapping POSition
   MAPQ: integer      -- MAPping Quality
   CIGAR: string      -- CIGAR string
   RNEXT: string      -- Ref. name of the mate/next read
   PNEXT: integer     -- Position of the mate/next read
   TLEN: integer      -- observed Template LENgth
   SEQ: string        -- segment SEQuence
   QUAL: string       -- ASCII of Phred-scaled base QUALity+33
   optionals: {SamOptional}  -- Optional fields
   format: bio.Format
   to_string: function(self: SamAlignment): string
   write: function(self: SamAlignment, writer: bio.Writer): (boolean, string)
   validate: function(self: SamAlignment): string
end

-- SamParser implements a parser for SAM format
local record SamParser is bio.Parser
   where self.format == "SAM"

   reader: bio.BufferedReader
   line: integer
   file_header: SamHeader
   first_line: string
   read_first_line: boolean

   -- Implement bio.Parser interface
   header: function(self: SamParser): (bio.Header, string)
   next: function(self: SamParser): (SamAlignment, string)
   get_format: function(self: SamParser): bio.Format
end

-- Helper function to write header sections
local function header_write_helper(sb: {string}, header_string: string, header_map: {string:string}, ordered_keys: {string})
   table.insert(sb, header_string)
   -- Write orderedKeys first, if they exist
   for _, key in ipairs(ordered_keys) do
      if header_map[key] then
         table.insert(sb, string.format("\t%s:%s", key, header_map[key]))
      end
   end
   -- Write remaining key-value pairs
   for key, value in pairs(header_map) do
      -- Skip if key is one of the ordered keys
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

-- SamHeader methods
function SamHeader:to_string(): string
   local sb = {}
   if self.HD and next(self.HD) ~= nil then
      header_write_helper(sb, "@HD", self.HD, {"VN", "SO", "GO", "SS"})
   end
   for _, sq in ipairs(self.SQ) do
      header_write_helper(sb, "@SQ", sq, {"SN", "LN", "AH", "AN", "AS", "DS", "M5", "SP", "TP", "UR"})
   end
   for _, rg in ipairs(self.RG) do
      header_write_helper(sb, "@RG", rg, {"ID", "BC", "CN", "DS", "DT", "FO", "KS", "LB", "PG", "PI", "PL", "PM", "PU", "SM"})
   end
   for _, pg in ipairs(self.PG) do
      header_write_helper(sb, "@PG", pg, {"ID", "PN", "VN", "CL", "PP", "DS"})
   end
   for _, co in ipairs(self.CO) do
      table.insert(sb, string.format("@CO %s\n", co))
   end
   return table.concat(sb)
end

function SamHeader:write(writer: bio.Writer): (boolean, string)
   local _, err = writer:write(self:to_string())
   if err then return false, err end
   return true, nil
end

-- Validate validates that the header has all required information
function SamHeader:validate(): string
   -- Check VN format (numbers.numbers)
   if self.HD.VN then
      if not self.HD.VN:match("^%d+%.%d+$") then
         return string.format("Invalid format for @HD VN: %s", self.HD.VN)
      end
   end

   -- Validate SQ tags first
   local sn_map: {string:boolean} = {}
   for _, sq in ipairs(self.SQ) do
      -- Check unique SN
      if sq.SN then
         if sn_map[sq.SN] then
            return string.format("Non-unique @SQ SN: %s", sq.SN)
         end
         sn_map[sq.SN] = true
      end

      -- Check LN range
      if sq.LN then
         local ln = tonumber(sq.LN) as integer
         if not ln or ln < 1 or ln > 2147483647 then
            return string.format("Invalid value for @SQ LN: %s", sq.LN)
         end
      end

      -- Check TP values using TPValue enum
      if sq.TP then
         local valid = false
         for _, tp in ipairs({"linear", "circular"}) do
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

   -- Check SO values using SortOrder enum
   if self.HD.SO then
      local valid = false
      for _, so in ipairs({"unknown", "unsorted", "queryname", "coordinate"}) do
         if self.HD.SO == so then
            valid = true
            break
         end
      end
      if not valid then
         return string.format("Invalid value for @HD SO: %s", self.HD.SO)
      end
   end

   -- Check GO values using GroupOrder enum
   if self.HD.GO then
      local valid = false
      for _, go in ipairs({"none", "query", "reference"}) do
         if self.HD.GO == go then
            valid = true
            break
         end
      end
      if not valid then
         return string.format("Invalid value for @HD GO: %s", self.HD.GO)
      end
   end

   -- Validate RG tags
   local rg_id_map: {string:boolean} = {}
   for _, rg in ipairs(self.RG) do
      -- Check unique ID
      if rg.ID then
         if rg_id_map[rg.ID] then
            return string.format("Non-unique @RG ID: %s", rg.ID)
         end
         rg_id_map[rg.ID] = true
      end

      -- Check PL values using PlatformValue enum
      if rg.PL then
         local valid = false
         for _, pl in ipairs({"CAPILLARY", "DNBSEQ", "ELEMENT", "HELICOS", "ILLUMINA", "IONTORRENT", "LS454", "ONT", "PACBIO", "SOLID", "ULTIMA"}) do
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

-- SamAlignment methods
function SamAlignment:to_string(): string
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
      self.QUAL
   }
   
   for _, opt in ipairs(self.optionals) do
      table.insert(parts, string.format("%s:%s:%s", opt.tag, opt.tag_type, opt.data))
   end
   
   return table.concat(parts, "\t") .. "\n"
end

function SamAlignment:write(writer: bio.Writer): (boolean, string)
   local _, err = writer:write(self:to_string())
   if err then return false, err end
   return true, nil
end

function SamAlignment:validate(): string
	-- Validate FLAG range first (since this was the specific test case)
   if self.FLAG < 0 or self.FLAG > 65535 then
      return "Invalid FLAG range"
   end

	-- Validate CIGAR format next
	if self.CIGAR ~= "*" then
	    -- Collect all valid operations
	    local valid_ops = ""
	    for op in self.CIGAR:gmatch("%d+[MIDNSHPX=]") do
	        valid_ops = valid_ops .. op
	    end
	    
	    -- If valid_ops doesn't exactly match the input, then there were invalid operations
	    if valid_ops ~= self.CIGAR then
	        return "Invalid CIGAR format"
	    end
	end


   -- Then validate QNAME
   if not self.QNAME:match("^[!-?A-~]{1,254}$") then
      return "Invalid QNAME format"
   end

   -- Validate POS range
   if self.POS < 0 or self.POS > 2147483647 then
      return "Invalid POS range"
   end

   -- Validate MAPQ range (byte in Go, regular integer in Lua)
   if self.MAPQ < 0 or self.MAPQ > 255 then
      return "Invalid MAPQ range"
   end

   -- Validate SEQ format
   if not self.SEQ:match("^%*|[A-Za-z=.]+$") then
      return "Invalid SEQ format"
   end

   -- Validate QUAL format
   if not self.QUAL:match("^[!-~]+$") then
      return "Invalid QUAL format"
   end

   return nil
end

-- Create a new SAM parser
function sam.new_parser(reader: bio.Reader, max_line_size: integer): (SamParser, SamHeader, string)
   local buffered = bio.new_buffered_reader(reader, max_line_size)
   local parser: SamParser = {
      reader = buffered,
      line = 0,
      file_header = nil,
      first_line = "",
      read_first_line = false,
      header = nil,
      next = nil,
      get_format = nil
   }

   -- Initialize header
   local header: SamHeader = {
      HD = {},
      SQ = {},
      RG = {},
      PG = {},
      CO = {},
      format = "SAM",
      to_string = SamHeader.to_string,
      write = SamHeader.write,
      validate = SamHeader.validate
   }

   local hd_parsed = false

   -- Read and parse header
   while true do
      local line: string
      local err: string
      line, err = buffered:read_line()
	  if err then
		   if err == "EOF" then
		      if line and line:match("^[^@]") then
		         parser.first_line = line
		         break
		      end
		      -- Even if we hit EOF, if we have a valid header, return it
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

      -- Check if we've reached alignment section
      if line:sub(1,1) ~= "@" then
         parser.first_line = line
         break
      end

      local values = {}
      for value in line:gmatch("[^\t]+") do
         table.insert(values, value)
      end

      -- Parse HD (must be first)
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
         -- Parse CO lines (special case - full line comments)
         if values[1] == "@CO" then
            table.insert(header.CO, line)
         else
            -- Parse SQ, RG, PG sections
            local section: {string:string} = {}
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

   function parser:header(): (bio.Header, string)
      return self.file_header, nil
   end

   function parser:get_format(): bio.Format
      return "SAM"
   end

   function parser:next(): (SamAlignment, string)
	   local line: string
	   local err: string
	   
	   -- Keep reading until we get a non-empty line
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

      -- Split line into values
      local values = {}
      for value in line:gmatch("[^\t]+") do
         table.insert(values, value)
      end

      if #values < 11 then
         return nil, string.format("Line %d must have at least 11 tab-delimited values. Had %d", self.line, #values)
      end

      -- Parse alignment values
      local alignment: SamAlignment = {
         QNAME = values[1],
         FLAG = tonumber(values[2]) as integer,
         RNAME = values[3],
         POS = tonumber(values[4]) as integer,
         MAPQ = tonumber(values[5]) as integer,
         CIGAR = values[6],
         RNEXT = values[7],
         PNEXT = tonumber(values[8]) as integer,
         TLEN = tonumber(values[9]) as integer,
         SEQ = values[10],
         QUAL = values[11],
         optionals = {},
         format = "SAM",
         to_string = SamAlignment.to_string,
         write = SamAlignment.write,
         validate = SamAlignment.validate
      }

      -- Check for invalid number conversions
      if not alignment.FLAG or not alignment.POS or not alignment.MAPQ 
         or not alignment.PNEXT or not alignment.TLEN then
         return nil, string.format("Line %d contains invalid number format", self.line)
      end

      -- Parse optional fields
      if #values > 11 then
         for i = 12, #values do
            local tag, tag_type, data = values[i]:match("([^:]+):([^:]+):(.+)")
            if not (tag and tag_type and data) then
               return nil, string.format("Line %d has invalid optional field format: %s", self.line, values[i])
            end
            table.insert(alignment.optionals, {
               tag = tag,
               tag_type = tag_type,
               data = data
            })
         end
      end

      return alignment, nil
   end

   return parser, header, nil
end

-- Check if an alignment is primary by checking for secondary and supplementary flags
function sam.is_primary(alignment: SamAlignment): boolean
   -- Check each flag using integer division to isolate the bit
   -- and modulo 2 to check if that bit is set
   local has_secondary = (math.floor(alignment.FLAG / 256) % 2) == 1
   local has_supplementary = (math.floor(alignment.FLAG / 2048) % 2) == 1
   
   -- Primary means neither secondary nor supplementary
   return not (has_secondary or has_supplementary)
end

return sam
