-- examples/fasta_examples.lua
local dnadesign = require("dnadesign")
local bio = dnadesign.bio
local fasta = dnadesign.fasta

describe("FASTA Examples", function()
    it("demonstrates basic FASTA parsing", function()
        local fasta_string = ">testing\nATGC\n"
        local reader = bio.new_string_reader(fasta_string)
        local parser = fasta.new_parser(reader, 32*1024)
        
        local records = {}
        while true do
            local record, err = parser:next()
            if err then
                if err == "EOF" then
                    break
                end
                error("Got error: " .. err)
            end
            table.insert(records, record)
        end
        
        assert.are.equal("ATGC", records[1].sequence)
    end)

	it("demonstrates FASTA writing and reading", function()
        -- Create a FASTA record
        local fasta_string = ">testing\nATGC\n"
        local reader = bio.new_string_reader(fasta_string)
        local parser = fasta.new_parser(reader, 32*1024)
        local record, err = parser:next()
        assert.is_nil(err)

        -- Write it to a string writer
        local writer = bio.new_string_writer()
        local success, write_err = record:write(writer)
        assert.is_true(success)
        assert.is_nil(write_err)
        
        -- Read it back from the written content
        local new_reader = bio.new_string_reader(writer:get_content())
        local new_parser = fasta.new_parser(new_reader, 32*1024)
        local new_record, read_err = new_parser:next()
        assert.is_nil(read_err)

        assert.are.equal("testing", new_record.identifier)
	end)

    it("demonstrates parsing multiple FASTA sequences", function()
        local multi_fasta = ">seq1\nGATTACA\n>seq2\nCGCGCGC\n"
        local reader = bio.new_string_reader(multi_fasta)
        local parser = fasta.new_parser(reader, 32*1024)
        
        local record1, err1 = parser:next()
        assert.is_nil(err1)
        assert.are.equal("seq1", record1.identifier)
        assert.are.equal("GATTACA", record1.sequence)
        
        local record2, err2 = parser:next()
        assert.is_nil(err2)
        assert.are.equal("seq2", record2.identifier)
        assert.are.equal("CGCGCGC", record2.sequence)
    end)
end)
