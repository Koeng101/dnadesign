--[[ Module codon
Module codon is a module for optimizing codons for expression when synthesizing DNA.

This module contains almost everything you need to do standard codon optimization.

Biological context: certain cells favor certain codons and will reject or under
express sequences that don't use a similar ratio of codons.
This is called codon bias: https://en.wikipedia.org/wiki/Codon_usage_bias

Furthermore, different ribosomes in different organisms will interpret codons differently.
What may be a start codon for one ribosome may be a stop in the other.
Heck, apparently nucleomorphs contain 4 different kinds of ribosomes.
https://en.wikipedia.org/wiki/Nucleomorph <- Thanks Keoni for mentioning this example!

TTFN,
Tim
]]

-- Interface for codon operations
local record codon
	-- core table generation
    default_tables: {string:TranslationTable}
	translation_tables_by_number: {integer:TranslationTable} -- blank translation tables
    new_translation_table: function(index: integer): TranslationTable
    copy_translation_table: function(t: TranslationTable): TranslationTable

	-- codon table creation functions
    get_codon_frequency: function(sequence: string): {string:integer}
    get_stochastic_codon: function(aa: AminoAcid): string
    weight_amino_acids: function(sequence: string, amino_acids: {AminoAcid}): {AminoAcid}

	-- compromise tables
	update_weights: function(t: TranslationTable, new_weights: {AminoAcid}): string
    compromise_codon_table: function(first_t: TranslationTable, second_t: TranslationTable, cut_off: number): (TranslationTable, string)
    add_codon_table: function(first_t: TranslationTable, second_t: TranslationTable): (TranslationTable, string)

	-- generating from NCBI strings
    generate_codon_table: function(amino_acids: string, starts: string): TranslationTable

	TranslationTable: TranslationTable
end

-- Error messages
local errEmptyAminoAcidString <const> = "empty amino acid string"
local errEmptySequenceString <const> = "empty sequence string"

-- Core records
local record Codon
    triplet: string
    weight: integer
end

local record AminoAcid
    letter: string
    codons: {Codon}
end

local record TranslationTable
    start_codons: {string}
    stop_codons: {string}
    amino_acids: {AminoAcid}
    translation_map: {string:string}  -- maps codons to amino acids
    start_codon_table: {string:string}  -- maps start codons to their amino acids
	translate: function(self: TranslationTable, dna_seq: string): (string, string)
    optimize: function(self: TranslationTable, amino_acids: string): (string, string)
    standardize_last_codon: function(self: TranslationTable, dna_seq: string): (string, string)
end

-- Copy method for TranslationTable
function codon.copy_translation_table(t: TranslationTable): TranslationTable
    local new_t: TranslationTable = {
        start_codons = {},
        stop_codons = {},
        amino_acids = {},
        translation_map = {},
        start_codon_table = {},
		translate = t.translate,
        optimize = t.optimize,
        standardize_last_codon = t.standardize_last_codon
    }
    
    -- Copy arrays
    for _, cdn in ipairs(t.start_codons) do
        table.insert(new_t.start_codons, cdn)
    end
    for _, cdn in ipairs(t.stop_codons) do
        table.insert(new_t.stop_codons, cdn)
    end
    
    -- Copy amino acids
    for _, aa in ipairs(t.amino_acids) do
        local new_codons: {Codon} = {}
        for _, cdn in ipairs(aa.codons) do
            table.insert(new_codons, {
                triplet = cdn.triplet,
                weight = cdn.weight
            })
        end
        table.insert(new_t.amino_acids, {
            letter = aa.letter,
            codons = new_codons
        })
    end
    
    -- Copy maps
    for k, v in pairs(t.translation_map) do
        new_t.translation_map[k] = v
    end
    for k, v in pairs(t.start_codon_table) do
        new_t.start_codon_table[k] = v
    end
    
    return new_t
end

-- GetStochasticCodon gets a random codon based on weights
function codon.get_stochastic_codon(aa: AminoAcid): string
    local total_weight = 0
    local cumulative_weights: {integer} = {}
    
    -- Calculate cumulative weights
    for i, cdn in ipairs(aa.codons) do
        total_weight = total_weight + cdn.weight
        cumulative_weights[i] = total_weight
    end
    
    if total_weight == 0 then 
        return nil
    end
    
    -- Generate a random number in the range of totalWeight
    local r = math.random(0, total_weight - 1)
    
    -- Find the first element where r is less than or equal to its cumulative weight
    for i, cdn in ipairs(aa.codons) do
        if r < cumulative_weights[i] then
            return cdn.triplet
        end
    end
    
    return nil
end

