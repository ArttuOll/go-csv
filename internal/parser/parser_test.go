package parser

import (
	"errors"
	"fmt"
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
	input := "apple,orange,banana\r\n1,2,34,5,6"
	parser := NewCsvParser(strings.NewReader(input))
	records, err := parser.ParseAll()

	if err == nil {
		t.Errorf("didn't return an error on missing line break")
	}

	var csvParseError *CsvParseError
	if errors.As(err, &csvParseError) {
		expectedErrorLine := 2
		if csvParseError.Line != expectedErrorLine {
			t.Errorf("%s", fmt.Sprintf("missing line break reported on wrong line. expected %v, got %v", expectedErrorLine, csvParseError.Line))
		}
	} else {
		t.Errorf("wrong error type returned on missing line break. expected a CsvParseError.")
	}

	expectedRecord := []string{"apple", "orange", "banana"}
	if len(records) != 1 || !slices.Equal(records[0], expectedRecord) {
		t.Errorf("%s", fmt.Sprintf("unexpected record parsed before missing line break. expected [%v], got %v", expectedRecord, records))
	}

}

/**
* 2. The last record in the file may or may not have an ending line break
 */
func TestParseRecordMissingLineBreakLastLine(t *testing.T) {
	input := "apple,orange,banana\r\n1,2,3"
	parser := NewCsvParser(strings.NewReader(input))
	records, err := parser.ParseAll()

	if err != nil {
		t.Errorf("last record shouldn't need to have a line break")
	}

	expectedRecord1 := []string{"apple", "orange", "banana"}
	expectedRecord2 := []string{"1", "2", "3"}
	if len(records) != 2 || !slices.Equal(records[0], expectedRecord1) || !slices.Equal(records[1], expectedRecord2) {
		t.Errorf("%s", fmt.Sprintf("unexpected records parsed. expected [%v] and [%v], got %v", expectedRecord1, expectedRecord2, records))
	}
}
