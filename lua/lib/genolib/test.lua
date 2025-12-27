local dnadesign = require("dnadesign")
local genolib = require("genolib")
local transform = dnadesign.transform
local mash = dnadesign.mash
local hash = dnadesign.hash

local pOpen_v3 = "TAACTATCGTCTTGAGTCCAACCCGGTAAGACACGACTTATCGCCACTGGCAGCAGCCACTGGTAACAGGATTAGCAGAGCGAGGTATGTAGGCGGTGCTACAGAGTTCTTGAAGTGGTGGCCTAACTACGGCTACACTAGAAGAACAGTATTTGGTATCTGCGCTCTGCTGAAGCCAGTTACCTTCGGAAAAAGAGTTGGTAGCTCTTGATCCGGCAAACAAACCACCGCTGGTAGCGGTGGTTTTTTTGTTTGCAAGCAGCAGATTACGCGCAGAAAAAAAGGATCTCAAGAAGGCCTACTATTAGCAACAACGATCCTTTGATCTTTTCTACGGGGTCTGACGCTCAGTGGAACGAAAACTCACGTTAAGGGATTTTGGTCATGAGATTATCAAAAAGGATCTTCACCTAGATCCTTTTAAATTAAAAATGAAGTTTTAAATCAATCTAAAGTATATATGAGTAAACTTGGTCTGACAGTTACCAATGCTTAATCAGTGAGGCACCTATCTCAGCGATCTGTCTATTTCGTTCATCCATAGTTGCCTGACTCCCCGTCGTGTAGATAACTACGATACGGGAGGGCTTACCATCTGGCCCCAGTGCTGCAATGATACCGCGAGAACCACGCTCACCGGCTCCAGATTTATCAGCAATAAACCAGCCAGCCGGAAGGGCCGAGCGCAGAAGTGGTCCTGCAACTTTATCCGCCTCCATCCAGTCTATTAATTGTTGCCGGGAAGCTAGAGTAAGTAGTTCGCCAGTTAATAGTTTGCGCAACGTTGTTGCCATTGCTACAGGCATCGTGGTGTCACGCTCGTCGTTTGGTATGGCTTCATTCAGCTCCGGTTCCCAACGATCAAGGCGAGTTACATGATCCCCCATGTTGTGCAAAAAAGCGGTTAGCTCCTTCGGTCCTCCGATCGTTGTCAGAAGTAAGTTGGCCGCAGTGTTATCACTCATGGTTATGGCAGCACTGCATAATTCTCTTACTGTCATGCCATCCGTAAGATGCTTTTCTGTGACTGGTGAGTACTCAACCAAGTCATTCTGAGAATAGTGTATGCGGCGACCGAGTTGCTCTTGCCCGGCGTCAATACGGGATAATACCGCGCCACATAGCAGAACTTTAAAAGTGCTCATCATTGGAAAACGTTCTTCGGGGCGAAAACTCTCAAGGATCTTACCGCTGTTGAGATCCAGTTCGATGTAACCCACTCGTGCACCCAACTGATCTTCAGCATCTTTTACTTTCACCAGCGTTTCTGGGTGAGCAAAAACAGGAAGGCAAAATGCCGCAAAAAAGGGAATAAGGGCGACACGGAAATGTTGAATACTCATACTCTTCCTTTTTCAATATTATTGAAGCATTTATCAGGGTTATTGTCTCATGAGCGGATACATATTTGAATGTATTTAGAAAAATAAACAAATAGGGGTTCCGCGCACCTGCACCAGTCAGTAAAACGACGGCCAGTAGTCAAAAGCCTCCGACCGGAGGCTTTTGACTTGGTTCAGGTGGAGTGGGAGTAgtcttcGCcatcgCtACTAAAagccagataacagtatgcgtatttgcgcgctgatttttgcggtataagaatatatactgatatgtatacccgaagtatgtcaaaaagaggtatgctatgaagcagcgtattacagtgacagttgacagcgacagctatcagttgctcaaggcatatatgatgtcaatatctccggtctggtaagcacaaccatgcagaatgaagcccgtcgtctgcgtgccgaacgctggaaagcggaaaatcaggaagggatggctgaggtcgcccggtttattgaaatgaacggctcttttgctgacgagaacagggGCTGGTGAAATGCAGTTTAAGGTTTACACCTATAAAAGAGAGAGCCGTTATCGTCTGTTTGTGGATGTACAGAGTGATATTATTGACACGCCCGGGCGACGGATGGTGATCCCCCTGGCCAGTGCACGTCTGCTGTCAGATAAAGTCTCCCGTGAACTTTACCCGGTGGTGCATATCGGGGATGAAAGCTGGCGCATGATGACCACCGATATGGCCAGTGTGCCGGTCTCCGTTATCGGGGAAGAAGTGGCTGATCTCAGCCACCGCGAAAATGACATCAAAAACGCCATTAACCTGATGTTCTGGGGAATATAAATGTCAGGCTCCCTTATACACAGgcgatgttgaagaccaCGCTGAGGTGTCAATCGTCGGAGCCGCTGAGCAATAACTAGCATAACCCCTTGGGGCCTCTAAACGGGTCTTGAGGGGTTTTTTGCATGGTCATAGCTGTTTCCTGAGAGCTTGGCAGGTGATGACACACATTAACAAATTTCGTGAGGAGTCTCCAGAAGAATGCCATTAATTTCCATAGGCTCCGCCCCCCTGACGAGCATCACAAAAATCGACGCTCAAGTCAGAGGTGGCGAAACCCGACAGGACTATAAAGATACCAGGCGTTTCCCCCTGGAAGCTCCCTCGTGCGCTCTCCTGTTCCGACCCTGCCGCTTACCGGATACCTGTCCGCCTTTCTCCCTTCGGGAAGCGTGGCGCTTTCTCATAGCTCACGCTGTAGGTATCTCAGTTCGGTGTAGGTCGTTCGCTCCAAGCTGGGCTGTGTGCACGAACCCCCCGTTCAGCCCGACCGCTGCGCCTTATCCGG"

