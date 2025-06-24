--[[ Module fix
Module fix fixes synthetic DNA molecules in preparation for synthesis.

Many synthesis companies have restrictions on the DNA they can synthesize. This
synthesis fixer takes advantage of synonymous codons in protein coding
sequences (CDS) to remove problematic sequences that either users don't want
(like restriction enzymes sites) or that would cause DNA synthesis companies to
reject a synthesis project.

This synthesis fixer is meant to cover the majority of use cases for DNA
fixing. It is not intended to cover all possible use cases, since the majority
of DNA design does not actually have these edge cases.

For most users, using `cds_simple` will be sufficient to prepare a sequence
for synthesis (you may want to add in restriction enzyme sites to remove).

Cds does not guarantee that all requested features will be removed. If you
have use case that Cds cannot properly fix, please put an issue in the poly
github.

NOTE: if you aim to fix TTATT at an isoleucine isoleucine alanine junction
as well as CG (ATT ATT GCT) - you will fail. we need to handle this case.
]]

-- Core records
local record DnaSuggestion
    start: integer        -- Position where fix should start
    stop: integer         -- Position where fix should end 
    bias: string         -- "NA", "GC", or "AT"
    quantity_fixes: integer
    suggestion_type: string
end

local record Change
    position: integer
    step: integer
    from: string  
    to: string
    reason: string
end

local record fix
    -- Core functions that generate sequence checkers
    remove_sequence: function(sequences_to_remove: {string}, reason: string): function(string): {DnaSuggestion}
    remove_repeat: function(repeat_len: integer): function(string): {DnaSuggestion}  
    gc_content_fixer: function(upper_bound: number, lower_bound: number): function(string): {DnaSuggestion}
    
    -- Main CDS fixing functions
    cds: function(sequence: string, codon_table: codon.TranslationTable, problematic_sequence_funcs: {function(string): {DnaSuggestion}}): (string, {Change}, string)
    cds_simple: function(sequence: string, codon_table: codon.TranslationTable, sequences_to_remove: {string}): (string, {Change}, string)
end

local transform = require("transform")
local codon = require("codon")

local function gc_content(sequence: string): number
    sequence = string.upper(sequence)
    local guanine_count = 0
    local cytosine_count = 0
    
    -- Count G's and C's
    for i = 1, #sequence do
        local base = string.sub(sequence, i, i)
        if base == "G" then
            guanine_count = guanine_count + 1
        elseif base == "C" then
            cytosine_count = cytosine_count + 1
        end
    end

    -- Calculate percentage
    local guanine_and_cytosine_percentage = (guanine_count + cytosine_count) / #sequence
    return guanine_and_cytosine_percentage
end

-- RemoveSequence is a generator for a problematicSequenceFuncs for specific sequences.
function fix.remove_sequence(sequences_to_remove: {string}, reason: string): function(string): {DnaSuggestion}
    return function(sequence: string): {DnaSuggestion}
        local suggestions: {DnaSuggestion} = {}
        local sequences_to_remove_for_reverse: {string} = {}
        
        for _, seq in ipairs(sequences_to_remove) do
            local reverse_complement_to_remove = transform.reverse_complement(seq)
            -- Palindromes only have to be fixed once, so add a check here for palindromes
            if reverse_complement_to_remove == seq then
                sequences_to_remove_for_reverse = {seq}
            else
                sequences_to_remove_for_reverse = {seq, reverse_complement_to_remove}
            end
            
            for _, site in ipairs(sequences_to_remove_for_reverse) do
                -- Find all occurrences of the site (similar to regexp in Go)
                local start_pos = 1
                while true do
                    local start_idx, end_idx = string.find(sequence, site, start_pos, true)
                    if not start_idx then break end
                    
                    local codon_length = 3
					local position = math.floor((start_idx - 1) / codon_length) + 1
                    table.insert(suggestions, {
                        start = position,
                        stop = math.floor((end_idx - 1) / codon_length) + 1,
                        bias = "NA",
                        quantity_fixes = 1,
                        suggestion_type = reason
                    })
                    start_pos = start_idx + 1
                end
            end
        end
        return suggestions
    end
end

