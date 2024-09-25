from typing import List, Optional
from .cffi_bindings import ffi, lib

class Part:
    def __init__(self, sequence: str, circular: bool):
        self.sequence = sequence
        self.circular = circular

class Fragment:
    def __init__(self, sequence: str, forward_overhang: str, reverse_overhang: str):
        self.sequence = sequence
        self.forward_overhang = forward_overhang
        self.reverse_overhang = reverse_overhang

def _create_c_string(python_string: str):
    return ffi.new("char[]", python_string.encode('utf-8'))

def _create_c_part(part: Part):
    return {"sequence": _create_c_string(part.sequence), "circular": ffi.cast("int", int(part.circular))}

def _create_c_fragment(fragment: Fragment):
    return {
        "sequence": _create_c_string(fragment.sequence),
        "forward_overhang": _create_c_string(fragment.forward_overhang),
        "reverse_overhang": _create_c_string(fragment.reverse_overhang)
    }

def _fragment_from_c(c_fragment):
    return Fragment(
        ffi.string(c_fragment.sequence).decode('utf-8'),
        ffi.string(c_fragment.forward_overhang).decode('utf-8'),
        ffi.string(c_fragment.reverse_overhang).decode('utf-8')
    )

def cut_with_enzyme_by_name(part: Part, directional: bool, name: str, methylated: bool) -> List[Fragment]:
    c_part = ffi.new("Part*", _create_c_part(part))
    c_name = _create_c_string(name)
    c_directional = ffi.cast("int", int(directional))
    c_methylated = ffi.cast("int", int(methylated))

    result = lib.CutWithEnzymeByName(c_part[0], c_directional, c_name, c_methylated)
    if result.error != ffi.NULL:
        raise Exception(ffi.string(result.error).decode('utf-8'))
    
    fragments = [_fragment_from_c(result.fragments[i]) for i in range(result.size)]
    return fragments

def ligate(fragments: List[Fragment], circular: bool) -> str:
    c_fragments = ffi.new("Fragment[]", [_create_c_fragment(f) for f in fragments])
    c_fragment_count = ffi.cast("int", len(fragments))
    c_circular = ffi.cast("int", int(circular))

    result = lib.Ligate(c_fragments, c_fragment_count, c_circular)
    if result.error != ffi.NULL:
        raise Exception(ffi.string(result.error).decode('utf-8'))
    
    return ffi.string(result.ligation).decode('utf-8')

def golden_gate(sequences: List[Part], cutting_enzyme_name: str, methylated: bool) -> str:
    c_parts = ffi.new("Part[]", [_create_c_part(part) for part in sequences])
    c_sequence_count = ffi.cast("int", len(sequences))
    c_cutting_enzyme_name = _create_c_string(cutting_enzyme_name)
    c_methylated = ffi.cast("int", int(methylated))

    result = lib.GoldenGate(c_parts, c_sequence_count, c_cutting_enzyme_name, c_methylated)
    if result.error != ffi.NULL:
        raise Exception(ffi.string(result.error).decode('utf-8'))
    
    return ffi.string(result.ligation).decode('utf-8')
