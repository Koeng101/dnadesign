-- examples/slow5_examples.lua
local dnadesign = require("dnadesign")
local bio = dnadesign.bio
local slow5 = dnadesign.slow5

describe("SLOW5 Examples", function()
    local example_slow5 = [[#slow5_version	0.2.0
#num_read_groups	1
@asic_id	4175987214
@asic_id_eeprom	5910407
@asic_temp	31.649540
@asic_version	IA02D
@auto_update	0
@auto_update_source	https://mirror.oxfordnanoportal.com/software/MinKNOW/
@device_id	MN33517
@sample_frequency	4000
@sequencing_kit	sqk-lsk109
#char*	uint32_t	double	double	double	double	uint64_t	int16_t*	uint64_t	int32_t	uint8_t	double	enum{unknown,partial,mux_change,unblock_mux_change,data_service_unblock_mux_change,signal_positive,signal_negative}	char*
#read_id	read_group	digitisation	offset	range	sampling_rate	len_raw_signal	raw_signal	start_time	read_number	start_mux	median_before	end_reason	channel_number
0026631e-33a3-49ab-aa22-3ab157d71f8b	0	8192	16	1489.52832	4000	10	430,472,463,467,454,465,463,450,450,449	8318394	5383	1	219.133423	5	10
]]

    it("demonstrates basic SLOW5 parsing", function()
        local reader = bio.new_string_reader(example_slow5)
        local parser, err = slow5.new_parser(reader, 2 * 32 * 1024)
        assert.is_nil(err)
        
        -- Check header information
        local header = parser:header()
        assert.equals("4175987214", header.header_values[1].attributes["@asic_id"])
        assert.equals("31.649540", header.header_values[1].attributes["@asic_temp"])
        
        -- Read data
        local read, read_err = parser:next()
        assert.is_nil(read_err)
        assert.equals("0026631e-33a3-49ab-aa22-3ab157d71f8b", read.read_id)
        assert.equals(430, read.raw_signal[1])
        assert.equals(449, read.raw_signal[10])
    end)

    it("demonstrates raw signal access", function()
        local reader = bio.new_string_reader(example_slow5)
        local parser, err = slow5.new_parser(reader, 2 * 32 * 1024)
        assert.is_nil(err)
        
        local read, read_err = parser:next()
        assert.is_nil(read_err)
        
        -- Check first 10 values of raw signal
        local expected = {430, 472, 463, 467, 454, 465, 463, 450, 450, 449}
        for i = 1, 10 do
            assert.equals(expected[i], read.raw_signal[i])
        end
    end)

    it("demonstrates reading multiple records", function()
        local multi_record = [[#slow5_version	0.2.0
#num_read_groups	1
@asic_id	4175987214
#char*	uint32_t	double	double	double	double	uint64_t	int16_t*	uint64_t	int32_t	uint8_t	double	enum{unknown,partial,mux_change,unblock_mux_change,data_service_unblock_mux_change,signal_positive,signal_negative}	char*
#read_id	read_group	digitisation	offset	range	sampling_rate	len_raw_signal	raw_signal	start_time	read_number	start_mux	median_before	end_reason	channel_number
read1	0	8192	16	1489.52832	4000	3	430,472,463	8318394	5383	1	219.133423	5	10
read2	0	8192	16	1489.52832	4000	3	467,454,465	8318395	5384	2	219.133423	5	11]]
        local reader = bio.new_string_reader(multi_record)
        local parser, err = slow5.new_parser(reader, 2 * 32 * 1024)
        assert.is_nil(err)
        
        -- Read first record
        local read1, err1 = parser:next()
        assert.is_nil(err1)
        assert.equals("read1", read1.read_id)
        assert.equals(430, read1.raw_signal[1])
        
        -- Read second record
        local read2, err2 = parser:next()
        assert.is_nil(err2)
        assert.equals("read2", read2.read_id)
        assert.equals(467, read2.raw_signal[1])
        
        -- Verify EOF
        local read3, err3 = parser:next()
        assert.is_nil(read3)
        assert.is_not_nil(err3)
    end)

    it("demonstrates error handling for malformed data", function()
        local malformed_slow5 = [[#slow5_version	0.2.0
#num_read_groups	1
@asic_id	4175987214
#char*	uint32_t	double	double	double	double	uint64_t	int16_t*	uint64_t	int32_t	uint8_t	double	enum{unknown,partial,mux_change,unblock_mux_change,data_service_unblock_mux_change,signal_positive,signal_negative}	char*
#read_id	read_group	digitisation	offset	range	sampling_rate	len_raw_signal	raw_signal	start_time	read_number	start_mux	median_before	end_reason	channel_number
read1	bad	8192	16	1489.52832	4000	3	430,472,463	8318394	5383	1	219.133423	5	10]]
        local reader = bio.new_string_reader(malformed_slow5)
        local parser, err = slow5.new_parser(reader, 2 * 32 * 1024)
        assert.is_nil(err)
        
        local read, read_err = parser:next()
        assert.is_not_nil(read_err)
        assert.matches("Failed convert read_group 'bad' to number", read_err)
    end)

    it("demonstrates handling of large raw signals", function()
        local large_signal = [[#slow5_version	0.2.0
#num_read_groups	1
@asic_id	4175987214
#char*	uint32_t	double	double	double	double	uint64_t	int16_t*	uint64_t	int32_t	uint8_t	double	enum{unknown,partial,mux_change,unblock_mux_change,data_service_unblock_mux_change,signal_positive,signal_negative}	char*
#read_id	read_group	digitisation	offset	range	sampling_rate	len_raw_signal	raw_signal	start_time	read_number	start_mux	median_before	end_reason	channel_number
big_read	0	8192	16	1489.52832	4000	1000	]] .. string.rep("450,", 999) .. [[450	8318394	5383	1	219.133423	5	10]]
        local reader = bio.new_string_reader(large_signal)
        local parser, err = slow5.new_parser(reader, 2 * 32 * 1024)
        assert.is_nil(err)
        
        local read, read_err = parser:next()
        assert.is_nil(read_err)
        assert.equals(1000, #read.raw_signal)
        assert.equals(450, read.raw_signal[1])
        assert.equals(450, read.raw_signal[1000])
    end)
end)
