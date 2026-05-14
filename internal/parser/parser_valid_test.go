package parser

import (
	"slices"
	"strings"
	"testing"
)

func ParseAll_TestParseValidRecords(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  [][]string
	}{
		{
			name:  "MultipleRecords",
			input: "apple,orange,banana\r\n1,2,3\r\n",
			want:  [][]string{{"apple", "orange", "banana"}, {"1", "2", "3"}},
		},
		{
			name:  "MissingLineBreakLastLine",
			input: "apple,orange,banana\r\n1,2,3",
			want:  [][]string{{"apple", "orange", "banana"}, {"1", "2", "3"}},
		},
		{
			name:  "VariableNumberOfFields",
			input: "apple,orange,banana\r\n1,2,3,4",
			want:  [][]string{{"apple", "orange", "banana"}, {"1", "2", "3", "4"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewCSVParser(strings.NewReader(tt.input))
			got, err := parser.ParseAll()

			if err != nil {
				t.Fatalf("unexpected error when parsing input %v: %v", tt.input, err)
			}

			if !slices.EqualFunc(got, tt.want, slices.Equal[[]string]) {
				t.Fatalf("unexpected record parsed. wanted %v, but got %v", tt.want, got)
			}
		})
	}
}

func Parse_TestParseValidRecord(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "SingleRecord",
			input: "apple,orange,banana\r\n",
			want:  []string{"apple", "orange", "banana"},
		},
		{
			name:  "TrailingDelimiter",
			input: "apple,orange,banana,",
			want:  []string{"apple", "orange", "banana", ""},
		},
		{
			name:  "LineBreakWithinField",
			input: `"apple","oran\r\nge","banana"`,
			want:  []string{"apple", "oran\r\nge", "banana"},
		},
		{
			name:  "QuotedFieldWithEscapedDoubleQuote",
			input: `"apple","oran""ge","banana"`,
			want:  []string{"apple", `oran"ge`, "banana"},
		},
		{
			name:  "FieldContainingCommaEnclosedByDoubleQuotes",
			input: `"apple","ora,nge","banana"`,
			want:  []string{"apple", "ora,nge", "banana"},
		},
		{
			name:  "EmptyFile",
			input: "",
			want:  nil,
		},
		{
			name:  "QuotedEmptyFields",
			input: `"",""`,
			want:  []string{"", ""},
		},
		{
			name:  "EOFAfterQuote",
			input: `"orange"`,
			want:  []string{"orange"},
		},
		{
			name:  "EmptyField",
			input: "apple,,banana",
			want:  []string{"apple", "", "banana"},
		},
		{
			name:  "StartingDelimiter",
			input: ",orange,banana",
			want:  []string{"", "orange", "banana"},
		}, {
			name:  "DelimitersOnly",
			input: ",,",
			want:  []string{"", "", ""},
		},
		{
			name:  "RecordEndingAfterDelimiter",
			input: "apple,orange,\r\n",
			want:  []string{"apple", "orange", ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewCSVParser(strings.NewReader(tt.input))
			got, err := parser.Parse()

			if err != nil {
				t.Fatalf("unexpected error when parsing input %v: %v", tt.input, err)
			}

			if !slices.Equal(got, tt.want) {
				t.Fatalf("unexpected record parsed. wanted %v, but got %v", tt.want, got)
			}
		})
	}
}

func TestParse_RepeatedParsing(t *testing.T) {
	input := "apple,orange,banana\r\n1,2,3\r\n"
	parser := NewCSVParser(strings.NewReader(input))
	got1, _ := parser.Parse()
	got2, _ := parser.Parse()
	got3, err := parser.Parse()

	want1 := []string{"apple", "orange", "banana"}
	want2 := []string{"1", "2", "3"}

	if !slices.Equal(got1, want1) {
		t.Fatalf("unexpected record parsed when parsing repeatedly: want %v, got %v", want1, got1)
	}

	if !slices.Equal(got2, want2) {
		t.Fatalf("unexpected record parsed when parsing repeatedly: want %v, got %v", want2, got2)
	}

	if got3 != nil {
		t.Fatalf("unexpected record parsed when parsing repeatedly: want %v, got %v", nil, got3)
	}

	if err != nil {
		t.Fatalf("unexpected error when parsing repeatedly: %v", got3)
	}
}
