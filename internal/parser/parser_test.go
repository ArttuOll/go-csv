package parser

import (
	"slices"
	"strings"
	"testing"
)

func TestParse_MultipleRecords(t *testing.T) {
	input := "apple,orange,banana\r\n1,2,3\r\n"
	parser := NewCSVParser(strings.NewReader(input))
	got, err := parser.ParseAll()
	want := [][]string{{"apple", "orange", "banana"}, {"1", "2", "3"}}

	if len(got) != len(want) || !slices.EqualFunc(got, want, slices.Equal[[]string]) || err != nil {
		t.Fatalf("parseRecord: want %v, got %v", want, got)
	}
}

func TestParse_SingleRecord(t *testing.T) {
	input := "apple,orange,banana\r\n"
	parser := NewCSVParser(strings.NewReader(input))
	got, err := parser.Parse()
	want := []string{"apple", "orange", "banana"}

	if len(got) != len(want) || !slices.Equal(got, want) || err != nil {
		t.Fatalf("parseRecord: want %v, got %v", want, got)
	}
}

/**
* 2. The last record in the file may or may not have an ending line break
 */
func TestParse_RecordMissingLineBreakLastLine(t *testing.T) {
	input := "apple,orange,banana\r\n1,2,3"
	parser := NewCSVParser(strings.NewReader(input))
	records, err := parser.ParseAll()

	if err != nil {
		t.Fatalf("last record shouldn't need to have a line break")
	}

	expectedRecords := [][]string{{"apple", "orange", "banana"}, {"1", "2", "3"}}
	if !slices.EqualFunc(records, expectedRecords, slices.Equal[[]string]) {
		t.Fatalf("unexpected records parsed. expected [%v], got %v", expectedRecords, records)
	}
}

func TestParse_VariableNumberOfFields(t *testing.T) {
	input := "apple,orange,banana\r\n1,2,3,4"
	parser := NewCSVParser(strings.NewReader(input))
	records, err := parser.ParseAll()

	if err != nil {
		t.Fatalf("variable number of fields in records should be allowed, but got error: %v", err)
	}

	expectedRecords := [][]string{{"apple", "orange", "banana"}, {"1", "2", "3", "4"}}
	if !slices.EqualFunc(records, expectedRecords, slices.Equal[[]string]) {
		t.Fatalf("unexpected records parsed when records contains variable number of fields. expected [%v], got %v", expectedRecords, records)
	}

}

func TestParse_TrailingDelimiter(t *testing.T) {
	input := "apple,orange,banana,"
	parser := NewCSVParser(strings.NewReader(input))
	record, err := parser.Parse()

	if err != nil {
		t.Fatalf("last field in a record should be allowed to be followed by a comma. this should result in an empty field.")
	}

	expectedRecord := []string{"apple", "orange", "banana", ""}
	if !slices.Equal(record, expectedRecord) {
		t.Fatalf("unexpected record parsed when last field is followed by a comma. expected [%v], got %v", expectedRecord, record)
	}
}

func TestParse_UnescapedDoubleQuoteInQuotedField(t *testing.T) {
	input := `apple,"orange"",banana`
	parser := NewCSVParser(strings.NewReader(input))
	_, err := parser.Parse()

	if err == nil {
		t.Fatalf("unescaped double quotes within quoted fields shouldn't be allowed")
	}
}

func TestParse_QuoteInUnquotedField(t *testing.T) {
	input := `apple,ora"nge,banana`
	parser := NewCSVParser(strings.NewReader(input))
	_, err := parser.Parse()

	if err == nil {
		t.Fatalf("quotes in unquoted fields shouldn't be allowed")
	}
}

func TestParse_LineBreakWithinField(t *testing.T) {
	input := `"apple","oran\r\nge","banana"`
	parser := NewCSVParser(strings.NewReader(input))
	_, err := parser.Parse()

	if err != nil {
		t.Fatalf("line breaks within double quote enclosed fields should be allowed: %v", err)
	}
}

func TestParse_QuotedFieldWithEscapedDoubleQuote(t *testing.T) {
	input := `"apple","oran""ge","banana"`
	parser := NewCSVParser(strings.NewReader(input))
	got, err := parser.Parse()
	want := []string{"apple", `oran"ge`, "banana"}

	if err != nil {
		t.Fatalf("escaped double quotes within quoted fields should be allowed: %v", err)
	}

	if !slices.Equal(want, got) {
		t.Fatalf("returned unexpected record when parsing a field with an escaped double quote. expected: %v, got: %v", want, got)
	}
}

