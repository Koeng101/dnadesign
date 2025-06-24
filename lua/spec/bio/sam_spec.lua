local dnadesign = require("dnadesign")
local bio = dnadesign.bio
local sam = dnadesign.sam

describe("SAM Parser", function()
   describe("basic parsing", function()
      it("should parse valid SAM data with header", function()
         local content = [[
@HD	VN:1.6	SO:unsorted	GO:query
@SQ	SN:pOpen_V3_amplified	LN:2482
@PG	ID:minimap2	PN:minimap2	VN:2.24-r1155-dirty	CL:minimap2 -acLx map-ont
test_read	16	pOpen_V3_amplified	1	60	8S54M1D3M1D108M1D1M1D62M226S	*	0	0	AGCATGCCGCTTTTCTGTGACTGGTGAGTACTCAACCAAGT	!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
]]
         local reader = bio.new_string_reader(content)
         local parser, header, err = sam.new_parser(reader, sam.DEFAULT_MAX_LINE_SIZE)
         
         assert.is_nil(err)
         assert.equals("1.6", header.HD.VN)
         assert.equals("unsorted", header.HD.SO)
         assert.equals("query", header.HD.GO)
         
         local alignment, read_err = parser:next()
         assert.is_nil(read_err)
         assert.equals("test_read", alignment.QNAME)
         assert.equals(16, alignment.FLAG)
         assert.equals("pOpen_V3_amplified", alignment.RNAME)
         assert.equals(1, alignment.POS)
         assert.equals(60, alignment.MAPQ)
         assert.equals("8S54M1D3M1D108M1D1M1D62M226S", alignment.CIGAR)
      end)
   end)

   describe("header validation", function()
      it("should validate header VN format", function()
         local content = [[
@HD	VN:invalid	SO:unsorted	GO:query
@SQ	SN:test	LN:1000
]]
         local reader = bio.new_string_reader(content)
         local parser, header, err = sam.new_parser(reader, sam.DEFAULT_MAX_LINE_SIZE)
         
         assert.is_nil(err)
         local validate_err = header:validate()
         assert.matches("Invalid format for @HD VN", validate_err)
      end)

      it("should validate header SO values", function()
         local content = [[
@HD	VN:1.6	SO:invalid	GO:query
@SQ	SN:test	LN:1000
]]
         local reader = bio.new_string_reader(content)
         local parser, header, err = sam.new_parser(reader, sam.DEFAULT_MAX_LINE_SIZE)
         
         assert.is_nil(err)
         local validate_err = header:validate()
         assert.matches("Invalid value for @HD SO", validate_err)
      end)

      it("should validate unique SQ SN values", function()
         local content = [[
@HD	VN:1.6	SO:unsorted	GO:query
@SQ	SN:test	LN:1000
@SQ	SN:test	LN:2000
]]
         local reader = bio.new_string_reader(content)
         local parser, header, err = sam.new_parser(reader, sam.DEFAULT_MAX_LINE_SIZE)
         
         assert.is_nil(err)
         local validate_err = header:validate()
         assert.matches("Non%-unique @SQ SN", validate_err)
      end)
   end)

   describe("alignment validation", function()
      it("should validate QNAME format", function()
		   local content = [[
@HD	VN:1.6	SO:unsorted
test read with spaces	0	*	0	0	8S54M1D3M1D108M1D1M1D62M226S	*	0	0	*	*
]]
		   local reader = bio.new_string_reader(content)
		   local parser, _, err = sam.new_parser(reader, sam.DEFAULT_MAX_LINE_SIZE)
		   
		   assert.is_nil(err)
		   local alignment, read_err = parser:next()
		   assert.is_nil(read_err)
		   
		   local validate_err = alignment:validate()
		   assert.matches("Invalid QNAME format", validate_err)
		end)

      it("should validate FLAG range", function()
         local content = [[
@HD	VN:1.6	SO:unsorted
test	70000	*	0	0	*	*	0	0	*	*
]]
         local reader = bio.new_string_reader(content)
         local parser, _, err = sam.new_parser(reader, sam.DEFAULT_MAX_LINE_SIZE)
         
         assert.is_nil(err)
         local alignment, read_err = parser:next()
         assert.is_nil(read_err)
         
         local validate_err = alignment:validate()
         assert.matches("Invalid FLAG range", validate_err)
      end)

      it("should validate CIGAR format", function()
         local content = [[
@HD	VN:1.6	SO:unsorted
test	0	*	0	0	1X2Y	*	0	0	*	*
]]
         local reader = bio.new_string_reader(content)
         local parser, _, err = sam.new_parser(reader, sam.DEFAULT_MAX_LINE_SIZE)
         
         assert.is_nil(err)
         local alignment, read_err = parser:next()
         assert.is_nil(read_err)
         
         local validate_err = alignment:validate()
         assert.matches("Invalid CIGAR format", validate_err)
      end)
   end)

   describe("optional fields", function()
      it("should parse optional fields correctly", function()
         local content = [[
@HD	VN:1.6	SO:unsorted
test	0	*	0	0	*	*	0	0	*	*	NM:i:4	MD:Z:ACGT	AS:i:100
]]
         local reader = bio.new_string_reader(content)
         local parser, _, err = sam.new_parser(reader, sam.DEFAULT_MAX_LINE_SIZE)
         
         assert.is_nil(err)
         local alignment, read_err = parser:next()
         assert.is_nil(read_err)
         
         assert.equals(3, #alignment.optionals)
         assert.equals("NM", alignment.optionals[1].tag)
         assert.equals("i", alignment.optionals[1].tag_type)
         assert.equals("4", alignment.optionals[1].data)
      end)
   end)

   describe("primary alignment check", function()
      it("should identify primary alignments", function()
         local content = [[
@HD	VN:1.6	SO:unsorted
test1	0	*	0	0	*	*	0	0	*	*
test2	256	*	0	0	*	*	0	0	*	*
test3	2048	*	0	0	*	*	0	0	*	*
]]
         local reader = bio.new_string_reader(content)
         local parser, _, err = sam.new_parser(reader, sam.DEFAULT_MAX_LINE_SIZE)
         
         assert.is_nil(err)
         
         -- First alignment should be primary
         local aln1, err1 = parser:next()
         assert.is_nil(err1)
         assert.is_true(sam.is_primary(aln1))
         
         -- Second alignment (FLAG=256) should be secondary
         local aln2, err2 = parser:next()
         assert.is_nil(err2)
         assert.is_false(sam.is_primary(aln2))
         
         -- Third alignment (FLAG=2048) should be supplementary
         local aln3, err3 = parser:next()
         assert.is_nil(err3)
         assert.is_false(sam.is_primary(aln3))
      end)
   end)

   describe("error handling", function()
      it("should handle missing required fields", function()
         local content = [[
@HD	VN:1.6	SO:unsorted
test	0	*	0	0	*	*	0	0	*
]]
         local reader = bio.new_string_reader(content)
         local parser, _, err = sam.new_parser(reader, sam.DEFAULT_MAX_LINE_SIZE)
         
         assert.is_nil(err)
         local _, read_err = parser:next()
         assert.matches("must have at least 11 tab%-delimited values", read_err)
      end)

      it("should handle invalid number formats", function()
         local content = [[
@HD	VN:1.6	SO:unsorted
test	abc	*	0	0	*	*	0	0	*	*
]]
         local reader = bio.new_string_reader(content)
         local parser, _, err = sam.new_parser(reader, sam.DEFAULT_MAX_LINE_SIZE)
         
         assert.is_nil(err)
         local _, read_err = parser:next()
         assert.matches("contains invalid number format", read_err)
      end)
   end)
end)
