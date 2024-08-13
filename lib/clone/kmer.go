package clone

import (
	"errors"
	"strings"

	"github.com/koeng101/dnadesign/lib/align/megamash"
	"github.com/koeng101/dnadesign/lib/bio/fastq"
	"github.com/koeng101/dnadesign/lib/transform"
)

/*
This includes code to create and check kmers from ligation reactions.

Here is the problem: You have sequenced a GoldenGate or ligation you've run
to assemble some DNA. How do you quantify the efficiency of the ligation from
this raw data? You may want to do this to make sure your ligations are working
properly in a way that simple controls wouldn't: you can *directly* observe
single molecules of interest that are ligated and ones which are not.

How would you do this? The simplest version, which we implement here, is to
check whether or not kmers indicative of ligation exist within the sequenced
fragments. This is simple and computationally inexpensive.
*/

// KmerOverlap represents the overlap between two fragments
type KmerOverlap struct {
	Kmer      string
	Fragment1 Fragment
	Fragment2 Fragment
}

func FindKmerOverlaps(fragments []Fragment, ligationProduct string, ligationPattern []int, kmerSize int) ([]KmerOverlap, error) {
	if len(fragments) < 2 {
		return []KmerOverlap{}, errors.New("need at least two fragments to find overlaps")
	}
	bpFromEach := (kmerSize - 4) / 2 // bp needed from each side
	if bpFromEach < 4 {
		return []KmerOverlap{}, errors.New("need at least a kmer of 12")
	}

	ligation := ligationProduct + ligationProduct // double for circularization
	position := 0                                 // index position
	var kmer string
	var kmerOverlaps []KmerOverlap

	for i := 0; i < len(ligationPattern); i++ {
		var frag1, frag2 Fragment
		frag1 = fragments[ligationPattern[i]]
		if len(fragments)-1 == i { // Account for last fragment
			frag2 = fragments[0]
		} else {
			frag2 = fragments[ligationPattern[i+1]]
		}

		position = position + len(frag1.ForwardOverhang) + len(frag1.Sequence)
		kmer = megamash.StandardizeDNA(ligation[position-bpFromEach : position+4+bpFromEach])
		kmerOverlaps = append(kmerOverlaps, KmerOverlap{Kmer: kmer, Fragment1: frag1, Fragment2: frag2})
	}
	return kmerOverlaps, nil
}

func FindKmers(kmerOverlaps []KmerOverlap, read fastq.Read) []KmerOverlap {
	var outputKmerOverlaps []KmerOverlap
	sequence := strings.ToUpper(read.Sequence)
	for _, kmerOverlap := range kmerOverlaps {
		if strings.Contains(sequence, strings.ToUpper(kmerOverlap.Kmer)) || strings.Contains(sequence, strings.ToUpper(transform.ReverseComplement(kmerOverlap.Kmer))) {
			outputKmerOverlaps = append(outputKmerOverlaps, KmerOverlap{Kmer: kmerOverlap.Kmer, Fragment1: kmerOverlap.Fragment1, Fragment2: kmerOverlap.Fragment2})
		}
	}
	return outputKmerOverlaps
}
