from cffi import FFI
import os
import sys

ffi = FFI()

# Define common types based on the platform architecture
is_64b = sys.maxsize > 2**32
if is_64b:
    ffi.cdef("typedef long GoInt;\n")
else:
    ffi.cdef("typedef int GoInt;\n")

current_dir = os.path.dirname(__file__)

# Build the path to definitions.h and libdnadesign relative to the current script
definitions_path = os.path.join(current_dir, 'definitions.h')
lib_path = os.path.join(current_dir, 'libdnadesign.so')

# Read the C declarations from an external file
with open(definitions_path, 'r') as f:
    ffi.cdef(f.read())

# Load the shared library
lib = ffi.dlopen(lib_path)
