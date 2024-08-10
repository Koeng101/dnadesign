typedef struct FILE FILE;
FILE* fopen(const char* path, const char* mode);
int fclose(FILE* fp);

typedef struct {
    char* identifier;
    char* sequence;
} FastaRecord;

typedef struct {
    FastaRecord* records;
    GoInt numRecords;
    char* error;
} FastaResult;

FastaResult ParseFastaFromCFile(void* cfile);
FastaResult ParseFastaFromCString(char* cstring);
