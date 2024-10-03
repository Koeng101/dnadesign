from typing import List, Tuple
from .cffi_bindings import ffi, lib

class Assembly:
    def __init__(self, sequence: str, fragments: List[str], efficiency: float, sub_assemblies: List['Assembly']):
        self.sequence = sequence
        self.fragments = fragments
        self.efficiency = efficiency
        self.sub_assemblies = sub_assemblies

def _create_c_string_array(python_strings: List[str]):
    c_strings = [ffi.new("char[]", s.encode('utf-8')) for s in python_strings]
    c_array = ffi.new("char *[]", c_strings)
    return c_array, c_strings  # Return c_strings to keep them alive

def _c_string_array_to_python(c_array, size):
    return [ffi.string(c_array[i]).decode('utf-8') for i in range(size)]

def set_efficiency(overhangs: List[str]) -> float:
    c_overhangs, _ = _create_c_string_array(overhangs)
    return lib.SetEfficiency(c_overhangs, len(overhangs))

def next_overhangs(current_overhangs: List[str]) -> Tuple[List[str], List[float]]:
    c_overhangs, _ = _create_c_string_array(current_overhangs)
    result = lib.NextOverhangs(c_overhangs, len(current_overhangs))
    
    if result.error != ffi.NULL:
        raise Exception(ffi.string(result.error).decode('utf-8'))
    
    overhangs = _c_string_array_to_python(result.overhangs, result.size)
    efficiencies = [result.efficiencies[i] for i in range(result.size)]
    return overhangs, efficiencies

def next_overhang(current_overhangs: List[str]) -> str:
    c_overhangs, _ = _create_c_string_array(current_overhangs)
    result = lib.NextOverhang(c_overhangs, len(current_overhangs))
    return ffi.string(result).decode('utf-8')

def fragment(sequence: str, min_fragment_size: int, max_fragment_size: int, exclude_overhangs: List[str]) -> Tuple[List[str], float, str]:
    c_sequence = ffi.new("char[]", sequence.encode('utf-8'))
    c_exclude_overhangs, _ = _create_c_string_array(exclude_overhangs)
    
    result = lib.FragmentSequence(c_sequence, min_fragment_size, max_fragment_size, c_exclude_overhangs, len(exclude_overhangs))
    
    if result.error != ffi.NULL:
        error = ffi.string(result.error).decode('utf-8')
        return [], 0.0, error
    
    fragments = _c_string_array_to_python(result.fragments, result.size)
    return fragments, result.efficiency, None

def fragment_with_overhangs(sequence: str, min_fragment_size: int, max_fragment_size: int, 
                            exclude_overhangs: List[str], include_overhangs: List[str]) -> Tuple[List[str], float, str]:
    c_sequence = ffi.new("char[]", sequence.encode('utf-8'))
    c_exclude_overhangs, _ = _create_c_string_array(exclude_overhangs)
    c_include_overhangs, _ = _create_c_string_array(include_overhangs)
    
    result = lib.FragmentSequenceWithOverhangs(c_sequence, min_fragment_size, max_fragment_size, 
                                               c_exclude_overhangs, len(exclude_overhangs),
                                               c_include_overhangs, len(include_overhangs))
    
    if result.error != ffi.NULL:
        error = ffi.string(result.error).decode('utf-8')
        return [], 0.0, error
    
    fragments = _c_string_array_to_python(result.fragments, result.size)
    return fragments, result.efficiency, None

def _assembly_from_c(c_assembly) -> Assembly:
    sequence = ffi.string(c_assembly.sequence).decode('utf-8')
    fragments = _c_string_array_to_python(c_assembly.fragments, c_assembly.fragmentCount)
    efficiency = c_assembly.efficiency
    sub_assemblies = [_assembly_from_c(c_assembly.subAssemblies[i]) for i in range(c_assembly.subAssemblyCount)]
    return Assembly(sequence, fragments, efficiency, sub_assemblies)

def recursive_fragment(sequence: str, max_coding_size_oligo: int, assembly_pattern: List[int],
                       exclude_overhangs: List[str], include_overhangs: List[str],
                       forward_flank: str, reverse_flank: str) -> Assembly:
    c_sequence = ffi.new("char[]", sequence.encode('utf-8'))
    c_forward_flank = ffi.new("char[]", forward_flank.encode('utf-8'))
    c_reverse_flank = ffi.new("char[]", reverse_flank.encode('utf-8'))
    c_assembly_pattern = ffi.new("int[]", assembly_pattern)
    c_exclude_overhangs, _ = _create_c_string_array(exclude_overhangs)
    c_include_overhangs, _ = _create_c_string_array(include_overhangs)
    
    result = lib.RecursiveFragmentSequence(c_sequence, max_coding_size_oligo, c_assembly_pattern, len(assembly_pattern),
                                           c_exclude_overhangs, len(exclude_overhangs),
                                           c_include_overhangs, len(include_overhangs),
                                           c_forward_flank, c_reverse_flank)
    
    if result.error != ffi.NULL:
        raise Exception(ffi.string(result.error).decode('utf-8'))
    
    return _assembly_from_c(result.assembly)
