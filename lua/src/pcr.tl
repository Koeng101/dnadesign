--[[ Module pcr
Module pcr designs and simulates simple PCR reactions.

PCR, or polymerase chain reaction, is a method developed in 1983 to copy DNA
templates using small fragments of synthesized single-stranded DNA, amplifying
those DNA templates to ~x1,000,000,000 their starting concentration. These small
fragments, referred to as "primers" or "oligos", can be designed on a computer
and then synthesized for amplifying a variety of different templates.

This module allows users to simulate a PCR reaction or design new primers to
amplify a given template. This module assumes perfect annealing to template at
a target temperature, so should only be used for PCR reactions where this is a
reasonable assumption.

If you are trying to simulate amplification out of a large pool, such as an
oligo pool, use the `simulate` rather than `simulate_simple` function to detect
if there is concatemerization happening in your multiplex reaction. In most
other cases, use `simulate_simple`.

IMPORTANT! The target_tm in all functions is specifically for Taq polymerase.
]]

local transform = require("transform")
local primers = require("primers")

local record pcr
    -- Core PCR functions
    design_primers_with_overhangs: function(sequence: string, forward_overhang: string, reverse_overhang: string, target_tm: number): (string, string)
    design_primers: function(sequence: string, target_tm: number): (string, string)
    simulate_simple: function(sequences: {string}, target_tm: number, circular: boolean, primer_list: {string}): {string}
    simulate: function(sequences: {string}, target_tm: number, circular: boolean, primer_list: {string}): ({string}, string)
end

-- Minimum primer length based on https://doi.org/10.1089/dna.1994.13.75
local minimal_primer_length <const> = 15

-- Helper function to find all occurrences of a substring in a string
local function find_all(str: string, pattern: string): {integer}
    local positions: {integer} = {}
    local pos = 1
    while true do
        local start, finish = str:find(pattern, pos, true)
        if not start then break end
        table.insert(positions, start)
        pos = finish + 1
    end
    return positions
end

