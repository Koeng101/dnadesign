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

# Genbank time
class GenbankLocus:
    def __init__(self, name: str, sequence_length: str, molecule_type: str, genbank_division: str,
                 modification_date: str, sequence_coding: str, circular: bool):
        self.name = name
        self.sequence_length = sequence_length
        self.molecule_type = molecule_type
        self.genbank_division = genbank_division
        self.modification_date = modification_date
        self.sequence_coding = sequence_coding
        self.circular = circular

class GenbankReference:
    def __init__(self, authors: str, title: str, journal: str, pub_med: str, remark: str, range_: str, consortium: str):
        self.authors = authors
        self.title = title
        self.journal = journal
        self.pub_med = pub_med
        self.remark = remark
        self.range_ = range_
        self.consortium = consortium

class GenbankBaseCount:
    def __init__(self, base: str, count: int):
        self.base = base
        self.count = count

class GenbankMeta:
    def __init__(self, date: str, definition: str, accession: str, version: str, keywords: str,
                 organism: str, source: str, taxonomy: List[str], origin: str, locus: GenbankLocus,
                 references: List[GenbankReference], base_counts: List[GenbankBaseCount], other: Dict[str, str],
                 name: str, sequence_hash: str, hash_function: str):
        self.date = date
        self.definition = definition
        self.accession = accession
        self.version = version
        self.keywords = keywords
        self.organism = organism
        self.source = source
        self.taxonomy = taxonomy
        self.origin = origin
        self.locus = locus
        self.references = references
        self.base_counts = base_counts
        self.other = other
        self.name = name
        self.sequence_hash = sequence_hash
        self.hash_function = hash_function

# Update the Location class to match the C struct
class GenbankLocation:
    def __init__(self, start: int, end: int, complement: bool, join: bool, five_prime_partial: bool,
                 three_prime_partial: bool, gbk_location_string: str, sub_locations: List['GenbankLocation']):
        self.start = start
        self.end = end
        self.complement = complement
        self.join = join
        self.five_prime_partial = five_prime_partial
        self.three_prime_partial = three_prime_partial
        self.gbk_location_string = gbk_location_string
        self.sub_locations = sub_locations

class GenbankFeature:
    def __init__(self, type_: str, description: str, attributes: Dict[str, List[str]],
                 sequence_hash: str, hash_function: str, sequence: str, location: GenbankLocation):
        self.type_ = type_
        self.description = description
        self.attributes = attributes
        self.sequence_hash = sequence_hash
        self.hash_function = hash_function
        self.sequence = sequence
        self.location = location

class Genbank:
    def __init__(self, meta: GenbankMeta, features: List[GenbankFeature], sequence: str):
        self.meta = meta
        self.features = features
        self.sequence = sequence

def _safe_open_file(file_path: str):
    if not os.path.exists(file_path):
        raise FileNotFoundError(f"The file {file_path} does not exist.")
    cfile = lib.fopen(file_path.encode('utf-8'), "r".encode('utf-8'))
    if cfile == ffi.NULL:
        raise IOError(f"Failed to open the file {file_path}.")
    return cfile

def parse_genbank_from_c_file(file_path: str) -> List[Genbank]:
    try:
        cfile = _safe_open_file(file_path)
        result = lib.ParseGenbankFromCFile(cfile)
        return _process_genbank_result(result)
    finally:
        if 'cfile' in locals() and cfile != ffi.NULL:
            lib.fclose(cfile)

def parse_genbank_from_c_string(cstring: str) -> List[Genbank]:
    result = lib.ParseGenbankFromCString(cstring.encode('utf-8'))
    return _process_genbank_result(result)

def _process_genbank_result(result) -> List[Genbank]:
    if result.error != ffi.NULL:
        error_str = ffi.string(result.error).decode('utf-8')
        raise Exception("Error parsing Genbank: " + error_str)
    num_records = result.numRecords
    records = ffi.cast("Genbank*", result.records)
    return [_convert_genbank_record(records[i]) for i in range(num_records)]

