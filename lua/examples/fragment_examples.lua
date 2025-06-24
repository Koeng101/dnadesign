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
end)
