--[[ Module slow5
Module slow5 contains slow5 parsers and writers.

Right now, only parsing slow5 files is supported. Support for writing and blow5
coming soon.

slow5 is a file format alternative to fast5, which is the file format outputted
by Oxford Nanopore sequencing devices. fast5 uses hdf5, which is a complex file
format that can only be read and written with a single software library built
in 1998. On the other hand, slow5 uses a .tsv file format, which is easy to
both parse and write.

slow5 files contain both general metadata about the sequencing run and raw
signal reads from the sequencing run. This raw signal can be used directly or
basecalled and used for alignment.

More information on slow5 can be found here: https://github.com/hasindu2008/slow5tools
]]

local bio = require("bio")

local record slow5
   Slow5Parser: Slow5Parser
   Slow5HeaderValue: Slow5HeaderValue
   Slow5Header: Slow5Header
   Slow5Read: Slow5Read
   new_parser: function(reader: bio.Reader, max_line_size: integer): (Slow5Parser, string)
end

-- Slow5HeaderValue contains metadata about a particular read group
local record Slow5HeaderValue
   read_group_id: integer  
   slow5_version: string
   attributes: {string:string}
   end_reason_header_map: {string:integer}
end

-- Slow5Header contains metadata about the sequencing run in general
local record Slow5Header is bio.Header
   where self.format == "SLOW5"

   header_values: {Slow5HeaderValue}
   format: bio.Format
   to_string: function(self: Slow5Header): string
   write: function(self: Slow5Header, writer: bio.Writer): (boolean, string)
end

-- Slow5Read contains metadata and raw signal strengths for a single nanopore read
local record Slow5Read is bio.Data
   where self.format == "SLOW5"

   read_id: string
   read_group_id: integer
   digitisation: number
   offset: number
   read_range: number 
   sampling_rate: number
   len_raw_signal: integer
   raw_signal: {integer}
   
   -- Auxiliary fields
   channel_number: string
   median_before: number
   read_number: integer
   start_mux: integer
   start_time: integer
   end_reason: string
   end_reason_map: {string:integer}
   
   format: bio.Format
   to_string: function(self: Slow5Read): string
   write: function(self: Slow5Read, writer: bio.Writer): (boolean, string)
end

-- Slow5Parser implements a parser for SLOW5 format
local record Slow5Parser is bio.Parser
   where self.format == "SLOW5"

   reader: bio.BufferedReader
   line: integer
   headerMap: {integer:string}
   endReasonMap: {integer:string}
   endReasonSlow5HeaderMap: {string:integer}
   headerValue: Slow5Header
   hitEOF: boolean

   -- Implement bio.Parser interface
   header: function(self: Slow5Parser): (Slow5Header, string)
   next: function(self: Slow5Parser): (Slow5Read, string)
   get_format: function(self: Slow5Parser): bio.Format
end

-- Check if string key exists in known end reasons map
local knownend_reasons: {string:boolean} = {
   ["unknown"] = true,
   ["partial"] = true,
   ["mux_change"] = true,
   ["unblock_mux_change"] = true,
   ["data_service_unblock_mux_change"] = true,
   ["signal_positive"] = true,
   ["signal_negative"] = true
}

-- Read methods
function Slow5Read:to_string(): string
   local parts = {
      self.read_id,
      tostring(self.read_group_id),
      string.format("%g", self.digitisation),
      string.format("%g", self.offset),
      string.format("%g", self.read_range), 
      string.format("%g", self.sampling_rate),
      tostring(self.len_raw_signal)
   }

   -- Convert raw signals to comma-separated string
   local signals = {}
   for i = 1, #self.raw_signal do
      table.insert(signals, tostring(self.raw_signal[i]))
   end
   table.insert(parts, table.concat(signals, ","))

   -- Add auxiliary fields
   table.insert(parts, tostring(self.start_time))
   table.insert(parts, tostring(self.read_number))
   table.insert(parts, tostring(self.start_mux))
   table.insert(parts, string.format("%g", self.median_before))
   table.insert(parts, tostring(self.end_reason_map[self.end_reason]))
   table.insert(parts, self.channel_number)

   return table.concat(parts, "\t") .. "\n"
end

function Slow5Read:write(writer: bio.Writer): (boolean, string)
   local _, err = writer:write(self:to_string())
   if err then return false, err end
   return true, nil
end

