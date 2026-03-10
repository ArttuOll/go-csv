package parser

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type CsvParser struct {
	scanner         *bufio.Scanner
	fieldsInARecord int
	currentLine     int
}

func NewCsvParser(r io.Reader) *CsvParser {
	return &CsvParser{
		scanner: bufio.NewScanner(r),
	}
}

func (parser *CsvParser) ParseAll() (records [][]string, err error) {
	line, err := parser.parseFirstLine()

	records = append(records, line)

	for parser.scanner.Scan() {
		parser.currentLine++

		nextLine, err := parser.Parse()
		if err != nil {
			return nil, err
		}

		records = append(records, nextLine)
	}

	return records, nil
}

func (parser *CsvParser) Parse() (record []string, err error) {
	line := parser.scanner.Text()

	fields, err := parser.parseRecord(line)
	if err != nil {
		return nil, err
	}

	numberOfFields := len(fields)
	if numberOfFields != parser.fieldsInARecord {
		return nil, &CsvParseError{Line: parser.currentLine, Message: fmt.Sprintf("failed to parse line. too many fields in a record: %v, but should be %v", numberOfFields, parser.fieldsInARecord)}
	}

	return fields, nil
}

func (parser *CsvParser) parseFirstLine() (record []string, err error) {
	line := parser.scanner.Text()

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
