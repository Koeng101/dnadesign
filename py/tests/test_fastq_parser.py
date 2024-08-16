import pytest
import os
from dnadesign.parsers import parse_fastq_from_c_file, parse_fastq_from_c_string, FastqRecord

def test_parse_fastq_from_c_file():
    current_dir = os.path.dirname(__file__)
    example_path = os.path.join(current_dir, 'data/example.fastq')
    records = parse_fastq_from_c_file(example_path)
    assert len(records) > 0
    assert all(isinstance(r, FastqRecord) for r in records)

def test_parse_fastq_from_c_string():
    fastq_data = "@test\nATCG\n+\nIIII\n"
    records = parse_fastq_from_c_string(fastq_data)
    assert len(records) == 1
    assert records[0].identifier == "test"
    assert records[0].sequence == "ATCG"
    assert records[0].quality == "IIII"
    assert records[0].optionals == {}

def test_parse_fastq_with_optionals():
    fastq_data = "@test read=1 ch=2\nATCG\n+\nIIII\n"
    records = parse_fastq_from_c_string(fastq_data)
    assert len(records) == 1
    assert records[0].identifier == "test"
    assert records[0].sequence == "ATCG"
    assert records[0].quality == "IIII"
    assert records[0].optionals == {"read": "1", "ch": "2"}

def test_multiple_fastq_records():
    fastq_data = "@seq1\nACGT\n+\nHHHH\n@seq2\nTGCA\n+\nIIII\n"
    records = parse_fastq_from_c_string(fastq_data)
    assert len(records) == 2
    assert records[0].identifier == "seq1"
    assert records[0].sequence == "ACGT"
    assert records[0].quality == "HHHH"
    assert records[1].identifier == "seq2"
    assert records[1].sequence == "TGCA"
    assert records[1].quality == "IIII"

def test_invalid_fastq():
    invalid_fastq = "@test\nATCG\n+\nII\n"  # Quality string too short
    with pytest.raises(Exception):
        parse_fastq_from_c_string(invalid_fastq)
