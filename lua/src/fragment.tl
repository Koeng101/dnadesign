--[[ Module fragment
Module fragment optimally fragments DNA for GoldenGate systems.

Optimal fragmentation is accomplished by using empirical fidelity data derived
by NEB in the paper "Enabling one-pot Golden Gate assemblies of unprecedented
complexity using data-optimized assembly design". We use the BsaI-T4 ligase
data provided in table S1.

Paper link: https://doi.org/10.1371/journal.pone.0238592
Data link: https://doi.org/10.1371/journal.pone.0238592.s001
]]

local record fragment
    -- Core functions
    set_efficiency: function({string}): number
    next_overhangs: function({string}): ({string}, {number})
    next_overhang: function({string}): string
    fragment: function(string, integer, integer, {string}): ({string}, number, string)
    fragment_with_overhangs: function(string, integer, integer, {string}, {string}): ({string}, number, string)

    -- Recursive Assembly
    record Assembly
        sequence: string
        fragments: {string}
        efficiency: number
        sub_assemblies: {Assembly}
    end

    recursive_fragment: function(string, integer, {integer}, {string}, {string}, string, string): (Assembly, string)
end

-- Import key dependencies
local transform = require("transform")
local fragment_frequencies = require("fragment_frequencies")

local function is_palindromic(sequence: string): boolean
	return sequence == transform.reverse_complement(sequence)
end

-- SetEfficiency gets the estimated fidelity rate of a given set of
-- GoldenGate overhangs.
function fragment.set_efficiency(overhangs: {string}): number
    local efficiency = 1.0
    for _, overhang in ipairs(overhangs) do
        -- Make sure the table exists
        if not fragment_frequencies[overhang] then
            fragment_frequencies[overhang] = {}
        end
        
        local n_correct = (fragment_frequencies[overhang][overhang] or 0)
        local n_total = 0
        for _, overhang2 in ipairs(overhangs) do
            n_total = n_total + (fragment_frequencies[overhang][overhang2] or 0)
            n_total = n_total + (fragment_frequencies[overhang][transform.reverse_complement(overhang2)] or 0)
        end
        if n_total ~= n_correct and n_total > 0 then -- Add check for n_total > 0
            efficiency = efficiency * (n_correct / n_total)
        end
    end
    return efficiency
end

-- NextOverhangs gets a list of possible next overhangs to use for an overhang
-- list, along with their efficiencies. This can be used for more optimal
-- fragmentation of sequences with potential degeneracy.
function fragment.next_overhangs(current_overhangs: {string}): ({string}, {number})
    -- Create map of current overhangs for quick lookup
    local current_overhang_map: {string:boolean} = {}
    for _, overhang in ipairs(current_overhangs) do
        current_overhang_map[overhang] = true
    end

    -- These 4 for loops generate all combinations of 4 base pairs
    -- checking all 256 4mer combinations for palindromes. Palindromes
    -- can cause problems in large combinatorial reactions, so are
    -- removed here.
    local bases = {"A", "T", "G", "C"}
    local overhangs_to_test: {string} = {}
    for _, base1 in ipairs(bases) do
        for _, base2 in ipairs(bases) do
            for _, base3 in ipairs(bases) do
                for _, base4 in ipairs(bases) do
                    local new_overhang = base1 .. base2 .. base3 .. base4
                    local in_current = current_overhang_map[new_overhang]
                    local in_current_reverse = current_overhang_map[transform.reverse_complement(new_overhang)]
                    if not in_current and not in_current_reverse then
                        if not is_palindromic(new_overhang) then
                            table.insert(overhangs_to_test, new_overhang)
                        end
                    end
                end
            end
        end
    end

    local efficiencies: {number} = {}
    for _, overhang in ipairs(overhangs_to_test) do
        local current_with_overhang = {}
        for _, o in ipairs(current_overhangs) do table.insert(current_with_overhang, o) end
        table.insert(current_with_overhang, overhang)
        
        local strand_efficiency = fragment.set_efficiency(current_with_overhang)
        
        local current_with_complement = {}
        for _, o in ipairs(current_overhangs) do table.insert(current_with_complement, o) end
        table.insert(current_with_complement, transform.reverse_complement(overhang))
        
        local complement_efficiency = fragment.set_efficiency(current_with_complement)
        table.insert(efficiencies, (strand_efficiency + complement_efficiency) / 2)
    end
    
    return overhangs_to_test, efficiencies
