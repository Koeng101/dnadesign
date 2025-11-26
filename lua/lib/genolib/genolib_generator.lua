local dnadesign = dofile("../../dnadesign.lua")
local bio = dnadesign.bio
local fasta = dnadesign.fasta
local mash = dnadesign.mash
local hash = dnadesign.hash

-- Parse CSV file
local function parse_csv(filename)
    local file = io.open(filename, "r")
    if not file then
        error("Could not open CSV file: " .. filename)
    end
    
    local metadata = {}
    local header_line = file:read("*line")
    
    for line in file:lines() do
        -- Parse CSV line: sseqid,Feature,Type,Description
        local sseqid, feature, type_field, description = line:match("([^,]+),([^,]+),([^,]+),(.+)")
        if sseqid then
            metadata[sseqid] = {
                feature = feature,
                type = type_field,
                description = description
            }
        end
    end
    
    file:close()
    return metadata
end

-- Parse FASTA file and create Mash sketches
local function parse_fasta_with_mash(filename, metadata, kmer_size)
    local file = io.open(filename, "r")
    if not file then
        error("Could not open FASTA file: " .. filename)
    end
    
    local content = file:read("*all")
    file:close()
    
    local reader = bio.new_string_reader(content)
    local parser = fasta.new_parser(reader, 32*1024)
    
    local features = {}
    
    while true do
        local record, err = parser:next()
        if err then
            if err == "EOF" then
                break
            end
            error("Got error: " .. err)
        end
        
        -- Get metadata for this sequence
		id = record.identifier
		local id_trimmed = id:match("^%s*(.-)%s*$")
        local meta = metadata[id_trimmed]
        if meta then
			-- Create containment sketch for this sequence
			local hasher = hash.new_crc32()
        	local sketch = mash.new_containment_sketch(kmer_size, record.sequence, hasher)
        	
        	-- Store feature data
        	table.insert(features, {
        	    sseqid = record.identifier,
        	    feature = meta.feature,
        	    type = meta.type,
        	    description = meta.description,
        	    sequence = record.sequence,
        	    mash = {
        	        kmer_size = sketch.kmer_size,
        	        sketch_size = sketch.sketch_size,
        	        sketches = sketch.sketches
        	    }
        	})
		end
    end
    
    return features
end

-- Generate Teal type definition file
local function generate_teal_file(features, output_filename)
    local file = io.open(output_filename, "w")
    if not file then
        error("Could not create output file: " .. output_filename)
    end
    
    -- Write header and type definitions
    file:write([[
-- genolib.tl
-- Auto-generated feature library with cached Mash sketches
-- DO NOT EDIT - regenerate using genolib_generator.lua

-- Minimal Mash sketch type definition
local record Mash
    kmer_size: integer
    sketch_size: integer
    sketches: {integer}
end

-- Feature record with metadata and Mash sketch
local record Feature
    sseqid: string        -- Sequence identifier from FASTA
    feature: string       -- Feature name
    feature_type: string          -- Feature type (e.g., CDS, promoter)
    description: string   -- Feature description
    sequence: string      -- DNA/RNA sequence
    mash: Mash           -- Pre-computed Mash containment sketch
end

local record genolib
    features: {Feature}
    containment: function(feature_mash: Mash, query_mash: Mash): number
end

-- Containment function (calculates fraction of hashes in a that are present in b)
local function containment(a: Mash, b: Mash): number
    -- Handle empty sketches
    if #a.sketches == 0 or #b.sketches == 0 then
        return 0
    end
    
    local i, j = 1, 1
    local same_hashes = 0
    
    -- Iterate over sorted sketches
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

genolib.containment = containment

-- Feature database
genolib.features = {
]])
    
    -- Write feature data
    for i, feature in ipairs(features) do
        file:write("    {\n")
        file:write(string.format("        sseqid = %q,\n", feature.sseqid))
        file:write(string.format("        feature = %q,\n", feature.feature))
        file:write(string.format("        feature_type = %q,\n", feature.type))
        file:write(string.format("        description = %q,\n", feature.description))
        file:write(string.format("        sequence = %q,\n", feature.sequence))
        file:write("        mash = {\n")
        file:write(string.format("            kmer_size = %d,\n", feature.mash.kmer_size))
        file:write(string.format("            sketch_size = %d,\n", feature.mash.sketch_size))
        file:write("            sketches = {")
        
        -- Write sketch hashes
        for j, hash_val in ipairs(feature.mash.sketches) do
            if j > 1 then file:write(", ") end
            if (j - 1) % 10 == 0 and j > 1 then
                file:write("\n                ")
            end
            file:write(string.format("%d", hash_val))
        end
        
        file:write("}\n")
        file:write("        }\n")
        file:write("    }")
        
        if i < #features then
            file:write(",\n")
        else
            file:write("\n")
        end
    end
    
    file:write("}\n\nreturn genolib\n")
    file:close()
end

-- Main execution
local function main()
    print("GenOlib Generator - Creating cached Mash feature library")
    print("=======================================================")
    
    -- Configuration
    local csv_file = "snapgene.csv"
    local fasta_file = "genolib.fasta"
    local output_file = "genolib.tl"
    local kmer_size = 20  -- Adjust as needed
    
    print("Step 1: Parsing CSV metadata...")
    local metadata = parse_csv(csv_file)
    print(string.format("  Found metadata for %d features", 
        (function() local count = 0; for _ in pairs(metadata) do count = count + 1 end; return count end)()))
    
    print("Step 2: Parsing FASTA and creating Mash sketches...")
    local features = parse_fasta_with_mash(fasta_file, metadata, kmer_size)
    print(string.format("  Processed %d features with Mash sketches", #features))
    
    print("Step 3: Generating Teal type definition file...")
    generate_teal_file(features, output_file)
    print(string.format("  Generated %s", output_file))
    
    print("\nDone! You can now use 'local genolib = require(\"genolib\")' in your code.")
    print("Example usage:")
    print("  local my_mash = {...}  -- Your query mash sketch")
    print("  for _, feature in ipairs(genolib.features) do")
    print("    local score = genolib.containment(feature.mash, my_mash)")
    print("    if score > 0.9 then")
    print("      print(feature.feature, score)")
    print("    end")
    print("  end")
end

main()
