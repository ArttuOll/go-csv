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

func TestParseRecordMissingLineBreak(t *testing.T) {
	input := `apple,orange,banana
	1,2,3`
	parser := NewCsvParser(strings.NewReader(input))
	_, err := parser.parseRecord(input)

	if err == nil {
		t.Errorf("parseRecord: didn't return an error on missing line break")
	}
}

/**
* 2. The last record in the file may or may not have an ending line break
 */
func TestParseRecordMissingLineBreakLastLine(t *testing.T) {
	input := "apple,orange,banana\r\n1,2,3"
	parser := NewCsvParser(strings.NewReader(input))
	_, err := parser.parseRecord(input)

	if err != nil {
		t.Errorf("parseRecord: last record shouldn't need to have a line break")
	}
}
