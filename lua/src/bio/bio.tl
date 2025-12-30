--[[ Module bio
The bio module provides utilities for reading and writing sequence data.

There are many different biological file formats for different applications.
The bio package provides a consistent way to work with each of these file
formats. The underlying data returned by each parser is as raw as we can return
while still being easy to use for downstream applications, and should be
immediately recognizable as the original format.

TODO: new_file_reader would be convenient function
]]

local record bio
	Format: Format
    Reader: Reader
    Writer: Writer
	Header: Header
	Data: Data
	Parser: Parser
	DEFAULT_MAX_LENGTHS: {Format:integer}
	
	StringReader: StringReader
	StringWriter: StringWriter
	BufferedReader: BufferedReader
	new_string_reader: function(content: string): StringReader
	new_string_writer: function(): StringWriter
	new_buffered_reader: function(reader: Reader, max_line_size: integer): BufferedReader
end

-- Format is an enum of different parser formats.
local enum Format
   "FASTA"
   "FASTQ" 
   "GENBANK"
   "SLOW5"
   "SAM"
   "PILEUP"
end

--[[
In the original dnadesign/poly, which was coded in Go, we used the io.Reader
and io.Writer interfaces in order to handle data. These are very convenient
because they give a standardized way to get data from tons of different sources
(file, http, torrent, ftp, etc). They're super easy to implement, so any system
that you need to pass data into dnadesign from can implement these simple
interfaces.

By having a Reader instead of passing in a whole file, we can process much
larger files by only loading into memory a little bit at a time.
--]]

-- Reader interface for reading data
local interface Reader
    -- read reads up to n bytes from the source
    -- Returns the bytes read and any error
    read: function(self, n: integer): (string, string)

    -- close releases any resources
    close: function(self): (boolean, string)
end

-- Writer interface for writing data
local interface Writer
    -- write writes the given string to the destination
    -- Returns number of bytes written and any error
    write: function(self, data: string): (integer, string)

    -- close releases any resources
    close: function(self): (boolean, string)
end

--[[
DefaultMaxLineLength variables are defined for performance reasons. While
parsing, reading byte-by-byte takes far, far longer than reading many bytes
into a buffer. In Lua/Teal, we choose similar buffer sizes to Go's bufio
for consistency and proven performance characteristics.
]]
local DEFAULT_MAX_LINE_LENGTH = 65536 -- 64KB like Go's bufio.MaxScanTokenSize

--[[
DefaultMaxLengths maps each format to its maximum line length.
Some formats, particularly SLOW5 and FASTQ, can have very long lines
and need larger buffers.
]]
local DEFAULT_MAX_LENGTHS: {Format:integer} = {
   ["FASTA"] = DEFAULT_MAX_LINE_LENGTH,
   ["FASTQ"] = 8 * 1024 * 1024,    -- The longest single nanopore sequencing read so far is 4Mb. A 8mb buffer should be large enough.
   ["GENBANK"] = DEFAULT_MAX_LINE_LENGTH,
   ["SLOW5"] = 128 * 1024 * 1024,  -- 128MB is used because slow5 lines can be massive
   ["SAM"] = DEFAULT_MAX_LINE_LENGTH,
   ["PILEUP"] = DEFAULT_MAX_LINE_LENGTH
}
bio.DEFAULT_MAX_LENGTHS = DEFAULT_MAX_LENGTHS

--[[
Header is the base interface that all format-specific headers must implement.
Some formats (like FASTA) don't have headers, while others (like SAM) do.
]]
local interface Header
   -- The format this header is for
   format: Format
   
   -- Convert header to string representation
   to_string: function(self): string
end

--[[
Data is the base interface that all format-specific data records must implement.
Each format will extend this with its specific fields.
]]
local interface Data 
   -- The format this data is for
   format: Format
   
   -- Convert data to string representation
   to_string: function(self): string
end

--[[
Parser is the core interface that all format-specific parsers must implement.
It provides a consistent way to read headers and data records from any
supported format.
]]
local interface Parser
   -- Header returns the format-specific header if one exists.
   -- Some formats like FASTA don't have headers and will return nil.
   header: function(self): Header, string -- returns header and error
   
   -- Next returns the next data record from the parser.
   -- Returns nil,"EOF" when reaching end of file.
   next: function(self): Data, string -- returns data and error
   
   -- GetFormat returns the format this parser handles
   format: Format
end

--[[
The stringreader package provides a StringReader implementation for the bio package.
It allows reading string content line by line, simulating file-like behavior.
This is particularly useful for testing and for processing string data.
--]]

