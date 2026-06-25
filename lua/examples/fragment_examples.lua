-- examples/fragment_examples.lua
local dnadesign = require("dnadesign")
local fragment = dnadesign.fragment

describe("Fragment Examples", function()
    it("demonstrates basic fragmentation", function()
        local lacZ = "ATGACCATGATTACGCCAAGCTTGCATGCCTGCAGGTCGACTCTAGAGGATCCCCGGGTACCGAGCTCGAATTCACTGGCCGTCGTTTTACAACGTCGTGACTGGGAAAACCCTGGCGTTACCCAACTTAATCGCCTTGCAGCACATCCCCCTTTCGCCAGCTGGCGTAATAGCGAAGAGGCCCGCACCGATCGCCCTTCCCAACAGTTGCGCAGCCTGAATGGCGAATGGCGCCTGATGCGGTATTTTCTCCTTACGCATCTGTGCGGTATTTCACACCGCATATGGTGCACTCTCAGTACAATCTGCTCTGATGCCGCATAG"
        local fragments, _, err = fragment.fragment(lacZ, 95, 105, {"AAAA"})
        assert.is_nil(err)

        -- Expected fragments should match Go output
        local expected_fragments = {
            "ATGACCATGATTACGCCAAGCTTGCATGCCTGCAGGTCGACTCTAGAGGATCCCCGGGTACCGAGCTCGAATTCACTGGCCGTCGTTTTACAACGTCGTGACTGG",
            "CTGGGAAAACCCTGGCGTTACCCAACTTAATCGCCTTGCAGCACATCCCCCTTTCGCCAGCTGGCGTAATAGCGAAGAGGCCCGCACCGATCGCCCTTCCCAAC",
            "CAACAGTTGCGCAGCCTGAATGGCGAATGGCGCCTGATGCGGTATTTTCTCCTTACGCATC",
            "CATCTGTGCGGTATTTCACACCGCATATGGTGCACTCTCAGTACAATCTGCTCTGATGCCGCATAG"
        }

        assert.are.same(expected_fragments, fragments)
    end)

    it("demonstrates next overhang generation", function()
        local primer_overhangs = {"ATAA"}
        table.insert(primer_overhangs, fragment.next_overhang(primer_overhangs))
        table.insert(primer_overhangs, fragment.next_overhang(primer_overhangs))
        table.insert(primer_overhangs, fragment.next_overhang(primer_overhangs))

        local expected_overhangs = {"ATAA", "AAAT", "AATA", "AAGA"}
        assert.are.same(expected_overhangs, primer_overhangs)
    end)

    it("demonstrates fragment efficiency calculation", function()
        local lacZ = "ATGACCATGATTACGCCAAGCTTGCATGCCTGCAGGTCGACTCTAGAGGATCCCCGGGTACCGAGCTCGAATTCACTGGCCGTCGTTTTACAACGTCGTGACTGGGAAAACCCTGGCGTTACCCAACTTAATCGCCTTGCAGCACATCCCCCTTTCGCCAGCTGGCGTAATAGCGAAGAGGCCCGCACCGATCGCCCTTCCCAACAGTTGCGCAGCCTGAATGGCGAATGGCGCCTGATGCGGTATTTTCTCCTTACGCATCTGTGCGGTATTTCACACCGCATATGGTGCACTCTCAGTACAATCTGCTCTGATGCCGCATAG"
        local fragments, efficiency, err = fragment.fragment(lacZ, 95, 105, {})
        assert.is_nil(err)

        -- Check second fragment and efficiency matches Go output
        local expected_fragment = "CTGGGAAAACCCTGGCGTTACCCAACTTAATCGCCTTGCAGCACATCCCCCTTTCGCCAGCTGGCGTAATAGCGAAGAGGCCCGCACCGATCGCCCTTCCCAACA"
        local expected_efficiency = 1.0

        assert.are.equal(expected_fragment, fragments[2])
        assert.are.equal(expected_efficiency, efficiency)
    end)

    it("demonstrates recursive fragmentation", function()
        local gene = "ATGACCATGATTACGCCAAGCTTGCATGCCTGCAGGTCGACTCTAGAGGATCCCCGGGTACCGAGCTCGAATTCACTGGCCGTCGTTTTACAACGTCGTGACTGGGAAAACCCTGGCGTTACCCAACTTAATCGCCTTGCAGCACATCCCCCTTTCGCCAGCTGGCGTAATAGCGAAGAGGCCCGCACCGATCGCCCTTCCCAACAGTTGCGCAGCCTGAATGGCGAATGGCGCCTGATGCGGTATTTTCTCCTTACGCATCTGTGCGGTATTTCACACCGCATATGGTGCACTCTCAGTACAATCTGCTCTGATGCCGCATAG"
        local default_overhangs = {"GGGG", "AAAA", "AACT", "AATG", "ATCC"}  -- shortened for example
        local exclude_overhangs = {"CGAG", "GTCT"}
        local max_oligo_len = 174
        local assembly_pattern = {5, 4, 4, 5}

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
        assert.is_not_nil(result)
        assert.is_true(#result.fragments > 0)
    end)

    it("demonstrates fragmenting with a break to interrupt a gene", function()
        -- For full plasmid synthesis, the first assembly is ampR-selected while a
        -- kanR marker is deliberately interrupted so it is non-functional. Here
        -- the kanR marker spans the window [1962, 2485]; fragment_with_break
        -- forces one fragment junction into that window, splitting the marker
        -- across two assemblies. The "before" and "after" fragments only come
        -- together in a later GoldenGate, reconstituting full-length kanR.
        local plasmid = "TTTCGGGAGCGGATTATACACAAGCTTctagcataaccccttggggcctctaaacgggtcttgaggggttttttgGTCGTGGTTTGTCTGGTCAACCACCGCGGGCTCAGTGGTGTACGGTACAAACCCCGACTTTTAAATTTatgacacacattaacaaatttcgtgaggagtctccagaagaatgccattaaCGGGTATGCTTGCCTACTAATTAGGGATAACAGGGTAATCTCTCTTAAGGTAGCTTGAGGAGGTTTCTCTGTAAATAATAACTATAACGGTCCTAAGGTAGCGAggattttaacattttgcgttgttccaaaagttatcaacagcctagaacgtcataggaagcgattacagacactttatagctatcagcatgggaacataagggcaggatgaaatatgggtcttaaaacgcaaatggtgaggttttagaggtattttttgaagatgattaaggcggtttgtttttaaaatttttggcggctctcaggctgcttacattttaaccagttcagtgaaaagttctttttcagcaaatttctgtttagcaccatagctaaaacttgcgtggaacatattaagaattgaccgaaatgacacaatctcaattatattttttttgaaaagttttctttatcaaatattttaaatcattgatttatatataagtatgcattcattttaataattaatctttatttaacaatgatttatctatattcaattgtttaattattcttactaatattatctctatatcaatattttttatttaaaaacatatgtttagtagtgcttttgattaaagtaccagagggagggagcagagctgaatgggaaatactcaccctagagcgattcttaaaaatcaccctaaagtattcccattcgatgtaccgtcggtcggtcgctttcgcatcagggatgacatcactgtatcaagctgccactgttatgattacgattgatagcaccgcctgaacacgctcataaccgccaattaaatgactactcattgcgtccgctactctgttcagttcccctagtaatagcgtttttccgatgtgtgctagcgtcactgtacctcatcacccacacatggacagattgagttacagacattggctaaatttttggggtctatgcttgacaaagcgagttaaaaccttagaatttaagacaggtacattaagcccctgtggtgaaatcattagcggtcgtcaaaccttaatgctttcgattgccgatatgtcagtatcgaagatctgtaccccataaattacggtaaagccccaagcaattgcaaggggcttttatcttttttaacaaaaaaaatttataaatcaggattttataacactaaataatccaaagtacatgagtaagtcatgaccactcctgcgatgtgtgtagcactctgagtatccgtattatatcagtgatgtcatatacaaccacataatgcggatgtatcactaactccctagtatctttctgtctgcctgtgcgccccatcagtcgattatcaacaagtaaatctgtcttttcttcaattaaatcatcaatttcaatagctgctatagggttgcgttcttcaagataatcatagatatttctacgatcttggcTCTGCCCGGGattctcaccaataaaaaacgcccggcggcaaccgagcgttctgaacaaatccagatggagttctgaggtcattactggatctatcaacgggagtccaagcgagctcggtactaaaacaattcatccagtaaaatataatattttattttctcccaatcaggcttgatccccagtaagtcaaaaaatagctcgacatactgttcttccccgatatcctccctgatcgaccggacgcagaaggcaatgtcataccacttgtccgccctgccgcttctcccaagatcaataaagccacttactttgccatctttcacaaagatgttgctgtctcccaggtcgccgtgggaaaagacaagttcctcttcgggcttttccgtctttaaaaaatcatacagctcgcgcggatctttaaatggagtatcttcttcccagttttcgcaatccacatcggccagatcgttattcagtaagtaatccaattcggctaagcggccgtctaagctattcgtatagggacaatccgatatgtcgatggagtgaaagagcctgatgcactccgcatacagctcgatagtcttttcagggctttgttcatcttcatacccttccgagcaaaggacgccatcggcctcactcatgagcagattgctccagccatcatgccgttcaaagtgcaggacctttggaacaggcagctttccttccagccatagcatcatgtccttttcccgttccacatcataggtggtccctttataccggctgtccgtcatttttaaatataggatttcattttctcccaccagcttatataccttagcaggagacattccttccgtatcttttacgcagcggtattcttcgatcagttttttcaattccggtgatattctcattttagccatttattatttccttcctcttttctacagtatttaaagataccccaagaagctaattataacaagacgaactccaattcactgttccttgcattctaaaaccttaaatacagaaaacagccttttcaaagttgttttcaaagttggcgtataacatagtatcgacggagccgattttgaaaccacaattatgatagaatttgacgtccttttccgctgcataaccctgcttcggggtcattatagcgattttttcggtatatccatcctttttcgcacgatatacaggattttgccaaagggttcgtgtagactttccttggtgtatccaacggcgtcagccgggcaggataggtgaagtaggcccacccgcgagcgggtgttccttcttcactgtcccttattcgcacctggcggtgctcaacgggaatcctgctctgcgaggctggccgtaTTGACAGACAATCCGTAGGCACAATTTTCGAAAAAACCCGCTTCGGCGGGTTTTTTTATAGCTAAAAATGTTCCAGCGCTGGCACGCAACCTCTCATGCGCTACTTATCACGCCGCGCCAATTTATTACCGCTATGGCCAATTGATCGGCCGGCTTGTCGACGACGGCGGACTCCGTCGTCAGGATCATCCGGGCGAATTCCGTGTTATCCAGTCCCAGAA"

        -- Split the plasmid so a junction lands between positions 1962 and 2485,
        -- fragmenting each side into 95-105bp pieces.
        local before, after, efficiency, err = fragment.fragment_with_break(plasmid, 1962, 2485, 800, 1000, {})
        assert.is_nil(err)

        -- "before" holds the fragments up to and including the break overhang;
        -- "after" holds the fragments from the break overhang onward. The two
        -- lists share the break overhang as their junction.
        assert.is_true(#before > 0)
        assert.is_true(#after > 0)
        assert.are.equal(before[#before]:sub(-4), after[1]:sub(1, 4))

        -- The combined design assembles efficiently.
        assert.is_true(efficiency > 0.85)
    end)

    it("demonstrates recursing down from ~1kbp chunks to oligo-sized pieces", function()
        -- Two-level synthesis. The top level breaks the plasmid into ~1kbp chunks
        -- (800-1000bp) with a break in the kanR window, just like above. Each of
        -- those chunks is itself assembled from smaller 151-176bp pieces, so we
        -- recurse down by fragmenting each chunk again.
        local plasmid = "TTTCGGGAGCGGATTATACACAAGCTTctagcataaccccttggggcctctaaacgggtcttgaggggttttttgGTCGTGGTTTGTCTGGTCAACCACCGCGGGCTCAGTGGTGTACGGTACAAACCCCGACTTTTAAATTTatgacacacattaacaaatttcgtgaggagtctccagaagaatgccattaaCGGGTATGCTTGCCTACTAATTAGGGATAACAGGGTAATCTCTCTTAAGGTAGCTTGAGGAGGTTTCTCTGTAAATAATAACTATAACGGTCCTAAGGTAGCGAggattttaacattttgcgttgttccaaaagttatcaacagcctagaacgtcataggaagcgattacagacactttatagctatcagcatgggaacataagggcaggatgaaatatgggtcttaaaacgcaaatggtgaggttttagaggtattttttgaagatgattaaggcggtttgtttttaaaatttttggcggctctcaggctgcttacattttaaccagttcagtgaaaagttctttttcagcaaatttctgtttagcaccatagctaaaacttgcgtggaacatattaagaattgaccgaaatgacacaatctcaattatattttttttgaaaagttttctttatcaaatattttaaatcattgatttatatataagtatgcattcattttaataattaatctttatttaacaatgatttatctatattcaattgtttaattattcttactaatattatctctatatcaatattttttatttaaaaacatatgtttagtagtgcttttgattaaagtaccagagggagggagcagagctgaatgggaaatactcaccctagagcgattcttaaaaatcaccctaaagtattcccattcgatgtaccgtcggtcggtcgctttcgcatcagggatgacatcactgtatcaagctgccactgttatgattacgattgatagcaccgcctgaacacgctcataaccgccaattaaatgactactcattgcgtccgctactctgttcagttcccctagtaatagcgtttttccgatgtgtgctagcgtcactgtacctcatcacccacacatggacagattgagttacagacattggctaaatttttggggtctatgcttgacaaagcgagttaaaaccttagaatttaagacaggtacattaagcccctgtggtgaaatcattagcggtcgtcaaaccttaatgctttcgattgccgatatgtcagtatcgaagatctgtaccccataaattacggtaaagccccaagcaattgcaaggggcttttatcttttttaacaaaaaaaatttataaatcaggattttataacactaaataatccaaagtacatgagtaagtcatgaccactcctgcgatgtgtgtagcactctgagtatccgtattatatcagtgatgtcatatacaaccacataatgcggatgtatcactaactccctagtatctttctgtctgcctgtgcgccccatcagtcgattatcaacaagtaaatctgtcttttcttcaattaaatcatcaatttcaatagctgctatagggttgcgttcttcaagataatcatagatatttctacgatcttggcTCTGCCCGGGattctcaccaataaaaaacgcccggcggcaaccgagcgttctgaacaaatccagatggagttctgaggtcattactggatctatcaacgggagtccaagcgagctcggtactaaaacaattcatccagtaaaatataatattttattttctcccaatcaggcttgatccccagtaagtcaaaaaatagctcgacatactgttcttccccgatatcctccctgatcgaccggacgcagaaggcaatgtcataccacttgtccgccctgccgcttctcccaagatcaataaagccacttactttgccatctttcacaaagatgttgctgtctcccaggtcgccgtgggaaaagacaagttcctcttcgggcttttccgtctttaaaaaatcatacagctcgcgcggatctttaaatggagtatcttcttcccagttttcgcaatccacatcggccagatcgttattcagtaagtaatccaattcggctaagcggccgtctaagctattcgtatagggacaatccgatatgtcgatggagtgaaagagcctgatgcactccgcatacagctcgatagtcttttcagggctttgttcatcttcatacccttccgagcaaaggacgccatcggcctcactcatgagcagattgctccagccatcatgccgttcaaagtgcaggacctttggaacaggcagctttccttccagccatagcatcatgtccttttcccgttccacatcataggtggtccctttataccggctgtccgtcatttttaaatataggatttcattttctcccaccagcttatataccttagcaggagacattccttccgtatcttttacgcagcggtattcttcgatcagttttttcaattccggtgatattctcattttagccatttattatttccttcctcttttctacagtatttaaagataccccaagaagctaattataacaagacgaactccaattcactgttccttgcattctaaaaccttaaatacagaaaacagccttttcaaagttgttttcaaagttggcgtataacatagtatcgacggagccgattttgaaaccacaattatgatagaatttgacgtccttttccgctgcataaccctgcttcggggtcattatagcgattttttcggtatatccatcctttttcgcacgatatacaggattttgccaaagggttcgtgtagactttccttggtgtatccaacggcgtcagccgggcaggataggtgaagtaggcccacccgcgagcgggtgttccttcttcactgtcccttattcgcacctggcggtgctcaacgggaatcctgctctgcgaggctggccgtaTTGACAGACAATCCGTAGGCACAATTTTCGAAAAAACCCGCTTCGGCGGGTTTTTTTATAGCTAAAAATGTTCCAGCGCTGGCACGCAACCTCTCATGCGCTACTTATCACGCCGCGCCAATTTATTACCGCTATGGCCAATTGATCGGCCGGCTTGTCGACGACGGCGGACTCCGTCGTCAGGATCATCCGGGCGAATTCCGTGTTATCCAGTCCCAGAA"

        -- Top level: ~1kbp chunks, with the break in the kanR window [1962, 2485].
        local before, after, _, err = fragment.fragment_with_break(plasmid, 1962, 2485, 800, 1000, {})
        assert.is_nil(err)

        -- Recurse down: fragment each ~1kbp chunk into 151-176bp pieces.
        local chunks = {}
        for _, chunk in ipairs(before) do table.insert(chunks, chunk) end
        for _, chunk in ipairs(after) do table.insert(chunks, chunk) end

        for _, chunk in ipairs(chunks) do
            local pieces, _, piece_err = fragment.fragment(chunk, 151, 176, {})
            assert.is_nil(piece_err)
            assert.is_true(#pieces > 0)
        end
    end)
end)
