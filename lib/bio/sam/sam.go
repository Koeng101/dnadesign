/*
Package sam implements a SAM file parser and writer.

SAM is a tab-delimited text format for storing DNA/RNA sequence alignment data.
It is the most widely used alignment format, complementing its binary
equivalent, BAM, which stores the same data in a compressed format.

DNA sequencing works in the following way:

  - DNA is read in with some raw signal format from the sequencer machine.
  - Raw signal is converted to fastq reads using basecalling software.
  - Fastq reads are aligned to target template, producing SAM files.
  - SAM files are used to answer bioinformatic queries.

This parser allows parsing and writing of SAM files in Go. Unlike other SAM
parsers in Golang, we aim to be as close to underlying data types as possible,
with a goal of being as simple as possible, and no simpler.

Paper: https://doi.org/10.1093%2Fbioinformatics%2Fbtp352
Spec: http://samtools.github.io/hts-specs/SAMv1.pdf
Spec(locally): `dnadesign/lib/bio/sam/SAMv1.pdf`
*/
package sam

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

const DefaultMaxLineSize int = 1024 * 32 * 2 // // 32kB is a magic number often used by the Go stdlib for parsing. We multiply it by two.

// Each header in a SAM file begins with an @ followed by a two letter record
// code type. Each line is tab delimited, and contains TAG:VALUE pairs. HD, the
// first line, only occurs once, while SQ, RG, and PG can appear multiple
// times. Finally, @CO contains user generated comments.
//
// For more information, check section 1.3 of the reference document.
type Header struct {
	HD map[string]string   // File-level metadata. Optional. If present, there must be only one @HD line and it must be the first line of the file.
	SQ []map[string]string // Reference sequence dictionary. The order of @SQ lines defines the alignment sorting order.
	RG []map[string]string // Read group. Unordered multiple @RG lines are allowed.
	PG []map[string]string // Program.
	CO []string            // One-line text comment. Unordered multiple @CO lines are allowed. UTF-8 encoding may be used.
}

// headerWriteHelper helps write SAM headers in an ordered way.
func headerWriteHelper(sb io.StringWriter, headerString string, headerMap map[string]string, orderedKeys []string) {
	_, _ = sb.WriteString(headerString)
	// Write orderedKeys first, if they exist
	for _, key := range orderedKeys {
		if value, exists := headerMap[key]; exists {
			_, _ = sb.WriteString(fmt.Sprintf("\t%s:%s", key, value))
		}
	}
	// Write the remaining key-value pairs
	for key, value := range headerMap {
		// Skip if the key is one of the specific keys
		var skip bool
		for _, orderedKey := range orderedKeys {
			if key == orderedKey {
				skip = true
			}
		}
		if skip {
			continue
		}
		_, _ = sb.WriteString(fmt.Sprintf("\t%s:%s", key, value))
	}
	_, _ = sb.WriteString("\n")
}

// WriteTo writes a SAM header to an io.Writer.
func (header *Header) WriteTo(w io.Writer) (int64, error) {
	// Here we write the header into a SAM file. Please check the official
	// documentation for the meaning of each tag used as ordered keys.
	// Here, we iterate through each, and write it to a file.
	var sb strings.Builder
	if len(header.HD) > 0 {
		headerWriteHelper(&sb, "@HD", header.HD, []string{"VN", "SO", "GO", "SS"})
	}
	for _, sq := range header.SQ {
		headerWriteHelper(&sb, "@SQ", sq, []string{"SN", "LN", "AH", "AN", "AS", "DS", "M5", "SP", "TP", "UR"})
	}
	for _, rg := range header.RG {
		headerWriteHelper(&sb, "@RG", rg, []string{"ID", "BC", "CN", "DS", "DT", "FO", "KS", "LB", "PG", "PI", "PL", "PM", "PU", "SM"})
	}
	for _, pg := range header.PG {
		headerWriteHelper(&sb, "@PG", pg, []string{"ID", "PN", "VN", "CL", "PP", "DS"})
	}
	for _, co := range header.CO {
		_, _ = sb.WriteString(fmt.Sprintf("@CO %s\n", co))
	}

	newWrittenBytes, err := w.Write([]byte(sb.String()))
	return int64(newWrittenBytes), err
}