-- Timing utilities
local function time_operation(name, func)
    local start = os.clock()
    local result = func()
    local elapsed = os.clock() - start
    print(string.format("[TIMING] %s: %.4f ms", name, elapsed * 1000))
    return result, elapsed
end

-- Time sketch creation
local hasher = hash.new_crc32()
local plasmid, sketch_time = time_operation("Create plasmid sketch", function()
    return mash.new_containment_sketch(20, pOpen_v3 .. pOpen_v3, hasher)
end)

local plasmid_rev, sketch_rev_time = time_operation("Create reverse complement sketch", function()
    hasher = hash.new_crc32()  -- Fresh hasher
    return mash.new_containment_sketch(20, transform.reverse_complement(pOpen_v3 .. pOpen_v3), hasher)
end)

-- Time forward containment screening
local forward_matches = 0
local forward_time = 0
print("\nForward strand screening:")
for _, feature in ipairs(genolib.features) do
    local start = os.clock()
    local score = mash.containment(feature.mash, plasmid)
    forward_time = forward_time + (os.clock() - start)
    
    if score > 0.8 then
        print(string.format("  %s: %.2f%%", feature.feature, score * 100))
        forward_matches = forward_matches + 1
    end
end
print(string.format("[TIMING] Forward screening (%d features): %.4f ms (%.4f μs per feature)", 
    #genolib.features, forward_time * 1000, forward_time * 1000000 / #genolib.features))

-- Time reverse containment screening
local reverse_matches = 0
local reverse_time = 0
print("\nReverse strand screening:")
for _, feature in ipairs(genolib.features) do
    local start = os.clock()
    local score = mash.containment(feature.mash, plasmid_rev)
    reverse_time = reverse_time + (os.clock() - start)
    
    if score > 0.8 then
        print(string.format("  %s: %.2f%%", feature.feature, score * 100))
        reverse_matches = reverse_matches + 1
    end
end
print(string.format("[TIMING] Reverse screening (%d features): %.4f ms (%.4f μs per feature)", 
    #genolib.features, reverse_time * 1000, reverse_time * 1000000 / #genolib.features))

-- Summary
print("\n=== TIMING SUMMARY ===")
print(string.format("Sketch creation:        %.4f ms", (sketch_time + sketch_rev_time) * 1000))
print(string.format("Forward screening:      %.4f ms (%d matches)", forward_time * 1000, forward_matches))
print(string.format("Reverse screening:      %.4f ms (%d matches)", reverse_time * 1000, reverse_matches))
print(string.format("Total screening time:   %.4f ms", (forward_time + reverse_time) * 1000))
print(string.format("Avg per containment:    %.2f μs", (forward_time + reverse_time) * 1000000 / (#genolib.features * 2)))
print(string.format("Total time:             %.4f ms", (sketch_time + sketch_rev_time + forward_time + reverse_time) * 1000))

z = 0
for _, feature in ipairs(genolib.features) do
	z = z + #feature.sequence
end
print(#genolib.features)
print(z)
