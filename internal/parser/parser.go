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
	FinishRecord
)

type CSVParser struct {
	reader          *bufio.Reader
	fieldsInARecord int
	currentLine     int

	state ParserState

	readToIndex int
	position    int
	done        bool

	fieldBuffer  []byte
	recordBuffer []string
}

func NewCSVParser(r io.Reader) *CSVParser {
	return &CSVParser{
		reader:      bufio.NewReader(r),
		currentLine: 1,
		state:       StartField,
	}
}

type CSVParseError struct {
	Line    int
	Message string
}

func (e *CSVParseError) Error() string {
	return fmt.Sprintf("[Line %v]: %v", e.Line, e.Message)
}

func (p *CSVParser) ParseAll() ([][]string, error) {
	var records [][]string

	for {
		rec, err := p.parseLine()
		if err != nil {
			return records, err
		}

		if rec == nil {
			return records, nil
		}

		records = append(records, rec)
	}

}

func (p *CSVParser) Parse() ([]string, error) {
	return p.parseLine()
}

func (p *CSVParser) parseLine() ([]string, error) {
	for {
		record, complete, err := p.parse()
		if err != nil {
			return nil, err
		}

		if complete {
			return record, nil
		}

		if p.done {
			break
		}
	}

	return nil, nil
}

func (p *CSVParser) parse() ([]string, bool, error) {
	for {
		v, err := p.reader.ReadByte()

		if err != nil {
			if err == io.EOF {
				p.done = true
				break
			}
		}

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
				return nil, false, &CSVParseError{
					Line:    p.currentLine,
					Message: "unexpected quote in unquoted field",
				}

			case ',':
				p.pushField()
				p.state = StartField

			case '\n':
				return nil, false, &CSVParseError{
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
				return nil, false, &CSVParseError{
					Line:    p.currentLine,
					Message: "invalid character after closing quote",
				}
			}

		case FinishRecord:
			p.pushField()

			if v != '\n' {
				return nil, false, &CSVParseError{
					Line:    p.currentLine,
					Message: "invalid character after carriage return at the end of record",
				}
			}

			p.currentLine++

			return p.finishRecord()
		}

		p.position++
	}

	// EOF handling
	if p.done {
		switch p.state {
		case QuotedField:
			return nil, false, &CSVParseError{
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
		case FinishRecord:
			return nil, false, &CSVParseError{
				Line:    p.currentLine,
				Message: "line feed missing from end of record",
			}
		}
	}

	return nil, false, nil
}

func (p *CSVParser) pushField() {
	p.recordBuffer = append(p.recordBuffer, string(p.fieldBuffer))
	p.fieldBuffer = p.fieldBuffer[:0]
}

func (p *CSVParser) finishRecord() ([]string, bool, error) {
	record := p.recordBuffer

	if p.fieldsInARecord == 0 {
		p.fieldsInARecord = len(record)
	} else if len(record) != p.fieldsInARecord {
		return nil, false, &CSVParseError{
			Line:    p.currentLine,
			Message: fmt.Sprintf("unexpected number of fields in a record: got %d expected %d", len(record), p.fieldsInARecord),
		}
	}

	// reset state
	p.recordBuffer = nil
	p.fieldBuffer = p.fieldBuffer[:0]
	p.state = StartField

	return record, true, nil
}
