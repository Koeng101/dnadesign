package uniprot

import (
	"testing"
)

func TestIntListType_UnmarshalText(t *testing.T) {
	list := IntListType{}
	err := list.UnmarshalText([]byte("a"))
	if err.Error() != `strconv.Atoi: parsing "a": invalid syntax` {
		t.Errorf("Failed to get proper error")
	}
}
