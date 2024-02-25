package api

import (
	"github.com/koeng101/dnadesign/api/gen"
	"github.com/koeng101/dnadesign/lib/bio/genbank"
	"github.com/koeng101/dnadesign/lib/bio/pileup"
	"github.com/koeng101/dnadesign/lib/bio/slow5"
)

func ConvertPileupLineToGenPileupLine(pileupLine pileup.Line) gen.PileupLine {
	return gen.PileupLine{
		Position:      int(pileupLine.Position),
		Quality:       pileupLine.Quality,
		ReadCount:     int(pileupLine.ReadCount),
		ReadResults:   pileupLine.ReadResults,
		ReferenceBase: pileupLine.ReferenceBase,
		Sequence:      pileupLine.Sequence,
	}
}

func ConvertGenPileupLineToPileupLine(genPileupLine gen.PileupLine) pileup.Line {
	return pileup.Line{
		Sequence:      genPileupLine.Sequence,
		Position:      uint(genPileupLine.Position),
		ReferenceBase: genPileupLine.ReferenceBase,
		ReadCount:     uint(genPileupLine.ReadCount),
		ReadResults:   genPileupLine.ReadResults,
		Quality:       genPileupLine.Quality,
	}
}

//nolint:dupl
func ConvertSlow5ReadToRead(slow5Read gen.Slow5Read) slow5.Read {
	var read slow5.Read

	// Convert and assign each field
	read.ReadID = slow5Read.ReadID
	read.ReadGroupID = uint32(slow5Read.ReadGroupID)
	read.Digitisation = float64(slow5Read.Digitisation)
	read.Offset = float64(slow5Read.Offset)
	read.Range = float64(slow5Read.Range)
	read.SamplingRate = float64(slow5Read.SamplingRate)
	read.LenRawSignal = uint64(slow5Read.LenRawSignal)

	read.RawSignal = make([]int16, len(slow5Read.RawSignal))
	for i, v := range slow5Read.RawSignal {
		read.RawSignal[i] = int16(v)
	}

	// Auxiliary fields
	read.ChannelNumber = slow5Read.ChannelNumber
	read.MedianBefore = float64(slow5Read.MedianBefore)
	read.ReadNumber = int32(slow5Read.ReadNumber)
	read.StartMux = uint8(slow5Read.StartMux)
	read.StartTime = uint64(slow5Read.StartTime)
	read.EndReason = slow5Read.EndReason

	read.EndReasonMap = slow5Read.EndReasonMap

	return read
}

//nolint:dupl
func ConvertReadToSlow5Read(read slow5.Read) gen.Slow5Read {
	var slow5Read gen.Slow5Read

	// Convert and assign each field
	slow5Read.ReadID = read.ReadID
	slow5Read.ReadGroupID = int(read.ReadGroupID)
	slow5Read.Digitisation = float32(read.Digitisation)
	slow5Read.Offset = float32(read.Offset)
	slow5Read.Range = float32(read.Range)
	slow5Read.SamplingRate = float32(read.SamplingRate)
	slow5Read.LenRawSignal = int(read.LenRawSignal)

	slow5Read.RawSignal = make([]int, len(read.RawSignal))
	for i, v := range read.RawSignal {
		slow5Read.RawSignal[i] = int(v)
	}

	// Auxiliary fields
	slow5Read.ChannelNumber = read.ChannelNumber
	slow5Read.MedianBefore = float32(read.MedianBefore)
	slow5Read.ReadNumber = int(read.ReadNumber)
	slow5Read.StartMux = int(read.StartMux)
	slow5Read.StartTime = int(read.StartTime)
	slow5Read.EndReason = read.EndReason

	slow5Read.EndReasonMap = read.EndReasonMap

	return slow5Read
}

