package parser

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type CsvParser struct {
	reader          *bufio.Reader
	fieldsInARecord int
	currentLine     int
}

func NewCsvParser(r io.Reader) *CsvParser {
	return &CsvParser{
		reader: bufio.NewReader(r),
	}
}

func (parser *CsvParser) Parse() (records [][]string, err error) {
	line, err := parser.parseFirstLine()

	records = append(records, line)

	return records, nil
}

func (parser *CsvParser) parseFirstLine() (record []string, err error) {
	line, err := parser.reader.ReadSlice('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read the first line of the CSV file: %v", err)
	}

	// TODO: headers are parsed here

	fields := parser.parseRecord(string(line))
	parser.fieldsInARecord = len(fields)
	return fields, nil
}

func (parser *CsvParser) parseRecord(line string) []string {
	// TODO: parse fields
	return strings.Split(line, ",")
}

type CsvParseError struct {
	Line    int
	Message string
}

func (error *CsvParseError) Error() string {
	return fmt.Sprintf("[Line %v]: %v", error.Line, error.Message)
}
