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

-- Define the substitution matrix record type
local record SubstitutionMatrix
   data: {string:{string:integer}}
end

-- Scoring configuration record type
local record Scoring 
   substitution_matrix: SubstitutionMatrix
   gap_penalty: integer
end

-- Default DNA scoring matrix implementation
local default_matrix: SubstitutionMatrix = {
   data = {
      A = {A = 1, C = -1, G = -1, T = -1, U = -1, ["-"] = -1},
      C = {A = -1, C = 1, G = -1, T = -1, U = -1, ["-"] = -1},
      G = {A = -1, C = -1, G = 1, T = -1, U = -1, ["-"] = -1},
      T = {A = -1, C = -1, G = -1, T = 1, U = -1, ["-"] = -1},
      U = {A = -1, C = -1, G = -1, T = -1, U = 1, ["-"] = -1},
      ["-"] = {A = -1, C = -1, G = -1, T = -1, U = -1, ["-"] = -1}
   }
}
align.default_matrix = default_matrix

-- NewScoring returns a new Scoring struct with default values for DNA.
function align.new_scoring(substitution_matrix: SubstitutionMatrix | nil, gap_penalty: integer): Scoring
   if not substitution_matrix then
      substitution_matrix = default_matrix
   end
   return {
      substitution_matrix = substitution_matrix,
      gap_penalty = gap_penalty
   }
end

-- Score returns the match/mismatch score for two characters
local function score(scoring: Scoring, a: string, b: string): integer
   if not scoring.substitution_matrix.data[a] or not scoring.substitution_matrix.data[a][b] then
      error("Invalid characters for scoring: " .. a .. ", " .. b)
   end
   return scoring.substitution_matrix.data[a][b]
end

-- Helper function to get maximum of two values
local function max(a: integer, b: integer): integer
   if a > b then
      return a
   end
   return b
end

--[[
For developers:

Needleman–Wunsch traceback stops too early and truncates alignments when sequences differ in length.

So the algorithm computes the correct score, but returns a truncated alignment for unequal lengths.

This should be fixed in later versions.
--]]

-- NeedlemanWunsch performs global alignment between two strings using the Needleman-Wunsch algorithm.
-- It returns the final score and the optimal alignments of the two strings in O(nm) time and O(nm) space.
-- https://en.wikipedia.org/wiki/Needleman-Wunsch_algorithm
function align.needleman_wunsch(string_a: string, string_b: string, scoring: Scoring): (integer, string, string)
   -- Get the M and N dimensions of the matrix. The M x N matrix is standard linear algebra notation.
   -- But I added columns and rows to the variable name to make it more clear what the dimensions are.
   local column_length_m, row_length_n = #string_a, #string_b

   -- Initialize the matrix
   local matrix: {{integer}} = {}
   for i = 0, column_length_m do
      matrix[i] = {}
      for j = 0, row_length_n do
         matrix[i][j] = 0
      end
   end

   -- Fill in the first column with gap penalties
   for i = 1, column_length_m do
      matrix[i][0] = matrix[i-1][0] + scoring.gap_penalty
   end

   -- Fill in the first row with gap penalties
   for j = 1, row_length_n do
      matrix[0][j] = matrix[0][j-1] + scoring.gap_penalty
   end

   -- Fill in the rest of the matrix
   for i = 1, column_length_m do
      for j = 1, row_length_n do
         local match_score = score(scoring, string_a:sub(i,i), string_b:sub(j,j))
         matrix[i][j] = max(
            matrix[i-1][j-1] + match_score,
            max(
               matrix[i-1][j] + scoring.gap_penalty,
               matrix[i][j-1] + scoring.gap_penalty
            )
         )
      end
   end

   -- Traceback to find the optimal alignment
   local align_a = ""
   local align_b = ""
   local i, j = column_length_m, row_length_n

   while i > 0 and j > 0 do
      local match_score = score(scoring, string_a:sub(i,i), string_b:sub(j,j))
      if matrix[i][j] == matrix[i-1][j-1] + match_score then
         align_a = string_a:sub(i,i) .. align_a
         align_b = string_b:sub(j,j) .. align_b
         i = i - 1
         j = j - 1
      elseif matrix[i][j] == matrix[i-1][j] + scoring.gap_penalty then
         align_a = string_a:sub(i,i) .. align_a
         align_b = "-" .. align_b
         i = i - 1
      else
         align_a = "-" .. align_a
         align_b = string_b:sub(j,j) .. align_b
         j = j - 1
      end
   end

   return matrix[column_length_m][row_length_n], align_a, align_b
