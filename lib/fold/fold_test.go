package fold

import (
	"testing"

	"github.com/koeng101/dnadesign/lib/fold/linearfold"
	"github.com/koeng101/dnadesign/lib/fold/zuker"
)

func TestCONTRAfold(t *testing.T) {
	var newFolder SequenceFolder
	newFolder = linearfold.NewDefaultCONTRAfoldV2FoldWrapper()
	result, score, _ := newFolder.Fold("UGAGUUCUCGAUCUCUAAAAUCG", 37.0)

	expectedResult := "......................."
	if result != expectedResult {
		t.Errorf("Got unexpected result. Expected: %s\nGot: %s", expectedResult, result)
	}
	expectedScore := -0.22376699999999988
	if !(score >= -0.22600466999999988 && score <= -0.22152932999999989) {
		t.Errorf("Got unexpected score. Expected: %f +/- 1%%\nGot: %f", expectedScore, score)
	}
}

func TestViennaRNAFold(t *testing.T) {
	var newFolder SequenceFolder
	newFolder = linearfold.NewDefaultViennaRnaFoldWrapper()
	result, score, _ := newFolder.Fold("UCUAGACUUUUCGAUAUCGCGAAAAAAAAU", 37.0)

	expectedResult := ".......((((((......))))))....."
	if result != expectedResult {
		t.Errorf("Got unexpected result. Expected: %s\nGot: %s", expectedResult, result)
	}
	expectedScore := -3.90
	if !(score >= -3.939 && score <= -3.861) {
		t.Errorf("Got unexpected score. Expected: %f +/- 1%%\nGot: %f", expectedScore, score)
	}
}

func TestZuker(t *testing.T) {
	var newFolder SequenceFolder
	newFolder = zuker.NewZukerFoldWrapper()
	result, score, _ := newFolder.Fold("ACCCCCUCCUUCCUUGGAUCAAGGGGCUCAA", 37.0)

	expectedResult := ".((((.(((......)))....))))"
	if result != expectedResult {
		t.Errorf("Got unexpected result. Expected: %s\nGot: %s", expectedResult, result)
	}

	expectedScore := -9.422465
	if !(score >= -9.51668965 && score <= -9.32824035) {
		t.Errorf("Got unexpected score. Expected: %f +/- 1%%\nGot: %f", expectedScore, score)
	}
}
