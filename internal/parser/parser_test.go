package parser

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"
)

func TestParseMultipleRecords(t *testing.T) {
	input := "apple,orange,banana\r\n1,2,3\r\n"
	parser := NewCsvParser(strings.NewReader(input))
	got, err := parser.ParseAll()
	want := [][]string{{"apple", "orange", "banana"}, {"1", "2", "3"}}

	if !slices.Equal(got[0], want[0]) || !slices.Equal(got[1], want[1]) || err != nil {
		t.Errorf("parseRecord: want %v, got %v", want, got)
	}
}

func TestParseSingleRecord(t *testing.T) {
	input := "apple,orange,banana\r\n"
	parser := NewCsvParser(strings.NewReader(input))
	got, err := parser.parseRecord(input)
	want := []string{"apple", "orange", "banana"}

	if !slices.Equal(got, want) || err != nil {
		t.Errorf("parseRecord: want %v, got %v", want, got)
	}
}

func TestMissingLineBreak(t *testing.T) {
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

func TestTooManyFields(t *testing.T) {
	input := "apple,orange,banana\r\n1,2,3,4"
	parser := NewCsvParser(strings.NewReader(input))
	records, err := parser.ParseAll()

	if err == nil {
		t.Errorf("didn't return an error on too many fields")
	}

	var csvParseError *CsvParseError
	if errors.As(err, &csvParseError) {
		expectedErrorLine := 2
		if csvParseError.Line != expectedErrorLine {
			t.Errorf("%s", fmt.Sprintf("too many fields reported on wrong line. expected %v, got %v", expectedErrorLine, csvParseError.Line))
		}
	} else {
		t.Errorf("wrong error type returned on too many fields. expected a CsvParseError.")
	}

	expectedRecord := []string{"apple", "orange", "banana"}
	if len(records) != 1 || !slices.Equal(records[0], expectedRecord) {
		t.Errorf("%s", fmt.Sprintf("unexpected record parsed before encountering too many fields. expected [%v], got %v", expectedRecord, records))
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

/**
* 4. The last field in the record must not be followed by a comma.
 */
func TestLastFieldFollowedByComma(t *testing.T) {
	input := "apple,orange,banana,"
	parser := NewCsvParser(strings.NewReader(input))
	record, err := parser.Parse()

	if err == nil {
		t.Errorf("last field in a record shouldn't be allowed to be followed bya comma")
	}

	if record != nil {
		t.Errorf("shouldn't return malformed record")
	}

}
