-- dnadesign.d.tl

--[[ Module rng
rng.tl - A Deterministic Random Number Generator for Hermetic Environments

This module implements L'Ecuyer's Combined Linear Congruential Generator (LCG) to provide
a completely deterministic, reproducible random number generation system. It's designed
to replace Lua's built-in math.random functionality for environments requiring
hermeticity (complete reproducibility and isolation from system state).

Mathematical Background:
- Uses two independent LCG streams with carefully chosen prime moduli:
  1. m1 = 2147483647 (2^31 - 1, a Mersenne prime)
  2. m2 = 2147483399 (a prime slightly smaller than m1)

- Multipliers chosen for full-period generation:
  1. a1 = 48271 (for m1)
  2. a2 = 40692 (for m2)

These numbers were selected because they:
- Fit well within Lua's 53-bit precision for integers
- Provide a proven period of approximately 2^46
- Pass spectral tests for good distribution of generated numbers
- Allow for efficient modulo operations (especially m1 being a Mersenne prime)

The generator combines two independent streams to break up lattice structures
that can occur in single-stream LCGs. This provides better statistical properties
than simple LCGs while maintaining complete determinism.

Usage:
    local rng = require("hermetic_rng")
    local num = rng.random()     -- returns [0,1)
    local num = rng.random(N)    -- returns [1,N]
    local num = rng.random(M,N)  -- returns [M,N]

Environment Variables:
    SEED_1 - First seed for the generator (defaults to 12345)
    SEED_2 - Second seed for the generator (defaults to 67890)

Note: This module deliberately overrides math.random and math.randomseed to ensure
      hermetic behavior. Attempts to call math.randomseed will raise an error.

Question: Does this differ among hardware platforms?
Answer: By carefully keeping all arithmetic within 2^53, this PRNG remains
consistent across IEEE-754 double-precision platforms (including x86 and ARM).
The largest product here is on the order of 1×10^14, whereas 2^53 is about
9x10^15 - so intermediate multiplications (Before taking modulus) and the floor
operations used in '%' remain exact for all valid states. Consequently, the
combined LCG sequence and the final scaling for random() are reproducible under
any IEEE-754-compliant Lua VM, ensuring hermetic behavior without
hardware-dependent variation.
]]


local record RNGState
    x: integer
    y: integer
end

local record rng
    -- Core generator functions
    new: function(seed_x: integer, seed_y: integer): RNGState
    next: function(state: RNGState): (integer, RNGState)
    random: function(state: RNGState, min?: integer, max?: integer): (number, RNGState)
    
    -- Current state
    state: RNGState
    
	-- Matches all possible math.random call patterns:
    -- math.random()     -> [0,1)
    -- math.random(n)    -> [1,n]
    -- math.random(m,n)  -> [m,n]
    random_wrapper: function(m?: integer, n?: integer): number
end

--[[ Module hash
Module hash contains implementations of common hash algorithms. While normally
these would be imported from the system (and would be much faster there), these
are to ensure cross-platform compatibility.

TODO: sha256 should be able to return more than 32 bits
]]


local interface Hash32
    reset: function(Hash32)
    write: function(Hash32, string)
    sum: function(Hash32): {integer}
	sum32: function(Hash32): integer
end

local record hash
    bit32: bit32
    new_crc32: function(): CRC32
    new_sha256: function(): SHA256
    Hash32: Hash32
end

local record bit32
	band: function(a: integer, b: integer): integer
	bxor: function(a: integer, b: integer): integer
	rshift: function(a: integer, b: integer): integer
	lshift: function(a: integer, b: integer): integer
	bor: function(a: integer, b: integer): integer
	bnot: function(a: integer): integer
end

local record CRC32 is Hash32
    value: integer
end

local record SHA256State
    digest: {integer}
    count_lo: integer
    count_hi: integer
    data: {integer}
    local_: integer
    digestsize: integer
end

local record SHA256 is Hash32
    h: {integer}
    length: integer
    buffer: string
	state: SHA256State
end

--[[ Module transform
Module transform provides functions for transforming sequences.
]]


local record transform
    -- DNA/RNA sequence transformations
    reverse_complement: function(sequence: string): string
    complement: function(sequence: string): string
    reverse: function(sequence: string): string
    complement_base: function(base_pair: string): string
    
    -- RNA-specific transformations
    reverse_complement_rna: function(sequence: string): string
    complement_rna: function(sequence: string): string 
    complement_base_rna: function(base_pair: string): string

	-- Windowing functions
	get_window: function(target_sequence: string, left_flank: string, right_flank: string, circular: boolean): string
