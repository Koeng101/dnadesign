package main

/*
#include <stdio.h>
#include <stdlib.h>

// FastaRecord
typedef struct {
    char* identifier;
    char* sequence;
} FastaRecord;

// FastqOptional
typedef struct {
    char* key;
    char* value;
} FastqOptional;

// FastqRecord
typedef struct {
    char* identifier;
    FastqOptional* optionals;
    int optionals_count;
    char* sequence;
    char* quality;
} FastqRecord;

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

// Part
typedef struct {
    char* sequence;
    int circular;
} Part;

// Fragment
typedef struct {
    char* sequence;
    char* forward_overhang;
    char* reverse_overhang;
} Fragment;

// Assembly
typedef struct {
    char* sequence;
    char** fragments;
    int fragmentCount;
    double efficiency;
    void* subAssemblies;
    int subAssemblyCount;
} Assembly;
*/
import "C"
import (
	"fmt"
	"io"
	"strings"
	"unsafe"

	"github.com/koeng101/dnadesign/lib/bio"
	"github.com/koeng101/dnadesign/lib/bio/genbank"
	"github.com/koeng101/dnadesign/lib/clone"
	"github.com/koeng101/dnadesign/lib/synthesis/fragment"
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
Aug 16, 2024

Fastq

******************************************************************************/

// goFastqToCFastq converts an io.Reader to a C.FastqRecord
func goFastqToCFastq(reader io.Reader) (*C.FastqRecord, int, *C.char) {
	parser := bio.NewFastqParser(reader)
	records, err := parser.Parse()
	if err != nil {
		return nil, 0, C.CString(err.Error())
	}
	cRecords := (*C.FastqRecord)(C.malloc(C.size_t(len(records)) * C.size_t(unsafe.Sizeof(C.FastqRecord{}))))
	slice := (*[1<<30 - 1]C.FastqRecord)(unsafe.Pointer(cRecords))[:len(records):len(records)]

	for i, read := range records {
		slice[i].identifier = C.CString(read.Identifier)
		slice[i].sequence = C.CString(read.Sequence)
		slice[i].quality = C.CString(read.Quality)

		optionalsCount := len(read.Optionals)
		slice[i].optionals_count = C.int(optionalsCount)
		if optionalsCount > 0 {
			slice[i].optionals = (*C.FastqOptional)(C.malloc(C.size_t(optionalsCount) * C.size_t(unsafe.Sizeof(C.FastqOptional{}))))
			optionalsSlice := (*[1<<30 - 1]C.FastqOptional)(unsafe.Pointer(slice[i].optionals))[:optionalsCount:optionalsCount]

			j := 0
			for key, value := range read.Optionals {
				optionalsSlice[j].key = C.CString(key)
				optionalsSlice[j].value = C.CString(value)
				j++
			}
		}
	}

	return cRecords, len(records), nil
}

//export ParseFastqFromCFile
func ParseFastqFromCFile(cfile *C.FILE) (*C.FastqRecord, int, *C.char) {
	reader := readerFromCFile(cfile)
	return goFastqToCFastq(reader)
}

//export ParseFastqFromCString
func ParseFastqFromCString(cstring *C.char) (*C.FastqRecord, int, *C.char) {
	reader := strings.NewReader(C.GoString(cstring))
	return goFastqToCFastq(reader)
}

/******************************************************************************

Genbank

******************************************************************************/

// goGenbankToCGenbank converts a slice of genbank.Genbank to a C.Genbank array
func goGenbankToCGenbank(gbs []genbank.Genbank) (*C.Genbank, int, *C.char) {
	if len(gbs) == 0 {
		return nil, 0, C.CString("No genbank records provided")
	}

	cGenbanks := (*C.Genbank)(C.malloc(C.size_t(len(gbs)) * C.size_t(unsafe.Sizeof(C.Genbank{}))))
	slice := (*[1<<30 - 1]C.Genbank)(unsafe.Pointer(cGenbanks))[:len(gbs):len(gbs)]

	for i, gb := range gbs {
		// Convert Meta
		slice[i].meta = C.GenbankMeta{
			date:                   C.CString(gb.Meta.Date),
			definition:             C.CString(gb.Meta.Definition),
			accession:              C.CString(gb.Meta.Accession),
			version:                C.CString(gb.Meta.Version),
			keywords:               C.CString(gb.Meta.Keywords),
			organism:               C.CString(gb.Meta.Organism),
			source:                 C.CString(gb.Meta.Source),
			taxonomy:               (**C.char)(C.malloc(C.size_t(len(gb.Meta.Taxonomy)) * C.size_t(unsafe.Sizeof(uintptr(0))))),
			taxonomy_count:         C.int(len(gb.Meta.Taxonomy)),
			origin:                 C.CString(gb.Meta.Origin),
			name:                   C.CString(gb.Meta.Name),
			sequence_hash:          C.CString(gb.Meta.SequenceHash),
			sequence_hash_function: C.CString(gb.Meta.SequenceHashFunction),
		}

		// Convert Taxonomy
		taxonomySlice := (*[1 << 30]*C.char)(unsafe.Pointer(slice[i].meta.taxonomy))[:len(gb.Meta.Taxonomy):len(gb.Meta.Taxonomy)]
		for j, taxon := range gb.Meta.Taxonomy {
			taxonomySlice[j] = C.CString(taxon)
		}

		// Convert Locus
		slice[i].meta.locus = C.GenbankLocus{
			name:              C.CString(gb.Meta.Locus.Name),
			sequence_length:   C.CString(gb.Meta.Locus.SequenceLength),
			molecule_type:     C.CString(gb.Meta.Locus.MoleculeType),
			genbank_division:  C.CString(gb.Meta.Locus.GenbankDivision),
			modification_date: C.CString(gb.Meta.Locus.ModificationDate),
			sequence_coding:   C.CString(gb.Meta.Locus.SequenceCoding),
			circular:          C.int(boolToInt(gb.Meta.Locus.Circular)),
		}

		// Convert References
		slice[i].meta.references = (*C.GenbankReference)(C.malloc(C.size_t(len(gb.Meta.References)) * C.size_t(unsafe.Sizeof(C.GenbankReference{}))))
		slice[i].meta.reference_count = C.int(len(gb.Meta.References))
		refSlice := (*[1 << 30]C.GenbankReference)(unsafe.Pointer(slice[i].meta.references))[:len(gb.Meta.References):len(gb.Meta.References)]
		for j, ref := range gb.Meta.References {
			refSlice[j] = C.GenbankReference{
				authors:    C.CString(ref.Authors),
				title:      C.CString(ref.Title),
				journal:    C.CString(ref.Journal),
				pub_med:    C.CString(ref.PubMed),
				remark:     C.CString(ref.Remark),
				range_:     C.CString(ref.Range),
				consortium: C.CString(ref.Consortium),
			}
		}

		// Convert BaseCount
		slice[i].meta.base_counts = (*C.GenbankBaseCount)(C.malloc(C.size_t(len(gb.Meta.BaseCount)) * C.size_t(unsafe.Sizeof(C.GenbankBaseCount{}))))
		slice[i].meta.base_count_count = C.int(len(gb.Meta.BaseCount))
		baseCountSlice := (*[1 << 30]C.GenbankBaseCount)(unsafe.Pointer(slice[i].meta.base_counts))[:len(gb.Meta.BaseCount):len(gb.Meta.BaseCount)]
		for j, bc := range gb.Meta.BaseCount {
			baseCountSlice[j] = C.GenbankBaseCount{
				base:  C.char(bc.Base[0]),
				count: C.int(bc.Count),
			}
		}

		// Convert Other
		slice[i].meta.other_keys = (**C.char)(C.malloc(C.size_t(len(gb.Meta.Other)) * C.size_t(unsafe.Sizeof(uintptr(0)))))
		slice[i].meta.other_values = (**C.char)(C.malloc(C.size_t(len(gb.Meta.Other)) * C.size_t(unsafe.Sizeof(uintptr(0)))))
		slice[i].meta.other_count = C.int(len(gb.Meta.Other))
		otherKeysSlice := (*[1 << 30]*C.char)(unsafe.Pointer(slice[i].meta.other_keys))[:len(gb.Meta.Other):len(gb.Meta.Other)]
		otherValuesSlice := (*[1 << 30]*C.char)(unsafe.Pointer(slice[i].meta.other_values))[:len(gb.Meta.Other):len(gb.Meta.Other)]
		j := 0
		for k, v := range gb.Meta.Other {
			otherKeysSlice[j] = C.CString(k)
			otherValuesSlice[j] = C.CString(v)
			j++
		}

		// Convert Features
		slice[i].features = (*C.GenbankFeature)(C.malloc(C.size_t(len(gb.Features)) * C.size_t(unsafe.Sizeof(C.GenbankFeature{}))))
		slice[i].feature_count = C.int(len(gb.Features))
		featureSlice := (*[1 << 30]C.GenbankFeature)(unsafe.Pointer(slice[i].features))[:len(gb.Features):len(gb.Features)]
		for j, feature := range gb.Features {
			featureSlice[j] = C.GenbankFeature{
				type_:                  C.CString(feature.Type),
				description:            C.CString(feature.Description),
				sequence_hash:          C.CString(feature.SequenceHash),
				sequence_hash_function: C.CString(feature.SequenceHashFunction),
				sequence:               C.CString(feature.Sequence),
			}

			// Convert Attributes
			featureSlice[j].attribute_keys = (**C.char)(C.malloc(C.size_t(len(feature.Attributes)) * C.size_t(unsafe.Sizeof(uintptr(0)))))
			featureSlice[j].attribute_values = (***C.char)(C.malloc(C.size_t(len(feature.Attributes)) * C.size_t(unsafe.Sizeof(uintptr(0)))))
			featureSlice[j].attribute_value_counts = (*C.int)(C.malloc(C.size_t(len(feature.Attributes)) * C.size_t(unsafe.Sizeof(C.int(0)))))
			featureSlice[j].attribute_count = C.int(len(feature.Attributes))

			attrKeysSlice := (*[1 << 30]*C.char)(unsafe.Pointer(featureSlice[j].attribute_keys))[:len(feature.Attributes):len(feature.Attributes)]
			attrValuesSlice := (*[1 << 30]**C.char)(unsafe.Pointer(featureSlice[j].attribute_values))[:len(feature.Attributes):len(feature.Attributes)]
			attrValueCountsSlice := (*[1 << 30]C.int)(unsafe.Pointer(featureSlice[j].attribute_value_counts))[:len(feature.Attributes):len(feature.Attributes)]

			k := 0
			for key, values := range feature.Attributes {
				attrKeysSlice[k] = C.CString(key)
				attrValuesSlice[k] = (**C.char)(C.malloc(C.size_t(len(values)) * C.size_t(unsafe.Sizeof(uintptr(0)))))
				attrValueCountsSlice[k] = C.int(len(values))
				valueSlice := (*[1 << 30]*C.char)(unsafe.Pointer(attrValuesSlice[k]))[:len(values):len(values)]
				for l, val := range values {
					valueSlice[l] = C.CString(val)
				}
				k++
			}

			// Convert Location
			featureSlice[j].location = convertLocation(feature.Location)
		}

		// Convert Sequence
		slice[i].sequence = C.CString(gb.Sequence)
	}

	return cGenbanks, len(gbs), nil
}

// Convert Location
func convertLocation(location genbank.Location) C.GenbankLocation {
	cLocation := C.GenbankLocation{
		start:               C.int(location.Start),
		end:                 C.int(location.End),
		complement:          C.int(boolToInt(location.Complement)),
		join:                C.int(boolToInt(location.Join)),
		five_prime_partial:  C.int(boolToInt(location.FivePrimePartial)),
		three_prime_partial: C.int(boolToInt(location.ThreePrimePartial)),
		gbk_location_string: C.CString(location.GbkLocationString),
		sub_locations:       nil,
		sub_locations_count: 0,
	}

	if len(location.SubLocations) > 0 {
		cSubLocations := make([]C.GenbankLocation, len(location.SubLocations))
		for i, subLocation := range location.SubLocations {
			cSubLocations[i] = convertLocation(subLocation)
		}
		cLocation.sub_locations = &cSubLocations[0]
		cLocation.sub_locations_count = C.int(len(cSubLocations))
	}

	return cLocation
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

//export ParseGenbankFromCFile
func ParseGenbankFromCFile(cfile *C.FILE) (*C.Genbank, int, *C.char) {
	reader := readerFromCFile(cfile)
	parser := bio.NewGenbankParser(reader)
	genbanks, err := parser.Parse()
	if err != nil {
		return nil, len(genbanks), C.CString(err.Error())
	}
	return goGenbankToCGenbank(genbanks)
}

//export ParseGenbankFromCString
func ParseGenbankFromCString(cstring *C.char) (*C.Genbank, int, *C.char) {
	reader := strings.NewReader(C.GoString(cstring))
	parser := bio.NewGenbankParser(reader)
	genbanks, err := parser.Parse()
	if err != nil {
		return nil, len(genbanks), C.CString(err.Error())
	}
	return goGenbankToCGenbank(genbanks)
}

/******************************************************************************
Aug 28, 2024

Clone Package Functions

******************************************************************************/

//export CutWithEnzymeByName
func CutWithEnzymeByName(part C.Part, directional C.int, name *C.char, methylated C.int) (*C.Fragment, C.int, *C.char) {
	goPart := clone.Part{
		Sequence: C.GoString(part.sequence),
		Circular: part.circular != 0,
	}
	fragments, err := clone.CutWithEnzymeByName(goPart, directional != 0, C.GoString(name), methylated != 0)
	if err != nil {
		return nil, 0, C.CString(err.Error())
	}

	cFragments := (*C.Fragment)(C.malloc(C.size_t(len(fragments)) * C.size_t(unsafe.Sizeof(C.Fragment{}))))
	slice := (*[1<<30 - 1]C.Fragment)(unsafe.Pointer(cFragments))[:len(fragments):len(fragments)]

	for i, frag := range fragments {
		slice[i].sequence = C.CString(frag.Sequence)
		slice[i].forward_overhang = C.CString(frag.ForwardOverhang)
		slice[i].reverse_overhang = C.CString(frag.ReverseOverhang)
	}

	return cFragments, C.int(len(fragments)), nil
}

//export Ligate
func Ligate(fragments *C.Fragment, fragmentCount C.int, circular C.int) (*C.char, *C.int, C.int, *C.char) {
	goFragments := make([]clone.Fragment, int(fragmentCount))
	slice := (*[1<<30 - 1]C.Fragment)(unsafe.Pointer(fragments))[:fragmentCount:fragmentCount]

	for i := 0; i < int(fragmentCount); i++ {
		goFragments[i] = clone.Fragment{
			Sequence:        C.GoString(slice[i].sequence),
			ForwardOverhang: C.GoString(slice[i].forward_overhang),
			ReverseOverhang: C.GoString(slice[i].reverse_overhang),
		}
	}

	ligation, ligationPattern, err := clone.Ligate(goFragments, circular != 0)
	if err != nil {
		return nil, nil, 0, C.CString(err.Error())
	}

	cLigation := C.CString(ligation)
	cLigationPattern := (*C.int)(C.malloc(C.size_t(len(ligationPattern)) * C.sizeof_int))
	cLigationPatternSlice := (*[1<<30 - 1]C.int)(unsafe.Pointer(cLigationPattern))[:len(ligationPattern):len(ligationPattern)]
	for i, v := range ligationPattern {
		cLigationPatternSlice[i] = C.int(v)
	}

	return cLigation, cLigationPattern, C.int(len(ligationPattern)), nil
}

//export GoldenGate
func GoldenGate(sequences *C.Part, sequenceCount C.int, cuttingEnzymeName *C.char, methylated C.int) (*C.char, *C.int, C.int, *C.char) {
	goParts := make([]clone.Part, int(sequenceCount))
	slice := (*[1<<30 - 1]C.Part)(unsafe.Pointer(sequences))[:sequenceCount:sequenceCount]

	for i := 0; i < int(sequenceCount); i++ {
		goParts[i] = clone.Part{
			Sequence: C.GoString(slice[i].sequence),
			Circular: slice[i].circular != 0,
		}
	}

	// Look up the cutting enzyme by name
	enzymeName := C.GoString(cuttingEnzymeName)
	cuttingEnzyme, ok := clone.DefaultEnzymes[enzymeName]
	if !ok {
		return nil, nil, 0, C.CString(fmt.Sprintf("Unknown enzyme: %s", enzymeName))
	}

	result, pattern, err := clone.GoldenGate(goParts, cuttingEnzyme, methylated != 0)
	if err != nil {
		return nil, nil, 0, C.CString(err.Error())
	}

	cResult := C.CString(result)
	cPattern := (*C.int)(C.malloc(C.size_t(len(pattern)) * C.sizeof_int))
	cPatternSlice := (*[1<<30 - 1]C.int)(unsafe.Pointer(cPattern))[:len(pattern):len(pattern)]
	for i, v := range pattern {
		cPatternSlice[i] = C.int(v)
	}

	return cResult, cPattern, C.int(len(pattern)), nil
}

/******************************************************************************
Aug 28, 2024

Fragment Package Functions

******************************************************************************/

//export SetEfficiency
func SetEfficiency(overhangs **C.char, overhangCount C.int) C.double {
	goOverhangs := make([]string, int(overhangCount))
	slice := (*[1<<30 - 1]*C.char)(unsafe.Pointer(overhangs))[:overhangCount:overhangCount]

	for i := 0; i < int(overhangCount); i++ {
		goOverhangs[i] = C.GoString(slice[i])
	}

	return C.double(fragment.SetEfficiency(goOverhangs))
}

//export NextOverhangs
func NextOverhangs(currentOverhangs **C.char, overhangCount C.int) (**C.char, *C.double, C.int, *C.char) {
	goCurrentOverhangs := make([]string, int(overhangCount))
	slice := (*[1<<30 - 1]*C.char)(unsafe.Pointer(currentOverhangs))[:overhangCount:overhangCount]

	for i := 0; i < int(overhangCount); i++ {
		goCurrentOverhangs[i] = C.GoString(slice[i])
	}

	nextOverhangs, efficiencies := fragment.NextOverhangs(goCurrentOverhangs)

	cNextOverhangs := (**C.char)(C.malloc(C.size_t(len(nextOverhangs)) * C.size_t(unsafe.Sizeof(uintptr(0)))))
	cNextOverhangsSlice := (*[1<<30 - 1]*C.char)(unsafe.Pointer(cNextOverhangs))[:len(nextOverhangs):len(nextOverhangs)]

	cEfficiencies := (*C.double)(C.malloc(C.size_t(len(efficiencies)) * C.sizeof_double))
	cEfficienciesSlice := (*[1<<30 - 1]C.double)(unsafe.Pointer(cEfficiencies))[:len(efficiencies):len(efficiencies)]

	for i, overhang := range nextOverhangs {
		cNextOverhangsSlice[i] = C.CString(overhang)
		cEfficienciesSlice[i] = C.double(efficiencies[i])
	}

	return cNextOverhangs, cEfficiencies, C.int(len(nextOverhangs)), nil
}

//export NextOverhang
func NextOverhang(currentOverhangs **C.char, overhangCount C.int) *C.char {
	goCurrentOverhangs := make([]string, int(overhangCount))
	slice := (*[1<<30 - 1]*C.char)(unsafe.Pointer(currentOverhangs))[:overhangCount:overhangCount]

	for i := 0; i < int(overhangCount); i++ {
		goCurrentOverhangs[i] = C.GoString(slice[i])
	}

	return C.CString(fragment.NextOverhang(goCurrentOverhangs))
}

//export FragmentSequence
func FragmentSequence(sequence *C.char, minFragmentSize C.int, maxFragmentSize C.int, excludeOverhangs **C.char, excludeOverhangCount C.int) (**C.char, C.int, C.double, *C.char) {
	goSequence := C.GoString(sequence)
	goExcludeOverhangs := make([]string, int(excludeOverhangCount))
	slice := (*[1<<30 - 1]*C.char)(unsafe.Pointer(excludeOverhangs))[:excludeOverhangCount:excludeOverhangCount]

	for i := 0; i < int(excludeOverhangCount); i++ {
		goExcludeOverhangs[i] = C.GoString(slice[i])
	}

	fragments, efficiency, err := fragment.Fragment(goSequence, int(minFragmentSize), int(maxFragmentSize), goExcludeOverhangs)
	if err != nil {
		return nil, 0, 0, C.CString(err.Error())
	}

	cFragments := (**C.char)(C.malloc(C.size_t(len(fragments)) * C.size_t(unsafe.Sizeof(uintptr(0)))))
	cFragmentsSlice := (*[1<<30 - 1]*C.char)(unsafe.Pointer(cFragments))[:len(fragments):len(fragments)]

	for i, frag := range fragments {
		cFragmentsSlice[i] = C.CString(frag)
	}

	return cFragments, C.int(len(fragments)), C.double(efficiency), nil
}

//export FragmentSequenceWithOverhangs
func FragmentSequenceWithOverhangs(sequence *C.char, minFragmentSize C.int, maxFragmentSize C.int, excludeOverhangs **C.char, excludeOverhangCount C.int, includeOverhangs **C.char, includeOverhangCount C.int) (**C.char, C.int, C.double, *C.char) {
	goSequence := C.GoString(sequence)
	goExcludeOverhangs := make([]string, int(excludeOverhangCount))
	goIncludeOverhangs := make([]string, int(includeOverhangCount))

	excludeSlice := (*[1<<30 - 1]*C.char)(unsafe.Pointer(excludeOverhangs))[:excludeOverhangCount:excludeOverhangCount]
	includeSlice := (*[1<<30 - 1]*C.char)(unsafe.Pointer(includeOverhangs))[:includeOverhangCount:includeOverhangCount]

	for i := 0; i < int(excludeOverhangCount); i++ {
		goExcludeOverhangs[i] = C.GoString(excludeSlice[i])
	}

	for i := 0; i < int(includeOverhangCount); i++ {
		goIncludeOverhangs[i] = C.GoString(includeSlice[i])
	}

	fragments, efficiency, err := fragment.FragmentWithOverhangs(goSequence, int(minFragmentSize), int(maxFragmentSize), goExcludeOverhangs, goIncludeOverhangs)
	if err != nil {
		return nil, 0, 0, C.CString(err.Error())
	}

	cFragments := (**C.char)(C.malloc(C.size_t(len(fragments)) * C.size_t(unsafe.Sizeof(uintptr(0)))))
	cFragmentsSlice := (*[1<<30 - 1]*C.char)(unsafe.Pointer(cFragments))[:len(fragments):len(fragments)]

	for i, frag := range fragments {
		cFragmentsSlice[i] = C.CString(frag)
	}

	return cFragments, C.int(len(fragments)), C.double(efficiency), nil
}

//export RecursiveFragmentSequence
func RecursiveFragmentSequence(sequence *C.char, maxCodingSizeOligo C.int, assemblyPattern *C.int, patternCount C.int, excludeOverhangs **C.char, excludeCount C.int, includeOverhangs **C.char, includeCount C.int) (*C.Assembly, *C.char) {
	goSequence := C.GoString(sequence)
	goAssemblyPattern := make([]int, patternCount)
	goExcludeOverhangs := make([]string, excludeCount)
	goIncludeOverhangs := make([]string, includeCount)
	patternSlice := (*[1<<30 - 1]C.int)(unsafe.Pointer(assemblyPattern))[:patternCount:patternCount]
	excludeSlice := (*[1<<30 - 1]*C.char)(unsafe.Pointer(excludeOverhangs))[:excludeCount:excludeCount]
	includeSlice := (*[1<<30 - 1]*C.char)(unsafe.Pointer(includeOverhangs))[:includeCount:includeCount]
	for i := 0; i < int(patternCount); i++ {
		goAssemblyPattern[i] = int(patternSlice[i])
	}
	for i := 0; i < int(excludeCount); i++ {
		goExcludeOverhangs[i] = C.GoString(excludeSlice[i])
	}
	for i := 0; i < int(includeCount); i++ {
		goIncludeOverhangs[i] = C.GoString(includeSlice[i])
	}
	assembly, err := fragment.RecursiveFragment(goSequence, int(maxCodingSizeOligo), goAssemblyPattern, goExcludeOverhangs, goIncludeOverhangs)
	if err != nil {
		return nil, C.CString(err.Error())
	}
	return convertAssemblyToC(assembly), nil
}

func convertAssemblyToC(assembly fragment.Assembly) *C.Assembly {
	cAssembly := (*C.Assembly)(C.malloc(C.size_t(unsafe.Sizeof(C.Assembly{}))))
	cAssembly.sequence = C.CString(assembly.Sequence)

	// Convert fragments
	cFragments := (**C.char)(C.malloc(C.size_t(len(assembly.Fragments)) * C.size_t(unsafe.Sizeof(uintptr(0)))))
	cFragmentsSlice := (*[1<<30 - 1]*C.char)(unsafe.Pointer(cFragments))[:len(assembly.Fragments):len(assembly.Fragments)]
	for i, frag := range assembly.Fragments {
		cFragmentsSlice[i] = C.CString(frag)
	}
	cAssembly.fragments = cFragments
	cAssembly.fragmentCount = C.int(len(assembly.Fragments))

	cAssembly.efficiency = C.double(assembly.Efficiency)

	// Convert sub-assemblies recursively
	if len(assembly.SubAssemblies) > 0 {
		cSubAssemblies := (*C.Assembly)(C.malloc(C.size_t(len(assembly.SubAssemblies)) * C.size_t(unsafe.Sizeof(C.Assembly{}))))
		cSubAssembliesSlice := (*[1<<30 - 1]C.Assembly)(unsafe.Pointer(cSubAssemblies))[:len(assembly.SubAssemblies):len(assembly.SubAssemblies)]
		for i, subAssembly := range assembly.SubAssemblies {
			cSubAssembliesSlice[i] = *convertAssemblyToC(subAssembly)
		}
		cAssembly.subAssemblies = unsafe.Pointer(cSubAssemblies)
		cAssembly.subAssemblyCount = C.int(len(assembly.SubAssemblies))
	} else {
		cAssembly.subAssemblies = nil
		cAssembly.subAssemblyCount = 0
	}

	return cAssembly
}

/******************************************************************************

main.go

******************************************************************************/

func main() {}