-- RemoveRepeat is a generator to make a problematicSequenceFunc for repeats.
function fix.remove_repeat(repeat_len: integer): function(string): {DnaSuggestion}
    return function(sequence: string): {DnaSuggestion}
        local suggestions: {DnaSuggestion} = {}
        local kmers: {string:boolean} = {}
        
        -- Get a kmer list
        for sequence_position = 1, #sequence - repeat_len do
            local kmer = string.sub(sequence, sequence_position, sequence_position + repeat_len - 1)
            local already_found_forward = kmers[kmer]
            local already_found_reverse = kmers[transform.reverse_complement(kmer)]
            kmers[kmer] = true
            
            if already_found_forward or already_found_reverse then
                local codon_length = 3
				local position = math.floor((sequence_position - 1) / codon_length) + 1
                local leftover = sequence_position % codon_length
				local end_position = math.floor((sequence_position + repeat_len - 1) / codon_length) + 1
                
                if leftover == 0 then
                    table.insert(suggestions, {
                        start = position,
                        stop = end_position,
                        bias = "NA",
                        quantity_fixes = 1,
                        suggestion_type = "Repeat sequence"
                    })
                else
                    table.insert(suggestions, {
                        start = position,
                        stop = end_position - 1,
                        bias = "NA",
                        quantity_fixes = 1,
                        suggestion_type = "Repeat sequence"
                    })
                end
                sequence_position = sequence_position + leftover
            end
        end
        return suggestions
    end
end

