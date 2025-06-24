-- transform_spec.lua
local dnadesign = require("dnadesign")
local transform = dnadesign.transform

-- Helper function to generate random DNA sequence
local function random_dna_sequence(length)
  local bases = {"A", "T", "C", "G"}
  local sequence = {}
  for i = 1, length do
    sequence[i] = bases[math.random(1, 4)]
  end
  return table.concat(sequence)
end

-- Helper function to generate random RNA sequence
local function random_rna_sequence(length)
  local bases = {"A", "U", "C", "G"}
  local sequence = {}
  for i = 1, length do
    sequence[i] = bases[math.random(1, 4)]
  end
  return table.concat(sequence)
end

describe("transform", function()
  describe("reverse", function()
    it("should correctly reverse sequences of even and odd lengths", function()
      local sequence = random_dna_sequence(20)
      local test_sequences = {sequence, sequence:sub(1, #sequence-1)}
      
      for _, test_sequence in ipairs(test_sequences) do
        local reversed = transform.reverse(test_sequence)
        for i = 1, math.floor(#reversed/2 + 1) do
          local got_base = reversed:sub(i,i)
          local expect = test_sequence:sub(#test_sequence-i+1, #test_sequence-i+1)
          assert.are.equal(expect, got_base, 
            string.format("mismatch at pos %d, got %s, expect %s", i, got_base, expect))
        end
      end
    end)
  end)

  describe("complement", function()
    it("should correctly complement DNA sequences", function()
      local sequence = random_dna_sequence(20)
      local complement_sequence = transform.complement(sequence)
      
      for i = 1, #sequence do
        local base = sequence:sub(i,i)
        local complement = complement_sequence:sub(i,i)
        local expected_base = transform.complement_base(base)
        assert.are.equal(expected_base, complement,
          string.format("bad %s complement: got %s, expect %s", base, complement, expected_base))
      end
    end)
  end)

  describe("reverse_complement", function()
    it("should match individual reverse and complement operations", function()
      local sequence = random_dna_sequence(20)
      local test_sequences = {sequence, sequence:sub(1, #sequence-1)}
      
      for _, test_sequence in ipairs(test_sequences) do
        local got = transform.reverse_complement(test_sequence)
        local expect = transform.reverse(transform.complement(test_sequence))
        assert.are.equal(expect, got)
      end
    end)
  end)

  describe("complement_base", function()
    it("should correctly complement individual bases", function()
      local letters = {"A", "B", "C", "D", "G", "H", "K", "M", "N", "R", "S", "T", "V", "W", "Y",
                      "a", "b", "c", "d", "g", "h", "k", "m", "n", "r", "s", "t", "v", "w", "y"}
      
      for _, c in ipairs(letters) do
        local got = transform.complement_base(c)
        local got_i = transform.complement_base(got)
        local got_ii = transform.complement_base(got_i)
        assert.are.equal(c, got_i)
        assert.are.equal(got, got_ii)
      end
    end)

    it("should return space for invalid base", function()
      local complement_base = transform.complement_base("!")
      assert.are.equal(" ", complement_base)
    end)
  end)

  describe("complement_base_rna", function()
    it("should correctly complement RNA bases", function()
      local bases = {"A", "B", "C", "D", "G", "H", "K", "M", "N", "R", "S", "U", "V", "W", "Y",
                    "a", "b", "c", "d", "g", "h", "k", "m", "n", "r", "s", "u", "v", "w", "y"}
      
      for _, c in ipairs(bases) do
        local got = transform.complement_base_rna(c)
        local got_i = transform.complement_base_rna(got)
        local got_ii = transform.complement_base_rna(got_i)
        assert.are.equal(c, got_i)
        assert.are.equal(got, got_ii)
      end
    end)
  end)

  describe("complement_rna", function()
    it("should correctly complement RNA sequences", function()
      local sequence = random_rna_sequence(20)
      local complement_sequence = transform.complement_rna(sequence)
      
      for i = 1, #sequence do
        local base = sequence:sub(i,i)
        local complement = complement_sequence:sub(i,i)
        local expected_base = transform.complement_base_rna(base)
        assert.are.equal(expected_base, complement,
          string.format("bad %s complement: got %s, expect %s", base, complement, expected_base))
      end
    end)

    it("should correctly complement known RNA sequence", function()
      local sequence = "ABCDGHKMNRSUVWY"
      local expected = "UVGHCDMKNYSABWR"
      local result = transform.complement_rna(sequence)
      assert.are.equal(expected, result)
    end)
  end)

  describe("reverse_complement_rna", function()
    it("should match individual reverse and complement RNA operations", function()
      local sequence = random_rna_sequence(20)
      local test_sequences = {sequence, sequence:sub(1, #sequence-1)}
      
      for _, test_sequence in ipairs(test_sequences) do
        local got = transform.reverse_complement_rna(test_sequence)
        local expect = transform.reverse(transform.complement_rna(test_sequence))
        assert.are.equal(expect, got)
      end
    end)
  end)
end)
