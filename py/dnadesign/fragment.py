from typing import List, Tuple
from .cffi_bindings import ffi, lib

def set_efficiency(overhangs: List[str]) -> float:
    c_overhangs = [ffi.new("char[]", overhang.encode('utf-8')) for overhang in overhangs]
    c_overhang_array = ffi.new("char *[]", c_overhangs)
    return lib.SetEfficiency(c_overhang_array, len(overhangs))

def next_overhangs(current_overhangs: List[str]) -> Tuple[List[str], List[float]]:
    c_overhangs = [ffi.new("char[]", overhang.encode('utf-8')) for overhang in current_overhangs]
    c_overhang_array = ffi.new("char *[]", c_overhangs)
    result = lib.NextOverhangs(c_overhang_array, len(current_overhangs))
    if result.error != ffi.NULL:
        raise Exception(ffi.string(result.error).decode('utf-8'))
    overhangs = [ffi.string(result.overhangs[i]).decode('utf-8') for i in range(result.size)]
    efficiencies = [result.efficiencies[i] for i in range(result.size)]
    return overhangs, efficiencies

def next_overhang(current_overhangs: List[str]) -> str:
    c_overhangs = [ffi.new("char[]", overhang.encode('utf-8')) for overhang in current_overhangs]
    c_overhang_array = ffi.new("char *[]", c_overhangs)
    result = lib.NextOverhang(c_overhang_array, len(current_overhangs))
    return ffi.string(result).decode('utf-8')

def fragment_sequence(sequence: str, min_fragment_size: int, max_fragment_size: int, exclude_overhangs: List[str]) -> Tuple[List[str], float]:
    c_exclude_overhangs = [ffi.new("char[]", overhang.encode('utf-8')) for overhang in exclude_overhangs]
    c_exclude_overhang_array = ffi.new("char *[]", c_exclude_overhangs)
    result = lib.FragmentSequence(sequence.encode('utf-8'), min_fragment_size, max_fragment_size, c_exclude_overhang_array, len(exclude_overhangs))
    if result.error != ffi.NULL:
        raise Exception(ffi.string(result.error).decode('utf-8'))
    fragments = [ffi.string(result.fragments[i]).decode('utf-8') for i in range(result.size)]
    return fragments, result.efficiency

def fragment_sequence_with_overhangs(sequence: str, min_fragment_size: int, max_fragment_size: int, exclude_overhangs: List[str], include_overhangs: List[str]) -> Tuple[List[str], float]:
    c_exclude_overhangs = [ffi.new("char[]", overhang.encode('utf-8')) for overhang in exclude_overhangs]
    c_exclude_overhang_array = ffi.new("char *[]", c_exclude_overhangs)
    c_include_overhangs = [ffi.new("char[]", overhang.encode('utf-8')) for overhang in include_overhangs]
    c_include_overhang_array = ffi.new("char *[]", c_include_overhangs)
    result = lib.FragmentSequenceWithOverhangs(sequence.encode('utf-8'), min_fragment_size, max_fragment_size, c_exclude_overhang_array, len(exclude_overhangs), c_include_overhang_array, len(include_overhangs))
    if result.error != ffi.NULL:
        raise Exception(ffi.string(result.error).decode('utf-8'))
    fragments = [ffi.string(result.fragments[i]).decode('utf-8') for i in range(result.size)]
    return fragments, result.efficiency

def recursive_fragment_sequence(sequence: str, max_coding_size_oligo: int, assembly_pattern: List[int], exclude_overhangs: List[str], include_overhangs: List[str]) -> 'Assembly':
    c_assembly_pattern = ffi.new("int[]", assembly_pattern)
    c_exclude_overhangs = [ffi.new("char[]", overhang.encode('utf-8')) for overhang in exclude_overhangs]
    c_exclude_overhang_array = ffi.new("char *[]", c_exclude_overhangs)
    c_include_overhangs = [ffi.new("char[]", overhang.encode('utf-8')) for overhang in include_overhangs]
    c_include_overhang_array = ffi.new("char *[]", c_include_overhangs)
    result = lib.RecursiveFragmentSequence(sequence.encode('utf-8'), max_coding_size_oligo, c_assembly_pattern, len(assembly_pattern), c_exclude_overhang_array, len(exclude_overhangs), c_include_overhang_array, len(include_overhangs))
    if result.error != ffi.NULL:
        raise Exception(ffi.string(result.error).decode('utf-8'))
    return Assembly.from_c(result)

class Assembly:
    def __init__(self, sequence: str, fragments: List[str], efficiency: float, sub_assemblies: List['Assembly']):
        self.sequence = sequence
        self.fragments = fragments
        self.efficiency = efficiency
        self.sub_assemblies = sub_assemblies

    @classmethod
    def from_c(cls, c_assembly):
        sequence = ffi.string(c_assembly.sequence).decode('utf-8')
        fragments = [ffi.string(c_assembly.fragments[i]).decode('utf-8') for i in range(c_assembly.fragmentCount)]
        efficiency = c_assembly.efficiency
        sub_assemblies = [cls.from_c(c_assembly.subAssemblies[i]) for i in range(c_assembly.subAssemblyCount)]
        return cls(sequence, fragments, efficiency, sub_assemblies)
