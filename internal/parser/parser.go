package parser

import (
	"bufio"
	"errors"
	"fmt"
	"io"
)

type ParserState int

const (
	ParsingField = iota
	ParsingRecord
)

type CsvParser struct {
	reader          io.Reader
	fieldsInARecord int
	currentLine     int
	line            string
	state           ParserState
	buffer          []byte
	readToIndex     int
	position        int
	fieldBuffer     []string
	done            bool
}

const BUFFER_SIZE = 8

func NewCsvParser(r io.Reader) *CsvParser {
	return &CsvParser{
		reader:      bufio.NewReader(r),
		buffer:      make([]byte, BUFFER_SIZE),
		currentLine: 1,
	}
}

type CsvParseError struct {
	Line    int
	Message string
}

func (error *CsvParseError) Error() string {
	return fmt.Sprintf("[Line %v]: %v", error.Line, error.Message)
}

func (parser *CsvParser) ParseAll() (records [][]string, err error) {
	for !parser.done {
		nextLine, err := parser.parseLine()
		if err != nil {
			return records, err
		}

		records = append(records, nextLine)
	}

	return records, nil
}

func (parser *CsvParser) Parse() (record []string, err error) {
	return parser.parseLine()
}

func (parser *CsvParser) parseLine() (record []string, err error) {
	for {
		if parser.readToIndex >= len(parser.buffer) {
			newBuffer := make([]byte, len(parser.buffer)*2)
			copy(newBuffer, parser.buffer)
			parser.buffer = newBuffer
		}

		numberOfBytesInChunk, err := parser.reader.Read(parser.buffer[parser.readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				parser.done = true
			} else {
				return nil, err
			}
		}

		parser.readToIndex += numberOfBytesInChunk

		// Attempt to parse data received so far
		record, charsParsed, err := parser.parseRecord(string(parser.buffer[:parser.readToIndex]))
		if err != nil {
			return nil, err
		}

		// Need more data
		if charsParsed == 0 {
			continue
		}

		parser.currentLine++

		// Succeeded in parsing record. Remove the parsed record from the buffer.
		copy(parser.buffer, parser.buffer[charsParsed:])
		parser.readToIndex -= charsParsed

		parser.position = 0
		parser.fieldBuffer = []string{}

		return record, nil
	}
}

func (parser *CsvParser) parseRecord(input string) ([]string, int, error) {
	for {
		field, charsRead, err := parser.parseField(input[parser.position:])
		if err != nil {
			return nil, 0, err
		}

		// Need more data
		if charsRead == 0 {
			return nil, 0, nil
		}

		parser.fieldBuffer = append(parser.fieldBuffer, field)
		parser.position += len(field)

		if parser.peek(input) == ',' {
			if parser.position+1 == len(input) {
				return nil, 0, &CsvParseError{Line: parser.currentLine, Message: "failed to parse record. trailing commas are not allowed"}
			}

			parser.position++
			continue
		}

		if parser.peek(input) == '\r' {
			parser.position += 2
		}

		numberOfFields := len(parser.fieldBuffer)

		if parser.fieldsInARecord == 0 {
			parser.fieldsInARecord = len(parser.fieldBuffer)
		} else if numberOfFields != parser.fieldsInARecord {
			return nil, 0, &CsvParseError{Line: parser.currentLine, Message: fmt.Sprintf("failed to parse record. too many fields in a record: %v, but should be %v", numberOfFields, parser.fieldsInARecord)}
		}

		return parser.fieldBuffer, parser.position, nil
	}
}

func (parser *CsvParser) parseField(input string) (string, int, error) {
	escaped := false
	escapingDoubleQuote := false

	for i, v := range input {
		if i == 0 && v == '"' {
			escaped = true
			continue
		}

		// Final field ends
		if !escaped && (v == '\r' || v == '\n') {
			return input[:i], i, nil
		}

		if !escaped && v == '"' {
			return "", 0, &CsvParseError{Line: parser.currentLine, Message: "fields containing double quotes must be enclosed by double quotes. the contained double quote must be escaped with a preceding double quote."}
		}

		// Field ends
		if !escaped && v == ',' {
			return input[:i], i, nil
		}

		if v == '"' {
			// Closing double quote, field ends
			if i+1 == len(input) {
				return input[1 : len(input)-1], i + 1, nil
			}

			// Escaped double quote sequence
			if input[i+1] == '"' {
				escapingDoubleQuote = true
				continue
			}

			if escapingDoubleQuote {
				escapingDoubleQuote = false
				continue
			}

			return "", 0, &CsvParseError{Line: parser.currentLine, Message: "double quotes within double quote enclosed fields must be escaped with a preceding double quote"}
		}
	}

	if parser.done {
		return input, len(input), nil
	}

	// The input doesn't contain a complete field. We need more data.
	return "", 0, nil
}

func (parser *CsvParser) peek(input string) byte {
	if parser.position >= len(input) {
		return 0
	}

	return input[parser.position]
}
