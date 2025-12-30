-- examples/fastq_examples.lua
local dnadesign = require("dnadesign")
local bio = dnadesign.bio
local fastq = dnadesign.fastq

describe("FASTQ Examples", function()
    it("demonstrates basic FASTQ parsing with optionals", function()
        local fastq_string = "@e3cc70d5-90ef-49b6-bbe1-cfef99537d73 runid=test123 ch=53 start_time=2020-11-11T01:49:01Z\n" ..
                            "GATTACA\n" ..
                            "+\n" ..
                            "IIIIIII\n"
        local reader = bio.new_string_reader(fastq_string)
        local parser = fastq.new_parser(reader, 32*1024)
        
        local record, err = parser:next()
        assert.is_nil(err)
        assert.are.equal("e3cc70d5-90ef-49b6-bbe1-cfef99537d73", record.identifier)
        assert.are.equal("GATTACA", record.sequence)
        assert.are.equal("IIIIIII", record.quality)
        assert.are.equal("53", record.optionals["ch"])
        assert.are.equal("test123", record.optionals["runid"])
    end)

    it("demonstrates parsing multiple FASTQ sequences", function()
        local multi_fastq = "@read1\nATCG\n+\nIIII\n" ..
                           "@read2\nGCTA\n+\nAAAA\n"
        local reader = bio.new_string_reader(multi_fastq)
        local parser = fastq.new_parser(reader, 32*1024)
        
        local identifiers = {}
        while true do
            local record, err = parser:next()
            if err then
                if err == "EOF" then
                    break
                end
                error("Got error: " .. err)
            end
            table.insert(identifiers, record.identifier)
        end
        
        assert.are.same({"read1", "read2"}, identifiers)
    end)

    it("demonstrates deep copying FASTQ records", function()
        local fastq_string = "@read1 sample=test1 length=4\nATCG\n+\nIIII\n"
        local reader = bio.new_string_reader(fastq_string)
        local parser = fastq.new_parser(reader, 32*1024)
        
        local original, err = parser:next()
        assert.is_nil(err)
        
        local copy = original:deep_copy()
        
        -- Modify original
        original.optionals["new_field"] = "value"
        
        -- Verify copy remains unchanged
        assert.are.equal(original.sequence, copy.sequence)
        assert.are.equal(original.quality, copy.quality)
        assert.are.equal("test1", copy.optionals["sample"])
        assert.is_nil(copy.optionals["new_field"])
    end)
end)
