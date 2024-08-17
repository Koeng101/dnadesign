import pytest
import os
from dnadesign.parsers import parse_genbank_from_c_file, parse_genbank_from_c_string, Genbank, GenbankMeta, GenbankFeature

def test_parse_genbank_from_c_file():
    current_dir = os.path.dirname(__file__)
    example_path = os.path.join(current_dir, 'data/example.gb')
    records = parse_genbank_from_c_file(example_path)
    assert len(records) > 0
    assert all(isinstance(r, Genbank) for r in records)

    # Test the first record
    first_record = records[0]
    assert isinstance(first_record.meta, GenbankMeta)
    assert isinstance(first_record.features, list)
    assert all(isinstance(f, GenbankFeature) for f in first_record.features)
    assert isinstance(first_record.sequence, str)

    # Test some fields of the first record
    assert first_record.meta.accession
    assert first_record.meta.version
    assert first_record.meta.organism
    assert len(first_record.features) > 0
    assert first_record.sequence

def test_parse_genbank_from_c_string():
    genbank_data = """LOCUS       SCU49845     5028 bp    DNA             PLN       21-JUN-1999
DEFINITION  Saccharomyces cerevisiae TCP1-beta gene, partial cds, and Axl2p
            (AXL2) and Rev7p (REV7) genes, complete cds.
ACCESSION   U49845
VERSION     U49845.1  GI:1293613
KEYWORDS    .
SOURCE      Saccharomyces cerevisiae (baker's yeast)
  ORGANISM  Saccharomyces cerevisiae
            Eukaryota; Fungi; Ascomycota; Saccharomycotina; Saccharomycetes;
            Saccharomycetales; Saccharomycetaceae; Saccharomyces.
REFERENCE   1  (bases 1 to 5028)
  AUTHORS   Torpey,L.E., Gibbs,P.E., Nelson,J. and Lawrence,C.W.
  TITLE     Cloning and sequence of REV7, a gene whose function is required for
            DNA damage-induced mutagenesis in Saccharomyces cerevisiae
  JOURNAL   Yeast 10 (11), 1503-1509 (1994)
  PUBMED    7871890
FEATURES             Location/Qualifiers
     source          1..5028
                     /organism="Saccharomyces cerevisiae"
                     /db_xref="taxon:4932"
                     /chromosome="IX"
                     /map="9"
ORIGIN
        1 gatcctccat atacaacggt atctccacct caggtttaga tctcaacaac ggaaccattg
       61 ccgacatgag acagttaggt atcgtcgaga gttacaagct aaaacgagca gtagtcagct
      121 ctgcatctga agccgctgaa gttctactaa gggtggataa catcatccgt gcaagaccaa
//
"""
    records = parse_genbank_from_c_string(genbank_data)
    assert len(records) == 1
    record = records[0]
    assert record.meta.locus.name == "SCU49845"
    assert record.meta.accession == "U49845"
    assert record.meta.organism == "Saccharomyces cerevisiae"
    assert len(record.features) == 1
    assert record.features[0].type_ == "source"
    assert record.sequence.startswith("gatcctccat")
    assert len(record.sequence) == 180  # Based on the ORIGIN section in the example
