package base58

import (
	"reflect"
	"testing"
)

func TestEncode(t *testing.T) {
	tests := []struct {
		input  []byte
		output string
	}{
		{[]byte(""), ""},
		{[]byte("hello world"), "StV1DL6CwTryKyV"},
		{[]byte{0x00, 0x00, 0x28}, "11h"},
	}

	for _, test := range tests {
		if encoded := Encode(test.input); encoded != test.output {
			t.Errorf("Encode(%x) = %s, want %s", test.input, encoded, test.output)
		}
	}
}

func TestDecode(t *testing.T) {
	tests := []struct {
		input  string
		output []byte
	}{
		{"", []byte("")},
		{"StV1DL6CwTryKyV", []byte("hello world")},
		{"11h", []byte{0x00, 0x00, 0x28}},
	}

	for _, test := range tests {
		if decoded := Decode(test.input); !reflect.DeepEqual(decoded, test.output) {
			t.Errorf("Decode(%s) = %x, want %x", test.input, decoded, test.output)
		}
	}

	invalidInputs := []string{"!@#", "I1O0"}
	for _, invalid := range invalidInputs {
		if decoded := Decode(invalid); decoded != nil {
			t.Errorf("Expected nil decoding result for invalid input %s but got %x", invalid, decoded)
		}
	}
}
