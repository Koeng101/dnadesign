package api

import (
	"github.com/koeng101/dnadesign/api/gen"
	"github.com/koeng101/dnadesign/lib/bio/genbank"
)

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
