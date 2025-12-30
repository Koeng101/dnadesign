local dnadesign = require("dnadesign")
local bio = dnadesign.bio

describe("StringReader", function()
    describe("creation", function()
        it("should create with valid string input", function()
            local reader = bio.new_string_reader("test content")
            assert.is_not_nil(reader)
        end)

        it("should error with non-string input", function()
            assert.has_error(function()
                bio.new_string_reader(123)
            end)
        end)
    end)

    describe("reading", function()
        it("should read exact number of bytes requested", function()
            local reader = bio.new_string_reader("hello world")
            local chunk, err = reader:read(5)
            assert.is_nil(err)
            assert.equals("hello", chunk)
        end)

        it("should read remaining bytes if less than requested", function()
            local reader = bio.new_string_reader("hello")
            local chunk, err = reader:read(10)
            assert.is_nil(err)
            assert.equals("hello", chunk)

            -- Next read should be EOF
            chunk, err = reader:read(1)
            assert.equals("EOF", err)
            assert.is_nil(chunk)
        end)

        it("should handle multiple reads", function()
            local reader = bio.new_string_reader("hello world")
            local chunk, err

            chunk, err = reader:read(5)
            assert.is_nil(err)
            assert.equals("hello", chunk)

            chunk, err = reader:read(1)
            assert.is_nil(err)
            assert.equals(" ", chunk)

            chunk, err = reader:read(5)
            assert.is_nil(err)
            assert.equals("world", chunk)
        end)

        it("should handle empty string", function()
            local reader = bio.new_string_reader("")
            local chunk, err = reader:read(1)
            assert.equals("EOF", err)
            assert.is_nil(chunk)
        end)

        it("should track line numbers across reads", function()
            local reader = bio.new_string_reader("line1\nline2\nline3")
            assert.equals(0, reader:get_line_number())
            
            reader:read(6)  -- reads "line1\n"
            assert.equals(1, reader:get_line_number())
            
            reader:read(6)  -- reads "line2\n"
            assert.equals(2, reader:get_line_number())
            
            reader:read(5)  -- reads "line3"
            assert.equals(2, reader:get_line_number())
        end)

        it("should handle consecutive reads after EOF", function()
            local reader = bio.new_string_reader("test")
            local chunk, err = reader:read(10)
            assert.is_nil(err)
            assert.equals("test", chunk)

            -- First EOF
            chunk, err = reader:read(1)
            assert.equals("EOF", err)
            assert.is_nil(chunk)

            -- Subsequent EOFs
            chunk, err = reader:read(1)
            assert.equals("EOF", err)
            assert.is_nil(chunk)
        end)
    end)

    describe("closing", function()
        it("should close successfully", function()
            local reader = bio.new_string_reader("test")
            local success, err = reader:close()
            assert.is_true(success)
            assert.is_nil(err)

            -- Reading after close should return EOF
            local chunk, read_err = reader:read(1)
            assert.equals("EOF", read_err)
            assert.is_nil(chunk)
        end)
    end)

    describe("edge cases", function()
        it("should handle zero-byte reads", function()
            local reader = bio.new_string_reader("test")
            local chunk, err = reader:read(0)
            assert.is_nil(err)
            assert.equals("", chunk)
        end)

        it("should handle non-ascii characters", function()
            local reader = bio.new_string_reader("hello⚡world")
            local chunk, err = reader:read(8)
            assert.is_nil(err)
            assert.equals("hello⚡", chunk)
        end)
        
        it("should handle large reads", function()
            local big_string = string.rep("x", 1000)
            local reader = bio.new_string_reader(big_string)
            local chunk, err = reader:read(2000)
            assert.is_nil(err)
            assert.equals(big_string, chunk)
        end)
    end)
end)