-- StringReader implements the bio.Reader interface
local record StringReader is bio.Reader
    content: string
    position: integer
    line_number: integer
    _eof_emitted: boolean
    
    read: function(self: StringReader, n: integer): (string, string)
    close: function(self: StringReader): (boolean, string)
    get_line_number: function(self: StringReader): number
end

-- Create a new string reader that implements the bio.Reader interface
function bio.new_string_reader(content: string): StringReader
    if type(content) ~= "string" then
        error("StringReader content must be a string")
    end

    local reader: StringReader = {
        content = content,
        position = 1,
        line_number = 0,
        _eof_emitted = false,
        
        read = nil,  -- Will be defined below
        close = nil, -- Will be defined below
        get_line_number = nil -- Will be defined below
    }

    -- Read implements bio.Reader read interface.
    -- Returns up to n bytes from the current position.
    function reader:read(n: integer): (string, string)
        -- Return EOF only once
        if self._eof_emitted then
            return nil, "EOF"
        end

        -- Check if we're at the end
        if self.position > #self.content then
            self._eof_emitted = true
            return nil, "EOF"
        end

        -- Calculate how many bytes we can actually read
        local bytes_available = #self.content - self.position + 1
        local bytes_to_read = math.min(n, bytes_available)
        
        -- Read the bytes
        local chunk = self.content:sub(self.position, self.position + bytes_to_read - 1)
        self.position = self.position + bytes_to_read
        
        -- Update line count by counting newlines in the chunk
		self.line_number = self.line_number + (select(2, chunk:gsub("\n", "")) as integer)
        
        return chunk, nil
    end

    -- Close implements Reader close interface
    function reader:close(): (boolean, string)
        self._eof_emitted = true
        return true, nil
    end

    -- Get current line number
    function reader:get_line_number(): number
        return self.line_number
    end

    return reader
end

-- StringWriter implements the bio.Writer interface for testing
local record StringWriter is bio.Writer
    content: string
    
    write: function(self: StringWriter, data: string): (integer, string)
    close: function(self: StringWriter): (boolean, string)
    get_content: function(self: StringWriter): string
end

-- Create a new string writer that implements the bio.Writer interface
function bio.new_string_writer(): StringWriter
    local writer: StringWriter = {
        content = "",
        write = nil,  -- Will be defined below
        close = nil,  -- Will be defined below
        get_content = nil -- Will be defined below
    }

    -- Write implements bio.Writer write interface
    function writer:write(data: string): (integer, string)
        self.content = self.content .. data
        return #data, nil
    end

    -- Close implements Writer close interface
    function writer:close(): (boolean, string)
        return true, nil
    end

    -- Get current content
    function writer:get_content(): string
        return self.content
    end

    return writer
end

local record BufferedReader is Reader
   wrapped: Reader
   buffer: string        
   buffer_pos: integer   
   buffer_size: integer  
   max_line_size: integer

   read: function(self: BufferedReader, n: integer): (string, string)
   close: function(self: BufferedReader): (boolean, string)
   read_line: function(self: BufferedReader): (string, string)
end

-- Create a new buffered reader that implements the bio.Reader interface
function bio.new_buffered_reader(reader: Reader, max_line_size: integer): BufferedReader
   local br: BufferedReader = {
      wrapped = reader,
      buffer = "",
      buffer_pos = 1,
      buffer_size = 0,
      max_line_size = max_line_size,
      read = nil,  -- Defined below
      close = nil, -- Defined below 
      read_line = nil -- Defined below
   }

   -- Implement the read method - delegates to wrapped reader
   function br:read(n: integer): (string, string)
      return self.wrapped:read(n)
   end

   -- Implement the close method - delegates to wrapped reader
   function br:close(): (boolean, string)
      return self.wrapped:close()
   end

   -- Efficiently read a line from the buffer, fetching more data if needed
   function br:read_line(): (string, string)
      local line_parts = {}
      local found_newline = false
      
      while not found_newline do
         -- If buffer is empty or exhausted, read more
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
         
         -- Look for newline in current buffer
         local newline_pos = self.buffer:find("\n", self.buffer_pos)
         if newline_pos then
            -- Found newline, extract up to it
            local part = self.buffer:sub(self.buffer_pos, newline_pos - 1)
            if part ~= "" then
               table.insert(line_parts, part)
            end
            self.buffer_pos = newline_pos + 1
            found_newline = true
         else
            -- No newline, take rest of buffer
            local part = self.buffer:sub(self.buffer_pos)
            if part ~= "" then
               table.insert(line_parts, part)
            end
            self.buffer_pos = self.buffer_size + 1
         end
      end
      
      local line = table.concat(line_parts)
      -- Remove trailing \r if present
      if line:sub(-1) == "\r" then
         line = line:sub(1, -2)
      end
      return line, nil
   end

   return br
end

return bio
