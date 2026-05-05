package parser

import (
	"bufio"
	"errors"
	"fmt"
	"io"
)

type ParserState int

const (
	StartField ParserState = iota
	UnquotedField
	QuotedField
	AfterQuote
	FinishRecord
)

type CsvParser struct {
	reader          io.Reader
	fieldsInARecord int
	currentLine     int

	state ParserState

	buffer      []byte
	readToIndex int
	position    int
	done        bool

	fieldBuffer  []byte
	recordBuffer []string
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

func (e *CsvParseError) Error() string {
	return fmt.Sprintf("[Line %v]: %v", e.Line, e.Message)
}

func (p *CsvParser) ParseAll() ([][]string, error) {
	var records [][]string

	for !p.done {
		rec, err := p.parseLine()
		if err != nil {
			return records, err
		}

		if rec != nil {
			records = append(records, rec)
		}
	}

	return records, nil
}

func (p *CsvParser) Parse() ([]string, error) {
	return p.parseLine()
}

func (p *CsvParser) parseLine() ([]string, error) {
	for {
		// grow buffer if needed
		if p.readToIndex >= len(p.buffer) {
			newBuf := make([]byte, len(p.buffer)*2)
			copy(newBuf, p.buffer)
			p.buffer = newBuf
		}

		bytesRead, err := p.reader.Read(p.buffer[p.readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				p.done = true
			} else {
				return nil, err
			}
		}

		p.readToIndex += bytesRead

		record, consumed, complete, err := p.parse()
		if err != nil {
			return nil, err
		}

		if complete {
			// shift buffer
			copy(p.buffer, p.buffer[consumed:])
			p.readToIndex -= consumed
			p.position = 0

			return record, nil
		}

		if p.done {
			return nil, nil
		}
	}
}

func (p *CsvParser) parse() ([]string, int, bool, error) {
	for p.position < p.readToIndex {
		v := p.buffer[p.position]

		switch p.state {

		case StartField:
			switch v {
			case '"':
				p.state = QuotedField

			case ',':
				p.recordBuffer = append(p.recordBuffer, "")

			case '\r':
				p.state = FinishRecord

			default:
				p.state = UnquotedField
				p.fieldBuffer = append(p.fieldBuffer, v)
			}

		case UnquotedField:
			switch v {
			case '"':
				return nil, 0, false, &CsvParseError{
					Line:    p.currentLine,
					Message: "unexpected quote in unquoted field",
				}

			case ',':
				p.pushField()
				p.state = StartField

			case '\n':
				return nil, 0, false, &CsvParseError{
					Line:    p.currentLine,
					Message: "unexpected newline in unquoted field",
				}

			case '\r':
				p.state = FinishRecord

			default:
				p.fieldBuffer = append(p.fieldBuffer, v)
			}

		case QuotedField:
			switch v {
			case '"':
				p.state = AfterQuote
			// Withing quoted fields the line break can be anything. \n is always present.
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
				p.state = FinishRecord

			default:
				return nil, 0, false, &CsvParseError{
					Line:    p.currentLine,
					Message: "invalid character after closing quote",
				}
			}

		case FinishRecord:
			p.pushField()

			if v != '\n' {
				return nil, 0, false, &CsvParseError{
					Line:    p.currentLine,
					Message: "invalid character after carriage return at the end of record",
				}
			}

			p.currentLine++

			return p.finishRecord(p.position + 1)
		}

		p.position++
	}

	// EOF handling
	if p.done {
		switch p.state {
		case QuotedField:
			return nil, 0, false, &CsvParseError{
				Line:    p.currentLine,
				Message: "unterminated quoted field",
			}

		// the last record doesn't need to end in a newline
		case UnquotedField, AfterQuote:
			p.pushField()
			return p.finishRecord(p.position)

		case StartField:
			// the record had a trailing comma. this translates to an empty field
			if len(p.recordBuffer) > 0 {
				p.recordBuffer = append(p.recordBuffer, "")
				return p.finishRecord(p.position)
			}
		case FinishRecord:
			return nil, 0, false, &CsvParseError{
				Line:    p.currentLine,
				Message: "line feed missing from end of record",
			}
		}
	}

	return nil, 0, false, nil
}

func (p *CsvParser) pushField() {
	p.recordBuffer = append(p.recordBuffer, string(p.fieldBuffer))
	p.fieldBuffer = nil
}

func (p *CsvParser) finishRecord(consumed int) ([]string, int, bool, error) {
	record := p.recordBuffer

	if p.fieldsInARecord == 0 {
		p.fieldsInARecord = len(record)
	} else if len(record) != p.fieldsInARecord {
		return nil, 0, false, &CsvParseError{
			Line:    p.currentLine,
			Message: fmt.Sprintf("unexpected number of fields in a record: got %d expected %d", len(record), p.fieldsInARecord),
		}
	}

	// reset state
	p.recordBuffer = nil
	p.fieldBuffer = nil
	p.state = StartField

	return record, consumed, true, nil
}

func (p *CsvParser) ensureNext(expected byte) bool {
	if p.position+1 >= p.readToIndex {
		return false
	}

	return p.buffer[p.position+1] == expected
}
