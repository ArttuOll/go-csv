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
}

const BUFFER_SIZE = 8

func NewCsvParser(r io.Reader) *CsvParser {
	return &CsvParser{
		reader: bufio.NewReader(r),
		buffer: make([]byte, BUFFER_SIZE),
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
	for {
		nextLine, err := parser.parseLine()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return records, nil
			}

			return nil, err
		}

		records = append(records, nextLine)
	}
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

		// Attempt to parse data received so far
		record, charsParsed, err := parser.parseRecord(string(parser.buffer[:parser.readToIndex]))
		if err != nil {
			return nil, err
		}

		// Not enough data in the buffer. Read more.
		if charsParsed == 0 {
			numberOfBytesInChunk, err := parser.reader.Read(parser.buffer[parser.readToIndex:])
			if err != nil {
				if errors.Is(err, io.EOF) {
					return record, err
				}

				return nil, err
			}

			parser.readToIndex += numberOfBytesInChunk
			continue
		}

		parser.currentLine++

		// Succeeded in parsing record. Remove the parsed record from the buffer.
		copy(parser.buffer, parser.buffer[charsParsed:])
		parser.readToIndex -= charsParsed
		return record, nil
	}
}

func (parser *CsvParser) parseRecord(input string) (fields []string, totalCharsRead int, err error) {
	position := 0
	for {
		field, charsRead, err := parser.parseField(input[position:])
		if err != nil {
			return nil, 0, err
		}

		// Need more data
		if charsRead == 0 {
			return nil, 0, nil
		}

		fields = append(fields, field)
		position += len(field)
		totalCharsRead += charsRead

		if input[position] == ',' {
			position++
			totalCharsRead++
		}

		if input[position] == '\r' {
			position++
			totalCharsRead++
		}

		if input[position] == '\n' {
			totalCharsRead++
			numberOfFields := len(fields)

			if parser.fieldsInARecord == 0 {
				parser.fieldsInARecord = len(fields)
			} else if numberOfFields != parser.fieldsInARecord {
				return nil, 0, &CsvParseError{Line: parser.currentLine, Message: fmt.Sprintf("failed to parse record. too many fields in a record: %v, but should be %v", numberOfFields, parser.fieldsInARecord)}
			}

			return fields, totalCharsRead, nil
		}

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

	// The input doesn't contain a complete field. We need more data.
	return "", 0, nil
}