// ConvertGenbankToGenbankRecord converts a genbank.Genbank object to a gen.GenbankRecord object.
func ConvertGenbankToGenbankRecord(gb genbank.Genbank) gen.GenbankRecord {
	var features []gen.Feature
	for _, f := range gb.Features {
		features = append(features, gen.Feature{
			Type:        f.Type,
			Description: f.Description,
			Attributes:  f.Attributes,
			Sequence:    f.Sequence,
			Location:    convertLocationGenbankToGen(f.Location),
		})
	}

	return gen.GenbankRecord{
		Features: features,
		Meta:     convertMetaGenbankToGen(gb.Meta),
		Sequence: gb.Sequence,
	}
}

// ConvertGenbankRecordToGenbank converts a gen.GenbankRecord object to a genbank.Genbank object.
func ConvertGenbankRecordToGenbank(gbr gen.GenbankRecord) genbank.Genbank {
	var features []genbank.Feature
	for _, f := range gbr.Features {
		features = append(features, genbank.Feature{
			Type:        f.Type,
			Description: f.Description,
			Attributes:  f.Attributes,
			Sequence:    f.Sequence,
			Location:    convertLocationGenToGenbank(f.Location),
		})
	}

	return genbank.Genbank{
		Features: features,
		Meta:     convertMetaGenToGenbank(gbr.Meta),
		Sequence: gbr.Sequence,
	}
}

// Helper functions for converting Location and Meta types
func convertLocationGenbankToGen(loc genbank.Location) gen.Location {
	var subLocations []gen.Location
	for _, sl := range loc.SubLocations {
		subLocations = append(subLocations, convertLocationGenbankToGen(sl))
	}

	return gen.Location{
		Start:             loc.Start,
		End:               loc.End,
		Complement:        loc.Complement,
		Join:              loc.Join,
		FivePrimePartial:  loc.FivePrimePartial,
		ThreePrimePartial: loc.ThreePrimePartial,
		GbkLocationString: loc.GbkLocationString,
		SubLocations:      subLocations,
	}
}

func convertLocationGenToGenbank(loc gen.Location) genbank.Location {
	var subLocations []genbank.Location
	for _, sl := range loc.SubLocations {
		subLocations = append(subLocations, convertLocationGenToGenbank(sl))
	}

	return genbank.Location{
		Start:             loc.Start,
		End:               loc.End,
		Complement:        loc.Complement,
		Join:              loc.Join,
		FivePrimePartial:  loc.FivePrimePartial,
		ThreePrimePartial: loc.ThreePrimePartial,
		GbkLocationString: loc.GbkLocationString,
		SubLocations:      subLocations,
	}
}

func convertBaseCountsGenbankToGen(bcs []genbank.BaseCount) []gen.BaseCount {
	var genBcs []gen.BaseCount
	for _, bc := range bcs {
		genBcs = append(genBcs, gen.BaseCount{
			Base:  bc.Base,
			Count: bc.Count,
		})
	}
	return genBcs
}

func convertBaseCountsGenToGenbank(bcs []gen.BaseCount) []genbank.BaseCount {
	var genbankBcs []genbank.BaseCount
	for _, bc := range bcs {
		genbankBcs = append(genbankBcs, genbank.BaseCount{
			Base:  bc.Base,
			Count: bc.Count,
		})
	}
	return genbankBcs
}

func convertLocusGenbankToGen(locus genbank.Locus) gen.Locus {
	return gen.Locus{
		Circular:         locus.Circular,
		GenbankDivision:  locus.GenbankDivision,
		ModificationDate: locus.ModificationDate,
		MoleculeType:     locus.MoleculeType,
		Name:             locus.Name,
		SequenceCoding:   locus.SequenceCoding,
		SequenceLength:   locus.SequenceLength,
	}
}

func convertLocusGenToGenbank(locus gen.Locus) genbank.Locus {
	return genbank.Locus{
		Circular:         locus.Circular,
		GenbankDivision:  locus.GenbankDivision,
		ModificationDate: locus.ModificationDate,
		MoleculeType:     locus.MoleculeType,
		Name:             locus.Name,
		SequenceCoding:   locus.SequenceCoding,
		SequenceLength:   locus.SequenceLength,
	}
}

