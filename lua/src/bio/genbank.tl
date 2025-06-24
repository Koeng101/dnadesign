--[[ Module genbank
Module genbank provides genbank parsers and writers.

GenBank is a flat text file format developed in the 1980s to annotate genetic
sequences, and has since become the standard for sharing annotated genetic
sequences.

This module provides a parser and writer to convert between the GenBank file
format and the more general Genbank struct.
]]

local bio = require("bio")
local transform = require("transform")

-- Core package definition 
local record genbank
    -- Types
    Genbank: Genbank
    Meta: Meta  
    Feature: Feature
    Reference: Reference
    Location: Location
    BaseCount: BaseCount
    GenbankParser: GenbankParser

    -- Functions exported by package
    new_parser: function(reader: bio.Reader, max_line_size: integer): GenbankParser
    get_feature_sequence: function(feature: Feature, location: Location): (string, string)
end

-- Core data structures with proper table handling
local record Location
    Start: integer
    End: integer 
    Complement: boolean
    Join: boolean
    FivePrimePartial: boolean
    ThreePrimePartial: boolean
    GbkLocationString: string
    SubLocations: {Location}
end

local record BaseCount
    Base: string
    Count: integer
end

local record Locus 
    Name: string
    SequenceLength: string
    MoleculeType: string
    GenbankDivision: string
    ModificationDate: string  
    SequenceCoding: string
    Circular: boolean
end

local record Reference
    Authors: string
    Title: string
    Journal: string  
    PubMed: string
    Remark: string
    Range: string
    Consortium: string
end

-- Note: Meta requires careful table handling
local record Meta
    Date: string
    Definition: string
    Accession: string
    Version: string
    Keywords: string
    Organism: string
    Source: string
    Taxonomy: {string}
    Origin: string
    Locus: Locus
    References: {Reference}
    BaseCount: {BaseCount}  
    Other: {string:string} -- Map type
    Name: string
    SequenceHash: string
    SequenceHashFunction: string
end

-- Note: Feature requires careful table handling for Attributes
local record Feature
    Type: string
    Description: string
    Attributes: {string:{string}} -- MultiMap
    SequenceHash: string
    SequenceHashFunction: string
    Sequence: string
    Location: Location
    ParentSequence: Genbank

    -- Methods that need careful table handling
    get_sequence: function(self: Feature): (string, string)
    store_sequence: function(self: Feature): (string, string) 
    copy: function(self: Feature): Feature
end

-- Main Genbank record
local record Genbank is bio.Data
    where self.format == "GENBANK"
    
    Meta: Meta
    Features: {Feature}
    Sequence: string

    -- Methods that need careful table handling
    store_feature_sequences: function(self: Genbank): string
    add_feature: function(self: Genbank, feature: Feature): string
    write_to: function(self: Genbank, writer: bio.Writer): (integer, string)
end

local record GenbankParser is bio.Parser
    where self.format == "GENBANK"

    reader: bio.BufferedReader
    parameters: ParseLoopParameters -- We also need to define this
    
    -- Methods required by bio.Parser interface
    header: function(self: GenbankParser): (bio.Header, string) 
    next: function(self: GenbankParser): (Genbank, string)
    get_format: function(self: GenbankParser): bio.Format
end

-- Also need to define ParseLoopParameters
local record ParseLoopParameters
    -- Current parsing context
    parsing_mode: string  -- "metadata", "features", "sequence"
    in_feature: boolean
    in_qualifier: boolean
    line_number: integer
    
    -- Current data accumulation
    current_line: string
    prev_line: string
    feature_text: {string}
    sequence_builder: {string}
    metadata_buffer: {string}
    current_tag: string
	feature_start: boolean
    
    -- Feature parsing state
    current_feature: Feature
    current_qualifier_key: string
    current_qualifier_value: string
    features_accumulated: {Feature}
    
    -- Result building
    result: Genbank

    -- Methods
    init: function(self: ParseLoopParameters)
    save_last_attribute: function(self: ParseLoopParameters)
end

