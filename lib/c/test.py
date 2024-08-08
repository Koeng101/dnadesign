from __future__ import print_function
import sys
import os
from cffi import FFI

ffi = FFI()

# Define common types based on the platform architecture
is_64b = sys.maxsize > 2**32
if is_64b:
    ffi.cdef("typedef long GoInt;\n")
else:
    ffi.cdef("typedef int GoInt;\n")

# Define the FastqRead struct and the function signature for ParseFastqFromCFile
ffi.cdef("""
typedef struct FILE FILE;
FILE *fopen(const char *path, const char *mode);
int fclose(FILE *fp);


typedef struct {
    char* identifier;
    char* sequence;
    char* quality;
    char* optionals;  // Serialized JSON string of the map.
} FastqRead;

typedef struct {
    FastqRead* reads;
    GoInt numReads;
    char* error;
} FastqResult;

FastqResult ParseFastqFromCFile(void* cfile);
""")


# Load the shared library compiled from Go
lib = ffi.dlopen("./libawesome.so")

file_path = "example.fastq".encode('utf-8')  # Convert the string to bytes
mode = "r".encode('utf-8')  # Convert the mode to bytes as well
cfile = lib.fopen(file_path, mode)

# Call the function from the shared library
result = lib.ParseFastqFromCFile(cfile)

# Check if there was an error
if result.error != ffi.NULL:
    error_str = ffi.string(result.error).decode('utf-8')
    print("Error parsing FASTQ:", error_str)
else:
    # Process the reads
    num_reads = result.numReads
    reads = ffi.cast("FastqRead*", result.reads)
    for i in range(num_reads):
        identifier = ffi.string(reads[i].identifier).decode('utf-8')
        sequence = ffi.string(reads[i].sequence).decode('utf-8')
        quality = ffi.string(reads[i].quality).decode('utf-8')
        optionals = ffi.string(reads[i].optionals).decode('utf-8')
        print(f"Read {i+1}: Identifier={identifier}, Sequence={sequence}, Quality={quality}, Optionals={optionals}")