end

-- SmithWaterman performs local alignment between two strings using the Smith-Waterman algorithm.
-- It returns the max score and optimal local alignments between two strings alignments of the two strings in O(nm) time and O(nm) space.
-- https://en.wikipedia.org/wiki/Smith-Waterman_algorithm
function align.smith_waterman(string_a: string, string_b: string, scoring: Scoring): (integer, string, string)
   local column_length_m, row_length_n = #string_a, #string_b

   -- Initialize the alignment matrix
   local matrix: {{integer}} = {}
   for i = 0, column_length_m do
      matrix[i] = {}
      for j = 0, row_length_n do
         matrix[i][j] = 0
      end
   end

   -- Initialize variables to keep track of the maximum score and its position
   local max_score = 0
   local max_score_row = 0
   local max_score_col = 0

   -- Fill the alignment matrix
   for i = 1, column_length_m do
      for j = 1, row_length_n do
         local match_score = score(scoring, string_a:sub(i,i), string_b:sub(j,j))
         local diag_score = matrix[i-1][j-1] + match_score
         local up_score = matrix[i-1][j] + scoring.gap_penalty
         local left_score = matrix[i][j-1] + scoring.gap_penalty
         matrix[i][j] = max(0, max(diag_score, max(up_score, left_score)))

         if matrix[i][j] > max_score then
            max_score = matrix[i][j]
            max_score_row = i
            max_score_col = j
         end
      end
   end

   -- Traceback to construct the aligned strings
   local align_a = ""
   local align_b = ""
   local i, j = max_score_row, max_score_col

   while matrix[i][j] > 0 do
      local match_score = score(scoring, string_a:sub(i,i), string_b:sub(j,j))
      if matrix[i][j] == matrix[i-1][j-1] + match_score then
         align_a = string_a:sub(i,i) .. align_a
         align_b = string_b:sub(j,j) .. align_b
         i = i - 1
         j = j - 1
      elseif matrix[i][j] == matrix[i-1][j] + scoring.gap_penalty then
         align_a = string_a:sub(i,i) .. align_a
         align_b = "-" .. align_b
         i = i - 1
      else
         align_a = "-" .. align_a
         align_b = string_b:sub(j,j) .. align_b
         j = j - 1
      end
   end

   return max_score, align_a, align_b
end

--[[
align_many is a helper function for figuring out, given an insert, which
candidate matches the very best. It is much slower than methods like minimap2,
but it is also more accurate: it is deterministic, does not use heuristics, and
will give the actual optimal alignment.

More advanced algorithms would be needed for large sequences, but given
sufficient computation, smith-waterman or needleman-wunsh will give optimal
results.

TODO: have {integer, string, string} be a type
--]]
function align.align_many(func: function(string_a: string, string_b: string, scoring: Scoring): (integer, string, string), target: string, candidates: {string}, scoring: Scoring, ntop: integer): { {integer, string, string, integer} }
    local results: { {integer, string, string, integer} } = {}
    for i = 1, #candidates do
        local alignment_score, align_a, align_b = func(target, candidates[i], scoring)
        results[#results + 1] = { alignment_score, align_a, align_b, i }  -- Add index here
    end

    -- sort by score descending
    table.sort(results, function(a:{integer, string, string, integer}, b: {integer, string, string, integer}): boolean
        return a[1] > b[1]
    end)

    -- clamp ntop to the number of results
    if ntop < 0 then
        ntop = 0
    end
    if ntop > #results then
        ntop = #results
    end

    -- return only the top ntop entries
    local out: { {integer, string, string, integer} } = {}
    for i = 1, ntop do
        out[i] = results[i]
    end

    return out
end

return align
