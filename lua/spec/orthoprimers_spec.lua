-- spec/orthoprimers_spec.lua
local dnadesign = require("dnadesign")
local orthoprimers = dnadesign.orthoprimers

describe("OrthoPrimers", function()
    describe("NewOrthogonalPrimerSet", function()
        it("creates primer set correctly", function()
            local primers = {"AAACACGTGGCAAACATTCC", "AAACCGGAGCCATACAGTAC", "AAAGCACTCTTAGGCCTCTG"}
            local ops = orthoprimers.new_orthogonal_primer_set(primers)

            assert.are.equal(3, #ops.primers)

            for _, primer in ipairs(primers) do
                assert.is_not_nil(ops.primer_use_quantity[primer])
                assert.are.equal(0, ops.primer_use_quantity[primer])
            end

            -- Check that primer_pairs is initially empty
            local pair_count = 0
            for _ in pairs(ops.primer_pairs) do
                pair_count = pair_count + 1
            end
            assert.are.equal(0, pair_count)
        end)
    end)

    describe("NewDefaultOrthogonalPrimerSet", function()
        it("creates default set with 96 primers", function()
            local ops = orthoprimers.new_default_orthogonal_primer_set()

            assert.are.equal(96, #ops.primers)

            for i = 1, #ops.primers do
                assert.are.equal(orthoprimers.ortho_primers[i], ops.primers[i])
            end
        end)
    end)

    describe("NewPrimerSet", function()
        it("generates multiple primer pairs correctly", function()
            local ops = orthoprimers.new_default_orthogonal_primer_set()

            -- Test multiple primer pairs
            for i = 1, 100 do
                local forward, reverse, err = ops:new_primer_set()

                assert.is_nil(err)
                assert.is_not.equal("", forward)
                assert.is_not.equal("", reverse)
                assert.is_not.equal(forward, reverse)

                local pair_key = orthoprimers.make_primer_pair_key(forward, reverse)
                assert.is_true(ops.primer_pairs[pair_key])

                assert.is_true(ops.primer_use_quantity[forward] > 0)
                assert.is_true(ops.primer_use_quantity[reverse] > 0)
            end
        end)

        it("returns error when primers are exhausted", function()
            local small_ops = orthoprimers.new_orthogonal_primer_set({
                "AAACACGTGGCAAACATTCC", 
                "AAACCGGAGCCATACAGTAC", 
                "AAAGCACTCTTAGGCCTCTG"
            })
            
            -- Use up available pairs
            small_ops:new_primer_set()  -- Use up one pair
            small_ops:new_primer_set()  -- Use up another pair
            small_ops:new_primer_set()  -- Use up a final pair
            
            local _, _, err = small_ops:new_primer_set()  -- Should return an error
            assert.is_not_nil(err)
        end)
    end)

    describe("MakePrimerPairKey", function()
        it("creates consistent primer pair keys", function()
            local test_cases = {
                {
                    forward = "AAACACGTGGCAAACATTCC",
                    reverse = "AAACCGGAGCCATACAGTAC",
                    expected = "AAACCGGAGCCATACAGTAC|AAACACGTGGCAAACATTCC"
                },
                {
                    forward = "AAACCGGAGCCATACAGTAC",
                    reverse = "AAACACGTGGCAAACATTCC",
                    expected = "AAACCGGAGCCATACAGTAC|AAACACGTGGCAAACATTCC"
                },
                {
                    forward = "AAAGCACTCTTAGGCCTCTG",
                    reverse = "AAAGCACTCTTAGGCCTCTG",
                    expected = "AAAGCACTCTTAGGCCTCTG|AAAGCACTCTTAGGCCTCTG"
                }
            }

            for _, tc in ipairs(test_cases) do
                local result = orthoprimers.make_primer_pair_key(tc.forward, tc.reverse)
                assert.are.equal(tc.expected, result)
            end
        end)
    end)

    describe("PrimerDistribution", function()
        it("distributes primer usage evenly", function()
            local ops = orthoprimers.new_default_orthogonal_primer_set()
            local iterations = 1000
            local primer_counts = {}

            for i = 1, iterations do
                local forward, reverse, _ = ops:new_primer_set()
                primer_counts[forward] = (primer_counts[forward] or 0) + 1
                primer_counts[reverse] = (primer_counts[reverse] or 0) + 1
            end

            -- Check if all primers are used
            for _, primer in ipairs(ops.primers) do
                assert.is_true(primer_counts[primer] and primer_counts[primer] > 0)
            end

            -- Check if primer usage is relatively evenly distributed
            -- (allowing for some variation due to randomness)
            local expected_avg = (iterations * 2) / #ops.primers
            local tolerance = 0.2 * expected_avg  -- 20% tolerance

            for primer, count in pairs(primer_counts) do
                assert.is_true(count >= expected_avg - tolerance and count <= expected_avg + tolerance)
            end
        end)
    end)
end)