-- Implement the GenbankParser methods
function GenbankParser:header(): (bio.Header, string)
    return {format = "GENBANK"}, nil
end

function GenbankParser:get_format(): bio.Format
    return "GENBANK"
end

-- Helper functions for deep table copying
local function deep_copy_location(loc: Location): Location
    if not loc then return nil end
    
    local copy: Location = {
        Start = loc.Start,
        End = loc.End,
        Complement = loc.Complement,
        Join = loc.Join,
        FivePrimePartial = loc.FivePrimePartial,
        ThreePrimePartial = loc.ThreePrimePartial,
        GbkLocationString = loc.GbkLocationString,
        SubLocations = {}
    }

    -- Deep copy sub-locations
    if loc.SubLocations then
        for _, subloc in ipairs(loc.SubLocations) do
            table.insert(copy.SubLocations, deep_copy_location(subloc))
        end
    end

    return copy
end

local function deep_copy_attributes(attrs: {string:{string}}): {string:{string}}
    if not attrs then return {} end
    
    local copy: {string:{string}} = {}
    for k, v in pairs(attrs) do
        copy[k] = {}
        for _, val in ipairs(v) do
            table.insert(copy[k], val)
        end
    end
    return copy
end

-- Constants
local QUALIFIER_INDENT = 21
local FEATURE_INDENT = 5
--local LINE_LENGTH = 80

-- String cleaning helpers
local function clean_whitespace(s: string): string
    local result, _ = s:gsub("^%s+", "")
    result, _ = result:gsub("%s+$", "")
    result, _ = result:gsub("%s+", " ")
    return result
end

local function count_leading_spaces(line: string): integer
    local trimmed, _ = line:gsub("^%s+", "")
    return #line - #trimmed
end

-- Location parsing implementation
local function parse_simple_location(loc_str: string): (Location, string)
    -- Create new location with explicit initialization
    local location: Location = {
        Start = 0,
        End = 0,
        Complement = false,
        Join = false,
        FivePrimePartial = false,
        ThreePrimePartial = false,
        GbkLocationString = loc_str,
        SubLocations = {}
    }

    -- Handle partial indicators
    if loc_str:find("<") then
        location.FivePrimePartial = true
        loc_str = loc_str:gsub("<", "")
    end
    if loc_str:find(">") then
        location.ThreePrimePartial = true
        loc_str = loc_str:gsub(">", "")
    end

    -- Parse the location numbers
    if not loc_str:find("%.%.") then
        -- Single position
        local pos = tonumber(loc_str) as integer
        if not pos then
            return nil, "Invalid position in location string"
        end
        -- Convert to 0-based indexing
        location.Start = pos - 1
        location.End = pos
    else
        -- Range
        local start_str, end_str = loc_str:match("(%d+)%.%.(%d+)")
        if not start_str or not end_str then
            return nil, "Invalid range format"
        end
        local start = tonumber(start_str) as integer
        local end_pos = tonumber(end_str) as integer
        if not start or not end_pos then
            return nil, "Invalid range numbers"
        end
        -- Convert to 0-based indexing
        location.Start = start - 1
        location.End = end_pos
    end

    return location, nil
end

-- Split location string handling nested parentheses
local function split_location_parts(inner: string): ({string}, string)
    local parts: {string} = {}
    local current = ""
    local paren_count = 0
    
    for i = 1, #inner do
        local char = inner:sub(i,i)
        if char == "(" then
            paren_count = paren_count + 1
            current = current .. char
        elseif char == ")" then
            paren_count = paren_count - 1
            current = current .. char
        elseif char == "," and paren_count == 0 then
            -- Only split on commas at top level
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

    -- Don't forget the last part
    local cleaned = clean_whitespace(current)
    if #cleaned > 0 then
        table.insert(parts, cleaned)
    end

    return parts, nil
end

