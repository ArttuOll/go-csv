package parser

import (
	"strings"
	"testing"
)

func Parse_TestParseInvalidRecord(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "UnescapedDoubleQuoteInQuotedField",
			input: `apple,"orange"",banana`,
		},
		{
			name:  "QuoteInUnquotedField",
			input: `apple,ora"nge,banana`,
		},
		{
			name:  "RecordTerminatedWithJustCarriageReturn",
			input: "apple,orange\r",
		},
		{
			name:  "MalformedLineBreak",
			input: "apple,orange\rX",
		},
		{

			name:  "RecordTerminatedWithJustLineFeed",
			input: "apple,orange\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewCSVParser(strings.NewReader(tt.input))
			_, err := parser.Parse()

			if err == nil {
				t.Fatalf("unexpectedly didn't error on input %v", tt.input)
			}
		})
	}
}
