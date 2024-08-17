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

// Genbank definitions
// GenbankLocation
typedef struct GenbankLocation {
    int start;
    int end;
    int complement;
    int join;
    int five_prime_partial;
    int three_prime_partial;
    char* gbk_location_string;
    struct GenbankLocation* sub_locations;
    int sub_locations_count;
} GenbankLocation;

// GenbankFeature
typedef struct {
    char* type_;
    char* description;
    char** attribute_keys;
    char*** attribute_values;
    int* attribute_value_counts;
    int attribute_count;
    char* sequence_hash;
    char* sequence_hash_function;
    char* sequence;
    GenbankLocation location;
} GenbankFeature;

// GenbankReference
typedef struct {
    char* authors;
    char* title;
    char* journal;
    char* pub_med;
    char* remark;
    char* range_;
    char* consortium;
} GenbankReference;

// GenbankLocus
typedef struct {
    char* name;
    char* sequence_length;
    char* molecule_type;
    char* genbank_division;
    char* modification_date;
    char* sequence_coding;
    int circular;
} GenbankLocus;

// GenbankBaseCount
typedef struct {
    char base;
    int count;
} GenbankBaseCount;

// GenbankMeta
typedef struct {
    char* date;
    char* definition;
    char* accession;
    char* version;
    char* keywords;
    char* organism;
    char* source;
    char** taxonomy;
    int taxonomy_count;
    char* origin;
    GenbankLocus locus;
    GenbankReference* references;
    int reference_count;
    GenbankBaseCount* base_counts;
    int base_count_count;
    char** other_keys;
    char** other_values;
    int other_count;
    char* name;
    char* sequence_hash;
    char* sequence_hash_function;
} GenbankMeta;

// Genbank
typedef struct {
    GenbankMeta meta;
    GenbankFeature* features;
    int feature_count;
    char* sequence;
} Genbank;

typedef struct {
    Genbank* records;
    int numRecords;
    char* error;
} GenbankResult;

GenbankResult ParseGenbankFromCFile(void* cfile);
GenbankResult ParseGenbankFromCString(char* cstring);