-- Recursive location parsing with proper table handling
local function parse_location_recursive(expr: string, depth: integer): (Location, string)
    if depth > 10 then
        return nil, "Maximum recursion depth exceeded"
    end

    -- Clean input
    local cleaned = clean_whitespace(expr)
    if #cleaned == 0 then
        return nil, "Empty location string"
    end

    -- Check for basic location (no parentheses)
    if not cleaned:find("%(") then
        return parse_simple_location(cleaned)
    end

    -- Parse complex location (join/complement)
    local command, inner = cleaned:match("^(%w+)%((.+)%)$")
    if not command or not inner then
        return nil, "Invalid location format"
    end

    -- Create new location with explicit initialization
    local location: Location = {
        Start = 0,
        End = 0,
        Complement = false,
        Join = false,
        FivePrimePartial = false,
        ThreePrimePartial = false,
        GbkLocationString = expr,
        SubLocations = {}
    }

    if command == "complement" then
        local subloc, err = parse_location_recursive(inner, depth + 1)
        if err then
            return nil, err
        end
        
        -- Create new complement location
        local complement_loc = {
            Start = subloc.Start,
            End = subloc.End,
            Complement = true,
            Join = subloc.Join,
            FivePrimePartial = subloc.FivePrimePartial,
            ThreePrimePartial = subloc.ThreePrimePartial,
            GbkLocationString = expr,
            SubLocations = {}
        }
        
        -- Deep copy sublocations
        for _, sub in ipairs(subloc.SubLocations) do
            table.insert(complement_loc.SubLocations, deep_copy_location(sub))
        end
        
        table.insert(location.SubLocations, complement_loc)
        return location, nil

    elseif command == "join" then
        location.Join = true
        
        -- Split parts handling nested parentheses
        local parts, err = split_location_parts(inner)
        if err then
            return nil, err
        end

        -- Parse each part recursively
        for _, part in ipairs(parts) do
            local subloc, err2 = parse_location_recursive(part, depth + 1)
            if err2 then
                return nil, err2
            end
            -- Add deep copy of sublocation
            table.insert(location.SubLocations, deep_copy_location(subloc))
        end

        return location, nil
    else
        return nil, "Unknown location command: " .. command
    end
end

local function parse_location(location_string: string): (Location, string)
    if not location_string or #location_string == 0 then
        return nil, "Empty location string"
    end

    local location, err = parse_location_recursive(location_string, 0)
    if err then
        return nil, err
    end

    -- Simplify single sublocation case
    if location.Start == 0 and location.End == 0 
       and not location.Join and not location.Complement 
       and #location.SubLocations == 1 then
        -- Create new copy of the sublocation
        local subloc = location.SubLocations[1]
        location = deep_copy_location(subloc)
    end

    location.GbkLocationString = location_string
    return location, nil
end

-- Feature handling with proper table management
local function finalize_current_feature(state: ParseLoopParameters)
    --if state.current_feature.Type ~= "" then
        -- Parse location string into actual location structure
        local loc_str = state.current_feature.Location.GbkLocationString 
        if loc_str and #loc_str > 0 then
            local parsed_loc, perr = parse_location(loc_str)
            if not perr and parsed_loc then
                -- Preserve the GbkLocationString while updating numeric fields
                parsed_loc.GbkLocationString = loc_str
                state.current_feature.Location = parsed_loc
            end
        end

        -- Create deep copy of the parsed feature
        local feature_copy = {
            Type = state.current_feature.Type,
            Description = state.current_feature.Description,
            Attributes = deep_copy_attributes(state.current_feature.Attributes),
            Location = deep_copy_location(state.current_feature.Location),
            Sequence = state.current_feature.Sequence,
            SequenceHash = state.current_feature.SequenceHash,
            SequenceHashFunction = state.current_feature.SequenceHashFunction,
            ParentSequence = nil  -- Will be set when adding to final Genbank
        }
        
        table.insert(state.features_accumulated, feature_copy)
        
        -- Reset current feature
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
                GbkLocationString = ""
            },
            Sequence = "",
            SequenceHash = "",
            SequenceHashFunction = ""
        }
        
        state.current_qualifier_key = ""
        state.current_qualifier_value = ""
        state.in_qualifier = false
    --end
end

