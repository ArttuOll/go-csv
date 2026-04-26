package parser

import (
	"bufio"
	"errors"
	"fmt"
	"io"
)

type ParserState int

const (
	StartField = iota
	UnquotedField
	EndField
	ParsingRecord
	QuotedField
	QuoteInQuotedField
	EndRecord
	Delimiter
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
	done            bool
}

const BUFFER_SIZE = 8

func NewCsvParser(r io.Reader) *CsvParser {
	return &CsvParser{
		reader:      bufio.NewReader(r),
		buffer:      make([]byte, BUFFER_SIZE),
		currentLine: 1,
		state:       StartField,
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
		record, charsParsed, err := parser.parseRecord()
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

		return record, nil
	}
}

func (parser *CsvParser) parseRecord() ([]string, int, error) {
	var fields []string
	for parser.position < parser.readToIndex {
		if parser.state != ParsingRecord {
			field, err := parser.parseField()
			if err != nil {
				return nil, 0, err
			}

			fields = append(fields, field)
			parser.state = ParsingRecord
			continue
		}

		if parser.buffer[parser.position] == ',' {
			parser.position++
			parser.state = StartField
			continue
		}

		// We've hit a delimiter (a newline and we know we're not inside a field.)
		if parser.buffer[parser.position] == '\r' {
			parser.position += 2
			parser.state = EndRecord
			break
		}

		return nil, 0, &CsvParseError{Line: parser.currentLine, Message: "failed to parse record. unexpected field separator."}
	}

	if parser.done && parser.state == StartField {
		return nil, 0, &CsvParseError{Line: parser.currentLine, Message: "failed to parse record. trailing comma"}
	}

	if parser.done || parser.state == EndRecord {
		numberOfFields := len(fields)

		if parser.fieldsInARecord == 0 {
			parser.fieldsInARecord = numberOfFields
		} else if numberOfFields != parser.fieldsInARecord {
			return nil, 0, &CsvParseError{Line: parser.currentLine, Message: fmt.Sprintf("failed to parse record. too many fields in a record: %v, but should be %v", numberOfFields, parser.fieldsInARecord)}
		}

		parser.state = StartField
		return fields, parser.position, nil
	}

	// The buffer doesn't contain a complete record. We need to read more data.
	parser.position = 0
	parser.state = StartField
	return nil, 0, nil
}

func (parser *CsvParser) parseField() (string, error) {
	var field []byte

parserLoop:
	for parser.position < parser.readToIndex {
		v := parser.buffer[parser.position]

		switch parser.state {
		case StartField:
			if v == '"' {
				// First character of the field is a double quote. The field is quoted.
				parser.state = QuotedField
				parser.position++
				continue
			}

			parser.state = UnquotedField

			field = append(field, v)
			parser.position++

		case UnquotedField:
			if v == '"' {
				return "", &CsvParseError{Line: parser.currentLine, Message: "fields containing double quotes must be enclosed by double quotes. the contained double quote must be escaped with a preceding double quote."}
			}

			if v == '\r' || v == ',' {
				parser.state = Delimiter
				continue
			}

			field = append(field, v)
			parser.position++

		case QuotedField:
			if v == '"' {
				parser.state = QuoteInQuotedField
				parser.position++
				continue
			}

			field = append(field, v)
			parser.position++

		case QuoteInQuotedField:
			if v == '"' {
				parser.state = QuotedField
				parser.position++
				continue
			}

			if v == '\r' || v == ',' {
				parser.state = Delimiter
				continue
			}

			return "", &CsvParseError{Line: parser.currentLine, Message: "double quotes within double quote enclosed fields must be escaped with a preceding double quote"}

		case Delimiter:
			parser.state = ParsingRecord
			break parserLoop
		}
	}

	if parser.state == ParsingRecord || parser.state == QuoteInQuotedField {
		return string(field), nil
	}

	if !parser.done {
		// The buffer didn't contain a complete field
		return "", nil
	}

	if parser.state == QuotedField {
		return "", &CsvParseError{Line: parser.currentLine, Message: "failed to parse field. unmatched double quote"}
	}

	// The buffer didn't contain a complete field
	return string(field), nil
}