end

--[[ Module align
Module align is a module for aligning (comparing) DNA, RNA, and protein sequences.

Biology is fickle and full of quirks that make it hard to do even the most basic of tasks
which we would normally take for granted when working with other kinds of data.

Comparing two biological sequences to see if they're roughly equivalent is one of those tasks.

Essentially two almost identical sequences with almost identical functionality can contain
small insertions or deletions that shift the entire string such that a meaningful comparison via
hamming distance or levenshtein distance becomes impossible.

For example:

Timothy Stiles
||||||| ||||||
Timothy Stiles

is an easy match with hamming or levenshtein distance.

However, say we introduce a new character to the beginning of the sequence.

Timothy Stiles
xxxxxxxxxxxxxxxx
A Timothy Stiles

Now our edit distance via levenshtein is maximized at 16 and we wouldn't
be able to tell that semantically these strings are almost identical.

This frame shifting seen above is incredibly common within biological sequences and alignment
algorithms are designed in part to deal with these shifts so that when we compare two sequences
like the two below we can get a more useful edit distance.

GAAAAAAT
GAA----T

As of writing this module includes the two most basic algorithms for alignment,
Needleman-Wunsch and Smith-Waterman. Needleman-Wunsch is used when you are
looking for global alignment between two full-length sequences, while
Smith-Waterman is better at smaller sequences with local similarities and
handling sequences with long non-homologous regions. BLAST, on the other hand,
takes advantage of more heuristic techniques to speed up alignment, and is better
at finding similar sequences in large database, sacrificing precision for faster
results.

Both are "dynamic programming algorithms" which is a fancy 1980's term for they use
matrices. If you're familiar with kernel operations, linear filters, or whatever term
ML researchers are using nowadays for, "slide a window over a matrix and determine that
entry's values using its neighbor's values", then this should be pretty easy to grok.

If not these algorithms essentially compare every character in one sequence with another
sequence and create an edit distance along with human readable string to show gaps like the
previous example.

I'm not really an expert on alignment so if you want to learn more about this class of algorithms
wikipedia has a decent overview.

https://en.wikipedia.org/wiki/Sequence_alignment

Even if I may not know the answer to your alignment questions please ask and I'll do my best
to help!

TTFN,
Tim
]]


local record align
	-- alignments
    default_matrix: SubstitutionMatrix
    new_scoring: function(substitution_matrix: SubstitutionMatrix | nil, gap_penalty: integer): Scoring
    needleman_wunsch: function(string_a: string, string_b: string, scoring: Scoring): (integer, string, string)
    smith_waterman: function(string_a: string, string_b: string, scoring: Scoring): (integer, string, string)

	-- align_many compares one string against many, using a provided alignment function
	align_many: function(func: function(string_a: string, string_b: string, scoring: Scoring): (integer, string, string), 
		target: string, candidates: {string}, scoring: Scoring, ntop: integer): { {integer, string, string, integer} }
end

local record SubstitutionMatrix
   data: {string:{string:integer}}
end

local record Scoring 
   substitution_matrix: SubstitutionMatrix
   gap_penalty: integer
end

--[[ Module mash
Module mash is for sketching sequence data to make it easier to compare to other sequence.

The module is named mash after the mash sketching algorithm, which is based on the MinHash algorithm.

Mash: fast genome and metagenome distance estimation using MinHash.
Ondov, B.D., Treangen, T.J., Melsted, P. et al.
Genome Biol 17, 132 (2016).
https://doi.org/10.1186/s13059-016-0997-x

Mash Screen: high-throughput sequence containment estimation for genome discovery.
Ondov, B., Starrett, G., Sappington, A. et al.
Genome Biol 20, 232 (2019).
https://doi.org/10.1186/s13059-019-1841-x

tl;dr for the papers above:

Comparing biological sequences is really hard because similar sequences can have frameshifts that make it impossible
to measure similarity using more common metric distances like hamming distance or levenshtein distance.

Bioinformatics and nlp researchers have come up with tons of string comparison algorithms that are better suited for
comparing biological sequences. For example poly already implements a few of them like the Needleman-Wunsch and Smith-Waterman
algorithms in our "align" module.

Mash is a different approach to comparing biological sequences. It uses a technique called sketching to reduce the
complexity of the sequence to a vector of hashes. The hashes are generated by sliding a window of size k along the
sequence and hashing each kmer. The hash is then stored in a vector of size s. The vector is sorted and the smallest
hash is kept. The process is repeated until the vector is full. The vector of hashes is the sketch.

The sketch is then compared to other sketches by counting the number of hashes that are the same between the two sketches.
The number of hashes that are the same is divided by the size of the sketch to get a distance between 0 and 1.

Hash vectors can only be compared to other hash vectors that use the same sliding window size.
Sketch size limits how many hashes can be stored in the vector and the return vector
will always be of length of the sketch size and filled the smallest hashes that were generated
and sorted from smallest to largest.

The larger the sketch size the more accurate the distance calculation will be but the longer it will take to calculate.

TTFN,
Tim

Containment is a slightly different algorithm from similarity. It returns the
amount of kmers from a which are contained in b. This is useful for doing
things like plasmid auto-annotation: features are checked for containment in b.
]]


