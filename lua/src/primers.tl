--[[ Module primers
Module primers provides utilities for creating primers and DNA barcodes.

Primers are short sequences of DNA that can be used to amplify DNA sequences
and they are the workhorse of modern molecular biology.

Essentially primers are short pieces of single stranded DNA that can
bind to a target sequence of single stranded DNA. These primers serve as a
marker for polymerases to bind and start adding free floating nucleotides
(ACTGs) to a single strand piece of DNA to form a double stranded piece of DNA.
]]

local record primers
    -- Core functions for primer analysis
    santa_lucia: function(sequence: string, primer_concentration: number, salt_concentration: number, magnesium_concentration: number): number, number, number
    marmur_doty: function(sequence: string): number
    melting_temp: function(sequence: string): number

    -- Functions for generating DNA barcodes
    nucleobase_debruijn_sequence: function(substring_length: number): string
    create_barcodes: function(length: integer, max_subsequence: integer): {string}
    create_barcodes_with_banned_sequences: function(length: integer, max_subsequence: integer, banned_sequences: {string}, banned_functions: {function(string): boolean}): {string}
    create_barcodes_gc_range: function(length: integer, max_subsequence: integer, min_gc_content: number, max_gc_content: number): {string}

	-- Types
	Thermodynamics: Thermodynamics
end

local transform = require("transform")

-- Record to store thermodynamic values for nucleotide pairs
local record Thermodynamics
    h: number  -- enthalpy (dH, kcal/mol)
    s: number  -- entropy (dS, cal/mol-K)
end

--[[
This section contains various penalties applied when calculating primer melting
temperature using the SantaLucia algorithm.
--]]

-- Penalties for nearest neighbor effects
local nearest_neighbors_thermodynamics: {string:Thermodynamics} = {
    ["AA"] = { h = -7.6, s = -21.3 },
    ["TT"] = { h = -7.6, s = -21.3 },
    ["AT"] = { h = -7.2, s = -20.4 },
    ["TA"] = { h = -7.2, s = -21.3 },
    ["CA"] = { h = -8.5, s = -22.7 },
    ["TG"] = { h = -8.5, s = -22.7 },
    ["GT"] = { h = -8.4, s = -22.4 },
    ["AC"] = { h = -8.4, s = -22.4 },
    ["CT"] = { h = -7.8, s = -21.0 },
    ["AG"] = { h = -7.8, s = -21.0 },
    ["GA"] = { h = -8.2, s = -22.2 },
    ["TC"] = { h = -8.2, s = -22.2 },
    ["CG"] = { h = -10.6, s = -27.2 },
    ["GC"] = { h = -9.8, s = -24.4 },
    ["GG"] = { h = -8.0, s = -19.9 },
    ["CC"] = { h = -8.0, s = -19.9 }
}

local initial_thermodynamic_penalty: Thermodynamics = { h = 0.2, s = -5.7 }      -- penalty for initiating helix
local symmetry_thermodynamic_penalty: Thermodynamics = { h = 0, s = -1.4 }       -- penalty for self-complementarity
local terminal_at_thermodynamic_penalty: Thermodynamics = { h = 2.2, s = 6.9 }   -- penalty for 3' AT

