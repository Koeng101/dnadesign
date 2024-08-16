typedef struct FILE FILE;
FILE* fopen(const char* path, const char* mode);
int fclose(FILE* fp);

// FASTA definitions
typedef struct {
    char* identifier;
    char* sequence;
} FastaRecord;

typedef struct {
    FastaRecord* records;
    int numRecords;
    char* error;
} FastaResult;

FastaResult ParseFastaFromCFile(void* cfile);
FastaResult ParseFastaFromCString(char* cstring);

// FASTQ definitions
typedef struct {
    char* key;
    char* value;
} FastqOptional;

typedef struct {
    char* identifier;
    FastqOptional* optionals;
    int optionals_count;
    char* sequence;
    char* quality;
} FastqRecord;

typedef struct {
    FastqRecord* records;
    int numRecords;
    char* error;
} FastqResult;

FastqResult ParseFastqFromCFile(void* cfile);
FastqResult ParseFastqFromCString(char* cstring);