-- Helper function to generate PCR fragments
local function generate_pcr_fragments(
    sequence: string,
    forward_location: integer,
    reverse_location: integer,
    forward_primer_indices: {integer},
    reverse_primer_indices: {integer},
    minimal_primers: {string},
    primer_list: {string}
): {string}
    local pcr_fragments: {string} = {}
    
    for _, forward_primer_index in ipairs(forward_primer_indices) do
        local minimal_primer = minimal_primers[forward_primer_index]
        local full_primer_forward = primer_list[forward_primer_index]
        
        for _, reverse_primer_index in ipairs(reverse_primer_indices) do
            local full_primer_reverse = transform.reverse_complement(primer_list[reverse_primer_index])
            local pcr_fragment = full_primer_forward:sub(1, #full_primer_forward - #minimal_primer) ..
                                sequence:sub(forward_location, reverse_location - 1) ..
                                full_primer_reverse
            table.insert(pcr_fragments, pcr_fragment)
        end
    end
    
    return pcr_fragments
end

--[[
design_primers_with_overhangs designs two primers to amplify a target sequence,
adding on an overhang to the forward and reverse strand. This overhang can
contain additional DNA needed for assembly, like Gibson assembly overhangs
or restriction enzyme sites.
]]
function pcr.design_primers_with_overhangs(sequence: string, forward_overhang: string, reverse_overhang: string, target_tm: number): (string, string)
    sequence = string.upper(sequence)
    local forward_primer = sequence:sub(1, minimal_primer_length)
    
    -- Extend forward primer until target Tm is reached
    local additional_nucleotides = 0
    while primers.melting_temp(forward_primer) < target_tm do
        additional_nucleotides = additional_nucleotides + 1
        forward_primer = sequence:sub(1, minimal_primer_length + additional_nucleotides)
    end
    
    -- Design reverse primer
    local reverse_primer = transform.reverse_complement(sequence:sub(#sequence - minimal_primer_length + 1))
    additional_nucleotides = 0
    while primers.melting_temp(reverse_primer) < target_tm do
        additional_nucleotides = additional_nucleotides + 1
        reverse_primer = transform.reverse_complement(sequence:sub(#sequence - (minimal_primer_length + additional_nucleotides) + 1))
    end
    
    -- Add overhangs
    forward_primer = forward_overhang .. forward_primer
    reverse_primer = transform.reverse_complement(reverse_overhang) .. reverse_primer
    
    return forward_primer, reverse_primer
end

--[[
design_primers designs two primers to amplify a target sequence without any overhangs.
]]
function pcr.design_primers(sequence: string, target_tm: number): (string, string)
    return pcr.design_primers_with_overhangs(sequence, "", "", target_tm)
end

--[[
simulate_simple simulates a PCR reaction. It takes in a list of sequences and
a list of primers, with support for complex multiplex reactions, produces
a list of all possible PCR fragments from such a reaction. It does not
detect concatemerization. The variable 'circular' is for if the target template 
is circular, like a plasmid.
]]
function pcr.simulate_simple(sequences: {string}, target_tm: number, circular: boolean, primer_list: {string}): {string}
    -- Convert primers to uppercase
    local upper_primers: {string} = {}
    for _, primer in ipairs(primer_list) do
        table.insert(upper_primers, string.upper(primer))
    end
    
    local pcr_fragments: {string} = {}
    
    for _, sequence in ipairs(sequences) do
        sequence = string.upper(sequence)
        
        local forward_locations: {integer:{integer}} = {}
        local reverse_locations: {integer:{integer}} = {}
        local minimal_primers: {string} = {}
        
        -- Find binding sites for each primer
        for primer_index, primer in ipairs(upper_primers) do
            local minimal_length = minimal_primer_length
            while primers.melting_temp(primer:sub(-minimal_length)) < target_tm do
                minimal_length = minimal_length + 1
                if minimal_length > #primer then break end
            end
            
            local minimal_primer = primer:sub(-minimal_length)
			if minimal_primer ~= primer then  
				minimal_primers[primer_index] = minimal_primer
			end
            
            -- Find forward binding sites
            for _, loc in ipairs(find_all(sequence, minimal_primer)) do
                if not forward_locations[loc] then forward_locations[loc] = {} end
                table.insert(forward_locations[loc], primer_index)
            end
            
            -- Find reverse binding sites
            local rev_minimal_primer = transform.reverse_complement(minimal_primer)
            for _, loc in ipairs(find_all(sequence, rev_minimal_primer)) do
                if not reverse_locations[loc] then reverse_locations[loc] = {} end
                table.insert(reverse_locations[loc], primer_index)
            end
        end
        
        -- Process locations to generate fragments
        local forward_locs: {integer} = {}
        local reverse_locs: {integer} = {}
        for loc, _ in pairs(forward_locations) do table.insert(forward_locs, loc) end
        for loc, _ in pairs(reverse_locations) do table.insert(reverse_locs, loc) end
        table.sort(forward_locs)
        table.sort(reverse_locs)
        
        -- Generate PCR fragments
        for i, forward_loc in ipairs(forward_locs) do
            if i < #forward_locs then
                for _, reverse_loc in ipairs(reverse_locs) do
                    if forward_loc < reverse_loc and reverse_loc < forward_locs[i + 1] then
                        local new_fragments = generate_pcr_fragments(
                            sequence,
                            forward_loc,
                            reverse_loc,
                            forward_locations[forward_loc],
                            reverse_locations[reverse_loc],
                            minimal_primers,
                            upper_primers
                        )
                        for _, fragment in ipairs(new_fragments) do
                            table.insert(pcr_fragments, fragment)
                        end
                        break
                    end
                end
            else
                local found_fragment = false
                for _, reverse_loc in ipairs(reverse_locs) do
                    if forward_loc < reverse_loc then
                        local new_fragments = generate_pcr_fragments(
                            sequence,
                            forward_loc,
                            reverse_loc,
                            forward_locations[forward_loc],
                            reverse_locations[reverse_loc],
                            minimal_primers,
                            upper_primers
                        )
                        for _, fragment in ipairs(new_fragments) do
                            table.insert(pcr_fragments, fragment)
                        end
                        found_fragment = true
                    end
                end
                
                -- Handle circular templates
                if circular and not found_fragment then
                    for _, reverse_loc in ipairs(reverse_locs) do
                        if forward_locs[1] > reverse_loc then
                            local rotated_sequence = sequence:sub(forward_loc) .. sequence:sub(1, forward_loc - 1)
                            local rotated_forward_loc = 1
                            local rotated_reverse_loc = #sequence - forward_loc + reverse_loc + 1
                            local new_fragments = generate_pcr_fragments(
                                rotated_sequence,
                                rotated_forward_loc,
                                rotated_reverse_loc,
                                forward_locations[forward_loc],
                                reverse_locations[reverse_loc],
                                minimal_primers,
                                upper_primers
                            )
                            for _, fragment in ipairs(new_fragments) do
                                table.insert(pcr_fragments, fragment)
                            end
                        end
                    end
                end
            end
        end
    end
    
    return pcr_fragments
end

--[[
simulate simulates a PCR reaction, including concatemerization analysis. It
takes in a list of sequences and list of primers, produces all possible PCR
fragments in a given reaction, and then attempts to see if the output
fragments can amplify themselves. The variable 'circular' is for if the 
target template is circular, like a plasmid.
]]
function pcr.simulate(sequences: {string}, target_tm: number, circular: boolean, primer_list: {string}): ({string}, string)
    local initial_amplification = pcr.simulate_simple(sequences, target_tm, circular, primer_list)
    
    -- Create combined primer list
    local combined_primers: {string} = {}
    for _, primer in ipairs(primer_list) do
        table.insert(combined_primers, primer)
    end
    for _, fragment in ipairs(initial_amplification) do
        table.insert(combined_primers, fragment)
    end
    
    local subsequent_amplification = pcr.simulate_simple(sequences, target_tm, circular, combined_primers)
    
    if #initial_amplification ~= #subsequent_amplification then
        return initial_amplification, "Concatemerization detected in PCR."
    end
    
    return initial_amplification, ""
end

return pcr
