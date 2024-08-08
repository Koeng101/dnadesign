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