-- Parse a qualifier line, being careful about string handling
local function parse_qualifier(line: string): (string, string, boolean)
    local trimmed = clean_whitespace(line)
    if #trimmed < 2 or trimmed:sub(1,1) ~= "/" then
        return "", "", false
    end
    
    -- Remove leading slash
    trimmed = trimmed:sub(2)
    
    -- Split on equals if present
    local equals_pos = trimmed:find("=")
    if not equals_pos then
        return clean_whitespace(trimmed), "", true
    end
    
    local key = clean_whitespace(trimmed:sub(1, equals_pos-1))
    local value = trimmed:sub(equals_pos+1)
    
    -- Handle quoted values
    if value:sub(1,1) == '"' then
        value = value:sub(2, -2) -- Remove quotes
    end
    
    return key, clean_whitespace(value), false
end

-- Location parsing helpers
local function is_location_syntax(text: string): boolean
    if not text or #text == 0 then return false end
    
    -- Check for location indicators
    if text:find("%.%.") then return true end    -- range
    if text:find("join%(") then return true end  -- join
    if text:find("complement%(") then return true end -- complement
    if text:find("[<>]") then return true end    -- partial
    if text:match("^%d+$") then return true end  -- single position
    
    return false
end

-- Constants used in parsing
local METADATA_SECTIONS: {string:boolean} = {
    ["LOCUS"] = true,
    ["DEFINITION"] = true,
    ["ACCESSION"] = true,
    ["VERSION"] = true,
    ["KEYWORDS"] = true,
    ["SOURCE"] = true,
    ["REFERENCE"] = true,
    ["COMMENT"] = true,
    ["FEATURES"] = true,
    ["ORIGIN"] = true
}

-- Handle metadata line processing with careful state management
local function process_metadata_line(state: ParseLoopParameters, line: string): string
    if #line == 0 then
        return "unexpected empty metadata line"
    end

    -- Check if this is a new metadata section
    if line:sub(1,1) ~= " " or state.current_tag == "FEATURES" then
        -- Process the previous metadata section if it exists
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
                -- Special handling for SOURCE section
                local source_parts = {}
                local organism = ""
                local taxonomy: {string} = {}
                
                for _, line_data in ipairs(state.metadata_buffer) do
                    if line_data:find("ORGANISM") then
                        local _, org_start = line_data:find("ORGANISM%s*")
                        if org_start then
                            organism = clean_whitespace(line_data:sub(org_start + 1))
                        end
                    elseif line_data:find(";") then
                        -- Handle taxonomy line
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
                -- Create new reference object
                local ref: Reference = {
                    Authors = "",
                    Title = "",
                    Journal = "",
                    PubMed = "",
                    Remark = "",
                    Range = "",
                    Consortium = ""
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
                        -- Continue previous field
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
                
                -- Extract range if present in first line
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
                -- Store in Other if not a recognized section
                state.result.Meta.Other[state.current_tag] = content
            end
        end

        -- Start new metadata section
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
        -- Continue current metadata section
        table.insert(state.metadata_buffer, line)
    end

    return nil
end

-- Process a feature line with proper state management
local function process_feature_line(state: ParseLoopParameters, line: string): string
    local trimmed = clean_whitespace(line)
    if #trimmed == 0 then return nil end

    local indent = count_leading_spaces(line)
    
    -- Check for section transitions
    if line:find("BASE COUNT") then
        -- Parse base counts
        for count_str, base in line:gmatch("(%d+)%s+([a-zA-Z])") do
            local count = tonumber(count_str) as integer
            if count then
                table.insert(state.result.Meta.BaseCount, {
                    Base = base,
                    Count = count
                })
            end
        end
        return nil
    elseif line:find("ORIGIN") or line:find("CONTIG") then
        state.parsing_mode = "sequence"
        finalize_current_feature(state)
        
        -- Process accumulated features
        for _, feature in ipairs(state.features_accumulated) do
            -- Copy feature to avoid reference issues
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
				copy = Feature.copy
            }
			new_feature:store_sequence()
            table.insert(state.result.Features, new_feature)
        end
        
        if line:find("CONTIG") then
            local contig_parts: {string} = {}
            for part in line:gmatch("%S+") do
                table.insert(contig_parts, part)
            end
            if #contig_parts > 1 then
                table.remove(contig_parts, 1) -- Remove "CONTIG"
                state.result.Meta.Other["CONTIG"] = table.concat(contig_parts, " ")
            end
        end
        return nil
    end

	-- Handle new feature or continuation
	if (indent <= QUALIFIER_INDENT and count_leading_spaces(state.prev_line) > indent) or state.feature_start then
        -- Save previous feature if exists
		if state.feature_start == nil then
			finalize_current_feature(state)
		else
			state.feature_start = nil
		end
        
        -- Parse new feature line
        local parts: {string} = {}
        for part in trimmed:gmatch("%S+") do
            table.insert(parts, part)
        end
        
        if #parts < 2 then
            return "malformed feature line"
        end

        -- Initialize new feature with explicit table creation
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
                GbkLocationString = parts[#parts]
            },
            Sequence = "",
            SequenceHash = "",
            SequenceHashFunction = "",
            ParentSequence = nil -- Will be set when adding to final Genbank
        }
        
    elseif line:sub(indent + 1, indent + 1) == "/" then
        -- Handle qualifier line
        local key, value, is_empty = parse_qualifier(trimmed)
        if #key > 0 then
            -- Initialize attribute array if needed
            if not state.current_feature.Attributes[key] then
                state.current_feature.Attributes[key] = {}
            end
            
            if is_empty then
                -- Handle empty qualifiers (like /pseudo)
                table.insert(state.current_feature.Attributes[key], "")
            else
                -- Store non-empty qualifier
                if value and #value > 0 then
                    table.insert(state.current_feature.Attributes[key], value)
                end
            end
            
            state.current_qualifier_key = key
            state.current_qualifier_value = value
            state.in_qualifier = true
        end
        
    else
        -- Handle continuation line
        if indent >= QUALIFIER_INDENT then
            if is_location_syntax(trimmed) then
                -- Append to location string
                if #state.current_feature.Location.GbkLocationString > 0 then
                    state.current_feature.Location.GbkLocationString = 
                        state.current_feature.Location.GbkLocationString .. " " .. trimmed
                else
                    state.current_feature.Location.GbkLocationString = trimmed
                end
            elseif state.in_qualifier and state.current_qualifier_key ~= "" then
                -- Continue qualifier value
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

