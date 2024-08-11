package main

/*
#include <stdio.h>
#include <stdlib.h>

// FastaRecord
typedef struct {
    char* identifier;
    char* sequence;
} FastaRecord;
*/
import "C"
import (
	"io"
	"strings"
	"unsafe"

	"github.com/koeng101/dnadesign/lib/bio"
)

/******************************************************************************
Aug 10, 2024

Interoperation with CFile

******************************************************************************/

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

/******************************************************************************
Aug 10, 2024

Fasta

******************************************************************************/

// goFastaToCFasta converts an io.Reader to a C.FastaResult
func goFastaToCFasta(reader io.Reader) (*C.FastaRecord, int, *C.char) {
	parser := bio.NewFastaParser(reader)
	records, err := parser.Parse()
	if err != nil {
		return nil, 0, C.CString(err.Error())
	}

	cRecords := (*C.FastaRecord)(C.malloc(C.size_t(len(records)) * C.size_t(unsafe.Sizeof(C.FastaRecord{}))))
	slice := (*[1<<30 - 1]C.FastaRecord)(unsafe.Pointer(cRecords))[:len(records):len(records)]

	for i, read := range records {
		slice[i].identifier = C.CString(read.Identifier)
		slice[i].sequence = C.CString(read.Sequence)
	}

	return cRecords, len(records), nil
}

//export ParseFastaFromCFile
func ParseFastaFromCFile(cfile *C.FILE) (*C.FastaRecord, int, *C.char) {
	reader := readerFromCFile(cfile)
	return goFastaToCFasta(reader)
}

//export ParseFastaFromCString
func ParseFastaFromCString(cstring *C.char) (*C.FastaRecord, int, *C.char) {
	reader := strings.NewReader(C.GoString(cstring))
	return goFastaToCFasta(reader)
}

/******************************************************************************

main.go

******************************************************************************/

func main() {}