-- GetCodonFrequency takes a DNA sequence and returns a hashmap of its codons and their frequencies
function codon.get_codon_frequency(sequence: string): {string:integer}
    local cdn_freq_map: {string:integer} = {}
    sequence = string.upper(sequence)
    
    for i = 1, #sequence - 2, 3 do
        local cdn = sequence:sub(i, i + 2)
        cdn_freq_map[cdn] = (cdn_freq_map[cdn] or 0) + 1
    end
    
    return cdn_freq_map
end

-- WeightAminoAcids weights each codon in a codon table according to input string codon frequency
function codon.weight_amino_acids(sequence: string, amino_acids: {AminoAcid}): {AminoAcid}
    sequence = string.upper(sequence)
    local cdn_freq_map = codon.get_codon_frequency(sequence)
    
    local weighted_amino_acids: {AminoAcid} = {}
    
    for _, amino_acid in ipairs(amino_acids) do
        local new_codons: {Codon} = {}
        for _, cdn in ipairs(amino_acid.codons) do
            table.insert(new_codons, {
                triplet = cdn.triplet,
                weight = cdn_freq_map[cdn.triplet] or 0
            })
        end
        table.insert(weighted_amino_acids, {
            letter = amino_acid.letter,
            codons = new_codons
        })
    end
    
    return weighted_amino_acids
end

-- Add translate method
local function translate(self: TranslationTable, dna_seq: string): (string, string)
    if dna_seq == "" then
        return "", errEmptySequenceString
    end
    
    local current_amino_acids: {string} = {}
    dna_seq = string.upper(dna_seq)
    
    for i = 1, #dna_seq - 2, 3 do
        local cdn = dna_seq:sub(i, i + 2)
        local aa = self.translation_map[cdn]
        if aa then
            table.insert(current_amino_acids, aa)
        end
    end
    
    return table.concat(current_amino_acids), nil
end

-- Add optimize method
local function optimize(self: TranslationTable, aas: string): (string, string)
    -- Finding any given aminoAcid is dependent upon it being capitalized
    local amino_acids = string.upper(aas)
    
    if amino_acids == "" then
        return "", errEmptyAminoAcidString
    end
    
    local codons: {string} = {}
    for i = 1, #amino_acids do
        local amino_acid = amino_acids:sub(i, i)
        local found = false
        
        for _, aa in ipairs(self.amino_acids) do
            if amino_acid == aa.letter then
                table.insert(codons, codon.get_stochastic_codon(aa))
                found = true
                break
            end
        end
        
        if not found then
            return "", string.format('amino acid "%s" is missing from codon table', amino_acid)
        end
    end
    
    return table.concat(codons), nil
end

-- Standardized codon mapping
local standard_codons: {string:string} = {
    K = "AAA", L = "TTA", R = "CGA", F = "TTT", Q = "CAA", W = "TGG",
    H = "CAC", P = "CCG", V = "GTG", C = "TGC", T = "ACA", S = "TCC", 
    Y = "TAC", I = "ATA", G = "GGT", N = "AAC", A = "GCC", M = "ATG", 
    D = "GAC", E = "GAA"
}

-- Standardize last codon function
local function standardize_last_codon(self: TranslationTable, dna_seq: string): (string, string)
    if dna_seq == "" then
        return "", errEmptySequenceString
    end
    
    dna_seq = string.upper(dna_seq)
    local seq_len = #dna_seq
    
    if seq_len < 3 then
        return dna_seq, nil
    end
    
    local last_codon = dna_seq:sub(seq_len - 2, seq_len)
    local aa = self.translation_map[last_codon]
    
    if not aa then
        return dna_seq, string.format('unknown codon "%s"', last_codon)
    end
    
    local standard_codon = standard_codons[aa]
    if not standard_codon then
        return dna_seq, string.format('no standard codon for amino acid "%s"', aa)
    end
    
    return dna_seq:sub(1, seq_len - 3) .. standard_codon, nil
end