end

-- NextOverhang gets next most efficient overhang to use for a given set of
-- overhangs. This is useful for when developing a new set of standard
-- overhangs. Note: NextOverhang is biased towards high AT overhangs, but this
-- will not affect fidelity at all.
function fragment.next_overhang(current_overhangs: {string}): string
    local overhangs_to_test, efficiencies = fragment.next_overhangs(current_overhangs)
    local max_efficiency = 0.0
    local new_overhang = ""
    
    for i, overhang in ipairs(overhangs_to_test) do
        local efficiency = efficiencies[i]
        if efficiency > max_efficiency then
            max_efficiency = efficiency
            new_overhang = overhang
        end
    end
    
    return new_overhang
end

-- optimizeOverhangIteration takes in a sequence and optimally fragments it.
local function optimize_overhang_iteration(sequence: string, min_fragment_size: integer, max_fragment_size: integer, existing_fragments: {string}, exclude_overhangs: {string}, include_overhangs: {string}): ({string}, number, string)
    local recurse_max_fragment_size = max_fragment_size

    -- If the sequence is smaller than maxFragment size, stop iteration.
    if #sequence < max_fragment_size then
        table.insert(existing_fragments, sequence)
        return existing_fragments, fragment.set_efficiency(exclude_overhangs), nil
    end

    -- Make sure minFragmentSize > maxFragmentSize
    if min_fragment_size > max_fragment_size then
        return {}, 0, string.format("min_fragment_size (%d) larger than max_fragment_size (%d)", min_fragment_size, max_fragment_size)
    end

    -- Minimum lengths (given oligos) for assembly is 8 base pairs
    -- https://doi.org/10.1186/1756-0500-3-291
    -- For GoldenGate, 2 8bp oligos create 12 base pairs (4bp overhangs on two sides of 4bp),
    -- so we check for minimal size of 12 base pairs.
    if min_fragment_size < 12 then
        return {}, 0, string.format("min_fragment_size must be equal to or greater than 12. Got size of %d", min_fragment_size)
    end

    -- If our iteration is approaching the end of the sequence, that means we need to gracefully handle
    -- the end so we aren't left with a tiny fragment that cannot be synthesized. For example, if our goal
    -- is fragments of 100bp, and we have 110 base pairs left, we want each final fragment to be 55bp, not
    -- 100 and 10bp
    if #sequence < 2 * max_fragment_size then
        local max_and_min_difference = max_fragment_size - min_fragment_size
        local max_fragment_size_buffer = math.floor(#sequence / 2)
        min_fragment_size = max_fragment_size_buffer - max_and_min_difference
        if min_fragment_size < 12 then
            min_fragment_size = 12
        end
        max_fragment_size = max_fragment_size_buffer
    end

    -- Get all sets of 4 between the min and max FragmentSize
    local best_overhang_efficiency = 0.0
    local best_overhang_position = 0
    
    for overhang_offset = 0, max_fragment_size - min_fragment_size do
        local overhang_position = max_fragment_size - overhang_offset
        if overhang_position <= #sequence - 4 then
            local overhang_to_test = sequence:sub(overhang_position-3, overhang_position)
            
            -- Make sure overhang isn't already in set
            local already_exists = false
            for _, exclude_overhang in ipairs(exclude_overhangs) do
                if exclude_overhang == overhang_to_test or transform.reverse_complement(exclude_overhang) == overhang_to_test then
                    already_exists = true
                    break
                end
            end
            
            -- Make sure overhang is in set of includeOverhangs. If includeOverhangs is
            -- blank, skip this check.
            local build_available = #include_overhangs == 0
            for _, include_overhang in ipairs(include_overhangs) do
                if include_overhang == overhang_to_test or transform.reverse_complement(include_overhang) == overhang_to_test then
                    build_available = true
                    break
                end
            end
            
            if not already_exists and build_available and not is_palindromic(overhang_to_test) then
                local temp_exclude = {}
                for _, o in ipairs(exclude_overhangs) do table.insert(temp_exclude, o) end
                table.insert(temp_exclude, overhang_to_test)
                local set_efficiency = fragment.set_efficiency(temp_exclude)
                
                if set_efficiency > best_overhang_efficiency then
                    best_overhang_efficiency = set_efficiency
                    best_overhang_position = overhang_position
                end
            end
        end
    end

    if best_overhang_position == 0 then
        return {}, 0, "best_overhang_position failed by equaling zero"
    end

    table.insert(existing_fragments, sequence:sub(1, best_overhang_position))
    table.insert(exclude_overhangs, sequence:sub(best_overhang_position-3, best_overhang_position))
    
    return optimize_overhang_iteration(
        sequence:sub(best_overhang_position-3),
        min_fragment_size,
        recurse_max_fragment_size,
        existing_fragments,
        exclude_overhangs,
        include_overhangs
    )
end

-- Fragment fragments a sequence into fragments between the min and max size,
-- choosing fragment ends for optimal assembly efficiency. Since fragments will
-- be inserted into either a vector or primer binding sites, the first 4 and
-- last 4 base pairs are the initial overhang set.
function fragment.fragment(sequence: string, min_fragment_size: integer, max_fragment_size: integer, exclude_overhangs: {string}): ({string}, number, string)
    sequence = string.upper(sequence)
    local initial_overhangs = {sequence:sub(1,4), sequence:sub(#sequence-3)}
    for _, o in ipairs(exclude_overhangs) do
        table.insert(initial_overhangs, o)
    end
    return optimize_overhang_iteration(sequence, min_fragment_size, max_fragment_size, {}, initial_overhangs, {})
end

-- FragmentWithOverhangs fragments a sequence with only a certain overhang set.
-- This is useful if you are constraining the set of possible overhangs when
-- doing more advanced forms of cloning.
function fragment.fragment_with_overhangs(sequence: string, min_fragment_size: integer, max_fragment_size: integer, exclude_overhangs: {string}, include_overhangs: {string}): ({string}, number, string)
    sequence = string.upper(sequence)
    local initial_overhangs = {sequence:sub(1,4), sequence:sub(#sequence-3)}
    for _, o in ipairs(exclude_overhangs) do
        table.insert(initial_overhangs, o)
    end
    return optimize_overhang_iteration(sequence, min_fragment_size, max_fragment_size, {}, initial_overhangs, include_overhangs)
end

--[[
                            Higher level assembly

Practically speaking, if we are synthesizing DNA, there is an upper limit to
the quantity of DNA we can assemble in a single reaction because of the
mutation rate of the synthesis reaction and because assemble efficiency drops
as we have more fragments.

So, the functions here help with that. They allow DNA to be broken up into
sub-fragments, then re-assembled from there.
]]

-- Helper function for recursive fragment iteration
local function recursive_fragment_iteration(sequence: string, max_coding_size_oligo: integer, assembly_pattern: {integer}, exclude_overhangs: {string}, include_overhangs: {string}, forward_flank: string, reverse_flank: string, iteration: integer): (fragment.Assembly, string)
    --[[
    Ok, so this is a note for you hackers out there: this algorithm can be
    greatly improved. The optimal way to do this would be to do a continuous
    fragmentation, so that each of the smallest possible assemblies are as
    large as possible. This reduces the number of total assembly reactions
    you'd have to do. However, this requires some special programming that is,
    frankly, difficult (some dynamic programming), so I just implemented the
    simplest thing that would work. I'd love to get contributions to improve
    this function, though.

    What we want to optimize for: as FEW assemblies as possible, with the
    smallest assembly always using the maximal number of bases per oligo.

    - Keoni
    ]]

    -- Constants for fragment size calculations
    local SMALLEST_MIN_FRAGMENT_SIZE_SUBTRACTION = 60
    local MIN_FRAGMENT_SIZE_SUBTRACTION = 100

    -- Remove any newlines from sequence
    sequence = string.gsub(sequence, "\n", "")
    local sequence_len = #sequence

    -- Calculate size pattern
    local append_length = #forward_flank + #reverse_flank
    local sizes: {integer} = {}
    local max_size = (max_coding_size_oligo - append_length) * assembly_pattern[1]
    
    for i = 1, #assembly_pattern do
        if i == 1 then
            sizes[i] = max_size
        else
            sizes[i] = sizes[i-1] * assembly_pattern[i] - SMALLEST_MIN_FRAGMENT_SIZE_SUBTRACTION
        end
    end

    local target_sequence = sequence
    if iteration ~= 0 then
        target_sequence = forward_flank .. sequence .. reverse_flank
    end

    -- Handle base case
    if sequence_len <= sizes[1] then
        local fragments, efficiency, err = fragment.fragment_with_overhangs(
            target_sequence,
            max_coding_size_oligo - 60,
            max_coding_size_oligo,
            exclude_overhangs,
            include_overhangs
        )
        if err ~= nil then
            return fragment.Assembly, err
        end
        return {
            sequence = sequence,
            fragments = fragments,
            efficiency = efficiency,
            sub_assemblies = {}
        }, nil
    end

    -- Handle recursive cases
    for i = 1, #sizes-1 do
        if sequence_len <= sizes[i+1] then
            local fragments, efficiency, err = fragment.fragment_with_overhangs(
                target_sequence,
                sizes[i] - MIN_FRAGMENT_SIZE_SUBTRACTION,
                sizes[i],
                exclude_overhangs,
                include_overhangs
            )
            if err ~= nil then
                return fragment.Assembly, err
            end

            local sub_assemblies: {fragment.Assembly} = {}
            for _, frag in ipairs(fragments) do
                local sub_assembly, sub_err = recursive_fragment_iteration(
                    frag,
                    max_coding_size_oligo,
                    assembly_pattern,
                    exclude_overhangs,
                    include_overhangs,
                    forward_flank,
                    reverse_flank,
                    iteration + 1
                )
                if sub_err ~= "" then
                    return fragment.Assembly, sub_err
                end
                table.insert(sub_assemblies, sub_assembly)
            end

            return {
                sequence = sequence,
                fragments = {},
                efficiency = efficiency,
                sub_assemblies = sub_assemblies
            }, nil
        end
    end

    return fragment.Assembly, "Fragment too long!"
end

-- RecursiveFragment fragments a sequence recursively into an assembly, which
-- can be created from sub-assemblies. This function is for designing large
-- pieces of DNA.
--
-- maxCodingSizeOligo should be for the max oligo size. If you are synthesizing
-- from an oligo pool, I'd recommend about 56 shorter so you can add primers,
-- BsaI, and a CA for recursive assembly.
--
-- The assemblyPattern should be for how the oligo should be fragmented up.
-- For example: []int{5,4,4,5} is a very reasonable standard if you have
-- oligos with a 1/2000 mutation rate that are approximately 174bp - you
-- would assemble ~870bp fragments, which should have a 64.72% success rate,
-- or a ~95% success rate over 3 colonies. Assembly pattern is also just a
-- rough... recommendation. Often times the lowest level of oligo has +1 in
-- order to fit the right overhangs in. This doesn't matter that much because
-- the limiting factor in assemblies is typically mutation rate at that size.
--
-- The forwardFlank and reverseFlank are for preparing the sequences for
-- recursive assembly. Generally, this involves appending a certain sequence
-- to each oligo, and also to the edges of each subassembly. Do not add these
-- to the maxCodingSizeOligo: that is done within the function. The flanks are
-- NOT added to the first iteration of the sequence: if they are desired there,
-- add them manually.
function fragment.recursive_fragment(sequence: string, max_coding_size_oligo: integer, assembly_pattern: {integer}, exclude_overhangs: {string}, include_overhangs: {string}, forward_flank: string, reverse_flank: string): (fragment.Assembly, string)
    return recursive_fragment_iteration(sequence, max_coding_size_oligo, assembly_pattern, exclude_overhangs, include_overhangs, forward_flank, reverse_flank, 0)
end

return fragment
