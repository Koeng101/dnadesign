local dnadesign = require("dnadesign")
local bio = dnadesign.bio
local slow5 = dnadesign.slow5

describe("SLOW5 Parser", function()
   -- Example SLOW5 file content as string
   local example_slow5 = [[
#slow5_version	0.2.0
#num_read_groups	1
@asic_id	4175987214
@asic_id_eeprom	5910407
@asic_temp	31.649540
@asic_version	IA02D
#char*	uint32_t	double	double	double	double	uint64_t	int16_t*	uint64_t	int32_t	uint8_t	double	enum{unknown,partial,mux_change,unblock_mux_change,data_service_unblock_mux_change,signal_positive,signal_negative}	char*
#read_id	read_group	digitisation	offset	range	sampling_rate	len_raw_signal	raw_signal	start_time	read_number	start_mux	median_before	end_reason	channel_number
0026631e-33a3-49ab-aa22-3ab157d71f8b	0	8192	16	1489.52832	4000	3	430,472,463	8318394	5383	1	219.133423	5	10
]]

   local header_without_tabs = [[
#slow5_version	0.2.0
#num_read_groups	1
@bad
@asic_id	4175987214
]]

   local header_bad_num_groups = [[
#slow5_version	0.2.0
#num_read_groups	bad!
@asic_id	4175987214
]]

   local header_not_enough_attrs = [[
#slow5_version	0.2.0
#num_read_groups	2
@asic_id	4175987214
]]

   describe("basic parsing", function()
      it("should parse valid SLOW5 data", function()
         local reader = bio.new_string_reader(example_slow5)
         local parser, err = slow5.new_parser(reader, 2 * 32 * 1024)
         
         assert.is_nil(err)
         
         local header = parser:header()
         assert.equals(1, #header.header_values)
         assert.equals("4175987214", header.header_values[1].attributes["@asic_id"])
         
         local read, read_err = parser:next()
         assert.is_nil(read_err)
         assert.equals(430, read.raw_signal[1])
         assert.equals("0026631e-33a3-49ab-aa22-3ab157d71f8b", read.read_id)
      end)
   end)

   describe("header validation", function()
      it("should detect header without tabs", function()
         local reader = bio.new_string_reader(header_without_tabs)
         local parser, err = slow5.new_parser(reader, 2 * 32 * 1024)
         assert.is_not_nil(err)
         assert.matches("Got following line without tabs: @bad", err)
      end)

      it("should detect invalid num_read_groups", function()
         local reader = bio.new_string_reader(header_bad_num_groups)
         local parser, err = slow5.new_parser(reader, 2 * 32 * 1024)
         assert.is_not_nil(err)
         assert.matches("Invalid num_read_groups value", err)
      end)

      it("should detect not enough attributes", function()
         local reader = bio.new_string_reader(header_not_enough_attrs)
         local parser, err = slow5.new_parser(reader, 2 * 32 * 1024)
         assert.is_not_nil(err)
         assert.matches("Improper amount of information for read groups. Needed 3, got 2", err)
      end)
   end)

   describe("read validation", function()
      local function test_read_error(content, expected_error)
         local reader = bio.new_string_reader(content)
         local parser, _ = slow5.new_parser(reader, 2 * 32 * 1024)
         
         local _, err = parser:next()
         assert.is_not_nil(err)
         assert.matches(expected_error, err)
      end

      it("should validate read_group", function()
         local content = example_slow5:gsub("0	8192", "bad	8192")
         test_read_error(content, "Failed convert read_group 'bad' to number on line 9")
      end)

      it("should validate digitisation", function()
         local content = example_slow5:gsub("8192	16", "bad	16")
         test_read_error(content, "Failed to convert digitisation 'bad' to number on line 9")
      end)

      it("should validate offset", function()
         local content = example_slow5:gsub("16	1489", "bad	1489")
         test_read_error(content, "Failed to convert offset 'bad' to number on line 9")
      end)

      it("should validate range", function()
         local content = example_slow5:gsub("1489.52832	4000", "bad	4000")
         test_read_error(content, "Failed to convert range 'bad' to number on line 9")
      end)

      it("should validate sampling_rate", function()
         local content = example_slow5:gsub("4000	3", "bad	3")
         test_read_error(content, "Failed to convert sampling_rate 'bad' to number on line 9")
      end)

      it("should validate raw_signal", function()
         local content = example_slow5:gsub("430,472,463", "430,bad,463")
         test_read_error(content, "Failed to convert raw signal 'bad' to number on line 9")
      end)
   end)

   describe("simple example", function()
      it("should parse minimal SLOW5 data", function()
         local content = [[
#slow5_version	0.2.0
#num_read_groups	1
@asic_id	4175987214
#char*	uint32_t	double	double	double	double	uint64_t	int16_t*	uint64_t	int32_t	uint8_t	double	enum{unknown,partial,mux_change,unblock_mux_change,data_service_unblock_mux_change,signal_positive,signal_negative}	char*
#read_id	read_group	digitisation	offset	range	sampling_rate	len_raw_signal	raw_signal	start_time	read_number	start_mux	median_before	end_reason	channel_number
0026631e-33a3-49ab-aa22-3ab157d71f8b	0	8192	16	1489.52832	4000	3	430,472,463	8318394	5383	1	219.133423	5	10
]]
         local reader = bio.new_string_reader(content)
         local parser, err = slow5.new_parser(reader, 2 * 32 * 1024)
         
         assert.is_nil(err)
         local read, read_err = parser:next()
         assert.is_nil(read_err)
         assert.equals(430, read.raw_signal[1])
      end)
   end)
end)
