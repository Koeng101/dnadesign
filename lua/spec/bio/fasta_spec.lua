local dnadesign = require("dnadesign")
local bio = dnadesign.bio
local fasta = dnadesign.fasta

describe("FASTA Parser", function()
   describe("header", function()
      it("should return empty header without error", function()
         local reader = bio.new_string_reader("")
         local parser = fasta.new_parser(reader, 256)
         local header, err = parser:header()
         assert.is_nil(err)
         assert.equals("FASTA", header.format)
         assert.equals("", header:to_string())
      end)
   end)

   describe("parser", function()
      it("should parse EOF-ended FASTA", function()
         local reader = bio.new_string_reader(">humen\nGATTACA\nCATGAT")
         local parser = fasta.new_parser(reader, 256)
         
         local record, err = parser:next()
         assert.is_nil(err)
         assert.equals("humen", record.identifier)
         assert.equals("GATTACACATGAT", record.sequence)
         
         local _, eof_err = parser:next()
         assert.equals("EOF", eof_err)
      end)

      it("should parse FASTA with newline ending", function()
         local reader = bio.new_string_reader(">humen\nGATTACA\nCATGAT\n")
         local parser = fasta.new_parser(reader, 256)
         
         local record, err = parser:next()
         assert.is_nil(err)
         assert.equals("humen", record.identifier)
         assert.equals("GATTACACATGAT", record.sequence)
         
         local _, eof_err = parser:next()
         assert.equals("EOF", eof_err)
      end)

      it("should parse multiple sequences with comments", function()
         local content = ">doggy or something\nGATTACA\n\nCATGAT\n\n;a fun comment\n" ..
                        ">homunculus\nAAAA\n"
         local reader = bio.new_string_reader(content)
         local parser = fasta.new_parser(reader, 256)
         
         local record1, err1 = parser:next()
         assert.is_nil(err1)
         assert.equals("doggy or something", record1.identifier)
         assert.equals("GATTACACATGAT", record1.sequence)
         
         local record2, err2 = parser:next()
         assert.is_nil(err2)
         assert.equals("homunculus", record2.identifier)
         assert.equals("AAAA", record2.sequence)
         
         local _, eof_err = parser:next()
         assert.equals("EOF", eof_err)
      end)
   end)

   describe("empty FASTA handling", function()
      it("should handle invalid FASTA without identifier", function()
         local content = "testing\natagtagtagtagtagatgatgatgatgagatg\n\n\n\n\n\n\n\n\n\n\n"
         local reader = bio.new_string_reader(content)
         local parser = fasta.new_parser(reader, 256)
         
         local record, err = parser:next()
         assert.is_not_nil(err)
         assert.matches("invalid input: missing sequence identifier", err)
      end)

      it("should handle empty sequence", function()
         local content = ">testing\natagtagtagtagtagatgatgatgatgagatg\n>testing2\n\n\n\n\n\n\n\n\n\n"
         local reader = bio.new_string_reader(content)
         local parser = fasta.new_parser(reader, 256)
         
         local record1, err1 = parser:next()
         assert.is_nil(err1)
         assert.equals("testing", record1.identifier)
         assert.equals("atagtagtagtagtagatgatgatgatgagatg", record1.sequence)
         
         local _, err2 = parser:next()
         assert.equal(err2, "EOF")
      end)
   end)

   describe("buffer handling", function()
      it("should NOT handle small buffer sizes", function()
         local content = ">test\natagtagtagtagtagatgatgatgatgagatg\n>test2\n\n\n\n\n\n\n\n\n\n"
         local reader = bio.new_string_reader(content)
         local parser = fasta.new_parser(reader, 8)
         
         local record1, err1 = parser:next()
         assert.is_nil(err1)
         
         local _, err2 = parser:next()
         assert.equal(err2, "EOF")
      end)
   end)
end)