// Validate validates that the header has all required information, as
// described in the SAMv1 specification document. Not implemented yet.
func (header *Header) Validate() error {
	/* The following rules apply:
	1. @HD.VN: Format version. Accepted format: /^[0-9]+\.[0-9]+$/.
	2. @HD.SO: Valid values: unknown (default), unsorted, queryname and coordinate
	3. @HD.GO: Valid values: none (default), query (alignments are grouped by QNAME), and reference (alignments are grouped by RNAME/POS)
	4. @HD.SS: Regular expression: (coordinate|queryname|unsorted)(:[A-Za-z0-9_-]+)+
	5. @SQ.SN: Regular expression: [:rname:^*=][:rname:]*
	6. @SQ.SN/AN: The SN tags and all individual AN names in all @SQ lines must be distinct
	7. @SQ.LN: Reference sequence length. Range: [1, 2^31 − 1]
	8. @SQ.AN: Regular expression: name(,name)* where name is [:rname:^*=][:rname:]* (definition of 6)
	9. @SQ.TP: Valid values: linear (default) and circular
	10. @RG.ID: Each @RG line must have a unique ID
	11. @RG.DT: Date the run was produced (ISO8601 date or date/time).
	12. @RG.FO: Format: /\*|[ACMGRSVTWYHKDBN]+/
	13. @RG.PL: Valid values: CAPILLARY, DNBSEQ (MGI/BGI), ELEMENT, HELICOS, ILLUMINA, IONTORRENT, LS454, ONT (Oxford Nanopore), PACBIO (Pacific Bio-sciences), SOLID, and ULTIMA
	14. @PG.ID: Each @PG line must have a unique ID.
	15. @PG.PP: Previous @PG-ID. Must match another @PG header’s ID tag. @PG records may be chained using PP tag, with the last record in the chain having no PP tag
	*/

	// Validate @HD tags
	if len(header.HD) > 0 {
		// Accessing HD map directly as it's not a function returning two values
		hd := header.HD

		// 1. @HD VN
		if vn, ok := hd["VN"]; ok {
			matched, _ := regexp.MatchString(`^[0-9]+\.[0-9]+$`, vn)
			if !matched {
				return fmt.Errorf("Invalid format for @HD VN. Accepted format: /^[0-9]+\\.[0-9]+$/.\nGot: %s", vn)
			}
		}
		// 2. @HD SO
		if so, ok := hd["SO"]; ok {
			validValues := map[string]bool{"unknown": true, "unsorted": true, "queryname": true, "coordinate": true}
			if _, valid := validValues[so]; !valid {
				return fmt.Errorf("Invalid value for @HD SO. Valid values: unknown (default), unsorted, queryname and coordinate. Got: %s", so)
			}
		}
		// 3. @HD GO
		if goTag, ok := hd["GO"]; ok {
			validValues := map[string]bool{"none": true, "query": true, "reference": true}
			if _, valid := validValues[goTag]; !valid {
				return fmt.Errorf("Invalid value for @HD GO. Valid values: none (default), query (alignments are grouped by QNAME), and reference (alignments are grouped by RNAME/POS). Got: %s", goTag)
			}
		}
		// 4. @HD SS
		if ss, ok := hd["SS"]; ok {
			matched, _ := regexp.MatchString(`(coordinate|queryname|unsorted)(:[A-Za-z0-9_-]+)+`, ss)
			if !matched {
				return fmt.Errorf("Invalid format for @HD SS. Needs to match: Regular expression: (coordinate|queryname|unsorted)(:[A-Za-z0-9_-]+)+\nGot: %s", ss)
			}
		}
	}

	// Validate @SQ tags
	snMap := make(map[string]bool)
	for _, sq := range header.SQ {
		// 5. @SQ SN
		if sn, ok := sq["SN"]; ok {
			// [:rname:^*=][:rname:]* isn't actually a valid regexp, so I'm not
			// sure why they've used this as the definition. We skip this check
			// because it doesn't make much sense.
			if snMap[sn] {
				return fmt.Errorf("Non-unique @SQ SN: %s", sn)
			}
			snMap[sn] = true
		}
		// 7. @SQ LN
		if ln, ok := sq["LN"]; ok {
			lnInt, err := strconv.Atoi(ln)
			if err != nil || lnInt < 1 || lnInt > 2147483647 {
				return fmt.Errorf("Invalid value for @SQ LN. Range: [1, 231 − 1], Got: %d", lnInt)
			}
		}
		// 9. @SQ TP
		if tp, ok := sq["TP"]; ok {
			validValues := map[string]bool{"linear": true, "circular": true}
			if _, valid := validValues[tp]; !valid {
				return fmt.Errorf("Invalid value for @SQ TP. Valid values: linear (default) and circular, Got: %s", tp)
			}
		}
	}

	// Validate @RG tags
	rgIDMap := make(map[string]bool)
	rgFoRegexp := regexp.MustCompile(`\*|[ACMGRSVTWYHKDBN]+`)
	for _, rg := range header.RG {
		// 10. @RG ID
		if id, ok := rg["ID"]; ok {
			if rgIDMap[id] {
				return fmt.Errorf("Non-unique @RG ID. Got: %s", id)
			}
			rgIDMap[id] = true
		}
		// 12. @RG FO
		if fo, ok := rg["FO"]; ok {
			matched := rgFoRegexp.MatchString(fo)
			if !matched {
				return fmt.Errorf("Invalid format for @RG FO. Required regexp format: /\\*|[ACMGRSVTWYHKDBN]+/\nGot: %s", fo)
			}
		}
		// 13. @RG PL
		if pl, ok := rg["PL"]; ok {
			validValues := map[string]bool{
				"CAPILLARY": true, "DNBSEQ": true, "ELEMENT": true, "HELICOS": true, "ILLUMINA": true,
				"IONTORRENT": true, "LS454": true, "ONT": true, "PACBIO": true, "SOLID": true, "ULTIMA": true,
			}
			if _, valid := validValues[pl]; !valid {
				return fmt.Errorf("Invalid value for @RG PL. Valid values: CAPILLARY, DNBSEQ (MGI/BGI), ELEMENT, HELICOS, ILLUMINA, IONTORRENT, LS454, ONT (Oxford Nanopore), PACBIO (Pacific Bio-sciences), SOLID, and ULTIMA. Got: %s", pl)
			}
		}
	}

	// Validate @PG tags
	pgIDMap := make(map[string]bool)
	for _, pg := range header.PG {
		// 14. @PG ID
		if id, ok := pg["ID"]; ok {
			if pgIDMap[id] {
				return fmt.Errorf("Non-unique @PG ID. Got: %s", id)
			}
			pgIDMap[id] = true
		}
	}
	return nil
}

