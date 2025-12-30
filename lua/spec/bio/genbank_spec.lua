local dnadesign = require("dnadesign")
local genbank = dnadesign.genbank
local transform = dnadesign.transform
local bio = dnadesign.bio

describe("Genbank", function()

local function print_table(t, indent, printed)
    indent = indent or ""
    printed = printed or {}
    
    if printed[t] then
        print(indent .. "<circular>")
        return
    end
    printed[t] = true
    
    if type(t) ~= "table" then
        print(indent .. tostring(t))
        return
    end
    
    for k, v in pairs(t) do
        if type(v) == "table" then
            print(indent .. tostring(k) .. ":")
            print_table(v, indent .. "  ", printed)
        else
            print(indent .. tostring(k) .. ": " .. tostring(v))
        end
    end
end

-- Helper function for deep table comparison
local function deep_compare_tables(t1, t2)
    if type(t1) ~= "table" or type(t2) ~= "table" then
        return t1 == t2
    end
    
    -- First, check if all keys in t1 exist in t2 with same values
    for k, v1 in pairs(t1) do
        local v2 = t2[k]
        if type(v1) == "table" and type(v2) == "table" then
            if not deep_compare_tables(v1, v2) then
                print("DEBUG: Nested table mismatch for key:", k)
                return false
            end
        elseif v1 ~= v2 then
            print("DEBUG: Value mismatch for key:", k)
            print("t1 value:", v1)
            print("t2 value:", v2)
            return false
        end
    end
    
    -- Then check if t2 has any extra keys
    for k, _ in pairs(t2) do
        if t1[k] == nil then
            print("DEBUG: Extra key in second table:", k)
            return false
        end
    end
    
    return true
end

local function compare_genbanks(gb1, gb2)
    if not gb1 or not gb2 then
        print("DEBUG: One of the genbank objects is nil")
        return false
    end
    
    -- Compare Meta fields with deep comparison
    if not deep_compare_tables(gb1.Meta, gb2.Meta) then
        print("DEBUG: Meta fields differ (detailed comparison)")
		print("GB1 References:")
		print_table(gb1.Meta)
		print("\nGB2 References:")
		print_table(gb2.Meta)
        return false
    end
    
    if gb1.Sequence ~= gb2.Sequence then
        print("DEBUG: Sequences differ")
        print("GB1 Sequence length:", #gb1.Sequence)
        print("GB2 Sequence length:", #gb2.Sequence)
        -- Print first 50 chars of each sequence if they differ
        print("GB1 Sequence (first 50):", gb1.Sequence:sub(1, 50))
        print("GB2 Sequence (first 50):", gb2.Sequence:sub(1, 50))
        return false
    end
    
    -- Compare features count
    if #gb1.Features ~= #gb2.Features then
        print("DEBUG: Feature count mismatch")
        print("GB1 Features:", #gb1.Features)
        print("GB2 Features:", #gb2.Features)
        return false
    end
    
    -- Compare features with detailed output
    for i, f1 in ipairs(gb1.Features) do
        local f2 = gb2.Features[i]
        
        if f1.Type ~= f2.Type then
            print(string.format("Feature %d Type mismatch:", i))
            print("GB1:", f1.Type)
            print("GB2:", f2.Type)
            return false
        end
        
        if f1.Description ~= f2.Description then
            print(string.format("Feature %d Description mismatch:", i))
            print("GB1:", f1.Description)
            print("GB2:", f2.Description)
            return false
        end
        
        if f1.Sequence ~= f2.Sequence then
            print(string.format("Feature %d Sequence mismatch:", i))
            print("GB1 length:", #f1.Sequence)
            print("GB2 length:", #f2.Sequence)
            print("GB1 (first 50):", f1.Sequence:sub(1, 50))
            print("GB2 (first 50):", f2.Sequence:sub(1, 50))
            return false
        end
        
        -- Compare Location with deep comparison since it might be a table
        if type(f1.Location) == "table" and type(f2.Location) == "table" then
            if not deep_compare_tables(f1.Location, f2.Location) then
                print(string.format("Feature %d Location table mismatch:", i))
                print("GB1 GbkLocationString:", f1.Location.GbkLocationString)
                print("GB2 GbkLocationString:", f2.Location.GbkLocationString)
                return false
            end
        elseif f1.Location ~= f2.Location then
            print(string.format("Feature %d Location mismatch:", i))
            print("GB1:", f1.Location)
            print("GB2:", f2.Location)
            return false
        end
        
        -- Compare attributes with deep comparison
        if not deep_compare_tables(f1.Attributes, f2.Attributes) then
            print(string.format("Feature %d Attributes mismatch:", i))
            -- Print all attribute keys for comparison
            local attr1_keys, attr2_keys = {}, {}
            for k in pairs(f1.Attributes) do table.insert(attr1_keys, k) end
            for k in pairs(f2.Attributes) do table.insert(attr2_keys, k) end
            print("GB1 Attribute keys:", table.concat(attr1_keys, ", "))
            print("GB2 Attribute keys:", table.concat(attr2_keys, ", "))
            return false
        end
    end
    
    return true
end

    describe("Core IO", function()
        local single_gbk_paths = {
            "./spec/bio/data/t4_intron.gb",
            "./spec/bio/data/puc19.gbk", 
            "./spec/bio/data/puc19_snapgene.gb",
            "./spec/bio/data/benchling.gb",
            "./spec/bio/data/phix174.gb",
            "./spec/bio/data/sample.gbk"
        }

        it("should handle read/write roundtrip", function()
            for _, gbk_path in ipairs(single_gbk_paths) do
                -- Read file content
                local f = io.open(gbk_path, "r")
                assert.is_not_nil(f)
                local content = f:read("*a")
                f:close()
                
                -- Create parser with StringReader
                local reader = bio.new_string_reader(content)
                local parser = genbank.new_parser(reader)
                local gbk, err = parser:next()
                assert.is_nil(err)
                assert.is_not_nil(gbk)

                -- Write to string buffer
                local writer = bio.new_string_writer()
                local written, err = gbk:write_to(writer)
                assert.is_nil(err)
                assert.is_true(written > 0)

                -- Read back and compare
                local write_test_reader = bio.new_string_reader(reader.content)
                local write_test_parser = genbank.new_parser(write_test_reader)
                local write_test_gbk, err = write_test_parser:next()
                assert.is_nil(err)
                assert.is_true(compare_genbanks(gbk, write_test_gbk))
            end
        end)
    end)

    it("should handle multiline feature parsing", function()
        local f = io.open("./spec/bio/data/pichia_chr1_head.gb", "r")
        assert.is_not_nil(f)
        local content = f:read("*a")
        f:close()

        local reader = bio.new_string_reader(content)
        local parser = genbank.new_parser(reader, -1)
        local pichia, err = parser:next()
        assert.is_nil(err)

        -- Check the location string 
        local multiline_output
        for _, feature in ipairs(pichia.Features) do
            multiline_output = feature.Location.GbkLocationString
        end

        assert.equals("join(<459260..459456,459556..459637,459685..459739, 459810..>460126)", 
                     multiline_output)
    end)

    it("should handle multi-genbank IO", function()
        -- Read original multi-genbank file
        local f = io.open("./spec/bio/data/multiGbk_test.seq", "r")
        assert.is_not_nil(f)
        local content = f:read("*a")
        f:close()

        local reader = bio.new_string_reader(content)
        local parser = genbank.new_parser(reader, -1)
        
        local multi_gbk = {}
        while true do
            local gbk, err = parser:next()
            if err == "EOF" then
                break
            end
            assert.is_nil(err)
            table.insert(multi_gbk, gbk)
        end

        -- Write to string buffer
        local writer = bio.new_string_writer()
        for _, gb in ipairs(multi_gbk) do
            local _, err = gb:write_to(writer)
            assert.is_nil(err)
        end

        -- Read back and compare
        local write_test_reader = bio.new_string_reader(writer.content)
        local write_test_parser = genbank.new_parser(write_test_reader, -1)
        local write_test_gbk = {}
        
        while true do
            local gbk, err = write_test_parser:next()
            if err == "EOF" then
                break
            end
            assert.is_nil(err)
            table.insert(write_test_gbk, gbk)
        end
        
        assert.equals(#multi_gbk, #write_test_gbk)
        for i=1, #multi_gbk do
            assert.is_true(compare_genbanks(multi_gbk[i], write_test_gbk[i]))
        end
    end)

    it("should handle location string building", function()
        local f = io.open("./spec/bio/data/sample.gbk", "r")
        assert.is_not_nil(f)
        local content = f:read("*a")
        f:close()

        local reader = bio.new_string_reader(content)
        local parser = genbank.new_parser(reader, -1)
        local gbk, err = parser:next()
        assert.is_nil(err)

        -- Clear GbkLocationString to test building
        for i, feature in ipairs(gbk.Features) do
            gbk.Features[i].Location.GbkLocationString = ""
        end

        -- Write to string buffer
        local writer = bio.new_string_writer()
        local _, err = gbk:write_to(writer)
        assert.is_nil(err)

        -- Compare original vs rebuilt
        local test_input_reader = bio.new_string_reader(content)
        local test_output_reader = bio.new_string_reader(writer.content)
        
        local test_input_parser = genbank.new_parser(test_input_reader, -1)
        local test_output_parser = genbank.new_parser(test_output_reader, -1)
        
        local test_input_gbk = test_input_parser:next()
        local test_output_gbk = test_output_parser:next()

        assert.is_true(compare_genbanks(test_input_gbk, test_output_gbk))
    end)

    it("should handle Snapgene genbank regression", function()
        local f = io.open("./spec/bio/data/puc19_snapgene.gb", "r")
        assert.is_not_nil(f)
        local content = f:read("*a")
        f:close()

        local reader = bio.new_string_reader(content)
        local parser = genbank.new_parser(reader, -1)
        local snapgene, err = parser:next()
        assert.is_nil(err)
        assert.is_not_equal("", snapgene.Sequence)
    end)

    it("should parse various location types correctly", function()
        local f = io.open("./spec/bio/data/t4_intron.gb", "r")
        assert.is_not_nil(f)
        local content = f:read("*a")
        f:close()

        local reader = bio.new_string_reader(content)
        local parser = genbank.new_parser(reader)
        local gbk, err = parser:next()
        assert.is_nil(err)

		-- gotta + for go->lua
        -- Test simple range (Feature 2)
        local seq1, err = gbk.Features[2]:get_sequence()
        assert.is_nil(err)
        assert.equals("atgagattacaacgccagagcatcaaagattcagaagttagaggtaaatggtattttaatatcatcggtaaagattctgaacttgttgaaaaagctgaacatcttttacgtgatatgggatgggaagatgaatgcgatggatgtcctctttatgaagacggagaaagcgcaggattttggatttaccattctgacgtcgagcagtttaaagctgattggaaaattgtgaaaaagtctgtttga", seq1)

        -- Test join (Feature 7)
        local seq2, err = gbk.Features[7]:get_sequence()
        assert.is_nil(err)
        assert.equals("atgaaacaataccaagatttaattaaagacatttttgaaaatggttatgaaaccgatgatcgtacaggcacaggaacaattgctctgttcggatctaaattacgctgggatttaactaaaggttttcctgcggtaacaactaagaagctcgcctggaaagcttgcattgctgagctaatatggtttttatcaggaagcacaaatgtcaatgatttacgattaattcaacacgattcgttaatccaaggcaaaacagtctgggatgaaaattacgaaaatcaagcaaaagatttaggataccatagcggtgaacttggtccaatttatggaaaacagtggcgtgattttggtggtgtagaccaaattatagaagttattgatcgtattaaaaaactgccaaatgataggcgtcaaattgtttctgcatggaatccagctgaacttaaatatatggcattaccgccttgtcatatgttctatcagtttaatgtgcgtaatggctatttggatttgcagtggtatcaacgctcagtagatgttttcttgggtctaccgtttaatattgcgtcatatgctacgttagttcatattgtagctaagatgtgtaatcttattccaggggatttgatattttctggtggtaatactcatatctatatgaatcacgtagaacaatgtaaagaaattttgaggcgtgaacctaaagagctttgtgagctggtaataagtggtctaccttataaattccgatatctttctactaaagaacaattaaaatatgttcttaaacttaggcctaaagatttcgttcttaacaactatgtatcacaccctcctattaaaggaaagatggcggtgtaa", seq2)

        -- Test complement (Feature 11)
        local seq3, err = gbk.Features[11]:get_sequence()
        assert.is_nil(err)
        assert.equals("ttattcactacccggcatagacggcccacgctggaataattcgtcatattgtttttccgttaaaacagtaatatcgtagtaacagtcagaagaagttttaactgtggaaattttattatcaaaatactcacgagtcattttatgagtatagtattttttaccataaatggtaataggctgttctggtcctggaacttctaactcgcttgggttaggaagtgtaaaaagaactacaccagaagtatctttaaatcgtaaaatcat", seq3)

        -- Test join(complement(), complement()) (Feature 4)
        local seq4, err = gbk.Features[4]:get_sequence()
        assert.is_nil(err)
        assert.equals("ataccaatttaatcattcatttatatactgattccgtaagggttgttacttcatctattttataccaatgcgtttcaaccatttcacgcttgcttatatcatcaagaaaacttgcgtctaattgaactgttgaattaacacgatgccttttaacgatgcgagaaacaactacttcatctgcataaggtaatgcagcatataacagagcaggcccgccaattacacttactttagaattctgatcaagcatagtttcgaatggtgcattagggcttgacacttgaatttcgccgccagaaatgtaagttatatattgctcccaagtaatatagaaatgtgctaaatcgccgtctttagttacaggataatcacgcgcaaggtcacacaccacaatatggctacgaccaggaagtaatgtaggcaatgactggaacgttttagcacccataatcataattgtgccttcagtacgagctttaaaattctggaggtcctttttaactcgtccccatggtaaaccatcacctaaaccgaatgctaattcattaaagccgtcgaccgttttagttggaga", seq4)
    end)

    it("should handle consortium regression", function()
        local f = io.open("./spec/bio/data/puc19_consrtm.gbk", "r")
        assert.is_not_nil(f)
        local content = f:read("*a")
        f:close()

        local reader = bio.new_string_reader(content)
        local parser = genbank.new_parser(reader, -1)
        local _, err = parser:next()
        assert.is_nil(err)
    end)
end)