-- Function to generate default codon tables from NCBI
function codon.generate_codon_table(amino_acids: string, starts: string): TranslationTable
    -- Define base sequences for codon generation
    local base1 = "TTTTTTTTTTTTTTTTCCCCCCCCCCCCCCCCAAAAAAAAAAAAAAAAGGGGGGGGGGGGGGGG"
    local base2 = "TTTTCCCCAAAAGGGGTTTTCCCCAAAAGGGGTTTTCCCCAAAAGGGGTTTTCCCCAAAAGGGG"
    local base3 = "TCAGTCAGTCAGTCAGTCAGTCAGTCAGTCAGTCAGTCAGTCAGTCAGTCAGTCAGTCAGTCAG"
    
    -- Initialize maps and lists
    local amino_acid_map: {string:{Codon}} = {}
    local start_codons: {string} = {}
    local stop_codons: {string} = {}
    local translation_map: {string:string} = {}
    
    -- Process the amino acids and starts strings
    for i = 1, #amino_acids do
        local aa = amino_acids:sub(i, i)
        local start = starts:sub(i, i)
        local triplet = base1:sub(i, i) .. base2:sub(i, i) .. base3:sub(i, i)
        
        -- Initialize amino acid entry if it doesn't exist
        if not amino_acid_map[aa] then
            amino_acid_map[aa] = {}
        end
        
        -- Add codon to appropriate amino acid
        table.insert(amino_acid_map[aa], {triplet = triplet, weight = 1})
        translation_map[triplet] = aa
        
        -- Check if this is a start codon
        if start == "M" then
            table.insert(start_codons, triplet)
        end
        
        -- Check if this is a stop codon
        if aa == "*" then
            table.insert(stop_codons, triplet)
        end
    end
    
    -- Convert amino acid map to array format
    local amino_acids_array: {AminoAcid} = {}
    for aa, codons in pairs(amino_acid_map) do
        table.insert(amino_acids_array, {
            letter = aa,
            codons = codons
        })
    end
    
    -- Generate start codon table
    local start_codon_table: {string:string} = {}
    for _, cdn in ipairs(start_codons) do
        start_codon_table[cdn] = "M"
    end
    
    -- Create and return the translation table
    return {
        start_codons = start_codons,
        stop_codons = stop_codons,
        amino_acids = amino_acids_array,
        translation_map = translation_map,
        start_codon_table = start_codon_table,
		translate = translate,
		optimize = optimize,
        standardize_last_codon = standardize_last_codon
	    }
end

-- CompromiseCodonTable takes 2 CodonTables and makes a new codonTable
-- that is an equal compromise between the two tables
function codon.compromise_codon_table(first_t: TranslationTable, second_t: TranslationTable, cut_off: number): (TranslationTable, string)
    -- Check if cutOff is too high or low (this is converted to a percent)
    if cut_off < 0 then
        return first_t, "cut off too low, cannot be less than 0"
    end
    if cut_off > 1 then
        return first_t, "cut off too high, cannot be greater than 1"
    end
    
    -- Copy first table to base our merge on
    local merged_t = codon.copy_translation_table(first_t)
    local final_amino_acids: {AminoAcid} = {}
    
    -- Loop over all AminoAcids in the first table
    for _, first_aa in ipairs(first_t.amino_acids) do
        local first_triplets: {string} = {}
        local first_weights: {integer} = {}
        local first_total = 0
        
        local second_weights: {integer} = {}
        local second_total = 0
        
        -- Get triplets and weights from first table
        for _, first_cdn in ipairs(first_aa.codons) do
            table.insert(first_triplets, first_cdn.triplet)
            table.insert(first_weights, first_cdn.weight)
            first_total = first_total + first_cdn.weight
            
            -- Find matching codon in second table
            for _, second_aa in ipairs(second_t.amino_acids) do
                if second_aa.letter == first_aa.letter then
                    for _, second_cdn in ipairs(second_aa.codons) do
                        if second_cdn.triplet == first_cdn.triplet then
                            table.insert(second_weights, second_cdn.weight)
                            second_total = second_total + second_cdn.weight
                            break
                        end
                    end
                end
            end
        end
        
        local final_codons: {Codon} = {}
        local cut_off_weight = math.floor(10000 * cut_off)
        
        -- Calculate final weights and create new codons
        for i, triplet in ipairs(first_triplets) do
            local first_weight = math.floor((first_weights[i] / first_total) * 10000)
            local second_weight = math.floor((second_weights[i] / second_total) * 10000)
            
            if first_weight < cut_off_weight or second_weight < cut_off_weight then
                table.insert(final_codons, {triplet = triplet, weight = 0})
            else
                table.insert(final_codons, {
                    triplet = triplet,
                    weight = math.floor((first_weight + second_weight) / 2)
                })
            end
        end
        
        table.insert(final_amino_acids, {
            letter = first_aa.letter,
            codons = final_codons
        })
    end
    
    merged_t.amino_acids = final_amino_acids
    return merged_t, nil
end

-- AddCodonTable takes 2 CodonTables and adds them together
function codon.add_codon_table(first_t: TranslationTable, second_t: TranslationTable): (TranslationTable, string)
    local final_amino_acids: {AminoAcid} = {}
    
    for _, first_aa in ipairs(first_t.amino_acids) do
        local final_codons: {Codon} = {}
        
        for _, first_cdn in ipairs(first_aa.codons) do
            for _, second_aa in ipairs(second_t.amino_acids) do
                for _, second_cdn in ipairs(second_aa.codons) do
                    if first_cdn.triplet == second_cdn.triplet then
                        table.insert(final_codons, {
                            triplet = first_cdn.triplet,
                            weight = first_cdn.weight + second_cdn.weight
                        })
                    end
                end
            end
        end
        
        table.insert(final_amino_acids, {
            letter = first_aa.letter,
            codons = final_codons
        })
    end
    
    local merged_t = codon.copy_translation_table(first_t)
    merged_t.amino_acids = final_amino_acids
    return merged_t, nil
