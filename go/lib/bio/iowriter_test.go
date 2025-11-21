package bio

import (
	"io"
	"testing"

	"github.com/koeng101/dnadesign/lib/bio/fasta"
	"github.com/koeng101/dnadesign/lib/bio/fastq"
	"github.com/koeng101/dnadesign/lib/bio/genbank"
	"github.com/koeng101/dnadesign/lib/bio/pileup"
	"github.com/koeng101/dnadesign/lib/bio/sam"
	"github.com/koeng101/dnadesign/lib/bio/slow5"
	"github.com/koeng101/dnadesign/lib/bio/uniprot"
)

func TestAllTypesImplementWriterTo(t *testing.T) {
	var _ io.WriterTo = &genbank.Genbank{}
	var _ io.WriterTo = &fasta.Record{}
	var _ io.WriterTo = &fastq.Read{}
	var _ io.WriterTo = &slow5.Read{}
	var _ io.WriterTo = &sam.Alignment{}
	var _ io.WriterTo = &pileup.Line{}
	var _ io.WriterTo = &uniprot.Entry{}
}
