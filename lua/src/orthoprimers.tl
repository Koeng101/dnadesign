--[[ Module orthoprimers
Package orthoprimers contains functions for working with dial-out primers.

The idea of dial-out primers is that, given we have a set of mutually
orthogonal primer binding sites, we can use a relatively small number of
oligos to specify a certain amplicon or set of amplicons from a large mixed
pool.

While they identify 166 orthogonal primers, we only use 96 of them, mostly
because it is physically easier to handle a single 96 well plate for index
creation. It also lowers the quantity of oligos that needs to be synthesized.
]]

local record orthoprimers
    -- Core functions for primer pair management
    make_primer_pair_key: function(forward: string, reverse: string): string
    new_orthogonal_primer_set: function(primers: {string}): OrthogonalPrimerSet
    new_default_orthogonal_primer_set: function(): OrthogonalPrimerSet
    
    -- Types
    PrimerPair: PrimerPair
    OrthogonalPrimerSet: OrthogonalPrimerSet
    
    -- Data
    ortho_primers: {string}
end

-- PrimerPair is a struct containing a pair of primers. Make sure to use
-- MakePrimerPair to produce, since we assume a certain DNA directionality
-- to correctly sort the maps. Note: normally, you should not be using these
-- structs at all, you should be just calling
-- `OrthogonalPrimerSet.NewPrimerSet()`.
local record PrimerPair
    forward: string
    reverse: string
end

-- OrthogonalPrimerSet is a struct containing a list of mutually orthogonal
-- primers and their current usage statistics in order to pick the next most
-- optimal pairing.
local record OrthogonalPrimerSet
    primers: {string}
    primer_use_quantity: {string:integer}
    primer_pairs: {string:boolean}
    
    -- Method for generating new primer pairs
    new_primer_set: function(self: OrthogonalPrimerSet): string, string, string
end