--[[
SantaLucia calculates the melting point of a short DNA sequence (15-200 bp), 
using the Nearest Neighbors method [SantaLucia, J. (1998) PNAS, doi:10.1073/pnas.95.4.1460]
]]
function primers.santa_lucia(sequence: string, primer_concentration: number, salt_concentration: number, magnesium_concentration: number): number, number, number
    sequence = string.upper(sequence)
    
    local gas_constant: number = 1.9872  -- gas constant (cal / mol - K)
    local symmetry_factor: number = 1    -- symmetry factor
    local dh: number = 0                 -- enthalpy
    local ds: number = 0                 -- entropy
    
    -- Apply initialization penalty
    dh = dh + initial_thermodynamic_penalty.h
    ds = ds + initial_thermodynamic_penalty.s
    
    -- Apply symmetry penalty if sequence is self-complementary
    if sequence == transform.reverse_complement(sequence) then
        dh = dh + symmetry_thermodynamic_penalty.h
        ds = ds + symmetry_thermodynamic_penalty.s
        symmetry_factor = 1
    else
        symmetry_factor = 4
    end
    
    -- Apply penalty if 3' nucleotides are A or T
    local last_base = sequence:sub(-1)
    if last_base == "A" or last_base == "T" then
        dh = dh + terminal_at_thermodynamic_penalty.h
        ds = ds + terminal_at_thermodynamic_penalty.s
    end
    
    -- Apply salt penalty (von Ahsen et al 1999)
    local salt_effect = salt_concentration + (magnesium_concentration * 140)
    ds = ds + (0.368 * (#sequence - 1) * math.log(salt_effect))
    
    -- Calculate penalty for nearest neighbor effects
    for i = 1, #sequence - 1 do
        local pair = sequence:sub(i, i + 1)
        local dt = nearest_neighbors_thermodynamics[pair]
        if dt then
            dh = dh + dt.h
            ds = ds + dt.s
        end
    end
    
    local melting_temp = dh * 1000 / (ds + gas_constant * math.log(primer_concentration/symmetry_factor)) - 273.15
    return melting_temp, dh, ds
end

--[[
MarmurDoty calculates the melting point of an extremely short DNA sequence (<15 bp) 
using a modified Marmur Doty formula [Marmur J & Doty P (1962). Determination of the base 
composition of deoxyribonucleic acid from its thermal denaturation temperature. J Mol Biol, 5, 109-118.]
]]
function primers.marmur_doty(sequence: string): number
    sequence = string.upper(sequence)
    
    local a_count = 0
    local t_count = 0
    local c_count = 0
    local g_count = 0
    
    for i = 1, #sequence do
        local base = sequence:sub(i, i)
        if base == "A" then a_count = a_count + 1
        elseif base == "T" then t_count = t_count + 1
        elseif base == "C" then c_count = c_count + 1
        elseif base == "G" then g_count = g_count + 1
        end
    end
    
    local melting_temp = 2 * (a_count + t_count) + 4 * (c_count + g_count) - 7.0
    return melting_temp
end

--[[
MeltingTemp calls SantaLucia with default inputs for primer and salt concentration.
]]
function primers.melting_temp(sequence: string): number
    local primer_concentration = 500e-9  -- 500 nM (nanomolar) primer concentration
    local salt_concentration = 50e-3     -- 50 mM (millimolar) sodium concentration
    local magnesium_concentration = 0.0  -- 0 mM (millimolar) magnesium concentration
    
    local melting_temp = primers.santa_lucia(sequence, primer_concentration, salt_concentration, magnesium_concentration)
    return melting_temp
end

--[[
May 23 2021

Start of the De Bruijn stuff

=== Barcode basics ===

We're rapidly getting better at sequencing a lot of DNA. At their core, most
DNA sequencing technologies pool together many samples and sequence them all
at once. For example, let's say we have 2 samples of DNA whose true sequence
is as follows:

DNA-1 := ATGC
DNA-2 := AGGC

If we pooled these two samples together into a single tube, and sequenced
them, we would not be able to tell if ATGC came from DNA-1 or DNA-2. In order
to tell the difference, we would have to go through the process of DNA
barcoding. Let's attach(2) two small barcodes to each DNA fragment separately
in their own tubes and then pool them together:

Barcode-1 + DNA-1 = GC + ATGC = GCATGC
Barcode-2 + DNA-2 = AT + AGGC = ATAGGC

When we sequence this pool together, we will end up with two sequences,
GCATGC and ATAGGC. If we correlate the first 2 base pairs with the tube
the sample came from, we can derive DNA-1 is ATGC and DNA-2 is AGGC.



=== Redundancy and start sites ===

Now, let's say we have the need for N number of samples to be pooled
together. The minimal barcode length could be expressed as:

n = number of samples
b = bases required in a minimal barcode

4^b = n
OR
log4n = b

In our perfect case, we would only need 8 base pair barcodes to represent 65536
different samples. Reality is a little different, however.

1. Failure of DNA sequencers to accurately sequence the barcode, leading for
   one barcode to be mistaken for a different barcode
2. Misalignment of the barcode to the sequence. We cannot guarantee that the
   DNA sequencer will begin sequencing our fragment at an exact base pair.
3. Misreading of sequence as barcode. If our barcode is only 8 base pairs, on
   average, it will occur once within 65536 base pairs, and that occurrence may
   be misread as a barcode.

These challenges force us to build a barcode that has the following features:

1. Any barcode must be different enough from any other barcode that there will
   be no misreading, even with mutated base pairs.
2. Any barcode must be large enough that, on average, it will not occur in a
   natural piece of DNA.

While the second feature is quite easy (use ~20-30 base pair barcodes), the
first can be challenging. When developing a large quantity of barcodes, how
do you guarantee that they are optimally distanced from each other so that
there will be no cross-talk?

=== Our solution to distanced barcodes ===

De Bruijn sequences are an interesting data structure where every possible
substring of length N occurs exactly once as a substring(1). For example, a De
Bruijn sequence of length 3 will only have ATG occur once in the entire
sequence.

By constructing a nucleobase De Bruijn sequence, and selecting barcodes from
within that De Bruijn sequence, we can guarantee that each barcode will never
share any N length substring, since it only occurs once within the whole De
Bruijn sequence.

For example, a nucleobase De Bruijn sequence of substring length 6 is 4101
base pairs long (4^n + (n-1)). You can generate 205 20 base pair barcodes with
each barcode guaranteed to never share any 6 base pairs. This makes it very
easy to unambiguously parse which samples came from where, while maintaining
a guarantee of optimal distancing between your barcodes.

Good luck with barcoding,

Keoni

(1) https://en.wikipedia.org/wiki/De_Bruijn_sequence
(2) Barcodes are usually added in a process called "ligation" using an enzyme
    called ligase, which basically just glues together DNA fragments. Wikipedia
    has a good introduction:
    https://en.wikipedia.org/wiki/Ligation_(molecular_biology)

--]]

--[[
NucleobaseDeBruijnSequence generates a DNA DeBruijn sequence with alphabet ATGC.
DeBruijn sequences are basically a string with all unique substrings of an 
alphabet represented exactly once.
]]
function primers.nucleobase_debruijn_sequence(substring_length: number): string
    local alphabet = "ATGC"
    local alphabet_length = #alphabet
    local a: {number} = {}

    -- Initialize array with zeros
    for i = 1, alphabet_length * substring_length do
        a[i as integer] = 0
    end

    local seq: {number} = {}

    local function construct_debruijn(t: number, p: number)
        if t > substring_length then
            if substring_length % p == 0 then
                for i = 1, p do
                    table.insert(seq, a[i as integer] or 0)
                end
            end
        else
            local tp_idx = (t - p) as integer
            a[t as integer] = a[tp_idx] or 0
            construct_debruijn(t + 1, p)
            local start_val = (a[tp_idx] or 0)
            for j = start_val + 1, alphabet_length - 1 do
                a[t as integer] = j
                construct_debruijn(t + 1, t)
            end
        end
    end

    construct_debruijn(1, 1)

    local result = ""
    for _, i in ipairs(seq) do
        local idx: integer = (i + 1) as integer
        result = result .. alphabet:sub(idx, idx)
    end

    return result .. result:sub(1, (substring_length - 1) as integer)
end

--[[
CreateBarcodesWithBannedSequences creates a list of barcodes given a desired barcode length, 
the maxSubSequence shared in each barcode. Sequences may be marked as banned by passing a 
static list, `banned_sequences`, or through a list of `banned_functions` that dynamically 
generates banned sequences.
]]
function primers.create_barcodes_with_banned_sequences(length: integer, max_subsequence: integer, banned_sequences: {string}, banned_functions: {function(string): boolean}): {string}
    local barcodes: {string} = {}
    local start = 1
    local finish = 1
    local debruijn = primers.nucleobase_debruijn_sequence(max_subsequence)
    
    local barcode_num = 0
    while (barcode_num * (length - (max_subsequence - 1))) + length <= #debruijn do
        start = barcode_num * (length - (max_subsequence - 1)) + 1
        finish = start + length - 1
        barcode_num = barcode_num + 1
        
        -- Check banned sequences
        for _, banned_sequence in ipairs(banned_sequences) do
            -- Check if current range contains banned sequence
            while debruijn:sub(start, finish):find(banned_sequence, 1, true) do
                if finish + 1 > #debruijn then
                    return barcodes
                end
                start = start + 1
                finish = finish + 1
                barcode_num = barcode_num + 1
            end
            
            -- Check reverse complement
            while debruijn:sub(start, finish):find(transform.reverse_complement(banned_sequence), 1, true) do
                if finish + 1 > #debruijn then
                    return barcodes
                end
                start = start + 1
                finish = finish + 1
                barcode_num = barcode_num + 1
            end
        end
        
        -- Check banned functions
        for _, banned_function in ipairs(banned_functions) do
            while not banned_function(debruijn:sub(start, finish)) do
                if finish + 1 > #debruijn then
                    return barcodes
                end
                start = start + 1
                finish = finish + 1
                barcode_num = barcode_num + 1
            end
        end
        
        table.insert(barcodes, debruijn:sub(start, finish))
    end
    
    return barcodes
end

--[[
CreateBarcodes is a simplified version of CreateBarcodesWithBannedSequences with sane defaults.
]]
function primers.create_barcodes(length: integer, max_subsequence: integer): {string}
    return primers.create_barcodes_with_banned_sequences(length, max_subsequence, {}, {})
end

--[[
CreateBarcodesGcRange creates a list of barcodes within a given GC range.
]]
function primers.create_barcodes_gc_range(length: integer, max_subsequence: integer, min_gc_content: number, max_gc_content: number): {string}
    local function gc_content(sequence: string): number
        local gc_count = 0
        sequence = string.upper(sequence)
        for i = 1, #sequence do
            local base = sequence:sub(i, i)
            if base == "G" or base == "C" then
                gc_count = gc_count + 1
            end
        end
        return gc_count / #sequence
    end
    
    local function gc_barcode_func(barcode_to_check: string): boolean
        local gc = gc_content(barcode_to_check)
        if gc < min_gc_content or gc > max_gc_content then
            return false
        end
        return true
    end
    
    return primers.create_barcodes_with_banned_sequences(length, max_subsequence, {}, {gc_barcode_func})
end

return primers
