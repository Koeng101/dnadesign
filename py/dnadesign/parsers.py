from typing import List, Optional, Dict
from .cffi_bindings import ffi, lib
import os

class FastaRecord:
    def __init__(self, identifier: str, sequence: str):
        self.identifier = identifier
        self.sequence = sequence

class FastqRecord:
    def __init__(self, identifier: str, sequence: str, quality: str, optionals: Dict[str, str]):
        self.identifier = identifier
        self.sequence = sequence
        self.quality = quality
        self.optionals = optionals

def _safe_open_file(file_path: str):
    if not os.path.exists(file_path):
        raise FileNotFoundError(f"The file {file_path} does not exist.")
    cfile = lib.fopen(file_path.encode('utf-8'), "r".encode('utf-8'))
    if cfile == ffi.NULL:
        raise IOError(f"Failed to open the file {file_path}.")
    return cfile

def parse_fasta_from_c_file(file_path: str) -> List[FastaRecord]:
    try:
        cfile = _safe_open_file(file_path)
        result = lib.ParseFastaFromCFile(cfile)
        return _process_fasta_result(result)
    finally:
        if 'cfile' in locals() and cfile != ffi.NULL:
            lib.fclose(cfile)

def parse_fasta_from_c_string(cstring: str) -> List[FastaRecord]:
    result = lib.ParseFastaFromCString(cstring.encode('utf-8'))
    return _process_fasta_result(result)

def _process_fasta_result(result) -> List[FastaRecord]:
    if result.error != ffi.NULL:
        error_str = ffi.string(result.error).decode('utf-8')
        raise Exception("Error parsing FASTA: " + error_str)
    num_records = result.numRecords
    records = ffi.cast("FastaRecord*", result.records)
    return [FastaRecord(ffi.string(records[i].identifier).decode('utf-8'),
                        ffi.string(records[i].sequence).decode('utf-8'))
            for i in range(num_records)]

def parse_fastq_from_c_file(file_path: str) -> List[FastqRecord]:
    try:
        cfile = _safe_open_file(file_path)
        result = lib.ParseFastqFromCFile(cfile)
        return _process_fastq_result(result)
    finally:
        if 'cfile' in locals() and cfile != ffi.NULL:
            lib.fclose(cfile)

def parse_fastq_from_c_string(cstring: str) -> List[FastqRecord]:
    result = lib.ParseFastqFromCString(cstring.encode('utf-8'))
    return _process_fastq_result(result)

def _process_fastq_result(result) -> List[FastqRecord]:
    if result.error != ffi.NULL:
        error_str = ffi.string(result.error).decode('utf-8')
        raise Exception("Error parsing FASTQ: " + error_str)
    num_records = result.numRecords
    records = ffi.cast("FastqRecord*", result.records)
    fastq_records = []
    for i in range(num_records):
        optionals = {}
        for j in range(records[i].optionals_count):
            key = ffi.string(records[i].optionals[j].key).decode('utf-8')
            value = ffi.string(records[i].optionals[j].value).decode('utf-8')
            optionals[key] = value
        fastq_records.append(FastqRecord(
            ffi.string(records[i].identifier).decode('utf-8'),
            ffi.string(records[i].sequence).decode('utf-8'),
            ffi.string(records[i].quality).decode('utf-8'),
            optionals
        ))
    return fastq_records