local ortho_primers: {string} = {
    "AAACACGTGGCAAACATTCC",
    "AAACCGGAGCCATACAGTAC",
    "AAAGCACTCTTAGGCCTCTG",
    "AAAGGGGCCGTCAATATCAG",
    "AAATAAGACGACGACCCTCG",
    "AACGATGATGCTCACTCTCG",
    "AAGAATTACTGACCCCTCGG",
    "AAGACGATCCGAGCCATTAC",
    "AAGGAACTATGGCATCGAGC",
    "AAGGACTGCATACCAGGTTG",
    "AAGGATATGTAGACACCGCC",
    "AAGGCCCAGAAGGATACAAC",
    "AAGGCGCTCGGATAATACTC",
    "AAGGTATGTATAGCGACCGC",
    "AATAGGAACCTCTTACGCGG",
    "AATATCACGCAAAAGCACCG",
    "AATCAGTTTCTTTGGCAGCC",
    "AATGCAAAGCTATTAGCGCG",
    "AATGCGTCATTTTACACGGC",
    "AATGTCCTTAGGCAGTCGTC",
    "ACAACGAGCAGACCGAATAG",
    "ACAAGGAGTCGGCATATCAC",
    "ACAGAACGAACAGGCACTAC",
    "ACAGGAAGCAAGGTATACGC",
    "ACAGGGTATATTGAGTGCCC",
    "ACATAAGCGATCCCAAGGTC",
    "ACATCGCATACCAGAACAGG",
    "ACATTAAATTTCGCCGTGGC",
    "ACCACAGGTCAAGATTCACG",
    "ACCCGTATCGCATAAGGATG",
    "ACGAGATGATGCACCGATAG",
    "ACGATGGGGACATAGAACAC",
    "ACGGAGCCCTTATTGTAACC",
    "ACGTATGGGGAACACTACAC",
    "ACGTGAAACTGTATCGAGCC",
    "ACGTTCAGTTTTCCAATGGC",
    "ACTAGATTAGCAAGGCACCC",
    "ACTGGACCCAATAAAAGGCC",
    "ACTTCGATTGGCAAGGACTG",
    "AGAACATAGCATTCACGGGG",
    "AGACAACAATCTGAGGCTGG",
    "AGACAAGCCTTAACCGTAGG",
    "AGACACAAGGCTGATTCCAG",
    "AGACATGGGATTGACCACAC",
    "AGAGAGGCATGATTGACCTC",
    "AGAGTTGCACCTAGAATCCG",
    "AGATAGATGCTCCGTCAAGC",
    "AGATAGTCACGCACAAGACC",
    "AGATTAGCCGACTTTCCTGG",
    "AGATTAGCTGCCGATACTGG",
    "AGATTGTTACTCCGACGGAC",
    "AGATTTCCGACGAGATTCCC",
    "AGCATCCGTCTAAATCTCGG",
    "AGCTATAAGAATTGCCGGGC",
    "AGCTATGATCCCGGTGTAAC",
    "AGCTCAATCTAACAGTGGGG",
    "AGGACACCAGACCAATGAAG",
    "AGGGCTAATTACCATCAGCG",
    "AGGTGATCTGACGAATGTCC",
    "AGTAAAGCATAGTGCCCAGC",
    "AGTAGTATCCGAATCGCTGC",
    "AGTATCTCAGCAAGGGCAAC",
    "AGTATTAGGCGTCAAGGTCC",
    "AGTATTCTTACAGCCAGCCG",
    "AGTATTGCCGGACTAAACCC",
    "AGTCCCAAGTTCAGACGTAC",
    "AGTCCGACACAATGTGACAC",
    "AGTGAACTGACCGAATCCTC",
    "AGTGGTCTGTAAACCGTACC",
    "AGTGTTTTCCATTTTCCGCG",
    "AGTTATAAGGGTCCGATGCC",
    "AGTTGCAGTATCTAACCCGC",
    "AGTTGTAATATCACCCGCGC",
    "ATACGTGGCTAGCATGAGAC",
    "ATACTGTAAGAACCACGCGG",
    "ATAGATCATGTCGGCAGTCG",
    "ATAGATGGTGCCTACATGCG",
    "ATCACAACAAAGGACGGGTC",
    "ATCAGACAACACAGAGGCTG",
    "ATCCAGGAGGTCTAGGAACC",
    "ATCCTAGAAAAGGCGAAGGC",
    "ATGCCATGACGACAACTAGC",
    "ATGCTAGCTGGAACTATCGG",
    "ATTAGGATTGCGAGCGACAC",
    "ATTAGTACACTCCGTGAGCG",
    "ATTCAAGGGTTGGACGACTC",
    "ATTCTCACGACGCAAGATGG",
    "ATTGACGGGAACTACACTCG",
    "CACTCGATAGGTACAACCGG",
    "CAGACCTACGGATCTTAGCG",
    "CCACGAGATAAGAGGATGGC",
    "CCAGAGCTTAGGGGACATAC",
    "CCCGAGGGGAGAAATATACC",
    "CCGAGGGAACCATGATACAG",
    "CCGGGAGGAAGATATAGCAC",
    "CCGGTTGTACCTATCGAGTG",
    "CCGTGCGACAAGATTTCAAG",
    "CCTTTAACAGGACATGCAGC",
    "CGAACGCAAAAGTCCTCAAG",
    "CGATAGAACGACCAGGTAGC",
    "CGGATCGAACTTAGGTAGCC",
    "CGGGAGGAAGTCTTTAGACC",
    "CTAATATCCCTGAGCGACGG",
    "CTAGGGAACCAGGCTTAACG",
    "CTAGGGGATGGTCCAATACG",
    "CTATAGAATCCGGGCTGGTC",
    "CTGCTAGGGGCTACTTATCG",
    "GAAAAGTCCCAATGAGTGCC",
    "GAAGTGGTTTGCCTAAACGC",
    "GACCATGCAAGGAGAGGTAC",
    "GATACATAGACTTGGCCCCG",
    "GCACGCAAAAGGACATAACC",
    "GCAGCGTTTTAGCCTACAAG",
    "GCATAAAGTTGACAGGCCAG",
    "GCTAAATAGAGGGAAGCCCC",
    "GGAAAACTAAGACAAGGCGC",
    "GGAAACAATAACCATCGGCG",
    "GGGCACCGATTAAGAAATGC",
    "GGGTTGTCTCCTCTGATAGC",
    "GTACTCAGAGATTGCCGGAG",
    "GTATAAGATCAGCCGGACCC",
    "GTATGTCGGCTCTCGTATCG",
    "GTTCAGAGGTACGAACCCTC",
    "GTTGCATCTAAGCCAAGTGC",
    "TAAAGAGAGGGCGTCCAATC",
    "TAACGACGTGCCGAACTTAG",
    "TAAGATAGCACCACGGATGG",
    "TAAGGATTCATCAGGTGCGC",
    "TAAGGGACGATGCTTAACCC",
    "TACCACGAAATGCACAGGAG",
    "TACTGATAATTCGGACGCCC",
    "TACTTGAATACCACGTGGCC",
    "TAGCCAGGCAAAAGAGATCC",
    "TAGCTCGATAATCAAGGGGC",
    "TAGTGACCTAATGCCATGGG",
    "TAGTTGAGAACACGAACCCG",
    "TATAACAGGCTGCTGAGACC",
    "TATACTGAAGAACGGCCCAG",
    "TATCAATCCGGAACCAGTGC",
    "TATCACGGAAGGACTCAACG",
    "TCAAAGGAGCACGAACCTAC",
    "TCAAGGTCCGTTATGGAACC",
    "TCACATAGAAGGACATGGCG",
    "TCACTTGGTATCGAGAACGG",
    "TCAGCCTTTCATTGATTGCG",
    "TCATCGACAAGATACAGGCG",
    "TCCAATTATACGGAGCAGGC",
    "TCGAATATGCTGTAACCCCG",
    "TCGACCAGGTTATCATGAGC",
    "TCGAGACAAGAACGATTCCC",
    "TCTAGGACTATCACCGGAGG",
    "TCTTCATAAGCCAGAGTGCC",
    "TCTTGCGATAGACACAAGCC",
    "TGAGCCATAAAAGCAAAGCG",
    "TGAGCGCAGAACTATCAGAC",
    "TGCATAGTATCCCAACAGGG",
    "TGCCAAAGGGTAGAGACATC",
    "TGCTGAATGAGAAACCTCGG",
    "TGGGGACGACTTATAATGCC",
    "TGTGGACCCTATCAAACGAG",
    "TTAGCTCAGGTCCAAAGTCC",
    "TTAGTAGGCAAGCATACCCG",
    "TTCGGGAGCGGATTATACAC",
    "TTCTGGGACTGGATAACACG",
    "TTGACAGACAATCCGTAGGC"
}

