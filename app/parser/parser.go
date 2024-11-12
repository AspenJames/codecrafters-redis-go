package parser

type Parser interface {
	Parse() ParseResponse
}

type ParseResponse = interface{}