-- Location string building with proper recursion handling
local function build_location_string(location: Location): string
    -- Helper to format position with partial indicators
    local function build_position_string(pos: integer, partial: boolean): string
        local s = tostring(pos + 1) -- Convert back to 1-based
        if partial then
            s = "<" .. s
        end
        return s
    end

    -- Helper to format range with partial indicators
    local function build_range_string(start: integer, stop: integer, five_prime: boolean, three_prime: boolean): string
        local s = build_position_string(start, five_prime) .. 
                 ".." .. tostring(stop)
        if three_prime then
            s = s .. ">"
        end
        return s
    end

    -- Handle join operation
    if location.Join then
        local parts: {string} = {}
        for _, subloc in ipairs(location.SubLocations) do
            table.insert(parts, build_location_string(subloc))
        end
        return "join(" .. table.concat(parts, ",") .. ")"
    
    -- Handle complement operation
    elseif location.Complement then
        -- Create temporary location without complement flag to avoid recursion
        local inner_loc = {
            Start = location.Start,
            End = location.End,
            Complement = false,
            Join = location.Join,
            FivePrimePartial = location.FivePrimePartial,
            ThreePrimePartial = location.ThreePrimePartial,
            GbkLocationString = "",
            SubLocations = location.SubLocations
        }
        return "complement(" .. build_location_string(inner_loc) .. ")"
    
    -- Handle simple location
    else
        return build_range_string(
            location.Start,
            location.End,
            location.FivePrimePartial,
            location.ThreePrimePartial
        )
    end
end