func convertReferencesGenbankToGen(refs []genbank.Reference) []gen.Reference {
	var genRefs []gen.Reference
	for _, ref := range refs {
		genRefs = append(genRefs, gen.Reference{
			Authors:    ref.Authors,
			Consortium: ref.Consortium,
			Journal:    ref.Journal,
			PubMed:     ref.PubMed,
			Range:      ref.Range,
			Remark:     ref.Remark,
			Title:      ref.Title,
		})
	}
	return genRefs
}

func convertReferencesGenToGenbank(refs []gen.Reference) []genbank.Reference {
	var genbankRefs []genbank.Reference
	for _, ref := range refs {
		genbankRefs = append(genbankRefs, genbank.Reference{
			Authors:    ref.Authors,
			Consortium: ref.Consortium,
			Journal:    ref.Journal,
			PubMed:     ref.PubMed,
			Range:      ref.Range,
			Remark:     ref.Remark,
			Title:      ref.Title,
		})
	}
	return genbankRefs
}

func convertMetaGenbankToGen(meta genbank.Meta) gen.Meta {
	return gen.Meta{
		Accession:  meta.Accession,
		BaseCount:  convertBaseCountsGenbankToGen(meta.BaseCount),
		Date:       meta.Date,
		Definition: meta.Definition,
		Keywords:   meta.Keywords,
		Locus:      convertLocusGenbankToGen(meta.Locus),
		Name:       meta.Name,
		Organism:   meta.Organism,
		Origin:     meta.Origin,
		Other:      meta.Other,
		References: convertReferencesGenbankToGen(meta.References),
		Source:     meta.Source,
		Taxonomy:   meta.Taxonomy,
		Version:    meta.Version,
	}
}

func convertMetaGenToGenbank(meta gen.Meta) genbank.Meta {
	return genbank.Meta{
		Accession:  meta.Accession,
		BaseCount:  convertBaseCountsGenToGenbank(meta.BaseCount),
		Date:       meta.Date,
		Definition: meta.Definition,
		Keywords:   meta.Keywords,
		Locus:      convertLocusGenToGenbank(meta.Locus),
		References: convertReferencesGenToGenbank(meta.References),
		Source:     meta.Source,
		Taxonomy:   meta.Taxonomy,
		Origin:     meta.Origin,
		Other:      meta.Other,
		Name:       meta.Name,
		Version:    meta.Version,
	}
}

func ConvertToSlow5Header(genHeader gen.Slow5Header) slow5.Header {
	var slow5HeaderValues []slow5.HeaderValue
	for _, hv := range genHeader.HeaderValues {
		slow5HV := slow5.HeaderValue{
			ReadGroupID:        uint32(hv.ReadGroupID),
			Slow5Version:       hv.Slow5Version,
			Attributes:         hv.Attributes,
			EndReasonHeaderMap: hv.EndReasonHeaderMap,
		}
		slow5HeaderValues = append(slow5HeaderValues, slow5HV)
	}

	return slow5.Header{HeaderValues: slow5HeaderValues}
}

func ConvertToGenSlow5Header(slow5Header slow5.Header) gen.Slow5Header {
	var genHeaderValues []gen.HeaderValue
	for _, hv := range slow5Header.HeaderValues {
		genHV := gen.HeaderValue{
			ReadGroupID:        int(hv.ReadGroupID),
			Slow5Version:       hv.Slow5Version,
			Attributes:         hv.Attributes,
			EndReasonHeaderMap: hv.EndReasonHeaderMap,
		}
		genHeaderValues = append(genHeaderValues, genHV)
	}

	return gen.Slow5Header{HeaderValues: genHeaderValues}
}
