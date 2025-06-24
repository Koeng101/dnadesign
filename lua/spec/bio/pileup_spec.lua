local dnadesign = require("dnadesign")
local bio = dnadesign.bio
local pileup = dnadesign.pileup

describe("PILEUP Parser", function()
   describe("basic parsing", function()
      it("should parse valid pileup data", function()
         -- Valid pileup data with read count 1401 at line 3
         local content = [[
seq1	272	T	24	,.$.....,,.,.,...,,,.,..^+.	<<<+;<<<<<<<<<<<=<;<;7<&
seq1	273	T	23	,.....,,.,.,...,,,.,..A	<<<;<<<<<<<<<3<=<<<;<<+
seq1	274	T	1401	,.$....,,.,.,...,,,.,... 	7<7;<;<<<<<<<<<=<;<;<<6
]]
         local reader = bio.new_string_reader(content)
         local parser = pileup.new_parser(reader, 2 * 32 * 1024)
         
         local pileup_reads = {}
         while true do
            local read, err = parser:next()
            if err == "EOF" then
               break
            end
            assert.is_nil(err)
            table.insert(pileup_reads, read)
         end
         
         assert.equals(1401, pileup_reads[3].read_count)
      end)
   end)

   describe("error handling", function()
      it("should detect not enough fields", function()
         local content = "seq1	272	T	24	,.$.....,,.,.,...,,,.,..^+."  -- Missing quality field
         local reader = bio.new_string_reader(content)
         local parser = pileup.new_parser(reader, 2 * 32 * 1024)
         
         local _, err = parser:next()
         assert.is_not_nil(err)
         assert.matches("values, expected 6", err)
      end)

      it("should detect non-integer position", function()
         local content = "seq1	abc	T	24	,.$.....,,.,.,...,,,.,..^+.	<<<+;<<<<<<<<<<<=<;<;7<&"
         local reader = bio.new_string_reader(content)
         local parser = pileup.new_parser(reader, 2 * 32 * 1024)
         
         local _, err = parser:next()
         assert.is_not_nil(err)
         assert.matches("Error on line", err)
      end)

      it("should detect non-integer read count", function()
         local content = "seq1	272	T	xyz	,.$.....,,.,.,...,,,.,..^+.	<<<+;<<<<<<<<<<<=<;<;7<&"
         local reader = bio.new_string_reader(content)
         local parser = pileup.new_parser(reader, 2 * 32 * 1024)
         
         local _, err = parser:next()
         assert.is_not_nil(err)
         assert.matches("Error on line", err)
      end)

      it("should detect invalid indel format", function()
         local content = "seq1	272	T	24	,.$.....,,.,.,...,,,.,..+X.	<<<+;<<<<<<<<<<<=<;<;7<&"  -- Invalid character in indel
         local reader = bio.new_string_reader(content)
         local parser = pileup.new_parser(reader, 2 * 32 * 1024)
         
         local _, err = parser:next()
         assert.is_not_nil(err)
         assert.matches("Error on line 1: Invalid indel format", err)
      end)

      it("should detect unknown characters", function()
         local content = "seq1	272	T	24	,.$.....,,.,.,...,,,.,.?^+.	<<<+;<<<<<<<<<<<=<;<;7<&"  -- Unknown character ?
         local reader = bio.new_string_reader(content)
         local parser = pileup.new_parser(reader, 2 * 32 * 1024)
         
         local _, err = parser:next()
         assert.is_not_nil(err)
         assert.matches("Invalid character in read results", err)
      end)
   end)

   describe("special cases", function()
      it("should handle insertions correctly", function()
         local content = "seq1	272	T	24	,.+3ATG..,,.,.,...,,,.,..^+.	<<<+;<<<<<<<<<<<=<;<;7<&"
         local reader = bio.new_string_reader(content)
         local parser = pileup.new_parser(reader, 2 * 32 * 1024)
         
         local read, err = parser:next()
         assert.is_nil(err)
         assert.is_true(#read.read_results > 0)
         -- Check that the insertion "+3ATG" is captured correctly
         local found_insertion = false
         for _, result in ipairs(read.read_results) do
            if result == "+3ATG" then
               found_insertion = true
               break
            end
         end
         assert.is_true(found_insertion)
      end)

      it("should handle deletions correctly", function()
         local content = "seq1	272	T	24	,.-3ATG..,,.,.,...,,,.,..^+.	<<<+;<<<<<<<<<<<=<;<;7<&"
         local reader = bio.new_string_reader(content)
         local parser = pileup.new_parser(reader, 2 * 32 * 1024)
         
         local read, err = parser:next()
         assert.is_nil(err)
         assert.is_true(#read.read_results > 0)
         -- Check that the deletion "-3ATG" is captured correctly
         local found_deletion = false
         for _, result in ipairs(read.read_results) do
            if result == "-3ATG" then
               found_deletion = true
               break
            end
         end
         assert.is_true(found_deletion)
      end)

      it("should handle start/end markers correctly", function()
         local content = "seq1	272	T	24	,.$....^k.,,.,.,...,,,.,..^+.	<<<+;<<<<<<<<<<<=<;<;7<&"
         local reader = bio.new_string_reader(content)
         local parser = pileup.new_parser(reader, 2 * 32 * 1024)
         
         local read, err = parser:next()
         assert.is_nil(err)
         -- Verify that both $ and ^ markers are handled
         local found_start = false
         local found_end = false
         for _, result in ipairs(read.read_results) do
            if result:find("^") then
               found_start = true
            end
            if result:find("$") then
               found_end = true
            end
         end
         assert.is_true(found_start)
         assert.is_true(found_end)
      end)
   end)

   describe("mutation detection", function()
      it("should detect point mutations", function()
         local read_results = {".", ",", "A", "A", "A"}  -- 3/5 are A mutations
         local mutation = pileup.call_mutations(read_results, "T", 0.5)
         assert.equals("point", mutation.type)
         assert.equals("A", mutation.to)
         assert.equals(2, mutation.total_correct)
         assert.equals(3, mutation.total_mutated)
      end)

      it("should detect indels", function()
         local read_results = {".", ",", "-2AT", "-2AT", "-2AT"}  -- 3/5 are deletions
         local mutation = pileup.call_mutations(read_results, "T", 0.5)
         assert.equals("indel", mutation.type)
         assert.equals(2, mutation.length)
         assert.equals(3, mutation.total_mutated)
      end)

      it("should handle noisy reads", function()
         local read_results = {".", "A", "T", "G", "C"}  -- Very noisy, no clear consensus
         local mutation = pileup.call_mutations(read_results, "T", 0.2)
         assert.equals("noisy", mutation.type)
         assert.equals("?", mutation.to)
      end)

      it("should detect no mutation when changes are below threshold", function()
         local read_results = {".", ",", ",", "A"}  -- Only 1/4 is mutated
         local mutation = pileup.call_mutations(read_results, "T", 0.5)
         assert.equals("no_mutation", mutation.type)
      end)
   end)
end)