-- Slow5Header methods
function Slow5Header:to_string(): string
   local sb = {}
   table.insert(sb, "#slow5_version\t" .. self.header_values[1].slow5_version .. "\n")
   table.insert(sb, "#num_read_groups\t" .. tostring(#self.header_values) .. "\n")
   
   -- Write attributes in sorted order
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

function Slow5Header:write(writer: bio.Writer): (boolean, string)
   local _, err = writer:write(self:to_string())
   if err then return false, err end
   return true, nil
end

-- Create a new SLOW5 parser
function slow5.new_parser(reader: bio.Reader, max_line_size: integer): (Slow5Parser, string)
   local buffered = bio.new_buffered_reader(reader, max_line_size)
   
   local parser: Slow5Parser = {
      reader = buffered,
      line = 0,
      headerMap = {},
      endReasonMap = {},
      endReasonSlow5HeaderMap = {},
      header = nil,
      hitEOF = false
   }

   -- Initialize empty header
   local header: Slow5Header = {
      header_values = {},
      format = "SLOW5",
      to_string = Slow5Header.to_string,
      write = Slow5Header.write
   }

   local slow5Version = ""
   local numReadGroups = 0

   -- Parse header section
   while true do
      local lineBytes, err = parser.reader:read_line()
      if err then
         if err == "EOF" then
            return nil, err
         end
         return nil, err
      end

      parser.line = parser.line + 1
      local line = lineBytes:gsub("^%s*(.-)%s*$", "%1") -- trim whitespace

      -- Split line by tabs
      local values = {}
      for value in line:gmatch("[^\t]+") do
         table.insert(values, value)
      end

      if #values < 2 then
         return nil, string.format("Got following line without tabs: %s", line)
      end

      -- Handle special header lines
      if numReadGroups == 0 then
         if values[1] == "#slow5_version" then
            slow5Version = values[2]
         elseif values[1] == "#num_read_groups" then
            numReadGroups = tonumber(values[2]) as integer
            if not numReadGroups then
               return nil, "Invalid num_read_groups value"
            end
            
            -- Initialize header values
            for id = 0, numReadGroups-1 do
               table.insert(header.header_values, {
                  read_group_id = id,
                  slow5_version = slow5Version,
                  attributes = {},
                  end_reason_header_map = {}
               })
            end
         end
      else
         -- Handle end reason enums
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
                     parser.endReasonMap[index-1] = reason
                     parser.endReasonSlow5HeaderMap[reason] = index-1
                  end
                  
                  -- Add endReasonMap to each header
                  for i = 1, #header.header_values do
                     header.header_values[i].end_reason_header_map = parser.endReasonSlow5HeaderMap
                  end
               end
            end
         elseif values[1] == "#read_id" then
            -- Store header map
            parser.headerMap[1] = "read_id"
            for i = 2, #values do
               parser.headerMap[i] = values[i]
            end
            break
         else
            -- Regular attribute line
            if #values ~= numReadGroups + 1 then
               return nil, string.format("Improper amount of information for read groups. Needed %d, got %d", numReadGroups+1, #values)
            end
            
            for id = 1, numReadGroups do
               header.header_values[id].attributes[values[1]] = values[id+1]
            end
         end
      end
   end

   parser.headerValue = header

   function parser:header(): (Slow5Header, string)
      return self.headerValue, nil
   end

   function parser:get_format(): bio.Format
      return "SLOW5"
   end

   function parser:next(): (Slow5Read, string) 
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

      -- Split line by tabs
      local values = {}
      for value in line:gmatch("[^\t]+") do
         table.insert(values, value)
      end

      local read: Slow5Read = {
         format = "SLOW5",
         to_string = Slow5Read.to_string,
         write = Slow5Read.write
      }

      -- Parse values based on header map
      for valueIndex, value in ipairs(values) do
         local fieldValue = self.headerMap[valueIndex]
         if value ~= "." then -- Invert the continue condition
            if fieldValue == "read_id" then
               read.read_id = value
            elseif fieldValue == "read_group" then
               read.read_group_id = tonumber(value) as integer
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
               read.len_raw_signal = tonumber(value) as integer
               if not read.len_raw_signal then
                  return nil, string.format("Failed to convert len_raw_signal '%s' to number on line %d", value, self.line)
               end
            elseif fieldValue == "raw_signal" then
               read.raw_signal = {}
               for signal in value:gmatch("[^,]+") do
                  local num = tonumber(signal) as integer
                  if not num then
                     return nil, string.format("Failed to convert raw signal '%s' to number on line %d", signal, self.line)
                  end
                  table.insert(read.raw_signal, num)
               end
            elseif fieldValue == "start_time" then
               read.start_time = tonumber(value) as integer
               if not read.start_time then
                  return nil, string.format("Failed to convert start_time '%s' to number on line %d", value, self.line)
               end
            elseif fieldValue == "read_number" then
               read.read_number = tonumber(value) as integer
               if not read.read_number then
                  return nil, string.format("Failed to convert read_number '%s' to number on line %d", value, self.line)
               end
            elseif fieldValue == "start_mux" then
               read.start_mux = tonumber(value) as integer
               if not read.start_mux then
                  return nil, string.format("Failed to convert start_mux '%s' to number on line %d", value, self.line)
               end
            elseif fieldValue == "median_before" then
               read.median_before = tonumber(value)
               if not read.median_before then
                  return nil, string.format("Failed to convert median_before '%s' to number on line %d", value, self.line)
               end
            elseif fieldValue == "end_reason" then
               local endReasonIndex = tonumber(value) as integer
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

return slow5