end

-- Initialize default translation tables
local translation_tables_by_number = {
    [1] = codon.generate_codon_table("FFLLSSSSYY**CC*WLLLLPPPPHHQQRRRRIIIMTTTTNNKKSSRRVVVVAAAADDEEGGGG", "---M------**--*----M---------------M----------------------------"),
    [2] = codon.generate_codon_table("FFLLSSSSYY**CCWWLLLLPPPPHHQQRRRRIIMMTTTTNNKKSS**VVVVAAAADDEEGGGG", "----------**--------------------MMMM----------**---M------------"),
    [3] = codon.generate_codon_table("FFLLSSSSYY**CCWWTTTTPPPPHHQQRRRRIIMMTTTTNNKKSSRRVVVVAAAADDEEGGGG", "----------**----------------------MM---------------M------------"),
    [4] = codon.generate_codon_table("FFLLSSSSYY**CCWWLLLLPPPPHHQQRRRRIIIMTTTTNNKKSSRRVVVVAAAADDEEGGGG", "--MM------**-------M------------MMMM---------------M------------"),
    [5] = codon.generate_codon_table("FFLLSSSSYY**CCWWLLLLPPPPHHQQRRRRIIMMTTTTNNKKSSSSVVVVAAAADDEEGGGG", "---M------**--------------------MMMM---------------M------------"),
    [6] = codon.generate_codon_table("FFLLSSSSYYQQCC*WLLLLPPPPHHQQRRRRIIIMTTTTNNKKSSRRVVVVAAAADDEEGGGG", "--------------*--------------------M----------------------------"),
    [9] = codon.generate_codon_table("FFLLSSSSYY**CCWWLLLLPPPPHHQQRRRRIIIMTTTTNNNKSSSSVVVVAAAADDEEGGGG", "----------**-----------------------M---------------M------------"),
    [10] = codon.generate_codon_table("FFLLSSSSYY**CCCWLLLLPPPPHHQQRRRRIIIMTTTTNNKKSSRRVVVVAAAADDEEGGGG", "----------**-----------------------M----------------------------"),
    [11] = codon.generate_codon_table("FFLLSSSSYY**CC*WLLLLPPPPHHQQRRRRIIIMTTTTNNKKSSRRVVVVAAAADDEEGGGG", "---M------**--*----M------------MMMM---------------M------------"),
    [12] = codon.generate_codon_table("FFLLSSSSYY**CC*WLLLSPPPPHHQQRRRRIIIMTTTTNNKKSSRRVVVVAAAADDEEGGGG", "----------**--*----M---------------M----------------------------"),
    [13] = codon.generate_codon_table("FFLLSSSSYY**CCWWLLLLPPPPHHQQRRRRIIMMTTTTNNKKSSGGVVVVAAAADDEEGGGG", "---M------**----------------------MM---------------M------------"),
    [14] = codon.generate_codon_table("FFLLSSSSYYY*CCWWLLLLPPPPHHQQRRRRIIIMTTTTNNNKSSSSVVVVAAAADDEEGGGG", "-----------*-----------------------M----------------------------"),
    [16] = codon.generate_codon_table("FFLLSSSSYY*LCC*WLLLLPPPPHHQQRRRRIIIMTTTTNNKKSSRRVVVVAAAADDEEGGGG", "----------*---*--------------------M----------------------------"),
    [21] = codon.generate_codon_table("FFLLSSSSYY**CCWWLLLLPPPPHHQQRRRRIIMMTTTTNNNKSSSSVVVVAAAADDEEGGGG", "----------**-----------------------M---------------M------------"),
    [22] = codon.generate_codon_table("FFLLSS*SYY*LCC*WLLLLPPPPHHQQRRRRIIIMTTTTNNKKSSRRVVVVAAAADDEEGGGG", "------*---*---*--------------------M----------------------------"),
    [23] = codon.generate_codon_table("FF*LSSSSYY**CC*WLLLLPPPPHHQQRRRRIIIMTTTTNNKKSSRRVVVVAAAADDEEGGGG", "--*-------**--*-----------------M--M---------------M------------"),
    [24] = codon.generate_codon_table("FFLLSSSSYY**CCWWLLLLPPPPHHQQRRRRIIIMTTTTNNKKSSSKVVVVAAAADDEEGGGG", "---M------**-------M---------------M---------------M------------"),
    [25] = codon.generate_codon_table("FFLLSSSSYY**CCGWLLLLPPPPHHQQRRRRIIIMTTTTNNKKSSRRVVVVAAAADDEEGGGG", "---M------**-----------------------M---------------M------------"),
    [26] = codon.generate_codon_table("FFLLSSSSYY**CC*WLLLAPPPPHHQQRRRRIIIMTTTTNNKKSSRRVVVVAAAADDEEGGGG", "----------**--*----M---------------M----------------------------"),
    [27] = codon.generate_codon_table("FFLLSSSSYYQQCCWWLLLLPPPPHHQQRRRRIIIMTTTTNNKKSSRRVVVVAAAADDEEGGGG", "--------------*--------------------M----------------------------"),
    [28] = codon.generate_codon_table("FFLLSSSSYYQQCCWWLLLLPPPPHHQQRRRRIIIMTTTTNNKKSSRRVVVVAAAADDEEGGGG", "----------**--*--------------------M----------------------------"),
    [29] = codon.generate_codon_table("FFLLSSSSYYYYCC*WLLLLPPPPHHQQRRRRIIIMTTTTNNKKSSRRVVVVAAAADDEEGGGG", "--------------*--------------------M----------------------------"),
    [30] = codon.generate_codon_table("FFLLSSSSYYEECC*WLLLLPPPPHHQQRRRRIIIMTTTTNNKKSSRRVVVVAAAADDEEGGGG", "--------------*--------------------M----------------------------"),
    [31] = codon.generate_codon_table("FFLLSSSSYYEECCWWLLLLPPPPHHQQRRRRIIIMTTTTNNKKSSRRVVVVAAAADDEEGGGG", "----------**-----------------------M----------------------------"),
    [33] = codon.generate_codon_table("FFLLSSSSYYY*CCWWLLLLPPPPHHQQRRRRIIIMTTTTNNKKSSSKVVVVAAAADDEEGGGG", "---M-------*-------M---------------M---------------M------------")
}
codon.translation_tables_by_number = translation_tables_by_number

