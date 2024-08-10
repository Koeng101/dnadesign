from typing import List, Optional
from .cffi_bindings import ffi, lib

class FastaRecord:
    def __init__(self, identifier: str, sequence: str):
        self.identifier = identifier
        self.sequence = sequence

def parse_fasta_from_c_file(file_path: str) -> List[FastaRecord]:
    cfile = lib.fopen(file_path.encode('utf-8'), "r".encode('utf-8'))
    result = lib.ParseFastaFromCFile(cfile)
    return _process_result(result)

def parse_fasta_from_c_string(cstring: str) -> List[FastaRecord]:
    result = lib.ParseFastaFromCString(cstring.encode('utf-8'))
    return _process_result(result)

def _process_result(result) -> List[FastaRecord]:
    if result.error != ffi.NULL:
        error_str = ffi.string(result.error).decode('utf-8')
        raise Exception("Error parsing FASTA: " + error_str)
    num_records = result.numRecords
    records = ffi.cast("FastaRecord*", result.records)
    return [FastaRecord(ffi.string(records[i].identifier).decode('utf-8'),
                        ffi.string(records[i].sequence).decode('utf-8'))
            for i in range(num_records)]
