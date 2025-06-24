local dnadesign = require("dnadesign")
local fragment = dnadesign.fragment
local transform = dnadesign.transform

describe("fragment", function()
    it("should fragment gene sequence", function()
        local gene = "atgaaaaaatttaactggaagaaaatagtcgcgccaattgcaatgctaattattggcttactaggtggtttacttggtgcctttatcctactaacagcagccggggtatcttttaccaatacaacagatactggagtaaaaacggctaagaccgtctacaccaatataacagatacaactaaggctgttaagaaagtacaaaatgccgttgtttctgtcatcaattatcaagaaggttcatcttcagattctctaaatgacctttatggccgtatctttggcggaggggacagttctgattctagccaagaaaattcaaaagattcagatggtctacaggtcgctggtgaaggttctggagtcatctataaaaaagatggcaaagaagcctacatcgtaaccaataaccatgttgtcgatggggctaaaaaacttgaaatcatgctttcggatggttcgaaaattactggtgaacttgttggtaaagacacttactctgacctagcagttgtcaaagtatcttcagataaaataacaactgttgcagaatttgcagactcaaactcccttactgttggtgaaaaagcaattgctatcggtagcccacttggtaccgaatacgccaactcagtaacagaaggaatcgtttctagccttagccgtactataacgatgcaaaacgataatggtgaaactgtatcaacaaacgctatccaaacagatgcagccattaaccctggtaactctggtggtgccctagtcaatattgaaggacaagttatcggtattaattcaagtaaaatttcatcaacgtctgcagtcgctggtagtgctgttgaaggtatggggtttgccattccatcaaacgatgttgttgaaatcatcaatcaattagaaaaagatggtaaagttacacgaccagcactaggaatctcaatagcagatcttaatagcctttctagcagcgcaacttctaaattagatttaccagatgaggtcaaatccggtgttgttgtcggtagtgttcagaaaggtatgccagctgacggtaaacttcaagaatatgatgttatcactgagattgatggtaagaaaatcagctcaaaaactgatattcaaaccaatctttacagccatagtatcggagatactatcaaggtaaccttctatcgtggtaaagataagaaaactgtagatcttaaattaacaaaatctacagaagacatatctgattaa"
        
        local fragments, efficiency, err = fragment.fragment(gene, 90, 110, {})
        assert.is_nil(err)
    end)

    it("should fail to fragment polyA", function()
        local polyA = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
        local fragments, efficiency, err = fragment.fragment(polyA, 40, 80, {})
        assert.is_not_nil(err)
    end)

    it("should check fragment sizes", function()
        local lacZ = "ATGACCATGATTACGCCAAGCTTGCATGCCTGCAGGTCGACTCTAGAGGATCCCCGGGTACCGAGCTCGAATTCACTGGCCGTCGTTTTACAACGTCGTGACTGGGAAAACCCTGGCGTTACCCAACTTAATCGCCTTGCAGCACATCCCCCTTTCGCCAGCTGGCGTAATAGCGAAGAGGCCCGCACCGATCGCCCTTCCCAACAGTTGCGCAGCCTGAATGGCGAATGGCGCCTGATGCGGTATTTTCTCCTTACGCATCTGTGCGGTATTTCACACCGCATATGGTGCACTCTCAGTACAATCTGCTCTGATGCCGCATAG"
        
        -- Test minSize > maxSize
        local fragments, efficiency, err = fragment.fragment(lacZ, 105, 95, {})
        assert.is_not_nil(err)
        
        -- Test minSize < 12
        local fragments2, efficiency2, err2 = fragment.fragment(lacZ, 7, 95, {})
        assert.is_not_nil(err2)
    end)

    it("should handle small fragment size", function()
        local lacZ = "ATGACCATGATTACGCCAAGCTTGCATGCCTGCAGGTCGACTCTAGAGGATCCCCGGGTACCGAGCTCGAATTCACTGGCCGTCGTTTTACAACGTCGTGACTGGGAAAACCCTGGCGTTACCCAACTTAATCGCCTTGCAGCACATCCCCCTTTCGCCAGCTGGCGTAATAGCGAAGAGGCCCGCACCGATCGCCCTTCCCAACAGTTGCGCAGCCTGAATGGCGAATGGCGCCTGATGCGGTATTTTCTCCTTACGCATCTGTGCGGTATTTCACACCGCATATGGTGCACTCTCAGTACAATCTGCTCTGATGCCGCATAG"
        local fragments, efficiency, err = fragment.fragment(lacZ, 12, 30, {})
        assert.is_nil(err)
    end)

    it("should handle long fragments correctly", function()
        local gene = "GGAGGGTCTCAATGCTGGACGATCGCAAATTCAGCGAACAGGAGCTGGTCCGTCGCAACAAATACAAAACGCTGGTCGAGCAAAACAAAGACCCGTACAAGATTACGAACTGGAAACGCAATACCACCCTGCTGAAACTGAATGAGAAATACAAAGACTATAGCAAGGAGGACCTGTTGAACCTGAATCAAGAACTGGTCGTTGTTGCAGGTCGTATCAAACTGTATCGTGAAGCCGGTAAAAAAGCTGCCTTTGTGAACATTGATGATCAAGACTCCTCTATTCAGTTGTACGTGCGCCTGGATGAGATCGGTGATCAGAGCTTCGAGGATTTCCGCAATTTCGACCTGGGTGACATCATTGGTGTTAAAGGTATCATGATGCGCACCGACCACGGCGAGTTGAGCATCCGTTGTAAGGAAGTCGTGCTGCTGAGCAAGGCCCTGCGTCCGCTGCCGGATAAACACGCGGGCATTCAGGATATTGAGGAAAAGTACCGCCGTCGCTATGTGGACCTGATTATGAATCACGACGTGCGCAAGACGTTCCAGGCGCGTACCAAGATCATTCGTACCTTGCAAAACTTTCTGGATAATAAGGGTTACATGGAGGTCGAAACCCCGATCCTGCATAGCCTGAAGGGTGGCGCGAGCGCGAAACCGTTTATTACCCACTACAATGTGCTGAATACGGATGTGTATCTGCGTATCGCGACCGAGCTGCACCTGAAACGCCTGATTGTTGGCGGTTTCGAGGGTGTGTATGAGATCGGTCGCATCTTTCGCAATGAAGGTATGTCCACGCGTCACAATCCGGAATTCACGTCTATCGAACTGTATGTCGCCTATGAGGACATGTTCTTTTTGATGGATCTGACCGAAGAGATTTTTCGCGTTTGTAATGCCGCAGTCAACAGCTCCAGCATCATTGAGTATAACAACGTGAAAATTGACCTGAGCAAGCCGTTTAAGCGCCTGCATATGGTTGACGGTATTAAACAGGTGACCGGCGTCGACTTCTGGCAGGAGATGACGGTCCAACAGGCTCTGGAGCTGGCCAAAAAGCATAAAGTGCACGTTGAAAAACATCAAGAGTCTGTTGGTCACATTATCAATTTGTTCTATGAGGAGTTCGTGGAGTCCACGATTGTTGAGCCGACGTTCGTGTACGGTCACCCGAAGGAAATCTCTCCGCTGGCTAAGAGCAATCCGTCTGACCCGCGTTTCACGGACCGTTTCGAGCTGTTCATTCTGGGTCGTGAGTATGCGAATGCGTTTAGCGAGCTGAATGACCCGATTGACCAGTACGAACGCTTCAAGGCTCAGATTGAGGAGGAAAGCAAGGGCAACGATGAAGCCAACGACATGGACATTGATTTCATCGAGGCTCTGGAACACGCCATGCCGCCGACCGCGGGTATTGGTATCGGCATTGATCGCTTGGTTATGCTGCTGACGAATAGCGAATCCATCAAAGACGTGCTGTTGTTCCCGCAAATGAAGCCGCGCGAATGAAGAGCTTAGAGACCCGCT"
        local fragments, efficiency, err = fragment.fragment(gene, 79, 94, {})
        assert.is_nil(err)
        
        -- Check fragment lengths
        for _, frag in ipairs(fragments) do
            assert.is_true(#frag <= 94)
        end
    end)

    it("should check long regression behavior", function()
        local overhangs = {"AGAC"}
        local new_overhangs, _ = fragment.next_overhangs(overhangs)
        
        -- Check that GTCT is not in new_overhangs
        local found_GTCT = false
        for _, overhang in ipairs(new_overhangs) do
            if overhang == "GTCT" then
                found_GTCT = true
                break
            end
        end
        assert.is_false(found_GTCT)
    end)

    it("should match NEB ligase fidelity viewer efficiency", function()
        local overhangs = {"CGAG", "GTCT", "TACT", "AATG", "ATCC", "CGCT", "AAAA", "AAGT", "ATAG", "ATTA", "ACAA", "ACGC", "TATC", "TAGA", "TTAC", "TTCA", "TGTG", "TCGG", "TCCC", "GAAG", "GTGC", "GCCG", "CAGG", "TACG"}
        local efficiency = fragment.set_efficiency(overhangs)
        assert.is_true(efficiency <= 1 and efficiency >= 0.965)
    end)

	it("should fragment with overhangs", function()
        local default_overhangs = {"CGAG", "GTCT", "GGGG", "AAAA", "AACT", "AATG", "ATCC", "CGCT", "TTCT", "AAGC", "ATAG", "ATTA", "ATGT", "ACTC", "ACGA", "TATC", "TAGG", "TACA", "TTAC", "TTGA", "TGGA", "GAAG", "GACC", "GCCG", "TCTG", "GTTG", "GTGC", "TGCC", "CTGG", "TAAA", "TGAG", "AAGA", "AGGT", "TTCG", "ACTA", "TTAG", "TCTC", "TCGG", "ATAA", "ATCA", "TTGC", "CACG", "AATA", "ACAA", "ATGG", "TATG", "AAAT", "TCAC"}
        local gene = "atgaaaaaatttaactggaagaaaatagtcgcgccaattgcaatgctaattattggcttactaggtggtttacttggtgcctttatcctactaacagcagccggggtatcttttaccaatacaacagatactggagtaaaaacggctaagaccgtctacaccaatataacagatacaactaaggctgttaagaaagtacaaaatgccgttgtttctgtcatcaattatcaagaaggttcatcttcagattctctaaatgacctttatggccgtatctttggcggaggggacagttctgattctagccaagaaaattcaaaagattcagatggtctacaggtcgctggtgaaggttctggagtcatctataaaaaagatggcaaagaagcctacatcgtaaccaataaccatgttgtcgatggggctaaaaaacttgaaatcatgctttcggatggttcgaaaattactggtgaacttgttggtaaagacacttactctgacctagcagttgtcaaagtatcttcagataaaataacaactgttgcagaatttgcagactcaaactcccttactgttggtgaaaaagcaattgctatcggtagcccacttggtaccgaatacgccaactcagtaacagaaggaatcgtttctagccttagccgtactataacgatgcaaaacgataatggtgaaactgtatcaacaaacgctatccaaacagatgcagccattaaccctggtaactctggtggtgccctagtcaatattgaaggacaagttatcggtattaattcaagtaaaatttcatcaacgtctgcagtcgctggtagtgctgttgaaggtatggggtttgccattccatcaaacgatgttgttgaaatcatcaatcaattagaaaaagatggtaaagttacacgaccagcactaggaatctcaatagcagatcttaatagcctttctagcagcgcaacttctaaattagatttaccagatgaggtcaaatccggtgttgttgtcggtagtgttcagaaaggtatgccagctgacggtaaacttcaagaatatgatgttatcactgagattgatggtaagaaaatcagctcaaaaactgatattcaaaccaatctttacagccatagtatcggagatactatcaaggtaaccttctatcgtggtaaagataagaaaactgtagatcttaaattaacaaaatctacagaagacatatctgattaa"
        
        local fragments, efficiency, err = fragment.fragment_with_overhangs(gene, 90, 110, {}, default_overhangs)
        assert.is_nil(err)
    end)

    it("should handle recursive fragmentation", function()
        local default_overhangs = {"GGGG", "AAAA", "AACT", "AATG", "ATCC", "CGCT", "TTCT", "AAGC", "ATAG", "ATTA", "ATGT", "ACTC", "ACGA", "TATC", "TAGG", "TACA", "TTAC", "TTGA", "TGGA", "GAAG", "GACC", "GCCG", "TCTG", "GTTG", "GTGC", "TGCC", "CTGG", "TAAA", "TGAG", "AAGA", "AGGT", "TTCG", "ACTA", "TTAG", "TCTC", "TCGG", "ATAA", "ATCA", "TTGC", "CACG", "AATA", "ACAA", "ATGG", "TATG", "AAAT", "TCAC"}
        local exclude_overhangs = {"CGAG", "GTCT"} -- These are the recursive BsaI definitions
        local gene = "ATGACCATGATTACGCCAAGCTTGCATGCCTGCAGGTCGACTCTAGAGGATCCCCGGGTACCGAGCTCGAATTCACTGGCCGTCGTTTTACAACGTCGTGACTGGGAAAACCCTGGCGTTACCCAACTTAATCGCCTTGCAGCACATCCCCCTTTCGCCAGCTGGCGTAATAGCGAAGAGGCCCGCACCGATCGCCCTTCCCAACAGTTGCGCAGCCTGAATGGCGAATGGCGCCTGATGCGGTATTTTCTCCTTACGCATCTGTGCGGTATTTCACACCGCATATGGTGCACTCTCAGTACAATCTGCTCTGATGCCGCATAG"
        local max_oligo_len = 174 -- for Agilent oligo pools
        local assembly_pattern = {5, 4, 4, 5} -- seems reasonable enough
        
        local result, err = fragment.recursive_fragment(
            gene,
            max_oligo_len,
            assembly_pattern,
            exclude_overhangs,
            default_overhangs,
            "GTCTCT",
            "CGAG"
        )
        assert.is_nil(err)

        -- Test specific fragment structure
        local expected_fragments = {
            "ATGACCATGATTACGCCAAGCTTGCATGCCTGCAGGTCGACTCTAGAGGATCCCCGGGTACCGAGCTCGAATTCACTGGCCGTCGTTTTACAACGTCGTGACTGGGAAAACCCTGGCGTTACCCAACTTAATCGCCTTGCAGCACATCCCCCTTTCGCCAG",
            "CCAGCTGGCGTAATAGCGAAGAGGCCCGCACCGATCGCCCTTCCCAACAGTTGCGCAGCCTGAATGGCGAATGGCGCCTGATGCGGTATTTTCTCCTTACGCATCTGTGCGGTATTTCACACCGCATATGGTGCACTCTCAGTACAATCTGCTCTGATGCCGCATAG"
        }

        -- Check that we have the same number of fragments
        assert.are.equal(#expected_fragments, #result.fragments)

        -- Check each fragment matches expected
        for i, expected_fragment in ipairs(expected_fragments) do
            assert.are.equal(expected_fragment, result.fragments[i])
        end
    end)
end)