// Optional fields in SAM alignments are structured as TAG:TYPE:DATA, where
// the type identifiers the typing of the data.
//
// For more information, check section 1.5 of http://samtools.github.io/hts-specs/SAMv1.pdf.
type Optional struct {
	Tag  string // Tag is typically a two letter tag corresponding to what the optional represents.
	Type rune   // The type may be one of A (character), B (general array), f (real number), H (hexadecimal array), i (integer), or Z (string).
	Data string // Optional data
}

// Each alignment is a single line of a SAM file, representing a linear
// alignment of a segment, consisting of 11 or more tab delimited fields. The
// 11 fields (QNAME -> QUAL) are always available (if the data isn't there, a
// placeholder '0' or '*' is used instead), with additional optional fields
// following.
//
// For more information, check section 1.4 of the reference document.
type Alignment struct {
	QNAME     string     // Query template NAME
	FLAG      uint16     // bitwise FLAG
	RNAME     string     // References sequence NAME
	POS       int32      // 1- based leftmost mapping POSition
	MAPQ      byte       // MAPping Quality
	CIGAR     string     // CIGAR string
	RNEXT     string     // Ref. name of the mate/next read
	PNEXT     int32      // Position of the mate/next read
	TLEN      int32      // observed Template LENgth
	SEQ       string     // segment SEQuence
	QUAL      string     // ASCII of Phred-scaled base QUALity+33
	Optionals []Optional // Map of TAG to {TYPE:DATA}
}

