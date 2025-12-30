dnadesign = require("dnadesign")
local align = dnadesign.align

describe("Sequence alignment", function()
  describe("Needleman-Wunsch algorithm", function()
    local scoring

    setup(function()
      -- Create DNA/RNA substitution matrix
      local substitution_matrix = {
        data = {
          A = {A = 1, C = -1, G = -1, T = -1, U = -1},
          C = {A = -1, C = 1, G = -1, T = -1, U = -1},
          G = {A = -1, C = -1, G = 1, T = -1, U = -1},
          T = {A = -1, C = -1, G = -1, T = 1, U = -1},
          U = {A = -1, C = -1, G = -1, T = -1, U = 1}
        }
      }
      scoring = align.new_scoring(substitution_matrix, -1)
    end)

    it("should align different sequences correctly", function()
      local score, align_a, align_b = align.needleman_wunsch("GATTACA", "GCATGCU", scoring)
      assert.equals(0, score)
      -- Note: We don't check exact alignments as there can be multiple valid alignments with same score
    end)

    it("should align identical sequences perfectly", function()
      local score, align_c, align_d = align.needleman_wunsch("GATTACA", "GATTACA", scoring)
      assert.equals(7, score)
      assert.equals("GATTACA", align_c)
      assert.equals("GATTACA", align_d)
    end)

    it("should align with short sequence", function()
      local score, align_e, align_f = align.needleman_wunsch("GATTACA", "GAT", scoring)
      assert.equals(-1, score)
    end)

    it("should handle empty string", function()
      local score, align_g, align_h = align.needleman_wunsch("", "GAT", scoring)
      assert.equals(-3, score)
    end)

    it("should handle empty strings", function()
      local score, align_i, align_j = align.needleman_wunsch("", "", scoring)
      assert.equals(0, score)
      assert.equals("", align_i)
      assert.equals("", align_j)
    end)

    it("should align single different characters", function()
      local score, align_k, align_l = align.needleman_wunsch("G", "A", scoring)
      assert.equals(-1, score)
    end)

    it("should align single identical characters", function()
      local score, align_m, align_n = align.needleman_wunsch("G", "G", scoring)
      assert.equals(1, score)
    end)

    it("should align single char with longer sequence", function()
      local score, align_o, align_p = align.needleman_wunsch("G", "GATTACA", scoring)
      assert.equals(-5, score)
    end)
  end)

  describe("Smith-Waterman algorithm", function()
    local scoring

    setup(function()
      -- Create DNA substitution matrix with gap character
      local substitution_matrix = {
        data = {
          ["-"] = {["-"] = 0, A = 0, C = 0, G = 0, T = 0},
          A = {["-"] = 0, A = 3, C = -3, G = -3, T = -3},
          C = {["-"] = 0, A = -3, C = 3, G = -3, T = -3},
          G = {["-"] = 0, A = -3, C = -3, G = 3, T = -3},
          T = {["-"] = 0, A = -3, C = -3, G = -3, T = 3}
        }
      }
      scoring = align.new_scoring(substitution_matrix, -2)
    end)

    it("should align Wikipedia example correctly", function()
      local score, align_a, align_b = align.smith_waterman("TGTTACGG", "GGTTGACTA", scoring)
      assert.equals(13, score)
      assert.equals("GTT-AC", align_a)
      assert.equals("GTTGAC", align_b)
    end)

    it("should find local alignment in similar sequences", function()
      local score, align_c, align_d = align.smith_waterman("ACACACTA", "AGCACACA", scoring)
      assert.equals(17, score)
      assert.equals("A-CACACTA", align_c)
      assert.equals("AGCACAC-A", align_d)
    end)

    it("should handle empty string", function()
      local score, align_e, align_f = align.smith_waterman("", "GAT", scoring)
      assert.equals(0, score)
      assert.equals("", align_e)
      assert.equals("", align_f)
    end)

    it("should handle empty strings", function()
      local score, align_g, align_h = align.smith_waterman("", "", scoring)
      assert.equals(0, score)
      assert.equals("", align_g)
      assert.equals("", align_h)
    end)

    it("should handle single different characters", function()
      local score, align_i, align_j = align.smith_waterman("G", "A", scoring)
      assert.equals(0, score)
      assert.equals("", align_i)
      assert.equals("", align_j)
    end)

    it("should handle single identical characters", function()
      local score, align_k, align_l = align.smith_waterman("G", "G", scoring)
      assert.equals(3, score)
      assert.equals("G", align_k)
      assert.equals("G", align_l)
    end)

    it("should find local alignment with single char vs long sequence", function()
      local score, align_m, align_n = align.smith_waterman("G", "GATTACA", scoring)
      assert.equals(3, score)
      assert.equals("G", align_m)
      assert.equals("G", align_n)
    end)
  end)

  describe("align_many", function()
	  local scoring
	
	  setup(function()
	    local substitution_matrix = {
	      data = {
	        ["-"] = {["-"] = 0, A = 0, C = 0, G = 0, T = 0},
	        A = {["-"] = 0, A = 3, C = -3, G = -3, T = -3},
	        C = {["-"] = 0, A = -3, C = 3, G = -3, T = -3},
	        G = {["-"] = 0, A = -3, C = -3, G = 3, T = -3},
	        T = {["-"] = 0, A = -3, C = -3, G = -3, T = 3}
	      }
	    }
	    scoring = align.new_scoring(substitution_matrix, -2)
	  end)
	
	  it("should return top N results sorted by score", function()
	    local target = "GATTACA"
	    local candidates = {"GATTACA", "GCATGCT", "ACGT", "TTTTTTT"}
	    
	    local results = align.align_many(align.smith_waterman, target, candidates, scoring, 3)
	    
	    -- Should return exactly 3 results
	    assert.equals(3, #results)
	    
	    -- Results should be sorted by score (descending)
	    assert.is_true(results[1][1] >= results[2][1])
	    assert.is_true(results[2][1] >= results[3][1])
	    
	    -- Best match should be identical sequence
	    assert.equals(21, results[1][1]) -- Perfect match: 7 chars * 3 points

		-- Best match should be 1
		assert.equals(1, results[1][4])
	  end)
	
	  it("should handle ntop larger than number of candidates", function()
	    local target = "GAT"
	    local candidates = {"GAT", "GCT"}
	    
	    local results = align.align_many(align.smith_waterman, target, candidates, scoring, 10)
	    
	    -- Should return only 2 results even though ntop=10
	    assert.equals(2, #results)
	  end)
	
	  it("should handle ntop of 0", function()
	    local target = "GAT"
	    local candidates = {"GAT", "GCT", "ACG"}
	    
	    local results = align.align_many(align.smith_waterman, target, candidates, scoring, 0)
	    
	    assert.equals(0, #results)
	  end)
	
	  it("should handle negative ntop", function()
	    local target = "GAT"
	    local candidates = {"GAT", "GCT"}
	    
	    local results = align.align_many(align.smith_waterman, target, candidates, scoring, -5)
	    
	    assert.equals(0, #results)
	  end)
	end)
end)
