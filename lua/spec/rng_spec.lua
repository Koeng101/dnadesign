-- rng_spec.lua
describe("hermetic RNG", function()
    local original_random

    before_each(function()
        -- Save original random
        original_random = math.random
        
        -- Clear global seeds
        SEED_1 = nil
        SEED_2 = nil
        
        -- Reset modules and reload
        package.loaded["dnadesign"] = nil
        require("dnadesign")
    end)

    after_each(function()
        -- Restore original random
        math.random = original_random
    end)

    describe("initialization", function()
        it("uses default seeds when globals are not set", function()
            local first_random = math.random()
            local second_random = math.random()

            -- Reset to initial state
            package.loaded["dnadesign"] = nil
            require("dnadesign")

            -- Should get same sequence
            assert.are.equal(first_random, math.random())
            assert.are.equal(second_random, math.random())
        end)

        it("uses global values when set", function()
            -- Get sequence with default seeds
            local default1 = math.random()
            local default2 = math.random()
            
            -- Set new seeds and reload
            _G.SEED_1 = 54321
            _G.SEED_2 = 98765
            package.loaded["dnadesign"] = nil
            require("dnadesign")
            
            -- Get sequence with new seeds
            local new1 = math.random()
            local new2 = math.random()
            
            -- Should be different sequences
            assert.are_not.equal(default1, new1)
            assert.are_not.equal(default2, new2)
            
            -- Reset with same seeds
            package.loaded["dnadesign"] = nil
            require("dnadesign")
            
            -- Should get same sequence again
            assert.are.equal(new1, math.random())
            assert.are.equal(new2, math.random())
        end)
    end)

    describe("range handling", function()
        it("handles math.random() -> [0,1)", function()
            local result = math.random()
            assert.is_true(result >= 0 and result < 1)
        end)

        it("handles math.random(N) -> [1,N]", function()
            local first = math.random(10)
            assert.is_true(first >= 1 and first <= 10)
            assert.is_true(math.floor(first) == first)

            -- Reset and check consistency
            package.loaded["dnadesign"] = nil
            require("dnadesign")
            assert.are.equal(first, math.random(10))
        end)

        it("handles math.random(M,N) -> [M,N]", function()
            local first = math.random(5, 10)
            assert.is_true(first >= 5 and first <= 10)
            assert.is_true(math.floor(first) == first)

            -- Reset and check consistency
            package.loaded["dnadesign"] = nil
            require("dnadesign")
            assert.are.equal(first, math.random(5, 10))
        end)
    end)

    describe("hermetic properties", function()
        it("disables math.randomseed", function()
            assert.has_error(function()
                math.randomseed(123)
            end, "math.randomseed is disabled in hermetic environment. Use SEED_1 and SEED_2 environment variables instead.")
        end)

        it("maintains state correctly across resets", function()
            local num1 = math.random()
            local num2 = math.random()
            local num3 = math.random()
            
            -- Reset module
            package.loaded["dnadesign"] = nil
            require("dnadesign")
            
            -- Should get same sequence
            assert.are.equal(num1, math.random())
            assert.are.equal(num2, math.random())
            assert.are.equal(num3, math.random())
        end)
    end)

    describe("global replacement", function()
        it("replaces global math.random", function()
            assert.are_not.equal(original_random, math.random)
        end)

        it("affects math.random calls in other modules", function()
            local first = math.random()
            local second = math.random()
            
            -- Create a new module that uses math.random
            local other_module = {}
            function other_module.generate_numbers()
                return {math.random(), math.random()}
            end
            
            -- Get numbers from other module
            local other_results = other_module.generate_numbers()
            
            -- Reset module
            package.loaded["dnadesign"] = nil
            require("dnadesign")
            
            -- Both direct and other module calls should be consistent
            assert.are.equal(first, math.random())
            assert.are.equal(second, math.random())
            
            local new_other_results = other_module.generate_numbers()
            assert.are.equal(other_results[1], new_other_results[1])
            assert.are.equal(other_results[2], new_other_results[2])
        end)
    end)

    describe("statistical properties", function()
        it("generates numbers with reasonable distribution", function()
            local buckets = {}
            for i = 1, 10 do buckets[i] = 0 end
            
            -- Generate 10000 numbers and bucket them
            for _ = 1, 10000 do
                local n = math.random()
                local bucket = math.floor(n * 10) + 1
                buckets[bucket] = buckets[bucket] + 1
            end
            
            -- Check that each bucket has roughly 1000 ±20% entries
            for _, count in ipairs(buckets) do
                assert.is_true(count > 800 and count < 1200)
            end
        end)

        it("generates integers with reasonable distribution", function()
            local buckets = {}
            for i = 1, 6 do buckets[i] = 0 end
            
            -- Generate 6000 dice rolls
            for _ = 1, 6000 do
                local n = math.random(1, 6)
                buckets[n] = buckets[n] + 1
            end
            
            -- Check that each number appears roughly 1000 ±20% times
            for _, count in ipairs(buckets) do
                assert.is_true(count > 800 and count < 1200)
            end
        end)
    end)
end)