// Alignment_WriteTo implements the io.WriterTo interface. It writes an
// alignment line.
func (alignment *Alignment) WriteTo(w io.Writer) (int64, error) {
	var sb strings.Builder
	_, _ = sb.WriteString(fmt.Sprintf("%s\t%d\t%s\t%d\t%d\t%s\t%s\t%d\t%d\t%s\t%s", alignment.QNAME, alignment.FLAG, alignment.RNAME, alignment.POS, alignment.MAPQ, alignment.CIGAR, alignment.RNEXT, alignment.PNEXT, alignment.TLEN, alignment.SEQ, alignment.QUAL))
	for _, optional := range alignment.Optionals {
		_, _ = sb.WriteString(fmt.Sprintf("\t%s:%c:%s", optional.Tag, optional.Type, optional.Data))
	}
	_, _ = sb.WriteString("\n")
	newWrittenBytes, err := w.Write([]byte(sb.String()))
	return int64(newWrittenBytes), err
}

// Alignment_Validate validates an alignment as valid, given the REGEXP/range
// defined in the SAM document. Not implemented yet.
func (alignment *Alignment) Validate() error {
	/* The following rules apply:

	1 QNAME	String	[!-?A-~]{1,254}				Query template NAME
	2 FLAG	Int		[0, 216 − 1]				bitwise FLAG
	3 RNAME	String	\*|[:rname:∧*=][:rname:]*	Reference sequence NAME11
	4 POS	Int		[0, 231 − 1]				1-based leftmost mapping POSition
	5 MAPQ	Int		[0, 28 − 1]					MAPping Quality
	6 CIGAR	String	\*|([0-9]+[MIDNSHPX=])+		CIGAR string
	7 RNEXT	String	\*|=|[:rname:∧*=][:rname:]*	Reference name of the mate/next read
	8 PNEXT	Int		[0, 231 − 1]				Position of the mate/next read
	9 TLEN	Int		[−231 + 1, 231 − 1]			observed Template LENgth
	10 SEQ	String	\*|[A-Za-z=.]+				segment SEQuence
	11 QUAL	String	[!-~]+						ASCII of Phred-scaled base QUALity+33
	*/
	// 1. Validate QNAME
	qnameRegex := `^[!-?A-~]{1,254}$`
	if matched, _ := regexp.MatchString(qnameRegex, alignment.QNAME); !matched {
		return errors.New("Invalid QNAME: must match " + qnameRegex)
	}

	// 2. FLAG is validated through uint16 typing.

	// 3. Validate RNAME
	rnameRegex := `^\*|[:rname:^\*=][:rname:]*$`
	if matched, _ := regexp.MatchString(rnameRegex, alignment.RNAME); !matched {
		return errors.New("Invalid RNAME: must match " + rnameRegex)
	}

	// 4. Validate POS
	if alignment.POS < 0 || alignment.POS > 2147483647 { // 2^31 - 1
		return errors.New("Invalid POS: must be in range [0, 2147483647]")
	}

	// 5. MAPQ is validated through byte typing.

	// 6. Validate CIGAR
	cigarRegex := `^\*|([0-9]+[MIDNSHPX=])+$`
	if matched, _ := regexp.MatchString(cigarRegex, alignment.CIGAR); !matched {
		return errors.New("Invalid CIGAR: must match " + cigarRegex)
	}

	// 7. Validate RNEXT
	rnextRegex := `^\*|=\|[:rname:^\*=][:rname:]*$`
	if matched, _ := regexp.MatchString(rnextRegex, alignment.RNEXT); !matched {
		return errors.New("Invalid RNEXT: must match " + rnextRegex)
	}

	// 8. Validate PNEXT
	if alignment.PNEXT < 0 || alignment.PNEXT > 2147483647 { // 2^31 - 1
		return errors.New("Invalid PNEXT: must be in range [0, 2147483647]")
	}

	// 9. TLEN is validated through int32 typing.

	// 10. Validate SEQ
	seqRegex := `^\*|[A-Za-z=.]+$`
	if matched, _ := regexp.MatchString(seqRegex, alignment.SEQ); !matched {
		return errors.New("Invalid SEQ: must match " + seqRegex)
	}

	// 11. Validate QUAL
	qualRegex := `^[!-~]+$`
	if matched, _ := regexp.MatchString(qualRegex, alignment.QUAL); !matched {
		return errors.New("Invalid QUAL: must match " + qualRegex)
	}

	return nil
}