def _convert_genbank_record(record) -> Genbank:
    meta = _convert_meta(record.meta)
    features = [_convert_feature(record.features[i]) for i in range(record.feature_count)]
    sequence = ffi.string(record.sequence).decode('utf-8')
    return Genbank(meta, features, sequence)

def _convert_meta(meta) -> GenbankMeta:
    locus = GenbankLocus(
        ffi.string(meta.locus.name).decode('utf-8'),
        ffi.string(meta.locus.sequence_length).decode('utf-8'),
        ffi.string(meta.locus.molecule_type).decode('utf-8'),
        ffi.string(meta.locus.genbank_division).decode('utf-8'),
        ffi.string(meta.locus.modification_date).decode('utf-8'),
        ffi.string(meta.locus.sequence_coding).decode('utf-8'),
        meta.locus.circular
    )
    
    references = [
        GenbankReference(
            ffi.string(ref.authors).decode('utf-8'),
            ffi.string(ref.title).decode('utf-8'),
            ffi.string(ref.journal).decode('utf-8'),
            ffi.string(ref.pub_med).decode('utf-8'),
            ffi.string(ref.remark).decode('utf-8'),
            ffi.string(ref.range_).decode('utf-8'),
            ffi.string(ref.consortium).decode('utf-8')
        ) for ref in meta.references[0:meta.reference_count]
    ]
    
    base_counts = [
        GenbankBaseCount(
            ffi.string(bc.base).decode('utf-8'),
            bc.count
        ) for bc in meta.base_counts[0:meta.base_count_count]
    ]
    
    other = {}
    for i in range(meta.other_count):
        key = ffi.string(meta.other_keys[i]).decode('utf-8')
        value = ffi.string(meta.other_values[i]).decode('utf-8')
        other[key] = value
    
    return GenbankMeta(
        ffi.string(meta.date).decode('utf-8'),
        ffi.string(meta.definition).decode('utf-8'),
        ffi.string(meta.accession).decode('utf-8'),
        ffi.string(meta.version).decode('utf-8'),
        ffi.string(meta.keywords).decode('utf-8'),
        ffi.string(meta.organism).decode('utf-8'),
        ffi.string(meta.source).decode('utf-8'),
        [ffi.string(tax).decode('utf-8') for tax in meta.taxonomy[0:meta.taxonomy_count]],
        ffi.string(meta.origin).decode('utf-8'),
        locus,
        references,
        base_counts,
        other,
        ffi.string(meta.name).decode('utf-8'),
        ffi.string(meta.sequence_hash).decode('utf-8'),
        ffi.string(meta.sequence_hash_function).decode('utf-8')
    )

def _convert_location(loc) -> GenbankLocation:
    sub_locations = []

    if loc.sub_locations_count > 0:
        sub_locations_array = ffi.cast("GenbankLocation *", loc.sub_locations)
        for i in range(loc.sub_locations_count):
            sub_loc = sub_locations_array[i]
            sub_locations.append(_convert_location(sub_loc))

    return GenbankLocation(
        loc.start,
        loc.end,
        bool(loc.complement),
        bool(loc.join),
        bool(loc.five_prime_partial),
        bool(loc.three_prime_partial),
        ffi.string(loc.gbk_location_string).decode('utf-8') if loc.gbk_location_string else "",
        sub_locations
    )

def _convert_feature(feature) -> GenbankFeature:
    attributes = {}
    for i in range(feature.attribute_count):
        key = ffi.string(feature.attribute_keys[i]).decode('utf-8')
        values = [ffi.string(feature.attribute_values[i][j]).decode('utf-8')
                  for j in range(feature.attribute_value_counts[i])]
        attributes[key] = values
    
    location = _convert_location(feature.location)
    
    return GenbankFeature(
        ffi.string(feature.type_).decode('utf-8'),
        ffi.string(feature.description).decode('utf-8'),
        attributes,
        ffi.string(feature.sequence_hash).decode('utf-8'),
        ffi.string(feature.sequence_hash_function).decode('utf-8'),
        ffi.string(feature.sequence).decode('utf-8'),
        location
    )
