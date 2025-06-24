-- build.lua
-- Builds both dnadesign.tl (implementation) and dnadesign.d.tl (type definitions)

-- Files in dependency order with their module names
local files = {
    { path = "../rng.tl", module = "rng"},
    { path = "src/hash.tl", module = "hash" },
    { path = "src/transform.tl", module = "transform" },
    { path = "src/align.tl", module = "align" },
    { path = "src/mash.tl", module = "mash" },
    { path = "src/seqhash.tl", module = "seqhash"},
    { path = "src/primers.tl", module = "primers"},
    { path = "src/pcr.tl", module = "pcr"},
    { path = "src/bio/bio.tl", module = "bio"},
    { path = "src/bio/fasta.tl", module = "fasta"},
    { path = "src/bio/fastq.tl", module = "fastq"},
    { path = "src/bio/pileup.tl", module = "pileup"},
    { path = "src/bio/sam.tl", module = "sam"},
    { path = "src/bio/slow5.tl", module = "slow5"},
    { path = "src/bio/genbank.tl", module = "genbank"},
    { path = "src/fragment_frequencies.tl", module = "fragment_frequencies"},
    { path = "src/fragment.tl", module = "fragment"},
    { path = "src/clone.tl", module = "clone"},
    { path = "src/codon.tl", module = "codon"},
    { path = "src/fix.tl", module = "fix"},
}

-- Initialize both output files with headers
local combined = "-- dnadesign.tl\n\n"
local type_defs = "-- dnadesign.d.tl\n\n"

-- Process each module file
for _, file in ipairs(files) do
    -- Read the file content
    local f = io.open(file.path, "r")
    if not f then error("Could not open file: " .. file.path) end
    local content = f:read("*a")
    f:close()

    -- Extract module documentation for type definitions
    local doc_start = content:find("%-%-%[%[ Module")
    if doc_start then
        local doc_end = content:find("%]%]", doc_start)
        if doc_end then
            type_defs = type_defs .. content:sub(doc_start, doc_end + 2) .. "\n\n"
        end
    end

	-- Find complete record and interface definitions (handling inline comments)
	for type_def in content:gmatch("(\nlocal%s+(interface%s+[%w_]+.-\nend))") do
        type_def = type_def:sub(2)
        type_defs = type_defs .. type_def .. "\n\n"
    end
	for type_def in content:gmatch("(\nlocal%s+(enum%s+[%w_]+.-\nend))") do
        type_def = type_def:sub(2)
        type_defs = type_defs .. type_def .. "\n\n"
    end
	for type_def in content:gmatch("(\nlocal%s+(record%s+[%w_]+.-\nend))") do
	    type_def = type_def:sub(2)
	    type_defs = type_defs .. type_def .. "\n\n"
	end

    -- Process implementation file
    -- Comment out require statements
    content = content:gsub('(local%s+[%w_]+%s*=%s*require%(["\'].-["\']%)[^\n]*)', '-- %1')
    -- Remove the return statement
    content = content:gsub(string.format("(\n)return %s", file.module), "%1")
    combined = combined .. content .. "\n\n"
end

-- Add the final return type definition
type_defs = type_defs .. "local record dnadesign\n"
for _, file in ipairs(files) do
	if file.module ~= "fragment_frequencies" then
		type_defs = type_defs .. string.format("    %s: %s\n", file.module, file.module)
	end
end
type_defs = type_defs .. "end\n\nreturn dnadesign\n"

-- Add the final return implementation
combined = combined .. "-- Main return table combining all modules\nreturn {\n"
for _, file in ipairs(files) do
	combined = combined .. string.format("    %s = %s,\n", file.module, file.module)
end
combined = combined .. "}\n"

-- Write both output files
local out_file = io.open("dnadesign.tl", "w")
out_file:write(combined)
out_file:close()

local def_file = io.open("dnadesign.d.tl", "w")
def_file:write(type_defs)
def_file:close()

-- Run teal compiler and tests
os.execute("tl check dnadesign.tl")
os.execute("tl gen dnadesign.tl")
os.execute("busted --lua=luajit")
