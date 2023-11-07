package parser

import "encoding/xml"

func NewXmlParser(data []byte) Parser {
	return &xmlParser{
		data: data,
	}
}

type xmlParser struct {
	data []byte
}

func (p *xmlParser) Parse(dst interface{}) error {
	return xml.Unmarshal(p.data, dst)
}