local function new_primer_set(self: OrthogonalPrimerSet): string, string, string
    -- Pick two primers
    local forward_primer: string = ""
    local reverse_primer: string = ""

    -- First pick forward primer from least used in list
    local least_used: integer = 1000000
    for i = #self.primers, 1, -1 do
        if self.primer_use_quantity[self.primers[i]] <= least_used then
            forward_primer = self.primers[i]
            least_used = self.primer_use_quantity[self.primers[i]]
        end
    end
    
    -- Now that we have a forward primer, get a list of reverse primers
    -- that are available for that particular forward primer
    local possible_reverse_primers: {string} = {}
    for _, primer in ipairs(self.primers) do
        local pair_key = orthoprimers.make_primer_pair_key(forward_primer, primer)
        if not self.primer_pairs[pair_key] and forward_primer ~= primer then
            table.insert(possible_reverse_primers, primer)
        end
    end
    
    -- From those possibleReversePrimers, get the least used
    least_used = 1000000
    for i = #possible_reverse_primers, 1, -1 do
        if self.primer_use_quantity[possible_reverse_primers[i]] <= least_used then
            reverse_primer = possible_reverse_primers[i]
            least_used = self.primer_use_quantity[possible_reverse_primers[i]]
        end
    end
    
    -- Double check we have a match
    if forward_primer == "" or reverse_primer == "" then
        return "", "", "Not enough primers for genes in pool"
    end
    
    -- If there is a match, add both to list!
    self.primer_use_quantity[forward_primer] = self.primer_use_quantity[forward_primer] + 1
    self.primer_use_quantity[reverse_primer] = self.primer_use_quantity[reverse_primer] + 1
    self.primer_pairs[orthoprimers.make_primer_pair_key(forward_primer, reverse_primer)] = true
    
    return forward_primer, reverse_primer, nil
end

-- MakePrimerPairKey makes a primer pair key from a forward and reverse primer. It
-- is done within a function to make sure of consistent forward+reverse
-- pairing, which is essential for Lua to map the pair properly. Since Lua
-- tables are compared by reference, not value, we use a string key instead.
function orthoprimers.make_primer_pair_key(forward: string, reverse: string): string
    if forward > reverse then
        return forward .. "|" .. reverse
    end
    return reverse .. "|" .. forward
end

-- NewOrthogonalPrimerSet generates a new primer set from a given set of
-- orthogonal primers.
function orthoprimers.new_orthogonal_primer_set(primers: {string}): OrthogonalPrimerSet
    local quantity_map: {string:integer} = {}
    for _, primer in ipairs(primers) do
        quantity_map[primer] = 0
    end
    local primer_pairs: {string:boolean} = {}
    return {primers = primers, primer_use_quantity = quantity_map, primer_pairs = primer_pairs, new_primer_set=new_primer_set}
end

-- NewDefaultOrthogonalPrimerSet generates a default set of primers. The
-- default primer set is the first 96 orthogonal primers from the paper "A set
-- of experimentally validated, mutually orthogonal primers for combinatorially
-- specifying genetic components". The first 96 are used because it makes
-- handling the plate easier.
--
--	https://doi.org/10.1093/synbio/ysx008
function orthoprimers.new_default_orthogonal_primer_set(): OrthogonalPrimerSet
    local first_96: {string} = {}
    for i = 1, 96 do
        first_96[i] = ortho_primers[i]
    end
    return orthoprimers.new_orthogonal_primer_set(first_96)
end

orthoprimers.ortho_primers = ortho_primers

return orthoprimers

