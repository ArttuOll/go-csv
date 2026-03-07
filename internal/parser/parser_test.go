package parser

import (
	"slices"
	"strings"
	"testing"
)

func TestParseRecord(t *testing.T) {
	input := "apple,orange,banana\r\n"
	parser := NewCsvParser(strings.NewReader(input))
	got, err := parser.parseRecord(input)
	want := []string{"apple", "orange", "banana"}

	if !slices.Equal(got, want) || err != nil {
		t.Errorf("parseRecord: want %v, got %v", want, got)
	}
}

// TODO: Last record may not have a line break
func TestParseRecordMissingLineBreak(t *testing.T) {
	input := "apple,orange,banana"
	parser := NewCsvParser(strings.NewReader(input))
	_, err := parser.parseRecord(input)

	if err == nil {
		t.Errorf("parseRecord: didn't return an error on missing line break")
	}
}
