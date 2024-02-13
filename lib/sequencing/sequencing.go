/*
Package sequencing contains functions associated with handling sequencing data.
*/
package sequencing

import (
	"context"

	"github.com/koeng101/dnadesign/lib/align/megamash"
	"github.com/koeng101/dnadesign/lib/bio/fastq"
	"github.com/koeng101/dnadesign/lib/sequencing/barcoding"
)

func MegamashFastq(ctx context.Context, megamashMap megamash.MegamashMap, input <-chan fastq.Read, output chan<- fastq.Read) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case data, ok := <-input:
			if !ok {
				return nil
			}
			matches := megamashMap.Match(data.Sequence)
			jsonStr, _ := megamash.MatchesToJSON(matches)
			readCopy := data.DeepCopy()
			readCopy.Optionals["megamash"] = jsonStr
			select {
			case output <- readCopy:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}

func DualBarcodeFastq(ctx context.Context, primerSet barcoding.DualBarcodePrimerSet, input <-chan fastq.Read, output chan<- fastq.Read) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case data, ok := <-input:
			if !ok {
				return nil
			}
			well, err := barcoding.DualBarcodeSequence(data.Sequence, primerSet)
			if err != nil {
				return err
			}
			readCopy := data.DeepCopy()
			readCopy.Optionals["dual_barcode"] = well
			select {
			case output <- readCopy:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}
