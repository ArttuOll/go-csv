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
	line, err := parser.reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read the first line of the CSV file: %v", err)
	}

	// TODO: headers are parsed here

	fields, err := parser.parseRecord(line)
	if err != nil {
		return nil, err
	}

	parser.fieldsInARecord = len(fields)
	return fields, nil
}

func (parser *CsvParser) parseRecord(line string) ([]string, error) {
	// TODO: parse fields
	withoutLineBreak, found := strings.CutSuffix(line, "\r\n")
	if !found {
		// TODO: Last record may not have a line break
		return nil, fmt.Errorf("failed to parse record %v. missing line break.", line)
	}

	return strings.Split(withoutLineBreak, ","), nil
}

type CsvParseError struct {
	Line    int
	Message string
}

func (error *CsvParseError) Error() string {
	return fmt.Sprintf("[Line %v]: %v", error.Line, error.Message)
}