-- Set up new_translation_table function
function codon.new_translation_table(index: integer): TranslationTable
    return codon.copy_translation_table(translation_tables_by_number[index])
end

-- Initialize default tables with data from freqB.json, pichia.json, and scerevisiae.json
codon.default_tables = {
    ["ecoli"] = {
        start_codons = {"TTG", "CTG", "ATT", "ATC", "ATA", "ATG", "GTG"},
        stop_codons = {"TAA", "TAG", "TGA"},
        amino_acids = {
            { letter = "Y", codons = {{ triplet = "TAT", weight = 42 }, { triplet = "TAC", weight = 58 }}},
            { letter = "C", codons = {{ triplet = "TGT", weight = 42 }, { triplet = "TGC", weight = 58 }}},
            { letter = "I", codons = {{ triplet = "ATT", weight = 49 }, { triplet = "ATC", weight = 51 }, { triplet = "ATA", weight = 0 }}},
            { letter = "V", codons = {{ triplet = "GTT", weight = 35 }, { triplet = "GTC", weight = 28 }, { triplet = "GTA", weight = 3 }, { triplet = "GTG", weight = 34 }}},
            { letter = "G", codons = {{ triplet = "GGT", weight = 60 }, { triplet = "GGC", weight = 39 }, { triplet = "GGA", weight = 0 }, { triplet = "GGG", weight = 0 }}},
            { letter = "L", codons = {{ triplet = "TTA", weight = 3 }, { triplet = "TTG", weight = 14 }, { triplet = "CTT", weight = 2 }, { triplet = "CTC", weight = 3 }, { triplet = "CTA", weight = 0 }, { triplet = "CTG", weight = 78 }}},
            { letter = "W", codons = {{ triplet = "TGG", weight = 1 }}},
            { letter = "K", codons = {{ triplet = "AAA", weight = 51 }, { triplet = "AAG", weight = 49 }}},
            { letter = "S", codons = {{ triplet = "TCT", weight = 10 }, { triplet = "TCC", weight = 13 }, { triplet = "TCA", weight = 2 }, { triplet = "TCG", weight = 5 }, { triplet = "AGT", weight = 0 }, { triplet = "AGC", weight = 68 }}},
            { letter = "*", codons = {{ triplet = "TAA", weight = 2015 }, { triplet = "TAG", weight = 1667 }, { triplet = "TGA", weight = 1300 }}},
            { letter = "Q", codons = {{ triplet = "CAA", weight = 45 }, { triplet = "CAG", weight = 55 }}},
            { letter = "R", codons = {{ triplet = "CGT", weight = 62 }, { triplet = "CGC", weight = 35 }, { triplet = "CGA", weight = 0 }, { triplet = "CGG", weight = 0 }, { triplet = "AGA", weight = 3 }, { triplet = "AGG", weight = 0 }}},
            { letter = "A", codons = {{ triplet = "GCT", weight = 12 }, { triplet = "GCC", weight = 19 }, { triplet = "GCA", weight = 24 }, { triplet = "GCG", weight = 44 }}},
            { letter = "E", codons = {{ triplet = "GAA", weight = 43 }, { triplet = "GAG", weight = 57 }}},
            { letter = "F", codons = {{ triplet = "TTT", weight = 45 }, { triplet = "TTC", weight = 55 }}},
            { letter = "P", codons = {{ triplet = "CCT", weight = 9 }, { triplet = "CCC", weight = 0 }, { triplet = "CCA", weight = 10 }, { triplet = "CCG", weight = 81 }}},
            { letter = "H", codons = {{ triplet = "CAT", weight = 38 }, { triplet = "CAC", weight = 62 }}},
            { letter = "M", codons = {{ triplet = "ATG", weight = 1 }}},
            { letter = "T", codons = {{ triplet = "ACT", weight = 10 }, { triplet = "ACC", weight = 57 }, { triplet = "ACA", weight = 0 }, { triplet = "ACG", weight = 33 }}},
            { letter = "N", codons = {{ triplet = "AAT", weight = 47 }, { triplet = "AAC", weight = 53 }}},
            { letter = "D", codons = {{ triplet = "GAT", weight = 46 }, { triplet = "GAC", weight = 54 }}}
        },
        translation_map = {},  -- Will be filled by init
        start_codon_table = {} -- Will be filled by init
    },
    ["pichia"] = {
        start_codons = {"TTG", "CTG", "ATT", "ATC", "ATA", "ATG", "GTG"},
        stop_codons = {"TAA", "TAG", "TGA"},
        amino_acids = {
            { letter = "Y", codons = {{ triplet = "TAT", weight = 40017 }, { triplet = "TAC", weight = 37740 }}},
            { letter = "C", codons = {{ triplet = "TGT", weight = 17099 }, { triplet = "TGC", weight = 10242 }}},
            { letter = "I", codons = {{ triplet = "ATT", weight = 68516 }, { triplet = "ATC", weight = 43651 }, { triplet = "ATA", weight = 35059 }}},
            { letter = "V", codons = {{ triplet = "GTT", weight = 54750 }, { triplet = "GTC", weight = 30526 }, { triplet = "GTA", weight = 25054 }, { triplet = "GTG", weight = 30581 }}},
            { letter = "G", codons = {{ triplet = "GGT", weight = 42959 }, { triplet = "GGC", weight = 18853 }, { triplet = "GGA", weight = 43541 }, { triplet = "GGG", weight = 14618 }}},
            { letter = "L", codons = {{ triplet = "TTA", weight = 41481 }, { triplet = "TTG", weight = 68335 }, { triplet = "CTT", weight = 40288 }, { triplet = "CTC", weight = 20003 }, { triplet = "CTA", weight = 29034 }, { triplet = "CTG", weight = 35916 }}},
            { letter = "W", codons = {{ triplet = "TGG", weight = 23941 }}},
            { letter = "K", codons = {{ triplet = "AAA", weight = 83571 }, { triplet = "AAG", weight = 77197 }}},
            { letter = "S", codons = {{ triplet = "TCT", weight = 53665 }, { triplet = "TCC", weight = 35643 }, { triplet = "TCA", weight = 43185 }, { triplet = "TCG", weight = 19746 }, { triplet = "AGT", weight = 32769 }, { triplet = "AGC", weight = 21832 }}},
            { letter = "*", codons = {{ triplet = "TAA", weight = 2015 }, { triplet = "TAG", weight = 1667 }, { triplet = "TGA", weight = 1300 }}},
            { letter = "Q", codons = {{ triplet = "CAA", weight = 58688 }, { triplet = "CAG", weight = 38500 }}},
            { letter = "R", codons = {{ triplet = "CGT", weight = 14716 }, { triplet = "CGC", weight = 5515 }, { triplet = "CGA", weight = 12855 }, { triplet = "CGG", weight = 5643 }, { triplet = "AGA", weight = 47972 }, { triplet = "AGG", weight = 19381 }}},
            { letter = "A", codons = {{ triplet = "GCT", weight = 51452 }, { triplet = "GCC", weight = 30978 }, { triplet = "GCA", weight = 35840 }, { triplet = "GCG", weight = 10148 }}},
            { letter = "E", codons = {{ triplet = "GAA", weight = 93407 }, { triplet = "GAG", weight = 64293 }}},
            { letter = "F", codons = {{ triplet = "TTT", weight = 60424 }, { triplet = "TTC", weight = 43704 }}},
            { letter = "P", codons = {{ triplet = "CCT", weight = 35821 }, { triplet = "CCC", weight = 18924 }, { triplet = "CCA", weight = 39324 }, { triplet = "CCG", weight = 10585 }}},
            { letter = "H", codons = {{ triplet = "CAT", weight = 30739 }, { triplet = "CAC", weight = 19034 }}},
            { letter = "M", codons = {{ triplet = "ATG", weight = 42837 }}},
            { letter = "T", codons = {{ triplet = "ACT", weight = 47886 }, { triplet = "ACC", weight = 31320 }, { triplet = "ACA", weight = 36947 }, { triplet = "ACG", weight = 16313 }}},
            { letter = "N", codons = {{ triplet = "AAT", weight = 66744 }, { triplet = "AAC", weight = 57670 }}},
            { letter = "D", codons = {{ triplet = "GAT", weight = 84985 }, { triplet = "GAC", weight = 52486 }}}
        },
        translation_map = {},  -- Will be filled by init
        start_codon_table = {} -- Will be filled by init
    },
    ["scerevisiae"] = {
        start_codons = {"TTG", "CTG", "ATG"},
        stop_codons = {"TAA", "TAG", "TGA"},
        amino_acids = {
            { letter = "F", codons = {{ triplet = "TTT", weight = 69255 }, { triplet = "TTC", weight = 47162 }}},
            { letter = "L", codons = {{ triplet = "TTA", weight = 69592 }, { triplet = "TTG", weight = 69917 }, { triplet = "CTT", weight = 32571 }, { triplet = "CTC", weight = 14569 }, { triplet = "CTA", weight = 35726 }, { triplet = "CTG", weight = 27963 }}},
            { letter = "S", codons = {{ triplet = "TCT", weight = 61396 }, { triplet = "TCC", weight = 36989 }, { triplet = "TCA", weight = 50416 }, { triplet = "TCG", weight = 23012 }, { triplet = "AGT", weight = 38486 }, { triplet = "AGC", weight = 26233 }}},
            { letter = "Y", codons = {{ triplet = "TAT", weight = 50491 }, { triplet = "TAC", weight = 38471 }}},
            { letter = "K", codons = {{ triplet = "AAA", weight = 113360 }, { triplet = "AAG", weight = 80638 }}},
            { letter = "W", codons = {{ triplet = "TGG", weight = 27342 }}},
            { letter = "V", codons = {{ triplet = "GTT", weight = 56557 }, { triplet = "GTC", weight = 29536 }, { triplet = "GTA", weight = 31958 }, { triplet = "GTG", weight = 28127 }}},
            { letter = "*", codons = {{ triplet = "TAA", weight = 2561 }, { triplet = "TAG", weight = 1243 }, { triplet = "TGA", weight = 1614 }}},
            { letter = "C", codons = {{ triplet = "TGT", weight = 20749 }, { triplet = "TGC", weight = 12623 }}},
            { letter = "P", codons = {{ triplet = "CCT", weight = 35886 }, { triplet = "CCC", weight = 18238 }, { triplet = "CCA", weight = 47055 }, { triplet = "CCG", weight = 14424 }}},
            { letter = "Q", codons = {{ triplet = "CAA", weight = 71668 }, { triplet = "CAG", weight = 32701 }}},
            { letter = "I", codons = {{ triplet = "ATT", weight = 79418 }, { triplet = "ATC", weight = 44763 }, { triplet = "ATA", weight = 48474 }}},
            { letter = "M", codons = {{ triplet = "ATG", weight = 54773 }}},
            { letter = "T", codons = {{ triplet = "ACT", weight = 52985 }, { triplet = "ACC", weight = 32808 }, { triplet = "ACA", weight = 47831 }, { triplet = "ACG", weight = 21401 }}},
            { letter = "A", codons = {{ triplet = "GCT", weight = 53513 }, { triplet = "GCC", weight = 31917 }, { triplet = "GCA", weight = 42870 }, { triplet = "GCG", weight = 16338 }}},
            { letter = "D", codons = {{ triplet = "GAT", weight = 100401 }, { triplet = "GAC", weight = 53688 }}},
            { letter = "E", codons = {{ triplet = "GAA", weight = 120741 }, { triplet = "GAG", weight = 51544 }}},
            { letter = "G", codons = {{ triplet = "GGT", weight = 59816 }, { triplet = "GGC", weight = 25856 }, { triplet = "GGA", weight = 29566 }, { triplet = "GGG", weight = 16019 }}},
            { letter = "H", codons = {{ triplet = "CAT", weight = 36801 }, { triplet = "CAC", weight = 20528 }}},
            { letter = "R", codons = {{ triplet = "CGT", weight = 16654 }, { triplet = "CGC", weight = 7037 }, { triplet = "CGA", weight = 8246 }, { triplet = "CGG", weight = 4843 }, { triplet = "AGA", weight = 55776 }, { triplet = "AGG", weight = 25020 }}},
            { letter = "N", codons = {{ triplet = "AAT", weight = 96750 }, { triplet = "AAC", weight = 65451 }}}
        },
        translation_map = {},  -- Will be filled by init
        start_codon_table = {} -- Will be filled by init
    },
    ["homo_sapiens"] = {
        start_codons = {"ATG"},  -- Typically just ATG in humans
        stop_codons = {"TAA", "TAG", "TGA"},
        amino_acids = {
            { letter = "F", codons = {{ triplet = "TTT", weight = 714298 }, { triplet = "TTC", weight = 824692 }}},
            { letter = "L", codons = {{ triplet = "TTA", weight = 311881 }, { triplet = "TTG", weight = 525688 }, { triplet = "CTT", weight = 536515 }, { triplet = "CTC", weight = 796638 }, { triplet = "CTA", weight = 290751 }, { triplet = "CTG", weight = 1611801 }}},
            { letter = "I", codons = {{ triplet = "ATT", weight = 650473 }, { triplet = "ATC", weight = 846466 }, { triplet = "ATA", weight = 304565 }}},
            { letter = "M", codons = {{ triplet = "ATG", weight = 896005 }}},
            { letter = "V", codons = {{ triplet = "GTT", weight = 448607 }, { triplet = "GTC", weight = 588138 }, { triplet = "GTA", weight = 287712 }, { triplet = "GTG", weight = 1143534 }}},
            { letter = "S", codons = {{ triplet = "TCT", weight = 618711 }, { triplet = "TCC", weight = 718892 }, { triplet = "TCA", weight = 496448 }, { triplet = "TCG", weight = 179419 }, { triplet = "AGT", weight = 493429 }, { triplet = "AGC", weight = 791383 }}},
            { letter = "P", codons = {{ triplet = "CCT", weight = 713233 }, { triplet = "CCC", weight = 804620 }, { triplet = "CCA", weight = 688038 }, { triplet = "CCG", weight = 281570 }}},
            { letter = "T", codons = {{ triplet = "ACT", weight = 533609 }, { triplet = "ACC", weight = 768147 }, { triplet = "ACA", weight = 614523 }, { triplet = "ACG", weight = 246105 }}},
            { letter = "A", codons = {{ triplet = "GCT", weight = 750096 }, { triplet = "GCC", weight = 1127679 }, { triplet = "GCA", weight = 643471 }, { triplet = "GCG", weight = 299495 }}},
            { letter = "Y", codons = {{ triplet = "TAT", weight = 495699 }, { triplet = "TAC", weight = 622407 }}},
            { letter = "*", codons = {{ triplet = "TAA", weight = 40285 }, { triplet = "TAG", weight = 32109 }, { triplet = "TGA", weight = 63237 }}},
            { letter = "H", codons = {{ triplet = "CAT", weight = 441711 }, { triplet = "CAC", weight = 613713 }}},
            { letter = "Q", codons = {{ triplet = "CAA", weight = 501911 }, { triplet = "CAG", weight = 1391973 }}},
            { letter = "N", codons = {{ triplet = "AAT", weight = 689701 }, { triplet = "AAC", weight = 776603 }}},
            { letter = "K", codons = {{ triplet = "AAA", weight = 993621 }, { triplet = "AAG", weight = 1295568 }}},
            { letter = "D", codons = {{ triplet = "GAT", weight = 885429 }, { triplet = "GAC", weight = 1020595 }}},
            { letter = "E", codons = {{ triplet = "GAA", weight = 1177632 }, { triplet = "GAG", weight = 1609975 }}},
            { letter = "C", codons = {{ triplet = "TGT", weight = 430311 }, { triplet = "TGC", weight = 513028 }}},
            { letter = "W", codons = {{ triplet = "TGG", weight = 535595 }}},
            { letter = "R", codons = {{ triplet = "CGT", weight = 184609 }, { triplet = "CGC", weight = 423516 }, { triplet = "CGA", weight = 250760 }, { triplet = "CGG", weight = 464485 }, { triplet = "AGA", weight = 494682 }, { triplet = "AGG", weight = 486463 }}},
            { letter = "G", codons = {{ triplet = "GGT", weight = 437126 }, { triplet = "GGC", weight = 903565 }, { triplet = "GGA", weight = 669873 }, { triplet = "GGG", weight = 669768 }}}
        },
        translation_map = {},  -- Will be filled by init
        start_codon_table = {} -- Will be filled by init
    }
}

-- Initialize translation maps and start codon tables for all default tables
local function init_translation_maps(t: TranslationTable)
    -- Create translation map
    for _, aa in ipairs(t.amino_acids) do
        for _, cdn in ipairs(aa.codons) do
            t.translation_map[cdn.triplet] = aa.letter
        end
    end
    
    -- Create start codon table
    for _, cdn in ipairs(t.start_codons) do
        t.start_codon_table[cdn] = "M"
    end
	t.translate = translate
	t.optimize = optimize
    t.standardize_last_codon = standardize_last_codon

	return
end

-- Initialize all default tables
for _, t in pairs(codon.default_tables) do
    init_translation_maps(t)
end

return codon