// Parser is a sam file parser that provide sample control over reading sam
// alignments. It should be initialized with NewParser.
type Parser struct {
	reader        bufio.Reader
	line          uint
	FileHeader    Header
	firstLine     string
	readFirstLine bool
}

// Header returns the parsed sam header.
func (p *Parser) Header() (*Header, error) {
	return &p.FileHeader, nil
}

func checkIfValidSamLine(lineBytes []byte) bool {
	return len(strings.Split(strings.TrimSpace(string(lineBytes)), "\t")) >= 11
}

// NewParser creates a parser from an io.Reader for sam data. For larger
// alignments, you will want to increase the maxLineSize.
func NewParser(r io.Reader, maxLineSize int) (*Parser, Header, error) {
	parser := &Parser{
		reader: *bufio.NewReaderSize(r, maxLineSize),
	}
	var header Header
	var hdParsed bool
	// Initialize header maps
	header.HD = make(map[string]string)
	header.SQ = []map[string]string{}
	header.RG = []map[string]string{}
	header.PG = []map[string]string{}
	header.CO = []string{}

	// We need to first read the header before returning the parser to the
	// user for analyzing alignments.
	for {
		lineBytes, err := parser.reader.ReadSlice('\n')
		line := strings.TrimSpace(string(lineBytes))
		if err != nil {
			// Check if we have an EOF, if we have a validSamLine, and we are
			// not parsing a header. We do not check EOF + header line without
			// any validSamLine because that is useless.
			//
			// This, on the other hand, will catch if we have a single line sam
			// file with an EOF at the end, like we often have in tests.
			if err == io.EOF && checkIfValidSamLine(lineBytes) && line[0] != '@' {
				parser.firstLine = line
				break
			}
			return parser, Header{}, err
		}
		parser.line++
		if len(line) == 0 {
			return parser, Header{}, fmt.Errorf("Line %d is empty. Empty lines are not allowed in headers.", parser.line)
		}
		// If this line is the start of the alignments, set the firstLine
		// into memory, and then break this loop.
		if line[0] != '@' {
			parser.firstLine = line
			break
		}
		values := strings.Split(line, "\t")
		if len(values) < 1 {
			return parser, Header{}, fmt.Errorf("Line %d should contain at least 1 value. Got: %d. Line text: %s", parser.line, len(values), line)
		}

		// If we haven't parsed HD, it is always the first line: lets parse it.
		if !hdParsed {
			if values[0] != "@HD" {
				return parser, Header{}, fmt.Errorf("First line (%d) should always contain @HD first. Line text: %s", parser.line, line)
			}
			// Now parse the rest of the HD header
			for _, value := range values[1:] {
				valueSplit := strings.Split(value, ":")
				header.HD[valueSplit[0]] = valueSplit[1]
			}
			hdParsed = true
			continue
		}

		// CO lines are unique in that they are just strings. So we try to parse them
		// first. We include the entire comment line for these.
		if values[0] == "@CO" {
			header.CO = append(header.CO, line)
			continue
		}

		// HD/CO lines have been successfully parsed, now we work on SQ, RG, and PG.
		// Luckily, each one has an identical form ( TAG:DATA ), so we can parse that
		// first and then just apply it to the respect top level tag.
		genericMap := make(map[string]string)
		for _, value := range values[1:] {
			valueSplit := strings.Split(value, ":")
			genericMap[valueSplit[0]] = valueSplit[1]
		}
		switch values[0] {
		case "@SQ":
			header.SQ = append(header.SQ, genericMap)
		case "@RG":
			header.RG = append(header.RG, genericMap)
		case "@PG":
			header.PG = append(header.PG, genericMap)
		default:
			return parser, Header{}, fmt.Errorf("Line %d should contain @SQ, @RG, @PG or @CO as top level tags, but they weren't found. Line text: %s", parser.line, line)
		}
	}
	parser.FileHeader = header
	return parser, header, nil
}