func TestParse_FieldContainingCommaEnclosedByDoubleQuotes(t *testing.T) {
	input := `"apple","ora,nge","banana"`
	parser := NewCSVParser(strings.NewReader(input))
	got, err := parser.Parse()
	want := []string{"apple", "ora,nge", "banana"}

	if err != nil {
		t.Fatalf("commas within double quote enclosed fields should be allowed: %v", err)
	}

	if !slices.Equal(want, got) {
		t.Fatalf("returned unexpected record when parsing a field with a comma. expected: %v, got: %v", want, got)
	}
}

func TestParse_EmptyFile(t *testing.T) {
	input := ``
	parser := NewCSVParser(strings.NewReader(input))
	got, err := parser.Parse()

	if err != nil {
		t.Fatalf("empty file should be allowed: %v", err)
	}

	if got != nil {
		t.Fatalf("returned unexpected record when parsing an empty file. expected: %v, got: %v", nil, got)
	}
}

func TestParse_QuotedEmptyFields(t *testing.T) {
	input := `"",""`
	parser := NewCSVParser(strings.NewReader(input))
	got, err := parser.Parse()
	want := []string{"", ""}

	if err != nil {
		t.Fatalf("quoted empty fields should be allowed: %v", err)
	}

	if !slices.Equal(want, got) {
		t.Fatalf("returned unexpected record when parsing quoted empty fields. expected: %v, got: %v", want, got)
	}
}

func TestParse_RecordTerminatedWithJustCarriageReturn(t *testing.T) {
	input := "apple,orange\r"
	parser := NewCSVParser(strings.NewReader(input))
	_, err := parser.Parse()

	if err == nil {
		t.Fatalf("record isn't allowed to be terminated with just a carriage return")
	}
}

func TestParse_RecordTerminatedWithJustLineFeed(t *testing.T) {
	input := "apple,orange\n"
	parser := NewCSVParser(strings.NewReader(input))
	_, err := parser.Parse()

	if err == nil {
		t.Fatalf("record isn't allowed to be terminated with just a line feed")
	}
}

func TestParse_EmptyField(t *testing.T) {
	input := "apple,,banana"
	parser := NewCSVParser(strings.NewReader(input))
	got, err := parser.Parse()
	want := []string{"apple", "", "banana"}

	if err != nil {
		t.Fatalf("empty fields should be allowed, but got error: %v", err)
	}

	if !slices.Equal(want, got) {
		t.Fatalf("returned unexpected record when parsing a record with an empty field. expected: %v, got: %v", want, got)
	}
}

func TestParse_StartingDelimiter(t *testing.T) {
	input := ",orange,banana"
	parser := NewCSVParser(strings.NewReader(input))
	got, err := parser.Parse()
	want := []string{"", "orange", "banana"}

	if err != nil {
		t.Fatalf("starting delimiter should be allowed, but got error %v", err)
	}

	if !slices.Equal(want, got) {
		t.Fatalf("returned unexpected record when parsing a record with a starting delimiter. expected: %v, got: %v", want, got)
	}
}

func TestParse_DelimitersOnly(t *testing.T) {
	input := ",,"
	parser := NewCSVParser(strings.NewReader(input))
	got, err := parser.Parse()
	want := []string{"", "", ""}

	if err != nil {
		t.Fatalf("delimiters-only files should be allowed, but got error: %v", err)
	}

	if !slices.Equal(want, got) {
		t.Fatalf("returned unexpected record when parsing a record with only delimiters. expected: %v, got: %v", want, got)
	}
}

func TestParse_RecordEndingAfterDelimiter(t *testing.T) {
	input := "apple,orange,\r\n"
	parser := NewCSVParser(strings.NewReader(input))
	got, err := parser.Parse()
	want := []string{"apple", "orange", ""}

	if err != nil {
		t.Fatalf("records ending after delimiters should be allowed, but got error: %v", err)
	}

	if !slices.Equal(want, got) {
		t.Fatalf("returned unexpected record when parsing a record ending after a delimiter. expected: %v, got: %v", want, got)
	}
}