-- Feature sequence handling
local function get_feature_sequence(feature: Feature, location: Location): (string, string)
    if not feature.ParentSequence then
        return "", "Feature has no parent sequence"
    end

    if not feature.ParentSequence.Sequence then
        return "", "Parent sequence is missing"
    end

    local sequence_parts: {string} = {}
    
    -- Handle sublocations recursively
    if location.SubLocations and #location.SubLocations > 0 then
        for _, subloc in ipairs(location.SubLocations) do
            local seq, err = get_feature_sequence(feature, subloc)
            if err then
                return "", err
            end
            table.insert(sequence_parts, seq)
        end
    else
        -- Handle simple location
        if location.Start >= 0 and location.End > location.Start and 
           location.End <= #feature.ParentSequence.Sequence then
            -- Add 1 to Start for Lua's 1-based string indexing
            local seq = feature.ParentSequence.Sequence:sub(location.Start + 1, location.End)
            table.insert(sequence_parts, seq)
        else
            return "", "Invalid location bounds"
        end
    end

    -- Combine sequences
    local sequence = table.concat(sequence_parts)
    
    -- Handle complement if needed
    if location.Complement then
        sequence = transform.reverse_complement(sequence)
    end

    return sequence, nil
end

-- Feature methods implementation
function Feature:get_sequence(): (string, string)
    return get_feature_sequence(self, self.Location)
end

function Feature:store_sequence(): (string, string)
    if self.Sequence ~= "" then
        return self.Sequence, nil
    end
    
    local seq, err = get_feature_sequence(self, self.Location)
    if err == nil then
        self.Sequence = seq
    end
    return seq, err
end

function Feature:copy(): Feature
    -- Create new feature with deep copies of all fields
    local copy: Feature = {
        Type = self.Type,
        Description = self.Description,
        Attributes = deep_copy_attributes(self.Attributes),
        Location = deep_copy_location(self.Location),
        Sequence = self.Sequence,
        SequenceHash = self.SequenceHash,
        SequenceHashFunction = self.SequenceHashFunction,
        ParentSequence = self.ParentSequence, -- Note: keeping same reference to parent
       
        -- Copy methods
        get_sequence = self.get_sequence,
        store_sequence = self.store_sequence,
        copy = self.copy
    }
    
    return copy
end

-- Genbank feature sequence methods
function Genbank:store_feature_sequences(): string
    for _, feature in ipairs(self.Features) do
        local _, err = feature:store_sequence()
        if err then
            return err
        end
    end
    return nil
end

function Genbank:add_feature(feature: Feature): string
    -- Create deep copy of feature before adding
    local new_feature = feature:copy()
    new_feature.ParentSequence = self
    table.insert(self.Features, new_feature)
    return nil
end

-- Build feature string for output
local function build_feature_string(feature: Feature): string
    if not feature then return "" end

    -- Calculate spacing
    local type_space = 16 - #feature.Type
    if type_space < 0 then type_space = 0 end
    local white_space_trail = string.rep(" ", type_space)
    
    -- Get location string
    local location = feature.Location.GbkLocationString
    if #location == 0 then
        location = build_location_string(feature.Location)
    end

    -- Build feature header
    local feature_string = string.rep(" ", FEATURE_INDENT) .. 
                          feature.Type .. 
                          white_space_trail .. 
                          location .. "\n"

    -- Add qualifiers if present
    if feature.Attributes then
        -- Sort keys for consistent output
        local keys: {string} = {}
        for k in pairs(feature.Attributes) do
            table.insert(keys, k)
        end
        table.sort(keys)

        -- Add each qualifier
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

