package linearfold

import (
	"github.com/koeng101/dnadesign/lib/fold/mfe"
	"github.com/koeng101/dnadesign/lib/fold/mfe/energy_params"
)

/*
These interfaces are here to satisfy the `SequenceFolder` interface, which
looks like this:

type SequenceFolder interface {
    Fold(sequence string, temp float64) (dotBracket string, score float64, error)
}
*/

type CONTRAfoldV2FoldWrapper struct {
	BeamSize int
}

type ViennaRNAFoldWrapper struct {
	EnergyParamsSet   energy_params.EnergyParamsSet
	DanglingEndsModel mfe.DanglingEndsModel
	BeamSize          int
}

func NewDefaultCONTRAfoldV2FoldWrapper() CONTRAfoldV2FoldWrapper {
	return CONTRAfoldV2FoldWrapper{BeamSize: DefaultBeamSize}
}

func NewDefaultViennaRnaFoldWrapper() ViennaRNAFoldWrapper {
	return ViennaRNAFoldWrapper{EnergyParamsSet: DefaultEnergyParamsSet, DanglingEndsModel: mfe.DefaultDanglingEndsModel, BeamSize: DefaultBeamSize}
}

func (c CONTRAfoldV2FoldWrapper) Fold(sequence string, temp float64) (dotBracket string, score float64, err error) {
	// Assuming CONTRAfoldV2 is adjusted to accept a temperature parameter and return an error
	dotBracket, score = CONTRAfoldV2(sequence, c.BeamSize)
	return dotBracket, score, nil
}

func (v ViennaRNAFoldWrapper) Fold(sequence string, temp float64) (dotBracket string, score float64, err error) {
	dotBracket, score = ViennaRNAFold(sequence, temp, v.EnergyParamsSet, v.DanglingEndsModel, v.BeamSize)
	return dotBracket, score, nil
}
