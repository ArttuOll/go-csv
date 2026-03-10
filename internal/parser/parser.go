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

	for parser.scanner.Scan() {
		nextLine, err := parser.parseLine()
		if err != nil {
			return records, err
		}

		records = append(records, nextLine)
	}

	return records, nil
}

func (parser *CsvParser) Parse() (record []string, err error) {
	parser.scanner.Scan()

	return parser.parseLine()
}

func (parser *CsvParser) parseLine() (record []string, err error) {
	line := parser.scanner.Text()

	fields, err := parser.parseRecord(line)
	if err != nil {
		return nil, err
	}

	parser.currentLine++

	if parser.fieldsInARecord == 0 {
		parser.fieldsInARecord = len(fields)
	}

	numberOfFields := len(fields)
	if numberOfFields != parser.fieldsInARecord {
		return nil, &CsvParseError{Line: parser.currentLine, Message: fmt.Sprintf("failed to parse line. too many fields in a record: %v, but should be %v", numberOfFields, parser.fieldsInARecord)}
	}

	return fields, nil
}

func (parser *CsvParser) parseRecord(line string) ([]string, error) {
	// TODO: parse fields
	withoutLineBreak, _ := strings.CutSuffix(line, "\r\n")
	return strings.Split(withoutLineBreak, ","), nil
}

type CsvParseError struct {
	Line    int
	Message string
}

func (error *CsvParseError) Error() string {
	return fmt.Sprintf("[Line %v]: %v", error.Line, error.Message)
}
