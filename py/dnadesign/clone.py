from typing import List, Optional
from .cffi_bindings import ffi, lib

class Part:
    def __init__(self, sequence: str, circular: bool):
        self.sequence = sequence
        self.circular = circular

    def to_c(self):
        return ffi.new("Part*", [self.sequence.encode('utf-8'), int(self.circular)])

class Fragment:
    def __init__(self, sequence: str, forward_overhang: str, reverse_overhang: str):
        self.sequence = sequence
        self.forward_overhang = forward_overhang
        self.reverse_overhang = reverse_overhang

    @classmethod
    def from_c(cls, c_fragment):
        return cls(
            ffi.string(c_fragment.sequence).decode('utf-8'),
            ffi.string(c_fragment.forward_overhang).decode('utf-8'),
            ffi.string(c_fragment.reverse_overhang).decode('utf-8')
        )

def cut_with_enzyme_by_name(part: Part, directional: bool, name: str, methylated: bool) -> List[Fragment]:
    result = lib.CutWithEnzymeByName(part.to_c()[0], int(directional), name.encode('utf-8'), int(methylated))
    if result.error != ffi.NULL:
        raise Exception(ffi.string(result.error).decode('utf-8'))
    return [Fragment.from_c(result.fragments[i]) for i in range(result.size)]

def ligate(fragments: List[Fragment], circular: bool) -> str:
    c_fragments = ffi.new("Fragment[]", [ffi.new("Fragment*", [f.sequence.encode('utf-8'), f.forward_overhang.encode('utf-8'), f.reverse_overhang.encode('utf-8')]) for f in fragments])
    result = lib.Ligate(c_fragments, len(fragments), int(circular))
    if result.error != ffi.NULL:
        raise Exception(ffi.string(result.error).decode('utf-8'))
    return ffi.string(result.ligation).decode('utf-8')

def golden_gate(sequences: List[Part], cutting_enzyme_name: str, methylated: bool) -> str:
    c_parts = ffi.new("Part[]", [part.to_c()[0] for part in sequences])
    result = lib.GoldenGate(c_parts, len(sequences), cutting_enzyme_name.encode('utf-8'), int(methylated))
    if result.error != ffi.NULL:
        raise Exception(ffi.string(result.error).decode('utf-8'))
    return ffi.string(result.ligation).decode('utf-8')
