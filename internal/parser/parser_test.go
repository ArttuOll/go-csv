package parser

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"
)

func TestMultipleRecords(t *testing.T) {
	input := "apple,orange,banana\r\n1,2,3\r\n"
	parser := NewCsvParser(strings.NewReader(input))
	got, err := parser.ParseAll()
	want := [][]string{{"apple", "orange", "banana"}, {"1", "2", "3"}}

	if !slices.Equal(got[0], want[0]) || !slices.Equal(got[1], want[1]) || err != nil {
		t.Errorf("parseRecord: want %v, got %v", want, got)
	}
}

func TestSingleRecord(t *testing.T) {
	input := "apple,orange,banana\r\n"
	parser := NewCsvParser(strings.NewReader(input))
	got, err := parser.Parse()
	want := []string{"apple", "orange", "banana"}

	if !slices.Equal(got, want) || err != nil {
		t.Errorf("parseRecord: want %v, got %v", want, got)
	}
}

/**
* 2. The last record in the file may or may not have an ending line break
 */
func TestRecordMissingLineBreakLastLine(t *testing.T) {
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

func TestTrailingDelimiter(t *testing.T) {
	input := "apple,orange,banana,"
	parser := NewCsvParser(strings.NewReader(input))
	record, err := parser.Parse()

	if err != nil {
		t.Errorf("last field in a record should be allowed to be followed by a comma. this should result in an empty field.")
	}

	expectedRecord := []string{"apple", "orange", "banana", ""}
	if !slices.Equal(record, expectedRecord) {
		t.Errorf("%s", fmt.Sprintf("unexpected record parsed when last field is followed by a comma. expected [%v], got %v", expectedRecord, record))
	}
}

func TestUnescapedDoubleQuoteInQuotedField(t *testing.T) {
	input := `apple,"orange"",banana`
	parser := NewCsvParser(strings.NewReader(input))
	_, err := parser.Parse()

	if err == nil {
		t.Errorf("unescaped double quotes within quoted fields shouldn't be allowed")
	}
}

func TestQuoteInUnquotedField(t *testing.T) {
	input := `apple,ora"nge,banana`
	parser := NewCsvParser(strings.NewReader(input))
	_, err := parser.Parse()

	if err == nil {
		t.Errorf("quotes in unquoted fields shouldn't be allowed")
	}
}

func TestLineBreakWithinField(t *testing.T) {
	input := `"apple","oran\r\nge","banana"`
	parser := NewCsvParser(strings.NewReader(input))
	_, err := parser.Parse()

	if err != nil {
		t.Errorf("line breaks within double quote enclosed fields should be allowed: %v", err)
	}
}

func TestQuotedFieldWithEscapedDoubleQuote(t *testing.T) {
	input := `"apple","oran""ge","banana"`
	parser := NewCsvParser(strings.NewReader(input))
	got, err := parser.Parse()
	want := []string{"apple", `oran"ge`, "banana"}

	if err != nil {
		t.Errorf("escaped double quotes within quoted fields should be allowed: %v", err)
	}

	if !slices.Equal(want, got) {
		t.Errorf("returned unexpected record when parsing a field with an escaped double quote. expected: %v, got: %v", want, got)
	}
}

func TestFieldContainingCommaEnclosedByDoubleQuotes(t *testing.T) {
	input := `"apple","ora,nge","banana"`
	parser := NewCsvParser(strings.NewReader(input))
	got, err := parser.Parse()
	want := []string{"apple", "ora,nge", "banana"}

	if err != nil {
		t.Errorf("commas within double quote enclosed fields should be allowed: %v", err)
	}

	if !slices.Equal(want, got) {
		t.Errorf("returned unexpected record when parsing a field with a comma. expected: %v, got: %v", want, got)
	}
}

func TestEmptyFile(t *testing.T) {
	input := ``
	parser := NewCsvParser(strings.NewReader(input))
	got, err := parser.Parse()
	want := []string{}

	if err != nil {
		t.Errorf("empty file should be allowed: %v", err)
	}

	if !slices.Equal(want, got) {
		t.Errorf("returned unexpected record when parsing an empty file. expected: %v, got: %v", want, got)
	}
}

func TestQuotedEmptyFields(t *testing.T) {
	input := `"",""`
	parser := NewCsvParser(strings.NewReader(input))
	got, err := parser.Parse()
	want := []string{"", ""}

	if err != nil {
		t.Errorf("quoted empty fields should be allowed: %v", err)
	}

	if !slices.Equal(want, got) {
		t.Errorf("returned unexpected record when parsing quoted empty fields. expected: %v, got: %v", want, got)
	}
}

func TestRecordTerminatedWithJustCarriageReturn(t *testing.T) {
	input := "apple,orange\r"
	parser := NewCsvParser(strings.NewReader(input))
	_, err := parser.Parse()

	if err == nil {
		t.Errorf("record isn't allowed to be terminated with just a carriage return")
	}
}

func TestRecordTerminatedWithJustLineFeed(t *testing.T) {
	input := "apple,orange\n"
	parser := NewCsvParser(strings.NewReader(input))
	_, err := parser.Parse()

	if err == nil {
		t.Errorf("record isn't allowed to be terminated with just a line feed")
	}
}

func TestEmptyField(t *testing.T) {
	input := "apple,,banana"
	parser := NewCsvParser(strings.NewReader(input))
	got, err := parser.Parse()
	want := []string{"apple", "", "banana"}

	if err != nil {
		t.Errorf("empty fields should be allowed, but got error: %v", err)
	}

	if !slices.Equal(want, got) {
		t.Errorf("returned unexpected record when parsing a record with an empty field. expected: %v, got: %v", want, got)
	}
}

func TestStartingDelimiter(t *testing.T) {
	input := ",orange,banana"
	parser := NewCsvParser(strings.NewReader(input))
	got, err := parser.Parse()
	want := []string{"", "orange", "banana"}

	if err != nil {
		t.Errorf("starting delimiter should be allowed, but got error %v", err)
	}

	if !slices.Equal(want, got) {
		t.Errorf("returned unexpected record when parsing a record with a starting delimiter. expected: %v, got: %v", want, got)
	}
}

func TestDelimitersOnly(t *testing.T) {
	input := ",,"
	parser := NewCsvParser(strings.NewReader(input))
	got, err := parser.Parse()
	want := []string{"", "", ""}

	if err != nil {
		t.Errorf("delimiters-only files should be allowed, but got error: %v", err)
	}

	if !slices.Equal(want, got) {
		t.Errorf("returned unexpected record when parsing a record with only delimiters. expected: %v, got: %v", want, got)
	}
}

func TestRecordEndingAfterDelimiter(t *testing.T) {
	input := "apple,orange,\r\n"
	parser := NewCsvParser(strings.NewReader(input))
	got, err := parser.Parse()
	want := []string{"apple", "orange", ""}

	if err != nil {
		t.Errorf("records ending after delimiters should be allowed, but got error: %v", err)
	}

	if !slices.Equal(want, got) {
		t.Errorf("returned unexpected record when parsing a record ending after a delimiter. expected: %v, got: %v", want, got)
	}
}
