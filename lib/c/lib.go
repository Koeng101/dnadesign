package main

/*
#include <stdio.h>
#include <stdlib.h>

typedef struct {
    char* identifier;
    char* sequence;
    char* quality;
    char* optionals;  // Serialized JSON string of the map.
} FastqRead;
*/
import "C"
import (
	"encoding/json"
	"io"
	"unsafe"

	"github.com/koeng101/dnadesign/lib/bio"
)

// Function to create an io.Reader from a C FILE*.
func readerFromCFile(cfile *C.FILE) io.Reader {
	return &fileReader{file: cfile}
}

type fileReader struct {
	file *C.FILE
}

func (f *fileReader) Read(p []byte) (n int, err error) {
	buffer := (*C.char)(unsafe.Pointer(&p[0]))
	count := C.size_t(len(p))
	result := C.fread(unsafe.Pointer(buffer), 1, count, f.file)
	if result == 0 {
		if C.feof(f.file) != 0 {
			return 0, io.EOF
		}
		return 0, io.ErrUnexpectedEOF
	}
	return int(result), nil
}

//export ParseFastqFromCFile
func ParseFastqFromCFile(cfile *C.FILE) (*C.FastqRead, int, *C.char) {
	reader := readerFromCFile(cfile)
	parser := bio.NewFastqParser(reader)
	reads, err := parser.Parse()
	if err != nil {
		return nil, 0, C.CString(err.Error())
	}

	cReads := (*C.FastqRead)(C.malloc(C.size_t(len(reads)) * C.size_t(unsafe.Sizeof(C.FastqRead{}))))
	slice := (*[1<<30 - 1]C.FastqRead)(unsafe.Pointer(cReads))[:len(reads):len(reads)]

	for i, read := range reads {
		optionalsJSON, _ := json.Marshal(read.Optionals)
		slice[i].identifier = C.CString(read.Identifier)
		slice[i].sequence = C.CString(read.Sequence)
		slice[i].quality = C.CString(read.Quality)
		slice[i].optionals = C.CString(string(optionalsJSON))
	}

	return cReads, len(reads), nil
}

func main() {
	// Optionally add any test code or leave empty.
}