local record mash
	new: function(kmer_size: integer, sketch_size: integer, hash_algorithm: hash.Hash32): Mash
	sketch: function(sketch_mash: Mash, sequence: string)
	similarity: function(a: Mash, b: Mash): number
	distance: function(a: Mash, b: Mash): number

	new_containment_sketch: function(kmer_size: integer, sequence: string, hash: hash.Hash32): Mash
	containment: function(a: Mash, b: Mash): number
end

local record Mash
    kmer_size: integer  -- The kmer size is the size of the sliding window that is used to generate the hashes.
    sketch_size: integer  -- The sketch size is the number of hashes to store.
    sketches: {integer}  -- The sketches are the hashes of the kmers that we can compare to other sketches.
	hash_algorithm: hash.Hash32
end

--[[ Module seqhash
Module seqhash contains the seqhash algorithm.

This module contains the reference seqhash algorithm in Teal.

There is a big problem with current sequence databases - they all use different
identifiers and accession numbers. This means cross-referencing databases is
a complicated exercise, especially as the quantity of databases increases, or if
you need to compare "wild" DNA sequences.

Seqhash is a simple algorithm to produce consistent identifiers for any genetic sequence. The
basic premise of the Seqhash algorithm is to hash sequences with the hash being a robust
cross-database identifier. Sequences themselves shouldn't be used as a database index
(often, they're too big), so a hash based off of a sequence is the next best thing.

Usability wise, you should be able to Seqhash any rotation of a sequence in any direction and
get a consistent hash.

The Seqhash algorithm makes several opinionated design choices, primarily to make working
with Seqhashes more consistent and nice. The Seqhash algorithm only uses a single hash function,
SHA256 (different from Go implementation which uses Blake3), and only operates on DNA, RNA, and 
Protein sequences. These identifiers will be seen by human beings, so versioning and metadata 
is attached to the front of the hashes so that a human operator can quickly identify problems 
with hashing.

For DNA or RNA sequences, only ATUGCYRSWKMBDHVNZ characters are allowed. For Proteins,
only ACDEFGHIKLMNPQRSTVWYUO*BXZ characters are allowed in sequences. Selenocysteine (Sec; U) and pyrrolysine
(Pyl; O) are included in the protein character set - usually U and O don't occur within protein sequences,
but for certain organisms they do, and it is certainly a relevant amino acid for those particular proteins.
]]


local record seqhash
    version2_flag: function(version: integer, sequence_type: string, circularity: boolean, double_stranded: boolean): integer
    decode_flag: function(flag: integer): (integer, string, boolean, boolean)
    booth_least_rotation: function(sequence: string): integer
    rotate_sequence: function(sequence: string): string
	circular_equality: function(sequence1: string, sequence2: string): boolean
    hash2: function(sequence: string, sequence_type: string, circular: boolean, double_stranded: boolean): ({integer}, string)
    hash2_fragment: function(sequence: string, fwd_overhang_length: integer, rev_overhang_length: integer): ({integer}, string) 
    encode_hash2: function(hash: {integer}, err: string): (string, string)
    decode_hash2: function(encoded_string: string): ({integer}, string)
    encode_to_base58: function(input: {integer}): string
    decode_from_base58: function(input: string): ({integer}, string)

	hash2_metadata: {string:string}
end

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

local record Thermodynamics
    h: number  -- enthalpy (dH, kcal/mol)
    s: number  -- entropy (dS, cal/mol-K)
end

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
    new_orthogonal_primer_set: function(primer_set: {string}): OrthogonalPrimerSet
    new_default_orthogonal_primer_set: function(): OrthogonalPrimerSet
    
    -- Types
    PrimerPair: PrimerPair
    OrthogonalPrimerSet: OrthogonalPrimerSet
    
    -- Data
    ortho_primers: {string}
end

local record PrimerPair
    forward: string
    reverse: string
end

local record OrthogonalPrimerSet
    primers: {string}
    primer_use_quantity: {string:integer}
    primer_pairs: {string:boolean}
    
    -- Method for generating new primer pairs
    new_primer_set: function(self: OrthogonalPrimerSet): string, string, string
end

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


local record pcr
    -- Core PCR functions
    design_primers_with_overhangs: function(sequence: string, forward_overhang: string, reverse_overhang: string, target_tm: number): (string, string)
    design_primers: function(sequence: string, target_tm: number): (string, string)
    simulate_simple: function(sequences: {string}, target_tm: number, circular: boolean, primer_list: {string}): {string}
    simulate: function(sequences: {string}, target_tm: number, circular: boolean, primer_list: {string}): ({string}, string)
end

--[[ Module bio
The bio module provides utilities for reading and writing sequence data.

There are many different biological file formats for different applications.
The bio package provides a consistent way to work with each of these file
formats. The underlying data returned by each parser is as raw as we can return
while still being easy to use for downstream applications, and should be
immediately recognizable as the original format.

TODO: new_file_reader would be convenient function
]]


local interface Reader
    -- read reads up to n bytes from the source
    -- Returns the bytes read and any error
    read: function(self, n: integer): (string, string)

    -- close releases any resources
    close: function(self): (boolean, string)
end

local interface Writer
    -- write writes the given string to the destination
    -- Returns number of bytes written and any error
    write: function(self, data: string): (integer, string)

    -- close releases any resources
    close: function(self): (boolean, string)
end

local interface Header
   -- The format this header is for
   format: Format
   
   -- Convert header to string representation
   to_string: function(self): string
end

local interface Data 
   -- The format this data is for
   format: Format
   
   -- Convert data to string representation
   to_string: function(self): string
end

local interface Parser
   -- Header returns the format-specific header if one exists.
   -- Some formats like FASTA don't have headers and will return nil.
   header: function(self): Header, string -- returns header and error
   
   -- Next returns the next data record from the parser.
   -- Returns nil,"EOF" when reaching end of file.
   next: function(self): Data, string -- returns data and error
   
   -- GetFormat returns the format this parser handles
   format: Format
end

local enum Format
   "FASTA"
   "FASTQ" 
   "GENBANK"
   "SLOW5"
   "SAM"
   "PILEUP"
end

local record bio
	Format: Format
    Reader: Reader
    Writer: Writer
	Header: Header
	Data: Data
	Parser: Parser
	DEFAULT_MAX_LENGTHS: {Format:integer}
	
	StringReader: StringReader
	StringWriter: StringWriter
	BufferedReader: BufferedReader
	new_string_reader: function(content: string): StringReader
	new_string_writer: function(): StringWriter
	new_buffered_reader: function(reader: Reader, max_line_size: integer): BufferedReader
end

local record StringReader is bio.Reader
    content: string
    position: integer
    line_number: integer
    _eof_emitted: boolean
    
    read: function(self: StringReader, n: integer): (string, string)
    close: function(self: StringReader): (boolean, string)
    get_line_number: function(self: StringReader): number
end

local record StringWriter is bio.Writer
    content: string
    
    write: function(self: StringWriter, data: string): (integer, string)
    close: function(self: StringWriter): (boolean, string)
    get_content: function(self: StringWriter): string
end

local record BufferedReader is Reader
   wrapped: Reader
   buffer: string        
   buffer_pos: integer   
   buffer_size: integer  
   max_line_size: integer

   read: function(self: BufferedReader, n: integer): (string, string)
   close: function(self: BufferedReader): (boolean, string)
   read_line: function(self: BufferedReader): (string, string)
end

--[[ Module fasta
Module fasta contains fasta parsers and writers.

Fasta is a flat text file format developed in 1985 to store nucleotide and
amino acid sequences. It is extremely simple and well-supported across many
languages. However, this simplicity means that annotation of genetic objects
is not supported.

This module provides a parser and writer for working with Fasta formatted
genetic sequences.
]]


local record fasta
   FastaParser: FastaParser
   FastaRecord: FastaRecord
   FastaHeader: FastaHeader
   new_parser: function(reader: bio.Reader, max_line_size: integer): FastaParser
end

local record FastaRecord is bio.Data
   where self.format == "FASTA"

   identifier: string
   sequence: string
   format: bio.Format
   to_string: function(self: FastaRecord): string
   write: function(self: FastaRecord, writer: bio.Writer): (boolean, string)
end

local record FastaHeader is bio.Header
   where self.format == "FASTA"

   format: bio.Format
   to_string: function(self: FastaHeader): string
   write: function(self: FastaHeader, writer: bio.Writer): (boolean, string)
end

local record FastaParser is bio.Parser
   where self.format == "FASTA"

   reader: bio.BufferedReader
   identifier: string
   sequence_buffer: {string}  -- Accumulates sequence data
   start: boolean
   line: integer
   more: boolean

   -- Implement bio.Parser interface
   header: function(self: FastaParser): (bio.Header, string)
   next: function(self: FastaParser): (FastaRecord, string)
   get_format: function(self: FastaParser): bio.Format
end

--[[ Module fastq
Module fastq contains fastq parsers and writers.

Fastq is a flat text file format developed in ~2000 to store nucleotide
sequencing data. While similar to fastq, fastq has a few differences. First,
the sequence identifier begins with @ instead of >, and includes quality
values for a sequence.

This module provides a parser and writer for working with Fastq formatted
sequencing data.
]]


local record fastq
   FastqParser: FastqParser
   FastqRead: FastqRead
   FastqHeader: FastqHeader
   new_parser: function(reader: bio.Reader, max_line_size: integer): FastqParser
end

local record FastqRead is bio.Data
   where self.format == "FASTQ"

   identifier: string
   optionals: {string:string}  -- Nanopore, for example, carries along data like: `read=13956 ch=53 start_time=2020-11-11T01:49:01Z`
   sequence: string
   quality: string
   format: bio.Format
   to_string: function(self: FastqRead): string
   write: function(self: FastqRead, writer: bio.Writer): (boolean, string)
   deep_copy: function(self: FastqRead): FastqRead
end

local record FastqHeader is bio.Header
   where self.format == "FASTQ"

   format: bio.Format
   to_string: function(self: FastqHeader): string
   write: function(self: FastqHeader, writer: bio.Writer): (boolean, string)
end

local record FastqParser is bio.Parser
   where self.format == "FASTQ"

   reader: bio.BufferedReader
   line: integer
   at_eof: boolean

   -- Implement bio.Parser interface
   header: function(self: FastqParser): (bio.Header, string)
   next: function(self: FastqParser): (FastqRead, string)
   get_format: function(self: FastqParser): bio.Format
end

--[[ Module pileup
Module pileup contains pileup parsers and writers.

The pileup format is a text-based bioinformatics format to summarize aligned
reads against a reference sequence. In comparison to simply getting a consensus
sequence from sequencing data, pileup files can contain more context about the
mutations in a sequencing run, which is especially useful when analyzing
plasmid sequencing data from Nanopore sequencing runs.

Pileup files are basically tsv files with 6 columns: Sequence Identifier, Position,
Reference Base, Read Count, Read Results, and Quality. An example from
wikipedia (https://en.wikipedia.org/wiki/Pileup_format) is shown below:

    ```
    seq1    272    T    24    ,.$.....,,.,.,...,,,.,..^+.    <<<+;<<<<<<<<<<<=<;<;7<&
    seq1    273    T    23    ,.....,,.,.,...,,,.,..A    <<<;<<<<<<<<<3<=<<<;<<+
    seq1    274    T    23    ,.$....,,.,.,...,,,.,... 7<7;<;<<<<<<<<<=<;<;<<6
    seq1    275    A    23    ,$....,,.,.,...,,,.,...^l.    <+;9*<<<<<<<<<=<<:;<<<<
    seq1    276    G    22    ...T,,.,.,...,,,.,....    33;+<<7=7<<7<&<<1;<<6<
    seq1    277    T    22    ....,,.,.,.C.,,,.,..G.    +7<;<<<<<<<&<=<<:;<<&<
    seq1    278    G    23    ....,,.,.,...,,,.,....^k.    %38*<<;<7<<7<=<<<;<<<<<
    seq1    279    C    23    A..T,,.,.,...,,,.,.....    75&<<<<<<<<<=<<<9<<:<<<
    ```

    1. Sequence Identifier: The sequence identifier of the reference sequence
    2. Position: Position of row in the reference sequence (indexed at 1)
    3. Reference Base: Base pair in reference sequence
    4. Read Count: Number of aligned reads to this particular base pair
    5. Read Results: The resultant alignments
    6. Quality: Phred quality scores associated with each base

This module provides a parser and writer for working with pileup files.

Wikipedia reference: https://en.wikipedia.org/wiki/Pileup_format
]]


local enum MutationType
   "no_mutation"
   "point"
   "point_indel" 
   "indel"
   "insertion"
   "noisy"
end

local record pileup
   PileupParser: PileupParser
   PileupLine: PileupLine
   PileupHeader: PileupHeader
   new_parser: function(reader: bio.Reader, max_line_size: integer): PileupParser
   
   -- Mutation types
   MutationType: MutationType
   Mutation: Mutation
   call_mutations: function(read_results: {string}, reference_base: string, minimal_ratio: number): Mutation
end

local record PileupLine is bio.Data
   where self.format == "PILEUP"

   sequence: string
   position: integer
   reference_base: string  
   read_count: integer
   read_results: {string}
   quality: string
   format: bio.Format
   to_string: function(self: PileupLine): string
   write: function(self: PileupLine, writer: bio.Writer): (boolean, string)
end

local record PileupHeader is bio.Header
   where self.format == "PILEUP"
   
   format: bio.Format
   to_string: function(self: PileupHeader): string
   write: function(self: PileupHeader, writer: bio.Writer): (boolean, string)
end

local record PileupParser is bio.Parser
   where self.format == "PILEUP"

   reader: bio.BufferedReader
   line: integer
   at_eof: boolean

   -- Implement bio.Parser interface
   header: function(self: PileupParser): (bio.Header, string)
   next: function(self: PileupParser): (PileupLine, string)
   get_format: function(self: PileupParser): bio.Format
end

local record Mutation
   type: MutationType
   from: string
   to: string
   length: integer -- Only for Indel and Insertions
   total_correct: integer
   total_mutated: integer
   total_aligned: integer
end

--[[ Module sam
Module sam implements a SAM file parser and writer.

SAM is a tab-delimited text format for storing DNA/RNA sequence alignment data.
It is the most widely used alignment format, complementing its binary
equivalent, BAM, which stores the same data in a compressed format.

DNA sequencing works in the following way:

  - DNA is read in with some raw signal format from the sequencer machine.
  - Raw signal is converted to fastq reads using basecalling software.
  - Fastq reads are aligned to target template, producing SAM files.
  - SAM files are used to answer bioinformatic queries.

This parser allows parsing and writing of SAM files in Teal. Unlike other SAM
parsers, we aim to be as close to underlying data types as possible,
with a goal of being as simple as possible, and no simpler.

Paper: https://doi.org/10.1093%2Fbioinformatics%2Fbtp352
Spec: http://samtools.github.io/hts-specs/SAMv1.pdf
]]


local enum SortOrder
   "unknown"
   "unsorted"
   "queryname"
   "coordinate"
end

local enum GroupOrder
   "none"
   "query" 
   "reference"
end

local enum TPValue
   "linear"
   "circular"
end

local enum PlatformValue
   "CAPILLARY"
   "DNBSEQ"
   "ELEMENT"
   "HELICOS"
   "ILLUMINA"
   "IONTORRENT"
   "LS454"
   "ONT"
   "PACBIO"
   "SOLID"
   "ULTIMA"
end

local record sam
   SamParser: SamParser
   SamAlignment: SamAlignment 
   SamHeader: SamHeader
   SamOptional: SamOptional
   DEFAULT_MAX_LINE_SIZE: integer
   SortOrder: SortOrder
   GroupOrder: GroupOrder
   TPValue: TPValue
   PlatformValue: PlatformValue
   new_parser: function(reader: bio.Reader, max_line_size: integer): (SamParser, SamHeader, string)
   is_primary: function(alignment: SamAlignment): boolean
end

local record SamHeader is bio.Header
   where self.format == "SAM"

   HD: {string:string}  -- File-level metadata
   SQ: {{string:string}}  -- Reference sequence dictionary
   RG: {{string:string}}  -- Read group
   PG: {{string:string}}  -- Program
   CO: {string}  -- Comments
   format: bio.Format
   to_string: function(self: SamHeader): string
   write: function(self: SamHeader, writer: bio.Writer): (boolean, string)
   validate: function(self: SamHeader): string
end

local record SamOptional
   tag: string  -- Tag is typically a two letter tag
   tag_type: string  -- One of A (character), B (array), f (real), H (hex), i (integer), Z (string)
   data: string  -- Optional data
end

local record SamAlignment is bio.Data
   where self.format == "SAM"

   QNAME: string      -- Query template NAME
   FLAG: integer      -- bitwise FLAG  
   RNAME: string      -- References sequence NAME
   POS: integer       -- 1-based leftmost mapping POSition
   MAPQ: integer      -- MAPping Quality
   CIGAR: string      -- CIGAR string
   RNEXT: string      -- Ref. name of the mate/next read
   PNEXT: integer     -- Position of the mate/next read
   TLEN: integer      -- observed Template LENgth
   SEQ: string        -- segment SEQuence
   QUAL: string       -- ASCII of Phred-scaled base QUALity+33
   optionals: {SamOptional}  -- Optional fields
   format: bio.Format
   to_string: function(self: SamAlignment): string
   write: function(self: SamAlignment, writer: bio.Writer): (boolean, string)
   validate: function(self: SamAlignment): string
end

local record SamParser is bio.Parser
   where self.format == "SAM"

   reader: bio.BufferedReader
   line: integer
   file_header: SamHeader
   first_line: string
   read_first_line: boolean

   -- Implement bio.Parser interface
   header: function(self: SamParser): (bio.Header, string)
   next: function(self: SamParser): (SamAlignment, string)
   get_format: function(self: SamParser): bio.Format
end

--[[ Module slow5
Module slow5 contains slow5 parsers and writers.

Right now, only parsing slow5 files is supported. Support for writing and blow5
coming soon.

slow5 is a file format alternative to fast5, which is the file format outputted
by Oxford Nanopore sequencing devices. fast5 uses hdf5, which is a complex file
format that can only be read and written with a single software library built
in 1998. On the other hand, slow5 uses a .tsv file format, which is easy to
both parse and write.

slow5 files contain both general metadata about the sequencing run and raw
signal reads from the sequencing run. This raw signal can be used directly or
basecalled and used for alignment.

More information on slow5 can be found here: https://github.com/hasindu2008/slow5tools
]]


local record slow5
   Slow5Parser: Slow5Parser
   Slow5HeaderValue: Slow5HeaderValue
   Slow5Header: Slow5Header
   Slow5Read: Slow5Read
   new_parser: function(reader: bio.Reader, max_line_size: integer): (Slow5Parser, string)
end

local record Slow5HeaderValue
   read_group_id: integer  
   slow5_version: string
   attributes: {string:string}
   end_reason_header_map: {string:integer}
end

local record Slow5Header is bio.Header
   where self.format == "SLOW5"

   header_values: {Slow5HeaderValue}
   format: bio.Format
   to_string: function(self: Slow5Header): string
   write: function(self: Slow5Header, writer: bio.Writer): (boolean, string)
end

local record Slow5Read is bio.Data
   where self.format == "SLOW5"

   read_id: string
   read_group_id: integer
   digitisation: number
   offset: number
   read_range: number 
   sampling_rate: number
   len_raw_signal: integer
   raw_signal: {integer}
   
   -- Auxiliary fields
   channel_number: string
   median_before: number
   read_number: integer
   start_mux: integer
   start_time: integer
   end_reason: string
   end_reason_map: {string:integer}
   
   format: bio.Format
   to_string: function(self: Slow5Read): string
   write: function(self: Slow5Read, writer: bio.Writer): (boolean, string)
end

local record Slow5Parser is bio.Parser
   where self.format == "SLOW5"

   reader: bio.BufferedReader
   line: integer
   headerMap: {integer:string}
   endReasonMap: {integer:string}
   endReasonSlow5HeaderMap: {string:integer}
   headerValue: Slow5Header
   hitEOF: boolean

   -- Implement bio.Parser interface
   header: function(self: Slow5Parser): (Slow5Header, string)
   next: function(self: Slow5Parser): (Slow5Read, string)
   get_format: function(self: Slow5Parser): bio.Format
end

--[[ Module genbank
Module genbank provides genbank parsers and writers.

GenBank is a flat text file format developed in the 1980s to annotate genetic
sequences, and has since become the standard for sharing annotated genetic
sequences.

This module provides a parser and writer to convert between the GenBank file
format and the more general Genbank struct.
]]


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

--[[ Module fragment_frequencies

Module fragment_frequencies is simply a table of {string:{string:integer}}
defining the observed experimental ligation rates of different overhangs.

It shouldn't be used directly - just use in combination with the fragment
module.
]]


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

--[[ Module clone
Module clone provides functions for cloning DNA sequences.

Since 1973, the most common way to make recombinant DNA has been restriction
enzyme cloning (though lately, homologous recombination based methods like
Gibson assembly have attracted a lot of use). The cloning functions here allow
for simulation of restriction enzyme cloning.

For a historical review leading up to the discovery:
https://doi.org/10.1073/pnas.1313397110

The idea of restriction enzyme cloning is that you can cut DNA at specific
locations with restriction enzyme and then glue them back together in different
patterns using ligase. The final product is (99.9% of the time) a circular plasmid
that you can transform into a bacterial cell for propagation.

While simulation is simple for simple cases, there are a lot of edge cases to handle, for example:
  - Which input sequences are circular? How do we handle their rotations?
  - Is the enzyme that is cutting directional? How do we handle that directionality?
  - Are there multiple possible outputs of our ligation reaction? For example, ligations may be
    able to create a "library" of plasmids, in which there are millions of valid combinations.
  - How do we handle sequences that get ligated in multiple orientations?

These cloning functions handle all those problems so that they appear simple to the end user.

In particular, there is a focus here on GoldenGate Assembly:
https://en.wikipedia.org/wiki/Golden_Gate_Cloning
https://www.neb.com/applications/cloning-and-synthetic-biology/dna-assembly-and-cloning/golden-gate-assembly

GoldenGate is a particular kind of restriction enzyme cloning reaction that you can do
in a single tube and that is extraordinarily efficient (up to 50 parts) and is popular
for new modular DNA part toolkits. Users can easily simulate GoldenGate assembly reactions
with just their input fragments + the enzyme name.

Unlike many other GoldenGate simulators, we support simulating GoldenGate with
methylated DNA sequences, which are represented as lowercased sequences in user
inputted sequences. Normally, this can be turned off, but can be used in the
special case of recursive GoldenGate reactions.

Let's build some DNA!

PS: We do NOT (yet) handle restriction enzymes which recognize one site but cut
in multiple places (Type IIG enzymes) such as BcgI.
]]


local record clone
   -- Core cloning functions
   -- TODO most of these should have keyword inputs so they are less confusing.
   cut_with_enzyme_by_name: function(part: Part, directional: boolean, name: string, methylated: boolean): ({Fragment}, string)
   cut_with_enzyme: function(part: Part, directional: boolean, enzyme: Enzyme, methylated: boolean): {Fragment}
   ligate: function(fragments: {Fragment}, circular: boolean): (string, {integer}, string)
   golden_gate: function(sequences: {Part}, cutting_enzyme: Enzyme, methylated: boolean): (string, {integer}, string)

   default_enzymes: {string:Enzyme}

   -- Kmer analysis functions for analyzing ligation junctions
   find_kmer_overlaps: function(fragments: {Fragment}, ligation_product: string, ligation_pattern: {integer}, kmer_size: integer): ({KmerOverlap}, string)
   find_kmers: function(kmer_overlaps: {KmerOverlap}, sequence: string): {KmerOverlap}

   -- Get the window surrounding a given fragment. Useful to call on a vector
   -- to call transform.get_window on input sequences to see what sequence
   -- has been inserted into a vector.
   get_window_from_fragment: function(target_fragment: Fragment, left_flank_length: integer, right_flank_length: integer): string, string

   Overhang: Overhang
end

local record Part
   -- Part is a simple struct that can carry a circular or linear DNA sequence.
   -- In the field of synthetic biology, the term "DNA Part" was popularized by
   -- the iGEM competition http://parts.igem.org/Main_Page , so we use that term
   -- here.
   sequence: string
   circular: boolean
end

local record Overhang
   -- Overhang is a struct that represents the ends of a linearized sequence where Enzymes had cut.
   length: integer
   position: integer
   forward: boolean
   recognition_site_plus_skip_length: integer
end

local record Fragment
   -- Fragment is a struct that represents linear DNA sequences with sticky ends.
   sequence: string
   forward_overhang: string
   reverse_overhang: string
end

local record Enzyme
   -- Enzyme is a struct that represents restriction enzymes.
   name: string
   pattern_for: string  -- Lua pattern instead of regexp
   pattern_rev: string  -- Lua pattern instead of regexp
   skip: integer 
   overhead_length: integer
   recognition_site: string
end

local record KmerOverlap
   -- KmerOverlap represents the overlap between two fragments indicative of a
   -- ligation event.
   kmer: string
   fragment1: Fragment
   fragment2: Fragment
end

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

local record dnadesign
    rng: rng
    hash: hash
    transform: transform
    align: align
    mash: mash
    seqhash: seqhash
    primers: primers
    orthoprimers: orthoprimers
    pcr: pcr
    bio: bio
    fasta: fasta
    fastq: fastq
    pileup: pileup
    sam: sam
    slow5: slow5
    genbank: genbank
    fragment: fragment
    clone: clone
    codon: codon
    fix: fix
end

return dnadesign
