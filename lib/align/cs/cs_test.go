package cs

import (
	"reflect"
	"testing"
)

func TestParseAndDigestCS(t *testing.T) {
	// Test case from the reference documentation
	csString := ":6-ata:10+gtc:4*at:3"

	// Test ParseCS
	expectedCS := []CS{
		{Type: ':', Size: 6, Change: ""},
		{Type: '-', Size: 3, Change: "ata"},
		{Type: ':', Size: 10, Change: ""},
		{Type: '+', Size: 3, Change: "gtc"},
		{Type: ':', Size: 4, Change: ""},
		{Type: '*', Size: 1, Change: "at"},
		{Type: ':', Size: 3, Change: ""},
	}

	parsedCS := ParseCS(csString)

	if !reflect.DeepEqual(parsedCS, expectedCS) {
		t.Errorf("ParseCS(%s) = %v, want %v", csString, parsedCS, expectedCS)
	}

	// Test DigestCS
	expectedDigestedCS := []DigestedCS{
		{Position: 6, Type: '*'},
		{Position: 7, Type: '*'},
		{Position: 8, Type: '*'},
		{Position: 19, Type: '+'},
		{Position: 23, Type: 't'},
	}

	expectedDigestedInsertions := []DigestedInsertion{
		{Position: 19, Insertion: "gtc"},
	}

	digestedCS, digestedInsertions := DigestCS(parsedCS)

	if !reflect.DeepEqual(digestedCS, expectedDigestedCS) {
		t.Errorf("DigestCS() digestedCS = %v, want %v", digestedCS, expectedDigestedCS)
	}

	if !reflect.DeepEqual(digestedInsertions, expectedDigestedInsertions) {
		t.Errorf("DigestCS() digestedInsertions = %v, want %v", digestedInsertions, expectedDigestedInsertions)
	}
}

func TestParseAndDigestLongerCS(t *testing.T) {
	// Test case with a longer CS string
	csString := ":38-t:25-c:147*ag:29+gta:9*gc:3-gag:6-t:94-cgca:77-c*ag:16-at:4*ct:1+g:1*gc:98-c:191-c:200*ga:21*ga:2*ca:1-c:83*ga"

	// Test ParseCS
	expectedCS := []CS{
		{Type: ':', Size: 38, Change: ""},
		{Type: '-', Size: 1, Change: "t"},
		{Type: ':', Size: 25, Change: ""},
		{Type: '-', Size: 1, Change: "c"},
		{Type: ':', Size: 147, Change: ""},
		{Type: '*', Size: 1, Change: "ag"},
		{Type: ':', Size: 29, Change: ""},
		{Type: '+', Size: 3, Change: "gta"},
		{Type: ':', Size: 9, Change: ""},
		{Type: '*', Size: 1, Change: "gc"},
		{Type: ':', Size: 3, Change: ""},
		{Type: '-', Size: 3, Change: "gag"},
		{Type: ':', Size: 6, Change: ""},
		{Type: '-', Size: 1, Change: "t"},
		{Type: ':', Size: 94, Change: ""},
		{Type: '-', Size: 4, Change: "cgca"},
		{Type: ':', Size: 77, Change: ""},
		{Type: '-', Size: 1, Change: "c"},
		{Type: '*', Size: 1, Change: "ag"},
		{Type: ':', Size: 16, Change: ""},
		{Type: '-', Size: 2, Change: "at"},
		{Type: ':', Size: 4, Change: ""},
		{Type: '*', Size: 1, Change: "ct"},
		{Type: ':', Size: 1, Change: ""},
		{Type: '+', Size: 1, Change: "g"},
		{Type: ':', Size: 1, Change: ""},
		{Type: '*', Size: 1, Change: "gc"},
		{Type: ':', Size: 98, Change: ""},
		{Type: '-', Size: 1, Change: "c"},
		{Type: ':', Size: 191, Change: ""},
		{Type: '-', Size: 1, Change: "c"},
		{Type: ':', Size: 200, Change: ""},
		{Type: '*', Size: 1, Change: "ga"},
		{Type: ':', Size: 21, Change: ""},
		{Type: '*', Size: 1, Change: "ga"},
		{Type: ':', Size: 2, Change: ""},
		{Type: '*', Size: 1, Change: "ca"},
		{Type: ':', Size: 1, Change: ""},
		{Type: '-', Size: 1, Change: "c"},
		{Type: ':', Size: 83, Change: ""},
		{Type: '*', Size: 1, Change: "ga"},
	}

	parsedCS := ParseCS(csString)

	if !reflect.DeepEqual(parsedCS, expectedCS) {
		t.Errorf("ParseCS(%s) = %v, want %v", csString, parsedCS, expectedCS)
	}
}
