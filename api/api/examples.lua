-- NOTE: we use camelCase vs snake_case because sometimes AI systems get
-- confused about the underscores and escape them - breaking JSON marshallers.

--[[
*****************************************************************************

IO examples.

*****************************************************************************
--]]

-- START: fastaParse
-- QUESTION: Does the request require the parsing of FASTA formatted data? Return a boolean.
-- fastaParse parses a fasta file into a list of tables of "identifier" and
-- "sequence"
parsedFasta = fastaParse(">test\nATGC\n>test2\nGATC")
print(parsedFasta[1]["identifier"]) -- returns "test"
print(parsedFasta[2]["sequence"]) -- returns "GATC"
-- END

--[[
*****************************************************************************

CDS examples.

*****************************************************************************
--]]

--[[
*****************************************************************************

PCR examples.

*****************************************************************************
--]]

--[[
*****************************************************************************

IO

*****************************************************************************
--]]
