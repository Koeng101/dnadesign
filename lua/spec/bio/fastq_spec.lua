local dnadesign = require("dnadesign")
local bio = dnadesign.bio
local fastq = dnadesign.fastq

describe("FASTQ Parser", function()
   describe("header", function()
      it("should return empty header without error", function()
         local reader = bio.new_string_reader("")
         local parser = fastq.new_parser(reader, 256)
         local header, err = parser:header()
         assert.is_nil(err)
         assert.equals("FASTQ", header.format)
         assert.equals("", header:to_string())
      end)
   end)

   describe("parser", function()
      it("should parse EOF-ended FASTQ with optionals", function()
         local content = "@e3cc70d5-90ef-49b6-bbe1-cfef99537d73 runid=99790f25859e24307203c25273f3a8be8283e7eb read=13956 ch=53 start_time=2020-11-11T01:49:01Z\n" ..
                        "GATTACA\n" ..
                        "+\n" ..
                        "IIIIIII"
         local reader = bio.new_string_reader(content)
         local parser = fastq.new_parser(reader, 256)
         
         local record, err = parser:next()
         assert.is_nil(err)
         assert.equals("e3cc70d5-90ef-49b6-bbe1-cfef99537d73", record.identifier)
         assert.equals("GATTACA", record.sequence)
         assert.equals("IIIIIII", record.quality)
         assert.equals("53", record.optionals["ch"])
         assert.equals("13956", record.optionals["read"])
         
         local _, eof_err = parser:next()
         assert.equals("EOF", eof_err)
      end)

      it("should parse FASTQ with newline ending", function()
         local content = "@test1\nGATTACA\n+\nIIIIIII\n"
         local reader = bio.new_string_reader(content)
         local parser = fastq.new_parser(reader, 256)
         
         local record, err = parser:next()
         assert.is_nil(err)
         assert.equals("test1", record.identifier)
         assert.equals("GATTACA", record.sequence)
         assert.equals("IIIIIII", record.quality)
         
         local _, eof_err = parser:next()
         assert.equals("EOF", eof_err)
      end)

      it("should parse multiple sequences", function()
         local content = "@seq1\nGATTACA\n+\nIIIIIII\n" ..
                        "@seq2\nTGCAT\n+\nAAAAA\n"
         local reader = bio.new_string_reader(content)
         local parser = fastq.new_parser(reader, 256)
         
         local record1, err1 = parser:next()
         assert.is_nil(err1)
         assert.equals("seq1", record1.identifier)
         assert.equals("GATTACA", record1.sequence)
         assert.equals("IIIIIII", record1.quality)
         
         local record2, err2 = parser:next()
         assert.is_nil(err2)
         assert.equals("seq2", record2.identifier)
         assert.equals("TGCAT", record2.sequence)
         assert.equals("AAAAA", record2.quality)
         
         local _, eof_err = parser:next()
         assert.equals("EOF", eof_err)
      end)
   end)

   describe("error handling", function()
      it("should error on missing sequence", function()
         local content = "@test1\n+\nIIIIIII\n"
         local reader = bio.new_string_reader(content)
         local parser = fastq.new_parser(reader, 256)
         
         local _, err = parser:next()
         assert.is_not_nil(err)
         assert.matches("empty fastq sequence", err)
      end)

      it("should error on missing quality", function()
         local content = "@test1\nGATTACA\n+\n"
         local reader = bio.new_string_reader(content)
         local parser = fastq.new_parser(reader, 256)
         
         local _, err = parser:next()
         assert.is_not_nil(err)
         assert.matches("empty quality sequence", err)
      end)

      it("should error on missing identifier", function()
         local content = "GATTACA\n+\nIIIIIII\n"
         local reader = bio.new_string_reader(content)
         local parser = fastq.new_parser(reader, 256)
         
         local _, err = parser:next()
         assert.is_not_nil(err)
         assert.matches("did not find fastq start '@'", err)
      end)

      it("should error on empty sequence", function()
         local content = "@test1\n\n+\nIIIIIII\n"
         local reader = bio.new_string_reader(content)
         local parser = fastq.new_parser(reader, 256)
         
         local _, err = parser:next()
         assert.is_not_nil(err)
         assert.matches("empty fastq sequence", err)
      end)

      it("should error on missing plus line", function()
         local content = "@test1\nGATTACA\nIIIIIII\n"
         local reader = bio.new_string_reader(content)
         local parser = fastq.new_parser(reader, 256)
         
         local _, err = parser:next()
         assert.is_not_nil(err)
         assert.matches("empty quality sequence", err)
      end)

      it("should error on sequence/quality length mismatch", function()
         local content = "@test\nATCG\n+\nII\n"
         local reader = bio.new_string_reader(content)
         local parser = fastq.new_parser(reader, 256)
         
         local _, err = parser:next()
         assert.is_not_nil(err)
         assert.matches("Got different lengths for sequence", err)
      end)

      it("should error on invalid sequence characters", function()
         local content = "@test\nATXG\n+\nIIII\n"
         local reader = bio.new_string_reader(content)
         local parser = fastq.new_parser(reader, 256)
         
         local _, err = parser:next()
         assert.is_not_nil(err)
         assert.matches("Only letters ATGCN are allowed", err)
      end)
   end)

   describe("deep copy", function()
      it("should properly copy all fields including optionals", function()
         local content = "@test1 runid=12345 ch=53\nGATTACA\n+\nIIIIIII\n"
         local reader = bio.new_string_reader(content)
         local parser = fastq.new_parser(reader, 256)
         local read, err = parser:next()
         assert.is_nil(err)
         local copy = read:deep_copy()
         
         assert.equals(read.identifier, copy.identifier)
         assert.equals(read.sequence, copy.sequence)
         assert.equals(read.quality, copy.quality)
         assert.equals(read.optionals["runid"], copy.optionals["runid"])
         assert.equals(read.optionals["ch"], copy.optionals["ch"])
         
         -- Modify original, copy should remain unchanged
         read.optionals["new_field"] = "value"
         assert.is_nil(copy.optionals["new_field"])
      end)
   end)

   describe("buffer handling", function()
      it("should handle small buffer sizes", function()
         local content = "@test1\nGATTACA\n+\nIIIIIII\n@test2\nATCG\n+\nAAAA\n"
         local reader = bio.new_string_reader(content)
         local parser = fastq.new_parser(reader, 8)
         
         local record1, err1 = parser:next()
         assert.is_nil(err1)
         assert.equals("test1", record1.identifier)
         
         local record2, err2 = parser:next()
         assert.is_nil(err2)
         assert.equals("test2", record2.identifier)
         
         local _, eof_err = parser:next()
         assert.equals("EOF", eof_err)
      end)
   end)
end)