-- Locus parsing implementation
function genbank.parse_locus(locus_string: string): Locus
    local locus: Locus = {
        Name = "",
        SequenceLength = "",
        MoleculeType = "",
        GenbankDivision = "",
        ModificationDate = "",
        SequenceCoding = "",
        Circular = false
    }

    -- Split and filter locus parts
    local parts: {string} = {}
    for part in locus_string:gmatch("%S+") do
        table.insert(parts, part)
    end

    if #parts < 2 then
        return locus
    end

    -- Get name (always second token after LOCUS)
    locus.Name = parts[2]

    -- Find base sequence length and coding
    local base_sequence = locus_string:match("%s+(%d+)%s+([a-zA-Z][a-zA-Z])%s+")
    if base_sequence then
        local length, coding = base_sequence:match("(%d+)%s+([a-zA-Z][a-zA-Z])")
        if length and coding then
            locus.SequenceLength = length
            locus.SequenceCoding = coding
        end
    end

    -- Find molecule type from predefined list
    for _, mol_type in ipairs({
        "DNA", "genomic DNA", "genomic RNA", "mRNA", "tRNA", "rRNA",
        "other RNA", "other DNA", "transcribed RNA", "viral cRNA",
        "unassigned DNA", "unassigned RNA"
    }) do
        if locus_string:find("%s+" .. mol_type .. "%s+") then
            locus.MoleculeType = mol_type
            break
        end
    end

    -- Check circularity
    if locus_string:find("%s+circular%s+") then
        locus.Circular = true
    end

    -- Find GenBank division
    for _, division in ipairs({
        "PRI", "ROD", "MAM", "VRT", "INV", "PLN", "BCT", "VRL",
        "PHG", "SYN", "UNA", "EST", "PAT", "STS", "GSS", "HTG",
        "HTC", "ENV"
    }) do
        if locus_string:find("%s+" .. division .. "%s+") then
            locus.GenbankDivision = division
            break
        end
    end

    -- Find modification date
    local date = locus_string:match("%d%d%-[A-Z][A-Z][A-Z]-%d%d%d%d")
    if date then
        locus.ModificationDate = date
    end

    return locus
end

-- Text wrapping utility
local function wrap_text(text: string, width: integer): {string}
    local lines: {string} = {}
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

