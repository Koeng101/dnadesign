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

    it("should fragment with a break inside a window", function()
        -- A plasmid where a kanR marker spans the window [1962, 2485]. The break
        -- forces a fragment junction into that window so the marker is split
        -- across two assemblies (see the FASTA header fragment_between_1962_and_2485).
        local plasmid = "TTTCGGGAGCGGATTATACACAAGCTTctagcataaccccttggggcctctaaacgggtcttgaggggttttttgGTCGTGGTTTGTCTGGTCAACCACCGCGGGCTCAGTGGTGTACGGTACAAACCCCGACTTTTAAATTTatgacacacattaacaaatttcgtgaggagtctccagaagaatgccattaaCGGGTATGCTTGCCTACTAATTAGGGATAACAGGGTAATCTCTCTTAAGGTAGCTTGAGGAGGTTTCTCTGTAAATAATAACTATAACGGTCCTAAGGTAGCGAggattttaacattttgcgttgttccaaaagttatcaacagcctagaacgtcataggaagcgattacagacactttatagctatcagcatgggaacataagggcaggatgaaatatgggtcttaaaacgcaaatggtgaggttttagaggtattttttgaagatgattaaggcggtttgtttttaaaatttttggcggctctcaggctgcttacattttaaccagttcagtgaaaagttctttttcagcaaatttctgtttagcaccatagctaaaacttgcgtggaacatattaagaattgaccgaaatgacacaatctcaattatattttttttgaaaagttttctttatcaaatattttaaatcattgatttatatataagtatgcattcattttaataattaatctttatttaacaatgatttatctatattcaattgtttaattattcttactaatattatctctatatcaatattttttatttaaaaacatatgtttagtagtgcttttgattaaagtaccagagggagggagcagagctgaatgggaaatactcaccctagagcgattcttaaaaatcaccctaaagtattcccattcgatgtaccgtcggtcggtcgctttcgcatcagggatgacatcactgtatcaagctgccactgttatgattacgattgatagcaccgcctgaacacgctcataaccgccaattaaatgactactcattgcgtccgctactctgttcagttcccctagtaatagcgtttttccgatgtgtgctagcgtcactgtacctcatcacccacacatggacagattgagttacagacattggctaaatttttggggtctatgcttgacaaagcgagttaaaaccttagaatttaagacaggtacattaagcccctgtggtgaaatcattagcggtcgtcaaaccttaatgctttcgattgccgatatgtcagtatcgaagatctgtaccccataaattacggtaaagccccaagcaattgcaaggggcttttatcttttttaacaaaaaaaatttataaatcaggattttataacactaaataatccaaagtacatgagtaagtcatgaccactcctgcgatgtgtgtagcactctgagtatccgtattatatcagtgatgtcatatacaaccacataatgcggatgtatcactaactccctagtatctttctgtctgcctgtgcgccccatcagtcgattatcaacaagtaaatctgtcttttcttcaattaaatcatcaatttcaatagctgctatagggttgcgttcttcaagataatcatagatatttctacgatcttggcTCTGCCCGGGattctcaccaataaaaaacgcccggcggcaaccgagcgttctgaacaaatccagatggagttctgaggtcattactggatctatcaacgggagtccaagcgagctcggtactaaaacaattcatccagtaaaatataatattttattttctcccaatcaggcttgatccccagtaagtcaaaaaatagctcgacatactgttcttccccgatatcctccctgatcgaccggacgcagaaggcaatgtcataccacttgtccgccctgccgcttctcccaagatcaataaagccacttactttgccatctttcacaaagatgttgctgtctcccaggtcgccgtgggaaaagacaagttcctcttcgggcttttccgtctttaaaaaatcatacagctcgcgcggatctttaaatggagtatcttcttcccagttttcgcaatccacatcggccagatcgttattcagtaagtaatccaattcggctaagcggccgtctaagctattcgtatagggacaatccgatatgtcgatggagtgaaagagcctgatgcactccgcatacagctcgatagtcttttcagggctttgttcatcttcatacccttccgagcaaaggacgccatcggcctcactcatgagcagattgctccagccatcatgccgttcaaagtgcaggacctttggaacaggcagctttccttccagccatagcatcatgtccttttcccgttccacatcataggtggtccctttataccggctgtccgtcatttttaaatataggatttcattttctcccaccagcttatataccttagcaggagacattccttccgtatcttttacgcagcggtattcttcgatcagttttttcaattccggtgatattctcattttagccatttattatttccttcctcttttctacagtatttaaagataccccaagaagctaattataacaagacgaactccaattcactgttccttgcattctaaaaccttaaatacagaaaacagccttttcaaagttgttttcaaagttggcgtataacatagtatcgacggagccgattttgaaaccacaattatgatagaatttgacgtccttttccgctgcataaccctgcttcggggtcattatagcgattttttcggtatatccatcctttttcgcacgatatacaggattttgccaaagggttcgtgtagactttccttggtgtatccaacggcgtcagccgggcaggataggtgaagtaggcccacccgcgagcgggtgttccttcttcactgtcccttattcgcacctggcggtgctcaacgggaatcctgctctgcgaggctggccgtaTTGACAGACAATCCGTAGGCACAATTTTCGAAAAAACCCGCTTCGGCGGGTTTTTTTATAGCTAAAAATGTTCCAGCGCTGGCACGCAACCTCTCATGCGCTACTTATCACGCCGCGCCAATTTATTACCGCTATGGCCAATTGATCGGCCGGCTTGTCGACGACGGCGGACTCCGTCGTCAGGATCATCCGGGCGAATTCCGTGTTATCCAGTCCCAGAA"
        local break_start = 1962
        local break_end = 2485

        local before, after, efficiency, err = fragment.fragment_with_break(plasmid, break_start, break_end, 800, 1000, {})
        assert.is_nil(err)
        assert.is_true(#before > 0)
        assert.is_true(#after > 0)

        -- Helper: reassemble a fragment list, dropping the 4bp overhang each
        -- consecutive fragment shares with the previous one.
        local function reassemble(frags)
            local seq = frags[1]
            for i = 2, #frags do
                seq = seq .. frags[i]:sub(5)
            end
            return seq
        end

        -- The break overhang is the junction shared by the last before-fragment
        -- and the first after-fragment.
        local break_overhang = before[#before]:sub(-4)
        assert.are.equal(break_overhang, after[1]:sub(1, 4))

        -- That junction must sit within the requested window. The before region
        -- is plasmid[1..break_position], so its length is the break position.
        local break_position = #reassemble(before)
        assert.is_true(break_position >= break_start)
        assert.is_true(break_position <= break_end)

        -- No two junction overhangs may be identical or reverse complements,
        -- otherwise the two halves would mis-assemble when combined.
        local all_frags = {}
        for _, f in ipairs(before) do table.insert(all_frags, f) end
        for _, f in ipairs(after) do table.insert(all_frags, f) end
        local overhangs = {before[1]:sub(1, 4)}
        for _, f in ipairs(all_frags) do table.insert(overhangs, f:sub(-4)) end
        for i = 1, #overhangs do
            for j = i + 1, #overhangs do
                assert.are_not.equal(overhangs[i], overhangs[j])
                assert.are_not.equal(transform.reverse_complement(overhangs[i]), overhangs[j])
            end
        end

        -- The combined overhang set should assemble efficiently.
        assert.is_true(efficiency > 0.85)

        -- Round trip: concatenating all fragments (before then after) while
        -- dropping each shared overhang must reproduce the original sequence.
        assert.are.equal(string.upper(plasmid), reassemble(all_frags))
    end)
end)

