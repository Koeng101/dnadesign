import pytest
import os
from dnadesign.parsers import parse_fasta_from_c_file, parse_fasta_from_c_string, FastaRecord

def test_parse_fasta_from_c_file():
    current_dir = os.path.dirname(__file__)
    example_path = os.path.join(current_dir, 'data/example.fasta')
    records = parse_fasta_from_c_file(example_path)
    assert len(records) > 0
    assert all(isinstance(r, FastaRecord) for r in records)

def test_parse_fasta_from_c_string():
    fasta_data = ">test\nATCG\n"
    records = parse_fasta_from_c_string(fasta_data)
    assert len(records) == 1
    assert records[0].identifier == "test"
    assert records[0].sequence == "ATCG"