// Next parsers the next read from a parser. Returns an `io.EOF` upon EOF.
func (p *Parser) Next() (*Alignment, error) {
	var alignment Alignment
	var finalLine bool
	var line string

	// We need to handle the firstLine after the header, as well as EOF checks.
	if !p.readFirstLine {
		line = p.firstLine
		p.readFirstLine = true
	} else {
		lineBytes, err := p.reader.ReadSlice('\n')
		if err != nil {
			if err == io.EOF {
				// This checks if the EOF is at the end of a line. If there is a
				// final SAM line, skip the EOF till the next Next()
				if len(strings.Split(strings.TrimSpace(string(lineBytes)), "\t")) >= 11 {
					finalLine = true
				}
			}
		}
		if !finalLine {
			if err != nil {
				return nil, err
			}
		}
		line = strings.TrimSpace(string(lineBytes))
	}
	p.line++
	values := strings.Split(line, "\t")
	if len(values) < 11 {
		return nil, fmt.Errorf("Line %d had error: must have at least 11 tab-delimited values. Had %d", p.line, len(values))
	}
	alignment.QNAME = values[0]
	flag64, err := strconv.ParseUint(values[1], 10, 16) // convert string to uint16
	if err != nil {
		return nil, fmt.Errorf("Line %d had error: %s", p.line, err)
	}
	alignment.FLAG = uint16(flag64)
	alignment.RNAME = values[2]
	pos64, err := strconv.ParseInt(values[3], 10, 32) // convert string to int32
	if err != nil {
		return nil, fmt.Errorf("Line %d had error: %s", p.line, err)
	}
	alignment.POS = int32(pos64)
	mapq64, err := strconv.ParseUint(values[4], 10, 8) // convert string to uint8 (otherwise known as byte)
	if err != nil {
		return nil, fmt.Errorf("Line %d had error: %s", p.line, err)
	}
	alignment.MAPQ = uint8(mapq64)
	alignment.CIGAR = values[5]
	alignment.RNEXT = values[6]
	pnext64, err := strconv.ParseInt(values[7], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("Line %d had error: %s", p.line, err)
	}
	alignment.PNEXT = int32(pnext64)
	tlen64, err := strconv.ParseInt(values[8], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("Line %d had error: %s", p.line, err)
	}
	alignment.TLEN = int32(tlen64)
	alignment.SEQ = values[9]
	alignment.QUAL = values[10]

	var optionals []Optional
	for _, value := range values[11:] {
		valueSplit := strings.Split(value, ":")
		optionals = append(optionals, Optional{Tag: valueSplit[0], Type: rune(valueSplit[1][0]), Data: valueSplit[2]})
	}
	alignment.Optionals = optionals
	return &alignment, nil
}