-- Writer implementation
function Genbank:write_to(writer: bio.Writer): (integer, string)
    local written: integer = 0

    -- Helper to write and track bytes
    local function write_line(line: string): string
        local n, err = writer:write(line)
        if err then
            return err
        end
        written = written + n
        return nil
    end

    -- Write LOCUS line
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

    -- Write standard meta sections
    local meta_sections = {
        {"DEFINITION", self.Meta.Definition},
        {"ACCESSION", self.Meta.Accession},
        {"VERSION", self.Meta.Version},
        {"KEYWORDS", self.Meta.Keywords},
        {"SOURCE", self.Meta.Source},
        {"  ORGANISM", self.Meta.Organism}
    }

    for _, section in ipairs(meta_sections) do
        if section[2] and #section[2] > 0 then
            -- Build metadata string with proper wrapping
            local indent = section[1] == "  ORGANISM" and 12 or #section[1]
            local content = section[2]
            local wrapped = wrap_text(content, 80 - indent - 1)
            
            -- Write first line
            err = write_line(section[1] .. string.rep(" ", 12 - indent) .. wrapped[1] .. "\n")
            if err then return written, err end

            -- Write continuation lines if any
            for i = 2, #wrapped do
                err = write_line(string.rep(" ", 12) .. wrapped[i] .. "\n")
                if err then return written, err end
            end
        end
    end

    -- Write taxonomy if present
    if #self.Meta.Taxonomy > 0 then
        local taxonomy = table.concat(self.Meta.Taxonomy, "; ") .. "."
        local wrapped = wrap_text(taxonomy, 68)
        for _, line in ipairs(wrapped) do
            err = write_line(string.rep(" ", 12) .. line .. "\n")
            if err then return written, err end
        end
    end

    -- Write references
    for i, ref in ipairs(self.Meta.References) do
        -- Reference header
        local ref_header = string.format("REFERENCE   %d", i)
        if ref.Range and #ref.Range > 0 then
            ref_header = ref_header .. "  " .. ref.Range
        end
        err = write_line(ref_header .. "\n")
        if err then return written, err end

        -- Reference fields
        local ref_fields = {
            {"  AUTHORS", ref.Authors},
            {"  TITLE", ref.Title},
            {"  JOURNAL", ref.Journal},
            {"  PUBMED", ref.PubMed},
            {"  CONSRTM", ref.Consortium}
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
            -- Special handling for COMMENT to match GenBank format
            if tag == "COMMENT" then
                -- Wrap the comment text with proper indentation
                local wrapped = wrap_text(content, 68)
                err = write_line("COMMENT             " .. wrapped[1] .. "\n")
                if err then return written, err end
                
                -- Write continuation lines if any
                for i = 2, #wrapped do
                    err = write_line(string.rep(" ", 21) .. wrapped[i] .. "\n")
                    if err then return written, err end
                end
            else
                -- Handle other metadata sections with standard formatting
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

    -- Write features section
    err = write_line("FEATURES             Location/Qualifiers\n")
    if err then return written, err end

    for _, feature in ipairs(self.Features) do
        err = write_line(build_feature_string(feature))
        if err then return written, err end
    end

    -- Write base count if present
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

    -- Write sequence section
    err = write_line("ORIGIN\n")
    if err then return written, err end

    -- Format sequence in blocks of 60
    for i = 1, #self.Sequence, 60 do
        -- Format line number
        local line_num = tostring(i)
        local leading_spaces = string.rep(" ", 9 - #line_num)
        err = write_line(leading_spaces .. line_num .. " ")
        if err then return written, err end

        -- Write sequence block in groups of 10
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

    -- Write file terminator
    err = write_line("//\n")
    if err then return written, err end

    return written, nil
end

function GenbankParser:next(): (Genbank, string)
    -- Initialize state for new record if needed
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
                    GbkLocationString = ""
                },
                Sequence = "",
                SequenceHash = "",
                SequenceHashFunction = ""
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
                        Circular = false
                    },
                    References = {},
                    BaseCount = {},
                    Other = {},
                    Name = "",
                    SequenceHash = "",
                    SequenceHashFunction = ""
                },
                Features = {},
                Sequence = "",
                format = "GENBANK"
            }
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
                    -- Finalize last record
                    self.parameters.result.Sequence = table.concat(self.parameters.sequence_builder)
                    local result = self.parameters.result
                    -- Reset parser state for next record
                    self.parameters = nil
                    return result, nil
                else
                    return nil, "Unexpected EOF in " .. self.parameters.parsing_mode .. " section"
                end
            end
            return nil, err
        end

        -- Update line tracking
        self.parameters.prev_line = self.parameters.current_line
        self.parameters.current_line = line
        self.parameters.line_number = self.parameters.line_number + 1

        -- Handle initial LOCUS detection
        if not found_locus then
            if line:find("LOCUS") then
                found_locus = true
                self.parameters.result.Meta.Locus = genbank.parse_locus(line)
            end
        else
            -- Main parsing state machine
			local err2: string
            if self.parameters.parsing_mode == "metadata" then
                err2 = process_metadata_line(self.parameters, line)
            elseif self.parameters.parsing_mode == "features" then
                err2 = process_feature_line(self.parameters, line)
            elseif self.parameters.parsing_mode == "sequence" then
                -- Handle sequence section
                if line:sub(1,2) == "//" then
                    -- End of record reached
                    self.parameters.result.Sequence = table.concat(self.parameters.sequence_builder)
                    local result = self.parameters.result
					result.write_to = Genbank.write_to
                    -- Reset parser state for next record
                    self.parameters = nil
                    return result, nil
                else
                    -- Clean and accumulate sequence data
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

-- Create a new GenbankParser
function genbank.new_parser(reader: bio.Reader, max_line_size: integer): GenbankParser
    if not reader then
        error("reader cannot be nil")
    end
    
    -- Default max line size if not provided
    if not max_line_size or max_line_size < 1 then
        max_line_size = bio.DEFAULT_MAX_LENGTHS["GENBANK"]
    end

    -- Create buffered reader if needed
    local buffered_reader = bio.new_buffered_reader(reader, max_line_size)

    -- Create and return new parser instance
    local parser: GenbankParser = {
        reader = buffered_reader,
        parameters = nil,
        format = "GENBANK",
        
        -- Methods already defined
        header = GenbankParser.header,
        next = GenbankParser.next,
        get_format = GenbankParser.get_format
    }

    return parser
end

return genbank
