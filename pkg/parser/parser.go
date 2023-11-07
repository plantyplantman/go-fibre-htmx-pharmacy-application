package parser

import (
	"encoding/csv"
	"io"
	"os"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/plantyplantman/bcapi/pkg/report"
)

type InvalidFileError struct {
	msg string
}

func (e *InvalidFileError) Error() string {
	if e.msg != "" {
		return e.msg
	}
	return "invalid file"
}

type Parser interface {
	Parse(interface{}) error
}

func NewCsvParser(file *os.File, opts ...ParserOption) (Parser, error) {
	if file == nil {
		return nil, &InvalidFileError{msg: "file is nil"}
	}

	p := csvParser{
		file: file,
		opts: parserOpts{
			comma:          '\t',
			lazyQuotes:     true,
			fieldPerRecord: -1,
		},
	}
	for _, opt := range opts {
		opt(&p.opts)
	}
	return &p, nil
}

type csvParser struct {
	file *os.File
	opts parserOpts
}

type parserOpts struct {
	comma          rune
	lazyQuotes     bool
	fieldPerRecord int
	isMultistore   bool
}

type ParserOption func(*parserOpts)

func (p *csvParser) Parse(dst interface{}) error {
	if p.opts.isMultistore {
		t, err := time.Parse("060102", "231030")
		if err != nil {
			return err
		}
		r := csv.NewReader(p.file)
		r.Comma = p.opts.comma
		r.LazyQuotes = p.opts.lazyQuotes
		// r.FieldsPerRecord = p.opts.fieldPerRecord
		rM, err := ParseMultistoreInput(r, t, "ms")
		if err != nil {
			return err
		}
		*dst.(*map[string]*report.ProductRetailList) = rM
		return nil
	}

	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
		r := csv.NewReader(in)
		r.Comma = p.opts.comma
		r.LazyQuotes = p.opts.lazyQuotes
		r.FieldsPerRecord = p.opts.fieldPerRecord
		return r
	})

	if err := gocsv.UnmarshalFile(p.file, dst); err != nil {
		return err
	}

	return nil
}

func IsMultistore(b bool) ParserOption {
	return func(opts *parserOpts) {
		opts.isMultistore = b
	}
}

func WithComma(r rune) ParserOption {
	return func(opts *parserOpts) {
		opts.comma = r
	}
}

func WithLazyQuotes(b bool) ParserOption {
	return func(opts *parserOpts) {
		opts.lazyQuotes = b
	}
}

func WithFieldsPerRecord(n int) ParserOption {
	return func(opts *parserOpts) {
		opts.fieldPerRecord = n
	}
}
