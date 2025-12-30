--[[ Module rng
rng.tl - A Deterministic Random Number Generator for Hermetic Environments

This module implements L'Ecuyer's Combined Linear Congruential Generator (LCG) to provide
a completely deterministic, reproducible random number generation system. It's designed
to replace Lua's built-in math.random functionality for environments requiring
hermeticity (complete reproducibility and isolation from system state).

Mathematical Background:
- Uses two independent LCG streams with carefully chosen prime moduli:
  1. m1 = 2147483647 (2^31 - 1, a Mersenne prime)
  2. m2 = 2147483399 (a prime slightly smaller than m1)

- Multipliers chosen for full-period generation:
  1. a1 = 48271 (for m1)
  2. a2 = 40692 (for m2)

These numbers were selected because they:
- Fit well within Lua's 53-bit precision for integers
- Provide a proven period of approximately 2^46
- Pass spectral tests for good distribution of generated numbers
- Allow for efficient modulo operations (especially m1 being a Mersenne prime)

The generator combines two independent streams to break up lattice structures
that can occur in single-stream LCGs. This provides better statistical properties
than simple LCGs while maintaining complete determinism.

Usage:
    local rng = require("hermetic_rng")
    local num = rng.random()     -- returns [0,1)
    local num = rng.random(N)    -- returns [1,N]
    local num = rng.random(M,N)  -- returns [M,N]

Environment Variables:
    SEED_1 - First seed for the generator (defaults to 12345)
    SEED_2 - Second seed for the generator (defaults to 67890)

Note: This module deliberately overrides math.random and math.randomseed to ensure
      hermetic behavior. Attempts to call math.randomseed will raise an error.

Question: Does this differ among hardware platforms?
Answer: By carefully keeping all arithmetic within 2^53, this PRNG remains
consistent across IEEE-754 double-precision platforms (including x86 and ARM).
The largest product here is on the order of 1×10^14, whereas 2^53 is about
9x10^15 - so intermediate multiplications (Before taking modulus) and the floor
operations used in '%' remain exact for all valid states. Consequently, the
combined LCG sequence and the final scaling for random() are reproducible under
any IEEE-754-compliant Lua VM, ensuring hermetic behavior without
hardware-dependent variation.
]]

local record RNGState
    x: integer
    y: integer
end

local record rng
    -- Core generator functions
    new: function(seed_x: integer, seed_y: integer): RNGState
    next: function(state: RNGState): (integer, RNGState)
    random: function(state: RNGState, min?: integer, max?: integer): (number, RNGState)
    
    -- Current state
    state: RNGState
    
	-- Matches all possible math.random call patterns:
    -- math.random()     -> [0,1)
    -- math.random(n)    -> [1,n]
    -- math.random(m,n)  -> [m,n]
    random_wrapper: function(m?: integer, n?: integer): number
end

-- Mathematical constants for the combined LCG
local m1 <const> = 2147483647  -- 2^31 - 1, Mersenne prime
local m2 <const> = 2147483399  -- Prime slightly smaller than m1

local a1 <const> = 48271       -- Full-period multiplier for m1
local a2 <const> = 40692       -- Full-period multiplier for m2

-- Get seeds from global variables or use defaults
global SEED_1: integer
global SEED_2: integer
local seed1 = (_G.SEED_1 as integer) or 12345
local seed2 = (_G.SEED_2 as integer) or 67890

-- Create new RNG state
function rng.new(seed_x: integer, seed_y: integer): RNGState
    return {
        x = seed_x % m1,
        y = seed_y % m2
    }
end

-- Generate next random number and state
function rng.next(state: RNGState): (integer, RNGState)
    local new_x = (state.x * a1) % m1
    local new_y = (state.y * a2) % m2
    
    -- Combine streams and ensure positive result
    local result = (new_x - new_y) % m1
    if result <= 0 then
        result = result + m1 - 1
    end
    
    return result, { x = new_x, y = new_y }
end

-- Generate random number in range
function rng.random(state: RNGState, min?: integer, max?: integer): (number, RNGState)
    local result, new_state = rng.next(state)
    
    -- Scale result based on arguments (matching math.random behavior)
    if not min and not max then
        -- Case: random() -> [0,1)
        return result / m1, new_state
    elseif not max then
        -- Case: random(n) -> [1,n]
        max = min
        min = 1
    end
    -- Case: random(m,n) -> [m,n]
    return min + (result % (max - min + 1)), new_state
end

-- Initialize global state
rng.state = rng.new(seed1, seed2)

-- Implementation matching all patterns
function rng.random_wrapper(m?: integer, n?: integer): number
    if m == nil and n == nil then
        -- Case: math.random()
        local result, new_state = rng.random(rng.state)
        rng.state = new_state
        return result
    elseif n == nil then
        -- Case: math.random(n)
        local result, new_state = rng.random(rng.state, 1, m as integer)
        rng.state = new_state
        return result
    else
        -- Case: math.random(m,n)
        local result, new_state = rng.random(rng.state, m as integer, n)
        rng.state = new_state
        return result
    end
end

-- Replace math.random with our hermetic version
_G.math.random = rng.random_wrapper

-- Disable math.randomseed
_G.math.randomseed = function(_: any): (integer, integer)
    error("math.randomseed is disabled in hermetic environment. Use SEED_1 and SEED_2 environment variables instead.")
	return 0, 0
end

return rng
