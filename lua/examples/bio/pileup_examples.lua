-- examples/pileup_examples.lua
local dnadesign = require("dnadesign")
local bio = dnadesign.bio
local pileup = dnadesign.pileup

describe("PILEUP Examples", function()
    it("demonstrates basic PILEUP parsing", function()
        local pileup_string = "seq1\t272\tT\t24\t,.$.....,,.,.,...,,,.,..^+.\t<<<+;<<<<<<<<<<<=<;<;7<&\n"
        local reader = bio.new_string_reader(pileup_string)
        local parser = pileup.new_parser(reader, 32*1024)
        
        local record, err = parser:next()
        assert.is_nil(err)
        assert.are.equal("seq1", record.sequence)
        assert.are.equal(272, record.position)
        assert.are.equal("T", record.reference_base)
        assert.are.equal(24, record.read_count)
        assert.are.equal("<<<+;<<<<<<<<<<<=<;<;7<&", record.quality)
    end)

    it("demonstrates parsing PILEUP with insertions and deletions", function()
        local complex_pileup = "seq1\t100\tA\t4\t,.+2AT,-3GTC.\t<<<<<"
        local reader = bio.new_string_reader(complex_pileup)
        local parser = pileup.new_parser(reader, 32*1024)
        
        local record, err = parser:next()
        assert.is_nil(err)
        
        -- Look for the insertion and deletion in read_results
        local found_insertion = false
        local found_deletion = false
        for _, result in ipairs(record.read_results) do
            if result == "+2AT" then found_insertion = true end
            if result == "-3GTC" then found_deletion = true end
        end
        
        assert.is_true(found_insertion)
        assert.is_true(found_deletion)
    end)

    it("demonstrates mutation detection", function()
        -- Example showing how to detect different types of mutations
        local examples = {
            {
                reads = {".", ",", "A", "A", "A"},  -- Point mutation to A
                ref = "T",
                description = "point mutation"
            },
            {
                reads = {".", ",", "-2AT", "-2AT", "-2AT"},  -- Deletion
                ref = "T",
                description = "deletion"
            },
            {
                reads = {".", "A", "T", "G", "C"},  -- Noisy reads
                ref = "T",
                description = "noisy reads"
            }
        }
        
        for _, example in ipairs(examples) do
            local mutation = pileup.call_mutations(example.reads, example.ref, 0.5)
            if example.description == "point mutation" then
                assert.are.equal("point", mutation.type)
                assert.are.equal("A", mutation.to)
            elseif example.description == "deletion" then
                assert.are.equal("indel", mutation.type)
                assert.are.equal(2, mutation.length)
            elseif example.description == "noisy reads" then
                assert.are.equal("noisy", mutation.type)
                assert.are.equal("?", mutation.to)
            end
        end
    end)
end)
