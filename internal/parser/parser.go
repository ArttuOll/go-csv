// Package parser provides a strictly RFC-4180 compliant CSV parser.
package parser

import (
	"bufio"
	"fmt"
	"io"
)

type ParserState int

const (
	StartField ParserState = iota
	UnquotedField
	QuotedField
	AfterQuote
)

// CSVParser parses CSV formatted text (as defined by RFC-4180).
type CSVParser struct {
	reader          *bufio.Reader
	fieldsInARecord int
	currentLine     int

	state ParserState

	fieldBuffer  []byte
	recordBuffer []string
}

// NewCSVParser creates a new CSVParser that reads CSV formatted text
// from r.
func NewCSVParser(r io.Reader) *CSVParser {
	return &CSVParser{
		reader:      bufio.NewReader(r),
		currentLine: 1,
		state:       StartField,
	}
}

// CSVParseError represents an error that occurred
// during CSV parsing.
type CSVParseError struct {
	Line    int    // Line on which the error was encountered.
	Message string // Message that explains the error.
}

// Error unwraps a CSVParseError into a string.
func (e *CSVParseError) Error() string {
	return fmt.Sprintf("line %v: %v", e.Line, e.Message)
}

// ParseAll parses records from CSVParser's reader until EOF is hit.
// The parsed records are then returned. EOF is not considered an error.
func (p *CSVParser) ParseAll() ([][]string, error) {
	var records [][]string

	for {
		rec, err := p.Parse()
		if err != nil {
			return records, err
		}

		if rec == nil {
			return records, nil
		}

		records = append(records, rec)
	}

}

// Parse parses a single record from CSVParser's reader.
// EOF is not considered an error.
func (p *CSVParser) Parse() ([]string, error) {
	record, err := p.parse()
	if err != nil && err != io.EOF {
		return nil, err
	}

	return record, nil
}

func (p *CSVParser) parse() ([]string, error) {
	for {
		v, err := p.reader.ReadByte()

		if err != nil {
			if err == io.EOF {
				switch p.state {
				case QuotedField:
					return nil, &CSVParseError{
						Line:    p.currentLine,
						Message: "unterminated quoted field",
					}

				// the last record doesn't need to end in a newline
				case UnquotedField, AfterQuote:
					p.pushField()
					return p.finishRecord()

				case StartField:
					// the record had a trailing comma. this translates to an empty field
					if len(p.recordBuffer) > 0 {
						p.recordBuffer = append(p.recordBuffer, "")
						return p.finishRecord()
					}
				}
			}

			return nil, err
		}

		switch p.state {

		case StartField:
			switch v {
			case '"':
				p.state = QuotedField

			case ',':
				p.recordBuffer = append(p.recordBuffer, "")

			case '\r':
				return p.parseLineEnding()

			default:
				p.state = UnquotedField
				p.fieldBuffer = append(p.fieldBuffer, v)
			}

		case UnquotedField:
			switch v {
			case '"':
				return nil, &CSVParseError{
					Line:    p.currentLine,
					Message: "unexpected quote in unquoted field",
				}

			case ',':
				p.pushField()
				p.state = StartField

			case '\n':
				return nil, &CSVParseError{
					Line:    p.currentLine,
					Message: "unexpected newline in unquoted field",
				}

			case '\r':
				return p.parseLineEnding()

			default:
				p.fieldBuffer = append(p.fieldBuffer, v)
			}

		case QuotedField:
			switch v {
			case '"':
				p.state = AfterQuote
			// Within quoted fields the line break can be anything. \n is always present.
			case '\n':
				p.currentLine++
			default:
				p.fieldBuffer = append(p.fieldBuffer, v)
			}

		case AfterQuote:
			switch v {
			case '"':
				p.fieldBuffer = append(p.fieldBuffer, '"')
				p.state = QuotedField

			case ',':
				p.pushField()
				p.state = StartField

			case '\r':
				return p.parseLineEnding()

			default:
				return nil, &CSVParseError{
					Line:    p.currentLine,
					Message: "invalid character after closing quote",
				}
			}
		}
	}
}

func (p *CSVParser) pushField() {
	p.recordBuffer = append(p.recordBuffer, string(p.fieldBuffer))
	p.fieldBuffer = p.fieldBuffer[:0]
}

func (p *CSVParser) finishRecord() ([]string, error) {
	record := p.recordBuffer

	if p.fieldsInARecord == 0 {
		p.fieldsInARecord = len(record)
	} else if len(record) != p.fieldsInARecord {
		return nil, &CSVParseError{
			Line:    p.currentLine,
			Message: fmt.Sprintf("unexpected number of fields in a record: got %d expected %d", len(record), p.fieldsInARecord),
		}
	}

	// reset state
	p.recordBuffer = nil
	p.fieldBuffer = p.fieldBuffer[:0]
	p.state = StartField

	return record, nil
}

func (p *CSVParser) parseLineEnding() ([]string, error) {
	next, err := p.reader.ReadByte()
	if err != nil {
		if err == io.EOF {
			return nil, &CSVParseError{
				Line:    p.currentLine,
				Message: "line feed missing from end of record",
			}
		}

		return nil, err
	}

	p.pushField()

	if next != '\n' {
		return nil, &CSVParseError{
			Line:    p.currentLine,
			Message: "invalid character after carriage return at the end of record",
		}
	}

	p.currentLine++

	return p.finishRecord()
}
