package rebase

import (
	"errors"
	"io"
	"strings"
	"testing"
)

func TestRead(t *testing.T) {
	_, err := Read("data/FAKE.txt")
	if err == nil {
		t.Errorf("Failed to error on fake file")
	}
}

func TestParse_error(t *testing.T) {
	parseErr := errors.New("fake error")
	oldReadAllFn := readAllFn
	readAllFn = func(r io.Reader) ([]byte, error) {
		return nil, parseErr
	}
	defer func() {
		readAllFn = oldReadAllFn
	}()
	_, err := Parse(strings.NewReader(""))
	if err != parseErr {
		t.Errorf("err should equal parseErr")
	}
}

func TestRead_error(t *testing.T) {
	readErr := errors.New("fake error")
	oldParseFn := parseFn
	parseFn = func(file io.Reader) (map[string]Enzyme, error) {
		return nil, readErr
	}
	defer func() {
		parseFn = oldParseFn
	}()
	_, err := Read("data/rebase_test.txt")
	if err != readErr {
		t.Errorf("err should equal readErr")
	}
}

func TestRead_multipleRefs(t *testing.T) {
	enzymeMap, err := Read("data/rebase_test.txt")
	if err != nil {
		t.Error("Failed to read test file")
	}

	if enzymeMap["AcaI"].References != "Calleja, F., de Waard, A., Unpublished observations.\nHughes, S.G., Bruce, T., Murray, K., Unpublished observations." {
		t.Errorf("Failed to read references properly")
	}
}

func TestExport_error(t *testing.T) {
	exportErr := errors.New("fake error")
	oldMarshallFn := marshallFn
	marshallFn = func(v any) ([]byte, error) {
		return []byte{}, exportErr
	}
	defer func() {
		marshallFn = oldMarshallFn
	}()
	_, err := Export(map[string]Enzyme{})
	if err != exportErr {
		t.Errorf("err should equal exportErr")
	}
}

func TestRebase(t *testing.T) {
	_ = Rebase()
}