-- GcContentFixer is a generator to increase or decrease the overall GcContent of a CDS.
-- GcContent is defined as the percentage of guanine and cytosine base pairs in comparison
-- to adenine and thymine base pairs. Usually, you want the range to be somewhere around 50%,
-- with a decent upperBound being 80% GC and a decent lowerBound being 20%.
function fix.gc_content_fixer(upper_bound: number, lower_bound: number): function(string): {DnaSuggestion}
    return function(sequence: string): {DnaSuggestion}
        local suggestions: {DnaSuggestion} = {}
        local gc_content_percentage = gc_content(sequence)
        local codon_length = 3
        
        if gc_content_percentage > upper_bound then
            local number_of_changes = math.floor((gc_content_percentage - upper_bound) * #sequence) + 1
            table.insert(suggestions, {
                start = 1,
                stop = math.floor((#sequence - 1) / codon_length) + 1,
                bias = "AT",
                quantity_fixes = number_of_changes,
                suggestion_type = "GcContent too high"
            })
        end
        
        if gc_content_percentage < lower_bound then
            local number_of_changes = math.floor((lower_bound - gc_content_percentage) * #sequence) + 1
            table.insert(suggestions, {
                start = 1,
                stop = math.floor((#sequence - 1) / codon_length) + 1,
                bias = "GC",
                quantity_fixes = number_of_changes,
                suggestion_type = "GcContent too low"
            })
        end
        
        return suggestions
    end
end

-- findProblems is a helper function in cds that runs each sequence check 
-- and returns a list of all the suggested changes.
local function find_problems(sequence: string, problematic_sequence_funcs: {function(string): {DnaSuggestion}}): {DnaSuggestion}
    local all_suggestions: {DnaSuggestion} = {}
    
    -- Run each function and collect suggestions
    for _, func in ipairs(problematic_sequence_funcs) do
        local suggestions = func(sequence)
        for _, suggestion in ipairs(suggestions) do
            table.insert(all_suggestions, suggestion)
        end
    end
    
    return all_suggestions
end

--[[
# For developers

cds is the core function of the synthesis fixing module. It takes a CDS
and uses degenerate codons to fix up places with undesirable sequence.

The most important part of synthesis fixing is the functions that go into it
(problematicSequenceFuncs). These provide the range to be fixed as well as the
number of fixes that should be required in that range to fix whatever is
checked for within the function.

cds first builds the following maps:

1. Builds a map of codons at each position, with the last codon in the list
   being taken to rebuild the sequence.
2. Builds a map of potential codons that an input codon can be changed to. For
   example, CAC -> CAT or ATT -> ATA,ATC . There are also codon maps for GC or
   AT nucleotide bias.
3. Builds a map of codon weights. Perhaps the organism really likes CAT codons
   but doesn't code ATA or ATC very often. Given the sequence CACATT, the "CAT"
   change will have a greater relative weight than the ATA or ATC change.

From this map, cds does the following operations:

1. Runs problematicSequenceFuncs on the sequence to get change suggestions.
2. For each output suggestion, get a list of all potential changes to the
   sequence that fix it.
3. Sort each potential change by its codon weight.
4. Append the best change to the positionMap
5. GOTO 1
6. complete
]]
function fix.cds(sequence: string, codon_table: codon.TranslationTable, problematic_sequence_funcs: {function(string): {DnaSuggestion}}): (string, {Change}, string)
    local codon_length = 3
    if #sequence % codon_length ~= 0 then
        return "", {}, "this sequence isn't a complete CDS, please try to use a CDS without interrupted codons"
    end

    -- Setup maps
    local historical_map: {integer:{string}} = {}  -- position -> list of codons
    local weight_map: {string:number} = {}         -- codon -> weight
    local na_bias_map: {string:{string}} = {}      -- codon -> list of possible codons (no bias)
    local gc_bias_map: {string:{string}} = {}      -- codon -> list of possible codons (GC bias)
    local at_bias_map: {string:{string}} = {}      -- codon -> list of possible codons (AT bias)

    -- Build historical maps and full amino acid weights
    local amino_acid_weight_table: {string:integer} = {}
    for _, amino_acid in ipairs(codon_table.amino_acids) do
        local amino_acid_total = 0
        for _, cdn in ipairs(amino_acid.codons) do
            -- Get total weights of all codons for a given amino acid
            amino_acid_total = amino_acid_total + cdn.weight

            -- Build bias maps for potential codon changes
            local codon_gc_count = 0
            for i = 1, #cdn.triplet do
                local base = string.sub(cdn.triplet, i, i)
                if base == "G" or base == "C" then
                    codon_gc_count = codon_gc_count + 1
                end
            end

            for _, to_cdn in ipairs(amino_acid.codons) do
                if cdn.triplet ~= to_cdn.triplet then
                    local to_codon_gc_count = 0
                    for i = 1, #to_cdn.triplet do
                        local base = string.sub(to_cdn.triplet, i, i)
                        if base == "G" or base == "C" then
                            to_codon_gc_count = to_codon_gc_count + 1
                        end
                    end

                    if codon_gc_count > to_codon_gc_count then
                        if not at_bias_map[cdn.triplet] then at_bias_map[cdn.triplet] = {} end
                        table.insert(at_bias_map[cdn.triplet], to_cdn.triplet)
                    elseif codon_gc_count < to_codon_gc_count then
                        if not gc_bias_map[cdn.triplet] then gc_bias_map[cdn.triplet] = {} end
                        table.insert(gc_bias_map[cdn.triplet], to_cdn.triplet)
                    end
                    if not na_bias_map[cdn.triplet] then na_bias_map[cdn.triplet] = {} end
                    table.insert(na_bias_map[cdn.triplet], to_cdn.triplet)
                end
            end
        end
        
        -- Check for incomplete codon table
        if amino_acid_total == 0 then
            return "", {}, "incomplete codon table"
        end
        amino_acid_weight_table[amino_acid.letter] = amino_acid_total
    end

    -- Build weight map
	for _, amino_acid in ipairs(codon_table.amino_acids) do
        for _, cdn in ipairs(amino_acid.codons) do
            local codon_weight_ratio = cdn.weight / amino_acid_weight_table[amino_acid.letter]
            local normalized_codon_weight = 100 * codon_weight_ratio
            weight_map[cdn.triplet] = normalized_codon_weight
        end
    end

	-- At the start of the cds function, after map declarations
	local sequence_length = 0
	
	-- Building historical map
	for pos = 1, #sequence, codon_length do
	    local cdn = string.sub(sequence, pos, pos + codon_length - 1)
		local position = math.floor((pos - 1) / codon_length) + 1
	    if not historical_map[position] then historical_map[position] = {} end
	    table.insert(historical_map[position], cdn)
	    sequence_length = position
	end
	
	-- Modify get_sequence to use sequence_length
	local function get_sequence(history: {integer:{string}}): string
	    local result = {}
	    for pos = 1, sequence_length do
	        local codon_history = history[pos]
	        table.insert(result, codon_history[#codon_history])
	    end
	    return table.concat(result)
	end

    local changes: {Change} = {}
    local fix_iteration = 1
    
    while true do
        local suggestions = find_problems(sequence, problematic_sequence_funcs)
        
        -- If there are no suggestions, break the iteration!
        if #suggestions == 0 then
            -- Sort changes by fixIteration and position
            table.sort(changes, function(a: Change, b: Change): boolean
                if a.step == b.step then
                    return a.position < b.position
                end
                return a.step < b.step
            end)
            return sequence, changes, nil
        end

        for _, suggestion in ipairs(suggestions) do
            -- Check for a valid bias direction
            if suggestion.bias ~= "NA" and suggestion.bias ~= "GC" and suggestion.bias ~= "AT" then
                return sequence, {}, string.format("Invalid bias. Expected NA, GC, or AT, got %s", suggestion.bias)
            end

            -- For each suggestion, get a list of potential changes that could fix the problem
            local potential_changes: {Change} = {}
            
			for position_selector = suggestion.start, suggestion.stop do
				if position_selector > sequence_length then break end
    			
    			local codon_list = historical_map[position_selector]
    			local last_codon = codon_list[#codon_list]
                local unavailable_codons: {string:boolean} = {}
                
                for _, codon_site in ipairs(historical_map[position_selector]) do
                    unavailable_codons[codon_site] = true
                end

                -- Select bias map based on suggestion
                local bias_map: {string:{string}}
                if suggestion.bias == "NA" then
                    bias_map = na_bias_map
                elseif suggestion.bias == "GC" then
                    bias_map = gc_bias_map
                else -- AT
                    bias_map = at_bias_map
                end

                -- Get potential codons from bias map
                if bias_map[last_codon] then
                    for _, potential_codon in ipairs(bias_map[last_codon]) do
                        if not unavailable_codons[potential_codon] then
                            table.insert(potential_changes, {
                                position = position_selector,
                                step = fix_iteration,
                                from = last_codon,
                                to = potential_codon,
                                reason = suggestion.suggestion_type
                            })
                        end
                    end
                end
            end

            -- Sort potential changes by weight
            table.sort(potential_changes, function(a: Change, b: Change): boolean
                return weight_map[a.to] > weight_map[b.to]
            end)

            -- Remove sorted changes that target the same position
            local sorted_changes: {Change} = {}
            local used_positions: {integer:boolean} = {}
            
            for _, potential_change in ipairs(potential_changes) do
                if not used_positions[potential_change.position] then
                    used_positions[potential_change.position] = true
                    table.insert(sorted_changes, potential_change)
                end
            end
			
			if sorted_changes[1] ~= nil then
			    -- Make sure we have enough sorted changes after sorting/removal
        	    if #sorted_changes < suggestion.quantity_fixes then
        	        return sequence, {}, string.format(
        	            "Too many fixes required. Number of potential fixes: %d, number of required fixes: %d",
        	            #potential_changes,
        	            suggestion.quantity_fixes
        	        )
        	    end

        	    -- Apply the best changes up to quantity_fixes
        	    for i = 1, suggestion.quantity_fixes do
        	        local target_change = sorted_changes[i]
        	        table.insert(historical_map[target_change.position], target_change.to)
        	        table.insert(changes, target_change)
        	        sequence = get_sequence(historical_map)
        	    end
        	end
		end
        fix_iteration = fix_iteration + 1

        -- Safety check to prevent infinite loops (100 iterations max)
        if fix_iteration > 100 then
            return sequence, changes, "maximum iterations reached"
        end
    end
end

-- CdsSimple is cds with some defaults for normal usage, including
-- removing of homopolymers, removing any repeat larger than 18 base pairs, and
-- fixing if a CDS's gc content is above 80% or below 20%
function fix.cds_simple(sequence: string, codon_table: codon.TranslationTable, sequences_to_remove: {string}): (string, {Change}, string)
    local functions: {function(string): {DnaSuggestion}} = {}
    
    -- Remove homopolymers
    table.insert(functions, fix.remove_sequence({"AAAAAAAA", "GGGGGGGG"}, "Homopolymers"))
    
    -- Remove user defined sequences
    table.insert(functions, fix.remove_sequence(sequences_to_remove, "Removal requested by user"))
    
    -- Remove repeats
    table.insert(functions, fix.remove_repeat(18))
    
    -- Ensure normal GC range
    table.insert(functions, fix.gc_content_fixer(0.80, 0.20))
    
    return fix.cds(sequence, codon_table, functions)
end

return fix
