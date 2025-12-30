-- examples/sam_examples.lua
local dnadesign = require("dnadesign")
local bio = dnadesign.bio
local sam = dnadesign.sam

describe("SAM Examples", function()
    it("demonstrates basic SAM parsing", function()
        local sam_string = [[
@HD	VN:1.6	SO:unsorted
@SQ	SN:ref1	LN:1000
test_read	0	ref1	1	60	10M	*	0	0	ACGTACGTAC	!!!!!!!!!!
]]
        local reader = bio.new_string_reader(sam_string)
        local parser, header, err = sam.new_parser(reader, sam.DEFAULT_MAX_LINE_SIZE)
        assert.is_nil(err)
        
        -- Check header information
        assert.are.equal("1.6", header.HD.VN)
        assert.are.equal("unsorted", header.HD.SO)
        assert.are.equal("ref1", header.SQ[1].SN)
        assert.are.equal("1000", header.SQ[1].LN)
        
        -- Read alignment
        local alignment, read_err = parser:next()
        assert.is_nil(read_err)
        assert.are.equal("test_read", alignment.QNAME)
        assert.are.equal(0, alignment.FLAG)
        assert.are.equal("10M", alignment.CIGAR)
        assert.are.equal("ACGTACGTAC", alignment.SEQ)
    end)

    it("demonstrates SAM writing and reading", function()
        -- First create a SAM record with header
        local sam_string = [[
@HD	VN:1.6	SO:coordinate
@SQ	SN:chr1	LN:1000
read1	0	chr1	1	30	8M2I	*	0	0	ACGTACGT	!!!!!!!!
]]
        local reader = bio.new_string_reader(sam_string)
        local parser, header, err = sam.new_parser(reader, sam.DEFAULT_MAX_LINE_SIZE)
        assert.is_nil(err)
        
        -- Write header to a string writer
        local writer = bio.new_string_writer()
        local success, write_err = header:write(writer)
        assert.is_true(success)
        assert.is_nil(write_err)
        
        -- Get and write alignment
        local alignment, read_err = parser:next()
        assert.is_nil(read_err)
        success, write_err = alignment:write(writer)
        assert.is_true(success)
        assert.is_nil(write_err)
        
        -- Read it back
        local new_reader = bio.new_string_reader(writer:get_content())
        local new_parser, new_header, new_err = sam.new_parser(new_reader, sam.DEFAULT_MAX_LINE_SIZE)
        assert.is_nil(new_err)
        assert.are.equal("1.6", new_header.HD.VN)
        
        local new_alignment, new_read_err = new_parser:next()
        assert.is_nil(new_read_err)
        assert.are.equal("read1", new_alignment.QNAME)
    end)

    it("demonstrates working with optional fields", function()
        local sam_string = [[
@HD	VN:1.6	SO:unsorted
read1	0	*	0	0	*	*	0	0	*	*	NM:i:2	MD:Z:10A5C	AS:i:100
]]
        local reader = bio.new_string_reader(sam_string)
        local parser, _, err = sam.new_parser(reader, sam.DEFAULT_MAX_LINE_SIZE)
        assert.is_nil(err)
        
        local alignment, read_err = parser:next()
        assert.is_nil(read_err)
        
        -- Check optional fields
        assert.are.equal(3, #alignment.optionals)
        assert.are.equal("NM", alignment.optionals[1].tag)
        assert.are.equal("i", alignment.optionals[1].tag_type)
        assert.are.equal("2", alignment.optionals[1].data)
        
        assert.are.equal("MD", alignment.optionals[2].tag)
        assert.are.equal("Z", alignment.optionals[2].tag_type)
        assert.are.equal("10A5C", alignment.optionals[2].data)
    end)

    it("demonstrates checking primary alignments", function()
        local sam_string = [[
@HD	VN:1.6	SO:unsorted
primary	0	*	0	0	*	*	0	0	*	*
secondary	256	*	0	0	*	*	0	0	*	*
supplementary	2048	*	0	0	*	*	0	0	*	*
]]
        local reader = bio.new_string_reader(sam_string)
        local parser, _, err = sam.new_parser(reader, sam.DEFAULT_MAX_LINE_SIZE)
        assert.is_nil(err)
        
        -- Check primary alignment
        local aln1, err1 = parser:next()
        assert.is_nil(err1)
        assert.is_true(sam.is_primary(aln1))
        
        -- Check secondary alignment
        local aln2, err2 = parser:next()
        assert.is_nil(err2)
        assert.is_false(sam.is_primary(aln2))
        
        -- Check supplementary alignment
        local aln3, err3 = parser:next()
        assert.is_nil(err3)
        assert.is_false(sam.is_primary(aln3))
    end)

    it("demonstrates header validation", function()
        local sam_string = [[
@HD	VN:1.6	SO:coordinate
@SQ	SN:chr1	LN:1000	TP:linear
@RG	ID:group1	PL:ILLUMINA	SM:sample1
]]
        local reader = bio.new_string_reader(sam_string)
        local parser, header, err = sam.new_parser(reader, sam.DEFAULT_MAX_LINE_SIZE)
        assert.is_nil(err)
        
        -- Validate header
        local validate_err = header:validate()
        assert.is_nil(validate_err)
        
        -- Check header components
        assert.are.equal("1.6", header.HD.VN)
        assert.are.equal("coordinate", header.HD.SO)
        assert.are.equal("chr1", header.SQ[1].SN)
        assert.are.equal("linear", header.SQ[1].TP)
        assert.are.equal("ILLUMINA", header.RG[1].PL)
    end)
end)
